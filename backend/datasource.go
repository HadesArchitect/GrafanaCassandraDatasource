package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"strconv"
	"time"

	"github.com/gocql/gocql"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/net/context"
)

type CassandraDatasource struct {
	logger    log.Logger
	builder   *QueryBuilder
	processor *QueryProcessor
	sessions  map[int]*gocql.Session
	session   *gocql.Session
}

func (ds *CassandraDatasource) HandleMetricQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(req, ds.handleMetricQuery),
	}, nil
}

func (ds *CassandraDatasource) handleMetricQuery(ctx *backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	_, err := ds.connect(ctx)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("unable to establish connection with the database, err=%v", err))
	}

	frames, err := ds.metricQuery(&query)

	return dataResponse(frames, err)
}

func (ds *CassandraDatasource) HandleMetricFindQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(req, ds.handleMetricFindQuery),
	}, nil
}

func (ds *CassandraDatasource) handleMetricFindQuery(ctx *backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	_, err := ds.connect(ctx)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("unable to establish connection with the database, err=%v", err))
	}

	var cassQuery CassandraQuery
	err = json.Unmarshal(query.JSON, &cassQuery)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("unmarshal queries, err=%v", err))
	}

	if cassQuery.Keyspace == "" || cassQuery.Table == "" {
		return dataResponse(data.Frames{}, errors.New("keyspace or table is not set"))
	}

	keyspaceMetadata, err := ds.session.KeyspaceMetadata(cassQuery.Keyspace)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("retrieve keyspace metadata, err=%v, keyspace=%s", err, cassQuery.Keyspace))
	}

	tableMetadata, ok := keyspaceMetadata.Tables[cassQuery.Table]
	if !ok {
		return dataResponse(data.Frames{}, fmt.Errorf("table not found, table=%s", cassQuery.Table))
	}

	frame := data.NewFrame(
		query.RefID,
		data.NewField(query.RefID, nil, make([]string, len(tableMetadata.Columns))),
		data.NewField("type", nil, make([]string, len(tableMetadata.Columns))),
	)

	index := 0
	for name, column := range tableMetadata.Columns {
		frame.Set(0, index, name)
		frame.Set(1, index, column.Type.Type().String())

		index++
	}

	return dataResponse([]*data.Frame{frame}, nil)
}

func (ds *CassandraDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) error {
	_, err := ds.connect(&req.PluginContext)
	if err != nil {
		return fmt.Errorf("unable to establish connection with the database, err=%v", err)
	}

	return nil
}

func processQueries(req *backend.QueryDataRequest, handler func(*backend.PluginContext, backend.DataQuery) backend.DataResponse) backend.Responses {
	result := backend.Responses{}
	for _, query := range req.Queries {
		result[query.RefID] = handler(&req.PluginContext, query)
	}

	return result
}

func (ds *CassandraDatasource) metricQuery(query *backend.DataQuery) (data.Frames, error) {
	var cassQuery CassandraQuery
	err := json.Unmarshal(query.JSON, &cassQuery)
	if err != nil {
		return nil, fmt.Errorf("unmarshal queries, err=%v", err)
	}

	if cassQuery.RawQuery {
		return ds.processor.processRawMetricQuery(cassQuery.Target, ds)
	} else {
		from, to := timeRangeToStr(query.TimeRange)
		ds.logger.Debug(fmt.Sprintf("Timeframe from: %s to %s\n", from, to))

		preparedQuery := ds.builder.prepareStrictMetricQuery(&cassQuery, from, to)

		return ds.processor.processStrictMetricQuery(preparedQuery, cassQuery.ValueID, ds)
	}
}

func (ds *CassandraDatasource) getRequestOptions(jsonData []byte) (DataSourceOptions, error) {
	var options DataSourceOptions
	err := json.Unmarshal(jsonData, &options)
	if err != nil {
		return options, fmt.Errorf("get request JSON data, err=%v", err)
	}
	return options, nil
}

func (ds *CassandraDatasource) connect(context *backend.PluginContext) (bool, error) {
	options, err := ds.getRequestOptions(context.DataSourceInstanceSettings.JSONData)
	if err != nil {
		return false, fmt.Errorf("parse connection parameters, err=%v", err)
	}

	datasourceID := int(context.DataSourceInstanceSettings.ID)

	if ds.sessions == nil {
		ds.sessions = make(map[int]*gocql.Session)
	}

	if ds.sessions[datasourceID] != nil {
		ds.session = ds.sessions[datasourceID]

		return true, nil
	}

	host := context.DataSourceInstanceSettings.URL

	ds.logger.Debug(fmt.Sprintf("Connecting to %s...\n", host))

	cluster := gocql.NewCluster(host)

	if options.Timeout != nil {
		cluster.Timeout = time.Duration(*options.Timeout) * time.Second
	}

	consistency, err := parseConsistency(options.Consistency)
	if err != nil {
		return false, fmt.Errorf("parse Consistency, err=%v, consistency string=%s", err, options.Consistency)
	}

	cluster.Consistency = consistency
	cluster.Keyspace = options.Keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: options.User,
		Password: context.DataSourceInstanceSettings.DecryptedSecureJSONData["password"],
	}
	cluster.DisableInitialHostLookup = true // AWS Specific Required

	if options.UseCustomTLS {
		ds.logger.Debug("Setting TLS Configuration...")

		tlsConfig, err := PrepareTLSCfg(options.CertPath, options.RootPath, options.CaPath, options.AllowInsecureTLS)
		if err != nil {
			return false, fmt.Errorf("create TLS config, err=%v", err)
		}

		cluster.SslOpts = &gocql.SslOptions{Config: tlsConfig}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return false, fmt.Errorf("create Cassandra DB session, err=%v", err)
	}

	ds.sessions[datasourceID] = session
	ds.session = session

	ds.logger.Debug("Connection successful")
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

func timeRangeToStr(timeRange backend.TimeRange) (string, string) {
	from := strconv.FormatInt(timeRange.From.UnixNano()/int64(time.Millisecond), 10)
	to := strconv.FormatInt(timeRange.To.UnixNano()/int64(time.Millisecond), 10)

	return from, to
}

func dataResponse(frames data.Frames, err error) backend.DataResponse {
	if err != nil {
		return backend.DataResponse{
			Error: err,
		}
	}

	return backend.DataResponse{
		Frames: frames,
	}
}
