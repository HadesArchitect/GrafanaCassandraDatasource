---
'grafana-cassandra-datasource': minor
---

Added the ability to configure AllowedAuthenticators from frontend to support querying clusters with "custom" authenticators. Frontend string is a semicolon separted array which is split by the backend and passed on to gocql driver.
