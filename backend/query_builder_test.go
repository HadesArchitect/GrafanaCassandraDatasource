package main

import (
	"testing"
)

func TestMetricQuery(t *testing.T) {
	var builder QueryBuilder

	var expected string = `SELECT created_at, CAST(value as double) FROM test.test WHERE id = 99051fe9-6a9c-46c2-b949-38ef78858dd0 AND created_at >= '1607364105184' AND created_at <= '1607623305184'`

	query := &CassandraQuery{
		DatasourceID:   1,
		QueryType:      "query",
		RawQuery:       false,
		RefID:          "A",
		Target:         "select id, cast(value as double), created_at from test.test where id in (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd1) and created_at > $__timeFrom",
		ColumnTime:     "created_at",
		ColumnValue:    "value",
		Keyspace:       "test",
		Table:          "test",
		ColumnID:       "id",
		ValueID:        "99051fe9-6a9c-46c2-b949-38ef78858dd0",
		AllowFiltering: false,
	}

	var result string = builder.prepareStrictMetricQuery(query, "1607364105184", "1607623305184")

	if result != expected {
		t.Errorf("\nGot: %s\nExpected: %s", result, expected)
	}
}
