package handler

import (
	"encoding/json"
	"fmt"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/backend/plugin"
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
	Alias          string `json:"alias,omitempty"`
	AllowFiltering bool   `json:"filtering,omitempty"`
	Instant        bool   `json:"instant,omitempty"`
}

// parseDataQuery is a simple helper to unmarshal
// backend.DataQuery's JSON into the cassandra.Query type.
func parseDataQuery(q *backend.DataQuery) (*plugin.Query, error) {
	var dq dataQuery
	err := json.Unmarshal(q.JSON, &dq)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal: %w", err)
	}

	return &plugin.Query{
		RawQuery:       dq.RawQuery,
		Target:         dq.Target,
		Keyspace:       dq.Keyspace,
		Table:          dq.Table,
		ColumnValue:    dq.ColumnValue,
		ColumnID:       dq.ColumnID,
		ValueID:        dq.ValueID,
		AliasID:        dq.Alias,
		ColumnTime:     dq.ColumnTime,
		TimeFrom:       q.TimeRange.From,
		TimeTo:         q.TimeRange.To,
		AllowFiltering: dq.AllowFiltering,
		Instant:        dq.Instant,
	}, nil
}
