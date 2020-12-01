package main

import (
	// "crypto/tls"
	"encoding/json"
	"fmt"
	// "io/ioutil"
	// "net"
	// "strings"
	"time"
	"errors"
	"github.com/gocql/gocql"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"golang.org/x/net/context"
)

type CassandraDatasource struct {
	plugin.NetRPCUnsupportedPlugin
	logger hclog.Logger
	session *gocql.Session
}

type ColumnInfo struct {
	Name string
	Type string
}

func (ds *CassandraDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug(fmt.Sprintf("TSDB Request: %+v\n", tsdbReq))

	queries, err := ds.parseJSONQueries(tsdbReq)
	if err != nil {
		return nil, err
	}

	queryType, err := ds.GetRequestType(queries)
	if err != nil {
		return nil, err
	}

	options, err := ds.GetRequestOptions(tsdbReq)
	if err != nil {
		return nil, err
	}

    _, err = ds.Connect(
		tsdbReq.Datasource.Url,
		options["keyspace"],
		options["user"],
		tsdbReq.Datasource.DecryptedSecureJsonData["password"],
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
		return ds.SearchQuery(tsdbReq, queries)
	case "query":
		return ds.MetricQuery(tsdbReq, queries, options)
	case "connection":
		return &datasource.DatasourceResponse{}, nil;
	default:
		return nil, errors.New(fmt.Sprintf("Unknown query type '%s'", queryType))
	}
}

func (ds *CassandraDatasource) MetricQuery(tsdbReq *datasource.DatasourceRequest, jsonQueries []*simplejson.Json, options map[string]string) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug(fmt.Sprintf("Query[0]: %v\n", jsonQueries[0]))

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

			preparedQuery = queryData.Get("target").MustString()

		} else {

			allowFiltering := ""		
			if queryData.Get("filtering").MustBool() { 
				ds.logger.Warn("Allow filtering enabled")
				allowFiltering = " ALLOW FILTERING" 
			}
			preparedQuery = fmt.Sprintf(
				"SELECT %s, CAST(%s as double) FROM %s.%s WHERE %s = %s%s",
				queryData.Get("columnTime").MustString(),
				queryData.Get("columnValue").MustString(),
				queryData.Get("keyspace").MustString(),
				queryData.Get("table").MustString(),
				queryData.Get("columnId").MustString(),
				queryData.Get("valueId").MustString(),
				allowFiltering,
			)
			
		} 		

		ds.logger.Debug(fmt.Sprintf("Executing CQL query: '%s' ...\n", preparedQuery))

		iter := ds.session.Query(preparedQuery).Iter()

		var id string
		var timestamp time.Time
		var value float64


		if queryData.Get("rawQuery").MustBool() {

			series := make(map[string]datasource.TimeSeries)

			for iter.Scan(&id, &value, &timestamp) {
				if _, ok := series[id]; ok {
					series[id] = datasource.TimeSeries{
						Name: id,
						Points: append(series[id].Points, &datasource.Point{
							Timestamp: timestamp.UnixNano() / int64(time.Millisecond),
							Value:     value,
						}),
					}
				}				
			}

			if err := iter.Close(); err != nil {
				ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
				return &datasource.DatasourceResponse{
					Results: []*datasource.QueryResult{
						&datasource.QueryResult{
							Error: err.Error(),
						},
					},
				}, nil
			}

			for _, serie2 := range series{
				queryResult.Series = append(queryResult.Series, &serie2)
			}
			
			response.Results = append(response.Results, &queryResult)

		} else {

			serie := &datasource.TimeSeries{Name: queryData.Get("valueId").MustString()}

			for iter.Scan(&timestamp, &value) {
				serie.Points = append(serie.Points, &datasource.Point{
					Timestamp: timestamp.UnixNano() / int64(time.Millisecond),
					Value:     value,
				})
			}
			if err := iter.Close(); err != nil {
				ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
				return &datasource.DatasourceResponse{
					Results: []*datasource.QueryResult{
						&datasource.QueryResult{
							Error: err.Error(),
						},
					},
				}, nil
			}
	
			queryResult.Series = append(queryResult.Series, serie)
	
			response.Results = append(response.Results, &queryResult)

		}

	}

	return response, nil
}

func (ds *CassandraDatasource) SearchQuery(tsdbReq *datasource.DatasourceRequest, jsonQueries []*simplejson.Json) (*datasource.DatasourceResponse, error) {	
	keyspaceName, keyspaceOk := jsonQueries[0].CheckGet("keyspace")
	tableName, tableOk       := jsonQueries[0].CheckGet("table")

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
				RefId:  "search",
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

func (ds *CassandraDatasource) parseJSONQueries(tsdbReq *datasource.DatasourceRequest) ([]*simplejson.Json, error) {
	queries := make([]*simplejson.Json, 0)
	if len(tsdbReq.Queries) < 1 {
		ds.logger.Error("No queries to parse, unable to proceed")
		return nil, errors.New("No queries in TSDB Request")
	}
	for _, query := range tsdbReq.Queries {
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

func (ds *CassandraDatasource) GetRequestType(queries []*simplejson.Json) (string, error) {
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

func (ds *CassandraDatasource) GetRequestOptions(tsdbReq *datasource.DatasourceRequest) (map[string]string, error) {
	var options map[string]string
	err := json.Unmarshal([]byte(tsdbReq.Datasource.JsonData), &options)
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to get request JSON data: %s\n", err))
		return nil, err
	}
	return options, nil
}

func (ds *CassandraDatasource) Connect(host string, keyspace string, username string, password string) (bool, error) {
	if (ds.session != nil) {
		return true, nil
	}

	ds.logger.Debug(fmt.Sprintf("Connecting to %s...\n", host))
	cluster := gocql.NewCluster(host)
	cluster.Keyspace = keyspace
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: username,
		Password: password,
	}
	cluster.Consistency = gocql.One

	session, err := cluster.CreateSession()
	if err != nil {
		ds.logger.Error(fmt.Sprintf("Unable to establish connection with the database, %s\n", err.Error()))
		return false, err;
	}
	ds.session = session

	ds.logger.Debug(fmt.Sprintf("Connection successful.\n"))
	return true, nil;
}
