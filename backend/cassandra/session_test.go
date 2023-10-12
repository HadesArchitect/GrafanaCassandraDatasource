package cassandra

import (
	"fmt"
	"testing"

	"github.com/gocql/gocql"
	"github.com/stretchr/testify/assert"
)

func Test_toString(t *testing.T) {
	testCases := []struct {
		name    string
		input   interface{}
		want    string
		wantErr error
	}{
		{
			name:    "string",
			input:   "00000000-0000-0000-0000-000000000000",
			want:    "00000000-0000-0000-0000-000000000000",
			wantErr: nil,
		},
		{
			name:    "UUID",
			input:   gocql.UUID{},
			want:    "00000000-0000-0000-0000-000000000000",
			wantErr: nil,
		},
		{
			name:    "int64",
			input:   int64(123),
			want:    "123",
			wantErr: nil,
		},
		{
			name:    "float64",
			input:   float64(0.1),
			want:    "0.100000",
			wantErr: nil,
		},
		{
			name:    "bool",
			input:   true,
			want:    "true",
			wantErr: nil,
		},
		{
			name:    "unsupported",
			input:   struct{}{},
			want:    "",
			wantErr: fmt.Errorf("unsupported type: %T", struct{}{}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val, err := toString(tc.input)
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.wantErr.Error())
			}
			assert.Equal(t, tc.want, val)
		})
	}
}

func Test_isSelect(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "select string low",
			input: "select * from table;",
			want:  true,
		},
		{
			name:  "select string up",
			input: "SELECT * from table;",
			want:  true,
		},
		{
			name:  "insert string",
			input: "insert into table (id) values ('test');",
			want:  false,
		},
		{
			name:  "delete string",
			input: "delete from table where id = 'test';",
			want:  false,
		},
		{
			name:  "delete string with column",
			input: "delete column from table where id = 'test';",
			want:  false,
		},
		{
			name:  "drop table string",
			input: "drop table test;",
			want:  false,
		},
		{
			name:  "truncate string",
			input: "truncate table test;",
			want:  false,
		},
		{
			name:  "drop keyspace string",
			input: "drop keyspace test;",
			want:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isSelect(tc.input)
			assert.Equal(t, tc.want, result)
		})
	}
}
