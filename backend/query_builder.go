package main

import (
	"fmt"

	simplejson "github.com/bitly/go-simplejson"
)

type QueryBuilder struct{}

func (qb *QueryBuilder) MetricQuery(queryData *simplejson.Json) string {
	allowFiltering := ""
	if queryData.Get("filtering").MustBool() {
		allowFiltering = " ALLOW FILTERING"
	}

	preparedQuery := fmt.Sprintf(
		"SELECT %s, CAST(%s as double) FROM %s.%s WHERE %s = %s %s",
		queryData.Get("columnTime").MustString(),
		queryData.Get("columnValue").MustString(),
		queryData.Get("keyspace").MustString(),
		queryData.Get("table").MustString(),
		queryData.Get("columnId").MustString(),
		queryData.Get("valueId").MustString(),
		allowFiltering,
	)

	return preparedQuery
}
