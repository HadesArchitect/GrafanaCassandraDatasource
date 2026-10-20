---
'grafana-cassandra-datasource': minor
---

**Save & test now tells you what is wrong and how to fix it.** It used to answer every failure with "Error, check Grafana logs for more details", which is no help on Grafana Cloud or behind Private Data Source Connect, where the server log is out of reach. The check now recognises the common causes and shows a specific message with a hint: wrong user or password, an authenticator missing from **Allowed authenticators**, a refused, unreachable or unresolvable contact point, a timeout, a cluster that requires TLS, TLS certificate problems (untrusted CA, wrong host name, expired, client certificate rejected, server not using TLS), a secure socks proxy (PDC) that cannot be reached, and a default keyspace that does not exist. Every message is listed, with what to do about it, in the new [troubleshooting guide](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/troubleshooting.md). The same messages replace the generic error in panels when the data source cannot connect at all.

Settings that can never work are now reported as such instead of as a connection failure: an unknown consistency level, a negative timeout, an empty **Host** field, and a TLS certificate, key or CA certificate that cannot be read or parsed (the message names which one). A negative timeout, which used to silently disable the timeout, is now rejected.

Any user who can query a data source can run its health check, so the messages never include the underlying error text, user names, passwords, certificate contents, file paths or host addresses. The full error is still written to the Grafana server log for administrators.

A successful check now says what it connected to, for example `Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1).`, including the DSE version on DataStax Enterprise and whether the connection goes through the secure socks proxy. It also adds a warning, without failing, when:

- the configured Cassandra role can read `system_auth`, which holds every role's password hash, so that any Grafana user who can query the data source could read it too. Use a dedicated read-only role instead; the troubleshooting guide shows how.
- the default keyspace no longer exists.
- the secure socks proxy is enabled for the data source but not configured on the Grafana instance, so Grafana connected directly.
