package main

import (
	"context"
	"fmt"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// secureSocksDialer returns a dialer routing cluster connections through Grafana's
// secure SOCKS proxy, the transport behind Private Data Source Connect (PDC).
//
// It returns a nil dialer when the proxy is not requested by the datasource or
// not configured on the Grafana instance, in which case connections are made
// directly from the plugin.
func secureSocksDialer(ctx context.Context, settings backend.DataSourceInstanceSettings) (cassandra.Dialer, error) {
	proxyClient, err := settings.ProxyClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("settings.ProxyClient: %w", err)
	}

	if !proxyClient.SecureSocksProxyEnabled() {
		return nil, nil
	}

	dialer, err := proxyClient.NewSecureSocksProxyContextDialer()
	if err != nil {
		return nil, fmt.Errorf("proxyClient.NewSecureSocksProxyContextDialer: %w", err)
	}

	contextDialer, ok := dialer.(cassandra.Dialer)
	if !ok {
		return nil, fmt.Errorf("secure socks proxy dialer is not context aware: %T", dialer)
	}

	return contextDialer, nil
}
