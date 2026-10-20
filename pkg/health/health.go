// Package health turns connection and configuration errors into messages that
// are safe to show in Grafana and tell the user what to fix.
//
// Grafana lets every user with query permission on a data source run its
// health check, not only administrators, so the text shown there must never
// contain driver or Go error strings: they can carry host addresses, user
// names, file paths or fragments of server responses. Every error is therefore
// mapped to one of a fixed set of categories, each with a curated message and
// hint. The raw error is only ever written to the Grafana server log.
package health

import (
	"errors"
	"fmt"
)

// DocsURL is the troubleshooting guide that lists every message below.
const DocsURL = "https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/troubleshooting.md"

// Category is a class of problem with a fixed, user-facing explanation.
type Category int

// Categories are listed roughly in the order a connection attempt meets them:
// settings, network, TLS, authentication, then the session itself.
const (
	Unknown Category = iota

	InvalidSettings
	InvalidContactPoints
	InvalidConsistency
	InvalidTimeout
	InvalidTLSClientCert
	InvalidTLSCA

	SecureSocksProxy
	HostNotFound
	ConnectionRefused
	HostUnreachable
	Timeout
	ConnectionClosed
	NoConnection
	UnsupportedProtocol

	TLSUnknownAuthority
	TLSHostnameMismatch
	TLSCertificateInvalid
	TLSClientCertRejected
	TLSNotSupportedByServer
	TLSHandshake

	AuthenticationFailed
	AuthenticatorNotAllowed
	Unauthorized

	KeyspaceNotFound
)

// Problem is the user-facing description of a Category.
type Problem struct {
	// Message is a one-line summary, shown as the title of the health check result.
	Message string
	// Hint tells the user how to fix the problem.
	Hint string
}

// problems holds the only text that may reach the user for a failed check.
// docs/troubleshooting.md lists the same messages; keep them in sync.
var problems = map[Category]Problem{
	Unknown: {
		Message: "Could not connect to Cassandra.",
		Hint: "The cause was not recognised. A Grafana administrator can find the full error in the Grafana server log. " +
			"See the troubleshooting guide: " + DocsURL,
	},
	InvalidSettings: {
		Message: "The data source settings could not be read.",
		Hint: "Open the data source settings and save them again. " +
			"For a provisioned data source, check the jsonData section of the provisioning file.",
	},
	InvalidContactPoints: {
		Message: "No contact point is configured.",
		Hint: "Enter at least one contact point in the Host field, for example 10.0.0.1:9042. " +
			"Separate several contact points with a semicolon.",
	},
	InvalidConsistency: {
		Message: "The consistency level is not valid.",
		Hint: "Pick a level from the Consistency list, for example ONE, LOCAL_ONE, QUORUM or LOCAL_QUORUM. " +
			"For a provisioned data source, set jsonData.consistency to one of these names.",
	},
	InvalidTimeout: {
		Message: "The timeout must not be negative.",
		Hint: "Set the Timeout field to a positive number of seconds, " +
			"or leave it empty to use the default of 11 seconds.",
	},
	InvalidTLSClientCert: {
		Message: "The TLS client certificate or private key could not be loaded.",
		Hint: "The public key (certificate) and private key must be a matching PEM-encoded pair. " +
			"When using file paths, both files must exist on the Grafana server and be readable by it.",
	},
	InvalidTLSCA: {
		Message: "The TLS root CA certificate could not be loaded.",
		Hint: "The root CA certificate must be PEM-encoded. " +
			"When using a file path, the file must exist on the Grafana server and be readable by it.",
	},
	SecureSocksProxy: {
		Message: "Could not connect through the secure socks proxy (Private Data Source Connect).",
		Hint: "Check that the PDC agent of the selected network is running and connected to Grafana, " +
			"that the agent can reach the contact points on the CQL port, and, for a provisioned data source, " +
			"that jsonData.secureSocksProxyUsername is the PDC network ID.",
	},
	HostNotFound: {
		Message: "The contact point host name could not be resolved.",
		Hint: "Check the host names in the Host field. Use host:port without a scheme such as tcp:// or http://. " +
			"With the secure socks proxy, host names are resolved by Grafana, not by the PDC agent: use IP addresses or publicly resolvable names.",
	},
	ConnectionRefused: {
		Message: "Cassandra refused the connection.",
		Hint: "Nothing accepts connections on the configured port. Check that the port is the CQL native transport port " +
			"(usually 9042) and that Cassandra is running with the native transport enabled.",
	},
	HostUnreachable: {
		Message: "The contact point could not be reached.",
		Hint: "Check the address in the Host field and the network route and firewall rules between Grafana and the cluster. " +
			"A cluster in a private network can be reached from Grafana Cloud with Private Data Source Connect.",
	},
	Timeout: {
		Message: "Connecting to Cassandra timed out.",
		Hint: "The node did not answer in time. A firewall that drops packets, a wrong port or an overloaded node are the usual causes. " +
			"For a slow or distant cluster, raise the Timeout setting.",
	},
	ConnectionClosed: {
		Message: "Cassandra closed the connection before it was established.",
		Hint: "This usually means the cluster requires TLS (client encryption) but TLS is not enabled in the data source, " +
			"or that the port is not a CQL native transport port.",
	},
	NoConnection: {
		Message: "Could not connect to any contact point.",
		Hint: "Check the Host field and that Cassandra is running and reachable from the Grafana server " +
			"on the CQL native transport port (usually 9042).",
	},
	UnsupportedProtocol: {
		Message: "The server does not speak a supported CQL protocol version.",
		Hint: "Check that the port is a CQL native transport port (usually 9042, not the JMX, Thrift or inter-node port) " +
			"and that the server is Cassandra 3.0 or newer, or a CQL-compatible database. " +
			"If the cluster requires TLS, enable Custom TLS settings.",
	},
	TLSUnknownAuthority: {
		Message: "The server's TLS certificate is not signed by a trusted certificate authority.",
		Hint: "Provide the CA certificate that signed the server certificate as the root CA certificate in the TLS settings. " +
			"For testing only, Allow self-signed certificates skips the verification.",
	},
	TLSHostnameMismatch: {
		Message: "The server's TLS certificate does not match the contact point address.",
		Hint: "Connect using a host name or IP address listed in the certificate's subject alternative names, " +
			"or reissue the certificate for the address in the Host field.",
	},
	TLSCertificateInvalid: {
		Message: "The server's TLS certificate is not valid.",
		Hint: "The certificate may have expired, not be valid yet, or not be allowed for server authentication. " +
			"Renew the certificate, or check the clock of the Grafana server.",
	},
	TLSClientCertRejected: {
		Message: "Cassandra rejected the TLS client certificate.",
		Hint: "The cluster requires a client certificate (require_client_auth). Provide a public key and private key " +
			"in the TLS settings, signed by a CA the cluster trusts and not expired.",
	},
	TLSNotSupportedByServer: {
		Message: "The server did not answer with TLS.",
		Hint: "TLS is enabled in the data source but the server does not use TLS on this port. " +
			"Disable Custom TLS settings, or connect to the port where client encryption is enabled.",
	},
	TLSHandshake: {
		Message: "The TLS handshake with Cassandra failed.",
		Hint: "Compare the TLS settings of the data source with the cluster's client_encryption_options: " +
			"protocol versions, cipher suites, and whether a client certificate is required (require_client_auth).",
	},
	AuthenticationFailed: {
		Message: "Authentication failed: Cassandra rejected the credentials.",
		Hint: "Check the user and password in the Credentials field. Passwords are case-sensitive. " +
			"A provisioned password goes in secureJsonData.password.",
	},
	AuthenticatorNotAllowed: {
		Message: "The cluster's authenticator is not in the list of allowed authenticators.",
		Hint: "Add the authenticator class the cluster uses, for example org.apache.cassandra.auth.LDAPAuthenticator, " +
			"to the Allowed authenticators setting. The Grafana server log names the class the cluster announced. " +
			"See https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/authenticators.md",
	},
	Unauthorized: {
		Message: "Connected, but the Cassandra role is not allowed to run the health check query.",
		Hint: "The health check reads system.local, which Apache Cassandra lets every role read. " +
			"If your database restricts it, grant the role SELECT on system.local.",
	},
	KeyspaceNotFound: {
		Message: "The default keyspace does not exist.",
		Hint: "Check the Keyspace field: keyspace names are case-sensitive. " +
			"Leave it empty if queries always name their keyspace.",
	},
}

// Describe returns the user-facing problem for a category.
func Describe(c Category) Problem {
	if p, ok := problems[c]; ok {
		return p
	}

	return problems[Unknown]
}

// Error tags an error with a category. Its Error method returns the original
// message, so it must only be logged, never shown to users: use Public for
// that.
type Error struct {
	Category Category
	Err      error
}

// Wrap tags err with a category. It returns nil when err is nil.
func Wrap(c Category, err error) error {
	if err == nil {
		return nil
	}

	return &Error{Category: c, Err: err}
}

func (e *Error) Error() string { return e.Err.Error() }

func (e *Error) Unwrap() error { return e.Err }

// PublicError is an error whose message is the curated text of its category.
// It deliberately does not wrap the original error, so that nothing upstream
// can print it by accident.
type PublicError struct {
	Category Category
}

// Public classifies err and returns an error that is safe to show to users,
// for example from the data source instance factory, whose errors Grafana
// shows in panels as well as in the health check. It returns nil when err is
// nil. Log the original error before calling it.
func Public(err error) error {
	if err == nil {
		return nil
	}

	return &PublicError{Category: Classify(err)}
}

func (e *PublicError) Error() string {
	p := Describe(e.Category)
	return fmt.Sprintf("%s %s", p.Message, p.Hint)
}

// categoryOf returns the category of the first tagged error in err's chain.
func categoryOf(err error) (Category, bool) {
	var public *PublicError
	if errors.As(err, &public) {
		return public.Category, true
	}

	var tagged *Error
	if errors.As(err, &tagged) {
		return tagged.Category, true
	}

	return Unknown, false
}

// ErrKeyspaceNotFound reports that the configured default keyspace does not exist.
var ErrKeyspaceNotFound = errors.New("keyspace does not exist")
