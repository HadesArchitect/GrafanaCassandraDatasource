package main

import (
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	gflog "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	//"github.com/hashicorp/go-plugin"
	//"github.com/grafana/grafana-plugin-sdk-go/backend"
)

var logger = gflog.New()

type Handler struct {
	// The instance manager can help with lifecycle management
	// of datasource instances in plugins. It's not a requirement
	// but a best practice that we recommend that you follow.
	im instancemgmt.InstanceManager
}

/*func newInstanceOpts(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {

}*/

func main() {
	logger.Debug("Running Cassandra backend datasource...")

	//im := datasource.NewInstanceManager()
	/*ds := Handler{
		im: im,
	}*/

	datasource.Serve(datasource.ServeOpts{
		QueryDataHandler: &CassandraDatasource{
			logger: logger,
		},
	})

}
