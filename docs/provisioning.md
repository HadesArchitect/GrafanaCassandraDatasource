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