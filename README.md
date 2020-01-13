# Cassandra Grafana DataSource

**EARLY PHASE!!!**

Apache Cassandra & DataStax Enterprise Datasource for Grafana

## Install & build frontend

### Docker Way

* `docker run -v ${PWD}:/opt/gcds -w /opt/gcds node npm install`
* `docker run -v ${PWD}:/opt/gcds -w /opt/gcds node node_modules/grunt-cli/bin/grunt`

### Locally

* `npm install`
* `grunt`

## Install & build backend

### Docker Way

* `docker run -v ${PWD}:/go/src/github.com/hadesarchitect/grafana-cassandra-plugin -w /go/src/github.com/hadesarchitect/grafana-cassandra-plugin instrumentisto/dep ensure`
* `docker run -v ${PWD}:/go/src/github.com/hadesarchitect/grafana-cassandra-plugin -w /go/src/github.com/hadesarchitect/grafana-cassandra-plugin golang go build -i -o ./dist/cassandra-plugin_linux_amd64 ./pkg`

## Run grafana, cassandra & studio

`docker-compose up -d`
