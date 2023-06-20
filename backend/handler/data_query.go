package handler

import (
	"encoding/json"
	"fmt"

	"local_package/cassandra"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type dataQuery struct {
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
	Longitude      string `json:"longitude"`
	Latitude       string `json:"latitude"`
	Alias          string `json:"alias,omitempty"`
	AllowFiltering bool   `json:"filtering,omitempty"`
}

// parseDataQuery is a simple helper to unmarshal
// backend.DataQuery's JSON into the cassandra.Query type.
func parseDataQuery(q *backend.DataQuery) (*cassandra.Query, error) {
	var bq dataQuery
	err := json.Unmarshal(q.JSON, &bq)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &cassandra.Query{
		RawQuery:       bq.RawQuery,
		Target:         bq.Target,
		Keyspace:       bq.Keyspace,
		Table:          bq.Table,
		ColumnValue:    bq.ColumnValue,
		ColumnID:       bq.ColumnID,
		ValueID:        bq.ValueID,
		Longitude:      bq.Longitude,
		Latitude:       bq.Latitude,
		AliasID:        bq.Alias,
		ColumnTime:     bq.ColumnTime,
		TimeFrom:       q.TimeRange.From,
		TimeTo:         q.TimeRange.To,
		AllowFiltering: bq.AllowFiltering,
	}, nil
}
