package main

type CassandraQuery struct {
	DatasourceID int    `json:"datasourceId"`
	QueryType    string `json:"queryType"`
	RawQuery     bool   `json:"rawQuery"`
	RefID        string `json:"refId"`
	Target       string `json:"target"`

	ColumnTime     string `json:"columnTime"`
	ColumnValue    string `json:"columnValue"`
	Keyspace       string `json:"keyspace"`
	Table          string `json:"table"`
	ColumnID       string `json:"columnId"`
	ValueID        string `json:"valueId"`
	AllowFiltering bool   `json:"filterint,omitempty"`
}
