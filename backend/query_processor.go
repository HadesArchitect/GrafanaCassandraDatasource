package main

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type QueryProcessor struct{}

func (qp *QueryProcessor) processRawMetricQuery(query string, ds *CassandraDatasource) (data.Frames, error) {
	iter := ds.session.Query(query).Iter()

	var id string
	var timestamp time.Time
	var value float64

	series := make(map[string]*datasource.TimeSeries)

	for iter.Scan(&id, &value, &timestamp) {
		if _, ok := series[id]; !ok {
			series[id] = &datasource.TimeSeries{Name: id}
		}

		series[id].Points = append(series[id].Points, &datasource.Point{
			Timestamp: timestamp.UnixNano(),
			Value:     value,
		})
	}

	ds.logger.Info(fmt.Sprintf("%+v'n", series))
	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("process query, err=%v", err)
	}

	frames := make([]*data.Frame, len(series))
	i := 0
	for _, serie := range series {
		frame := timeSerieToFrame(serie)

		frames[i] = frame
		i = i + 1
	}

	return frames, nil
}

func (qp *QueryProcessor) processStrictMetricQuery(query string, valueId, alias string, ds *CassandraDatasource) (data.Frames, error) {
	iter := ds.session.Query(query).Iter()

	if alias == "" {
		alias = valueId
	}

	var timestamp time.Time
	var value float64

	serie := &datasource.TimeSeries{Name: alias}

	for iter.Scan(&timestamp, &value) {
		serie.Points = append(serie.Points, &datasource.Point{
			Timestamp: timestamp.UnixNano(),
			Value:     value,
		})
		ds.logger.Warn(fmt.Sprintf("%+v", serie.Points))
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("process query, err=%v", err)
	}

	frame := data.NewFrame(
		serie.Name,
		data.NewField("time", nil, make([]time.Time, 0)),
		data.NewField(valueId, serie.Tags, make([]float64, 0)),
	)

	for _, point := range serie.Points {
		frame.AppendRow(time.Unix(0, point.Timestamp), point.Value)
	}

	return []*data.Frame{frame}, nil
}

func (qp *QueryProcessor) processKeyspacesQuery(ds *CassandraDatasource) (data.Frames, error) {
	iter := ds.session.Query("SELECT keyspace_name FROM system_schema.keyspaces;").Iter()

	var keyspace string
	var keyspaces []string = make([]string, 0)

	for iter.Scan(&keyspace) {
		keyspaces = append(keyspaces, keyspace)
	}

	if err := iter.Close(); err != nil {
		return nil, fmt.Errorf("process query, err=%v", err)
	}

	frame := data.NewFrame(
		"Keyspaces",
		data.NewField("keyspaces", nil, make([]string, 0)),
	)

	for _, keyspace := range keyspaces {
		frame.AppendRow(keyspace)
	}

	return []*data.Frame{frame}, nil
}

func timeSerieToFrame(serie *datasource.TimeSeries) *data.Frame {
	frame := data.NewFrame(
		serie.Name,
		data.NewField("time", nil, make([]time.Time, 0)),
		data.NewField(serie.Name, serie.Tags, make([]float64, 0)),
	)

	for _, point := range serie.Points {
		frame.AppendRow(time.Unix(0, point.Timestamp), point.Value)
	}

	return frame
}
