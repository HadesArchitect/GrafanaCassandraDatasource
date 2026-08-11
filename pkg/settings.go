package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
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
}

// parseAllowedAuthenticators splits the semicolon-separated allowedAuthenticators
// setting into a slice, trimming whitespace and dropping empty entries. It
// returns nil when no authenticators are configured so that gocql falls back to
// its built-in default list.
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

// consistencyLevels are the levels the driver understands. Anything else is
// rejected when the config is saved rather than on the first query.
var consistencyLevels = []string{
	"ANY", "ONE", "TWO", "THREE", "QUORUM", "ALL",
	"LOCAL_QUORUM", "EACH_QUORUM", "LOCAL_ONE",
}

// validateConsistency reports whether the configured consistency level is one
// the driver accepts.
func validateConsistency(level string) error {
	for _, known := range consistencyLevels {
		if strings.EqualFold(level, known) {
			return nil
		}
	}

	return fmt.Errorf("unknown consistency level %q, expected one of: %s", level, strings.Join(consistencyLevels, ", "))
}

// prepareTLSCfgFromPaths creates a tls.Config using certificate file paths.
func prepareTLSCfgFromPaths(certPath, rootPath, caPath string, allowInsecureTLS bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: allowInsecureTLS}

	// Load client certificate and key from files
	if certPath != "" && rootPath != "" {
		cert, err := filepath.Abs(certPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve certificate path: %w", err)
		}
		key, err := filepath.Abs(rootPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve key path: %w", err)
		}
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("failed to load certificate from files: %w", err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}

	// Load CA certificate from file
	if caPath != "" {
		caCertPEMPath, err := filepath.Abs(caPath)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve CA certificate path: %w", err)
		}
		caCertPEM, err := ioutil.ReadFile(caCertPEMPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate file: %w", err)
		}
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate from file")
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
			return nil, fmt.Errorf("failed to parse certificate content: %w", err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}

	// Load CA certificate from content
	if caContent != "" {
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM([]byte(caContent)); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate content")
		}
		tlsConfig.RootCAs = roots
	}

	return tlsConfig, nil
}
