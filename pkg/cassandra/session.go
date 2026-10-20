package cassandra

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
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
	session  *gocql.Session
	connects *connectRecorder
}

// New creates a new cassandra cluster session using provided settings.
//
// Errors caused by the settings are tagged with a health.Category, and a
// failed session creation carries the error of the last failed connection
// attempt, so that the cause can be classified even though the driver
// flattens it into a string.
func New(cfg Settings) (*Session, error) {
	cluster, err := newClusterConfig(cfg)
	if err != nil {
		return nil, err
	}

	return createSession(cluster)
}

// createSession connects to the cluster, recording connection attempts to
// explain a failure.
func createSession(cluster *gocql.ClusterConfig) (*Session, error) {
	connects := &connectRecorder{}
	cluster.ConnectObserver = connects

	clusterSession, err := cluster.CreateSession()
	if err != nil {
		if last := connects.lastError(); last != nil {
			return nil, fmt.Errorf("cluster.CreateSession: %w (last connection attempt: %w)", err, last)
		}
		// Connections that succeed but cannot switch to the default keyspace
		// are not reported to the connect observer, and leave the pool empty.
		if errors.Is(err, gocql.ErrNoConnectionsStarted) && cluster.Keyspace != "" {
			if exists, probeErr := probeKeyspace(*cluster); probeErr == nil && !exists {
				return nil, fmt.Errorf("cluster.CreateSession: %w: %w", err, health.ErrKeyspaceNotFound)
			}
		}
		return nil, fmt.Errorf("cluster.CreateSession: %w", err)
	}

	return &Session{session: clusterSession, connects: connects}, nil
}

// connectRecorder is a gocql.ConnectObserver that keeps the outcome of the
// latest connection attempt. The driver reports why a connection failed only
// to its log and to connect observers: session creation returns a summary
// with the original error flattened into text.
type connectRecorder struct {
	mu   sync.Mutex
	last error
}

// ObserveConnect implements gocql.ConnectObserver.
func (r *connectRecorder) ObserveConnect(c gocql.ObservedConnect) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.last = c.Err
}

// lastError returns the error of the latest connection attempt, or nil if it succeeded.
func (r *connectRecorder) lastError() error {
	if r == nil {
		return nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.last
}

// probeTimeout bounds each optional query made to explain or enrich a health check.
const probeTimeout = 5 * time.Second

// probeKeyspace checks whether the keyspace configured in cluster exists, using
// a short-lived session without a default keyspace. It is only used to explain
// why a session could not be created.
func probeKeyspace(cluster gocql.ClusterConfig) (bool, error) {
	keyspace := cluster.Keyspace
	cluster.Keyspace = ""
	cluster.NumConns = 1
	cluster.ConnectObserver = nil

	session, err := cluster.CreateSession()
	if err != nil {
		return false, fmt.Errorf("cluster.CreateSession: %w", err)
	}
	defer session.Close()

	ctx, cancel := context.WithTimeout(context.Background(), probeTimeout)
	defer cancel()

	return keyspaceExists(ctx, session, keyspace)
}

func keyspaceExists(ctx context.Context, session *gocql.Session, keyspace string) (bool, error) {
	iter := session.Query("SELECT keyspace_name FROM system_schema.keyspaces WHERE keyspace_name = ?", keyspace).
		Consistency(gocql.LocalOne).IterContext(ctx)
	found := iter.NumRows() > 0
	if err := iter.Close(); err != nil {
		return false, fmt.Errorf("query processing: %w", err)
	}

	return found, nil
}

// newClusterConfig translates the plugin settings into a driver cluster config.
func newClusterConfig(cfg Settings) (*gocql.ClusterConfig, error) {
	hosts := make([]string, 0, len(cfg.Hosts))
	for _, host := range cfg.Hosts {
		if host = strings.TrimSpace(host); host != "" {
			hosts = append(hosts, host)
		}
	}
	if len(hosts) == 0 {
		return nil, health.Wrap(health.InvalidContactPoints, gocql.ErrNoHosts)
	}

	cluster := gocql.NewCluster(hosts...)
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
		return nil, health.Wrap(health.InvalidConsistency, fmt.Errorf("gocql.ParseConsistencyWrapper: %w", err))
	}
	cluster.Consistency = consistencyLevel

	if cfg.Timeout != nil {
		// gocql treats a timeout of zero or less as no timeout at all, so a
		// negative value would silently remove the bound it was meant to set.
		if *cfg.Timeout < 0 {
			return nil, health.Wrap(health.InvalidTimeout, fmt.Errorf("negative timeout: %d", *cfg.Timeout))
		}
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

// SelectResult holds the rows read by Select.
type SelectResult struct {
	// Rows are the rows grouped by ID, the value of the first column.
	Rows map[string][]Row
	// RowCount is the number of rows in Rows, across all IDs.
	RowCount int
	// Truncated is true when the query had more rows than the limit passed
	// to Select. Rows then holds the first rows only, up to the limit.
	Truncated bool
}

// defaultPageSize is the driver's default page size (gocql.ClusterConfig.PageSize).
const defaultPageSize = 5000

// Select queries the database with provided query string and returns result rows grouped by ID.
// ID must be a first requested column in query and must be convertable to a string.
//
// At most maxRows rows are read, counted across all IDs. When the query has
// more, reading stops there, the remaining pages are not fetched, and the
// result is marked as truncated. A maxRows of zero or less reads every row.
func (s *Session) Select(ctx context.Context, maxRows int, query string, values ...interface{}) (result SelectResult, err error) {
	if !isSelect(query) {
		return SelectResult{}, fmt.Errorf("query is not a SELECT statement: %s", query)
	}

	q := s.session.Query(query, values...)
	// A small limit does not need the driver's default page of 5000 rows.
	// The extra row tells whether the result goes beyond the limit.
	if maxRows > 0 && maxRows < defaultPageSize {
		q = q.PageSize(maxRows + 1)
	}

	iter := q.IterContext(ctx)
	defer func() {
		// Closing an iterator before its last page is not an error: the
		// pages that were not read are simply never fetched.
		if iterErr := iter.Close(); iterErr != nil {
			result = SelectResult{}
			err = fmt.Errorf("select query processing: %w", iterErr)
		}
	}()

	return scanRows(iter, maxRows)
}

// rowScanner is the part of *gocql.Iter that scanRows uses.
type rowScanner interface {
	Columns() []gocql.ColumnInfo
	MapScan(m map[string]interface{}) bool
}

// scanRows reads up to maxRows rows from iter, all of them when maxRows is
// zero or less, and groups them by the value of their first column. Errors
// of the iterator itself are reported by its Close method.
func scanRows(iter rowScanner, maxRows int) (SelectResult, error) {
	result := SelectResult{Rows: make(map[string][]Row)}
	for {
		if maxRows > 0 && result.RowCount >= maxRows {
			// Reading one more row, which is discarded, tells a result of
			// exactly maxRows rows apart from a larger one.
			result.Truncated = iter.MapScan(make(map[string]interface{}, len(iter.Columns())))
			break
		}

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
			return SelectResult{}, fmt.Errorf("row processing: %w", err)
		}

		row := Row{
			Columns: columnNames(iter.Columns()),
			Fields:  rowValues,
		}
		if err := row.normalize(); err != nil {
			return SelectResult{}, fmt.Errorf("row.normalize: %w", err)
		}
		result.Rows[id] = append(result.Rows[id], row)
		result.RowCount++
	}

	return result, nil
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
		// When no node can be reached, why is only known from the pool's
		// latest reconnection attempt.
		if errors.Is(err, gocql.ErrNoConnections) {
			if last := s.connects.lastError(); last != nil {
				return fmt.Errorf("session.Query: %w (last connection attempt: %w)", err, last)
			}
		}
		return fmt.Errorf("session.Query: %w", err)
	}

	return nil
}

// ClusterInfo holds facts about the node the health check talked to. Every
// field may be empty: AWS Keyspaces, Astra and other CQL-compatible services
// return a limited or emulated system.local.
type ClusterInfo struct {
	ClusterName    string
	ReleaseVersion string
	// DSEVersion is only set by DataStax Enterprise.
	DSEVersion string
	DataCenter string
}

// systemLocalColumns are the system.local columns ClusterInfo is read from.
var systemLocalColumns = map[string]func(info *ClusterInfo) *string{
	"cluster_name":    func(info *ClusterInfo) *string { return &info.ClusterName },
	"release_version": func(info *ClusterInfo) *string { return &info.ReleaseVersion },
	"dse_version":     func(info *ClusterInfo) *string { return &info.DSEVersion },
	"data_center":     func(info *ClusterInfo) *string { return &info.DataCenter },
}

// ClusterInfo reads cluster facts from system.local in a single query.
//
// The query selects every column, because which ones exist depends on the
// server, and only decodes the text columns listed in systemLocalColumns.
// Every other column, whatever its type, is skipped without being decoded.
func (s *Session) ClusterInfo(ctx context.Context) (ClusterInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	var info ClusterInfo
	iter := s.session.Query("SELECT * FROM system.local").Consistency(gocql.LocalOne).IterContext(ctx)

	columns := iter.Columns()
	dest := make([]interface{}, len(columns))
	for i, column := range columns {
		field, known := systemLocalColumns[column.Name]
		if known && isTextColumn(column) {
			dest[i] = field(&info)
		} else {
			dest[i] = skipColumn{}
		}
	}
	iter.Scan(dest...)

	if err := iter.Close(); err != nil {
		return ClusterInfo{}, fmt.Errorf("query processing: %w", err)
	}

	return info, nil
}

func isTextColumn(column gocql.ColumnInfo) bool {
	if column.TypeInfo == nil {
		return false
	}

	switch column.TypeInfo.Type() {
	case gocql.TypeVarchar, gocql.TypeText, gocql.TypeAscii:
		return true
	default:
		return false
	}
}

// skipColumn is a scan destination that discards the column.
type skipColumn struct{}

// UnmarshalCQL implements gocql.Unmarshaler.
func (skipColumn) UnmarshalCQL(gocql.TypeInfo, []byte) error { return nil }

// KeyspaceExists reports whether the keyspace exists.
func (s *Session) KeyspaceExists(ctx context.Context, keyspace string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	return keyspaceExists(ctx, s.session, keyspace)
}

// CanReadSystemAuth reports whether the session's role can read
// system_auth.roles, which holds the password hash of every role. A role used
// by Grafana should not be able to.
//
// Only a successful read returns true. Any error, an authorization error as
// well as system_auth not existing (AWS Keyspaces, Astra), means the role
// cannot read it. The rows read are discarded without being decoded or logged.
func (s *Session) CanReadSystemAuth(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, probeTimeout)
	defer cancel()

	err := s.session.Query("SELECT role FROM system_auth.roles LIMIT 1").Consistency(gocql.LocalOne).ExecContext(ctx)
	if err != nil {
		return false, fmt.Errorf("session.Query: %w", err)
	}

	return true, nil
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
