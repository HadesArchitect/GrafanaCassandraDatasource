package main

import (
	"context"
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/proxy"
	"github.com/stretchr/testify/require"
)

func Test_secureSocksDialer(t *testing.T) {
	proxyConfigured := map[string]string{
		proxy.PluginSecureSocksProxyEnabled:       "true",
		proxy.PluginSecureSocksProxyProxyAddress:  "localhost:5555",
		proxy.PluginSecureSocksProxyAllowInsecure: "true",
	}

	testCases := []struct {
		name       string
		grafanaCfg map[string]string
		jsonData   string
		wantDialer bool
		wantErr    bool
	}{
		{
			name:       "proxy requested and configured",
			grafanaCfg: proxyConfigured,
			jsonData:   `{"enableSecureSocksProxy":true}`,
			wantDialer: true,
		},
		{
			name:       "proxy not requested by the datasource",
			grafanaCfg: proxyConfigured,
			jsonData:   `{"enableSecureSocksProxy":false}`,
			wantDialer: false,
		},
		{
			name:       "proxy key absent from the datasource settings",
			grafanaCfg: proxyConfigured,
			jsonData:   `{"keyspace":"system"}`,
			wantDialer: false,
		},
		{
			name:       "proxy requested but not configured on the grafana instance",
			grafanaCfg: nil,
			jsonData:   `{"enableSecureSocksProxy":true}`,
			wantDialer: false,
		},
		{
			name:       "broken datasource settings",
			grafanaCfg: proxyConfigured,
			jsonData:   `{`,
			wantErr:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := backend.WithGrafanaConfig(context.Background(), backend.NewGrafanaCfg(tc.grafanaCfg))
			settings := backend.DataSourceInstanceSettings{JSONData: []byte(tc.jsonData)}

			dialer, err := secureSocksDialer(ctx, settings)
			if tc.wantErr {
				require.Error(t, err)
				require.Nil(t, dialer)
				return
			}

			require.NoError(t, err)
			if tc.wantDialer {
				require.NotNil(t, dialer)
			} else {
				require.Nil(t, dialer)
			}
		})
	}
}
