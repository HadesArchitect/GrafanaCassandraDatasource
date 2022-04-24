package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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
	sessions        map[int]*gocql.Session
	session         *gocql.Session
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
	ds.logger.Info("Process 'keyspaces' request")

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
	ds.logger.Info("Process 'tables' request")

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
	_, err := ds.connect(ctx)
	if err != nil {
		ds.logger.Warn("Failed to connect", "Message", err)
		return dataResponse(data.Frames{}, fmt.Errorf("Failed to connect to server, please inspect Grafana server log for details"))
	}

	frames, err := ds.metricQuery(&query)

	return dataResponse(frames, err)
}

func (ds *CassandraDatasource) getConnected(ctx *backend.PluginContext, getData func() ([]string, error)) ([]string, error) {
	_, err := ds.connect(ctx)
	if err != nil {
		ds.logger.Warn("Failed to connect", "Message", err)

		return nil, errors.New("Failed to connect to server, please inspect Grafana server log for details")
	}

	return getData()
}

func (ds *CassandraDatasource) getColumns(ctx *backend.PluginContext, keyspace, table, needType string) ([]string, error) {
	getColumns := func() ([]string, error) {
		if keyspace == "" {
			return nil, errors.New("Keyspace is not set")
		}

		if table == "" {
			return nil, errors.New("Table is not set")
		}

		keyspaceMetadata, err := ds.session.KeyspaceMetadata(keyspace)
		if err != nil {
			ds.logger.Error("Failed to retrieve keyspace metadata", "Message", err, "Keyspace", keyspace)

			return nil, fmt.Errorf("Failed to retrieve keyspace metadata, please inspect Grafana server log for details")
		}

		tableMetadata, ok := keyspaceMetadata.Tables[table]
		if !ok {
			return nil, fmt.Errorf("Table '%s' not found", table)
		}

		var columns []string = make([]string, 0)
		for name, column := range tableMetadata.Columns {
			if column.Type.Type().String() == needType {
				columns = append(columns, name)
			}
		}

		return columns, nil
	}

	return ds.getConnected(ctx, getColumns)
}

func (ds *CassandraDatasource) getTables(ctx *backend.PluginContext, keyspace string) ([]string, error) {
	getTables := func() ([]string, error) {
		if keyspace == "" {
			return nil, errors.New("Keyspace is not set")
		}

		keyspaceMetadata, err := ds.session.KeyspaceMetadata(keyspace)
		if err != nil {
			ds.logger.Error("Failed to retrieve keyspace metadata", "Message", err, "Keyspace", keyspace)

			return nil, errors.New("Failed to retrieve keyspace metadata")
		}

		var tables []string = make([]string, 0)
		for name, _ := range keyspaceMetadata.Tables {
			tables = append(tables, name)
		}

		return tables, nil
	}

	return ds.getConnected(ctx, getTables)
}

func (ds *CassandraDatasource) getKeyspaces(ctx *backend.PluginContext) ([]string, error) {
	getKeyspaces := func() ([]string, error) {
		keyspaces, err := ds.processor.processKeyspacesQuery(ds)
		if err != nil {
			ds.logger.Error("Failed to get keyspaces list", "Message", err)

			return nil, errors.New("Failed to get keyspaces list")
		}

		return keyspaces, nil
	}

	return ds.getConnected(ctx, getKeyspaces)
}

func (ds *CassandraDatasource) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) error {
	_, err := ds.connect(&req.PluginContext)
	if err != nil {
		ds.logger.Warn("Failed to connect", "Message", err)
		return fmt.Errorf("Failed to connect to server, please inspect Grafana server log for details")
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
		ds.logger.Error("Failed to parse queries", "Message", err)
		return nil, fmt.Errorf("Failed to parse queries, please inspect Grafana server log for details")
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
		return options, fmt.Errorf("Failed to parse request, please inspect Grafana server log for details")
	}
	return options, nil
}

func (ds *CassandraDatasource) connect(context *backend.PluginContext) (bool, error) {
	options, err := ds.getRequestOptions(context.DataSourceInstanceSettings.JSONData)
	if err != nil {
		ds.logger.Error("Failed to parse connection parameters", "Message", err)
		return false, fmt.Errorf("Failed to parse connection parameters, please inspect Grafana server log for details")
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
		ds.logger.Error("Failed to parse consistency", "Message", err, "Consistency", options.Consistency)
		return false, fmt.Errorf("Failed to parse consistency, please inspect Grafana server log for details")
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
			return false, fmt.Errorf("Failed to create TLS config, please inspect Grafana server log for details")
		}

		cluster.SslOpts = &gocql.SslOptions{Config: tlsConfig}
	}

	session, err := cluster.CreateSession()
	if err != nil {
		ds.logger.Warn("Failed to create session", "Message", err)
		return false, fmt.Errorf("Failed to create Cassandra session, please inspect Grafana server log for details")
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
