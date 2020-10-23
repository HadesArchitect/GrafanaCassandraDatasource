# Cassandra DataSource for Grafana 

Apache Cassandra & DataStax Enterprise Datasource for Grafana.

## Usage

### Installation 

1. Download the plugin using [latest release](https://github.com/HadesArchitect/grafana-cassandra-source/releases/tag/0.1.7), please download `cassandra-datasource-VERSION.zip` or `cassandra-datasource-VERSION.tar.gz` and uncompress a file into the Grafana plugins directory.
2. The plugin is yet unsigned by Grafana so it may require additional step to enable the plugin:
  2.1. If you use a local version, enable plugin in `/etc/grafana/grafana.ini`
    ```
    [plugins]
    allow_loading_unsigned_plugins = "hadesarchitect-cassandra-datasource"
    ```
    2.2 If you use dockerized Grafana, you need to set environment variable `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=hadesarchitect-cassandra-datasource`.
3. Add the Cassandra DataSource as a datasource at the configuration page.
4. Configure the datasource specifying contact point and port like "10.11.12.13:9042", username, password and keyspace. All the fields are required. It's recommended to use a dedicated user with read-only permissions only to the table you have to access.

### Panel Setup

[TBD]

## Development

*This part of the documentation relates only to development of the plugin and not required if you only intended to use it.*

Frontend part is implemented using *Typescript*, *WebPack*, *ESLint* and *NPM*, backend is written on *Golang* and uses *Dep* as a dependency manager. The plugin development uses docker actively and it's recommended to have at least basic understanding of docker and docker-compose.

### Installation and Build

First, clone the project. It has to be built with docker or with locally installed tools. 

#### Docker Way (Recommended)

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 npm install`
* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 node node_modules/webpack/bin/webpack.js`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp instrumentisto/dep ensure`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

#### Locally

* `npm install`
* `webpack`
* `dep ensure`
* `go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

#### Run Grafana and Cassandra

`docker-compose up -d`

docker-compose includes two services:

- *Grafana* by itself, the plugin is mounted as a volume to `/var/lib/grafana/plugins/cassandra`. Verbose logging is enabled. Grafana is available at http://localhost:3000, user `admin`, password `admin`
- *Apache Cassandra*, `localhost:9042`, user `cassandra`, password `cassandra`. `cqlsh` is available via `docker-compose exec cassandra cqlsh -u cassanrda -p cassandra`.

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

Test data uses keyspace name `test`

```
docker-compose exec cassandra cqlsh -u cassandra -p cassandra -f ./test_data.cql
```

### Making Changes

#### Frontend

Run `webpack` with `--watch` option to enable watching:

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 node node_modules/webpack/bin/webpack.js --watch`
* `docker-compose restart grafana`

#### Backend

With any changes done to backend, the binary file should be recompiled and *grafana* should be restarted:

* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`
* `docker-compose restart grafana`
