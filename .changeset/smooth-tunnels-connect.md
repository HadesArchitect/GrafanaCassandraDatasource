---
'grafana-cassandra-datasource': minor
---

Added support for Grafana's secure socks proxy, the transport behind [Private Data Source Connect (PDC)](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/pdc.md). Cassandra clusters that live in a private network — a Kubernetes cluster, a VPC, an on-premises data centre — can now be queried from Grafana Cloud without being exposed to the internet.

A new **Secure Socks Proxy** switch appears in the connection settings when the Grafana instance has the proxy configured, and the same is available to provisioning as the `enableSecureSocksProxy` key under `jsonData`. When it is off, or when the Grafana instance has no proxy configured, the plugin connects directly as before, so existing data sources are unaffected.

Note that the driver resolves contact point host names inside the plugin rather than through the proxy, so contact points must be IP addresses (or host names that resolve publicly) when the proxy is used.
