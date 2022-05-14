package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"strconv"
	"time"

	"github.com/gocql/gocql"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"golang.org/x/net/context"
)

type CassandraDatasource struct {
	logger          log.Logger
	resourceHandler backend.CallResourceHandler
	builder         *QueryBuilder
	processor       *QueryProcessor
	cluster         *gocql.ClusterConfig
}

func NewDataSource() *CassandraDatasource {
	ds := &CassandraDatasource{
		logger: logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/keyspaces", ds.handleKeyspaces)
	mux.HandleFunc("/tables", ds.handleTables)
	mux.HandleFunc("/columns", ds.handleColumns)

	ds.resourceHandler = httpadapter.New(mux)

	return ds
}

func (ds *CassandraDatasource) handleKeyspaces(rw http.ResponseWriter, req *http.Request) {
	ds.logger.Debug("Process 'keyspaces' request")

	ctx := httpadapter.PluginConfigFromContext(req.Context())

	keyspaces, err := ds.getKeyspaces(&ctx)
	if err != nil {
		ds.logger.Error("Failed to get keyspaces list", "Message", err)

		rw.WriteHeader(http.StatusInternalServerError)

		return
	}
	ds.writeHTTPResult(rw, keyspaces)
}

func (ds *CassandraDatasource) handleTables(rw http.ResponseWriter, req *http.Request) {
	ds.logger.Debug("Process 'tables' request")

	ctx := httpadapter.PluginConfigFromContext(req.Context())

	tables, err := ds.getTables(&ctx, req.URL.Query().Get("keyspace"))
	if err != nil {
		ds.logger.Error("Failed to get tables list", "Message", err)

		rw.WriteHeader(http.StatusInternalServerError)

		return
	}

	ds.writeHTTPResult(rw, tables)
}

func (ds *CassandraDatasource) handleColumns(rw http.ResponseWriter, req *http.Request) {
	ds.logger.Info("Process 'columns' request")

	ctx := httpadapter.PluginConfigFromContext(req.Context())

	keyspace := req.URL.Query().Get("keyspace")
	table := req.URL.Query().Get("table")
	needType := req.URL.Query().Get("needType")

	ds.logger.Debug("Params", keyspace, table, needType)

	columns, err := ds.getColumns(&ctx, keyspace, table, needType)
	if err != nil {
		ds.logger.Error("Failed to get columns list", "Message", err)

		rw.WriteHeader(http.StatusInternalServerError)

		return
	}

	ds.writeHTTPResult(rw, columns)
}

func (ds *CassandraDatasource) writeHTTPResult(rw http.ResponseWriter, object interface{}) {
	data, err := json.MarshalIndent(object, "", "    ")
	if err != nil {
		ds.logger.Error("Failed to unmarshal object", "Message", err, "Object", object)

		rw.WriteHeader(http.StatusInternalServerError)

		return
	}

	_, err = rw.Write(data)
	if err != nil {
		ds.logger.Error("Failed to get keyspaces list", "Message", err)

		rw.WriteHeader(http.StatusInternalServerError)

		return
	}

	rw.WriteHeader(http.StatusOK)
}

func (ds *CassandraDatasource) HandleMetricQueries(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	return &backend.QueryDataResponse{
		Responses: processQueries(req, ds.handleMetricQuery),
	}, nil
}

func (ds *CassandraDatasource) handleMetricQuery(ctx *backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	return dataResponse(ds.metricQuery(&query))
}

func (ds *CassandraDatasource) getColumns(ctx *backend.PluginContext, keyspace, table, needType string) ([]string, error) {
	errMessage := errors.New("failed to retrieve columns, please inspect Grafana server log for details")

	if keyspace == "" {
		ds.logger.Error("Keyspace is not set")

		return nil, errMessage
	}

	if table == "" {
		ds.logger.Error("Table is not set")

		return nil, errMessage
	}

	session, err := ds.Session()
	if err != nil {
		ds.logger.Error("Failed to create session", "Message", err)

		return nil, errMessage
	}

	keyspaceMetadata, err := session.KeyspaceMetadata(keyspace)
	if err != nil {
		ds.logger.Error("Failed to retrieve keyspace metadata", "Message", err, "Keyspace", keyspace)

		return nil, errMessage
	}

	tableMetadata, ok := keyspaceMetadata.Tables[table]
	if !ok {
		ds.logger.Error("Table not fount", "Table", table)

		return nil, errMessage
	}

	var columns []string = make([]string, 0)
	for name, column := range tableMetadata.Columns {
		if column.Type.Type().String() == needType {
			columns = append(columns, name)
		}
	}

	return columns, nil
}

func (ds *CassandraDatasource) getTables(ctx *backend.PluginContext, keyspace string) ([]string, error) {
	errMessage := errors.New("failed to retrieve tables, please inspect Grafana server log for details")

	if keyspace == "" {
		ds.logger.Error("Keyspace is not set")

		return nil, errMessage
	}

	session, err := ds.Session()
	if err != nil {
		ds.logger.Error("Failed to create session", "Message", err)

		return nil, errMessage
	}

	keyspaceMetadata, err := session.KeyspaceMetadata(keyspace)
	if err != nil {
		ds.logger.Error("Failed to retrieve keyspace metadata", "Message", err, "Keyspace", keyspace)

		return nil, errMessage
	}

	var tables []string = make([]string, 0)
	for name := range keyspaceMetadata.Tables {
		tables = append(tables, name)
	}

	return tables, nil
}

func (ds *CassandraDatasource) getKeyspaces(ctx *backend.PluginContext) ([]string, error) {
	keyspaces, err := ds.processor.processKeyspacesQuery(ds)
	if err != nil {
		ds.logger.Error("Failed to get keyspaces list", "Message", err)

		return nil, errors.New("failed to get keyspaces list, please inspect Grafana server log for details")
	}

	return keyspaces, nil
}

func (ds *CassandraDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	if req.PluginContext.DataSourceInstanceSettings.URL == "" {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Host field cannot be empty, please fill it with a proper value",
		}, nil
	}

	err := ds.connect(&req.PluginContext)
	if err != nil {
		ds.logger.Warn("Failed to connect", "Message", err)

		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Failed to connect to server, please inspect Grafana server log for details",
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Database connected",
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
		ds.logger.Error("Failed to parse queries", "Message", err)

		return nil, errors.New("failed to parse queries, please inspect Grafana server log for details")
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
		ds.logger.Error("Failed to parse request", "Message", err)

		return options, errors.New("failed to parse request, please inspect Grafana server log for details")
	}
	return options, nil
}

func (ds *CassandraDatasource) connect(context *backend.PluginContext) error {
	hosts := strings.Split(context.DataSourceInstanceSettings.URL, ";")

	err := ds.tryToConnect(hosts, context)
	if err != nil {
		ds.logger.Error("Failed to connect", "Message", err)

		return errors.New("failed to connect, please inspect Grafana server log for details")
	}

	return nil
}

func (ds *CassandraDatasource) tryToConnect(hosts []string, context *backend.PluginContext) error {
	options, err := ds.getRequestOptions(context.DataSourceInstanceSettings.JSONData)
	if err != nil {
		ds.logger.Error("Failed to parse connection parameters", "Message", err)

		return errors.New("failed to parse connection parameters, please inspect Grafana server log for details")
	}

	ds.logger.Debug("Connecting", "Hosts", hosts)

	cluster := gocql.NewCluster(hosts...)

	if options.Timeout != nil {
		cluster.Timeout = time.Duration(*options.Timeout) * time.Second
	}

	consistency, err := parseConsistency(options.Consistency)
	if err != nil {
		ds.logger.Error("Failed to parse consistency", "Message", err, "Consistency", options.Consistency)

		return errors.New("failed to parse consistency, please inspect Grafana server log for details")
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
			ds.logger.Error("Failed to create TLS config", "Message", err)

			return errors.New("failed to create TLS config, please inspect Grafana server log for details")
		}

		cluster.SslOpts = &gocql.SslOptions{Config: tlsConfig}
	}

	_, err = cluster.CreateSession()
	if err != nil {
		ds.logger.Warn("Failed to create session", "Message", err)

		return errors.New("failed to create Cassandra session, please inspect Grafana server log for details")
	}

	ds.cluster = cluster

	return nil
}

func (ds *CassandraDatasource) Session() (*gocql.Session, error) {
	session, err := ds.cluster.CreateSession()
	if err != nil {
		ds.logger.Warn("Failed to create session", "Message", err)

		return nil, errors.New("failed to create Cassandra session, please inspect Grafana server log for details")
	}

	return session, nil
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
