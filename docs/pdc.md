# Private Data Source Connect (PDC)

[Private Data Source Connect](https://grafana.com/docs/grafana-cloud/connect-externally-hosted/private-data-source-connect/)
lets Grafana Cloud query a data source that lives in a private network — a
Kubernetes cluster, a VPC, an on-premises data centre — without exposing it to
the internet. You run a small PDC agent next to the database, the agent dials
out to Grafana Cloud, and Grafana tunnels data source traffic back through it.

Under the hood PDC is a SOCKS5 proxy, the same one Grafana calls the *secure
socks proxy* in self-managed installations. This plugin supports it, so the same
switch covers both.

## Configuring in the UI

The **Secure Socks Proxy** switch appears under **Private Data Source Connect**
in the data source settings only when the Grafana instance has the secure socks
proxy configured — on Grafana Cloud that means a PDC network is attached to the
stack. If you do not see it, the proxy is not available to the instance and
there is nothing for the plugin to connect through.

Turn it on, make sure the **Host** field points at the cluster as seen *from the
PDC agent*, then **Save & test**.

## Configuring via provisioning

Set the `enableSecureSocksProxy` key under `jsonData`:

```yaml
apiVersion: 1
datasources:
  - name: "Cassandra (private)"
    type: hadesarchitect-cassandra-datasource
    access: proxy
    url: "10.0.3.17:9042;10.0.3.18:9042"
    jsonData:
      keyspace: system
      consistency: ONE
      user: cassandra
      enableSecureSocksProxy: true
    secureJsonData:
      password: cassandra
```

## Use IP addresses as contact points

The Cassandra driver resolves host names itself, inside the plugin, before any
connection is made — host name resolution does **not** go through the proxy. A
name that only resolves inside your private network (`cassandra.default.svc.cluster.local`,
for instance) therefore fails on Grafana Cloud with:

```text
cluster.CreateSession: gocql: unable to create session: failed to resolve any of the provided hostnames
```

Use the node IP addresses as contact points instead, or a host name that
resolves publicly to the private address. Connections to the addresses the
driver discovers from the cluster itself (`system.peers`) are dialled through
the proxy as usual, so the PDC agent must be able to reach the nodes on their
broadcast addresses.

## TLS

The secure socks proxy and the plugin's own **TLS Settings** are independent and
can be used together: the proxy carries the TCP connection, TLS is negotiated
end to end with the Cassandra node on top of it. The same applies to
[DataStax Astra](connections/astra.md) and
[AWS Keyspaces](connections/aws-keyspaces.md), although those are reachable over
the internet and rarely need a proxy.

## Testing locally

PDC can be simulated with any SOCKS5 server, which is useful when debugging a
connection problem without a Grafana Cloud stack:

```shell
microsocks -i 127.0.0.1 -p 5555
```

Start Grafana with the proxy pointed at it:

```ini
[feature_toggles]
secureSocksDSProxyEnabled = true

[secure_socks_datasource_proxy]
enabled = true
proxy_address = localhost:5555
allow_insecure = true
```

Turn the **Secure Socks Proxy** switch on for the data source and **Save &
test**. With `GF_LOG_LEVEL=debug` the plugin logs
`Connecting to Cassandra through the secure socks proxy` when the proxy is in
use. If it instead logs that the proxy is enabled but not configured on the
instance, Grafana did not pass the proxy settings to the plugin and the
connection was made directly.
