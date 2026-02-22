# Grafana Cassandra Datasource — Documentation

## Getting Started

| Doc | What's inside |
| ----- | --------------- |
| [Installation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/installation.md) | GUI, CLI, and manual plugin install |
| [Quick Demo](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/quick-demo.md) | Run the full stack locally with Docker Compose |
| [Provisioning](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/provisioning.md) | Auto-configure the datasource via YAML (basic & TLS) |

## Querying

| Doc | What's inside |
| ----- | --------------- |
| [Query Configurator](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/configurator.md) | Point-and-click column picker — good starting point |
| [Query Editor](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/editor.md) | Full CQL editor; required for UDFs, aggregations, etc. |
| [Table Mode](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/table.md) | Return tabular results instead of time series |
| [Variables](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/variables.md) | Dashboard variables, chained variables, multi-value selects |

## Advanced Topics

| Doc | What's inside |
| ----- | --------------- |
| [Partitions](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/partitions.md) | Fat-partition problem and time-bucketing strategy |
| [Unix Epoch Time](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/unix-epoch.md) | Querying `bigint` timestamps stored as seconds or milliseconds |

## Connections

| Doc | What's inside |
| ----- | --------------- |
| [DataStax Astra](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/connections/astra.md) | Connect via Secure Connect Bundle and TLS tokens |
| [AWS Keyspaces](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/connections/aws-keyspaces.md) | IAM credentials, service endpoints, CA certificate setup |

## Development

| Doc | What's inside |
| ----- | --------------- |
| [Dev Setup](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/development/setup.md) | Local build, test, and lint workflow |
| [Release Process](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/development/release.md) | Versioning, changelog, and publishing steps |

## Questions & Support

Use [GitHub Discussions](https://github.com/HadesArchitect/GrafanaCassandraDatasource/discussions) to ask questions, share ideas, or get help from the community.
