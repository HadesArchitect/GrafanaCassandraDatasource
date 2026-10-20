package plugin

import (
	"fmt"
	"strings"
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

// describeStatement returns the statement built by BuildStatement followed by
// a CQL comment listing the values bound to its placeholders, as shown in
// Grafana's query inspector. The values are the ID values and the time range
// of the query: nothing that is not already part of the query or the panel.
func (q *Query) describeStatement(statement string) string {
	const timeFormat = "2006-01-02T15:04:05.000Z07:00"

	return fmt.Sprintf("%s\n-- bound values: %s IN (%s), %s >= %s, %s <= %s",
		statement,
		q.ColumnID, strings.Join(splitIDs(q.ValueID), ", "),
		q.ColumnTime, q.TimeFrom.UTC().Format(timeFormat),
		q.ColumnTime, q.TimeTo.UTC().Format(timeFormat),
	)
}
