package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"local_package/cassandra"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type plugin interface {
	ExecQuery(ctx context.Context, q *cassandra.Query) (data.Frames, error)
	GetKeyspaces(ctx context.Context) ([]string, error)
	GetTables(keyspace string) ([]string, error)
	GetColumns(keyspace, table, needType string) ([]string, error)
	CheckHealth(ctx context.Context) error
	Dispose()
}

// handler controls plugin instance manager and handles requests.
type handler struct {
	instanceManager instancemgmt.InstanceManager
}

// New initializes plugin instance, http multiplexers, and sets http request handlers.
func New(fn datasource.InstanceFactoryFunc) datasource.ServeOpts {
	h := &handler{instanceManager: datasource.NewInstanceManager(fn)}

	// CallResourceHandler
	mux := http.NewServeMux()
	mux.HandleFunc("/keyspaces", h.getKeyspaces)
	mux.HandleFunc("/tables", h.getTables)
	mux.HandleFunc("/columns", h.getColumns)

	// QueryDataHandler
	queryTypeMux := datasource.NewQueryTypeMux()
	queryTypeMux.HandleFunc("query", h.queryMetricData)

	return datasource.ServeOpts{
		CheckHealthHandler:  h,
		CallResourceHandler: httpadapter.New(mux),
		QueryDataHandler:    queryTypeMux,
	}
}

// queryMetricData is a handle to process metric requests.
func (h *handler) queryMetricData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	p, err := h.getPluginInstance(req.PluginContext)
	if err != nil {
		backend.Logger.Error("Failed to get plugin instance", "Message", err)
		return nil, fmt.Errorf("failed to get plugin instance: %w", err)
	}

	responses := backend.Responses{}
	for _, q := range req.Queries {
		cassQuery, err := parseDataQuery(&q)
		if err != nil {
			backend.Logger.Error("Failed to parse query", "Message", err)
			responses[q.RefID] = backend.DataResponse{Error: fmt.Errorf("json.Unmarshal: %w", err)}
			continue
		}

		dataFrames, err := p.ExecQuery(ctx, cassQuery)
		if err != nil {
			backend.Logger.Error("Failed to execute query", "Message", err)
			responses[q.RefID] = backend.DataResponse{Error: fmt.Errorf("p.ExecQuery: %w", err)}
			continue
		}

		responses[q.RefID] = backend.DataResponse{Frames: dataFrames}
	}

	return &backend.QueryDataResponse{Responses: responses}, nil
}

// getKeyspaces is a handle to fetch keyspaces list.
func (h *handler) getKeyspaces(rw http.ResponseWriter, req *http.Request) {
	backend.Logger.Debug("Process 'keyspaces' request")

	pluginCtx := httpadapter.PluginConfigFromContext(req.Context())
	p, err := h.getPluginInstance(pluginCtx)
	if err != nil {
		backend.Logger.Error("Failed to get plugin instance", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	keyspaces, err := p.GetKeyspaces(req.Context())
	if err != nil {
		backend.Logger.Error("Failed to get keyspaces list", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeHTTPResult(rw, keyspaces)
}

// getTables is a handle to fetch tables list.
func (h *handler) getTables(rw http.ResponseWriter, req *http.Request) {
	backend.Logger.Debug("Process 'tables' request")

	pluginCtx := httpadapter.PluginConfigFromContext(req.Context())
	p, err := h.getPluginInstance(pluginCtx)
	if err != nil {
		backend.Logger.Error("Failed to get plugin instance", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	keyspace := req.URL.Query().Get("keyspace")

	tables, err := p.GetTables(keyspace)
	if err != nil {
		backend.Logger.Error("Failed to get tables list", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeHTTPResult(rw, tables)
}

// getColumns is a handle to fetch columns list.
func (h *handler) getColumns(rw http.ResponseWriter, req *http.Request) {
	backend.Logger.Debug("Process 'columns' request")

	pluginCtx := httpadapter.PluginConfigFromContext(req.Context())
	p, err := h.getPluginInstance(pluginCtx)
	if err != nil {
		backend.Logger.Error("Failed to get plugin instance", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	keyspace := req.URL.Query().Get("keyspace")
	table := req.URL.Query().Get("table")
	needType := req.URL.Query().Get("needType")

	columns, err := p.GetColumns(keyspace, table, needType)
	if err != nil {
		backend.Logger.Error("Failed to get columns list", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeHTTPResult(rw, columns)
}

// getPluginInstance fetches plugin instance from instance manager, then
// returns it if it has been successfully asserted that it is a plugin type.
func (h *handler) getPluginInstance(pluginCtx backend.PluginContext) (plugin, error) {
	instance, err := h.instanceManager.Get(pluginCtx)
	if err != nil {
		return nil, fmt.Errorf("instanceManager.Get: %w", err)
	}

	p, ok := instance.(plugin)
	if !ok {
		return nil, fmt.Errorf("invalid instance type: %T", instance)
	}

	return p, nil
}

// CheckHealth is a handle to check database connection status.
func (h *handler) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	p, err := h.getPluginInstance(req.PluginContext)
	if err != nil {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusUnknown,
			Message: "Error, check Grafana logs for more details",
		}, nil
	}

	err = p.CheckHealth(ctx)
	if err != nil {
		backend.Logger.Error("Failed to connect to server", "Message", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Error, check Grafana logs for more details",
		}, nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: "Connected",
	}, nil
}

// writeHTTPResult is a simple helper to serialize the
// list of strings and put it in a http response.
func writeHTTPResult(rw http.ResponseWriter, list []string) {
	jsonBytes, err := json.MarshalIndent(list, "", "    ")
	if err != nil {
		backend.Logger.Error("Failed to marshal list", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = rw.Write(jsonBytes)
	if err != nil {
		backend.Logger.Error("Failed to write response", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	rw.WriteHeader(http.StatusOK)
}
