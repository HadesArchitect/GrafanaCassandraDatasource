# Cassandra Grafana DataSource

**EARLY PHASE!!!**

Apache Cassandra & DataStax Enterprise Datasource for Grafana

# Development

## Install & build frontend

### Docker Way

* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 npm install`
* `docker run --rm -v ${PWD}:/opt/gcds -w /opt/gcds node:11 node node_modules/webpack-cli/bin/cli.js`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/ha/gcp instrumentisto/dep ensure`
* `docker run --rm -v ${PWD}:/go/src/github.com/ha/gcp -w /go/src/github.com/hadesarchitect/grafana-cassandra-plugin golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./backend`

### Locally

* `npm install`
* `webpack`
* `dep ensure`
* `go build -i -o ./dist/cassandra-plugin_linux_amd64 ./pkg`

## Run grafana, cassandra & studio

`docker-compose up -d`
