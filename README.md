# Apache Cassandra Datasource for Grafana

A Grafana data source for visualizing time-series data stored in Apache Cassandra and CQL-compatible databases. This well-established plugin with over a million downloads supports Cassandra 3.x-5.x, DataStax Enterprise, DataStax Astra, AWS Keyspaces, and more. Key features include authentication, TLS, query configurator, CQL editor, table mode, variables, annotations, alerting, and provisioning.

See the [Quick Demo](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/quick-demo.md) for a quick start, or refer to our [Documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki) for detailed information.

[Installation options](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/installation.md)

## Compatibility

* Grafana 
    * 7.4 - 12.x are fully supported (plugin version 3.x)
    * 5.x, 6.x, 7.0-7.3 are deprecated (works with plugin versions 1.x/2.x, but we recommend upgrading)
* Cassandra 3.x, 4.x, 5.x
* DataStax Enterprise 6.x
* DataStax Astra ([docs](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/connections/astra.md))
* AWS Keyspaces (limited support) ([docs](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/connections/aws-keyspaces.md))
* Linux, OSX (incl. M series), Windows

## Features

* Connect to Cassandra using auth credentials and TLS
* Query Configurator
* Raw CQL query editor
* Table mode
* Variables
* Annotations
* Alerting
* Provisioning

## Contacts

* [![Github discussions](https://img.shields.io/badge/github-discussions-green)](https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions)

## Configuration

1. Add the `Apache Cassandra Data Source` as a data source at the datasource configuration page.
2. Configure the datasource specifying contact point and port like `10.11.12.13:9042`, username and password. It's strongly recommended to use a dedicated user with read-only permissions only to the table you have to access.
3. Push the "Save and Test" button, if there is an error message, check the credentials and connection.

For a quick setup with Grafana automatic provisioning, see [docs/provisioning.md](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/provisioning.md).

![Datasource Configuration](https://user-images.githubusercontent.com/1742301/148654400-3ac4a477-8ca3-4606-86e7-5d10cbdc4ea9.png)

## Usage

There are **two ways** to query data from Cassandra/DSE, **Query Configurator** and **Query Editor**. Configurator is easier to use but has limited capabilities, Editor is more powerful but requires understanding of [CQL](https://cassandra.apache.org/doc/latest/cassandra/developing/cql/index.html). 

### Query Configurator

![Query Configurator](https://user-images.githubusercontent.com/1742301/148654262-b9cb7253-4086-4367-8aae-35ea458fcbb6.png)

Query Configurator is the easiest way to query data. At first, enter the keyspace and table name, then pick proper columns. If keyspace and table names are given correctly, the datasource will suggest the column names automatically.

* **Time Column** - the column storing the timestamp value, it's used to answer "when" question. 
* **Value Column** - the column storing the value you'd like to show. It can be the `value`, `temperature` or whatever property you need.
* **ID Column** - the column to uniquely identify the source of the data, e.g. `sensor_id`, `shop_id` or whatever allows you to identify the origin of data.

After that, you have to specify the `ID Value`, the particular ID of the data origin you want to show. You may need to enable "ALLOW FILTERING" although we recommend to avoid it.

More information on [Query Configurator](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/configurator.md).

### Query Editor

Query Editor unlocks all possibilities of CQL including Used-Defined Functions, aggregations etc. To enable query editor, press 'toggle editor mode' button.

Example (using the sample table from the Query Configurator case):

```
SELECT sensor_id, temperature, registered_at, location FROM test.test WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) AND registered_at > $__timeFrom and registered_at < $__timeTo
```

1. Order of fields in the SELECT expression doesn't matter except `ID` field. This field used to distinguish different time series, so it is important to keep it or any other column with low cardinality on the first position.
* **Identifier** - the first property in the SELECT expression should be the ID, something that uniquely identifies the data (e.g. `sensor_id`)
* **Value** - There should be at least one numeric value among returned fields, if query result will be used to draw graph.
* **Timestamp** - There should be one timestamp value, if query result will be used to draw graph.
* There could be any number of additional fields, however be cautious when using multiple numeric fields as they are interpreted as values by grafana and therefore are drawn on TimeSeries graph.

![103153625-1fd85280-4792-11eb-9c00-085297802117](https://user-images.githubusercontent.com/1742301/148654522-8e50617d-0ba9-4c5a-a3f0-7badec92e31f.png)

More information on [Query Editor](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/editor.md).

### Table Mode

In addition to `TimeSeries` mode datasource supports `Table` mode to draw tables using Cassandra query results. Use `Merge`, `Sort by`, `Organize fields` and other transformations to shape the table in any desirable way.
There are two ways to plot not a whole timeseries but only last(most rescent) values.

More information on [Table Mode](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/table.md).

### Variables
- [Configuring variables in Cassandra Datasource](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/variables.md)
- [Grafana Variables documentation](https://grafana.com/docs/grafana/latest/dashboards/variables/)

### Aliases
Using aliases explained in [documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/Aliases)

### Annotations
[Grafana Annotations documentation](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/annotate-visualizations/)

### Alerting

Alerting is supported, however it has some limitations. Grafana does not support long(narrow) series in alerting, so query result must be converted to wide series before handing it over to grafana. Datasource performs it in pretty simple way - it creates labels using all the non-timeseries field and then removes that fields from response.

Basically, this query (using example table) will produce two wide series for alerting

```
SELECT sensor_id, temperature, registered_at, location
FROM test.test
WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd0)
AND registered_at > $__timeFrom AND registered_at < $__timeTo

99051fe9-6a9c-46c2-b949-38ef78858dd0 {location="kitchen", sensor_id="99051fe9-6a9c-46c2-b949-38ef78858dd0"}
99051fe9-6a9c-46c2-b949-38ef78858dd1 {location="bedroom", sensor_id="99051fe9-6a9c-46c2-b949-38ef78858dd1"}
```

- More information on series types in [grafana developers documentation](https://grafana.com/developers/plugin-tools/introduction/data-frames#data-frames-as-time-series).
- [Grafana Alerting documentation](https://grafana.com/docs/grafana/latest/alerting/alerting-rules/create-grafana-managed-rule/)

### Tips and tricks

- [Unix epoch time format](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/unix-epoch.md)
- [Cassandra fat partitions](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/partitions.md)

## Development

- [Developer documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/Developer-Guide)
- [Release Process](https://github.com/HadesArchitect/GrafanaCassandraDatasource/tree/main/docs/development/release.md)
