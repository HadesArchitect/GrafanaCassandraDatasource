package cassandra

import (
	"crypto/tls"
	"time"
)

type Settings struct {
	Hosts       []string
	Keyspace    string
	User        string
	Password    string
	Consistency string
	Timeout     *int
	TLSConfig   *tls.Config
}

type Query struct {
	RawQuery       bool
	Target         string
	Keyspace       string
	Table          string
	ColumnValue    string
	ColumnID       string
	ValueID        string
	Longitude      string
	Latitude       string
	AliasID        string
	ColumnTime     string
	TimeFrom       time.Time
	TimeTo         time.Time
	AllowFiltering bool
}

type TimeSeriesPoint struct {
	Timestamp time.Time
	Value     float64
	Longitude string
	Latitude  string
	Target    string
}
