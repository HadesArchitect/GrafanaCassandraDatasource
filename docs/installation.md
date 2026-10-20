# Data Source Installation

## Requirements

Plugin version 4.x requires **Grafana 10.4 or newer**. Grafana 7.4 - 10.3 are supported by plugin version 3.x: installations that already run it keep working, and version 3.3.0 stays published for them. To install it explicitly, pass the version to the console tool: `grafana-cli plugins install hadesarchitect-cassandra-datasource 3.3.0`.

## Installation options

- **GUI** Install the plugin via standard Grafana plugin management UI

- **Cli** Install the plugin using grafana console tool: `grafana-cli plugins install hadesarchitect-cassandra-datasource`. The plugin will be installed into your grafana plugins directory; the default is `/var/lib/grafana/plugins`.

- **Manual** Download the plugin using [latest release](https://github.com/HadesArchitect/GrafanaCassandraDatasource/releases/latest), download `cassandra-datasource-VERSION.zip` and uncompress a file into the Grafana plugins directory (`grafana/plugins`).

## Test the connection

After adding the data source, fill in the contact points, credentials and other
connection settings, then press **Save & test**:

- On success the message names what Grafana connected to, for example
  `Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1).`,
  followed by a warning when something needs attention, such as a role that can
  read `system_auth`.
- On failure the message says what went wrong, for example
  `Authentication failed: Cassandra rejected the credentials.`, with a hint on
  how to fix it.

Every message is listed in [Troubleshooting](troubleshooting.md), together with
how to set up a dedicated read-only role for Grafana.
