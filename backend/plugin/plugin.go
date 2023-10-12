package plugin

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/backend/cassandra"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

var aliasFormatRegexp = regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)

type repository interface {
	Select(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error)
	GetKeyspaces(ctx context.Context) ([]string, error)
	GetTables(keyspace string) ([]string, error)
	GetColumns(keyspace, table, needType string) ([]string, error)
	Ping(ctx context.Context) error
	Close()
}

// Plugin represents grafana datasource plugin.
type Plugin struct {
	repo repository
}

// New returns configured Plugin.
func New(repo repository) *Plugin {
	return &Plugin{
		repo: repo,
	}
}

// ExecQuery executes metric query based on provided query type.
func (p *Plugin) ExecQuery(ctx context.Context, q *Query) (data.Frames, error) {
	var (
		dataFrames data.Frames
		err        error
	)

	backend.Logger.Debug("ExecQuery", "query", q)
	switch q.RawQuery {
	case true:
		dataFrames, err = p.execRawMetricQuery(ctx, q)
	case false:
		dataFrames, err = p.execStrictMetricQuery(ctx, q)
	}

	if err != nil {
		return nil, fmt.Errorf("query processing: %w", err)
	}

	return dataFrames, nil
}

// execRawMetricQuery executes repository ExecRawQuery method and transforms reposonse to data.Frames.
func (p *Plugin) execRawMetricQuery(ctx context.Context, q *Query) (data.Frames, error) {
	rows, err := p.repo.Select(ctx, q.Target)
	if err != nil {
		return nil, fmt.Errorf("repo.Select: %w", err)
	}

	var frames data.Frames
	for id, points := range rows {
		frame := makeDataFrameFromPoints(id, q.AliasID, points)
		frames = append(frames, frame)
	}

	return frames, nil
}

// execStrictMetricQuery executes repository ExecStrictQuery method and transforms reposonse to data.Frames.
func (p *Plugin) execStrictMetricQuery(ctx context.Context, q *Query) (data.Frames, error) {
	rows, err := p.repo.Select(ctx, q.BuildStatement(), splitIDs(q.ValueID), q.TimeFrom, q.TimeTo)
	if err != nil {
		return nil, fmt.Errorf("repo.ExecStrictQuery: %w", err)
	}

	var frames data.Frames
	for id, points := range rows {
		frame := makeDataFrameFromPoints(id, q.AliasID, points)
		frames = append(frames, frame)
	}

	return frames, nil
}

// GetKeyspaces fetches and returns Cassandra's list of keyspaces.
func (p *Plugin) GetKeyspaces(ctx context.Context) ([]string, error) {
	keyspaces, err := p.repo.GetKeyspaces(ctx)
	if err != nil {
		return nil, fmt.Errorf("repo.GetKeyspaces: %w", err)
	}

	return keyspaces, nil
}

// GetTables fetches and returns Cassandra's list of tables for provided keyspace.
func (p *Plugin) GetTables(keyspace string) ([]string, error) {
	tables, err := p.repo.GetTables(keyspace)
	if err != nil {
		return nil, fmt.Errorf("repo.GetTables: %w", err)
	}

	return tables, nil
}

// GetColumns fetches and returns Cassandra's list of columns of given type for provided keyspace and table.
func (p *Plugin) GetColumns(keyspace, table, needType string) ([]string, error) {
	tables, err := p.repo.GetColumns(keyspace, table, needType)
	if err != nil {
		return nil, fmt.Errorf("repo.GetColumns: %w", err)
	}

	return tables, nil
}

// CheckHealth executes repository Ping method to check database health.
func (p *Plugin) CheckHealth(ctx context.Context) error {
	err := p.repo.Ping(ctx)
	if err != nil {
		return fmt.Errorf("repo.Ping: %w", err)
	}

	return nil
}

// Dispose closes all connections to Cassandra cluster.
func (p *Plugin) Dispose() {
	p.repo.Close()
}

func splitIDs(s string) []string {
	var ids []string
	for _, id := range strings.Split(s, ",") {
		ids = append(ids, strings.TrimSpace(id))
	}

	return ids
}

// makeDataFrameFromPoints creates data frames from time series points returned by repository.
func makeDataFrameFromPoints(id string, alias string, rows []cassandra.Row) *data.Frame {
	if len(rows) == 0 {
		return nil
	}

	frame := data.NewFrame(id, nil)

	alias = formatAlias(alias, rows[0].Fields)
	fields := make([]*data.Field, 0, len(rows[0].Columns))
	for _, colName := range rows[0].Columns {
		field := data.NewFieldFromFieldType(data.FieldTypeFor(rows[0].Fields[colName]), 0)
		field.Name = colName
		if alias != "" && field.Type().Numeric() {
			field.SetConfig(&data.FieldConfig{DisplayNameFromDS: alias})
		}
		fields = append(fields, field)
	}
	frame.Fields = fields

	for _, r := range rows {
		values := make([]interface{}, 0, len(r.Fields))
		for _, colName := range r.Columns {
			values = append(values, r.Fields[colName])
		}
		frame.AppendRow(values...)
	}

	return frame
}

func formatAlias(alias string, values map[string]interface{}) string {
	formattedAlias := aliasFormatRegexp.ReplaceAllFunc([]byte(alias), func(in []byte) []byte {
		fieldName := strings.Replace(string(in), "{{", "", 1)
		fieldName = strings.Replace(fieldName, "}}", "", 1)
		fieldName = strings.TrimSpace(fieldName)
		if val, exists := values[fieldName]; exists {
			switch v := val.(type) {
			case string:
				return []byte(v)
			case int8, int32, int64, int:
				return []byte(fmt.Sprintf("%d", v))
			case float32, float64:
				return []byte(fmt.Sprintf("%f", v))
			case bool:
				return []byte(fmt.Sprintf("%t", v))
			case time.Time:
				return []byte(v.String())
			default:
				return []byte{}
			}
		}
		return []byte{}
	})

	return string(formattedAlias)
}
