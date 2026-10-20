# Query Editor

Query Editor unlocks all possibilities of CQL including User-Defined Functions, aggregations etc. To enable query editor, select **Query Editor** in the mode switch at the top of the query.

## The editor

The CQL query is written in a code editor (the same Monaco editor Grafana uses for its own query languages):

* **Syntax highlighting and line numbers.** There is no dedicated CQL mode, so the editor highlights CQL as SQL. Keywords, strings and numbers are highlighted as expected; CQL-only keywords may not be.
* **Autocomplete.** Type `$` to get suggestions for the time range macros, Grafana's `$__from`/`$__to` and the variables of the dashboard. Each macro comes with a short description. `Ctrl+Space` opens the suggestions at any time.
* **Run with `Ctrl+Enter`** (`Cmd+Enter` on macOS). The query also runs when the editor loses focus, as before.
* The editor grows with the query, up to about 20 lines, and scrolls after that.

Saved queries are not changed by the editor: they are stored in the same format as before and open as they are.

A new query, or one whose CQL is still empty, is not run and shows no error. Other queries of the same panel run as usual.

## Time range macros

| Macro | Replaced with |
| --- | --- |
| `$__timeFrom` / `$__timeTo` | Start / end of the panel time range in milliseconds since the Unix epoch. Use them with `timestamp` columns, or `bigint` columns holding milliseconds. |
| `$__unixEpochFrom` / `$__unixEpochTo` | Start / end of the panel time range in seconds since the Unix epoch, for columns storing Unix time in seconds. See [Unix epoch time format](unix-epoch.md). |
| `$__from` / `$__to` | Grafana's global variables: start / end of the dashboard time range in milliseconds since the Unix epoch, with [formatting options](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to) such as `${__from:date:iso}`. |

## Usage

Example (using the sample table from the Query Configurator case):

```cql
SELECT sensor_id, temperature, registered_at, location FROM test.test WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) AND registered_at > $__timeFrom and registered_at < $__timeTo
```

1. Order of fields in the SELECT expression doesn't matter except `ID` field. This field used to distinguish different time series, so it is important to keep it or any other column with low cardinality on the first position.
    * **Identifier** - the first property in the SELECT expression should be the ID, something that uniquely identifies the data (e.g. `sensor_id`)
    * **Value** - There should be at least one numeric value among returned fields, if query result will be used to draw graph.
    * **Timestamp** - There should be one timestamp value, if query result will be used to draw graph.
    * There could be any number of additional fields, however be cautious when using multiple numeric fields as they are interpreted as values by grafana and therefore are drawn on TimeSeries graph.
    * Any field returned by query is available to use in `Alias` template, e.g. `{{ location }}`. Datasource interpolates such strings and updates graph legend. 
    * Datasource will try to keep all the fields, however it is not always possible since cassandra and grafana use different sets of supported types. Unsupported fields will be removed from response.
2. To filter data by time, use `$__timeFrom` and `$__timeTo` placeholders as in the example. The datasource will replace them with time values from the panel. **Notice** It's important to add the placeholders otherwise query will try to fetch data for the whole period of time. Don't try to specify the timeframe on your own, just put the placeholders. It's grafana's job to specify time limits. See [Queries without a time filter](#queries-without-a-time-filter).

![103153625-1fd85280-4792-11eb-9c00-085297802117](https://user-images.githubusercontent.com/1742301/148654522-8e50617d-0ba9-4c5a-a3f0-7badec92e31f.png)

## Queries without a time filter

When the CQL uses none of `$__timeFrom`, `$__timeTo`, `$__unixEpochFrom`, `$__unixEpochTo`, `$__from` or `$__to` (including forms such as `${__from:date:iso}`), the editor shows a warning under the query: **This query has no time filter**. Such a query ignores the dashboard time range and reads every matching row, which on a large table can be millions of rows.

The warning is advice only: the query still runs. Ignore it for queries that are meant to read all rows, for example a table of sensors. It is shown for the query as written, so a time range passed through a dashboard variable of your own is not recognised.

Whatever the query, the result is capped by the **Max rows** setting of the data source (100000 rows by default), unless Max rows is set to `0`, which removes the limit. A query that returns more is truncated and its panel shows a warning, see [Truncated query results](troubleshooting.md#truncated-query-results).

## Query inspector

Open the panel's **Query inspector** (or **Inspect > Query** / **Inspect > Stats**) to see what a query did:

* **Executed query.** The CQL that was sent to Cassandra, after the time range macros and dashboard variables were replaced. For a Query Configurator query it is the statement the plugin built, followed by a comment with the bound values: the ID values and the time range in UTC, for example:

  ```cql
  SELECT sensor_id, temperature, registered_at FROM test.test WHERE sensor_id IN ? AND registered_at >= ? AND registered_at <= ?
  -- bound values: sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1), registered_at >= 2024-03-01T09:00:00.000Z, registered_at <= 2024-03-01T10:00:00.000Z
  ```

* **Stats.** Under *Data source stats*: **Rows returned**, the number of rows read from Cassandra across all series of the query, and **Execution time**, how long running the query and reading its rows took, in milliseconds.

A query that returns no rows still reports both, which helps to check that the executed CQL is the one you expected.

## Variables

* [Configuring variables in Cassandra Datasource](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/variables.md)
