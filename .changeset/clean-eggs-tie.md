---
"grafana-cassandra-datasource": minor
---

**Added variable interpolation support for chained/dependent variables** - Variable queries now support template variable interpolation using `${variable}` syntax, enabling powerful cascading variable dependencies. Users can now create hierarchical variable relationships (e.g., Zone → Location → Sensor) where selecting a value in one variable automatically filters options in dependent variables (see [Variables documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/variables.md)).
