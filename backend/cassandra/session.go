package cassandra

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

// Settings is a set of Cassandra session settings.
type Settings struct {
	Hosts       []string
	Keyspace    string
	User        string
	Password    string
	Consistency string
	Timeout     *int
	TLSConfig   *tls.Config
}

// Session is a convenience wrapper for the gocql.Session.
type Session struct {
	session *gocql.Session
}

// New creates a new cassandra cluster session using provided settings.
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

// Select queries the database with provided query string and returns result rows grouped by ID.
// ID must be a first requested column in query and must be convertable to a string.
func (s *Session) Select(ctx context.Context, query string, values ...interface{}) (rows map[string][]Row, err error) {
	if !isSelect(query) {
		return nil, fmt.Errorf("query is not a SELECT statement: %s", query)
	}

	iter := s.session.Query(query, values...).WithContext(ctx).Iter()
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

		// id used to distinguish different timeseries, so it must have string type.
		// Here we are trying to convert id column value to a string or exit early
		// in case when such conversion is not supported.
		idFieldName := iter.Columns()[0].Name
		id, err := parseIDField(rowValues[idFieldName])
		if err != nil {
			return nil, fmt.Errorf("row processing: %w", err)
		}

		row := Row{
			Columns: columnNames(iter.Columns()),
			Fields:  rowValues,
		}
		row.filterUnsupportedTypes()
		rows[id] = append(rows[id], row)
	}

	return rows, nil
}

// GetKeyspaces queries the cassandra cluster for a list of existing keyspaces.
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

// Ping executes a simple query to check the connection status.
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

func parseIDField(val interface{}) (string, error) {
	var id string
	switch v := val.(type) {
	case string:
		id = v
	case gocql.UUID:
		id = v.String()
	case int8, int32, int64, int:
		id = fmt.Sprintf("%d", v)
	case float32, float64:
		id = fmt.Sprintf("%f", v)
	case time.Time:
		id = v.String()
	case bool:
		id = fmt.Sprintf("%t", v)
	default:
		return "", fmt.Errorf("unsupported type: %T", val)
	}

	return id, nil
}

func columnNames(columnInfo []gocql.ColumnInfo) []string {
	names := make([]string, 0, len(columnInfo))
	for _, col := range columnInfo {
		names = append(names, col.Name)
	}

	return names
}

func isSelect(query string) bool {
	stmt := strings.TrimSpace(query)
	if !strings.HasPrefix(strings.ToUpper(stmt), "SELECT ") {
		return false
	}

	return true
}
