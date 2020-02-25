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
}

//var session gocql.Session;

func (ds *CassandraDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug(fmt.Sprintf("TSDB Request: %+v\n", tsdbReq))

	queryType, err := GetQueryType(tsdbReq)
	if err != nil {
		return nil, err
	}

	queries, err := parseJSONQueries(tsdbReq)
	if err != nil {
		return nil, err
	}

	var options map[string]string
	json.Unmarshal([]byte(tsdbReq.Datasource.JsonData), &options)

	cluster := gocql.NewCluster(tsdbReq.Datasource.Url)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: options["user"],
		Password: tsdbReq.Datasource.DecryptedSecureJsonData["password"],
	}
	cluster.Keyspace = options["keyspace"]
	cluster.Consistency = gocql.One
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err;
	}
	defer session.Close()

	switch queryType {
	case "search":
		return ds.SearchQuery(ctx, tsdbReq, queries)
	case "query":
		return ds.MetricQuery(session, tsdbReq, queries, options)
	case "connection":
		return &datasource.DatasourceResponse{}, nil;
	default:
		return nil, errors.New(fmt.Sprintf("Unknown query type '%s'", queryType))
	}
}

func (ds *CassandraDatasource) MetricQuery(session *gocql.Session, tsdbReq *datasource.DatasourceRequest, jsonQueries []*simplejson.Json, options map[string]string) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug(fmt.Sprintf("Query[0]: %v\n", jsonQueries[0]))

	response := &datasource.DatasourceResponse{}

	for _, queryData := range jsonQueries {

		queryResult := datasource.QueryResult{
			RefId:  queryData.Get("refId").MustString(),
			Series: make([]*datasource.TimeSeries, 0),
		}

		serie := &datasource.TimeSeries{Name: queryData.Get("valueId").MustString()}

		var created_at time.Time
		var value int
		preparedQuery := fmt.Sprintf(
			"SELECT %s, %s FROM %s.%s WHERE %s = %s", 
			queryData.Get("columnTime").MustString(),
			queryData.Get("columnValue").MustString(),
			queryData.Get("keyspace").MustString(),
			queryData.Get("table").MustString(),
			queryData.Get("columnId").MustString(),
			queryData.Get("valueId").MustString(),
		)
		iter := session.Query(preparedQuery).Iter()
		for iter.Scan(&created_at, &value) {
			serie.Points = append(serie.Points, &datasource.Point{
				Timestamp: created_at.UnixNano() / int64(time.Millisecond),
				Value:     float64(value),
			})
		}
		if err := iter.Close(); err != nil {
			return nil, err;
		}

		queryResult.Series = append(queryResult.Series, serie)

		response.Results = append(response.Results, &queryResult)
	}

	return response, nil
}

func parseJSONQueries(tsdbReq *datasource.DatasourceRequest) ([]*simplejson.Json, error) {
	jsonQueries := make([]*simplejson.Json, 0)
	for _, query := range tsdbReq.Queries {
		json, err := simplejson.NewJson([]byte(query.ModelJson))
		if err != nil {
			return nil, err
		}

		jsonQueries = append(jsonQueries, json)
	}
	return jsonQueries, nil
}

// func (ds *CassandraDatasource) CreateMetricRequest(tsdbReq *datasource.DatasourceRequest) (*RemoteDatasourceRequest, error) {
	// jsonQueries, err := parseJSONQueries(tsdbReq)
	// if err != nil {
	// 	return nil, err
	// }

	// payload := simplejson.New()
	// payload.SetPath([]string{"range", "to"}, tsdbReq.TimeRange.ToRaw)
	// payload.SetPath([]string{"range", "from"}, tsdbReq.TimeRange.FromRaw)
	// payload.Set("targets", jsonQueries)

	// rbody, err := payload.MarshalJSON()
	// if err != nil {
	// 	return nil, err
	// }

	// url := tsdbReq.Datasource.Url + "/query"
	// req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rbody)))
	// if err != nil {
	// 	return nil, err
	// }

	// req.Header.Add("Content-Type", "application/json")

	// return &RemoteDatasourceRequest{
	// 	queryType: "query",
	// 	req:       req,
	// 	queries:   jsonQueries,
	// }, nil
// }

func (ds *CassandraDatasource) SearchQuery(ctx context.Context, tsdbReq *datasource.DatasourceRequest, jsonQueries []*simplejson.Json) (*datasource.DatasourceResponse, error) {
	return nil, errors.New("Not implemented yet")
}

func (ds *CassandraDatasource) Execute(ctx context.Context, remoteDsReq *RemoteDatasourceRequest) ([]byte, error) {
	return nil, errors.New("Not implemented yet")
}

func GetQueryType(tsdbReq *datasource.DatasourceRequest) (string, error) {
	queryType := "query"
	if len(tsdbReq.Queries) > 0 {
		firstQuery := tsdbReq.Queries[0]
		queryJson, err := simplejson.NewJson([]byte(firstQuery.ModelJson))
		if err != nil {
			return "", err
		}
		queryType = queryJson.Get("queryType").MustString("query")
	}
	return queryType, nil
}

func (ds *CassandraDatasource) ParseQueryResponse(queries []*simplejson.Json, body []byte) (*datasource.DatasourceResponse, error) {
	response := &datasource.DatasourceResponse{}
	responseBody := []TargetResponseDTO{}

	for i, r := range responseBody {
		refId := r.Target

		if len(queries) > i {
			refId = queries[i].Get("refId").MustString()
		}

		qr := datasource.QueryResult{
			RefId:  refId,
			Series: make([]*datasource.TimeSeries, 0),
		}

		serie := &datasource.TimeSeries{Name: r.Target}

		for _, p := range r.DataPoints {
			serie.Points = append(serie.Points, &datasource.Point{
				Timestamp: int64(p[1]),
				Value:     p[0],
			})
		}

		qr.Series = append(qr.Series, serie)
	
		response.Results = append(response.Results, &qr)
	}

	return response, nil
}

func (ds *CassandraDatasource) ParseSearchResponse(body []byte) (*datasource.DatasourceResponse, error) {
	return nil, errors.New("Not implemented yet")
}
