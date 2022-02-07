package main

import (
	// "crypto/tls"
	"encoding/json"
	"fmt"

	"strconv"
	"time"

	"github.com/gocql/gocql"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	gflog "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/net/context"
)

type CassandraDatasource struct {
	logger    gflog.Logger
	builder   *QueryBuilder
	processor *QueryProcessor
	sessions  map[int]*gocql.Session
	session   *gocql.Session
	settings  backend.DataSourceInstanceSettings
}

func (ds *CassandraDatasource) handleMetricQuery(ctx context.Context, request *backend.QueryDataRequest, query backend.DataQuery) (data.Frames, error) {
	_, err := ds.Connect(&request.PluginContext)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with the database, err=%v", err)
	}

	return ds.MetricQuery(&query)
}

func (ds *CassandraDatasource) handleConnectionQuery(ctx context.Context, request *backend.QueryDataRequest, query backend.DataQuery) (data.Frames, error) {
	_, err := ds.Connect(&request.PluginContext)
	if err != nil {
		return nil, fmt.Errorf("unable to establish connection with the database, err=%v", err)
	}

	return data.Frames{}, nil
}

func (ds *CassandraDatasource) finalMetricHandler(ctx context.Context, req *backend.QueryDataRequest, query backend.DataQuery) backend.DataResponse {
	return frameResponseHandler(ds.handleMetricQuery(ctx, req, query))
}

func (ds *CassandraDatasource) finalConnectionHandler(ctx context.Context, req *backend.QueryDataRequest, query backend.DataQuery) backend.DataResponse {
	return frameResponseHandler(ds.handleConnectionQuery(ctx, req, query))
}

func frameResponseHandler(frames data.Frames, err error) backend.DataResponse {
	if err != nil {
		return backend.DataResponse{
			Error: err,
		}
	}

	return backend.DataResponse{
		Frames: frames,
	}
}

func (ds *CassandraDatasource) HandleMetricQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(ctx, req, ds.finalMetricHandler),
	}, nil
}

func (ds *CassandraDatasource) HandleConnectionQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(ctx, req, ds.finalConnectionHandler),
	}, nil
}

func processQueries(ctx context.Context, req *backend.QueryDataRequest, handler func(context.Context, *backend.QueryDataRequest, backend.DataQuery) backend.DataResponse) backend.Responses {
	res := backend.Responses{}
	for _, v := range req.Queries {
		res[v.RefID] = handler(ctx, req, v)
	}

	return res
}

func (ds *CassandraDatasource) MetricQuery(query *backend.DataQuery) (data.Frames, error) {
	var cassQuery CassandraQuery
	err := json.Unmarshal(query.JSON, &cassQuery)
	if err != nil {
		return nil, fmt.Errorf("unmarshal queries, err=%v", err)
	}

	from, to := timeRangeToStr(query.TimeRange)
	ds.logger.Debug(fmt.Sprintf("Timeframe from: %s to %s\n", from, to))

	var preparedQuery string
	if cassQuery.RawQuery {
		preparedQuery = ds.builder.prepareRawMetricQuery(&cassQuery, from, to)
	} else {
		preparedQuery = ds.builder.prepareStrictMetricQuery(&cassQuery, from, to)
	}

	if cassQuery.RawQuery {
		return ds.processor.processRawMetricQuery(preparedQuery, ds)
	} else {
		return ds.processor.processStrictMetricQuery(preparedQuery, cassQuery.ValueID, ds)
	}
}

func timeRangeToStr(timeRange backend.TimeRange) (string, string) {
	from := strconv.FormatInt(timeRange.From.UnixNano()/int64(time.Millisecond), 10)
	to := strconv.FormatInt(timeRange.To.UnixNano()/int64(time.Millisecond), 10)

	return from, to
}

func (ds *CassandraDatasource) getRequestOptions(jsonData []byte) (DataSourceOptions, error) {
	var options DataSourceOptions
	err := json.Unmarshal(jsonData, &options)
	if err != nil {
		return options, fmt.Errorf("get request JSON data, err=%v", err)
	}
	return options, nil
}

func (ds *CassandraDatasource) Connect(context *backend.PluginContext) (bool, error) {
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

		tlsConfig, err := PrepareTLSCfg(options.CertPath, options.RootPath, options.CaPath)
		if err != nil {
			return false, fmt.Errorf("create TLS config, err=%v", err)
		}

		cluster.SslOpts = &gocql.SslOptions{Config: tlsConfig}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return false, fmt.Errorf("create Cassandra BD session, err=%v", err)
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
