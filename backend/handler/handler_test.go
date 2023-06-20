package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local_package/cassandra"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

type instanceManagerMock struct {
	plugin *pluginMock
}

func (i *instanceManagerMock) Get(_ backend.PluginContext) (instancemgmt.Instance, error) {
	return i.plugin, nil
}

func (i *instanceManagerMock) Do(_ backend.PluginContext, _ instancemgmt.InstanceCallbackFunc) error {
	return nil
}

type pluginMock struct {
	onExecQuery    func(ctx context.Context, q *cassandra.Query) (data.Frames, error)
	onGetKeyspaces func(ctx context.Context) ([]string, error)
	onGetTables    func(keyspace string) ([]string, error)
	onGetColumns   func(keyspace, table, needType string) ([]string, error)
	onCheckHealth  func(ctx context.Context) error
	onDispose      func()
}

func (p *pluginMock) ExecQuery(ctx context.Context, q *cassandra.Query) (data.Frames, error) {
	return p.onExecQuery(ctx, q)
}

func (p *pluginMock) GetKeyspaces(ctx context.Context) ([]string, error) {
	return p.onGetKeyspaces(ctx)
}

func (p *pluginMock) GetTables(keyspace string) ([]string, error) {
	return p.onGetTables(keyspace)
}

func (p *pluginMock) GetColumns(keyspace, table, needType string) ([]string, error) {
	return p.onGetColumns(keyspace, table, needType)
}

func (p *pluginMock) CheckHealth(ctx context.Context) error {
	return p.onCheckHealth(ctx)
}

func (p *pluginMock) Dispose() {}

func Test_queryMetricData(t *testing.T) {
	testCases := []struct {
		name    string
		plugin  *pluginMock
		request *backend.QueryDataRequest
		want    *backend.QueryDataResponse
	}{
		{
			name:    "empty",
			plugin:  &pluginMock{},
			request: &backend.QueryDataRequest{},
			want:    &backend.QueryDataResponse{Responses: backend.Responses{}},
		},
		{
			name: "one query",
			plugin: &pluginMock{
				onExecQuery: func(_ context.Context, q *cassandra.Query) (data.Frames, error) {
					return data.Frames{{
						Name: "1",
						Fields: []*data.Field{
							data.NewField("time", nil, []time.Time{time.Unix(1257894000, 0)}),
							data.NewField("1", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "Alias"}),
						}},
					}, nil
				},
			},
			request: &backend.QueryDataRequest{
				Queries: []backend.DataQuery{
					{
						RefID:     "123456789",
						TimeRange: backend.TimeRange{From: time.Unix(1257894000, 0), To: time.Unix(125789403, 0)},
						JSON: []byte(`{"datasourceId": 1, "queryType": "query", "rawQuery": true, "refId": "123456789",
							        "target": "SELECT * from Keyspace.Table", "columnTime": "Time", "columnValue": "Value",
								    "keyspace": "Keyspace", "table": "Table", "columnId": "ID", "valueId": "1", "longitude": "Longitude","latitude": "Latitude",
									"alias": "Alias", "filtering": true}`),
					},
				},
			},
			want: &backend.QueryDataResponse{Responses: backend.Responses{
				"123456789": backend.DataResponse{
					Frames: data.Frames{{
						Name: "1",
						Fields: []*data.Field{
							data.NewField("time", nil, []time.Time{time.Unix(1257894000, 0)}),
							data.NewField("1", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "Alias"}),
						}},
					},
					Error: nil,
				},
			}},
		},
		{
			name: "two queries",
			plugin: &pluginMock{
				onExecQuery: func(_ context.Context, q *cassandra.Query) (data.Frames, error) {
					return data.Frames{{
						Name: q.ValueID,
						Fields: []*data.Field{
							data.NewField("time", nil, []time.Time{time.Unix(1257894000, 0)}),
							data.NewField(q.ValueID, nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: q.AliasID}),
						}},
					}, nil
				},
			},
			request: &backend.QueryDataRequest{
				Queries: []backend.DataQuery{
					{
						RefID:     "123456789",
						TimeRange: backend.TimeRange{From: time.Unix(1257894000, 0), To: time.Unix(125789403, 0)},
						JSON: []byte(`{"datasourceId": 1, "queryType": "query", "rawQuery": true, "refId": "123456789",
							        "target": "SELECT * from Keyspace.Table", "columnTime": "Time", "columnValue": "Value",
								    "keyspace": "Keyspace", "table": "Table", "columnId": "ID", "valueId": "1", "longitude": "Longitude","latitude": "Latitude",
									"alias": "Alias1", "filtering": true}`),
					},
					{
						RefID:     "223456789",
						TimeRange: backend.TimeRange{From: time.Unix(1257894000, 0), To: time.Unix(125789403, 0)},
						JSON: []byte(`{"datasourceId": 1, "queryType": "query", "rawQuery": true, "refId": "223456789",
							        "target": "SELECT * from Keyspace.Table", "columnTime": "Time", "columnValue": "Value",
								    "keyspace": "Keyspace", "table": "Table", "columnId": "ID", "valueId": "2", "longitude": "Longitude","latitude": "Latitude",
									"alias": "Alias2", "filtering": false}`),
					},
				},
			},
			want: &backend.QueryDataResponse{Responses: backend.Responses{
				"123456789": backend.DataResponse{
					Frames: data.Frames{{
						Name: "1",
						Fields: []*data.Field{
							data.NewField("time", nil, []time.Time{time.Unix(1257894000, 0)}),
							data.NewField("1", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "Alias1"}),
						}},
					},
					Error: nil,
				},
				"223456789": backend.DataResponse{
					Frames: data.Frames{{
						Name: "2",
						Fields: []*data.Field{
							data.NewField("time", nil, []time.Time{time.Unix(1257894000, 0)}),
							data.NewField("2", nil, []float64{3.141}).SetConfig(&data.FieldConfig{DisplayNameFromDS: "Alias2"}),
						}},
					},
					Error: nil,
				},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := &handler{instanceManager: &instanceManagerMock{plugin: tc.plugin}}
			result, err := h.queryMetricData(context.TODO(), tc.request)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, tc.want, result)
		})
	}
}

func Test_CheckHealth(t *testing.T) {
	testCases := []struct {
		name   string
		plugin *pluginMock
		want   *backend.CheckHealthResult
	}{
		{
			name: "no error",
			plugin: &pluginMock{
				onCheckHealth: func(_ context.Context) error {
					return nil
				},
			},
			want: &backend.CheckHealthResult{
				Status:  backend.HealthStatusOk,
				Message: "Connected",
			},
		},
		{
			name: "error",
			plugin: &pluginMock{
				onCheckHealth: func(_ context.Context) error {
					return errors.New("some error")
				},
			},
			want: &backend.CheckHealthResult{
				Status:  backend.HealthStatusError,
				Message: "Error, check Grafana logs for more details",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			h := handler{instanceManager: &instanceManagerMock{plugin: tc.plugin}}
			result, err := h.CheckHealth(context.TODO(), &backend.CheckHealthRequest{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			assert.Equal(t, result, tc.want)
		})
	}
}

func Test_writeHTTPResult(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		status int
		want   string
	}{
		{
			name:   "nil",
			input:  nil,
			status: http.StatusOK,
			want:   "null",
		},
		{
			name:   "empty",
			input:  []string{},
			status: http.StatusOK,
			want:   "[]",
		},
		{
			name:   "one element",
			input:  []string{"ONE"},
			status: http.StatusOK,
			want: `[
    "ONE"
]`,
		},
		{
			name:   "multiple elements",
			input:  []string{"ONE", "TWO", "THREE"},
			status: http.StatusOK,
			want: `[
    "ONE",
    "TWO",
    "THREE"
]`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writeHTTPResult(recorder, tc.input)

			assert.Equal(t, recorder.Code, tc.status)
			assert.Equal(t, tc.want, recorder.Body.String())
		})
	}
}
