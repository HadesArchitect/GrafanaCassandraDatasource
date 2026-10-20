package plugin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestQuery_BuildStatement(t *testing.T) {
	testCases := []struct {
		name  string
		input *Query
		want  string
	}{
		{
			name: "without AllowFiltering and Instant",
			input: &Query{
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				ColumnTime:     "Time",
				AllowFiltering: false,
				Instant:        false,
			},
			want: "SELECT ID, Value, Time FROM Keyspace.Table WHERE ID IN ? AND Time >= ? AND Time <= ?",
		},
		{
			name: "with AllowFiltering",
			input: &Query{
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				ColumnTime:     "Time",
				AllowFiltering: true,
				Instant:        false,
			},
			want: "SELECT ID, Value, Time FROM Keyspace.Table WHERE ID IN ? AND Time >= ? AND Time <= ? ALLOW FILTERING",
		},
		{
			name: "with Instant",
			input: &Query{
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				ColumnTime:     "Time",
				AllowFiltering: false,
				Instant:        true,
			},
			want: "SELECT ID, Value, Time FROM Keyspace.Table WHERE ID IN ? AND Time >= ? AND Time <= ? PER PARTITION LIMIT 1",
		},
		{
			name: "with AllowFiltering and Instant",
			input: &Query{
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				ColumnTime:     "Time",
				AllowFiltering: true,
				Instant:        true,
			},
			want: "SELECT ID, Value, Time FROM Keyspace.Table WHERE ID IN ? AND Time >= ? AND Time <= ? PER PARTITION LIMIT 1 ALLOW FILTERING",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryString := tc.input.BuildStatement()
			assert.Equal(t, tc.want, queryString, tc.name)
		})
	}
}

func TestQuery_describeStatement(t *testing.T) {
	q := &Query{
		Keyspace: "ks", Table: "t", ColumnID: "id", ColumnValue: "v", ColumnTime: "ts",
		ValueID:  "a,b , c",
		TimeFrom: time.Date(2024, 3, 1, 10, 0, 0, 0, time.FixedZone("CET", 3600)),
		TimeTo:   time.Date(2024, 3, 1, 10, 30, 0, 123e6, time.UTC),
	}

	assert.Equal(t,
		"SELECT id, v, ts FROM ks.t WHERE id IN ? AND ts >= ? AND ts <= ?\n"+
			"-- bound values: id IN (a, b, c), ts >= 2024-03-01T09:00:00.000Z, ts <= 2024-03-01T10:30:00.123Z",
		q.describeStatement(q.BuildStatement()))
}
