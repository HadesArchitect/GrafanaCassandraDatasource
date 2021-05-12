package main

import (
	"context"
	"errors"

	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	gflog "github.com/grafana/grafana-plugin-sdk-go/backend/log"

	//"github.com/hashicorp/go-plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var logger = gflog.New()

type DataHandler struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirement
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

func (Handler DataHandler) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	datasource := mainInstance.datasources[req.PluginContext.DataSourceInstanceSettings.ID]
	if datasource == nil {
		return nil, errors.New("No datasource with such ID")
	}

	return datasource.QueryData(ctx, req), nil
}

/*func newInstanceOpts(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

}*/

var mainInstance Plugin

func newDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	if mainInstance.datasources == nil {
		mainInstance.datasources = make(map[int64]*CassandraDatasource)
	}

	datasource := CassandraDatasource{
		ID:     settings.ID,
		logger: logger,
	}

	mainInstance.datasources[settings.ID] = &datasource

	return datasource, nil
}

func main() {
	logger.Debug("Running Cassandra backend datasource...")

	//im := datasource.NewInstanceManager()
	/*ds := Handler{
		im: im,
	}*/
	datasource.Serve(datasource.ServeOpts{
		QueryDataHandler: &DataHandler{
			im: datasource.NewInstanceManager(newDataSource),
		},
	})

}

type Plugin struct {
	datasources map[int64]*CassandraDatasource
}
