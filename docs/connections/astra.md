# Connecting to DataStax Astra

1. Open `DataStax Astra` -> `Your Database` -> `Connect`
1. Create a `Database Administrator` token
1. Download SCB: `Get a Secure Connect Bundle` ([docs]((https://docs.datastax.com/en/astra-db-serverless/databases/secure-connect-bundle.html#download-the-secure-connect-bundle)))
1. Create a new datasource in Grafana, using the following details:
  - Host: specify the host and cql_port values of the config.json file from the SecureConnectBundle. It should look like `1234567890qwerty-eu-central-1.db.astra.datastax.com:29402` IMPORTANT Notice, it has to be the `cql_port`)
  - User: `clientID` of the API Token
  - Password: `secret` of the API Token
  - Enable `Custom TLS Settings`
  - Enable `Certificate input method` -> `Use Content`
  - Public Key Content: Use content of `cert` file from SecureConnectBundle
  - Private Key Content: use `key` file from SecureConnectBundle
  - RootCA Certificate Content: use `ca.crt` file from SecureConnectBundle

(Usage of TLS certificate file paths instead of content is also possible, in this case select `use file paths` and provide with paths accordingly)
