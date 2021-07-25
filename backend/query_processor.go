package main

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-model/go/datasource"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/data"
)

type QueryProcessor struct{}

func (qp *QueryProcessor) processRawMetricQuery(result *datasource.QueryResult, query string, ds *CassandraDatasource) {

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
			Timestamp: timestamp.UnixNano() / int64(time.Millisecond),
			Value:     value,
		})

	}

	if err := iter.Close(); err != nil {
		ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
		result.Error = err.Error()

		return
	}

	for _, serie2 := range series {
		result.Series = append(result.Series, serie2)
	}
}

func (qp *QueryProcessor) processRawMetricQuery1(result *backend.DataResponse, query string, ds *CassandraDatasource) {

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

	if err := iter.Close(); err != nil {
		ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
		result.Error = err

		return
	}

	frames := make([]*data.Frame, len(series))
	i := 0
	for _, serie := range series {

		frame := data.NewFrame(serie.Name,
			data.NewField("time", nil, make([]time.Time, len(serie.Points))),
			data.NewField(serie.Name, serie.Tags, make([]float64, len(serie.Points))),
		)

		for pIdx, point := range serie.Points {
			frame.Set(0, pIdx, time.Unix(0, point.Timestamp))
			frame.Set(1, pIdx, point.Value)
			//frame.RefID = "A"
		}

		frames[i] = frame
		i = i + 1
	}

	result.Frames = frames
}

func (qp *QueryProcessor) processStrictMetricQuery(result *datasource.QueryResult, query string, valueId string, ds *CassandraDatasource) {

	iter := ds.session.Query(query).Iter()

	var timestamp time.Time
	var value float64

	serie := &datasource.TimeSeries{Name: valueId}

	for iter.Scan(&timestamp, &value) {
		serie.Points = append(serie.Points, &datasource.Point{
			Timestamp: timestamp.UnixNano() / int64(time.Millisecond),
			Value:     value,
		})
	}
	if err := iter.Close(); err != nil {
		ds.logger.Error(fmt.Sprintf("Error while processing a query: %s\n", err.Error()))
		result.Error = err.Error()

		return
	}

	result.Series = append(result.Series, serie)
}
