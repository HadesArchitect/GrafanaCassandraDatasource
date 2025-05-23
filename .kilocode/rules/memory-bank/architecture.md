# Architecture Overview

## System Architecture

The Grafana Cassandra Data Source Plugin follows a standard Grafana plugin architecture with separate frontend and backend components:

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│   Grafana UI    │────▶│  Plugin Frontend │────▶│ Plugin Backend  │
│   (Browser)     │     │   (TypeScript)   │     │     (Go)        │
└─────────────────┘     └──────────────────┘     └────────┬────────┘
                                                           │
                                                           ▼
                                                  ┌─────────────────┐
                                                  │    Cassandra    │
                                                  │    Database     │
                                                  └─────────────────┘
```

## Frontend Architecture

### Key Components

- **ConfigEditor** ([`src/ConfigEditor.tsx`](src/ConfigEditor.tsx:1)): Configuration UI for datasource settings
  - Connection parameters (host, keyspace, credentials)
  - TLS configuration
  - Consistency level settings

- **QueryEditor** ([`src/QueryEditor.tsx`](src/QueryEditor.tsx:1)): Query building interface
  - Query Configurator mode: UI-based query builder
  - Raw Query mode: Direct CQL editor
  - Support for variables and aliases

- **Datasource** ([`src/datasource.ts`](src/datasource.ts:1)): Frontend datasource implementation
  - Extends [`DataSourceWithBackend`](src/datasource.ts:7)
  - Handles query preparation and template variable substitution
  - Manages communication with backend

- **Models** ([`src/models.ts`](src/models.ts:1)): TypeScript type definitions
  - [`CassandraQuery`](src/models.ts:4): Query structure
  - [`CassandraDataSourceOptions`](src/models.ts:24): Configuration options

### Frontend Technologies
- React 17.0.2
- TypeScript 4.9.5
- @grafana/ui, @grafana/data, @grafana/runtime (v10.4.18)
- Webpack 5 for bundling
- Jest for testing

## Backend Architecture

### Key Components

- **Main Entry** ([`backend/main.go`](backend/main.go:1)): Plugin initialization
  - Creates datasource instances
  - Configures TLS settings
  - Establishes Cassandra connections

- **Plugin Core** ([`backend/plugin/plugin.go`](backend/plugin/plugin.go:1)): Business logic
  - [`ExecQuery`](backend/plugin/plugin.go:39): Executes CQL queries
  - [`makeDataFrames`](backend/plugin/plugin.go:154): Transforms Cassandra results to Grafana data frames
  - Supports both raw CQL and configurator-based queries

- **Cassandra Session** ([`backend/cassandra/session.go`](backend/cassandra/session.go:1)): Database interaction
  - [`Select`](backend/cassandra/session.go:64): Executes SELECT queries with safety checks
  - [`GetKeyspaces`](backend/cassandra/session.go:106), [`GetTables`](backend/cassandra/session.go:126), [`GetColumns`](backend/cassandra/session.go:142): Metadata queries
  - Connection management with gocql driver

- **HTTP Handler** ([`backend/handler/handler.go`](backend/handler/handler.go:1)): Request routing
  - [`queryMetricData`](backend/handler/handler.go:56): Handles query requests
  - Resource endpoints for metadata (keyspaces, tables, columns)
  - Health check implementation

### Backend Technologies
- Go 1.21
- gocql v1.5.2 (Cassandra driver)
- grafana-plugin-sdk-go v0.172.0
- gRPC for plugin communication

## Data Flow

1. **Configuration Phase**
   - User configures connection in [`ConfigEditor`](src/ConfigEditor.tsx:1)
   - Settings stored in Grafana database
   - Backend creates Cassandra session on datasource initialization

2. **Query Execution**
   - User builds query in [`QueryEditor`](src/QueryEditor.tsx:1)
   - Frontend sends query to backend via gRPC
   - Backend [`ExecQuery`](backend/plugin/plugin.go:39) processes request
   - [`Session.Select`](backend/cassandra/session.go:64) executes CQL
   - Results transformed to Grafana data frames
   - Data returned to frontend for visualization

3. **Metadata Queries**
   - Frontend requests available keyspaces/tables/columns
   - Backend queries Cassandra system tables
   - Results cached and returned for UI suggestions

## Key Design Patterns

1. **Plugin Instance Management**: Each datasource configuration creates a separate plugin instance with its own Cassandra connection
2. **Query Safety**: Backend validates all queries are SELECT statements to prevent data modification
3. **Time Series Optimization**: First column in results used as series identifier for efficient grouping
4. **Alerting Support**: Narrow frames converted to wide format for Grafana alerting compatibility
5. **Template Variable Support**: Frontend handles variable substitution before sending to backend

## Security Considerations

- Credentials stored encrypted in Grafana
- TLS support with certificate validation
- Query validation to prevent non-SELECT operations
- Recommendation to use read-only Cassandra users