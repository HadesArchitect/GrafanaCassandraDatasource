# Local Build

## Prerequisites

- Docker and Docker Compose installed
- Make (optional, build can be done directly)

## Build

### Frontend

The frontend is built using webpack and can be built in two ways:

**Using Makefile (recommended, no local Node.js required):**
```bash
make frontend
```

**Using yarn (requires node, yarn):**
```bash
yarn install
yarn build
```

The frontend build outputs to the [`dist/`](../../dist) directory.

### Backend

The backend is built using Go and can be built in two ways:

**Using Docker (recommended):**
```bash
make backend
```

**Using golang (requires local golang and [Mage](https://magefile.org/)):**
```bash
cd pkg && go mod tidy
go mod vendor
cd .. && mage
```

The backend binary is output to [`dist/`](../../dist)

##  Start Grafana with the plugin

```bash
docker compose up
```
This starts:
- Grafana with the plugin loaded
- Cassandra 4 database
- Sample data loader

Grafana will be available at http://localhost:3000 Notice that Cassandra startup can take a few minutes, you can check its status via `docker compose logs cassandra`

### Testing with Specific Grafana Version

You can override the Grafana version using environment variables:

```bash
# Test with a specific version
GRAFANA_VERSION=7.4.0 docker compose up

# Test with Grafana OSS instead of Enterprise
GRAFANA_IMAGE=grafana GRAFANA_VERSION=12.0.0 docker compose up
```

## Docker Compose Configuration

The Docker setup mounts the [`dist/`](../../dist) directory into the Grafana container, allowing hot-reloading during development:

- **Plugin location:** `/var/lib/grafana/plugins/hadesarchitect-cassandra-datasource`
- **Provisioning:** [`provisioning/`](../../provisioning) directory is mounted for datasource configuration
- **Ports:**
  - `3000`: Grafana UI
  - `2345`: Delve debugger (for Go debugging)
  - `9042`: Cassandra (from cassandra service)

### Environment Variables

The Docker Compose setup configures Grafana with:

- `GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS`: Allows loading the unsigned development plugin
- `GF_LOG_FILTERS`: Enables debug logging for the plugin
- `GF_LOG_LEVEL`: Sets log level to debug
- `GF_DEFAULT_APP_MODE`: Sets development mode

## Development Workflow

### Watch Mode for Frontend

For active frontend development with auto-rebuild:

```bash
yarn dev
```

This watches for file changes and rebuilds automatically. Refresh your browser to see changes.

## Stopping the Environment

```bash
# Stop services
docker compose stop

# Stop and remove containers
docker compose down
```

## Verifying the Plugin

After starting Grafana:

1. The plugin should be available immediately
2. A datasource has to be created manually
