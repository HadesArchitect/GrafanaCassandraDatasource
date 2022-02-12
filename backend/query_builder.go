package main

import (
	"fmt"
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
