# Troubleshooting the connection (Save & test)

**Save & test** on the data source settings page checks that Grafana can connect
to the cluster and run a query. This page lists every message it can show, what
it means and how to fix it. It also explains the warning shown on panels whose
[query results were truncated](#truncated-query-results).

## What the check does

1. Reads the data source settings: contact points, consistency, timeout, max
   rows, TLS certificates and the secure socks proxy.
2. Connects to the contact points and authenticates.
3. Runs `SELECT key FROM system.local`. If this works, the check passes.
4. Collects facts and warnings. None of these steps can make a working
   connection fail:
   - `SELECT * FROM system.local` for the cluster name, Cassandra version, DSE
     version (DataStax Enterprise only) and data center. Services such as AWS
     Keyspaces and Astra return fewer or emulated values; whatever is missing
     is left out of the message.
   - When a default keyspace is configured, whether it exists in
     `system_schema.keyspaces`.
   - `SELECT role FROM system_auth.roles LIMIT 1`, to find out whether the role
     can read password hashes (see [Use a read-only role](#use-a-read-only-role)).
     Only whether the query succeeds is used; the rows are never read or logged.
     If your cluster audits access to `system_auth`, this query appears in the
     audit log.

## Why the messages do not show the actual error

Grafana lets every user who can query a data source run its health check, not
only administrators. Driver errors can contain user names, host addresses, file
paths or parts of server responses, so the plugin never shows them. Instead it
recognises the cause and shows one of the fixed messages below, with a hint on
how to fix it under the title.

The full error is always written to the Grafana server log, at error level,
with the message `Failed to create the data source instance` or
`Health check failed to query Cassandra`. On Grafana Cloud, where the server
log is not available, the messages below are meant to be enough to act on.

The same messages are shown in panels when a data source cannot connect at
all, for example `failed to get plugin instance: ... Authentication failed:
Cassandra rejected the credentials. ...`.

## Successful connection

A working connection reports what it connected to, for example:

```text
Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1).
Connected to cluster 'dse' (Cassandra 4.0.0.6851, DSE 6.8.51, data center DC1).
Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1) through the secure socks proxy (Private Data Source Connect).
Connected to Cassandra.
```

The last form is used when the server does not provide these facts.

### Warnings

The check still passes when one of these warnings is added to the message, but
you should act on it:

| Warning | What to do |
| --- | --- |
| Warning: the Cassandra role can read system_auth, which holds password hashes, and every user who can query this data source can read it too. Connect with a dedicated read-only role. | The role is over-privileged, typically a superuser such as the built-in `cassandra` role. Every Grafana user who can run queries against this data source runs them as this role. Create a dedicated role as described in [Use a read-only role](#use-a-read-only-role) and use it instead. |
| Warning: the default keyspace does not exist, so queries that do not name a keyspace will fail. Check the Keyspace field; keyspace names are case-sensitive. | The keyspace was dropped or renamed after the data source connected. Fix the **Keyspace** field, or leave it empty if your queries always use `keyspace.table`. |
| Warning: Secure Socks Proxy is enabled for this data source, but this Grafana instance has no secure socks proxy configured, so Grafana connected directly. | The data source asks for [Private Data Source Connect](pdc.md) but the Grafana instance has no proxy. The connection worked without it. Turn the switch off, or configure the proxy on the Grafana instance. |

## Error messages

The title of a failed check is one of the messages below. Grafana shows a hint
under the title; the second column repeats it, sometimes with more detail.

### Settings

These problems are found before connecting.

| Message | Meaning and fix |
| --- | --- |
| The data source settings could not be read. | Open the data source settings and save them again. For a provisioned data source, check the `jsonData` section of the provisioning file. |
| No contact point is configured. | Enter at least one contact point in the **Host** field, for example `10.0.0.1:9042`. Separate several contact points with a semicolon. |
| The consistency level is not valid. | Pick a level from the **Consistency** list, for example `ONE`, `LOCAL_ONE`, `QUORUM` or `LOCAL_QUORUM`. For a provisioned data source, set `jsonData.consistency` to one of these names. |
| The timeout must not be negative. | Set the **Timeout** field to a positive number of seconds, or leave it empty to use the default of 11 seconds. A negative value used to silently disable the timeout. |
| Max rows must not be negative. | Set the **Max rows** field to the largest number of rows a query may return, `0` for no limit, or leave it empty to use the default of 100000. For a provisioned data source, set `jsonData.maxRows`. |
| The TLS client certificate or private key could not be loaded. | The public key (certificate) and private key must be a matching PEM-encoded pair. When using file paths, both files must exist on the Grafana server and be readable by it. |
| The TLS root CA certificate could not be loaded. | The root CA certificate must be PEM-encoded. When using a file path, the file must exist on the Grafana server and be readable by it. |

### Network

| Message | Meaning and fix |
| --- | --- |
| Could not connect through the secure socks proxy (Private Data Source Connect). | Check that the PDC agent of the selected network is running and connected to Grafana, that the agent can reach the contact points on the CQL port, and, for a provisioned data source, that `jsonData.secureSocksProxyUsername` is the PDC network ID. See [Private Data Source Connect](pdc.md). |
| The contact point host name could not be resolved. | Check the host names in the **Host** field. Use `host:port` without a scheme such as `tcp://` or `http://`. With the secure socks proxy, host names are resolved by Grafana, not by the PDC agent: use IP addresses or publicly resolvable names. |
| Cassandra refused the connection. | Nothing accepts connections on the configured port. Check that the port is the CQL native transport port (usually 9042) and that Cassandra is running with the native transport enabled. |
| The contact point could not be reached. | Check the address in the **Host** field and the network route and firewall rules between Grafana and the cluster. A cluster in a private network can be reached from Grafana Cloud with [Private Data Source Connect](pdc.md). |
| Connecting to Cassandra timed out. | The node did not answer in time. A firewall that drops packets, a wrong port or an overloaded node are the usual causes. For a slow or distant cluster, raise the **Timeout** setting. |
| Cassandra closed the connection before it was established. | This usually means the cluster requires TLS (client encryption) but TLS is not enabled in the data source, or that the port is not a CQL native transport port. |
| Could not connect to any contact point. | Check the **Host** field and that Cassandra is running and reachable from the Grafana server on the CQL native transport port (usually 9042). |
| The server does not speak a supported CQL protocol version. | Check that the port is a CQL native transport port (usually 9042, not the JMX, Thrift or inter-node port) and that the server is Cassandra 3.0 or newer, or a CQL-compatible database. If the cluster requires TLS, enable **Custom TLS settings**. |

### TLS

| Message | Meaning and fix |
| --- | --- |
| The server's TLS certificate is not signed by a trusted certificate authority. | Provide the CA certificate that signed the server certificate as the root CA certificate in the TLS settings. For testing only, **Allow self-signed certificates** skips the verification. |
| The server's TLS certificate does not match the contact point address. | Connect using a host name or IP address listed in the certificate's subject alternative names, or reissue the certificate for the address in the **Host** field. |
| The server's TLS certificate is not valid. | The certificate may have expired, not be valid yet, or not be allowed for server authentication. Renew the certificate, or check the clock of the Grafana server. |
| Cassandra rejected the TLS client certificate. | The cluster requires a client certificate (`require_client_auth`). Provide a public key and private key in the TLS settings, signed by a CA the cluster trusts and not expired. |
| The server did not answer with TLS. | TLS is enabled in the data source but the server does not use TLS on this port. Disable **Custom TLS settings**, or connect to the port where client encryption is enabled. |
| The TLS handshake with Cassandra failed. | Compare the TLS settings of the data source with the cluster's `client_encryption_options`: protocol versions, cipher suites, and whether a client certificate is required (`require_client_auth`). Depending on the TLS version, a missing client certificate can be reported this way. |

### Authentication and authorization

| Message | Meaning and fix |
| --- | --- |
| Authentication failed: Cassandra rejected the credentials. | Check the user and password in the **Credentials** field. Passwords are case-sensitive. A provisioned password goes in `secureJsonData.password`. |
| The cluster's authenticator is not in the list of allowed authenticators. | Add the authenticator class the cluster uses, for example `org.apache.cassandra.auth.LDAPAuthenticator`, to the **Allowed authenticators** setting. The Grafana server log names the class the cluster announced. See [Custom Authenticators](authenticators.md). |
| Connected, but the Cassandra role is not allowed to run the health check query. | The health check reads `system.local`, which Apache Cassandra lets every role read. If your database restricts it, grant the role `SELECT` on `system.local`. |

### Keyspace

| Message | Meaning and fix |
| --- | --- |
| The default keyspace does not exist. | Check the **Keyspace** field: keyspace names are case-sensitive. Leave it empty if queries always name their keyspace. |

### Anything else

| Message | Meaning and fix |
| --- | --- |
| Could not connect to Cassandra. | The cause was not recognised. A Grafana administrator can find the full error in the Grafana server log. Please [open an issue](https://github.com/HadesArchitect/GrafanaCassandraDatasource/issues) with the log line, with addresses and names removed, so that it can be recognised in a future version. |

## Truncated query results

A panel can show this warning:

```text
Result truncated to 100000 rows (Max rows setting). Narrow the time range or the query, or raise Max rows in the data source settings.
```

The data source's **Max rows** setting (`jsonData.maxRows`, 100000 by default)
caps how many rows a single query may return, counted across all the series of
the query. When a query has more, the plugin stops reading at the limit, does
not fetch the remaining rows from Cassandra, and returns the rows read so far
with this warning. Without the cap, one query without a time filter on a large
table could load millions of rows and exhaust Grafana's memory.

The rows returned are the first ones in the order Cassandra returned them, so a
truncated graph can miss whole series or the end of the time range. To fix it:

- Limit the query to the dashboard time range with `$__timeFrom`/`$__timeTo`
  (see [Query Editor](editor.md#queries-without-a-time-filter)), or narrow the
  dashboard time range.
- Select fewer series, add a `LIMIT` or `PER PARTITION LIMIT`, or aggregate in
  CQL.
- If the query really needs more rows, raise **Max rows** in the data source
  settings. `0` removes the limit; only do this if every query of the data
  source is known to return a bounded number of rows.

**Inspect > Stats** shows the number of rows returned, and **Inspect > Query**
the CQL that was executed. Variable queries are capped by the same setting:
a truncated variable query offers the first values only, without a warning in
the dashboard; the Grafana server log records it. Alert rules evaluate the
truncated rows; the server log records that as well.

## Use a read-only role

The plugin runs every query, from every Grafana user who can query the data
source, as the one Cassandra role configured in the data source. Grafana roles
(Viewer, Editor, Admin) do not limit what that Cassandra role can read. Give
Grafana a dedicated role that can only read what dashboards need:

```sql
CREATE ROLE grafana WITH PASSWORD = 'choose-a-strong-password' AND LOGIN = true;

-- Repeat for every keyspace (or single table) dashboards read from.
GRANT SELECT ON KEYSPACE metrics TO grafana;
-- or, narrower:
-- GRANT SELECT ON TABLE metrics.temperature TO grafana;
```

That is all the plugin needs on Apache Cassandra: every role can read
`system.local` (used by the health check) and `system_schema` (used to list
keyspaces, tables and columns in the Query Configurator) without a grant.

Do not use a superuser such as the built-in `cassandra` role, and do not grant
anything on `system_auth`: it holds the password hash of every role, and a
role that can read it triggers the [warning above](#warnings). Do not grant
`MODIFY`, `CREATE`, `ALTER`, `DROP` or `AUTHORIZE` either: the plugin only
reads.
