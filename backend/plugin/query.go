package plugin

import (
	"fmt"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

type Query struct {
	RawQuery       bool
	Target         string
	Keyspace       string
	Table          string
	ColumnValue    string
	ColumnID       string
	ValueID        string
	AliasID        string
	ColumnTime     string
	TimeFrom       time.Time
	TimeTo         time.Time
	AllowFiltering bool
	Instant        bool
	IsAlertQuery   bool
}

// BuildStatement builds cassandra query statement with positional parameters.
func (q *Query) BuildStatement() string {
	var allowFiltering string
	if q.AllowFiltering {
		allowFiltering = " ALLOW FILTERING"
	}

	var perPartitionLimit string
	if q.Instant {
		perPartitionLimit = " PER PARTITION LIMIT 1"
	}

	statement := fmt.Sprintf(
		"SELECT %s, %s, %s FROM %s.%s WHERE %s IN ? AND %s >= ? AND %s <= ?%s%s",
		q.ColumnID,
		q.ColumnValue,
		q.ColumnTime,
		q.Keyspace,
		q.Table,
		q.ColumnID,
		q.ColumnTime,
		q.ColumnTime,
		perPartitionLimit,
		allowFiltering,
	)

	backend.Logger.Debug("Built strict statement", "statement", statement)

	return statement
}
