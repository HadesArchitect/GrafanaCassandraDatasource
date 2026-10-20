package main

import (
	"crypto/tls"
	"encoding/json"
	"testing"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseAllowedAuthenticators(t *testing.T) {
	testCases := []struct {
		name string
		raw  string
		want []string
	}{
		{
			name: "empty string returns nil (default allow-list)",
			raw:  "",
			want: nil,
		},
		{
			name: "whitespace only returns nil",
			raw:  "   ;  ; ",
			want: nil,
		},
		{
			name: "single value",
			raw:  "org.apache.cassandra.auth.LDAPAuthenticator",
			want: []string{"org.apache.cassandra.auth.LDAPAuthenticator"},
		},
		{
			name: "multiple values are trimmed",
			raw:  " org.apache.cassandra.auth.PasswordAuthenticator ; org.apache.cassandra.auth.LDAPAuthenticator ",
			want: []string{
				"org.apache.cassandra.auth.PasswordAuthenticator",
				"org.apache.cassandra.auth.LDAPAuthenticator",
			},
		},
		{
			name: "empty entries are dropped",
			raw:  "a;;b;",
			want: []string{"a", "b"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, parseAllowedAuthenticators(tc.raw))
		})
	}
}

func Test_prepareTLSCfgFromPaths(t *testing.T) {
	testCases := []struct {
		name             string
		certPath         string
		rootPath         string
		caPath           string
		allowInsecureTLS bool
		wantErr          bool
		validate         func(*testing.T, *tls.Config)
	}{
		{
			name:             "no certificates",
			certPath:         "",
			rootPath:         "",
			caPath:           "",
			allowInsecureTLS: false,
			wantErr:          false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Empty(t, cfg.Certificates)
				assert.Nil(t, cfg.RootCAs)
				assert.False(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:             "insecure TLS enabled",
			certPath:         "",
			rootPath:         "",
			caPath:           "",
			allowInsecureTLS: true,
			wantErr:          false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.True(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:             "file paths with non-existent files",
			certPath:         "/nonexistent/cert.pem",
			rootPath:         "/nonexistent/key.pem",
			caPath:           "",
			allowInsecureTLS: false,
			wantErr:          true,
			validate: func(t *testing.T, cfg *tls.Config) {
				// Should not reach here due to error
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := prepareTLSCfgFromPaths(
				tc.certPath,
				tc.rootPath,
				tc.caPath,
				tc.allowInsecureTLS,
			)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tc.validate(t, cfg)
			}
		})
	}
}

func Test_prepareTLSCfgFromContent(t *testing.T) {
	testCases := []struct {
		name             string
		certContent      string
		rootContent      string
		caContent        string
		allowInsecureTLS bool
		wantErr          bool
		validate         func(*testing.T, *tls.Config)
	}{
		{
			name:             "no certificates",
			certContent:      "",
			rootContent:      "",
			caContent:        "",
			allowInsecureTLS: false,
			wantErr:          false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Empty(t, cfg.Certificates)
				assert.Nil(t, cfg.RootCAs)
				assert.False(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:             "insecure TLS enabled",
			certContent:      "",
			rootContent:      "",
			caContent:        "",
			allowInsecureTLS: true,
			wantErr:          false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.True(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:             "invalid certificate content",
			certContent:      "invalid cert",
			rootContent:      "invalid key",
			caContent:        "",
			allowInsecureTLS: false,
			wantErr:          true,
			validate: func(t *testing.T, cfg *tls.Config) {
				// Should not reach here due to error
			},
		},
		{
			name:             "invalid CA certificate content",
			certContent:      "",
			rootContent:      "",
			caContent:        "invalid ca cert",
			allowInsecureTLS: false,
			wantErr:          true,
			validate: func(t *testing.T, cfg *tls.Config) {
				// Should not reach here due to error
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := prepareTLSCfgFromContent(
				tc.certContent,
				tc.rootContent,
				tc.caContent,
				tc.allowInsecureTLS,
			)

			if tc.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tc.validate(t, cfg)
			}
		})
	}
}

func Test_prepareTLSCfgFromContent_EmptyContent(t *testing.T) {
	// Test that empty certificate content is handled gracefully
	cfg, err := prepareTLSCfgFromContent(
		"",    // certContent (empty)
		"",    // rootContent (empty)
		"",    // caContent (empty)
		false, // allowInsecureTLS
	)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Empty(t, cfg.Certificates)
	assert.Nil(t, cfg.RootCAs)
}

func Test_prepareTLSCfgFromPaths_NonExistentFiles(t *testing.T) {
	// Test that non-existent file paths return an error
	cfg, err := prepareTLSCfgFromPaths(
		"/nonexistent/cert.pem", // certPath
		"/nonexistent/key.pem",  // rootPath
		"",                      // caPath (empty)
		false,                   // allowInsecureTLS
	)

	// Should get an error because the files don't exist
	assert.Error(t, err)
	assert.Nil(t, cfg)
}
func Test_resolveMaxRows(t *testing.T) {
	intPtr := func(v int) *int { return &v }

	testCases := map[string]struct {
		setting *int
		want    int
	}{
		"unset uses the default":  {setting: nil, want: 100000},
		"zero means no limit":     {setting: intPtr(0), want: 0},
		"explicit limit":          {setting: intPtr(500), want: 500},
		"limit above the default": {setting: intPtr(5000000), want: 5000000},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, err := resolveMaxRows(tc.setting)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}

	t.Run("negative is rejected", func(t *testing.T) {
		_, err := resolveMaxRows(intPtr(-1))
		require.Error(t, err)
		assert.Equal(t, health.InvalidMaxRows, health.Classify(err))
	})
}

func Test_dataSourceSettings_maxRows(t *testing.T) {
	var unset, zero, set dataSourceSettings
	require.NoError(t, json.Unmarshal([]byte(`{"consistency":"ONE"}`), &unset))
	require.NoError(t, json.Unmarshal([]byte(`{"maxRows":0}`), &zero))
	require.NoError(t, json.Unmarshal([]byte(`{"maxRows":250}`), &set))

	assert.Nil(t, unset.MaxRows, "an unset key must be told apart from 0, which means no limit")
	require.NotNil(t, zero.MaxRows)
	assert.Equal(t, 0, *zero.MaxRows)
	require.NotNil(t, set.MaxRows)
	assert.Equal(t, 250, *set.MaxRows)
}
