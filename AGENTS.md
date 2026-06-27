# AGENTS.md

## Project Overview

This is the **Apache Cassandra Datasource plugin for Grafana** — a Grafana data source plugin that enables visualizing time-series data stored in Apache Cassandra and CQL-compatible databases (DataStax Enterprise, DataStax Astra, AWS Keyspaces, etc.).

The plugin has two main components:

- **Backend** (Golang), path: `pkg/`: Handles Cassandra connections, authentication, TLS, and query execution via the Grafana backend plugin SDK.
- **Frontend** (TypeScript/React), path: `src/`: Provides the Grafana UI, including the Query Configurator, raw CQL Query Editor, Variable Query Editor, and Config Editor.

## Key Features

- Query Configurator (GUI-driven) and raw CQL Query Editor
- Table mode, variables, annotations, alerting, and provisioning support
- Authentication (username/password) and TLS
- Compatible with Grafana 7.4+ and Cassandra 3+

## Building & Running

The [`Makefile`](Makefile) is the primary entry point for building and running the project. All build steps run inside Docker containers — no local Go or Node installation required.

Full build process includes: install dependencies for frontend and backend, build both, start environment with docker compose (`make start`).

Key targets:

| Target | Description |
| --- | --- |
| `make build` | Build the full plugin (frontend + backend) |
| `make build` | Build the full plugin (frontend + backend) |
| `make frontend` | Install frontend deps and build frontend |
| `make backend` | Install backend deps and build backend |
| `make test` | Run all tests |
| `make start` | Launch the dev environment via Docker Compose |
| `make stop` | Stop the dev environment |

Build can be customised with `OS`, `ARCH`, `GOLANG`, and `NODE` variables (e.g. `make build OS=darwin ARCH=arm64`).
