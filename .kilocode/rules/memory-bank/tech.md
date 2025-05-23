# Technology Stack

## Frontend Technologies

### Core Framework
- **React** 17.0.2 - UI component framework
- **TypeScript** 4.9.5 - Type-safe JavaScript
- **Node.js** 18+ - JavaScript runtime (minimum required version)

### Grafana Integration
- **@grafana/ui** 10.4.18 - Grafana UI components
- **@grafana/data** 10.4.18 - Data manipulation utilities
- **@grafana/runtime** 10.4.18 - Runtime utilities
- **@grafana/schema** - Schema definitions

### Build Tools
- **Webpack** 5.90.3 - Module bundler
- **SWC** 1.4.6 - Fast TypeScript/JavaScript compiler
- **Sass** 1.72.0 - CSS preprocessor
- **ESLint** - Code linting with Grafana config

### Testing
- **Jest** 29.7.0 - Testing framework
- **@testing-library/react** 12.1.5 - React testing utilities
- **@testing-library/jest-dom** 6.4.2 - DOM testing matchers

## Backend Technologies

### Core Language & Framework
- **Go** 1.21 - Backend programming language
- **grafana-plugin-sdk-go** 0.172.0 - Grafana plugin SDK

### Database Driver
- **gocql** 1.5.2 - Cassandra driver for Go
  - Supports Apache Cassandra 3.x, 4.x, 5.x
  - Supports DataStax Enterprise 6.x
  - Limited support for AWS Keyspaces and DataStax Astra

### Communication
- **gRPC** - Plugin-to-Grafana communication protocol
- **HTTP** - REST endpoints for metadata queries

## Development Environment

### Containerization
- **Docker** & **Docker Compose** - Development environment setup
  - Grafana container (10.1.2)
  - Cassandra container (4.x)
  - Sample data loader container

### Package Management
- **Yarn** 1.22.19 - Frontend dependency management
- **Go Modules** - Backend dependency management

### Development Scripts (Makefile)
- `make install` - Install all dependencies
- `make build` - Build complete plugin
- `make frontend` - Build frontend only
- `make backend` - Build backend only
- `make test` - Run tests
- `make start` - Launch development environment
- `make stop` - Stop development environment

## Supported Platforms

### Operating Systems
- Linux (primary target)
- macOS (including M1)
- Windows

### Architectures
- amd64 (x86_64)
- arm64

### Grafana Versions
- 7.4+ through 10.x (full support)
- 5.x, 6.x, 7.0-7.3 (deprecated, use plugin v1.x/2.x)

## Configuration Options

### Connection Settings
- Multiple contact points support (semicolon-separated)
- Keyspace selection
- Username/password authentication
- Consistency levels (ONE, TWO, THREE, QUORUM, ALL, LOCAL_QUORUM, EACH_QUORUM, LOCAL_ONE)
- Connection timeout configuration

### TLS/SSL Configuration
- Custom TLS settings toggle
- Certificate paths (cert, root, CA)
- Allow self-signed certificates option
- Required for cloud providers (Astra, AWS Keyspaces)

## Query Capabilities

### Query Modes
1. **Query Configurator** - UI-based query builder
   - Keyspace/table selection
   - Column selection (time, value, ID)
   - ID value filtering
   - ALLOW FILTERING toggle

2. **Raw Query Editor** - Direct CQL input
   - Full CQL support (SELECT only)
   - Template variables: `$__timeFrom`, `$__timeTo`, `$__unixEpochFrom`, `$__unixEpochTo`
   - Alias templating with `{{ column_name }}` syntax

### Features
- Variables support for dynamic dashboards
- Annotations for event marking
- Alerting with wide frame conversion
- Table mode for non-time-series data
- Instant queries (PER PARTITION LIMIT 1)

## Build Output

The plugin produces platform-specific binaries:
- `cassandra-plugin_linux_amd64`
- `cassandra-plugin_darwin_amd64`
- `cassandra-plugin_windows_amd64`
- Additional architectures as needed

All build artifacts are placed in the `dist/` directory along with compiled frontend assets.

## Version Management

When updating the plugin version, it must be synchronized in two places:
1. `package.json` - The `version` field
2. `src/plugin.json` - The `info.version` field

Both files must have the same version number. The `info.updated` field in plugin.json should also be updated to the current date when releasing a new version.