package main

type RequestOptions struct {
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
