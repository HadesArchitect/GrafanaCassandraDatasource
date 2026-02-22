---
"grafana-cassandra-datasource": patch
---

**Added support for Cassandra VARINT columns (#235)** - Querying a table with a `varint` column caused the plugin to crash with `field value has unsupported type *big.Int`. Cassandra's `varint` type is an arbitrary-precision integer backed by Go's `*big.Int`, which was not handled during row normalisation. The fix adds explicit conversion of `*big.Int` values so that `varint` columns are returned as numeric data instead of producing an error.
