package main

import (
	"testing"

	simplejson "github.com/bitly/go-simplejson"
)

func Errorf(t *testing.T, result *string, expected *string) {
	t.Errorf("\nGot: %s\nExpected: %s", *result, *expected)
}

func TestMetricQuery(t *testing.T) {
	var builder QueryBuilder

	var expected string = `SELECT created_at, CAST(value as double) FROM test.test WHERE id = 99051fe9-6a9c-46c2-b949-38ef78858dd0 AND created_at >= '1607364105184' AND created_at <= '1607623305184'`

	var rawString string = `{"queryType":"query","target":"select id, cast(value as double), created_at from test.test where id in (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd1) and created_at > $__timeFrom","refId":"A","rawQuery":false,"type":"timeserie","datasourceId":1,"keyspace":"test","table":"test","columnTime":"created_at","columnValue":"value","columnId":"id","valueId":"99051fe9-6a9c-46c2-b949-38ef78858dd0"}`

	queryData, _ := simplejson.NewJson([]byte(rawString))
	var result string = builder.MetricQuery(queryData, "1607364105184", "1607623305184")

	if result != expected {
		Errorf(t, &result, &expected)
	}
}

func TestRawMetricQuery(t *testing.T) {
	var builder QueryBuilder

	var expected string = `select id, cast(value as double), created_at from test.test where id in (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd1) and created_at > 1607364105184`

	var rawString string = `{"queryType":"query","target":"select id, cast(value as double), created_at from test.test where id in (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd1) and created_at > $__timeFrom","refId":"A","rawQuery":true,"type":"timeserie","datasourceId":1,"keyspace":"test","table":"test","columnTime":"created_at","columnValue":"value","columnId":"id","valueId":"99051fe9-6a9c-46c2-b949-38ef78858dd0"}`

	queryData, _ := simplejson.NewJson([]byte(rawString))
	var result string = builder.RawMetricQuery(queryData, "1607364105184", "1607623305184")

	if result != expected {
		Errorf(t, &result, &expected)
	}
}
