package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"strconv"
)

var (
	CertPath           string
	KeyPath            string
	RootCA             string
	InsecureSkipVerify string
)

func PrepareTLSCfg() (*tls.Config, error) {
	skipVerify, _ := strconv.ParseBool(InsecureSkipVerify)
	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipVerify,
	}
	if CertPath != "" && KeyPath != "" {
		cert, err := Asset(CertPath)
		if err != nil {
			return nil, err
		}
		key, err := Asset(KeyPath)
		if err != nil {
			return nil, err
		}
		certificate, err := tls.X509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, certificate)
	}
	if RootCA != "" {
		caCertPEM, err := Asset(RootCA)
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
