---
'grafana-cassandra-datasource': minor
---

Added support for Cassandra clusters that use non-standard authenticators (e.g. LDAP). Previously, connecting to such clusters would fail during the initial handshake with an `unexpected authenticator` error.

A new **Allowed authenticators** field in the connection settings lets you specify which authenticator class names the driver should accept, as a semicolon-separated list. Leaving it empty preserves the original behaviour, so existing data sources are unaffected.
