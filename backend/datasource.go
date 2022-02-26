package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
	host      string
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
		data.NewField(query.RefID, nil, make([]string, 0)),
		data.NewField("type", nil, make([]string, 0)),
	)

	for name, column := range tableMetadata.Columns {
		frame.AppendRow(name, column.Type.Type().String())
	}

	return dataResponse([]*data.Frame{frame}, nil)
}

func (ds *CassandraDatasource) HandleKeyspacesQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(req, ds.handleKeyspacesQuery),
	}, nil
}

func (ds *CassandraDatasource) handleKeyspacesQuery(ctx *backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	_, err := ds.connect(ctx)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("unable to establish connection with the database, err=%v", err))
	}

	keyspaces, err := ds.processor.processKeyspacesQuery(ds)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("get keyspaces list, err=%v", err))
	}

	return dataResponse(keyspaces, nil)
}

func (ds *CassandraDatasource) HandleTablesQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(req, ds.handleTablesQuery),
	}, nil
}

func (ds *CassandraDatasource) handleTablesQuery(ctx *backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	_, err := ds.connect(ctx)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("unable to establish connection with the database, err=%v", err))
	}

	var cassQuery CassandraQuery
	err = json.Unmarshal(query.JSON, &cassQuery)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("unmarshal queries, err=%v", err))
	}

	if cassQuery.Keyspace == "" {
		return dataResponse(data.Frames{}, errors.New("keyspace is not set"))
	}

	keyspaceMetadata, err := ds.session.KeyspaceMetadata(cassQuery.Keyspace)
	if err != nil {
		return dataResponse(data.Frames{}, fmt.Errorf("retrieve keyspace metadata, err=%v, keyspace=%s", err, cassQuery.Keyspace))
	}

	frame := data.NewFrame(
		query.RefID,
		data.NewField(query.RefID, nil, make([]string, 0)),
	)

	for name := range keyspaceMetadata.Tables {
		frame.AppendRow(name)
	}

	return dataResponse([]*data.Frame{frame}, nil)
}

func (ds *CassandraDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	_, err := ds.connect(&req.PluginContext)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: fmt.Sprintf("Connection test failed, error = unable to establish connection with the database, err=%v", err),
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: fmt.Sprintf("Database connected, host: %s", ds.host),
	}, nil
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

		return ds.processor.processStrictMetricQuery(preparedQuery, cassQuery.ValueID, cassQuery.Alias, ds)
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
	if ds.session != nil {
		return true, nil
	}

	hosts := strings.Split(context.DataSourceInstanceSettings.URL, ";")

	for _, host := range hosts {
		err := ds.tryToConnect(host, context)
		if err == nil {
			ds.logger.Debug(fmt.Sprintf("Connected to host %s", host))

			return true, nil
		} else if err != nil {
			ds.logger.Warn(fmt.Sprintf("Failed to connect to host, host: %s, error: %+v", host, err))
		}
	}

	return false, fmt.Errorf("connect to hosts, hosts=%+v", hosts)
}

func (ds *CassandraDatasource) tryToConnect(host string, context *backend.PluginContext) error {
	options, err := ds.getRequestOptions(context.DataSourceInstanceSettings.JSONData)
	if err != nil {
		return fmt.Errorf("parse connection parameters, err=%v", err)
	}

	ds.logger.Debug(fmt.Sprintf("Connecting to %s...\n", host))

	cluster := gocql.NewCluster(host)

	if options.Timeout != nil {
		cluster.Timeout = time.Duration(*options.Timeout) * time.Second
	}

	consistency, err := parseConsistency(options.Consistency)
	if err != nil {
		return fmt.Errorf("parse Consistency, err=%v, consistency string=%s", err, options.Consistency)
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
			return fmt.Errorf("create TLS config, err=%v", err)
		}

		cluster.SslOpts = &gocql.SslOptions{Config: tlsConfig}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		return fmt.Errorf("create Cassandra DB session, err=%v", err)
	}

	ds.host = host
	ds.session = session

	return nil
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
