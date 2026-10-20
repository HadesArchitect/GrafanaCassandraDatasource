package handler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"syscall"
	"testing"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/plugin"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkHealth(t *testing.T, im *instanceManagerMock) *backend.CheckHealthResult {
	t.Helper()

	h := handler{instanceManager: im}
	result, err := h.CheckHealth(context.Background(), &backend.CheckHealthRequest{})
	require.NoError(t, err, "health check problems are reported in the result, not as an error")
	require.NotNil(t, result)

	return result
}

func healthyPlugin(report plugin.HealthReport) *pluginMock {
	return &pluginMock{onCheckHealth: func(context.Context) (plugin.HealthReport, error) { return report, nil }}
}

func failingPlugin(err error) *pluginMock {
	return &pluginMock{onCheckHealth: func(context.Context) (plugin.HealthReport, error) { return plugin.HealthReport{}, err }}
}

func Test_CheckHealth_success(t *testing.T) {
	testCluster := cassandra.ClusterInfo{ClusterName: "Test Cluster", ReleaseVersion: "4.1.12", DataCenter: "dc1"}

	testCases := map[string]struct {
		report plugin.HealthReport
		want   string
	}{
		"cassandra": {
			plugin.HealthReport{Cluster: testCluster},
			"Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1).",
		},
		"datastax enterprise": {
			plugin.HealthReport{Cluster: cassandra.ClusterInfo{
				ClusterName: "dse", ReleaseVersion: "4.0.0.6851", DSEVersion: "6.8.51", DataCenter: "DC1",
			}},
			"Connected to cluster 'dse' (Cassandra 4.0.0.6851, DSE 6.8.51, data center DC1).",
		},
		"aws keyspaces style": {
			plugin.HealthReport{Cluster: cassandra.ClusterInfo{ReleaseVersion: "3.11.2", DataCenter: "us-east-1"}},
			"Connected to Cassandra (Cassandra 3.11.2, data center us-east-1).",
		},
		"no facts available": {
			plugin.HealthReport{},
			"Connected to Cassandra.",
		},
		"secure socks proxy": {
			plugin.HealthReport{Cluster: testCluster, SecureSocksProxy: true},
			"Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1) through the secure socks proxy (Private Data Source Connect).",
		},
		"secure socks proxy unavailable": {
			plugin.HealthReport{Cluster: testCluster, SecureSocksProxyUnavailable: true},
			"Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1). " +
				"Warning: Secure Socks Proxy is enabled for this data source, but this Grafana instance has no secure socks proxy configured, so Grafana connected directly.",
		},
		"missing keyspace": {
			plugin.HealthReport{Cluster: testCluster, KeyspaceMissing: true},
			"Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1). " +
				"Warning: the default keyspace does not exist, so queries that do not name a keyspace will fail. Check the Keyspace field; keyspace names are case-sensitive.",
		},
		"over-privileged role (#245)": {
			plugin.HealthReport{Cluster: testCluster, SystemAuthReadable: true},
			"Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1). " +
				"Warning: the Cassandra role can read system_auth, which holds password hashes, and every user who can query this data source can read it too. " +
				"Connect with a dedicated read-only role: https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/troubleshooting.md#use-a-read-only-role",
		},
		"all warnings": {
			plugin.HealthReport{SecureSocksProxyUnavailable: true, KeyspaceMissing: true, SystemAuthReadable: true},
			"Connected to Cassandra. " + warningSecureSocksProxyUnavailable + " " + warningKeyspaceMissing + " " + warningSystemAuthReadable,
		},
		"cluster facts are sanitised": {
			plugin.HealthReport{Cluster: cassandra.ClusterInfo{
				ClusterName:    "line\none\x00\t" + strings.Repeat("x", 100),
				ReleaseVersion: " 4.1.12\r\n",
			}},
			"Connected to cluster 'lineone" + strings.Repeat("x", 57) + "...' (Cassandra 4.1.12).",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := checkHealth(t, &instanceManagerMock{plugin: healthyPlugin(tc.report)})

			assert.Equal(t, &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: tc.want}, result)
		})
	}
}

// Test_CheckHealth_failure checks the exact result for one error of every
// category, whether it comes from the health check query or from creating
// the data source instance.
func Test_CheckHealth_failure(t *testing.T) {
	for c := health.Unknown; c <= health.KeyspaceNotFound; c++ {
		problem := health.Describe(c)
		want := &backend.CheckHealthResult{
			Status:      backend.HealthStatusError,
			Message:     problem.Message,
			JSONDetails: []byte(fmt.Sprintf(`{"message":%q}`, problem.Hint)),
		}

		t.Run(fmt.Sprintf("category %d from the query", c), func(t *testing.T) {
			err := fmt.Errorf("repo.Ping: %w", health.Wrap(c, errors.New("raw")))
			assert.Equal(t, want, checkHealth(t, &instanceManagerMock{plugin: failingPlugin(err)}))
		})

		t.Run(fmt.Sprintf("category %d from the instance factory", c), func(t *testing.T) {
			err := health.Public(health.Wrap(c, errors.New("raw")))
			assert.Equal(t, want, checkHealth(t, &instanceManagerMock{err: err}))
		})
	}
}

func Test_CheckHealth_failureExamples(t *testing.T) {
	testCases := map[string]struct {
		err         error
		wantMessage string
		wantDetails string
	}{
		"wrong password": {
			fmt.Errorf("repo.Ping: %w", &gocql.RequestErrCredentials{}),
			"Authentication failed: Cassandra rejected the credentials.",
			`{"message":"Check the user and password in the Credentials field. Passwords are case-sensitive. A provisioned password goes in secureJsonData.password."}`,
		},
		"nothing listening": {
			fmt.Errorf("cluster.CreateSession: gocql: unable to create session: x (last connection attempt: %w)",
				fmt.Errorf("dial tcp 10.0.0.1:9042: %w", syscall.ECONNREFUSED)),
			"Cassandra refused the connection.",
			`{"message":"Nothing accepts connections on the configured port. Check that the port is the CQL native transport port (usually 9042) and that Cassandra is running with the native transport enabled."}`,
		},
		"unrecognised error": {
			errors.New("something nobody anticipated"),
			"Could not connect to Cassandra.",
			`{"message":"The cause was not recognised. A Grafana administrator can find the full error in the Grafana server log. See the troubleshooting guide: https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/troubleshooting.md"}`,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			result := checkHealth(t, &instanceManagerMock{plugin: failingPlugin(tc.err)})

			assert.Equal(t, backend.HealthStatusError, result.Status)
			assert.Equal(t, tc.wantMessage, result.Message)
			assert.JSONEq(t, tc.wantDetails, string(result.JSONDetails))
		})
	}
}

// Test_CheckHealth_noLeaks checks that nothing from the raw error reaches the
// result: not credentials, certificate content, file paths or addresses.
func Test_CheckHealth_noLeaks(t *testing.T) {
	const (
		user     = "grafana_reader_7f3a"
		password = "s3cr3t-P@ss-91c2"
		cert     = "-----BEGIN CERTIFICATE-----MIIBszCCAVmgAwIBAgIU-----END CERTIFICATE-----"
		keyPath  = "/etc/grafana/secret-dir/client-7f3a.key"
		host     = "10.13.37.42"
	)
	secrets := []string{user, password, "MIIBszCCAVmgAwIBAgIU", "BEGIN CERTIFICATE", keyPath, host}

	causes := []error{
		errors.New("unrecognised"),
		&gocql.RequestErrCredentials{},
		syscall.ECONNREFUSED,
		errors.New(`unexpected authenticator "com.example.Auth"`),
		errors.New("tls: failed to find any PEM data in certificate input"),
		gocql.ErrNoConnectionsStarted,
	}
	for c := health.Unknown; c <= health.KeyspaceNotFound; c++ {
		causes = append(causes, health.Wrap(c, errors.New("tagged")))
	}

	for i, cause := range causes {
		raw := fmt.Errorf("connect %s:9042 as %s/%s using %s and %s: %w", host, user, password, keyPath, cert, cause)

		for name, im := range map[string]*instanceManagerMock{
			"query":   {plugin: failingPlugin(raw)},
			"factory": {err: health.Public(raw)},
		} {
			t.Run(fmt.Sprintf("%s %d", name, i), func(t *testing.T) {
				result := checkHealth(t, im)
				shown := result.Message + string(result.JSONDetails)
				for _, s := range secrets {
					assert.NotContains(t, shown, s)
				}
			})
		}
	}
}
