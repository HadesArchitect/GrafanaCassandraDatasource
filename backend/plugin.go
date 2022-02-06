package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	gflog "github.com/grafana/grafana-plugin-sdk-go/backend/log"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var logger = gflog.New()

type QueryHandler struct {
	im  instancemgmt.InstanceManager
	mux *datasource.QueryTypeMux
}

func newDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger.Debug(fmt.Sprintf("Created datasource, ID: %d\n", settings.ID))

	return &CassandraDatasource{
		logger:   logger,
		settings: settings,
	}, nil
}

func (handler *QueryHandler) getDataSource(req *backend.QueryDataRequest) (*CassandraDatasource, error) {
	instance, err := handler.im.Get(req.PluginContext)

	logger.Debug(fmt.Sprintf("Handle request: %+v\n", req))

	if err != nil {
		return nil, fmt.Errorf("can not found datasource instance with ID: %d", req.PluginContext.DataSourceInstanceSettings.ID)
	}

	datasource, ok := instance.(*CassandraDatasource)

	if !ok {
		return nil, errors.New("can not convert datasource instance to Cassandra Datasource")
	}

	return datasource, nil
}

func (handler *QueryHandler) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	instance, err := handler.im.Get(req.PluginContext)

	logger.Debug(fmt.Sprintf("Handle request: %+v\n", req))

	if err != nil {
		return nil, fmt.Errorf("can not found datasource instance with ID: %d", req.PluginContext.DataSourceInstanceSettings.ID)
	}

	datasource, ok := instance.(*CassandraDatasource)

	if !ok {
		return nil, errors.New("can not convert datasource instance to Cassandra Datasource")
	}

	responses, err := datasource.QueryData(ctx, req)
	if err != nil {
		return nil, err
	}

	return &backend.QueryDataResponse{
		Responses: responses,
	}, nil
}

func main() {
	handler := &QueryHandler{
		im: datasource.NewInstanceManager(newDataSource),
	}

	mux := datasource.NewQueryTypeMux()
	mux.HandleFunc("query", func(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
		datasource, err := handler.getDataSource(req)
		if err != nil {
			return nil, fmt.Errorf("get datasource for query, err=%v", err)
		}

		return datasource.HandleMetricQueries(ctx, req)
	})

	handler.mux = mux

	logger.Debug("Running Cassandra backend datasource...")

	datasource.Serve(datasource.ServeOpts{
		QueryDataHandler: handler,
	})
}
