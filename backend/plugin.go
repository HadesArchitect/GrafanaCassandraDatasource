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
	im instancemgmt.InstanceManager
}

func newDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	logger.Debug(fmt.Sprintf("Created datasource, ID: %d\n", settings.ID))

	return CassandraDatasource{
		logger:   logger,
		settings: settings,
	}, nil
}

func (handler QueryHandler) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	instance, err := handler.im.Get(req.PluginContext)

	logger.Debug(fmt.Sprintf("Handle request: %+v\n", req))

	if err != nil {
		return nil, errors.New(fmt.Sprintf("Can not found datasource instance with ID: %d\n"))
	}

	datasource, ok := instance.(CassandraDatasource)

	if !ok {
		return nil, errors.New("Can not convert datasource instance to Cassandra Datasource")
	}

	return datasource.QueryData(ctx, req)
}

func main() {
	logger.Debug("Running Cassandra backend datasource...")

	datasource.Serve(datasource.ServeOpts{
		QueryDataHandler: &QueryHandler{
			im: datasource.NewInstanceManager(newDataSource),
		},
	})
}
