package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	"local_package/cassandra"
	"local_package/handler"
	"local_package/plugin"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

func newDataSource(settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	var dss dataSourceSettings
	err := json.Unmarshal(settings.JSONData, &dss)
	if err != nil {
		backend.Logger.Error("Failed to parse connection parameter", "Message", err)
		return nil, fmt.Errorf("failed to parse connection parameters: %w", err)
	}

	var tlsConfig *tls.Config
	if dss.UseCustomTLS {
		backend.Logger.Debug("Setting TLS Configuration...")

		tlsConfig, err = prepareTLSCfg(dss.CertPath, dss.RootPath, dss.CaPath, dss.AllowInsecureTLS)
		if err != nil {
			backend.Logger.Error("Failed to create TLS config", "Message", err)
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
	}

	sessionSettings := cassandra.Settings{
		Hosts:       strings.Split(settings.URL, ";"),
		Keyspace:    dss.Keyspace,
		User:        dss.User,
		Password:    settings.DecryptedSecureJSONData["password"],
		Consistency: dss.Consistency,
		Timeout:     dss.Timeout,
		TLSConfig:   tlsConfig,
	}

	session, err := cassandra.New(sessionSettings)
	if err != nil {
		backend.Logger.Error("Failed to create Cassandra connection", "Message", err)
		return nil, fmt.Errorf("failed to create Cassandra connection, check Grafana logs for more details")
	}

	return plugin.New(session), nil
}

func main() {
	backend.Logger.Debug("Running Cassandra backend datasource...")

	err := datasource.Serve(handler.New(newDataSource))
	if err != nil {
		backend.Logger.Error("Error serving cassandra requests: ", err.Error())
	}
}
