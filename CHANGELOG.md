# v3.1.1

* Updated frontend dependencies to latest compatible versions:
  - @grafana/data, @grafana/runtime, @grafana/ui remain at 10.4.19 (latest 10.x)
  - @grafana/e2e remains at 10.4.12 (latest available for 10.x)
  - @grafana/e2e-selectors remains at 10.4.19
  - @grafana/eslint-config remains at 6.0.1
  - @grafana/tsconfig remains at 2.0.0
  - TypeScript updated to 5.8.3
  - Webpack updated to 5.99.9
  - webpack-cli updated to 6.0.1
  - @swc/helpers updated to 0.5.13
  - @testing-library/jest-dom updated to 6.6.3
  - @types/jest updated to 29.5.14
  - @types/lodash updated to 4.17.13
  - @types/node updated to 22.10.6
  - css-loader updated to 7.1.2
  - sass-loader updated to 16.0.4
  - style-loader updated to 4.0.0
  - glob updated to 11.0.0
  - prettier updated to 3.4.2
  - Other dependencies remain at current stable versions
* Updated backend dependencies to latest compatible versions:
  - Updated Go version from 1.21 to 1.24.1
  - Updated core dependencies:
    - github.com/gocql/gocql: v1.5.2 → v1.7.0
    - github.com/grafana/grafana-plugin-sdk-go: v0.172.0 → v0.277.1
    - github.com/stretchr/testify: v1.8.4 → v1.10.0
  - Notable indirect dependency updates:
    - OpenTelemetry packages updated to v1.36.0
    - gRPC updated to v1.72.1
    - Protocol Buffers updated to v1.36.6
  - Compatibility fixes:
    - Downgraded tablewriter from v1.0.6 to v0.0.5 for SDK compatibility
    - Updated datasource factory function to include context parameter
* Updated vendor directory to sync with go.mod
* Updated Docker Compose Grafana image from 10.1.2 to 10.4.19 (latest 10.x)
* Fixed compatibility issues with missing @grafana/e2e version
* Note: @grafana/e2e is deprecated in favor of @grafana/plugin-e2e for newer Grafana versions

# v3.1.0

* Updated frontend dependencies to latest compatible versions
* Updated @grafana packages to 10.4.18
* Updated testing and build dependencies
* Increased minimum Node.js version to 18

# v3.0.0

**IMPORTANT** v3 supports Grafana versions 7.4+ through 10.x

* Added support for Grafana 10.x
* Enhanced security features including TLS support
* Support for various Cassandra implementations (Apache Cassandra, DataStax Enterprise, DataStax Astra, AWS Keyspaces)
* Modernized plugin architecture with backend and frontend components

# v2.0.0

**IMPORTANT** v2 does NOT support older grafana versions (any version older than 7.0)

* Added support for Grafana 8.x (#89)
* Added Alerting (#91)
* Added table format support (#66)
* Added aliases (#92)
* UX Query Editor Improvements (#93)

All credits to [@futuarmo](https://www.linkedin.com/in/armen-khachkinaev)

# v1.1.4

* Configurable connection timeout
* Configurable TLS setting (allow/disallow self-signed certs)
* UI configuration improvements
* Fronted dependencies update

# v1.0.1

* Supports linux ARM64 platform
* Updated dependencies

# v1.0.0 Initial

* First implementation