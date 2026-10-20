package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/stretchr/testify/assert"
)

var testClusterInfo = cassandra.ClusterInfo{
	ClusterName:    "Test Cluster",
	ReleaseVersion: "4.1.12",
	DataCenter:     "dc1",
}

// healthyRepo returns a repository mock for a reachable cluster with a
// least-privilege role, and records the keyspaces it was asked about.
func healthyRepo(checkedKeyspaces *[]string) *repositoryMock {
	return &repositoryMock{
		onClusterInfo: func(context.Context) (cassandra.ClusterInfo, error) { return testClusterInfo, nil },
		onKeyspaceExists: func(_ context.Context, keyspace string) (bool, error) {
			*checkedKeyspaces = append(*checkedKeyspaces, keyspace)
			return true, nil
		},
		onCanReadSystemAuth: func(context.Context) (bool, error) {
			return false, errors.New("User grafana has no SELECT permission on <table system_auth.roles>")
		},
	}
}

func TestPlugin_CheckHealth(t *testing.T) {
	t.Run("ping failure is the only error", func(t *testing.T) {
		pingErr := errors.New("gocql: no hosts available in the pool")
		repo := &repositoryMock{onPing: func(context.Context) error { return pingErr }}
		p := New(repo, Options{Keyspace: "metrics"})

		report, err := p.CheckHealth(context.Background())

		assert.ErrorIs(t, err, pingErr)
		assert.Equal(t, HealthReport{}, report)
	})

	t.Run("healthy least-privilege role", func(t *testing.T) {
		var checked []string
		p := New(healthyRepo(&checked), Options{Keyspace: "metrics"})

		report, err := p.CheckHealth(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, HealthReport{Cluster: testClusterInfo}, report)
		assert.Equal(t, []string{"metrics"}, checked)
	})

	t.Run("no keyspace configured skips the keyspace check", func(t *testing.T) {
		var checked []string
		p := New(healthyRepo(&checked), Options{})

		_, err := p.CheckHealth(context.Background())

		assert.NoError(t, err)
		assert.Empty(t, checked)
	})

	t.Run("secure socks proxy state is reported", func(t *testing.T) {
		var checked []string

		report, err := New(healthyRepo(&checked), Options{SecureSocksProxy: true}).CheckHealth(context.Background())
		assert.NoError(t, err)
		assert.True(t, report.SecureSocksProxy)
		assert.False(t, report.SecureSocksProxyUnavailable)

		report, err = New(healthyRepo(&checked), Options{SecureSocksProxyUnavailable: true}).CheckHealth(context.Background())
		assert.NoError(t, err)
		assert.False(t, report.SecureSocksProxy)
		assert.True(t, report.SecureSocksProxyUnavailable)
	})

	t.Run("missing keyspace is a warning, not an error", func(t *testing.T) {
		var checked []string
		repo := healthyRepo(&checked)
		repo.onKeyspaceExists = func(context.Context, string) (bool, error) { return false, nil }

		report, err := New(repo, Options{Keyspace: "metrcs"}).CheckHealth(context.Background())

		assert.NoError(t, err)
		assert.True(t, report.KeyspaceMissing)
	})

	t.Run("keyspace check failure does not warn", func(t *testing.T) {
		var checked []string
		repo := healthyRepo(&checked)
		repo.onKeyspaceExists = func(context.Context, string) (bool, error) { return false, errors.New("unauthorized") }

		report, err := New(repo, Options{Keyspace: "metrics"}).CheckHealth(context.Background())

		assert.NoError(t, err)
		assert.False(t, report.KeyspaceMissing)
	})

	t.Run("cluster information failure keeps the connection healthy", func(t *testing.T) {
		var checked []string
		repo := healthyRepo(&checked)
		repo.onClusterInfo = func(context.Context) (cassandra.ClusterInfo, error) {
			return cassandra.ClusterInfo{ClusterName: "partial"}, errors.New("unconfigured table local")
		}

		report, err := New(repo, Options{}).CheckHealth(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, cassandra.ClusterInfo{}, report.Cluster, "partial results of a failed read are dropped")
	})

	t.Run("readable system_auth is a warning (#245)", func(t *testing.T) {
		var checked []string
		repo := healthyRepo(&checked)
		repo.onCanReadSystemAuth = func(context.Context) (bool, error) { return true, nil }

		report, err := New(repo, Options{}).CheckHealth(context.Background())

		assert.NoError(t, err)
		assert.True(t, report.SystemAuthReadable)
	})

	t.Run("system_auth probe errors mean not readable", func(t *testing.T) {
		for name, probeErr := range map[string]error{
			"unauthorized":             errors.New("User grafana has no SELECT permission on <table system_auth.roles>"),
			"missing on AWS Keyspaces": errors.New("unconfigured table roles"),
			"timeout":                  context.DeadlineExceeded,
		} {
			t.Run(name, func(t *testing.T) {
				var checked []string
				repo := healthyRepo(&checked)
				repo.onCanReadSystemAuth = func(context.Context) (bool, error) { return true, probeErr }

				report, err := New(repo, Options{}).CheckHealth(context.Background())

				assert.NoError(t, err)
				assert.False(t, report.SystemAuthReadable)
			})
		}
	})
}
