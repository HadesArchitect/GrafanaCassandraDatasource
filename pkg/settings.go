package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// dataSourceSettings is a convenient presentation of a
// backend.DataSourceInstanceSettings JSON data.
type dataSourceSettings struct {
	Keyspace         string `json:"keyspace"`
	User             string `json:"user"`
	Password         string `json:"password"`
	Consistency      string `json:"consistency"`
	CertPath         string `json:"certPath"`
	RootPath         string `json:"rootPath"`
	CaPath           string `json:"caPath"`
	UseCertContent   bool   `json:"useCertContent"`
	Timeout          *int   `json:"timeout"`
	UseCustomTLS     bool   `json:"UseCustomTLS"`
	AllowInsecureTLS bool   `json:"allowInsecureTLS"`
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
