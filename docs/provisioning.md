# Grafana Provisioning with Cassandra Data Source

This setup allows you to quickly start Grafana with a preconfigured Cassandra data source using Docker Compose.

## Quick Start

1. Create the following files in your project directory:

```docker-compose.yaml
services:
  grafana:
    image: grafana/grafana-oss:main
    ports:
      - "3000:3000"
    volumes:
      - ./datasource:/etc/grafana/provisioning/datasources
    environment:
      GF_PLUGINS_PREINSTALL: hadesarchitect-cassandra-datasource
      GF_LOG_LEVEL: debug
      GF_AUTH_ANONYMOUS_ORG_ROLE: Admin
      GF_AUTH_ANONYMOUS_ENABLED: "true"
      GF_AUTH_BASIC_ENABLED: "false"
      GF_DEFAULT_APP_MODE: development
  cassandra:
    image: cassandra:4
    ports:
      - "9042:9042"
    volumes:
      - /var/lib/cassandra
```

2. Create a `datasource/` directory and add the data source configuration:

```datasource/cassandra.yaml
apiVersion: 1
datasources:
  - name: "Cassandra"
    type: hadesarchitect-cassandra-datasource
    access: proxy
    url: "cassandra:9042"
    jsonData:
      keyspace: system
      consistency: ONE
      user: cassandra
    secureJsonData:
      password: cassandra
```

3. Start the services:
```bash
docker-compose up -d
```

4. Access Grafana at http://localhost:3000 (no login required)

The Cassandra data source will be automatically configured and ready to use.