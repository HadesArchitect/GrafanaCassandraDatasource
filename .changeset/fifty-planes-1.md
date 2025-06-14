---
'grafana-cassandra-datasource': minor
---

* Updated frontend dependencies to latest compatible versions
* Updated backend dependencies to latest compatible versions
  - Updated Go version from 1.21 to 1.24.1
  - Updated core dependencies
  - Compatibility fixes:
    - Downgraded tablewriter from v1.0.6 to v0.0.5 for SDK compatibility
    - Updated datasource factory function to include context parameter
* Updated Docker Compose Grafana image from 10.1.2 to 10.4.19 (latest 10.x)
* Updated @grafana packages to 10.4.18
* Increased minimum Node.js version to 18