package cassandra

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_buildStatement(t *testing.T) {
	testCases := []struct {
		name  string
		input *Query
		want  string
	}{
		{
			name: "without AllowFiltering",
			input: &Query{
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				Longitude:      "Longitude",
				Latitude:       "Latitude",
				ColumnTime:     "Time",
				AllowFiltering: false,
			},
			want: "SELECT Time, CAST(Value as double), Longitude, Latitude FROM Keyspace.Table WHERE ID = ? AND Time >= ? AND Time <= ?",
		},
		{
			name: "with AllowFiltering",
			input: &Query{
				Keyspace:       "Keyspace",
				Table:          "Table",
				ColumnValue:    "Value",
				ColumnID:       "ID",
				Longitude:      "Longitude",
				Latitude:       "Latitude",
				ColumnTime:     "Time",
				AllowFiltering: true,
			},
			want: "SELECT Time, CAST(Value as double), Longitude, Latitude FROM Keyspace.Table WHERE ID = ? AND Time >= ? AND Time <= ? ALLOW FILTERING",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			queryString := buildStatement(tc.input)
			assert.Equal(t, tc.want, queryString, tc.name)
		})
	}
}
