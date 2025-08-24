package main

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_prepareTLSCfgFromPaths(t *testing.T) {
	testCases := []struct {
		name            string
		certPath        string
		rootPath        string
		caPath          string
		allowInsecureTLS bool
		wantErr         bool
		validate        func(*testing.T, *tls.Config)
	}{
		{
			name:            "no certificates",
			certPath:        "",
			rootPath:        "",
			caPath:          "",
			allowInsecureTLS: false,
			wantErr:         false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Empty(t, cfg.Certificates)
				assert.Nil(t, cfg.RootCAs)
				assert.False(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:            "insecure TLS enabled",
			certPath:        "",
			rootPath:        "",
			caPath:          "",
			allowInsecureTLS: true,
			wantErr:         false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.True(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:            "file paths with non-existent files",
			certPath:        "/nonexistent/cert.pem",
			rootPath:        "/nonexistent/key.pem",
			caPath:          "",
			allowInsecureTLS: false,
			wantErr:         true,
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
		name            string
		certContent     string
		rootContent     string
		caContent       string
		allowInsecureTLS bool
		wantErr         bool
		validate        func(*testing.T, *tls.Config)
	}{
		{
			name:            "no certificates",
			certContent:     "",
			rootContent:     "",
			caContent:       "",
			allowInsecureTLS: false,
			wantErr:         false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.Empty(t, cfg.Certificates)
				assert.Nil(t, cfg.RootCAs)
				assert.False(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:            "insecure TLS enabled",
			certContent:     "",
			rootContent:     "",
			caContent:       "",
			allowInsecureTLS: true,
			wantErr:         false,
			validate: func(t *testing.T, cfg *tls.Config) {
				assert.NotNil(t, cfg)
				assert.True(t, cfg.InsecureSkipVerify)
			},
		},
		{
			name:            "invalid certificate content",
			certContent:     "invalid cert",
			rootContent:     "invalid key",
			caContent:       "",
			allowInsecureTLS: false,
			wantErr:         true,
			validate: func(t *testing.T, cfg *tls.Config) {
				// Should not reach here due to error
			},
		},
		{
			name:            "invalid CA certificate content",
			certContent:     "",
			rootContent:     "",
			caContent:       "invalid ca cert",
			allowInsecureTLS: false,
			wantErr:         true,
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