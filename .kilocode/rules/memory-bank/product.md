# Product Overview

## What is this?

The Grafana Cassandra Data Source Plugin is a specialized connector that enables Grafana users to visualize time-series data stored in Apache Cassandra databases. It bridges the gap between Cassandra's distributed database architecture and Grafana's powerful visualization capabilities.

## Problem it Solves

Organizations storing time-series data in Cassandra (sensor data, metrics, logs, etc.) need a way to create real-time dashboards and visualizations. This plugin eliminates the need for intermediate data processing or ETL pipelines by providing direct connectivity between Grafana and Cassandra.

## How it Works

1. **Connection**: Users configure the plugin with Cassandra connection details (hosts, keyspace, credentials, TLS settings)
2. **Query Building**: Two modes are available:
   - Query Configurator: Simple UI-based query builder for basic time-series queries
   - Query Editor: Raw CQL editor for complex queries with full Cassandra query capabilities
3. **Data Retrieval**: The backend (Go) executes CQL queries and transforms results into Grafana-compatible data frames
4. **Visualization**: Data is rendered in Grafana panels (graphs, tables, alerts)

## User Experience Goals

- **Easy Setup**: Simple configuration with clear connection parameters
- **Flexible Querying**: Support both novice users (configurator) and power users (raw CQL)
- **Performance**: Efficient data retrieval with proper time filtering
- **Reliability**: Robust error handling and connection management
- **Compatibility**: Support for various Cassandra distributions (Apache Cassandra, DataStax Enterprise, Astra, AWS Keyspaces)

## Key Use Cases

1. **IoT Monitoring**: Visualizing sensor data (temperature, pressure, location)
2. **Application Metrics**: Tracking application performance metrics over time
3. **Log Analysis**: Creating dashboards from time-series log data
4. **Business Analytics**: Visualizing business metrics stored in Cassandra

## Target Users

- DevOps engineers monitoring infrastructure
- Data analysts creating business dashboards
- IoT engineers tracking sensor networks
- Application developers monitoring system performance