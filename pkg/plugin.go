package main

import (
	"github.com/grafana/grafana_plugin_model/go/datasource"
	hclog "github.com/hashicorp/go-hclog"
	plugin "github.com/hashicorp/go-plugin"
)

var logger = hclog.New(&hclog.LoggerOptions{
	Name:  "cassandra-backend-datasource",
	Level: hclog.LevelFromString("DEBUG"),
})

func main() {
	logger.Debug("Running Cassandra backend datasource...")

	plugin.Serve(&plugin.ServeConfig{

		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "grafana_plugin_type",
			MagicCookieValue: "datasource",
		},
		Plugins: map[string]plugin.Plugin{
			"cassandra-backend-datasource": &datasource.DatasourcePluginImpl{Plugin: &JsonDatasource{
				logger: logger,
			}},
		},

		GRPCServer: plugin.DefaultGRPCServer,
	})
}