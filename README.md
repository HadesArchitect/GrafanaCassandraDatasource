# Apache Cassandra Datasource for Grafana

This datasource is to visualise **time-series data** stored in Cassandra/DSE, if you are looking for Cassandra **metrics**, you may need [datastax/metric-collector-for-apache-cassandra](https://github.com/datastax/metric-collector-for-apache-cassandra) instead.

![Release Status](https://github.com/HadesArchitect/GrafanaCassandraDatasource/workflows/Handle%20Release/badge.svg)
 ![CodeQL](https://github.com/HadesArchitect/grafana-cassandra-source/workflows/CodeQL/badge.svg?branch=master)

To see the datasource in action, please follow the [Quick Demo](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/Quick-Demo) steps. Documentation is available [here](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki)

**Supports**:

* Grafana 
    * 7.x, 8.x, 9.x, 10.x are fully supported (plugin version 2.x)
    * 5.x, 6.x are deprecated (works with plugin versions 1.x, but we recommend upgrading)
* Cassandra 3.x, 4.x
* DataStax Enterprise 6.x
* DataStax Astra ([docs](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/DataStax-Astra))
* AWS Keyspaces (limited support)  ([docs](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/AWS-Keyspaces))
* Linux, OSX (incl. M1), Windows

**Features**:
* Connect to Cassandra using auth credentials and TLS
* Query configurator
* Raw CQL query editor
* Table mode
* Variables
* Annotations
* Alerting

**Contacts**:

* [![Github discussions](https://img.shields.io/badge/github-discussions-green)](https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions)

**TOC**
- [About](#about)
- [Usage](#usage)
- [Installation](#installation)
- [Building query](#building-query)
- [Table Mode](#table-mode)
- [Alerting](#alerting)
- [Tips and tricks](#tips-and-tricks)
- [Development](#tips-and-tricks)

## Usage

You can find more detailed instructions in [the datasource wiki](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki).

### Installation

1. Install the plugin using grafana console tool: `grafana-cli plugins install hadesarchitect-cassandra-datasource`. The plugin will be installed into your grafana plugins directory; the default is `/var/lib/grafana/plugins`. Alternatively, download the plugin using [latest release](https://github.com/HadesArchitect/GrafanaCassandraDatasource/releases/latest), please download `cassandra-datasource-VERSION.zip` and uncompress a file into the Grafana plugins directory (`grafana/plugins`).
2. Add the Apache Cassandra Data Source as a data source at the datasource configuration page.
3. Configure the datasource specifying contact point and port like "10.11.12.13:9042", username and password. It's strongly recommended to use a dedicated user with read-only permissions only to the table you have to access.
4. Push the "Save and Test" button, if there is an error message, check the credentials and connection. 

![Datasource Configuration](https://user-images.githubusercontent.com/1742301/148654400-3ac4a477-8ca3-4606-86e7-5d10cbdc4ea9.png)

### Building query

There are **two ways** to query data from Cassandra/DSE, **Query Configurator** and **Query Editor**. Configurator is easier to use but has limited capabilities, Editor is more powerful but requires understanding of [CQL](https://cassandra.apache.org/doc/latest/cql/). 

#### Query Configurator

![Query Configurator](https://user-images.githubusercontent.com/1742301/148654262-b9cb7253-4086-4367-8aae-35ea458fcbb6.png)

Query Configurator is the easiest way to query data. At first, enter the keyspace and table name, then pick proper columns. If keyspace and table names are given correctly, the datasource will suggest the column names automatically.

* **Time Column** - the column storing the timestamp value, it's used to answer "when" question. 
* **Value Column** - the column storing the value you'd like to show. It can be the `value`, `temperature` or whatever property you need.
* **ID Column** - the column to uniquely identify the source of the data, e.g. `sensor_id`, `shop_id` or whatever allows you to identify the origin of data.

After that, you have to specify the `ID Value`, the particular ID of the data origin you want to show. You may need to enable "ALLOW FILTERING" although we recommend to avoid it.

**Example** Imagine you want to visualise reports of a temperature sensor installed in your smart home. Given the sensor reports its ID, time, location and temperature every minute, we create a table to store the data and put some values there:

```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    location text,
    PRIMARY KEY ((sensor_id), registered_at)
);

insert into temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:21:59.001+0000, 18, "kitchen");
insert into temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:22:59.001+0000, 19, "kitchen");
insert into temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:23:59.001+0000, 20, "kitchen");
```

In this case, we have to fill the configurator fields the following way to get the results:

* **Keyspace** - smarthome *(keyspace name)*
* **Table** - temperature *(table name)*
* **Time Column** - registered_at *(occurence)*
* **Value Column** - temperature *(value to show)*
* **ID Column** - sensor_id *(ID of the data origin)*
* **ID Value** - 99051fe9-6a9c-46c2-b949-38ef78858dd0 *ID of the sensor*
* **ALLOW FILTERING** - FALSE *(not required, so we are happy to avoid)*

In case of a few origins (multiple sensors) you will need to add more rows. If your case is as simple as that, query configurator will be a good choice, otherwise  please proceed to the query editor.

#### Query Editor

Query Editor is more powerful way to query data. To enable query editor, press "toggle text edit mode" button.

![102781863-a8bd4b80-4398-11eb-8c28-4d06a1f29279](https://user-images.githubusercontent.com/1742301/148654475-6718f3ff-1290-4d7a-a40b-dc107c52ac15.png)

Query Editor unlocks all possibilities of CQL including Used-Defined Functions, aggregations etc. 

Example (using the sample table from the Query Configurator case):

```
SELECT sensor_id, temperature, registered_at, location FROM test.test WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) AND registered_at > $__timeFrom and registered_at < $__timeTo
```

1. Order of fields in the SELECT expression doesn't matter except `ID` field. This field used to distinguish different time series, so it is important to keep it or any other column with low cardinality on the first position.
* **Identifier** - the first property in the SELECT expression should be the ID, something that uniquely identifies the data (e.g. `sensor_id`)
* **Value** - There should be at least one numeric value among returned fields, if query result will be used to draw graph.
* **Timestamp** - There should be one timestamp value, if query result will be used to draw graph.
* There could be any number of additional fields, however be cautious when using multiple numeric fields as they are interpreted as values by grafana and therefore are drawn on TimeSeries graph.
* Any field returned by query is available to use in `Alias` template, e.g. `{{ location }}`. Datasource interpolates such strings and updates graph legend. 
* Datasource will try to keep all the fields, however it is not always possible since cassandra and grafana use different sets of supported types. Unsupported fields will be removed from response.

2. To filter data by time, use `$__timeFrom` and `$__timeTo` placeholders as in the example. The datasource will replace them with time values from the panel. **Notice** It's important to add the placeholders otherwise query will try to fetch data for the whole period of time. Don't try to specify the timeframe on your own, just put the placeholders. It's grafana's job to specify time limits.

![103153625-1fd85280-4792-11eb-9c00-085297802117](https://user-images.githubusercontent.com/1742301/148654522-8e50617d-0ba9-4c5a-a3f0-7badec92e31f.png)

### Table Mode
In addition to TimeSeries mode datasource supports Table mode to draw tables using Cassandra query results. Use `Merge`, `Sort by`, `Organize fields` and other transformations to shape the table in any desirable way.
There are two ways to plot not a whole timeseries but only last(most rescent) values.
1. Inefficient way

In case if table created with default ascending ordering the most recent value is always stored in the end of partition. To retrieve it `ORDER BY` and `LIMIT` clauses must be used in query:
```
SELECT sensor_id, temperature, registered_at, location
FROM test.test
WHERE sensor_id = 99051fe9-6a9c-46c2-b949-38ef78858dd0
AND registered_at > $__timeFrom and registered_at < $__timeTo
ORDER BY registered_at
LIMIT 1
```
Note that `WHERE IN ()` clause could not be used with `ORDER BY`, so query must be duplicated for any additional `sensor_id`.

2. Efficient way

To query the most recent values efficiently ordering must be specified during the table creation:
```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    location text,
    PRIMARY KEY ((sensor_id), registered_at)
) WITH CLUSTERING ORDER BY (registered_at DESC);
```
After that the most recent value will always be stored in the beginning of partition and could be queried with just `LIMIT` clause:
```
SELECT sensor_id, temperature, registered_at, room_name
FROM test.test
WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd0)
AND registered_at > $__timeFrom AND registered_at < $__timeTo
PER PARTITION LIMIT 1
```
Note that `PER PARTITION LIMIT 1` used instead of `LIMIT 1` to query one row for each partition and not just one row total.

### Variables
[Grafana Variables documentation](https://grafana.com/docs/grafana/latest/dashboards/variables/)

### Annotations
[Grafana Annotations documentation](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/annotate-visualizations/)

### Alerting
Alerting is supported, however it has some limitations. Grafana does not support long(narrow) series in alerting, 
so query result must be converted to wide series before handing it over to grafana. Datasource performs it in pretty
simple way - it creates labels using all the non-timeseries field and then removes that fields from response.
Basically, this query(using example table)
```
SELECT sensor_id, temperature, registered_at, location
FROM test.test
WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd0, 99051fe9-6a9c-46c2-b949-38ef78858dd0)
AND registered_at > $__timeFrom AND registered_at < $__timeTo
```
will produce two wide series for alerting
```
99051fe9-6a9c-46c2-b949-38ef78858dd0 {location="kitchen", sensor_id="99051fe9-6a9c-46c2-b949-38ef78858dd0"}
99051fe9-6a9c-46c2-b949-38ef78858dd1 {location="bedroom", sensor_id="99051fe9-6a9c-46c2-b949-38ef78858dd1"}


```
More information on series types in [grafana developers documentation](https://grafana.com/developers/plugin-tools/introduction/data-frames#data-frames-as-time-series).

[Grafana Alerting documentation](https://grafana.com/docs/grafana/latest/alerting/alerting-rules/create-grafana-managed-rule/)

### Tips and tricks

#### Unix epoch time format
Usually there are no problems - Cassandra can store timestamps using different formats as shown in [documentation](https://cassandra.apache.org/doc/latest/cassandra/cql/types.html#timestamps).
However, it is not always enough. One of possible cases could be unix time, which is just number of seconds or milliseconds and usually stored as integer type.
1. If time is stored as a number of milliseconds in a `bigint` column, then it should be converted into the `timestamp` type before return the data to grafana:
```
SELECT sensor_id, temperature, dateOf(maxTimeuuid(registered_at)), location
FROM test.test WHERE sensor_id = 99051fe9-6a9c-46c2-b949-38ef78858dd0
AND registered_at > $__timeFrom AND registered_at < $__timeTo
```
This query returns proper timestamp even if it stored as number of milliseconds.

2. If time is stored as a number of seconds, then it is not possible to convert it into the timestamp natively, but there is a trick:
```
SELECT sensor_id, temperature, dateOf(maxTimeuuid(registered_at*1000)), location
FROM test.test WHERE sensor_id = 99051fe9-6a9c-46c2-b949-38ef78858dd0
AND registered_at > $__unixEpochFrom AND registered_at < $__unixEpochTo
```
* There are two important parts in this query:
  * `dateOf(maxTimeuuid(registered_at*1000))` used to convert seconds to milliseconds(`registered_at*1000`) and then to convert milliseconds to `timestamp` type, which is handed over to grafana.
  * `$__unixEpochFrom` and `$__unixEpochTo` are variables with unix time in the seconds format that are used to fill out conditions part of the query.

#### Cassandra fat partitions
Cassandra stores data in `partitions` which are minimal storage units for the DB. It means that using the example table
```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    PRIMARY KEY ((sensor_id), registered_at)
);
```
will lead to partitions bloating and performance degradation, because all the data for all time for specific `sensor_id` is stored in just one partition(first part of `PRIMARY KEY` is `PARTITION KEY`).
To avoid that there is a technique called `bucketing`, which basically means that partitions are split up into smaller pieces.
For instance, we can split that example table partitions by time: year, month, day, or even hour and less. What to choose depends on how
much data stored in each partition. To achieve that the example table has to be modified like this:
```
CREATE TABLE IF NOT EXISTS temperature (
    sensor_id uuid,
    date date,
    registered_at timestamp,
    temperature int,
    PRIMARY KEY ((sensor_id, date), registered_at)
);
```
After that change the database schema became more effective because of bucketing by date, and queries will have a form of
```
SELECT sensor_id, temperature, registered_at
FROM temperature
WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) 
AND date = '${__from:date:YYYY-MM-DD}'
AND registered_at > $__timeFrom 
AND registered_at < $__timeTo
```
Note that `$__from`/`$__to` variables are used. They are [grafana built-in variables](https://grafana.com/docs/grafana/latest/dashboards/variables/add-template-variables/#__from-and-__to), and they have formatting capabilities which are perfect for our case.
In case when time range includes more than one day, each day has to be added into `AND date IN (...)` predicate. Another way to make it more convenient is to consider using larger buckets, e.g. month instead of day-size.

## Development

[Developer documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/Developer-Guide)
