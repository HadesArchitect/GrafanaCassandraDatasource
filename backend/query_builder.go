package main

import (
	"fmt"
	"strings"
)

type QueryBuilder struct{}

func (qb *QueryBuilder) prepareStrictMetricQuery(query *CassandraQuery, timeRangeFrom string, timeRangeTo string) string {
	allowFiltering := ""
	if query.AllowFiltering {
		allowFiltering = " ALLOW FILTERING"
	}

	preparedQuery := fmt.Sprintf(
		"SELECT %s, CAST(%s as double) FROM %s.%s WHERE %s = %s AND %s >= '%s' AND %s <= '%s'%s",
		query.ColumnTime,
		query.ColumnValue,
		query.Keyspace,
		query.Table,
		query.ColumnID,
		query.ValueID,
		query.ColumnTime,
		timeRangeFrom,
		query.ColumnTime,
		timeRangeTo,
		allowFiltering,
	)

	return preparedQuery
}

func (qb *QueryBuilder) prepareRawMetricQuery(query *CassandraQuery, timeRangeFrom string, timeRangeTo string) string {
	if !query.RawQuery {
		return qb.prepareStrictMetricQuery(query, timeRangeFrom, timeRangeTo)
	}

	timeRangeReplacer := strings.NewReplacer("$__timeFrom", timeRangeFrom, "$__timeTo", timeRangeTo)

	return timeRangeReplacer.Replace(query.Target)
}
