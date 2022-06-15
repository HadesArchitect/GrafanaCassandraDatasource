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
	Timeout          *int   `json:"timeout"`
	UseCustomTLS     bool   `json:"UseCustomTLS"`
	AllowInsecureTLS bool   `json:"allowInsecureTLS"`
}

// prepareTLSCfg is a helper to create a tls.Config using provided file paths.
func prepareTLSCfg(certPath string, rootPath string, caPath string, allowInsecureTLS bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: allowInsecureTLS}
	if certPath != "" && rootPath != "" {
		cert, err := filepath.Abs(certPath)
		if err != nil {
			return nil, err
		}
		key, err := filepath.Abs(rootPath)
		if err != nil {
			return nil, err
		}
		certificate, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}

	if caPath != "" {
		caCertPEMPath, err := filepath.Abs(caPath)
		if err != nil {
			return nil, err
		}
		caCertPEM, err := ioutil.ReadFile(caCertPEMPath)
		if err != nil {
			return nil, err
		}
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, fmt.Errorf("failed to parse root certificate")
		}
		tlsConfig.RootCAs = roots
	}

	return tlsConfig, nil
}