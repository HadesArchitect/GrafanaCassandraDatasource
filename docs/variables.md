# Variables

Variables allow you to create dynamic filters for your dashboards. They can be used to query values from your Cassandra database and reference them in your queries.

## Basic Variables

Define variables in dashboard Settings > Variables. Each variable executes a CQL query to fetch available values:

![Variables example](assets/variables.png)

The query is written in the same code editor as the [Query Editor](editor.md), with CQL highlighted as SQL and autocomplete for Grafana's `$__from`/`$__to` and the other dashboard variables (type `$`). The query is applied when the editor loses focus or with `Ctrl+Enter` (`Cmd+Enter` on macOS), and Grafana then refreshes the preview of the values.

The data query macros `$__timeFrom`, `$__timeTo`, `$__unixEpochFrom` and `$__unixEpochTo` are not replaced in variable queries. To limit a variable query to the dashboard time range, use `$__from` and `$__to` and set the variable to refresh **On time range change**.

Use variables in your queries by referencing them with `$variable_name` syntax.

## Chained Variables

Variables can reference other variables in their queries, creating a chain of dependent values:

![Chained variables example](assets/chained-variables.png)

In this example, the `location` variable depends on the `zone` variable using `$zone` in its query.

**Important:** Variable order matters. Variables can only reference variables defined **above** them in the list. Lower variables can depend on higher variables, but not vice versa.

<https://github.com/user-attachments/assets/7793578b-1bb8-46be-ac23-dcc1a0c5ce41>

## Multi-Value Variables

When a variable is configured to allow **multiple values** or the **"All"** option, you must use the `:singlequote` formatter so Grafana wraps each selected value in single quotes — as required by CQL's `IN` clause.

```sql
SELECT sensor_id, temperature, registered_at, location
FROM test.test
WHERE location IN (${location:singlequote})
  AND registered_at > $__timeFrom
  AND registered_at < $__timeTo
ALLOW FILTERING
```

> **Note:** Use `${variable:singlequote}` (curly-brace syntax with the `:singlequote` modifier) instead of the plain `$variable` shorthand. Without this formatter, multi-value selections will not produce valid CQL and the query will fail.

## Demo Data

The [`demo/sample_data.sh`](../demo/sample_data.sh) script creates example tables including `test.zones_locations` which demonstrates the chained variables feature with zone and location hierarchies.
