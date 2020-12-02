package main


type QueryBuilder struct {}

func (qb *QueryBuilder) MetricQuery(queryData simplejson.Json) (string){
	allowFiltering := func() string { if queryData.Get("filtering").MustString() == "Disallow" { return "" } else { return "ALLOW FILTERING" } }()	

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