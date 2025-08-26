package plugin

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

type repositoryMock struct {
	onSelect       func(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error)
	onGetKeyspaces func(ctx context.Context) ([]string, error)
	onGetTables    func(keyspace string) ([]string, error)
	onGetColumns   func(keyspace, table, needType string) ([]string, error)
}

func (m *repositoryMock) Select(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error) {
	return m.onSelect(ctx, query, values...)
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
		query *Query
		want  data.Frames
	}{
		{
			name: "Raw Query",
			repo: &repositoryMock{
				onSelect: func(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error) {
					return map[string][]cassandra.Row{
						"1": {
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 6.283, "Time": time.UnixMilli(1257894001000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 2.718, "Time": time.UnixMilli(1257894002000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 1.618, "Time": time.UnixMilli(1257894003000).UTC()},
							},
						},
						"2": {
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "2", "Value": 3.142, "Time": time.UnixMilli(1257894000000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "2", "Value": 6.284, "Time": time.UnixMilli(1257894001000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "2", "Value": 2.719, "Time": time.UnixMilli(1257894002000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "2", "Value": 1.619, "Time": time.UnixMilli(1257894003000).UTC()},
							},
						},
					}, nil
				},
			},
			query: &Query{
				RawQuery: true,
				Target:   "SELECT ID, Value, Time FROM Keyspace.Table WHERE ID IN (1, 2) AND Time >= 1257894000 AND Time <= 1257894003",
			},
			want: data.Frames{
				{
					Name: "1",
					Fields: []*data.Field{
						data.NewField("ID", nil, []string{"1", "1", "1", "1"}),
						data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
						data.NewField("Time", nil, []time.Time{
							time.UnixMilli(1257894000000).UTC(),
							time.UnixMilli(1257894001000).UTC(),
							time.UnixMilli(1257894002000).UTC(),
							time.UnixMilli(1257894003000).UTC(),
						}),
					},
				},
				{
					Name: "2",
					Fields: []*data.Field{
						data.NewField("ID", nil, []string{"2", "2", "2", "2"}),
						data.NewField("Value", nil, []float64{3.142, 6.284, 2.719, 1.619}),
						data.NewField("Time", nil, []time.Time{
							time.UnixMilli(1257894000000).UTC(),
							time.UnixMilli(1257894001000).UTC(),
							time.UnixMilli(1257894002000).UTC(),
							time.UnixMilli(1257894003000).UTC(),
						}),
					},
				},
			},
		},
		{
			name: "Strict Query",
			repo: &repositoryMock{
				onSelect: func(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error) {
					return map[string][]cassandra.Row{
						"1": {
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 6.283, "Time": time.UnixMilli(1257894001000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 2.718, "Time": time.UnixMilli(1257894002000).UTC()},
							},
							{
								Columns: []string{"ID", "Value", "Time"},
								Fields:  map[string]interface{}{"ID": "1", "Value": 1.618, "Time": time.UnixMilli(1257894003000).UTC()},
							},
						},
					}, nil
				},
			},
			query: &Query{
				RawQuery:    false,
				Keyspace:    "Keyspace",
				Table:       "Table",
				ColumnValue: "Value",
				ColumnID:    "ID",
				ValueID:     "1",
				ColumnTime:  "Time",
				TimeFrom:    time.UnixMilli(1257894000000).UTC(),
				TimeTo:      time.UnixMilli(1257894003000).UTC(),
			},
			want: data.Frames{
				{
					Name: "1",
					Fields: []*data.Field{
						data.NewField("ID", nil, []string{"1", "1", "1", "1"}),
						data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
						data.NewField("Time", nil, []time.Time{
							time.UnixMilli(1257894000000).UTC(),
							time.UnixMilli(1257894001000).UTC(),
							time.UnixMilli(1257894002000).UTC(),
							time.UnixMilli(1257894003000).UTC(),
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
			sort.Slice(dataFrames, func(i, j int) bool {
				return dataFrames[i].Name < dataFrames[j].Name
			})
			sort.Slice(tc.want, func(i, j int) bool {
				return tc.want[i].Name < tc.want[j].Name
			})
			assert.NoError(t, err)
			assert.EqualValues(t, tc.want, dataFrames)
		})
	}
}

func TestPlugin_GetVariables(t *testing.T) {
	testCases := []struct {
		name string
		repo *repositoryMock
		want []Variable
	}{
		{
			name: "response with labels",
			repo: &repositoryMock{
				onSelect: func(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error) {
					return map[string][]cassandra.Row{
						"1": {
							{
								Columns: []string{"Value", "Label"},
								Fields:  map[string]interface{}{"Value": "1", "Label": "Text1"},
							},
						},
						"2": {
							{
								Columns: []string{"Value", "Label"},
								Fields:  map[string]interface{}{"Value": "2", "Label": "Text2"},
							},
						},
					}, nil
				},
			},
			want: []Variable{
				{Value: "1", Label: "Text1"},
				{Value: "2", Label: "Text2"},
			},
		},
		{
			name: "response without labels",
			repo: &repositoryMock{
				onSelect: func(ctx context.Context, query string, values ...interface{}) (rows map[string][]cassandra.Row, err error) {
					return map[string][]cassandra.Row{
						"1": {
							{
								Columns: []string{"Value"},
								Fields:  map[string]interface{}{"Value": "1"},
							},
						},
						"2": {
							{
								Columns: []string{"Value"},
								Fields:  map[string]interface{}{"Value": "2"},
							},
						},
					}, nil
				},
			},
			want: []Variable{
				{Value: "1", Label: "1"},
				{Value: "2", Label: "2"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			p := &Plugin{repo: tc.repo}
			vars, err := p.GetVariables(context.TODO(), "SELECT * FROM keyspace.table")
			assert.NoError(t, err)

			sort.Slice(vars, func(i, j int) bool {
				return vars[i].Value < vars[j].Value
			})
			assert.Equal(t, tc.want, vars)
		})
	}
}

func Test_makeDataFrameFromRows(t *testing.T) {
	testCases := []struct {
		name  string
		id    string
		alias string
		rows  []cassandra.Row
		want  *data.Frame
	}{
		{
			name:  "nil points",
			id:    "test",
			alias: "",
			rows:  nil,
			want:  nil,
		},
		{
			name:  "empty points",
			id:    "test",
			alias: "",
			rows:  []cassandra.Row{},
			want:  nil,
		},
		{
			name:  "one point",
			id:    "test",
			alias: "",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test"}),
					data.NewField("Value", nil, []float64{3.141}),
					data.NewField("Time", nil, []time.Time{time.UnixMilli(1257894000000).UTC()}),
				},
			},
		},
		{
			name:  "one point with string alias",
			id:    "test",
			alias: "alias",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test"}),
					data.NewField("Value", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "alias"}),
					data.NewField("Time", nil, []time.Time{time.UnixMilli(1257894000000).UTC()}),
				},
			},
		},
		{
			name:  "one point with template alias",
			id:    "test",
			alias: "{{ ID }}",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test"}),
					data.NewField("Value", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("Time", nil, []time.Time{time.UnixMilli(1257894000000).UTC()}),
				},
			},
		},
		{
			name:  "one point with additional fields",
			id:    "test",
			alias: "{{ ID }}",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Another Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Another Value": int64(111), "Time": time.UnixMilli(1257894000000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test"}),
					data.NewField("Value", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("Another Value", nil, []int64{111}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test"}),
					data.NewField("Time", nil, []time.Time{time.UnixMilli(1257894000000).UTC()}),
				},
			},
		},
		{
			name:  "multi points",
			id:    "test",
			alias: "",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 6.283, "Time": time.UnixMilli(1257894001000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 2.718, "Time": time.UnixMilli(1257894002000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 1.618, "Time": time.UnixMilli(1257894003000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name:  "multi points with alias",
			id:    "test",
			alias: "alias",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 6.283, "Time": time.UnixMilli(1257894001000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 2.718, "Time": time.UnixMilli(1257894002000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 1.618, "Time": time.UnixMilli(1257894003000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "alias"}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name:  "multi points with template alias",
			id:    "test",
			alias: "{{ ID }} alias",
			rows: []cassandra.Row{
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 3.141, "Time": time.UnixMilli(1257894000000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 6.283, "Time": time.UnixMilli(1257894001000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 2.718, "Time": time.UnixMilli(1257894002000).UTC()},
				},
				{
					Columns: []string{"ID", "Value", "Time"},
					Fields:  map[string]interface{}{"ID": "test", "Value": 1.618, "Time": time.UnixMilli(1257894003000).UTC()},
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "test alias"}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dataFrame := makeDataFrameFromRows(tc.id, tc.alias, tc.rows)
			assert.EqualValues(t, tc.want, dataFrame)
		})
	}
}

func Test_narrowFrameToWideFrame(t *testing.T) {
	testCases := []struct {
		name  string
		input *data.Frame
		want  *data.Frame
	}{
		{
			name:  "empty",
			input: &data.Frame{},
			want:  &data.Frame{},
		},
		{
			name: "multi points without labels",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", map[string]string{"ID": "test"}, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name: "multiple non-TS fields",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
					data.NewField("Location", nil, []string{"room", "room", "room", "room"}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", map[string]string{"ID": "test", "Location": "room"}, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name: "multi points with labels",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", map[string]string{"SomeLabel": "SomeValue"}, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", map[string]string{"SomeLabel": "SomeValue", "ID": "test"}, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name: "multi points with labels conflict",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", map[string]string{"ID": "SomeID"}, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", map[string]string{"ID": "test"}, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frame := narrowFrameToWideFrame(tc.input)
			sort.Slice(frame.Fields, func(i, j int) bool {
				return frame.Fields[i].Name < frame.Fields[j].Name
			})
			sort.Slice(tc.want.Fields, func(i, j int) bool {
				return tc.want.Fields[i].Name < tc.want.Fields[j].Name
			})
			assert.Equal(t, tc.want, frame)
		})
	}
}

func Test_makeLabelsFromNonTSFields(t *testing.T) {
	testCases := []struct {
		name  string
		input *data.Frame
		want  map[string]string
	}{
		{
			name:  "empty",
			input: &data.Frame{},
			want:  map[string]string{},
		},
		{
			name:  "nil",
			input: nil,
			want:  map[string]string{},
		},
		{
			name: "no string fields",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: map[string]string{},
		},
		{
			name: "one string field",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: map[string]string{"ID": "test"},
		},
		{
			name: "one string field with different values",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test1", "test2", "test3", "test4"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: map[string]string{"ID": "test1"},
		},
		{
			name: "multiple string fields, simple",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test1", "test2", "test3", "test4"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
					data.NewField("Location", nil, []string{"room1", "room2", "room3", "room4"}),
				},
			},
			want: map[string]string{"ID": "test1", "Location": "room1"},
		},
		{
			name: "multiple non TS fields",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test1", "test2", "test3", "test4"}),
					data.NewField("TestField1", nil, []string{"testVal1", "testVal2", "testVal3", "testVal4"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("TestField2", nil, []bool{true, false, true, false}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
					data.NewField("Location", nil, []string{"room1", "room2", "room3", "room4"}),
					data.NewField("TestField3", nil, []string{"testVal31", "testVal32", "testVal33", "testVal34"}),
				},
			},
			want: map[string]string{"ID": "test1", "TestField1": "testVal1", "TestField2": "true", "Location": "room1", "TestField3": "testVal31"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			labels := makeLabelsFromNonTSFields(tc.input)
			assert.Equal(t, tc.want, labels)
		})
	}
}

func Test_removeNonTSFields(t *testing.T) {
	testCases := []struct {
		name  string
		input *data.Frame
		want  *data.Frame
	}{
		{
			name:  "empty",
			input: &data.Frame{},
			want:  &data.Frame{},
		},
		{
			name:  "nil",
			input: nil,
			want:  nil,
		},
		{
			name: "no string fields",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name: "one string field",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test", "test", "test", "test"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name: "multiple string fields",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test1", "test2", "test3", "test4"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
					data.NewField("Location", nil, []string{"room1", "room2", "room3", "room4"}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
		{
			name: "multiple non TS fields",
			input: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("ID", nil, []string{"test1", "test2", "test3", "test4"}),
					data.NewField("TestField1", nil, []string{"testVal1", "testVal2", "testVal3", "testVal4"}),
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("TestField2", nil, []bool{true, false, true, false}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
					data.NewField("Location", nil, []string{"room1", "room2", "room3", "room4"}),
					data.NewField("TestField3", nil, []string{"testVal31", "testVal32", "testVal33", "testVal34"}),
				},
			},
			want: &data.Frame{
				Name: "test",
				Fields: []*data.Field{
					data.NewField("Value", nil, []float64{3.141, 6.283, 2.718, 1.618}),
					data.NewField("Time", nil, []time.Time{
						time.UnixMilli(1257894000000).UTC(),
						time.UnixMilli(1257894001000).UTC(),
						time.UnixMilli(1257894002000).UTC(),
						time.UnixMilli(1257894003000).UTC(),
					}),
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			frame := removeNonTSFields(tc.input)
			if frame != nil {
				sort.Slice(frame.Fields, func(i, j int) bool {
					return frame.Fields[i].Name < frame.Fields[j].Name
				})
			}

			if tc.want != nil {
				sort.Slice(tc.want.Fields, func(i, j int) bool {
					return tc.want.Fields[i].Name < tc.want.Fields[j].Name
				})
			}

			assert.Equal(t, tc.want, frame)
		})
	}
}

func Test_formatAlias(t *testing.T) {
	testCases := []struct {
		name   string
		alias  string
		values map[string]interface{}
		want   string
	}{
		{
			name:   "empty alias",
			alias:  "",
			values: map[string]interface{}{"K1": "V1"},
			want:   "",
		},
		{
			name:   "nil values",
			alias:  "{{ K1 }}",
			values: nil,
			want:   "",
		},
		{
			name:   "empty values",
			alias:  "{{ K1 }}",
			values: map[string]interface{}{},
			want:   "",
		},
		{
			name:   "simple string",
			alias:  "K1",
			values: map[string]interface{}{"K1": "V1"},
			want:   "K1",
		},
		{
			name:   "simple template",
			alias:  "{{ K1 }}",
			values: map[string]interface{}{"K1": "V1"},
			want:   "V1",
		},
		{
			name:   "template with two keys",
			alias:  "{{ K1 }}{{ K2 }}",
			values: map[string]interface{}{"K1": "V1", "K2": "V2"},
			want:   "V1V2",
		},
		{
			name:   "template with not existing key",
			alias:  "{{ K1 }}{{ K2 }}",
			values: map[string]interface{}{"K1": "V1", "K3": "V3"},
			want:   "V1",
		},
		{
			name:   "template with keys and strings",
			alias:  "{{ K1 }}:{{ K2 }} ALIAS",
			values: map[string]interface{}{"K1": "V1", "K2": "V2"},
			want:   "V1:V2 ALIAS",
		},
		{
			name:   "template with not existing key and string",
			alias:  "{{ K1 }}:{{ K2 }} ALIAS",
			values: map[string]interface{}{"K1": "V1", "K3": "V3"},
			want:   "V1: ALIAS",
		},
		{
			name:   "simple template with int64",
			alias:  "{{ K1 }}",
			values: map[string]interface{}{"K1": int64(123)},
			want:   "123",
		},
		{
			name:   "simple template with float64",
			alias:  "{{ K1 }}",
			values: map[string]interface{}{"K1": float64(0.1)},
			want:   "0.100000",
		},
		{
			name:   "simple template with bool",
			alias:  "{{ K1 }}",
			values: map[string]interface{}{"K1": true},
			want:   "true",
		},
		{
			name:   "simple template with time",
			alias:  "{{ K1 }}",
			values: map[string]interface{}{"K1": time.UnixMilli(1257894000000).UTC()},
			want:   "2009-11-10 23:00:00 +0000 UTC",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			alias := formatAlias(tc.alias, tc.values)
			assert.Equal(t, tc.want, alias)
		})
	}
}
