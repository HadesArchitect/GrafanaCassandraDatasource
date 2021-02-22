package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
)

func PrepareTLSCfg(certPath string, rootPath string, caPath string) (*tls.Config, error) {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	if certPath != "" && rootPath != "" {
		cert, err := Asset(certPath)
		if err != nil {
			return nil, err
		}
		key, err := Asset(rootPath)
		if err != nil {
			return nil, err
		}
		certificate, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}
	if caPath != "" {
		caCertPEM, err := Asset(caPath)
		if err != nil {
			return nil, err
		}
		roots := x509.NewCertPool()
		if ok := roots.AppendCertsFromPEM(caCertPEM); !ok {
			return nil, errors.New("failed to parse root certificate")
		}
		tlsConfig.RootCAs = roots
	}
	return tlsConfig, nil
}
