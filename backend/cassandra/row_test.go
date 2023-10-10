package cassandra

import (
	"sort"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_filterUnsupportedTypes(t *testing.T) {
	testCases := []struct {
		name  string
		input *Row
		want  *Row
	}{
		{
			name:  "empty",
			input: &Row{},
			want:  &Row{},
		},
		{
			name: "normal row",
			input: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
		},
		{
			name: "normal row with nil in the beginning",
			input: &Row{
				Columns: []string{"field1", "id", "time", "value"},
				Fields:  map[string]interface{}{"field1": nil, "id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
		},
		{
			name: "normal row with nil in the end",
			input: &Row{
				Columns: []string{"id", "time", "value", "field1"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1, "field1": nil},
			},
			want: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
		},
		{
			name: "normal row with unsupported",
			input: &Row{
				Columns: []string{"id", "field1", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "field1": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
		},
		{
			name: "normal row with multiple unsupported fields",
			input: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": struct{}{}, "id": "id", "field2": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "field3": struct{}{}, "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.input.filterUnsupportedTypes()
			sort.Slice(tc.input.Columns, func(i, j int) bool {
				return tc.input.Columns[i] < tc.input.Columns[j]
			})
			sort.Slice(tc.want.Columns, func(i, j int) bool {
				return tc.want.Columns[i] < tc.want.Columns[j]
			})
			assert.Equal(t, tc.want, tc.input, tc.name)
		})
	}
}
