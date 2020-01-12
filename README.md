# Cassandra Grafana DataSource

**EARLY PHASE!!!**

Apache Cassandra & DataStax Enterprise Datasource for Grafana

## Install JS dependencies

* **Docker Way** `docker run -v ${PWD}:/opt/gcds -w /opt/gcds node npm install`
* **Locally** `npm install`

## Build 

* **Docker Way** `docker run -v ${PWD}:/opt/gcds -w /opt/gcds node node_modules/grunt-cli/bin/grunt`
* **Locally** `grunt`

## Run grafana, cassandra & studio

`docker-compose up -d`
