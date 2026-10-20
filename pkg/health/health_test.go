package health

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Secrets injected into raw errors by the tests: none may ever reach the
// text shown to users.
const (
	secretUser     = "grafana_reader_7f3a"
	secretPassword = "s3cr3t-P@ss-91c2"
	secretCert     = "MIIBszCCAVmgAwIBAgIUfakecertcontent"
	secretKeyPath  = "/etc/grafana/secret-dir/client-7f3a.key"
	secretHost     = "10.13.37.42"
)

var secrets = []string{secretUser, secretPassword, secretCert, secretKeyPath, secretHost, "BEGIN CERTIFICATE", "PRIVATE KEY"}

// withSecrets wraps err in the kind of context a careless error chain could
// carry: credentials, certificate content, file paths and host addresses.
func withSecrets(err error) error {
	return fmt.Errorf("connect %s:9042 as %s/%s with key %s and cert -----BEGIN CERTIFICATE-----%s-----END CERTIFICATE----- (-----BEGIN PRIVATE KEY-----): %w",
		secretHost, secretUser, secretPassword, secretKeyPath, secretCert, err)
}

func assertNoSecrets(t *testing.T, text string) {
	t.Helper()
	for _, s := range secrets {
		assert.NotContains(t, text, s)
	}
}

// expectedProblems pins the exact text of every category. The same text is
// listed in docs/troubleshooting.md.
var expectedProblems = map[Category]Problem{
	Unknown: {
		"Could not connect to Cassandra.",
		"The cause was not recognised. A Grafana administrator can find the full error in the Grafana server log. See the troubleshooting guide: https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/troubleshooting.md",
	},
	InvalidSettings: {
		"The data source settings could not be read.",
		"Open the data source settings and save them again. For a provisioned data source, check the jsonData section of the provisioning file.",
	},
	InvalidContactPoints: {
		"No contact point is configured.",
		"Enter at least one contact point in the Host field, for example 10.0.0.1:9042. Separate several contact points with a semicolon.",
	},
	InvalidConsistency: {
		"The consistency level is not valid.",
		"Pick a level from the Consistency list, for example ONE, LOCAL_ONE, QUORUM or LOCAL_QUORUM. For a provisioned data source, set jsonData.consistency to one of these names.",
	},
	InvalidTimeout: {
		"The timeout must not be negative.",
		"Set the Timeout field to a positive number of seconds, or leave it empty to use the default of 11 seconds.",
	},
	InvalidTLSClientCert: {
		"The TLS client certificate or private key could not be loaded.",
		"The public key (certificate) and private key must be a matching PEM-encoded pair. When using file paths, both files must exist on the Grafana server and be readable by it.",
	},
	InvalidTLSCA: {
		"The TLS root CA certificate could not be loaded.",
		"The root CA certificate must be PEM-encoded. When using a file path, the file must exist on the Grafana server and be readable by it.",
	},
	SecureSocksProxy: {
		"Could not connect through the secure socks proxy (Private Data Source Connect).",
		"Check that the PDC agent of the selected network is running and connected to Grafana, that the agent can reach the contact points on the CQL port, and, for a provisioned data source, that jsonData.secureSocksProxyUsername is the PDC network ID.",
	},
	HostNotFound: {
		"The contact point host name could not be resolved.",
		"Check the host names in the Host field. Use host:port without a scheme such as tcp:// or http://. With the secure socks proxy, host names are resolved by Grafana, not by the PDC agent: use IP addresses or publicly resolvable names.",
	},
	ConnectionRefused: {
		"Cassandra refused the connection.",
		"Nothing accepts connections on the configured port. Check that the port is the CQL native transport port (usually 9042) and that Cassandra is running with the native transport enabled.",
	},
	HostUnreachable: {
		"The contact point could not be reached.",
		"Check the address in the Host field and the network route and firewall rules between Grafana and the cluster. A cluster in a private network can be reached from Grafana Cloud with Private Data Source Connect.",
	},
	Timeout: {
		"Connecting to Cassandra timed out.",
		"The node did not answer in time. A firewall that drops packets, a wrong port or an overloaded node are the usual causes. For a slow or distant cluster, raise the Timeout setting.",
	},
	ConnectionClosed: {
		"Cassandra closed the connection before it was established.",
		"This usually means the cluster requires TLS (client encryption) but TLS is not enabled in the data source, or that the port is not a CQL native transport port.",
	},
	NoConnection: {
		"Could not connect to any contact point.",
		"Check the Host field and that Cassandra is running and reachable from the Grafana server on the CQL native transport port (usually 9042).",
	},
	UnsupportedProtocol: {
		"The server does not speak a supported CQL protocol version.",
		"Check that the port is a CQL native transport port (usually 9042, not the JMX, Thrift or inter-node port) and that the server is Cassandra 3.0 or newer, or a CQL-compatible database. If the cluster requires TLS, enable Custom TLS settings.",
	},
	TLSUnknownAuthority: {
		"The server's TLS certificate is not signed by a trusted certificate authority.",
		"Provide the CA certificate that signed the server certificate as the root CA certificate in the TLS settings. For testing only, Allow self-signed certificates skips the verification.",
	},
	TLSHostnameMismatch: {
		"The server's TLS certificate does not match the contact point address.",
		"Connect using a host name or IP address listed in the certificate's subject alternative names, or reissue the certificate for the address in the Host field.",
	},
	TLSCertificateInvalid: {
		"The server's TLS certificate is not valid.",
		"The certificate may have expired, not be valid yet, or not be allowed for server authentication. Renew the certificate, or check the clock of the Grafana server.",
	},
	TLSClientCertRejected: {
		"Cassandra rejected the TLS client certificate.",
		"The cluster requires a client certificate (require_client_auth). Provide a public key and private key in the TLS settings, signed by a CA the cluster trusts and not expired.",
	},
	TLSNotSupportedByServer: {
		"The server did not answer with TLS.",
		"TLS is enabled in the data source but the server does not use TLS on this port. Disable Custom TLS settings, or connect to the port where client encryption is enabled.",
	},
	TLSHandshake: {
		"The TLS handshake with Cassandra failed.",
		"Compare the TLS settings of the data source with the cluster's client_encryption_options: protocol versions, cipher suites, and whether a client certificate is required (require_client_auth).",
	},
	AuthenticationFailed: {
		"Authentication failed: Cassandra rejected the credentials.",
		"Check the user and password in the Credentials field. Passwords are case-sensitive. A provisioned password goes in secureJsonData.password.",
	},
	AuthenticatorNotAllowed: {
		"The cluster's authenticator is not in the list of allowed authenticators.",
		"Add the authenticator class the cluster uses, for example org.apache.cassandra.auth.LDAPAuthenticator, to the Allowed authenticators setting. The Grafana server log names the class the cluster announced. See https://github.com/HadesArchitect/GrafanaCassandraDatasource/blob/main/docs/authenticators.md",
	},
	Unauthorized: {
		"Connected, but the Cassandra role is not allowed to run the health check query.",
		"The health check reads system.local, which Apache Cassandra lets every role read. If your database restricts it, grant the role SELECT on system.local.",
	},
	KeyspaceNotFound: {
		"The default keyspace does not exist.",
		"Check the Keyspace field: keyspace names are case-sensitive. Leave it empty if queries always name their keyspace.",
	},
}

func TestDescribe(t *testing.T) {
	assert.Len(t, problems, len(expectedProblems), "every category is pinned by this test")

	for c, want := range expectedProblems {
		assert.Equal(t, want, Describe(c), "category %d", c)
	}

	assert.Equal(t, expectedProblems[Unknown], Describe(Category(9999)), "unknown categories fall back to Unknown")
}

// classifyCases maps every category to errors that must be classified as it.
// Most are produced by the real driver, net, crypto/tls or x/net/proxy code
// in TestClassify_realErrors; these cover the remaining shapes.
func classifyCases(t *testing.T) map[string]struct {
	err  error
	want Category
} {
	authenticatorErr := func() error {
		_, _, err := gocql.PasswordAuthenticator{
			AllowedAuthenticators: []string{"org.apache.cassandra.auth.PasswordAuthenticator"},
		}.Challenge([]byte("org.apache.cassandra.auth.LDAPAuthenticator"))
		require.Error(t, err)
		return err
	}()

	return map[string]struct {
		err  error
		want Category
	}{
		"untagged unknown error": {errors.New("something odd"), Unknown},

		"settings: json": {Wrap(InvalidSettings, errors.New("invalid character")), InvalidSettings},
		"settings: no hosts tagged": {
			Wrap(InvalidContactPoints, gocql.ErrNoHosts), InvalidContactPoints,
		},
		"settings: no hosts from driver": {fmt.Errorf("cluster.CreateSession: %w", gocql.ErrNoHosts), InvalidContactPoints},
		"settings: consistency":          {Wrap(InvalidConsistency, errors.New(`invalid consistency "QORUM"`)), InvalidConsistency},
		"settings: timeout":              {Wrap(InvalidTimeout, errors.New("negative timeout: -1")), InvalidTimeout},
		"settings: client cert":          {Wrap(InvalidTLSClientCert, errors.New("tls: failed to find any PEM data")), InvalidTLSClientCert},
		"settings: CA":                   {Wrap(InvalidTLSCA, errors.New("failed to parse CA certificate content")), InvalidTLSCA},
		"public error keeps category": {
			fmt.Errorf("instanceManager.Get: %w", Public(Wrap(InvalidTLSCA, errors.New("x")))), InvalidTLSCA,
		},
		"tag wins over the wrapped cause": {
			Wrap(SecureSocksProxy, &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}), SecureSocksProxy,
		},

		"keyspace not found": {fmt.Errorf("cluster.CreateSession: %w: %w", gocql.ErrNoConnectionsStarted, ErrKeyspaceNotFound), KeyspaceNotFound},

		"server: bad credentials": {fmt.Errorf("x: %w", &gocql.RequestErrCredentials{}), AuthenticationFailed},
		"server: unauthorized":    {fmt.Errorf("x: %w", &gocql.RequestErrUnauthorized{}), Unauthorized},
		"server: read timeout":    {fmt.Errorf("x: %w", &gocql.RequestErrReadTimeout{}), Timeout},
		"authenticator not allowed": {
			fmt.Errorf("cluster.CreateSession: gocql: unable to create session: %v (last connection attempt: %w)", authenticatorErr, authenticatorErr),
			AuthenticatorNotAllowed,
		},

		"proxy wrapping a refused connection": {
			&net.OpError{Op: "socks connect", Net: "tcp", Err: &net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}},
			SecureSocksProxy,
		},

		"dns error": {&net.OpError{Op: "dial", Err: &net.DNSError{Err: "no such host", Name: secretHost, IsNotFound: true}}, HostNotFound},
		"dns flattened by the driver": {
			errors.New("cluster.CreateSession: gocql: unable to create session: failed to resolve any of the provided hostnames"),
			HostNotFound,
		},

		"context deadline":       {fmt.Errorf("x: %w", context.DeadlineExceeded), Timeout},
		"i/o deadline":           {&net.OpError{Op: "read", Err: os.ErrDeadlineExceeded}, Timeout},
		"driver no response":     {fmt.Errorf("x: %w", gocql.ErrTimeoutNoResponse), Timeout},
		"driver startup":         {errors.New("gocql: no response to connection startup within timeout"), Timeout},
		"refused":                {&net.OpError{Op: "dial", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}, ConnectionRefused},
		"host unreachable":       {&net.OpError{Op: "dial", Err: os.NewSyscallError("connect", syscall.EHOSTUNREACH)}, HostUnreachable},
		"network unreachable":    {&net.OpError{Op: "dial", Err: os.NewSyscallError("connect", syscall.ENETUNREACH)}, HostUnreachable},
		"other dial error":       {&net.OpError{Op: "dial", Err: errors.New("weird")}, HostUnreachable},
		"eof":                    {fmt.Errorf("x: %w", io.EOF), ConnectionClosed},
		"reset":                  {&net.OpError{Op: "read", Err: os.NewSyscallError("read", syscall.ECONNRESET)}, ConnectionClosed},
		"driver closed":          {fmt.Errorf("x: %w", gocql.ErrConnectionClosed), ConnectionClosed},
		"protocol error":         {fmt.Errorf("x: %w", gocql.NewErrProtocol("bad frame")), UnsupportedProtocol},
		"protocol discovery":     {errors.New("gocql: unable to create session: unable to discover protocol version: x"), UnsupportedProtocol},
		"no connections":         {fmt.Errorf("x: %w", gocql.ErrNoConnections), NoConnection},
		"no connections started": {fmt.Errorf("x: %w", gocql.ErrNoConnectionsStarted), NoConnection},

		"tls alert: client cert": {&net.OpError{Op: "remote error", Err: tlsAlert(t, 116)}, TLSClientCertRejected},
		"tls alert: other":       {&net.OpError{Op: "remote error", Err: tlsAlert(t, 70)}, TLSHandshake},
		"tls AlertError":         {fmt.Errorf("x: %w", tls.AlertError(42)), TLSClientCertRejected},
		"tls verification": {
			&tls.CertificateVerificationError{Err: errors.New("x509: something new")}, TLSHandshake,
		},
		"tls untyped": {errors.New("tls: server selected unsupported protocol version 301"), TLSHandshake},
	}
}

// tlsAlert returns crypto/tls's unexported alert value, as found in a
// "remote error" *net.OpError, by provoking a real one.
func tlsAlert(t *testing.T, code uint8) error {
	t.Helper()

	// crypto/tls only exposes the alert type through the errors it returns:
	// recover it from a real handshake, then build one with the wanted code.
	pki := newTestPKI(t)
	addr := serveTLS(t, &tls.Config{
		Certificates: []tls.Certificate{pki.validLeaf(t)},
		ClientAuth:   tls.RequireAnyClientCert,
		MinVersion:   tls.VersionTLS13,
	})
	err := clientHandshake(t, addr, &tls.Config{RootCAs: pki.pool})

	var opErr *net.OpError
	require.True(t, errors.As(err, &opErr), "got %v", err)
	require.Equal(t, "remote error", opErr.Op)

	return alertWithCode(opErr.Err, code)
}

func TestClassify(t *testing.T) {
	for name, tc := range classifyCases(t) {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, Classify(tc.err))
		})
	}

	assert.Equal(t, Unknown, Classify(nil))
}

// TestClassify_realErrors classifies errors produced by real connections.
func TestClassify_realErrors(t *testing.T) {
	pki := newTestPKI(t)

	t.Run("connection refused", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := ln.Addr().String()
		ln.Close()

		_, err = net.DialTimeout("tcp", addr, time.Second)
		require.Error(t, err)
		assert.Equal(t, ConnectionRefused, Classify(err))
	})

	t.Run("secure socks proxy not reachable", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		proxyAddr := ln.Addr().String()
		ln.Close()

		// The same dialer the plugin gets from the SDK for PDC.
		client := proxy.New(&proxy.Options{
			Enabled:   true,
			ClientCfg: &proxy.ClientCfg{ProxyAddress: proxyAddr, AllowInsecure: true},
			Timeouts:  &proxy.TimeoutOptions{Timeout: time.Second},
		})
		dialer, err := client.NewSecureSocksProxyContextDialer()
		require.NoError(t, err)
		contextDialer, ok := dialer.(interface {
			DialContext(ctx context.Context, network, addr string) (net.Conn, error)
		})
		require.True(t, ok)
		_, err = contextDialer.DialContext(context.Background(), "tcp", "10.0.0.1:9042")
		require.Error(t, err)
		assert.Equal(t, SecureSocksProxy, Classify(err), "got %v", err)
	})

	t.Run("unknown authority", func(t *testing.T) {
		addr := serveTLS(t, &tls.Config{Certificates: []tls.Certificate{pki.validLeaf(t)}})
		err := clientHandshake(t, addr, &tls.Config{RootCAs: x509.NewCertPool()})
		require.Error(t, err)
		assert.Equal(t, TLSUnknownAuthority, Classify(err), "got %v", err)
	})

	t.Run("hostname mismatch", func(t *testing.T) {
		addr := serveTLS(t, &tls.Config{Certificates: []tls.Certificate{pki.validLeaf(t)}})
		err := clientHandshake(t, addr, &tls.Config{RootCAs: pki.pool, ServerName: "cassandra.example.com"})
		require.Error(t, err)
		assert.Equal(t, TLSHostnameMismatch, Classify(err), "got %v", err)
	})

	t.Run("expired certificate", func(t *testing.T) {
		expired := pki.leaf(t, time.Now().Add(-48*time.Hour), time.Now().Add(-24*time.Hour))
		addr := serveTLS(t, &tls.Config{Certificates: []tls.Certificate{expired}})
		err := clientHandshake(t, addr, &tls.Config{RootCAs: pki.pool})
		require.Error(t, err)
		assert.Equal(t, TLSCertificateInvalid, Classify(err), "got %v", err)
	})

	t.Run("server does not speak TLS", func(t *testing.T) {
		addr := serve(t, func(conn net.Conn) {
			_, _ = conn.Write([]byte("this is not a TLS server\n"))
		})
		err := clientHandshake(t, addr, &tls.Config{RootCAs: pki.pool})
		require.Error(t, err)
		assert.Equal(t, TLSNotSupportedByServer, Classify(err), "got %v", err)
	})

	// A TLS 1.3 server says certificate_required. Over TLS 1.2 there is no
	// such alert and a server answers handshake_failure, which can mean other
	// things too: the generic TLS hint mentions client certificates.
	for name, tc := range map[string]struct {
		version uint16
		want    Category
	}{
		"TLS 1.2": {tls.VersionTLS12, TLSHandshake},
		"TLS 1.3": {tls.VersionTLS13, TLSClientCertRejected},
	} {
		t.Run("client certificate required, "+name, func(t *testing.T) {
			addr := serveTLS(t, &tls.Config{
				Certificates: []tls.Certificate{pki.validLeaf(t)},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    pki.pool,
				MinVersion:   tc.version,
				MaxVersion:   tc.version,
			})
			err := clientHandshake(t, addr, &tls.Config{RootCAs: pki.pool})
			require.Error(t, err)
			assert.Equal(t, tc.want, Classify(err), "got %v", err)
		})
	}

	t.Run("client certificate from an untrusted CA", func(t *testing.T) {
		other := newTestPKI(t)
		addr := serveTLS(t, &tls.Config{
			Certificates: []tls.Certificate{pki.validLeaf(t)},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    pki.pool,
		})
		err := clientHandshake(t, addr, &tls.Config{
			RootCAs:      pki.pool,
			Certificates: []tls.Certificate{other.validLeaf(t)},
		})
		require.Error(t, err)
		assert.Equal(t, TLSClientCertRejected, Classify(err), "got %v", err)
	})
}

func TestPublic(t *testing.T) {
	assert.Nil(t, Public(nil))

	err := Public(withSecrets(Wrap(InvalidTLSClientCert, errors.New("tls: failed to find any PEM data in key input"))))

	var public *PublicError
	require.True(t, errors.As(err, &public))
	assert.Equal(t, InvalidTLSClientCert, public.Category)
	assert.Equal(t,
		"The TLS client certificate or private key could not be loaded. The public key (certificate) and private key must be a matching PEM-encoded pair. When using file paths, both files must exist on the Grafana server and be readable by it.",
		err.Error())
	assert.Nil(t, errors.Unwrap(err), "the original error is not reachable from a public error")
}

func TestWrap(t *testing.T) {
	assert.Nil(t, Wrap(Timeout, nil))

	cause := errors.New("raw")
	err := Wrap(Timeout, cause)
	assert.ErrorIs(t, err, cause)
	assert.Equal(t, "raw", err.Error(), "a tagged error keeps its original text, for the log")
}

// TestNoSecretsReachUsers injects credentials, certificate content, paths and
// addresses into every kind of error and checks that none of it survives into
// the user-facing text.
func TestNoSecretsReachUsers(t *testing.T) {
	for name, tc := range classifyCases(t) {
		t.Run(name, func(t *testing.T) {
			err := withSecrets(tc.err)
			c := Classify(err)
			assert.Equal(t, tc.want, c, "context around the error does not change its category")

			p := Describe(c)
			assertNoSecrets(t, p.Message)
			assertNoSecrets(t, p.Hint)
			assertNoSecrets(t, Public(err).Error())
		})
	}

	for c, p := range problems {
		assertNoSecrets(t, p.Message)
		assertNoSecrets(t, p.Hint)
		assert.False(t, strings.Contains(p.Message+p.Hint, "%"), "category %d text is not a format string", c)
	}
}
