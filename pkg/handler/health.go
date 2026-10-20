package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/plugin"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// Warnings added to a successful health check. Grafana (10.x) drops the
// details of a successful check, so warnings are part of the message itself.
const (
	warningKeyspaceMissing = "Warning: the default keyspace does not exist, so queries that do not name a keyspace will fail. " +
		"Check the Keyspace field; keyspace names are case-sensitive."
	warningSystemAuthReadable = "Warning: the Cassandra role can read system_auth, which holds password hashes, " +
		"and every user who can query this data source can read it too. Connect with a dedicated read-only role: " +
		health.DocsURL + "#use-a-read-only-role"
	warningSecureSocksProxyUnavailable = "Warning: Secure Socks Proxy is enabled for this data source, " +
		"but this Grafana instance has no secure socks proxy configured, so Grafana connected directly."
)

// maxFactLength limits how much of a value read from system.local is shown.
const maxFactLength = 64

// CheckHealth is a handle to check database connection status.
//
// Every user with query permission on the data source can run the health
// check, so the result never contains error text: failures are reported with
// the curated message of their health.Category, and the raw error is logged.
func (h *handler) CheckHealth(ctx context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	p, err := h.getPluginInstance(ctx, req.PluginContext)
	if err != nil {
		// The instance factory logged the original error already.
		backend.Logger.Debug("Health check failed to get plugin instance", "Message", err)
		return failedHealthCheck(health.Classify(err)), nil
	}

	report, err := p.CheckHealth(ctx)
	if err != nil {
		backend.Logger.Error("Health check failed to query Cassandra", "Message", err)
		return failedHealthCheck(health.Classify(err)), nil
	}

	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusOk,
		Message: healthyMessage(report),
	}, nil
}

// failedHealthCheck builds the result for a failed check from the curated
// text of the category only. Grafana shows the details message under the
// title of a failed check.
func failedHealthCheck(c health.Category) *backend.CheckHealthResult {
	problem := health.Describe(c)

	details, err := json.Marshal(struct {
		Message string `json:"message"`
	}{Message: problem.Hint})
	if err != nil {
		details = nil
	}

	return &backend.CheckHealthResult{
		Status:      backend.HealthStatusError,
		Message:     problem.Message,
		JSONDetails: details,
	}
}

// healthyMessage describes a working connection, for example
// "Connected to cluster 'Test Cluster' (Cassandra 4.1.12, data center dc1)."
func healthyMessage(r plugin.HealthReport) string {
	var b strings.Builder

	if name := safeFact(r.Cluster.ClusterName); name != "" {
		fmt.Fprintf(&b, "Connected to cluster '%s'", name)
	} else {
		b.WriteString("Connected to Cassandra")
	}

	var facts []string
	if version := safeFact(r.Cluster.ReleaseVersion); version != "" {
		facts = append(facts, "Cassandra "+version)
	}
	if version := safeFact(r.Cluster.DSEVersion); version != "" {
		facts = append(facts, "DSE "+version)
	}
	if dc := safeFact(r.Cluster.DataCenter); dc != "" {
		facts = append(facts, "data center "+dc)
	}
	if len(facts) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(facts, ", "))
	}

	if r.SecureSocksProxy {
		b.WriteString(" through the secure socks proxy (Private Data Source Connect)")
	}
	b.WriteString(".")

	if r.SecureSocksProxyUnavailable {
		b.WriteString(" " + warningSecureSocksProxyUnavailable)
	}
	if r.KeyspaceMissing {
		b.WriteString(" " + warningKeyspaceMissing)
	}
	if r.SystemAuthReadable {
		b.WriteString(" " + warningSystemAuthReadable)
	}

	return b.String()
}

// safeFact makes a value read from the cluster fit for a one-line message:
// control and other non-printable characters are dropped and long values are
// shortened.
func safeFact(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, strings.TrimSpace(s))

	if runes := []rune(s); len(runes) > maxFactLength {
		s = string(runes[:maxFactLength]) + "..."
	}

	return s
}
