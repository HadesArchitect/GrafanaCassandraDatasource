package main

import (
	"github.com/grafana/grafana-plugin-model/go/datasource"
	gflog "github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/hashicorp/go-plugin"
)

var logger = gflog.New()

func main() {
	logger.Debug("Running Cassandra backend datasource...")

	plugin.Serve(&plugin.ServeConfig{

		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			"cassandra-backend-datasource": &datasource.DatasourcePluginImpl{Plugin: &CassandraDatasource{
				logger: logger,
			}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
