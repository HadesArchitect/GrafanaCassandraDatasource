# v3.0.0

## 3.2.0 (2026.02.22)

### Minor Changes

- **Added variable interpolation support for chained/dependent variables** - Variable queries now support template variable interpolation using `${variable}` syntax, enabling powerful cascading variable dependencies. Users can now create hierarchical variable relationships (e.g., Zone → Location → Sensor) where selecting a value in one variable automatically filters options in dependent variables (see [Variables documentation](https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/variables.md)). (issue #182) - Thanks @HadesArchitect

### Patch Changes

- **Improved UI with query mode switcher and contextual documentation links** - Replaced the plain "Toggle editor mode" button with a `RadioButtonGroup` (Configurator / Query Editor) for a clearer mode-switching experience and linked relevant docs page (configurator or editor) and GitHub Discussions.
- **Added support for Cassandra VARINT columns (#235)** - Querying a table with a `varint` column caused the plugin to crash with `field value has unsupported type *big.Int`. Cassandra's `varint` type is an arbitrary-precision integer backed by Go's `*big.Int`, which was not handled during row normalisation. The fix adds explicit conversion of `*big.Int` values so that `varint` columns are returned as numeric data instead of producing an error. (issue #235) - Thanks @HadesArchitect and @arturngomes
- **Fix ConfigEditor** - stop mutating React props directly instead of using proper state updates (issue #230) - Thanks @hugohaggmark

## 3.1.0

### Minor Changes

- d96a689: Added keyspace, table, column caching for faster GUI(@HadesArchitect)
- 532f4e3: Added support for TLS certificate configuration via direct content input alongside existing file path support (#210)
- Made 'toggle editor mode' button more visible to improve UI

### Patch Changes

- d96a689: Fixed #198 (@HadesArchitect)
- bcb51f2: Added frontend tests (@HadesArchitect)
- 74366e7: Added changesets
- Fixed TLS certificate fields names
- Updated frontend dependencies to latest compatible versions:
  - @grafana/data, @grafana/runtime, @grafana/ui remain at 10.4.19 (latest 10.x)
- Updated backend dependencies to latest compatible versions:
  - Updated Go version from 1.21 to 1.24.1
  - Compatibility fixes:
    - Downgraded tablewriter from v1.0.6 to v0.0.5 for SDK compatibility
    - Updated datasource factory function to include context parameter

## v3.0.0

**IMPORTANT** v3 supports Grafana versions 7.4+ through 10.x

- Added support for Grafana 10.x
- Enhanced security features including TLS support
- Support for various Cassandra implementations (Apache Cassandra, DataStax Enterprise, DataStax Astra, AWS Keyspaces)
- Modernized plugin architecture with backend and frontend components

## v2.0.0

**IMPORTANT** v2 does NOT support older grafana versions (any version older than 7.0)

- Added support for Grafana 8.x (#89)
- Added Alerting (#91)
- Added table format support (#66)
- Added aliases (#92)
- UX Query Editor Improvements (#93)

All credits to [@futuarmo](https://www.linkedin.com/in/armen-khachkinaev)

## v1.1.4

- Configurable connection timeout
- Configurable TLS setting (allow/disallow self-signed certs)
- UI configuration improvements
- Fronted dependencies update

## v1.0.1

- Supports linux ARM64 platform
- Updated dependencies

## v1.0.0 Initial

- First implementation
