package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/handler"
	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
)

// newDataSource is the instance factory. Its errors are shown to users, in
// the health check and in panels, so the original error is only logged and
// the returned one carries the curated message of its health.Category.
func newDataSource(ctx context.Context, settings backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	instance, err := newPlugin(ctx, settings)
	if err != nil {
		backend.Logger.Error("Failed to create the data source instance", "Message", err)
		return nil, health.Public(err)
	}

	return instance, nil
}

func newPlugin(ctx context.Context, settings backend.DataSourceInstanceSettings) (*plugin.Plugin, error) {
	var dss dataSourceSettings
	err := json.Unmarshal(settings.JSONData, &dss)
	if err != nil {
		return nil, health.Wrap(health.InvalidSettings, fmt.Errorf("failed to parse connection parameters: %w", err))
	}

	var tlsConfig *tls.Config
	if dss.UseCustomTLS {
		backend.Logger.Debug("Setting TLS Configuration...")

		if dss.UseCertContent {
			// Use certificate content from secure data
			certContent := settings.DecryptedSecureJSONData["certContent"]
			rootContent := settings.DecryptedSecureJSONData["rootContent"]
			caContent := settings.DecryptedSecureJSONData["caContent"]

			tlsConfig, err = prepareTLSCfgFromContent(certContent, rootContent, caContent, dss.AllowInsecureTLS)
		} else {
			// Use certificate file paths
			tlsConfig, err = prepareTLSCfgFromPaths(dss.CertPath, dss.RootPath, dss.CaPath, dss.AllowInsecureTLS)
		}

		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
	}

	allowedAuthenticators := parseAllowedAuthenticators(dss.AllowedAuthenticators)
	if len(allowedAuthenticators) > 0 {
		backend.Logger.Debug("Using custom authenticator", "authenticators", strings.Join(allowedAuthenticators, ";"))
	}

	dialer, err := secureSocksDialer(ctx, settings)
	if err != nil {
		return nil, health.Wrap(health.SecureSocksProxy, fmt.Errorf("failed to set up the secure socks proxy: %w", err))
	}

	proxyUnavailable := dialer == nil && dss.EnableSecureSocksProxy
	switch {
	case dialer != nil:
		backend.Logger.Debug("Connecting to Cassandra through the secure socks proxy")
	case proxyUnavailable:
		backend.Logger.Warn("Secure socks proxy is enabled for this datasource but not configured on this Grafana instance, connecting to Cassandra directly")
	}

	sessionSettings := cassandra.Settings{
		Hosts:                 strings.Split(settings.URL, ";"),
		Keyspace:              dss.Keyspace,
		User:                  dss.User,
		Password:              settings.DecryptedSecureJSONData["password"],
		Consistency:           dss.Consistency,
		Timeout:               dss.Timeout,
		TLSConfig:             tlsConfig,
		AllowedAuthenticators: allowedAuthenticators,
		Dialer:                dialer,
	}

	session, err := cassandra.New(sessionSettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cassandra connection: %w", err)
	}

	return plugin.New(session, plugin.Options{
		Keyspace:                    dss.Keyspace,
		SecureSocksProxy:            dialer != nil,
		SecureSocksProxyUnavailable: proxyUnavailable,
	}), nil
}

func main() {
	backend.Logger.Debug("Running Cassandra backend datasource...")

	err := datasource.Serve(handler.New(newDataSource))
	if err != nil {
		backend.Logger.Error("Error serving cassandra requests: ", err.Error())
	}
}
