---
'grafana-cassandra-datasource': minor
---

**The plugin now uses the Apache Cassandra GoCQL driver 2.x** (`github.com/apache/cassandra-gocql-driver/v2`), the Apache-maintained successor of the GoCQL 1.x driver used until now. Existing data sources keep working without changes: credentials, **Allowed authenticators**, TLS (certificate paths or content), consistency, timeout, AWS Keyspaces and the secure socks proxy (PDC) are configured exactly as before, and every column type that could be charted or shown in a table before is returned with the same values and formatting. Columns of type `decimal`, `time` and `duration` are still not supported, as before.

Connecting with an empty **Allowed authenticators** field still only accepts the same built-in list of server authenticators. The new driver would accept any authenticator by default, but the plugin keeps the old list so that credentials are never sent to an authenticator you have not allowed.

A few things behave slightly differently:

- On Cassandra 4.0 and newer the connection now uses native protocol v5. Clusters that only speak protocol v1 or v2 (Cassandra 2.0 and older) are no longer supported; Cassandra 3.x and newer, DataStax Enterprise, Astra and AWS Keyspaces are unaffected.
- The table and column lists in the Query Configurator are read from the cluster's schema tables each time they are requested instead of from a cache kept by the driver, so they always reflect the current schema.
- A `smallint` column can now be used as the first (ID) column of a query, which used to fail with `unsupported type: int16`.
- Driver warnings and errors, such as a node that cannot be reached, are now written to the Grafana server log as structured plugin log entries instead of raw standard-error output. Routine connection messages are logged at debug level only.
