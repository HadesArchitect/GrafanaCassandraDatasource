# Cassandra DataSource for Grafana 

Apache Cassandra & DataStax Enterprise Datasource for Grafana.

## Usage

[TBD]

## Development

*This part of the documentation relates only to development of the plugin and not required if you only intended to use it.*

Frontend part is implemented using *Typescript*, *WebPack*, *ESLint* and *NPM*, backend is written on *Golang* and uses *Dep* as a dependency manager. The plugin development uses docker actively and it's recommended to have at least basic understanding of docker and docker-compose.

### Installation and Build

First, clone the project. It has to be built with docker or with locally installed tools. 

#### Docker Way (Recommended)

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 npm install`
* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 node node_modules/webpack-cli/bin/cli.js`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp instrumentisto/dep ensure`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

#### Locally

* `npm install`
* `webpack`
* `dep ensure`
* `go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

### Run Grafana, Cassandra & Studio

`docker-compose up -d`

docker-compose includes three services:

- *Grafana* by itself, the plugin is mounted as a volume to `/var/lib/grafana/plugins/cassandra`
- DataStax Enterprise (Enterprise version of Apache Cassandra, will be replaced by the OSS Cassandra soon)
- DataStax Studio, web-based UI of the Cassandra to simplify development

After the startup, the datasource should be available in the list of datasources.

### Making Changes

#### Frontend

`watch` isn't implemented yet. After the frontend changes it should be rebuild and grafana container should be restarted. 

**NOTICE** Frontend build will wipe *dist/* folder that removes compiled backend binary file. 

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 node node_modules/webpack-cli/bin/cli.js`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`
* `docker-compose restart grafana`

#### Backend

With any changes done to backend, the binary file should be recompiled and *grafana* should be restarted:

* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`
* `docker-compose restart grafana`
