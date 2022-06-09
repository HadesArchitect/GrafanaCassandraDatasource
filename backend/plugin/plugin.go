package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/backend/cassandra"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type repository interface {
	ExecRawQuery(ctx context.Context, q *cassandra.Query) (map[string][]*cassandra.TimeSeriesPoint, error)
	ExecStrictQuery(ctx context.Context, q *cassandra.Query) ([]*cassandra.TimeSeriesPoint, error)
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
func (p *Plugin) ExecQuery(ctx context.Context, q *cassandra.Query) (data.Frames, error) {
	var (
		dataFrames data.Frames
		err        error
	)

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
func (p *Plugin) execRawMetricQuery(ctx context.Context, q *cassandra.Query) (data.Frames, error) {
	tsPointsMap, err := p.repo.ExecRawQuery(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repo.ExecRawQuery: %w", err)
	}

	var frames data.Frames
	for id, points := range tsPointsMap {
		frame := makeDataFrameFromPoints(id, points)
		frames = append(frames, frame)
	}

	return frames, nil
}

// execStrictMetricQuery executes repository ExecStrictQuery method and transforms reposonse to data.Frames.
func (p *Plugin) execStrictMetricQuery(ctx context.Context, q *cassandra.Query) (data.Frames, error) {
	tsPoints, err := p.repo.ExecStrictQuery(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("repo.ExecStrictQuery: %w", err)
	}

	id := q.AliasID
	if id == "" {
		id = q.ValueID
	}

	frame := makeDataFrameFromPoints(id, tsPoints)

	return data.Frames{frame}, nil
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

// makeDataFrameFromPoints creates data frames from time series points returned by repository.
func makeDataFrameFromPoints(id string, points []*cassandra.TimeSeriesPoint) *data.Frame {
	timeField := data.NewField("time", nil, make([]time.Time, 0, len(points)))
	valueField := data.NewField(id, nil, make([]float64, 0, len(points)))
	valueField.Config = &data.FieldConfig{DisplayNameFromDS: id}

	frame := data.NewFrame(id, timeField, valueField)
	for _, p := range points {
		frame.AppendRow(p.Timestamp, p.Value)
	}

	return frame
}
