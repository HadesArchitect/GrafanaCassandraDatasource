package cassandra

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
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
			name:    "int16 (smallint)",
			input:   int16(-12),
			want:    "-12",
			wantErr: nil,
		},
		{
			name:    "int (int)",
			input:   int(42),
			want:    "42",
			wantErr: nil,
		},
		{
			name:    "inet IPv4",
			input:   net.ParseIP("10.0.0.1").To4(),
			want:    "10.0.0.1",
			wantErr: nil,
		},
		{
			name:    "inet IPv6",
			input:   net.ParseIP("2001:db8::1"),
			want:    "2001:db8::1",
			wantErr: nil,
		},
		{
			name:    "inet NULL",
			input:   net.IP(nil),
			want:    "",
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

func Test_newClusterConfig(t *testing.T) {
	timeout := 30
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	dialer := &recordingDialer{}

	cluster, err := newClusterConfig(Settings{
		Hosts:       []string{"10.0.0.1:9042", "10.0.0.2:9042"},
		Keyspace:    "metrics",
		User:        "grafana",
		Password:    "secret",
		Consistency: "local_quorum",
		Timeout:     &timeout,
		TLSConfig:   tlsConfig,
		Dialer:      dialer,
	})

	assert.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.1:9042", "10.0.0.2:9042"}, cluster.Hosts)
	assert.Equal(t, "metrics", cluster.Keyspace)
	assert.True(t, cluster.DisableInitialHostLookup, "required for AWS Keyspaces")
	assert.Equal(t, gocql.Disabled, cluster.Metadata.CacheMode)
	assert.Equal(t, gocql.LocalQuorum, cluster.Consistency)
	assert.Equal(t, 30*time.Second, cluster.Timeout)
	assert.Same(t, tlsConfig, cluster.SslOpts.Config)
	assert.Same(t, dialer, cluster.Dialer)
	assert.NotNil(t, cluster.Logger)
	assert.Equal(t, gocql.PasswordAuthenticator{
		Username:              "grafana",
		Password:              "secret",
		AllowedAuthenticators: defaultAllowedAuthenticators,
	}, cluster.Authenticator)
}

func Test_newClusterConfig_defaults(t *testing.T) {
	cluster, err := newClusterConfig(Settings{
		Hosts:       []string{"127.0.0.1"},
		Consistency: "ONE",
	})

	assert.NoError(t, err)
	assert.Equal(t, gocql.One, cluster.Consistency)
	assert.Equal(t, gocql.NewCluster().Timeout, cluster.Timeout, "driver default timeout is kept")
	assert.Nil(t, cluster.SslOpts)
	assert.Nil(t, cluster.Dialer)
}

func Test_newClusterConfig_allowedAuthenticators(t *testing.T) {
	ldap := []string{
		"org.apache.cassandra.auth.PasswordAuthenticator",
		"org.apache.cassandra.auth.LDAPAuthenticator",
	}

	cluster, err := newClusterConfig(Settings{
		Hosts:                 []string{"127.0.0.1"},
		Consistency:           "ONE",
		AllowedAuthenticators: ldap,
	})

	assert.NoError(t, err)
	auth, ok := cluster.Authenticator.(gocql.PasswordAuthenticator)
	assert.True(t, ok)
	assert.Equal(t, ldap, auth.AllowedAuthenticators, "a configured list replaces the default one")
}

// Test_defaultAllowedAuthenticators guards the default allow-list: gocql v2
// accepts any server authenticator when the list is empty, so an empty default
// would silently relax what v1 enforced.
func Test_defaultAllowedAuthenticators(t *testing.T) {
	assert.Contains(t, defaultAllowedAuthenticators, "org.apache.cassandra.auth.PasswordAuthenticator")
	assert.Contains(t, defaultAllowedAuthenticators, "com.datastax.bdp.cassandra.auth.DseAuthenticator")
	assert.Contains(t, defaultAllowedAuthenticators, "com.amazon.helenus.auth.HelenusAuthenticator")
	assert.NotContains(t, defaultAllowedAuthenticators, "org.apache.cassandra.auth.LDAPAuthenticator")
	assert.Len(t, defaultAllowedAuthenticators, 10)
}

func Test_newClusterConfig_invalidConsistency(t *testing.T) {
	cluster, err := newClusterConfig(Settings{
		Hosts:       []string{"127.0.0.1"},
		Consistency: "SOMETIMES",
	})

	assert.Error(t, err)
	assert.Nil(t, cluster)
}

func Test_typeName(t *testing.T) {
	testCases := map[gocql.Type]string{
		gocql.TypeAscii:     "ascii",
		gocql.TypeBigInt:    "bigint",
		gocql.TypeBoolean:   "boolean",
		gocql.TypeCounter:   "counter",
		gocql.TypeDate:      "date",
		gocql.TypeDecimal:   "decimal",
		gocql.TypeDouble:    "double",
		gocql.TypeFloat:     "float",
		gocql.TypeInet:      "inet",
		gocql.TypeInt:       "int",
		gocql.TypeSmallInt:  "smallint",
		gocql.TypeText:      "text",
		gocql.TypeTimestamp: "timestamp",
		gocql.TypeTimeUUID:  "timeuuid",
		gocql.TypeTinyInt:   "tinyint",
		gocql.TypeUUID:      "uuid",
		gocql.TypeVarchar:   "varchar",
		gocql.TypeVarint:    "varint",
		gocql.TypeList:      "list",
		gocql.TypeMap:       "map",
		gocql.TypeSet:       "set",
		gocql.Type(0x00FF):  "unknown_type_255",
	}

	for typ, want := range testCases {
		t.Run(want, func(t *testing.T) {
			assert.Equal(t, want, typeName(typ))
		})
	}
}
