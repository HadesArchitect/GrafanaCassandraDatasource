package cassandra

import (
	"context"
	"errors"
	"fmt"
	"net"
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

// recordingDialer is a Dialer that records the addresses it was asked to dial
// and always fails, so that New returns before a real cluster is needed.
type recordingDialer struct {
	addrs []string
}

func (d *recordingDialer) DialContext(_ context.Context, _, addr string) (net.Conn, error) {
	d.addrs = append(d.addrs, addr)
	return nil, errors.New("dial refused by test dialer")
}

func Test_New_usesProvidedDialer(t *testing.T) {
	dialer := &recordingDialer{}

	session, err := New(Settings{
		Hosts:       []string{"127.0.0.1:19042"},
		Consistency: "ONE",
		Dialer:      dialer,
	})

	assert.Error(t, err)
	assert.Nil(t, session)
	assert.Contains(t, dialer.addrs, "127.0.0.1:19042")
}
