# Cassandra DataSource for Grafana 

Apache Cassandra Datasource for Grafana. This datasource is to visualise **time-series data** stored in Cassandra/DSE, if you are looking for Cassandra **metrics**, you may need [datastax/metric-collector-for-apache-cassandra](https://github.com/datastax/metric-collector-for-apache-cassandra) instead.

![Release Status](https://github.com/HadesArchitect/GrafanaCassandraDatasource/workflows/Handle%20Release/badge.svg)
 ![CodeQL](https://github.com/HadesArchitect/grafana-cassandra-source/workflows/CodeQL/badge.svg?branch=master) ![GitHub all releases](https://img.shields.io/github/downloads/hadesarchitect/grafanacassandradatasource/total?color=%2326c458&label=Downloads&logo=github)

To see the datasource in action, please follow the [Quick Demo](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/Quick-Demo) steps. Documentation is available [here](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki)

**Supports**:

* Grafana 
    * 7.x, 8.x are fully supported (plugin version 2.x)
    * 5.x, 6.x are deprecated (works with plugin versions 1.x, but we recommend upgrading)
* Cassandra 3.x, 4.x
* DataStax Enterprise 6.x
* DataStax Astra ([docs](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/DataStax-Astra))
* AWS Keyspaces (limited support)  ([docs](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/AWS-Keyspaces))
* Linux, OSX (incl. M1), Windows

**Contacts**:

* [![Discord Chat](https://img.shields.io/badge/discord-chat%20with%20us-green)](https://discord.gg/FU2Cb4KTyp) 
* [![Github discussions](https://img.shields.io/badge/github-discussions-green)](https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions) 

## Usage

You can find more detailed instructions in [the datasource wiki](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki).

### Installation 

1. Install the plugin using grafana console tool: `grafana-cli plugins install hadesarchitect-cassandra-datasource`. The plugin will be installed into your grafana plugins directory; the default is `/var/lib/grafana/plugins`. Alternatively, download the plugin using [latest release](https://github.com/HadesArchitect/GrafanaCassandraDatasource/releases/latest), please download `cassandra-datasource-VERSION.zip` and uncompress a file into the Grafana plugins directory (`grafana/plugins`).
2. Add the Apache Cassandra Data Source as a data source at the datasource configuration page.
3. Configure the datasource specifying contact point and port like "10.11.12.13:9042", username and password. It's strongly recommended to use a dedicated user with read-only permissions only to the table you have to access.
4. Push the "Save and Test" button, if there is an error message, check the credentials and connection. 

![Datasource Configuration](https://user-images.githubusercontent.com/1742301/148654400-3ac4a477-8ca3-4606-86e7-5d10cbdc4ea9.png)

### Panel Setup

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
SELECT sensor_id, CAST(temperature as double), registered_at FROM test.test WHERE sensor_id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) AND registered_at > $__timeFrom and registered_at < $__timeTo
```

1. Follow the order of the SELECT expressions, it's important! 
* **Identifier** - the first property in the SELECT expression must be the ID, something that uniquely identifies the data (e.g. `sensor_id`)
* **Value** - The second property must be the value what you are going to show 
* **Timestamp** - The third value must be timestamp of the value.
All other properties will be ignored

2. To filter data by time, use `$__timeFrom` and `$__timeTo` placeholders as in the example. The datasource will replace them with time values from the panel. **Notice** It's important to add the placeholders otherwise query will try to fetch data for the whole period of time. Don't try to specify the timeframe on your own, just put the placeholders. It's grafana's job to specify time limits.

![103153625-1fd85280-4792-11eb-9c00-085297802117](https://user-images.githubusercontent.com/1742301/148654522-8e50617d-0ba9-4c5a-a3f0-7badec92e31f.png)

## Development

[Developer documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/wiki/Developer-Guide)
