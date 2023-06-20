package plugin

import (
	"context"
	"testing"
	"time"

	"local_package/cassandra"

	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

type repositoryMock struct {
	onExecRawQuery    func(ctx context.Context, q *cassandra.Query) (map[string][]*cassandra.TimeSeriesPoint, error)
	onExecStrictQuery func(ctx context.Context, q *cassandra.Query) ([]*cassandra.TimeSeriesPoint, error)
	onGetKeyspaces    func(ctx context.Context) ([]string, error)
	onGetTables       func(keyspace string) ([]string, error)
	onGetColumns      func(keyspace, table, needType string) ([]string, error)
}

func (m *repositoryMock) ExecRawQuery(ctx context.Context, q *cassandra.Query) (map[string][]*cassandra.TimeSeriesPoint, error) {
	return m.onExecRawQuery(ctx, q)
}

func (m *repositoryMock) ExecStrictQuery(ctx context.Context, q *cassandra.Query) ([]*cassandra.TimeSeriesPoint, error) {
	return m.onExecStrictQuery(ctx, q)
}

func (m *repositoryMock) GetKeyspaces(ctx context.Context) ([]string, error) {
	return m.onGetKeyspaces(ctx)
}

func (m *repositoryMock) GetTables(keyspace string) ([]string, error) {
	return m.onGetTables(keyspace)
}

func (m *repositoryMock) GetColumns(keyspace, table, needType string) ([]string, error) {
	return m.onGetColumns(keyspace, table, needType)
}

func (m *repositoryMock) Ping(_ context.Context) error { return nil }

func (m *repositoryMock) Close() {}

func TestPlugin_ExecQuery(t *testing.T) {
	testCases := []struct {
		name  string
		repo  *repositoryMock
		query *cassandra.Query
		want  data.Frames
	}{
		{
			name: "Raw Query",
			repo: &repositoryMock{
				onExecRawQuery: func(ctx context.Context, q *cassandra.Query) (map[string][]*cassandra.TimeSeriesPoint, error) {
					return map[string][]*cassandra.TimeSeriesPoint{
						"1": {
							{Timestamp: time.Unix(1257894000, 0), Value: 3.141, Longitude: "", Latitude: ""},
							{Timestamp: time.Unix(1257894001, 0), Value: 6.283, Longitude: "", Latitude: ""},
							{Timestamp: time.Unix(1257894002, 0), Value: 2.718, Longitude: "", Latitude: ""},
							{Timestamp: time.Unix(1257894003, 0), Value: 1.618, Longitude: "", Latitude: ""},
						},
						"2": {
							{Timestamp: time.Unix(1257894000, 0), Value: 3.142, Longitude: "", Latitude: ""},
							{Timestamp: time.Unix(1257894001, 0), Value: 6.284, Longitude: "", Latitude: ""},
							{Timestamp: time.Unix(1257894002, 0), Value: 2.719, Longitude: "", Latitude: ""},
							{Timestamp: time.Unix(1257894003, 0), Value: 1.619, Longitude: "", Latitude: ""},
						},
					}, nil
				},
			},
			query: &cassandra.Query{
				RawQuery: true,
				Target:   "SELECT Time, CAST(Value as double), Longitude, Latitude FROM Keyspace.Table WHERE ID IN (1, 2) AND Time >= 1257894000 AND Time <= 1257894003",
			},
			want: data.Frames{
				{
					Name: "1",
					Fields: []*data.Field{
						data.NewField("time", nil, []time.Time{
							time.Unix(1257894000, 0),
							time.Unix(1257894001, 0),
							time.Unix(1257894002, 0),
							time.Unix(1257894003, 0),
						}),
						data.NewField("1", nil, []float64{
							3.141,
							6.283,
							2.718,
							1.618,
						}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "1"}),
						data.NewField("longitude", nil, []string{
							"",
							"",
							"",
							"",
						}),
						data.NewField("latitude", nil, []string{
							"",
							"",
							"",
							"",
						}),
					},
				},
				{
					Name: "2",
					Fields: []*data.Field{
						data.NewField("time", nil, []time.Time{
							time.Unix(1257894000, 0),
							time.Unix(1257894001, 0),
							time.Unix(1257894002, 0),
							time.Unix(1257894003, 0),
						}),
						data.NewField("2", nil, []float64{
							3.142,
							6.284,
							2.719,
							1.619,
						}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "2"}),
						data.NewField("longitude", nil, []string{
							"",
							"",
							"",
							"",
						}),
						data.NewField("latitude", nil, []string{
							"",
							"",
							"",
							"",
						}),
					},
				},
			},
		},
		{
			name: "Strict Query",
			repo: &repositoryMock{
				onExecStrictQuery: func(ctx context.Context, q *cassandra.Query) ([]*cassandra.TimeSeriesPoint, error) {
					return []*cassandra.TimeSeriesPoint{
						{Timestamp: time.Unix(1257894000, 0), Value: 3.141, Longitude: "", Latitude: ""},
						{Timestamp: time.Unix(1257894001, 0), Value: 6.283, Longitude: "", Latitude: ""},
						{Timestamp: time.Unix(1257894002, 0), Value: 2.718, Longitude: "", Latitude: ""},
						{Timestamp: time.Unix(1257894003, 0), Value: 1.618, Longitude: "", Latitude: ""},
					}, nil
				},
			},
			query: &cassandra.Query{
				RawQuery:    false,
				Keyspace:    "Keyspace",
				Table:       "Table",
				ColumnValue: "Value",
				ColumnID:    "ID",
				ValueID:     "1",
				Longitude:   "Longitude",
				Latitude:    "Latitude",
				ColumnTime:  "Time",
				TimeFrom:    time.Unix(1257894000, 0),
				TimeTo:      time.Unix(1257894003, 0),
			},
			want: data.Frames{
				{
					Name: "1",
					Fields: []*data.Field{
						data.NewField("time", nil, []time.Time{
							time.Unix(1257894000, 0),
							time.Unix(1257894001, 0),
							time.Unix(1257894002, 0),
							time.Unix(1257894003, 0),
						}),
						data.NewField("1", nil, []float64{
							3.141,
							6.283,
							2.718,
							1.618,
						}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "1"}),
						data.NewField("longitude", nil, []string{
							"",
							"",
							"",
							"",
						}),
						data.NewField("latitude", nil, []string{
							"",
							"",
							"",
							"",
						}),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Plugin{repo: tc.repo}
			dataFrames, err := p.ExecQuery(context.TODO(), tc.query)
			if err != nil {
				t.Fatalf("p.ExecQuery: %v", err)
			}

			assert.Equal(t, tc.want, dataFrames)
		})
	}
}

func Test_makeDataFrameFromPoints(t *testing.T) {
	testCases := []struct {
		name   string
		id     string
		points []*cassandra.TimeSeriesPoint
		want   *data.Frame
	}{
		{
			name:   "nil points",
			id:     "test",
			points: nil,
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("time", nil, make([]time.Time, 0)),
					data.NewField("test", nil, make([]float64, 0)).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("longitude", nil, make([]string, 0)),
					data.NewField("latitude", nil, make([]string, 0)),
				},
			},
		},
		{
			name:   "empty points",
			id:     "test",
			points: []*cassandra.TimeSeriesPoint{},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("time", nil, make([]time.Time, 0)),
					data.NewField("test", nil, make([]float64, 0)).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("longitude", nil, make([]string, 0)),
					data.NewField("latitude", nil, make([]string, 0)),
				},
			},
		},
		{
			name: "one point",
			id:   "test",
			points: []*cassandra.TimeSeriesPoint{
				{Timestamp: time.Unix(1257894000, 0), Value: 3.141, Longitude: "", Latitude: ""},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("time", nil, []time.Time{time.Unix(1257894000, 0)}),
					data.NewField("test", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("longitude", nil, make([]string, 0)),
					data.NewField("latitude", nil, make([]string, 0)),
				},
			},
		},
		{
			name: "multi points",
			id:   "test",
			points: []*cassandra.TimeSeriesPoint{
				{Timestamp: time.Unix(1257894000, 0), Value: 3.141, Longitude: "", Latitude: ""},
				{Timestamp: time.Unix(1257894001, 0), Value: 6.283, Longitude: "", Latitude: ""},
				{Timestamp: time.Unix(1257894002, 0), Value: 2.718, Longitude: "", Latitude: ""},
				{Timestamp: time.Unix(1257894003, 0), Value: 1.618, Longitude: "", Latitude: ""},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("time", nil, []time.Time{
						time.Unix(1257894000, 0),
						time.Unix(1257894001, 0),
						time.Unix(1257894002, 0),
						time.Unix(1257894003, 0),
					}),
					data.NewField("test", nil, []float64{
						3.141,
						6.283,
						2.718,
						1.618,
					}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("longitude", nil, []string{
						"",
						"",
						"",
						"",
					}),
					data.NewField("latitude", nil, []string{
						"",
						"",
						"",
						"",
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataFrame := makeDataFrameFromPoints(tc.id, tc.points)
			assert.Equal(t, tc.want, dataFrame)
		})
	}
}
