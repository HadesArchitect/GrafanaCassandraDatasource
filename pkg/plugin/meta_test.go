package plugin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const truncatedText = "Result truncated to 3 rows (Max rows setting). " +
	"Narrow the time range or the query, or raise Max rows in the data source settings."

func sensorRow(id, location string, value float64, ms int64) cassandra.Row {
	return cassandra.Row{
		Columns: []string{"sensor_id", "location", "value", "time"},
		Fields: map[string]interface{}{
			"sensor_id": id, "location": location, "value": value, "time": time.UnixMilli(ms).UTC(),
		},
	}
}

// selectRecorder is a repository whose Select returns result and records
// what it was called with.
type selectRecorder struct {
	result  cassandra.SelectResult
	err     error
	maxRows int
	query   string
	values  []interface{}
	calls   int
}

func (r *selectRecorder) repo() *repositoryMock {
	return &repositoryMock{
		onSelectResult: func(_ context.Context, maxRows int, query string, values ...interface{}) (cassandra.SelectResult, error) {
			r.calls++
			r.maxRows, r.query, r.values = maxRows, query, values
			return r.result, r.err
		},
	}
}

// statsOf returns the statistics of all frames by name, and how many frames carry them.
func statsOf(t *testing.T, frames data.Frames) (map[string]data.QueryStat, int) {
	t.Helper()
	stats := map[string]data.QueryStat{}
	carriers := 0
	for _, frame := range frames {
		require.NotNil(t, frame.Meta)
		if len(frame.Meta.Stats) > 0 {
			carriers++
		}
		for _, stat := range frame.Meta.Stats {
			stats[stat.DisplayName] = stat
		}
	}
	return stats, carriers
}

func twoSeriesResult(truncated bool) cassandra.SelectResult {
	return cassandra.SelectResult{
		Rows: map[string][]cassandra.Row{
			"a": {sensorRow("a", "kitchen", 1.5, 1000), sensorRow("a", "kitchen", 2.5, 2000)},
			"b": {sensorRow("b", "garage", 3.5, 1000)},
		},
		RowCount:  3,
		Truncated: truncated,
	}
}

func TestPlugin_ExecQuery_meta(t *testing.T) {
	t.Run("raw query: executed CQL, stats on one frame, no notice", func(t *testing.T) {
		rec := &selectRecorder{result: twoSeriesResult(false)}
		p := New(rec.repo(), Options{MaxRows: 3})
		cql := "SELECT sensor_id, location, value, time FROM test.test WHERE time > 1257894000000"

		frames, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: cql})
		require.NoError(t, err)
		require.Len(t, frames, 2)

		assert.Equal(t, 3, rec.maxRows, "the Max rows setting is passed to Select")
		assert.Equal(t, cql, rec.query)
		assert.Empty(t, rec.values)

		for _, frame := range frames {
			require.NotNil(t, frame.Meta)
			assert.Equal(t, cql, frame.Meta.ExecutedQueryString)
			assert.Empty(t, frame.Meta.Notices)
		}

		stats, carriers := statsOf(t, frames)
		assert.Equal(t, 1, carriers, "the inspector would list the stats of every frame")
		assert.Equal(t, 3.0, stats["Rows returned"].Value)
		require.Contains(t, stats, "Execution time")
		assert.Equal(t, "ms", stats["Execution time"].Unit)
		assert.GreaterOrEqual(t, stats["Execution time"].Value, 0.0)
	})

	t.Run("truncated result: warning notice on every frame", func(t *testing.T) {
		rec := &selectRecorder{result: twoSeriesResult(true)}
		p := New(rec.repo(), Options{MaxRows: 3})

		frames, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: "SELECT * FROM test.test"})
		require.NoError(t, err)
		require.Len(t, frames, 2)

		for _, frame := range frames {
			require.NotNil(t, frame.Meta)
			assert.Equal(t, []data.Notice{{Severity: data.NoticeSeverityWarning, Text: truncatedText}}, frame.Meta.Notices)
		}
		stats, _ := statsOf(t, frames)
		assert.Equal(t, 3.0, stats["Rows returned"].Value)
	})

	t.Run("configurator query: built statement and bound values", func(t *testing.T) {
		rec := &selectRecorder{result: twoSeriesResult(false)}
		p := New(rec.repo(), Options{MaxRows: 100000})
		from := time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)
		to := from.Add(90 * time.Minute)
		q := &Query{
			Keyspace: "test", Table: "test",
			ColumnID: "sensor_id", ColumnValue: "value", ColumnTime: "time",
			ValueID: "a, b", TimeFrom: from, TimeTo: to,
		}

		frames, err := p.ExecQuery(context.Background(), q)
		require.NoError(t, err)

		statement := "SELECT sensor_id, value, time FROM test.test WHERE sensor_id IN ? AND time >= ? AND time <= ?"
		assert.Equal(t, statement, rec.query)
		assert.Equal(t, []interface{}{[]string{"a", "b"}, from, to}, rec.values)
		assert.Equal(t, 100000, rec.maxRows)

		want := statement + "\n-- bound values: sensor_id IN (a, b), time >= 2009-11-10T23:00:00.000Z, time <= 2009-11-11T00:30:00.000Z"
		for _, frame := range frames {
			assert.Equal(t, want, frame.Meta.ExecutedQueryString)
		}
	})

	t.Run("no rows: an empty frame carries the metadata", func(t *testing.T) {
		rec := &selectRecorder{result: cassandra.SelectResult{Rows: map[string][]cassandra.Row{}}}
		p := New(rec.repo(), Options{MaxRows: 3})

		frames, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: "SELECT * FROM test.test"})
		require.NoError(t, err)
		require.Len(t, frames, 1)
		assert.Empty(t, frames[0].Fields)
		assert.Equal(t, "SELECT * FROM test.test", frames[0].Meta.ExecutedQueryString)

		stats, carriers := statsOf(t, frames)
		assert.Equal(t, 1, carriers)
		assert.Equal(t, 0.0, stats["Rows returned"].Value)
	})

	t.Run("no rows in an alert query: no frames, as before", func(t *testing.T) {
		rec := &selectRecorder{result: cassandra.SelectResult{Rows: map[string][]cassandra.Row{}}}
		p := New(rec.repo(), Options{MaxRows: 3})

		frames, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: "SELECT * FROM test.test", IsAlertQuery: true})
		require.NoError(t, err)
		assert.Empty(t, frames)
	})

	t.Run("truncated alert query: wide frames with labels and metadata", func(t *testing.T) {
		rec := &selectRecorder{result: twoSeriesResult(true)}
		p := New(rec.repo(), Options{MaxRows: 3})

		frames, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: "SELECT * FROM test.test", IsAlertQuery: true})
		require.NoError(t, err)
		require.Len(t, frames, 2)

		locations := map[string]bool{}
		for _, frame := range frames {
			// Only the time and value fields are left, the text columns are labels.
			require.Len(t, frame.Fields, 2)
			for _, field := range frame.Fields {
				assert.True(t, field.Type().Numeric() || field.Type().Time(), "field %s", field.Name)
				if field.Type().Numeric() {
					locations[field.Labels["location"]] = true
				}
			}
			require.NotNil(t, frame.Meta)
			assert.Equal(t, []data.Notice{{Severity: data.NoticeSeverityWarning, Text: truncatedText}}, frame.Meta.Notices)
		}
		assert.Equal(t, map[string]bool{"kitchen": true, "garage": true}, locations)

		// The frames, with their metadata, are sent to Grafana as Arrow.
		_, err = frames.MarshalArrow()
		assert.NoError(t, err)
	})

	t.Run("select error", func(t *testing.T) {
		rec := &selectRecorder{err: errors.New("boom")}
		p := New(rec.repo(), Options{MaxRows: 3})

		frames, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: "SELECT * FROM test.test"})
		assert.ErrorContains(t, err, "boom")
		assert.Nil(t, frames)
	})

	t.Run("unlimited", func(t *testing.T) {
		rec := &selectRecorder{result: twoSeriesResult(false)}
		p := New(rec.repo(), Options{MaxRows: 0})

		_, err := p.ExecQuery(context.Background(), &Query{RawQuery: true, Target: "SELECT * FROM test.test"})
		require.NoError(t, err)
		assert.Equal(t, 0, rec.maxRows)
	})
}

func TestPlugin_GetVariables_maxRows(t *testing.T) {
	rec := &selectRecorder{result: cassandra.SelectResult{
		Rows: map[string][]cassandra.Row{
			"1": {{Columns: []string{"id"}, Fields: map[string]interface{}{"id": "1"}}},
			"2": {{Columns: []string{"id"}, Fields: map[string]interface{}{"id": "2"}}},
		},
		RowCount:  2,
		Truncated: true,
	}}
	p := New(rec.repo(), Options{MaxRows: 2})

	vars, err := p.GetVariables(context.Background(), "SELECT id FROM test.test")
	require.NoError(t, err)

	assert.Equal(t, 2, rec.maxRows, "variable queries use the same row limit")
	assert.ElementsMatch(t, []Variable{{Value: "1", Label: "1"}, {Value: "2", Label: "2"}}, vars)
}
