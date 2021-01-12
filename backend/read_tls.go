package main

import (
	"crypto/tls"
	"strconv"
)

var (
	CertPath           string
	KeyPath            string
	InsecureSkipVerify string
)

func ReadTLSCert() (*tls.Config, error) {
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
	skipVerify, _ := strconv.ParseBool(InsecureSkipVerify)
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		InsecureSkipVerify: skipVerify,
	}
	return tlsConfig, nil
}
