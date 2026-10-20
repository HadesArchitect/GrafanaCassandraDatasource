package main

import (
	"context"
	"errors"
	"testing"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_newDataSource_settingsErrors checks that settings which can never work
// are reported by the instance factory with their own category and a curated
// message (issue #271), instead of a generic error.
func Test_newDataSource_settingsErrors(t *testing.T) {
	// A fake, unparsable key used only to check that it never leaks.
	const (
		fakeKeyBody   = "fake-test-key-do-not-use"
		secretContent = "-----BEGIN PRIVATE KEY-----" + fakeKeyBody + "-----END PRIVATE KEY-----"
	)

	testCases := map[string]struct {
		url        string
		jsonData   string
		secure     map[string]string
		grafanaCfg map[string]string
		want       health.Category
	}{
		"broken json": {
			url: "127.0.0.1:9042", jsonData: `{"consistency":`, want: health.InvalidSettings,
		},
		"no contact point": {
			url: "", jsonData: `{"consistency":"ONE"}`, want: health.InvalidContactPoints,
		},
		"invalid consistency": {
			url: "127.0.0.1:9042", jsonData: `{"consistency":"QORUM"}`, want: health.InvalidConsistency,
		},
		"negative timeout": {
			url: "127.0.0.1:9042", jsonData: `{"consistency":"ONE","timeout":-5}`, want: health.InvalidTimeout,
		},
		"unparsable client certificate content": {
			url:      "127.0.0.1:9042",
			jsonData: `{"consistency":"ONE","UseCustomTLS":true,"useCertContent":true}`,
			secure:   map[string]string{"certContent": secretContent, "rootContent": secretContent},
			want:     health.InvalidTLSClientCert,
		},
		"unparsable CA content": {
			url:      "127.0.0.1:9042",
			jsonData: `{"consistency":"ONE","UseCustomTLS":true,"useCertContent":true}`,
			secure:   map[string]string{"caContent": secretContent},
			want:     health.InvalidTLSCA,
		},
		"missing client certificate files": {
			url:      "127.0.0.1:9042",
			jsonData: `{"consistency":"ONE","UseCustomTLS":true,"certPath":"/nonexistent/secret-dir/c.crt","rootPath":"/nonexistent/secret-dir/c.key"}`,
			want:     health.InvalidTLSClientCert,
		},
		"missing CA file": {
			url:      "127.0.0.1:9042",
			jsonData: `{"consistency":"ONE","UseCustomTLS":true,"caPath":"/nonexistent/secret-dir/ca.crt"}`,
			want:     health.InvalidTLSCA,
		},
		"secure socks proxy without its CA": {
			url:      "127.0.0.1:9042",
			jsonData: `{"consistency":"ONE","enableSecureSocksProxy":true}`,
			grafanaCfg: map[string]string{
				proxy.PluginSecureSocksProxyEnabled:      "true",
				proxy.PluginSecureSocksProxyProxyAddress: "localhost:5555",
			},
			want: health.SecureSocksProxy,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := backend.WithGrafanaConfig(context.Background(), backend.NewGrafanaCfg(tc.grafanaCfg))
			settings := backend.DataSourceInstanceSettings{
				URL:                     tc.url,
				JSONData:                []byte(tc.jsonData),
				DecryptedSecureJSONData: tc.secure,
			}

			instance, err := newDataSource(ctx, settings)
			require.Error(t, err)
			assert.Nil(t, instance)

			var public *health.PublicError
			require.True(t, errors.As(err, &public), "the factory only returns public errors, got %T", err)
			assert.Equal(t, tc.want, public.Category)

			problem := health.Describe(tc.want)
			assert.Equal(t, problem.Message+" "+problem.Hint, err.Error())
			assert.NotContains(t, err.Error(), fakeKeyBody)
			assert.NotContains(t, err.Error(), "secret-dir")
		})
	}
}
