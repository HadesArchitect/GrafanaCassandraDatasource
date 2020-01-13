package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	simplejson "github.com/bitly/go-simplejson"
	"github.com/grafana/grafana_plugin_model/go/datasource"
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
	"golang.org/x/net/context"
	"golang.org/x/net/context/ctxhttp"
)

type JsonDatasource struct {
	plugin.NetRPCUnsupportedPlugin
	logger hclog.Logger
}

func (ds *JsonDatasource) Query(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	ds.logger.Debug("Query", "datasource", tsdbReq.Datasource.Name, "TimeRange", tsdbReq.TimeRange)

	queryType, err := GetQueryType(tsdbReq)
	if err != nil {
		return nil, err
	}

	ds.logger.Debug("createRequest", "queryType", queryType)

	switch queryType {
	case "search":
		return ds.SearchQuery(ctx, tsdbReq)
	default:
		return ds.MetricQuery(ctx, tsdbReq)
	}
}

func (ds *JsonDatasource) MetricQuery(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	remoteDsReq, err := ds.CreateMetricRequest(tsdbReq)
	if err != nil {
		return nil, err
	}

	body, err := ds.MakeHttpRequest(ctx, remoteDsReq)
	if err != nil {
		return nil, err
	}

	return ds.ParseQueryResponse(remoteDsReq.queries, body)
}

func (ds *JsonDatasource) CreateMetricRequest(tsdbReq *datasource.DatasourceRequest) (*RemoteDatasourceRequest, error) {
	jsonQueries, err := parseJSONQueries(tsdbReq)
	if err != nil {
		return nil, err
	}

	payload := simplejson.New()
	payload.SetPath([]string{"range", "to"}, tsdbReq.TimeRange.ToRaw)
	payload.SetPath([]string{"range", "from"}, tsdbReq.TimeRange.FromRaw)
	payload.Set("targets", jsonQueries)

	rbody, err := payload.MarshalJSON()
	if err != nil {
		return nil, err
	}

	url := tsdbReq.Datasource.Url + "/query"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rbody)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return &RemoteDatasourceRequest{
		queryType: "query",
		req:       req,
		queries:   jsonQueries,
	}, nil
}

func (ds *JsonDatasource) SearchQuery(ctx context.Context, tsdbReq *datasource.DatasourceRequest) (*datasource.DatasourceResponse, error) {
	remoteDsReq, err := ds.CreateSearchRequest(tsdbReq)
	if err != nil {
		return nil, err
	}

	body, err := ds.MakeHttpRequest(ctx, remoteDsReq)
	if err != nil {
		return nil, err
	}

	return ds.ParseSearchResponse(body)
}

func (ds *JsonDatasource) CreateSearchRequest(tsdbReq *datasource.DatasourceRequest) (*RemoteDatasourceRequest, error) {
	jsonQueries, err := parseJSONQueries(tsdbReq)
	if err != nil {
		return nil, err
	}

	payload := simplejson.New()
	payload.Set("target", jsonQueries[0].Get("target").MustString())

	rbody, err := payload.MarshalJSON()
	if err != nil {
		return nil, err
	}

	url := tsdbReq.Datasource.Url + "/search"
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(rbody)))
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

	return &RemoteDatasourceRequest{
		queryType: "search",
		req:       req,
		queries:   jsonQueries,
	}, nil
}

func (ds *JsonDatasource) MakeHttpRequest(ctx context.Context, remoteDsReq *RemoteDatasourceRequest) ([]byte, error) {
	res, err := ctxhttp.Do(ctx, httpClient, remoteDsReq.req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid status code. status: %v", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
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

func (ds *JsonDatasource) ParseQueryResponse(queries []*simplejson.Json, body []byte) (*datasource.DatasourceResponse, error) {
	response := &datasource.DatasourceResponse{}
	responseBody := []TargetResponseDTO{}
	err := json.Unmarshal(body, &responseBody)
	if err != nil {
		return nil, err
	}

	for i, r := range responseBody {
		refId := r.Target

		if len(queries) > i {
			refId = queries[i].Get("refId").MustString()
		}

		qr := datasource.QueryResult{
			RefId:  refId,
			Series: make([]*datasource.TimeSeries, 0),
			Tables: make([]*datasource.Table, 0),
		}

		if len(r.Columns) > 0 {
			table := datasource.Table{
				Columns: make([]*datasource.TableColumn, 0),
				Rows:    make([]*datasource.TableRow, 0),
			}

			for _, c := range r.Columns {
				table.Columns = append(table.Columns, &datasource.TableColumn{
					Name: c.Text,
				})
			}

			for _, row := range r.Rows {
				values := make([]*datasource.RowValue, 0)

				for i, cell := range row {
					rv := datasource.RowValue{}

					switch r.Columns[i].Type {
					case "time":
						if timeValue, ok := cell.(float64); ok {
							rv.Int64Value = int64(timeValue)
						}
						rv.Kind = datasource.RowValue_TYPE_INT64
					case "number":
						if numberValue, ok := cell.(float64); ok {
							rv.Int64Value = int64(numberValue)
						}
						rv.Kind = datasource.RowValue_TYPE_INT64
					case "string":
						if stringValue, ok := cell.(string); ok {
							rv.StringValue = stringValue
						}
						rv.Kind = datasource.RowValue_TYPE_STRING
					default:
						ds.logger.Debug(fmt.Sprintf("failed to parse value %v of type %T", cell, cell))
					}

					values = append(values, &rv)
				}

				table.Rows = append(table.Rows, &datasource.TableRow{Values: values})
			}

			qr.Tables = append(qr.Tables, &table)
		} else {
			serie := &datasource.TimeSeries{Name: r.Target}

			for _, p := range r.DataPoints {
				serie.Points = append(serie.Points, &datasource.Point{
					Timestamp: int64(p[1]),
					Value:     p[0],
				})
			}

			qr.Series = append(qr.Series, serie)
		}

		response.Results = append(response.Results, &qr)
	}

	return response, nil
}

func (ds *JsonDatasource) ParseSearchResponse(body []byte) (*datasource.DatasourceResponse, error) {
	jBody, err := simplejson.NewJson(body)
	if err != nil {
		return nil, err
	}

	metricCount := len(jBody.MustArray())
	table := datasource.Table{
		Columns: []*datasource.TableColumn{
			&datasource.TableColumn{Name: "text"},
		},
		Rows: make([]*datasource.TableRow, 0),
	}

	for n := 0; n < metricCount; n++ {
		values := make([]*datasource.RowValue, 0)
		jm := jBody.GetIndex(n)

		if text, found := jm.CheckGet("text"); found {
			values = append(values, &datasource.RowValue{
				Kind:        datasource.RowValue_TYPE_STRING,
				StringValue: text.MustString(),
			})
			values = append(values, &datasource.RowValue{
				Kind:       datasource.RowValue_TYPE_INT64,
				Int64Value: jm.Get("value").MustInt64(),
			})

			if len(table.Columns) == 1 {
				table.Columns = append(table.Columns, &datasource.TableColumn{Name: "value"})
			}
		} else {
			values = append(values, &datasource.RowValue{
				Kind:        datasource.RowValue_TYPE_STRING,
				StringValue: jm.MustString(),
			})
		}

		table.Rows = append(table.Rows, &datasource.TableRow{Values: values})
	}

	return &datasource.DatasourceResponse{
		Results: []*datasource.QueryResult{
			&datasource.QueryResult{
				RefId:  "search",
				Tables: []*datasource.Table{&table},
			},
		},
	}, nil
}

var httpClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation: tls.RenegotiateFreelyAsClient,
		},
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
	},
	Timeout: time.Duration(time.Second * 30),
}
