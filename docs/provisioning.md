# Grafana Provisioning with Cassandra Data Source

This setup allows you to quickly start Grafana with a preconfigured Cassandra data source using Docker Compose.

## Quick Start

1. Create the following files in your project directory:

```docker-compose.yaml
services:
  grafana:
    image: grafana/grafana-oss:main
    ports:
      - "3000:3000"
    volumes:
      - ./datasource:/etc/grafana/provisioning/datasources
    environment:
      GF_PLUGINS_PREINSTALL: hadesarchitect-cassandra-datasource
      GF_LOG_LEVEL: debug
      GF_AUTH_ANONYMOUS_ORG_ROLE: Admin
      GF_AUTH_ANONYMOUS_ENABLED: "true"
      GF_AUTH_BASIC_ENABLED: "false"
      GF_DEFAULT_APP_MODE: development
  cassandra:
    image: cassandra:4
    ports:
      - "9042:9042"
    volumes:
      - /var/lib/cassandra
```

2. Create a `datasource/` directory and add the data source configuration:

### Basic Configuration

```datasource/cassandra.yaml
apiVersion: 1
datasources:
  - name: "Cassandra"
    type: hadesarchitect-cassandra-datasource
    access: proxy
    url: "cassandra:9042"
    jsonData:
      keyspace: system
      consistency: ONE
      user: cassandra
    secureJsonData:
      password: cassandra
```

### TLS Configuration with File Paths

```datasource/cassandra-tls-files.yaml
apiVersion: 1
datasources:
  - name: "Cassandra TLS (Files)"
    type: hadesarchitect-cassandra-datasource
    access: proxy
    url: "cassandra:9042"
    jsonData:
      keyspace: system
      consistency: ONE
      user: cassandra
      useCustomTLS: true
      useCertContent: false
      certPath: "/path/to/client-cert.pem"
      rootPath: "/path/to/client-key.pem"
      caPath: "/path/to/ca-cert.pem"
      allowInsecureTLS: false
    secureJsonData:
      password: cassandra
```

### TLS Configuration with Certificate Content

```datasource/cassandra-tls-content.yaml
apiVersion: 1
datasources:
  - name: "Cassandra TLS (Content)"
    type: hadesarchitect-cassandra-datasource
    access: proxy
    url: "cassandra:9042"
    jsonData:
      keyspace: system
      consistency: ONE
      user: cassandra
      useCustomTLS: true
      useCertContent: true
      allowInsecureTLS: false
    secureJsonData:
      password: cassandra
      certContent: |
        -----BEGIN CERTIFICATE-----
        MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
        BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
        aWRnaXRzIFB0eSBMdGQwHhcNMTcwODI3MjM1NzU5WhcNMTgwODI3MjM1NzU5WjBF
        MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
        ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
        CgKCAQEAuuExKvlfqgE2pqbahyrCV2gOcwrVoJBNGit9HjyTU09RnNFzUDtrS7FgF
        -----END CERTIFICATE-----
      rootContent: |
        -----BEGIN PRIVATE KEY-----
        MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC64TEq+V+qATam
        ptqHKsJXaA5zCtWgkE0aK30ePJNTT1Gc0XNQOtLsWAX6AAOCAQ8AMIIBCgKCAQEA
        uuExKvlfqgE2pqbahyrCV2gOcwrVoJBNGit9HjyTU09RnNFzUDtrS7FgF6AAOCAQ
        8AMIIBCgKCAQEAuuExKvlfqgE2pqbahyrCV2gOcwrVoJBNGit9HjyTU09RnNFzUD
        -----END PRIVATE KEY-----
      caContent: |
        -----BEGIN CERTIFICATE-----
        MIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
        BAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX
        aWRnaXRzIFB0eSBMdGQwHhcNMTcwODI3MjM1NzU5WhcNMTgwODI3MjM1NzU5WjBF
        MQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50
        ZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB
        CgKCAQEAuuExKvlfqgE2pqbahyrCV2gOcwrVoJBNGit9HjyTU09RnNFzUDtrS7FgF
        -----END CERTIFICATE-----
```

3. Start the services:
```bash
docker-compose up -d
```

4. Access Grafana at http://localhost:3000 (no login required)

The Cassandra data source will be automatically configured and ready to use.