package plugin

import (
	"testing"

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
