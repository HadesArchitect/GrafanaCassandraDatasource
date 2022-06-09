package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/gocql/gocql"
)

// Session is a convenience wrapper for the gocql.Session.
type Session struct {
	session *gocql.Session
}

// New creates a new cassandra cluster session using provided settiongs.
func New(cfg Settings) (*Session, error) {
	cluster := gocql.NewCluster(cfg.Hosts...)
	cluster.DisableInitialHostLookup = true // required, AWS specific
	cluster.Keyspace = cfg.Keyspace

	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: cfg.User,
		Password: cfg.Password,
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

	clusterSession, err := cluster.CreateSession()
	if err != nil {
		return nil, fmt.Errorf("cluster.CreateSession: %w", err)
	}

	return &Session{clusterSession}, nil
}

// ExecRawQuery queries cassandra with a Query.Target query and returns
// a map of time series points grouped by a key column name.
func (s *Session) ExecRawQuery(ctx context.Context, q *Query) (map[string][]*TimeSeriesPoint, error) {
	iter := s.session.Query(q.Target).WithContext(ctx).Iter()

	ts := make(map[string][]*TimeSeriesPoint)
	var (
		id        string
		value     float64
		timestamp time.Time
	)

	for iter.Scan(&id, &value, &timestamp) {
		ts[id] = append(ts[id], &TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("raw query processing: %w", err)
	}

	return ts, nil
}

// ExecStrictQuery queries cassandra with passed Query parameters
// and returns a slice of time series points.
func (s *Session) ExecStrictQuery(ctx context.Context, q *Query) ([]*TimeSeriesPoint, error) {
	statement := buildStatement(q)
	iter := s.session.Query(statement, q.ValueID, q.TimeFrom, q.TimeTo).WithContext(ctx).Iter()

	var (
		ts        []*TimeSeriesPoint
		value     float64
		timestamp time.Time
	)

	for iter.Scan(&timestamp, &value) {
		ts = append(ts, &TimeSeriesPoint{
			Timestamp: timestamp,
			Value:     value,
		})
	}
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("strict query processing: %w", err)
	}

	return ts, nil
}

// GetKeyspaces queries the cassandra cluster for a list of an existing keyspaces.
func (s *Session) GetKeyspaces(ctx context.Context) ([]string, error) {
	statement := "SELECT keyspace_name FROM system_schema.keyspaces"
	iter := s.session.Query(statement).WithContext(ctx).Iter()

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
		if column.Type.Type().String() == needType {
			columns = append(columns, name)
		}
	}

	return columns, nil
}

// Ping executes simple query to check connection status.
func (s *Session) Ping(ctx context.Context) error {
	err := s.session.Query("SELECT key FROM system.local").WithContext(ctx).Exec()
	if err != nil {
		return fmt.Errorf("session.Query: %w", err)
	}

	return nil
}

// Close closes connections to cluster.
func (s *Session) Close() {
	s.session.Close()
}

// buildStatement builds cassandra query statement with positional parameters.
func buildStatement(q *Query) string {
	var allowFiltering string
	if q.AllowFiltering {
		allowFiltering = " ALLOW FILTERING"
	}

	statement := fmt.Sprintf(
		"SELECT %s, CAST(%s as double) FROM %s.%s WHERE %s = ? AND %s >= ? AND %s <= ?%s",
		q.ColumnTime,
		q.ColumnValue,
		q.Keyspace,
		q.Table,
		q.ColumnID,
		q.ColumnTime,
		q.ColumnTime,
		allowFiltering,
	)

	return statement
}
