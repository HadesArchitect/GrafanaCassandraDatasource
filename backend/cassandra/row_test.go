package cassandra

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

func TestRow_normalize(t *testing.T) {
	testCases := []struct {
		name    string
		input   *Row
		want    *Row
		wantErr error
	}{
		{
			name:    "empty",
			input:   &Row{},
			want:    &Row{},
			wantErr: nil,
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
			wantErr: nil,
		},
		{
			name: "normal row with nil in the beginning",
			input: &Row{
				Columns: []string{"field1", "id", "time", "value"},
				Fields:  map[string]interface{}{"field1": nil, "id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"field1", "id", "time", "value"},
				Fields:  map[string]interface{}{"field1": nil, "id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", nil),
		},
		{
			name: "normal row with nil in the end",
			input: &Row{
				Columns: []string{"id", "time", "value", "field1"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1, "field1": nil},
			},
			want: &Row{
				Columns: []string{"id", "time", "value", "field1"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1, "field1": nil},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", nil),
		},
		{
			name: "normal row with unsupported",
			input: &Row{
				Columns: []string{"id", "field1", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "field1": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "field1", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "field1": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", struct{}{}),
		},
		{
			name: "normal row with multiple unsupported fields",
			input: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": struct{}{}, "id": "id", "field2": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "field3": struct{}{}, "value": 0.1},
			},
			want: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": struct{}{}, "id": "id", "field2": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "field3": struct{}{}, "value": 0.1},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", struct{}{}),
		},
		{
			name: "normal row with conversion",
			input: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": gocql.UUID{}, "id": "id", "field2": net.ParseIP("127.0.0.1"), "time": time.UnixMilli(1257894000000).UTC(), "field3": []byte("some string"), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": "00000000-0000-0000-0000-000000000000", "id": "id", "field2": "127.0.0.1", "time": time.UnixMilli(1257894000000).UTC(), "field3": "some string", "value": 0.1},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.normalize()
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.wantErr.Error())
			}
			assert.Equal(t, tc.want, tc.input)
		})
	}
}
