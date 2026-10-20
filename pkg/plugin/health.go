package plugin

import (
	"context"
	"fmt"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/cassandra"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// HealthReport describes a working connection.
type HealthReport struct {
	// Cluster holds what could be read from system.local. Fields are empty
	// when the server does not provide them or they could not be read.
	Cluster cassandra.ClusterInfo
	// SecureSocksProxy is true when the connection goes through the secure
	// socks proxy (Private Data Source Connect).
	SecureSocksProxy bool
	// SecureSocksProxyUnavailable is true when the data source asks for the
	// secure socks proxy but Grafana has none, so the connection is direct.
	SecureSocksProxyUnavailable bool
	// KeyspaceMissing is true when the configured default keyspace was
	// checked and does not exist.
	KeyspaceMissing bool
	// SystemAuthReadable is true when the configured role can read
	// system_auth.roles, i.e. it is over-privileged for Grafana.
	SystemAuthReadable bool
}

// CheckHealth checks that the cluster answers a query, then gathers facts and
// warnings for the user. Only the first step can fail: the others are
// best-effort and their errors are logged, never returned, so that a working
// connection is always reported as working.
func (p *Plugin) CheckHealth(ctx context.Context) (HealthReport, error) {
	if err := p.repo.Ping(ctx); err != nil {
		return HealthReport{}, fmt.Errorf("repo.Ping: %w", err)
	}

	report := HealthReport{
		SecureSocksProxy:            p.opts.SecureSocksProxy,
		SecureSocksProxyUnavailable: p.opts.SecureSocksProxyUnavailable,
	}

	info, err := p.repo.ClusterInfo(ctx)
	if err != nil {
		backend.Logger.Warn("Health check could not read cluster information from system.local", "error", err)
	} else {
		report.Cluster = info
	}

	if p.opts.Keyspace != "" {
		exists, err := p.repo.KeyspaceExists(ctx, p.opts.Keyspace)
		if err != nil {
			backend.Logger.Warn("Health check could not verify that the default keyspace exists", "error", err)
		} else {
			report.KeyspaceMissing = !exists
		}
	}

	// See issue #245: a role that can read system_auth exposes every role's
	// password hash to anyone who can query this data source.
	readable, err := p.repo.CanReadSystemAuth(ctx)
	if err != nil {
		backend.Logger.Debug("The configured role cannot read system_auth.roles, as recommended", "error", err)
	}
	report.SystemAuthReadable = err == nil && readable
	if report.SystemAuthReadable {
		backend.Logger.Warn("The configured Cassandra role can read system_auth.roles; use a dedicated read-only role")
	}

	return report, nil
}
