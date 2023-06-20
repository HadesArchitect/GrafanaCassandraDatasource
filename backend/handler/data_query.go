package handler

import (
	"testing"
	"time"

	"local_package/cassandra"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
)

func Test_parseDataQuery(t *testing.T) {
	testCases := []struct {
		name      string
		timeRange backend.TimeRange
		jsonStr   []byte
		want      *cassandra.Query
	}{
		{
			name:      "all fields",
			timeRange: backend.TimeRange{From: time.Unix(1257894000, 0), To: time.Unix(1257894010, 0)},
			jsonStr: []byte(`{"datasourceId": 1, "queryType": "query", "rawQuery": true, "refId": "123456789",
							  "target": "SELECT * from Keyspace.Table", "columnTime": "Time", "columnValue": "Value",
							  "keyspace": "Keyspace", "table": "Table", "columnId": "ID", "valueId": "123","longitude": "Longitude","latitude": "Latitude",
							  "alias": "Alias", "filtering": true}`),
			want: &cassandra.Query{
				RawQuery:       true,
				Target:         "SELECT * from Keyspace.Table",
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				ValueID:        "123",
				Longitude:      "Longitude",
				Latitude:       "Latitude",
				AliasID:        "Alias",
				ColumnTime:     "Time",
				TimeFrom:       time.Unix(1257894000, 0),
				TimeTo:         time.Unix(1257894010, 0),
				AllowFiltering: true,
			},
		},
		{
			name:      "no optional fields",
			timeRange: backend.TimeRange{From: time.Unix(1257894000, 0), To: time.Unix(1257894010, 0)},
			jsonStr: []byte(`{"datasourceId": 1, "queryType": "query", "rawQuery": true, "refId": "123456789",
					   		  "target": "SELECT * from Keyspace.Table", "columnTime": "Time", "columnValue": "Value",
					   		  "keyspace": "Keyspace", "table": "Table", "columnId": "ID", "valueId": "123","longitude": "Longitude","latitude": "Latitude"}`),
			want: &cassandra.Query{
				RawQuery:       true,
				Target:         "SELECT * from Keyspace.Table",
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				ValueID:        "123",
				Longitude:      "Longitude",
				Latitude:       "Latitude",
				AliasID:        "",
				ColumnTime:     "Time",
				TimeFrom:       time.Unix(1257894000, 0),
				TimeTo:         time.Unix(1257894010, 0),
				AllowFiltering: false,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			query, err := parseDataQuery(&backend.DataQuery{TimeRange: tc.timeRange, JSON: tc.jsonStr})
			if err != nil {
				t.Fatalf("parseDataQuery: %v", err)
			}

			assert.Equal(t, tc.want, query)
		})
	}
}
