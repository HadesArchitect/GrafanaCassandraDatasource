---
'grafana-cassandra-datasource': major
---

**Grafana 10.0 is now the minimum supported version.** `grafanaDependency` moves from `>=7.4.0` to `>=10.0.0`. This is the only breaking change in this release: the Query Configurator, CQL editor, table mode, variables, annotations, alerting and every provisioning key behave exactly as before, so upgrading needs no migration.

Instances on Grafana 7.4-9.x are unaffected rather than broken — the plugin catalogue keeps serving them 3.3.0, they simply stop receiving new features. Grafana 7.4 through 9.x are all past end of life, and the plugin has been built against the Grafana 10 packages for some time, so the old floor no longer described anything that could be verified.

**Added support for Grafana's secure socks proxy, the transport behind [Private Data Source Connect (PDC)](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/pdc.md).** Cassandra clusters that live in a private network — a Kubernetes cluster, a VPC, an on-premises data centre — can now be queried from Grafana Cloud without being exposed to the internet. PDC itself requires Grafana 10 or newer, which is the other reason the two changes ship together.

A new **Secure Socks Proxy** switch appears in the connection settings when the Grafana instance has the proxy configured, and the same is available to provisioning as the `enableSecureSocksProxy` key under `jsonData`, alongside `secureSocksProxyUsername` for the PDC network ID. When the switch is off, or when the Grafana instance has no proxy configured, the plugin connects directly as before.

Note that the driver resolves contact point host names inside the plugin rather than through the proxy, so contact points must be IP addresses (or host names that resolve publicly) when the proxy is used.
