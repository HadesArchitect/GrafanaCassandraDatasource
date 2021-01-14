package main

import (
	simplejson "github.com/bitly/go-simplejson"
	"net/http"
)

type TargetResponseDTO struct {
	Target     string           `json:"target,omitempty"`
	DataPoints TimeSeriesPoints `json:"datapoints,omitempty"`
}

type TimePoint [2]float64
type TimeSeriesPoints []TimePoint

type TableColumn struct {
	Text string `json:"text"`
	Type string `json:"type"`
}

type RowValues []interface{}

type RemoteDatasourceRequest struct {
	queryType string
	req       *http.Request
	queries   []*simplejson.Json
}
