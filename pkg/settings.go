package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
)

// dataSourceSettings is a convenient presentation of a
// backend.DataSourceInstanceSettings JSON data.
type dataSourceSettings struct {
	Keyspace              string `json:"keyspace"`
	User                  string `json:"user"`
	Password              string `json:"password"`
	Consistency           string `json:"consistency"`
	CertPath              string `json:"certPath"`
	RootPath              string `json:"rootPath"`
	CaPath                string `json:"caPath"`
	UseCertContent        bool   `json:"useCertContent"`
	Timeout               *int   `json:"timeout"`
	UseCustomTLS          bool   `json:"UseCustomTLS"`
	AllowInsecureTLS      bool   `json:"allowInsecureTLS"`
	AllowedAuthenticators string `json:"allowedAuthenticators"`
	// MaxRows is the largest number of rows a query may return. Unset means
	// defaultMaxRows, 0 means no limit.
	MaxRows *int `json:"maxRows"`
	// EnableSecureSocksProxy mirrors the key the plugin SDK reads to decide
	// whether the datasource connection must go through Grafana's secure SOCKS
	// proxy (Private Data Source Connect). It is kept here only to be able to
	// tell "not requested" apart from "requested but unavailable" when logging.
	EnableSecureSocksProxy bool `json:"enableSecureSocksProxy"`
}

// defaultMaxRows is the row limit of a data source whose settings do not set
// maxRows. It keeps a query without a time filter from loading a whole table
// into Grafana's memory.
const defaultMaxRows = 100000

// resolveMaxRows returns the row limit for the maxRows setting: the default
// when it is not set, and no limit (0) when it is set to 0. A negative value
// is rejected rather than read as "no limit", as it is most likely a typo.
func resolveMaxRows(maxRows *int) (int, error) {
	if maxRows == nil {
		return defaultMaxRows, nil
	}
	if *maxRows < 0 {
		return 0, health.Wrap(health.InvalidMaxRows, fmt.Errorf("negative maxRows: %d", *maxRows))
	}

	return *maxRows, nil
}

// parseAllowedAuthenticators splits the semicolon-separated allowedAuthenticators
// setting into a slice, trimming whitespace and dropping empty entries. It
// returns nil when no authenticators are configured so that the cassandra package
// falls back to its default allow-list.
func parseAllowedAuthenticators(raw string) []string {
	parts := strings.Split(raw, ";")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return nil
	}

	return result
}

// prepareTLSCfgFromPaths creates a tls.Config using certificate file paths.
func prepareTLSCfgFromPaths(certPath, rootPath, caPath string, allowInsecureTLS bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: allowInsecureTLS}

	// Load client certificate and key from files
	if certPath != "" && rootPath != "" {
		cert, err := filepath.Abs(certPath)
		if err != nil {
			return nil, health.Wrap(health.InvalidTLSClientCert, fmt.Errorf("failed to resolve certificate path: %w", err))
		}
		key, err := filepath.Abs(rootPath)
		if err != nil {
			return nil, health.Wrap(health.InvalidTLSClientCert, fmt.Errorf("failed to resolve key path: %w", err))
		}
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, health.Wrap(health.InvalidTLSClientCert, fmt.Errorf("failed to load certificate from files: %w", err))
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}

	// Load CA certificate from file
	if caPath != "" {
		caCertPEMPath, err := filepath.Abs(caPath)
		if err != nil {
			return nil, health.Wrap(health.InvalidTLSCA, fmt.Errorf("failed to resolve CA certificate path: %w", err))
		}
		caCertPEM, err := ioutil.ReadFile(caCertPEMPath)
		if err != nil {
			return nil, health.Wrap(health.InvalidTLSCA, fmt.Errorf("failed to read CA certificate file: %w", err))
		}
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, health.Wrap(health.InvalidTLSCA, fmt.Errorf("failed to parse CA certificate from file"))
		}
		tlsConfig.RootCAs = roots
	}

	return tlsConfig, nil
}

// prepareTLSCfgFromContent creates a tls.Config using certificate content directly.
func prepareTLSCfgFromContent(certContent, rootContent, caContent string, allowInsecureTLS bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: allowInsecureTLS}

	// Load client certificate and key from content
	if certContent != "" && rootContent != "" {
		certificate, err := tls.X509KeyPair([]byte(certContent), []byte(rootContent))
		if err != nil {
			return nil, health.Wrap(health.InvalidTLSClientCert, fmt.Errorf("failed to parse certificate content: %w", err))
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}

	// Load CA certificate from content
	if caContent != "" {
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM([]byte(caContent)); !ok {
			return nil, health.Wrap(health.InvalidTLSCA, fmt.Errorf("failed to parse CA certificate content"))
		}
		tlsConfig.RootCAs = roots
	}

	return tlsConfig, nil
}
