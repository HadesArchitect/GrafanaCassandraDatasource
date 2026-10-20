package cassandra

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// defaultAllowedAuthenticators is the list of server-side authenticators the
// driver may authenticate against when no list is configured explicitly.
//
// gocql v1 enforced this list itself; v2 accepts any authenticator the server
// announces unless PasswordAuthenticator.AllowedAuthenticators is set. The list
// is kept here, unchanged, so that upgrading the driver does not silently widen
// the set of authenticators credentials are sent to.
var defaultAllowedAuthenticators = []string{
	"org.apache.cassandra.auth.PasswordAuthenticator",
	"com.instaclustr.cassandra.auth.SharedSecretAuthenticator",
	"com.datastax.bdp.cassandra.auth.DseAuthenticator",
	"io.aiven.cassandra.auth.AivenAuthenticator",
	"com.ericsson.bss.cassandra.ecaudit.auth.AuditPasswordAuthenticator",
	"com.amazon.helenus.auth.HelenusAuthenticator",
	"com.ericsson.bss.cassandra.ecaudit.auth.AuditAuthenticator",
	"com.scylladb.auth.SaslauthdAuthenticator",
	"com.scylladb.auth.TransitionalAuthenticator",
	"com.instaclustr.cassandra.auth.InstaclustrPasswordAuthenticator",
}

// Dialer establishes the TCP connections to the cluster nodes. It matches
// gocql.Dialer and is satisfied by the secure SOCKS proxy dialer provided by
// the Grafana plugin SDK, which is what Private Data Source Connect is built on.
type Dialer interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

// Settings is a set of Cassandra session settings.
type Settings struct {
	Hosts                 []string
	Keyspace              string
	User                  string
	Password              string
	Consistency           string
	Timeout               *int
	TLSConfig             *tls.Config
	AllowedAuthenticators []string
	Dialer                Dialer
}

// Session is a convenience wrapper for the gocql.Session.
type Session struct {
	session *gocql.Session
}

// New creates a new cassandra cluster session using provided settings.
func New(cfg Settings) (*Session, error) {
	cluster, err := newClusterConfig(cfg)
	if err != nil {
		return nil, err
	}

	clusterSession, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("cluster.CreateSession: %w", err)
	}

	return &Session{clusterSession}, nil
}

// newClusterConfig translates the plugin settings into a driver cluster config.
func newClusterConfig(cfg Settings) (*gocql.ClusterConfig, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.DisableInitialHostLookup = true // required, AWS specific
	cluster.Keyspace = cfg.Keyspace
	cluster.Logger = newDriverLogger(backend.Logger)

	// By default gocql v2 loads and caches the schema of every keyspace when
	// the session is created. The plugin only needs schema metadata to list
	// tables and columns in the Query Configurator, so it is read on demand
	// instead: session creation stays cheap, the lists always reflect the
	// current schema, and a failure to read the schema tables is reported to
	// the user instead of turning into "keyspace does not exist".
	cluster.Metadata.CacheMode = gocql.Disabled

	// The default allow-list is used when no authenticators are configured,
	// keeping backwards compatibility. When provided (e.g. to permit
	// org.apache.cassandra.auth.LDAPAuthenticator) the list replaces the default.
	allowedAuthenticators := cfg.AllowedAuthenticators
	if len(allowedAuthenticators) == 0 {
		allowedAuthenticators = defaultAllowedAuthenticators
	}
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username:              cfg.User,
		Password:              cfg.Password,
		AllowedAuthenticators: allowedAuthenticators,
	}

	consistencyLevel, err := gocql.ParseConsistencyWrapper(cfg.Consistency)
	if err != nil {
		return nil, fmt.Errorf("gocql.ParseConsistencyWrapper: %w", err)
	}
	cluster.Consistency = consistencyLevel

	if cfg.Timeout != nil {
		cluster.Timeout = time.Duration(*cfg.Timeout) * time.Second
	}

	if cfg.TLSConfig != nil {
		cluster.SslOpts = &gocql.SslOptions{Config: cfg.TLSConfig}
	}

	// gocql keeps applying SslOpts on top of a custom Dialer (it only ignores
	// them for a custom HostDialer), so TLS and the secure SOCKS proxy can be
	// used together.
	if cfg.Dialer != nil {
		cluster.Dialer = cfg.Dialer
	}

	return cluster, nil
}

// Select queries the database with provided query string and returns result rows grouped by ID.
// ID must be a first requested column in query and must be convertable to a string.
func (s *Session) Select(ctx context.Context, query string, values ...interface{}) (rows map[string][]Row, err error) {
	if !isSelect(query) {
		return nil, fmt.Errorf("query is not a SELECT statement: %s", query)
	}

	iter := s.session.Query(query, values...).IterContext(ctx)
	defer func() {
		if iterErr := iter.Close(); iterErr != nil {
			err = fmt.Errorf("select query processing: %w", iterErr)
		}
	}()

	rows = make(map[string][]Row)
	for {
		rowValues := make(map[string]interface{}, len(iter.Columns()))
		if !iter.MapScan(rowValues) {
			break
		}

		// first field is considered an id and used to distinguish different timeseries,
		// so it must have string type. We are trying to convert id field value to
		// a string or exit early in case when such conversion is not supported.
		idFieldName := iter.Columns()[0].Name
		id, err := toString(rowValues[idFieldName])
		if err != nil {
			return nil, fmt.Errorf("row processing: %w", err)
		}

		row := Row{
			Columns: columnNames(iter.Columns()),
			Fields:  rowValues,
		}
		if err := row.normalize(); err != nil {
			return nil, fmt.Errorf("row.normalize: %w", err)
		}
		rows[id] = append(rows[id], row)
	}

	return rows, nil
}

// GetKeyspaces queries the cassandra cluster for a list of existing keyspaces.
func (s *Session) GetKeyspaces(ctx context.Context) ([]string, error) {
	statement := "SELECT keyspace_name FROM system_schema.keyspaces"
	iter := s.session.Query(statement).IterContext(ctx)

	var (
		keyspace  string
		keyspaces []string
	)

	for iter.Scan(&keyspace) {
		keyspaces = append(keyspaces, keyspace)
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("query processing: %w", err)
	}

	return keyspaces, nil
}

// GetTables queries the cassandra cluster for a list of an existing tables in a given keyspace.
func (s *Session) GetTables(keyspace string) ([]string, error) {
	keyspaceMetadata, err := s.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("session.KeyspaceMetadata: %w", err)
	}

	tables := make([]string, 0, len(keyspaceMetadata.Tables))
	for tableName := range keyspaceMetadata.Tables {
		tables = append(tables, tableName)
	}

	return tables, nil
}

// GetColumns queries the cassandra cluster for a list of an
// existing columns of a given type for a given keyspace, table.
func (s *Session) GetColumns(keyspace, table, needType string) ([]string, error) {
	keyspaceMetadata, err := s.session.KeyspaceMetadata(keyspace)
	if err != nil {
		return nil, fmt.Errorf("session.KeyspaceMetadata: %w", err)
	}

	tableMetadata, ok := keyspaceMetadata.Tables[table]
	if !ok {
		return nil, fmt.Errorf("no such table: '%s'", table)
	}

	columns := make([]string, 0, len(tableMetadata.Columns))
	for name, column := range tableMetadata.Columns {
		if typeName(column.Type.Type()) == needType {
			columns = append(columns, name)
		}
	}

	return columns, nil
}

// Ping executes a simple query to check the connection status.
func (s *Session) Ping(ctx context.Context) error {
	err := s.session.Query("SELECT key FROM system.local").ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("session.Query: %w", err)
	}

	return nil
}

// Close closes connections to cluster.
func (s *Session) Close() {
	s.session.Close()
}

func isSelect(query string) bool {
	stmt := strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(stmt), "SELECT ") {
		return false
	}

	return true
}

func toString(val interface{}) (string, error) {
	var str string
	switch v := val.(type) {
	case string:
		str = v
	case gocql.UUID:
		str = v.String()
	case int8, int16, int32, int64, int:
		str = fmt.Sprintf("%d", v)
	case float32, float64:
		str = fmt.Sprintf("%f", v)
	case time.Time:
		str = v.String()
	case bool:
		str = fmt.Sprintf("%t", v)
	case net.IP:
		// gocql v2 returns inet columns as net.IP (v1 returned a string).
		str = ipString(v)
	default:
		return "", fmt.Errorf("unsupported type: %T", val)
	}

	return str, nil
}

func columnNames(columnInfo []gocql.ColumnInfo) []string {
	names := make([]string, 0, len(columnInfo))
	for _, col := range columnInfo {
		names = append(names, col.Name)
	}

	return names
}
