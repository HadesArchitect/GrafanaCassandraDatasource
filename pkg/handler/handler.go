package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/resource/httpadapter"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type ds interface {
	ExecQuery(ctx context.Context, q *plugin.Query) (data.Frames, error)
	GetKeyspaces(ctx context.Context) ([]string, error)
	GetTables(keyspace string) ([]string, error)
	GetColumns(keyspace, table, needType string) ([]string, error)
	GetVariables(ctx context.Context, query string) ([]plugin.Variable, error)
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
	mux.HandleFunc("/variables", h.getVariables)

	// QueryDataHandler
	queryTypeMux := datasource.NewQueryTypeMux()
	queryTypeMux.HandleFunc("query", h.queryMetricData)
	queryTypeMux.HandleFunc("alert", h.queryMetricData)

	return datasource.ServeOpts{
		CheckHealthHandler:  h,
		CallResourceHandler: httpadapter.New(mux),
		QueryDataHandler:    queryTypeMux,
	}
}

// queryMetricData is a handle to process metric requests.
func (h *handler) queryMetricData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	p, err := h.getPluginInstance(ctx, req.PluginContext)
	if err != nil {
		backend.Logger.Error("Failed to get plugin instance", "Message", err)
		return nil, fmt.Errorf("failed to get plugin instance: %w", err)
	}

	responses := backend.Responses{}
	for _, q := range req.Queries {
		backend.Logger.Debug("Process metrics request", "Request", q.JSON)
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
	p, err := h.getPluginInstance(req.Context(), pluginCtx)
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
	p, err := h.getPluginInstance(req.Context(), pluginCtx)
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
	p, err := h.getPluginInstance(req.Context(), pluginCtx)
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

// getVariables is a handle to fetch variable values.
func (h *handler) getVariables(rw http.ResponseWriter, req *http.Request) {
	backend.Logger.Debug("Process 'variables' request")

	pluginCtx := httpadapter.PluginConfigFromContext(req.Context())
	p, err := h.getPluginInstance(req.Context(), pluginCtx)
	if err != nil {
		backend.Logger.Error("Failed to get plugin instance", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	query := req.URL.Query().Get("query")

	variables, err := p.GetVariables(req.Context(), query)
	if err != nil {
		backend.Logger.Error("Failed to get variables", "Message", err)
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}

	writeHTTPResult(rw, variables)
}

// getPluginInstance fetches plugin instance from instance manager, then
// returns it if it has been successfully asserted that it is a plugin type.
func (h *handler) getPluginInstance(ctx context.Context, pluginCtx backend.PluginContext) (ds, error) {
	instance, err := h.instanceManager.Get(ctx, pluginCtx)
	if err != nil {
		return nil, fmt.Errorf("instanceManager.Get: %w", err)
	}

	p, ok := instance.(ds)
	if !ok {
		return nil, fmt.Errorf("invalid instance type: %T", instance)
	}

	return p, nil
}

// CheckHealth is a handle to check database connection status.
func (h *handler) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	p, err := h.getPluginInstance(ctx, req.PluginContext)
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

// writeHTTPResult is a simple helper to serialize data and put it in a http response.
func writeHTTPResult(rw http.ResponseWriter, val any) {
	jsonBytes, err := json.MarshalIndent(val, "", "    ")
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
