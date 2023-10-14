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

	return makeDataFrames(q, rows), nil
}

// execStrictMetricQuery executes repository ExecStrictQuery method and transforms reposonse to data.Frames.
func (p *Plugin) execStrictMetricQuery(ctx context.Context, q *Query) (data.Frames, error) {
	rows, err := p.repo.Select(ctx, q.BuildStatement(), splitIDs(q.ValueID), q.TimeFrom, q.TimeTo)
	if err != nil {
		return nil, fmt.Errorf("repo.ExecStrictQuery: %w", err)
	}

	return makeDataFrames(q, rows), nil
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

func makeDataFrames(q *Query, rows map[string][]cassandra.Row) data.Frames {
	var frames data.Frames
	for id, points := range rows {
		frame := makeDataFrameFromRows(id, q.AliasID, points)
		if q.IsAlertQuery {
			// alerting doesn't support narrow frames
			frame = narrowFrameToWideFrame(frame)
		}
		frames = append(frames, frame)
	}

	return frames
}

// makeDataFrameFromRows creates data frames from time series points returned by repository.
func makeDataFrameFromRows(id string, alias string, rows []cassandra.Row) *data.Frame {
	if len(rows) == 0 {
		return nil
	}

	frame := data.NewFrame(id, nil)

	// use first of the returned rows to interpolate legend alias.
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

// formatAlias performs legend alies interpolation.
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

// narrowFrameToWideFrame performs rudimentary frames conversion from narrow to wide format.
// It puts non-TS fields to labels and removes from fields list. Conflicting labels are replaced.
// Any other field is ignored and could cause grafana alerting error during alert query execution.
// https://grafana.com/developers/plugin-tools/introduction/data-frames#data-frames-as-time-series
func narrowFrameToWideFrame(frame *data.Frame) *data.Frame {
	if len(frame.Fields) == 0 {
		return frame
	}

	labels := makeLabelsFromNonTSFields(frame)
	for i := 0; i < len(frame.Fields); i++ {
		if frame.Fields[i].Type().Numeric() {
			if frame.Fields[i].Labels == nil {
				frame.Fields[i].Labels = make(map[string]string, len(labels))
			}
			for k, v := range labels {
				frame.Fields[i].Labels[k] = v
			}
		}
	}

	return removeNonTSFields(frame)
}

// makeLabelsFromNonTSFields creates map of labels and their corresponding
// values from all fields that are not numeric or timestamps. Only first
// value from each field used as label value.
func makeLabelsFromNonTSFields(frame *data.Frame) map[string]string {
	labels := make(map[string]string)
	if frame == nil || len(frame.Fields) == 0 || frame.Fields[0].Len() == 0 {
		return labels
	}

	for _, f := range frame.Fields {
		if !f.Type().Numeric() && !f.Type().Time() {
			labels[f.Name] = fmt.Sprintf("%v", f.CopyAt(0))
		}
	}

	return labels
}

// removeNonTSFields deletes all fields that are not numeric or timestamps
// from frame. These values should be previously stored to labels
// using makeLabelsFromNonTSFields method. Result is not stable, e.g.
// elements can change their positions during filtration.
func removeNonTSFields(frame *data.Frame) *data.Frame {
	if frame == nil {
		return nil
	}

	i, j := 0, len(frame.Fields)-1
	for i <= j {
		// keep numeric and timestamp fields
		if frame.Fields[i].Type().Numeric() || frame.Fields[i].Type().Time() {
			i++
			continue
		}
		// move all the other fields to the end of the Fields slice
		frame.Fields[i], frame.Fields[j] = frame.Fields[j], frame.Fields[i]
		j--
	}

	frame.Fields = frame.Fields[0:i]

	return frame
}
