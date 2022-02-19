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

func (qp *QueryProcessor) processStrictMetricQuery(query string, valueId string, ds *CassandraDatasource) (data.Frames, error) {
	iter := ds.session.Query(query).Iter()

	var timestamp time.Time
	var value float64

	serie := &datasource.TimeSeries{Name: valueId}

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

	return []*data.Frame{timeSerieToFrame(serie)}, nil
}

func timeSerieToFrame(serie *datasource.TimeSeries) *data.Frame {
	frame := data.NewFrame(
		serie.Name,
		data.NewField("time", nil, make([]time.Time, len(serie.Points))),
		data.NewField(serie.Name, serie.Tags, make([]float64, len(serie.Points))),
	)

	for pIdx, point := range serie.Points {
		frame.Set(0, pIdx, time.Unix(0, point.Timestamp))
		frame.Set(1, pIdx, point.Value)
	}

	return frame
}
