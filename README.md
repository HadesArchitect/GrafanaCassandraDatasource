# Cassandra DataSource for Grafana 

Apache Cassandra & DataStax Enterprise Datasource for Grafana. This datasource was built to visualise **time-series data** stored in Cassandra/DSE, if you are looking for Cassandra **metrics**, you may need [datastax/metric-collector-for-apache-cassandra](https://github.com/datastax/metric-collector-for-apache-cassandra) instead.

![Release Status](https://github.com/HadesArchitect/GrafanaCassandraDatasource/workflows/Handle%20Release/badge.svg)
 ![CodeQL](https://github.com/HadesArchitect/grafana-cassandra-source/workflows/CodeQL/badge.svg?branch=master)

Supports:

* Grafana 5.x, 6.x, 7.x (4.x not tested)
* Cassandra 3.x, 4.x (2.x not tested)
* DSE 6.x
* Linux, OSX (Windows not tested but should work)

## Usage

### Installation 

1. Download the plugin using [latest release](https://github.com/HadesArchitect/GrafanaCassandraDatasource/releases/latest), please download `cassandra-datasource-VERSION.zip` or `cassandra-datasource-VERSION.tar.gz` and uncompress a file into the Grafana plugins directory (`grafana/plugins`).
2. The plugin is yet unsigned by Grafana ([WiP #58](https://github.com/HadesArchitect/GrafanaCassandraDatasource/issues/58)) so it may require additional step to enable the plugin if you are using Grafana 7.x:

    2.1. If you use a local version, enable plugin in `/etc/grafana/grafana.ini`
    ```
    [plugins]
    allow_loading_unsigned_plugins = "hadesarchitect-cassandra-datasource"
    ```
    2.2 If you use dockerized Grafana, you need to set environment variable `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=hadesarchitect-cassandra-datasource`.
3. Add the Cassandra DataSource as a datasource at the datasource configuration page.
4. Configure the datasource specifying contact point and port like "10.11.12.13:9042", username and password, skip the keyspace. It's recommended to use a dedicated user with read-only permissions only to the table you have to access.
5. Push the "Save and Test" button, if there is an error message, check the credentials and connection. 

<img src="https://user-images.githubusercontent.com/1742301/103153542-8a3cc300-4791-11eb-9479-4a2e3ec94463.png" width="500">

### Panel Setup

There are **two ways** to query data from Cassandra/DSE, Query Configurator and Query Editor. Configurator is easier to use but has limited capabilities, Editor is more powerful but requires understanding of [CQL](https://cassandra.apache.org/doc/latest/cql/). 

#### Query Configurator

<img src="https://user-images.githubusercontent.com/1742301/103153577-d12ab880-4791-11eb-9a6b-50c86423134d.png" width="500">

Query Configurator is the easiest way to query data. At first, enter the keyspace and table name, then pick proper columns. If keyspace and table name are given correctly, the datasource will suggest you the column names automatically.

* **Time Column** - the column storing the timestamp value, it's used to answer "when" question. 
* **Value Column** - the column storing the value you'd like to show. It can be the `value`, `temperature` or whatever property you need.
* **ID Column** - the column to uniquely identify the source of the data, e.g. `sensor_id`, `shop_id` or whatever allows you to identify the origin of data.

After that, you have to specify the `ID Value`, the particular ID of the data origin you want to show. You may need to enable "ALLOW FILTERING" although we recommend to avoid it.

**Example** Imagine you want to visualise reports of a temperature sensor installed in your smart home. Given the sensor reports its ID, time, location and temperature every minute, we create a table to store the data and put some values there:

```
CREATE TABLE IF NOT EXISTS smarthome.temperature (
    sensor_id uuid,
    registered_at timestamp,
    temperature int,
    location text,
    PRIMARY KEY ((id), registered_at)
);

insert into smarthome.temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:21:59.001+0000, 18, "kitchen");
insert into smarthome.temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:22:59.001+0000, 19, "kitchen");
insert into smarthome.temperature (sensor_id, registered_at, temperature, location) values (99051fe9-6a9c-46c2-b949-38ef78858dd0, 2020-04-01T11:23:59.001+0000, 20, "kitchen");
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

<img src="https://user-images.githubusercontent.com/1742301/102781863-a8bd4b80-4398-11eb-8c28-4d06a1f29279.png" width="300">

Query Editor unlocks all possibilities of CQL including Used-Defined Functions, aggregations etc. 

Example using [test_data.cql](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/master/test_data.cql):

```
SELECT id, CAST(value as double), created_at FROM test.test WHERE id IN (99051fe9-6a9c-46c2-b949-38ef78858dd1, 99051fe9-6a9c-46c2-b949-38ef78858dd0) AND created_at > $__timeFrom and created_at < $__timeTo
```

1. Follow the order of the SELECT expressions, it's important! 
* **Identifier** - the first property in the SELECT expression must be the ID, something that uniquely identifies the data (e.g. `sensor_id`)
* **Value** - The second property must be the value what you are going to show 
* **Timestamp** - The third value must be timestamp of the value.
All other properties will be ignored

2. To filter data by time, use `$__timeFrom` and `$__timeTo` placeholders as in the example. The datasource will replace them with time values from the panel. **Notice** It's important to add the placeholders otherwise query will try to fetch data for the whole period of time. Don't try to specify the timeframe on your own, just put the placeholders. It's grafana's job to specify time limits.

<img src="https://user-images.githubusercontent.com/1742301/103153625-1fd85280-4792-11eb-9c00-085297802117.png" width="500">

## Development

*This part of the documentation relates only to development of the plugin and not required if you only intended to use it.*

Frontend part is implemented using *Typescript*, *WebPack*, *ESLint* and *NPM*, backend is written on *Golang* and uses *Dep* as a dependency manager. The plugin development uses docker actively and it's recommended to have at least basic understanding of docker and docker-compose.

### Installation and Build

First, clone the project. It has to be built with docker or with locally installed tools. 

#### Docker Way (Recommended)

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:12 npm install`
* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:12 node node_modules/webpack/bin/webpack.js`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp instrumentisto/dep ensure`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

#### Locally

* `npm install`
* `webpack`
* `dep ensure`
* `go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

#### Building the backend with TLS support:

1) `go get -u github.com/go-bindata/go-bindata/...` - downloading the bindata package
2) Place your tls certificate and tls key into `./backend/creds` folder
3) `cd ./backend && go-bindata -o assets.go ./creds && cd ..` - move credentials files as a `.go` files
4) `go build -i -ldflags "-X main.CertPath=/creds/cert_file_name -X main.KeyPath=/creds/key_file_name -X main.InsecureSkipVerify=true" -o ./dist/cassandra-plugin_linux_amd64 ./backend` - build binary with required variables filled in. 
   If you'd like to use rootCA, do `go build -i -ldflags "-X main.RootCA=/creds/root_ca_file_name" -o ./dist/cassandra-plugin_linux_amd64 ./backend`


#### Run Grafana and Cassandra

`docker-compose up -d`

docker-compose includes two services:

- *Grafana* by itself, the plugin is mounted as a volume to `/var/lib/grafana/plugins/cassandra`. Verbose logging is enabled. Grafana is available at http://localhost:3000, user `admin`, password `admin`
- *Apache Cassandra*, host `cassandra:9042`, user `cassandra`, password `cassandra`. `cqlsh` is available via `docker-compose exec cassandra cqlsh -u cassandra -p cassandra`.

After the startup, the datasource should be available in the list of datasources. Also, following lines should appear in grafana logs:

```
# Frontend part registered
lvl=info msg="Starting plugin search" logger=plugins
lvl=info msg="Registering plugin" logger=plugins name="Apache Cassandra"
...
# Backend part is started and running
msg="Plugins: Adding route" logger=http.server route=/public/plugins/hadesarchitect-cassandra-datasource dir=/var/lib/grafana/plugins/cassandra/dist
msg="starting plugin" logger=plugins plugin-id=hadesarchitect-cassandra-datasource path=/var/lib/grafana/plugins/cassandra/dist/cassandra-plugin_linux_amd64 args=[/var/lib/grafana/plugins/cassandra/dist/cassandra-plugin_linux_amd64]
msg="plugin started" logger=plugins plugin-id=hadesarchitect-cassandra-datasource path=/var/lib/grafana/plugins/cassandra/dist/cassandra-plugin_linux_amd64 pid=23
msg="waiting for RPC address" logger=plugins plugin-id=hadesarchitect-cassandra-datasource path=/var/lib/grafana/plugins/cassandra/dist/cassandra-plugin_linux_amd64
msg="2020-01-16T22:08:51.619Z [DEBUG] cassandra-backend-datasource: Running Cassandra backend datasource..." logger=plugins plugin-id=hadesarchitect-cassandra-datasource
msg="plugin address" logger=plugins plugin-id=hadesarchitect-cassandra-datasource address=/tmp/plugin991218850 network=unix timestamp=2020-01-16T22:08:51.622Z
msg="using plugin" logger=plugins plugin-id=hadesarchitect-cassandra-datasource version=1
```

To read the logs, use `docker-compose logs -f grafana`.

### Load Sample Data

```
docker-compose exec cassandra cqlsh -u cassandra -p cassandra -f ./test_data.cql
```

### Testing

#### Docker Way (Recommended)

Backend tests: `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go test ./backend`

#### Locally

Backend tests: `go test ./backend`

### Making Changes

#### Frontend

Run `webpack` with `--watch` option to enable watching:

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:12 node node_modules/webpack/bin/webpack.js --watch`
* `docker-compose restart grafana`

#### Backend

With any changes done to backend, the binary file should be recompiled and *grafana* should be restarted:

* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`
* `docker-compose restart grafana`
