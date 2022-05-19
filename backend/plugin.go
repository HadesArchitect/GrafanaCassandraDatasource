package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var logger = log.New()

type Handler struct {
	im  instancemgmt.InstanceManager
	mux *datasource.QueryTypeMux
}

func newDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger.Debug("Created datasource", "id", settings.ID)

	return NewDataSource(settings)
}

func (handler *Handler) getDataSource(ctx *backend.PluginContext) (*CassandraDatasource, error) {
	instance, err := handler.im.Get(*ctx)
	if err != nil {
		return nil, fmt.Errorf("can not found datasource instance with ID: %d", ctx.DataSourceInstanceSettings.ID)
	}

	datasource, ok := instance.(*CassandraDatasource)
	if !ok {
		return nil, errors.New("can not convert datasource instance to Cassandra Datasource")
	}

	return datasource, nil
}

func (handler *Handler) CallResource(ctx context.Context, req *backend.CallResourceRequest, sender backend.CallResourceResponseSender) error {
	datasource, err := handler.getDataSource(&req.PluginContext)
	if err != nil {
		logger.Error("Failed to get datasource", "error", err)
		return fmt.Errorf("error, check Grafana logs for more details")
	}

	return datasource.resourceHandler.CallResource(ctx, req, sender)
}

func (handler *Handler) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	datasource, err := handler.getDataSource(&req.PluginContext)
	if err != nil {
		logger.Error("Failed to get datasource", "error", err)
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusError,
			Message: "Error, check Grafana logs for more details",
		}, nil
	}

	return datasource.CheckHealth(ctx, req)
}

func NewHandler() *Handler {
	handler := &Handler{
		im: datasource.NewInstanceManager(newDataSource),
	}

	mux := datasource.NewQueryTypeMux()
	mux.HandleFunc("query", func(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
		datasource, err := handler.getDataSource(&req.PluginContext)
		if err != nil {
			logger.Error("Failed to get datasource", "error", err)
			return nil, fmt.Errorf("error, check Grafana logs for more details")
		}

		return datasource.HandleMetricQueries(ctx, req)
	})

	handler.mux = mux

	return handler
}

func main() {
	logger.Debug("Running Cassandra backend datasource...")

	handler := NewHandler()

	datasource.Serve(datasource.ServeOpts{
		CallResourceHandler: handler,
		CheckHealthHandler:  handler,
		QueryDataHandler:    handler.mux,
	})
}
