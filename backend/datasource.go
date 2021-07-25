package main

import (
	// "crypto/tls"
	"encoding/json"
	"fmt"

	// "io/ioutil"
	// "net"
	// "strings"
	"errors"
	"strconv"
	"time"

	"github.com/gocql/gocql"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/grafana/grafana-plugin-model/go/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	gflog "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	plugin "github.com/hashicorp/go-plugin"
	"golang.org/x/net/context"
)

type CassandraDatasource struct {
	plugin.NetRPCUnsupportedPlugin
	logger    gflog.Logger
	builder   *QueryBuilder
	processor *QueryProcessor
	sessions  map[int]*gocql.Session
	session   *gocql.Session
	settings  backend.DataSourceInstanceSettings
}

type ColumnInfo struct {
	Name string
	Type string
}

func (ds *CassandraDatasource) QueryData(ctx context.Context, request *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	ds.logger.Debug(fmt.Sprintf("TSDB Request: %+v\n", request))

	queries, err := ds.parseJSONQueries1(request.Queries)
	if err != nil {
		return nil, err
	}

	/*for _, query := range queries {
		logger.Debug(fmt.Sprintf("Queries: %+v\n", query))
	}*/

	queryType, err := ds.getRequestType(queries)
	if err != nil {
		return nil, err
	}

	options, err := ds.getRequestOptions1(request.PluginContext.DataSourceInstanceSettings.JSONData)
	if err != nil {
		return nil, err
	}

	_, err = ds.Connect(
		request.PluginContext.DataSourceInstanceSettings.URL,
		options["keyspace"],
		options["user"],
		request.PluginContext.DataSourceInstanceSettings.DecryptedSecureJSONData["password"],
		options["certPath"],
		options["rootPath"],
		options["caPath"],
		int(request.PluginContext.DataSourceInstanceSettings.ID),
		WithConsistency(options["consistency"]),
	)

	if err != nil {

		errResponse := backend.QueryDataResponse{
			Responses: make(map[string]backend.DataResponse),
		}

		errResponse.Responses[request.Queries[0].RefID] = backend.DataResponse{
			Error: errors.New("Unable to establish connection with the database"),
		}

		return &errResponse, nil
	}

	switch queryType {
	//case "search":
	//	return ds.searchQuery(queries)
	case "query":
		return ds.metricQuery1(queries, request.Queries[0].TimeRange.From, request.Queries[0].TimeRange.To)
	case "connection":
		return &backend.QueryDataResponse{}, nil
	default:
		return nil, fmt.Errorf("Unknown query type '%s'", queryType)
	}

	return nil, fmt.Errorf("Unknown query type")
}

func (ds *CassandraDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug(fmt.Sprintf("TSDB Request: %+v\n", tsdbReq))

	queries, err := ds.parseJSONQueries(tsdbReq.Queries)
	if err != nil {
		return nil, err
	}

	queryType, err := ds.getRequestType(queries)
	if err != nil {
		return nil, err
	}

	datasourceID, err := ds.getDatasourceID(queries)
	if err != nil {
		return nil, err
	}

	// reset connection in case "save & test" in datasource configuration
	if queryType == "connection" && ds.sessions != nil {
		ds.sessions[datasourceID] = nil
		ds.session = nil
	}

	options, err := ds.getRequestOptions(tsdbReq.Datasource.JsonData)
	if err != nil {
		return nil, err
	}

	_, err = ds.Connect(
		tsdbReq.Datasource.Url,
		options["keyspace"],
		options["user"],
		tsdbReq.Datasource.DecryptedSecureJsonData["password"],
		options["certPath"],
		options["rootPath"],
		options["caPath"],
		datasourceID,
		WithConsistency(options["consistency"]),
	)
	if err != nil {
		return &datasource.DatasourceResponse{
			Results: []*datasource.QueryResult{
				&datasource.QueryResult{
					Error: "Unable to establish connection with the database",
				},
			},
		}, nil
	}

	switch queryType {
	case "search":
		return ds.searchQuery(queries)
	case "query":
		return ds.metricQuery(queries, tsdbReq.TimeRange.FromRaw, tsdbReq.TimeRange.ToRaw)
	case "connection":
		return &datasource.DatasourceResponse{}, nil
	default:
		return nil, fmt.Errorf("Unknown query type '%s'", queryType)
	}
}

func (ds *CassandraDatasource) metricQuery(jsonQueries []*simplejson.Json, timeFrom string, timeTo string) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug(fmt.Sprintf("Query[0]: %v\n", jsonQueries[0]))

	ds.logger.Debug(fmt.Sprintf("Timeframe from: %+v\n", timeFrom))
	ds.logger.Debug(fmt.Sprintf("Timeframe to: %+v\n", timeTo))

	response := &datasource.DatasourceResponse{}

	for _, queryData := range jsonQueries {

		queryResult := datasource.QueryResult{
			RefId:  queryData.Get("refId").MustString(),
			Series: make([]*datasource.TimeSeries, 0),
		}

		ds.logger.Debug(fmt.Sprintf("rawQuery: %v\n", queryData.Get("rawQuery").MustBool()))
		ds.logger.Debug(fmt.Sprintf("target: %s\n", queryData.Get("target").MustString()))

		var preparedQuery string
		if queryData.Get("rawQuery").MustBool() {
			preparedQuery = ds.builder.prepareRawMetricQuery(queryData, timeFrom, timeTo)
		} else {
			preparedQuery = ds.builder.prepareStrictMetricQuery(queryData, timeFrom, timeTo)
		}

		ds.logger.Debug(fmt.Sprintf("Executing CQL query: '%s' ...\n", preparedQuery))

		if queryData.Get("rawQuery").MustBool() {
			ds.processor.processRawMetricQuery(&queryResult, preparedQuery, ds)
		} else {
			valueID := queryData.Get("valueId").MustString()
			ds.processor.processStrictMetricQuery(&queryResult, preparedQuery, valueID, ds)
		}

		response.Results = append(response.Results, &queryResult)
	}

	return response, nil
}

func (ds *CassandraDatasource) metricQuery1(jsonQueries []*simplejson.Json, from time.Time, to time.Time) (*backend.QueryDataResponse, error) {
	ds.logger.Debug(fmt.Sprintf("Query[0]: %v\n", jsonQueries[0]))

	ds.logger.Debug(fmt.Sprintf("Timeframe from: %+v\n", from.Unix()))
	ds.logger.Debug(fmt.Sprintf("Timeframe to: %+v\n", to.Unix()))

	response := &backend.QueryDataResponse{
		Responses: make(map[string]backend.DataResponse),
	}

	for _, query := range jsonQueries {

		ds.logger.Debug(fmt.Sprintf("rawQuery: %v\n", query.Get("rawQuery").MustBool()))
		ds.logger.Debug(fmt.Sprintf("target: %s\n", query.Get("target").MustString()))

		var preparedQuery string
		if query.Get("rawQuery").MustBool() {
			preparedQuery = ds.builder.prepareRawMetricQuery(query, strconv.FormatInt(from.Unix(), 10), strconv.FormatInt(to.Unix(), 10))
		} else {
			preparedQuery = ds.builder.prepareStrictMetricQuery(query, strconv.FormatInt(from.Unix(), 10), strconv.FormatInt(to.Unix(), 10))
		}

		ds.logger.Debug(fmt.Sprintf("Executing CQL query: '%s' ...\n", preparedQuery))

		dataResponse := backend.DataResponse{}
		if query.Get("rawQuery").MustBool() {
			ds.processor.processRawMetricQuery1(&dataResponse, preparedQuery, ds)
		} /* else {
			valueID := query.Get("valueId").MustString()
			ds.processor.processStrictMetricQuery(&queryResult, preparedQuery, valueID, ds)
		}*/

		ds.logger.Debug("%v", dataResponse)

		response.Responses[query.Get("refId").MustString()] = dataResponse

		ds.logger.Debug("%v", response)
		ds.logger.Debug("%v", dataResponse.Frames)
		for _, frame := range dataResponse.Frames {
			ds.logger.Debug("%v", frame)
		}
		//response.Results = append(response.Results, &queryResult)
	}

	return response, nil
}

func (ds *CassandraDatasource) searchQuery(jsonQueries []*simplejson.Json) (*datasource.DatasourceResponse, error) {
	keyspaceName, keyspaceOk := jsonQueries[0].CheckGet("keyspace")
	tableName, tableOk := jsonQueries[0].CheckGet("table")

	if !keyspaceOk || !tableOk {
		ds.logger.Warn("Unable to search as keyspace or table is not set")
		return nil, errors.New("Keyspace or table is not set")
	}

	keyspaceMetadata, err := ds.session.KeyspaceMetadata(keyspaceName.MustString())
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Error while retrieving keyspace metadata: %s\n", err.Error()))
		return nil, err
	}

	tableMetadata, ok := keyspaceMetadata.Tables[tableName.MustString()]
	if !ok {
		ds.logger.Debug(fmt.Sprintf("Unknown table '%s'\n", tableName))
		return nil, errors.New("Unknown table")
	}

	columns := make([]*ColumnInfo, 0)
	for name, column := range tableMetadata.Columns {
		columns = append(
			columns,
			&ColumnInfo{
				Name: name,
				Type: column.Type.Type().String(),
			},
		)
	}

	serialisedColumns, err := json.Marshal(columns)
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Error while serialising: %s\n", err.Error()))
		return nil, errors.New("Unable to process request, see logs for more details")
	}

	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			&datasource.QueryResult{
				RefId: "search",
				Tables: []*datasource.Table{
					&datasource.Table{
						Rows: []*datasource.TableRow{
							&datasource.TableRow{
								Values: []*datasource.RowValue{
									&datasource.RowValue{
										Kind:        datasource.RowValue_TYPE_STRING,
										StringValue: string(serialisedColumns),
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}

func (ds *CassandraDatasource) parseJSONQueries(rawQueries []*datasource.Query) ([]*simplejson.Json, error) {
	queries := make([]*simplejson.Json, 0)
	if len(rawQueries) < 1 {
		ds.logger.Error("No queries to parse, unable to proceed")
		return nil, errors.New("No queries in TSDB Request")
	}
	for _, query := range rawQueries {
		json, err := simplejson.NewJson([]byte(query.ModelJson))
		if err != nil {
			ds.logger.Error(fmt.Sprintf("Unable to parse json query: %s\n", err.Error()))
			return nil, err
		}

		queries = append(queries, json)
	}

	ds.logger.Debug(fmt.Sprintf("Parsed queries: %v\n", len(queries)))
	return queries, nil
}

func (ds *CassandraDatasource) parseJSONQueries1(rawQueries []backend.DataQuery) ([]*simplejson.Json, error) {
	queries := make([]*simplejson.Json, 0)
	if len(rawQueries) < 1 {
		ds.logger.Error("No queries to parse, unable to proceed")
		return nil, errors.New("No queries in TSDB Request")
	}
	for _, query := range rawQueries {
		json, err := simplejson.NewJson(query.JSON)
		if err != nil {
			ds.logger.Error(fmt.Sprintf("Unable to parse json query: %s\n", err.Error()))
			return nil, err
		}

		queries = append(queries, json)
	}

	ds.logger.Debug(fmt.Sprintf("Parsed queries: %v\n", len(queries)))
	return queries, nil
}

func (ds *CassandraDatasource) getRequestType(queries []*simplejson.Json) (string, error) {
	queryTypeProperty, exist := queries[0].CheckGet("queryType")
	if !exist {
		ds.logger.Error("Query type is not set, unable to proceed")
		return "", errors.New("No query type specified")
	}
	queryType, err := queryTypeProperty.String()
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to get QueryType: %s\n", err))
		return "", err
	}
	ds.logger.Debug(fmt.Sprintf("Query type: %s\n", queryType))
	return queryType, nil
}

func (ds *CassandraDatasource) getDatasourceID(queries []*simplejson.Json) (int, error) {
	datasourceIDRaw, idOk := queries[0].CheckGet("datasourceId")
	if !idOk {
		ds.logger.Error("ID is not set")
		return -1, errors.New("No datasource ID")
	}
	datasourceID, err := datasourceIDRaw.Int()
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to get datasource ID: %s\n", err))
		return -1, err
	}
	ds.logger.Debug(fmt.Sprintf("Datasource ID: %d\n", datasourceID))
	return datasourceID, nil
}

func (ds *CassandraDatasource) getRequestOptions(jsonData string) (map[string]string, error) {
	var options map[string]string
	err := json.Unmarshal([]byte(jsonData), &options)
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to get request JSON data: %s\n", err))
		return nil, err
	}
	return options, nil
}

func (ds *CassandraDatasource) getRequestOptions1(jsonData []byte) (map[string]string, error) {
	var options map[string]string
	err := json.Unmarshal(jsonData, &options)
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to get request JSON data: %s\n", err))
		return nil, err
	}
	return options, nil
}

type Option func(config *gocql.ClusterConfig) error

func WithConsistency(consistencyStr string) Option {
	return func(config *gocql.ClusterConfig) error {
		if consistencyStr == "" {
			consistencyStr = "QUORUM"
		}
		consistency, err := parseConsistency(consistencyStr)
		if err != nil {
			return err
		}
		config.Consistency = consistency
		return nil
	}
}

func (ds *CassandraDatasource) Connect(host, keyspace, username, password string, certPath string, rootPath string, caPath string, datasourceID int, opts ...Option) (bool, error) {

	if ds.sessions == nil {
		ds.sessions = make(map[int]*gocql.Session)
	}

	if ds.sessions[datasourceID] != nil {
		ds.session = ds.sessions[datasourceID]
		return true, nil
	}
	ds.logger.Debug(fmt.Sprintf("Connecting to %s...\n", host))
	cluster := gocql.NewCluster(host)
	for _, opt := range opts {
		if err := opt(cluster); err != nil {
			ds.logger.Error(fmt.Sprintf("Failed to apply option: %s\n", err.Error()))
			return false, err
		}
	}
	cluster.Keyspace = keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
	cluster.DisableInitialHostLookup = true // AWS Specific Required
	enableTLS := (certPath != "" && rootPath != "") || caPath != ""
	if enableTLS {
		tlsConfig, err := PrepareTLSCfg(certPath, rootPath, caPath)
		if err != nil {
			ds.logger.Error(fmt.Sprintf("Unable create tls config, %s\n", err.Error()))
			return false, err
		}
		cluster.SslOpts = &gocql.SslOptions{Config: tlsConfig}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to establish connection with the database, %s\n", err.Error()))
		return false, err
	}

	ds.sessions[datasourceID] = session
	ds.session = session

	ds.logger.Debug(fmt.Sprintf("Connection successful.\n"))
	return true, nil
}

func parseConsistency(consistencyStr string) (consistency gocql.Consistency, err error) {
	defer func() {
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("failed to parse consistency \"%s\": %v", consistencyStr, r)
			}
		}
	}()
	consistency = gocql.ParseConsistency(consistencyStr)
	return
}
