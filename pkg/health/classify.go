package health

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"os"
	"reflect"
	"strings"
	"syscall"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

// Classify maps an error to a Category.
//
// Typed errors are matched with errors.Is and errors.As wherever the driver,
// net, crypto/tls or crypto/x509 packages provide them. The driver flattens a
// few errors into plain strings (session creation wraps its cause with %v, and
// the password authenticator and host resolution return untyped errors); only
// those are matched by message, after every typed check.
func Classify(err error) Category {
	if err == nil {
		return Unknown
	}

	if c, ok := categoryOf(err); ok {
		return c
	}

	if errors.Is(err, ErrKeyspaceNotFound) {
		return KeyspaceNotFound
	}
	if errors.Is(err, gocql.ErrNoHosts) {
		return InvalidContactPoints
	}

	if c, ok := classifyServerError(err); ok {
		return c
	}

	// gocql.PasswordAuthenticator.Challenge returns an untyped error.
	if containsAny(err, "unexpected authenticator") {
		return AuthenticatorNotAllowed
	}

	// Checked before the network errors: a failure of the proxy wraps the
	// network error of the hop that failed.
	if isSecureSocksProxyError(err) {
		return SecureSocksProxy
	}

	if c, ok := classifyTLSError(err); ok {
		return c
	}

	if c, ok := classifyNetworkError(err); ok {
		return c
	}

	var protocolErr gocql.ErrProtocol
	if errors.As(err, &protocolErr) || containsAny(err, "discover protocol version") {
		return UnsupportedProtocol
	}

	if errors.Is(err, gocql.ErrNoConnectionsStarted) || errors.Is(err, gocql.ErrNoConnections) ||
		containsAny(err, gocql.ErrNoConnectionsStarted.Error(), gocql.ErrNoConnections.Error()) {
		return NoConnection
	}

	return Unknown
}

// classifyServerError handles errors returned by Cassandra itself.
func classifyServerError(err error) (Category, bool) {
	var credentials *gocql.RequestErrCredentials
	if errors.As(err, &credentials) {
		return AuthenticationFailed, true
	}

	var unauthorized *gocql.RequestErrUnauthorized
	if errors.As(err, &unauthorized) {
		return Unauthorized, true
	}

	var readTimeout *gocql.RequestErrReadTimeout
	if errors.As(err, &readTimeout) {
		return Timeout, true
	}

	var requestErr gocql.RequestError
	if errors.As(err, &requestErr) {
		switch requestErr.Code() {
		case gocql.ErrCodeCredentials:
			return AuthenticationFailed, true
		case gocql.ErrCodeUnauthorized:
			return Unauthorized, true
		case gocql.ErrCodeReadTimeout, gocql.ErrCodeWriteTimeout:
			return Timeout, true
		}
	}

	return Unknown, false
}

// isSecureSocksProxyError reports whether the error was returned by the SOCKS5
// dialer behind Private Data Source Connect. golang.org/x/net/proxy reports
// every failure, whether reaching the proxy or the proxy reaching the target,
// as a *net.OpError whose Op is "socks <command>".
func isSecureSocksProxyError(err error) bool {
	return walk(err, func(e error) bool {
		opErr, ok := e.(*net.OpError)
		return ok && strings.HasPrefix(opErr.Op, "socks ")
	})
}

// clientCertificateAlerts are the TLS alerts a server sends when it does not
// accept the client certificate, or when it requires one and got none.
var clientCertificateAlerts = map[uint8]bool{
	42:  true, // bad_certificate
	43:  true, // unsupported_certificate
	44:  true, // certificate_revoked
	45:  true, // certificate_expired
	46:  true, // certificate_unknown
	48:  true, // unknown_ca
	49:  true, // access_denied
	116: true, // certificate_required
}

// classifyTLSError handles failures of the TLS handshake.
func classifyTLSError(err error) (Category, bool) {
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return TLSUnknownAuthority, true
	}

	var hostname x509.HostnameError
	if errors.As(err, &hostname) {
		return TLSHostnameMismatch, true
	}

	var invalid x509.CertificateInvalidError
	if errors.As(err, &invalid) {
		return TLSCertificateInvalid, true
	}

	var recordHeader tls.RecordHeaderError
	if errors.As(err, &recordHeader) {
		return TLSNotSupportedByServer, true
	}

	if alert, ok := receivedTLSAlert(err); ok {
		if clientCertificateAlerts[alert] {
			return TLSClientCertRejected, true
		}
		return TLSHandshake, true
	}

	var verification *tls.CertificateVerificationError
	if errors.As(err, &verification) {
		return TLSHandshake, true
	}

	// Remaining crypto/tls errors are untyped and all start with "tls: ".
	if containsAny(err, "tls: ") {
		return TLSHandshake, true
	}

	return Unknown, false
}

// receivedTLSAlert returns the TLS alert sent by the server, if err is one.
//
// crypto/tls reports a received alert as a *net.OpError with Op "remote error"
// wrapping its unexported alert type, a uint8. tls.AlertError, the exported
// equivalent, is only used by QUIC connections but is accepted as well.
func receivedTLSAlert(err error) (uint8, bool) {
	var alertErr tls.AlertError
	if errors.As(err, &alertErr) {
		return uint8(alertErr), true
	}

	var alert uint8
	found := walk(err, func(e error) bool {
		opErr, ok := e.(*net.OpError)
		if !ok || opErr.Op != "remote error" || opErr.Err == nil {
			return false
		}
		v := reflect.ValueOf(opErr.Err)
		if v.Kind() != reflect.Uint8 || v.Type().PkgPath() != "crypto/tls" {
			return false
		}
		alert = uint8(v.Uint())
		return true
	})

	return alert, found
}

// classifyNetworkError handles name resolution, dialling and connection errors.
func classifyNetworkError(err error) (Category, bool) {
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return HostNotFound, true
	}
	// gocql logs and drops DNS errors while resolving contact points and
	// returns these untyped errors instead.
	if containsAny(err, "failed to resolve any of the provided hostnames", "no IP's returned from DNS lookup") {
		return HostNotFound, true
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, os.ErrDeadlineExceeded) ||
		errors.Is(err, gocql.ErrTimeoutNoResponse) {
		return Timeout, true
	}
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return Timeout, true
	}
	if containsAny(err, "no response to connection startup within timeout") {
		return Timeout, true
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return ConnectionRefused, true
	}
	if errors.Is(err, syscall.EHOSTUNREACH) || errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EHOSTDOWN) || errors.Is(err, syscall.ENETDOWN) {
		return HostUnreachable, true
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) || errors.Is(err, syscall.ECONNRESET) ||
		errors.Is(err, syscall.EPIPE) || errors.Is(err, gocql.ErrConnectionClosed) {
		return ConnectionClosed, true
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		return HostUnreachable, true
	}

	return Unknown, false
}

// walk calls fn for err and every error it wraps, depth first, and reports
// whether fn returned true for any of them.
func walk(err error, fn func(error) bool) bool {
	if err == nil {
		return false
	}
	if fn(err) {
		return true
	}

	switch e := err.(type) {
	case interface{ Unwrap() error }:
		return walk(e.Unwrap(), fn)
	case interface{ Unwrap() []error }:
		for _, inner := range e.Unwrap() {
			if walk(inner, fn) {
				return true
			}
		}
	}

	return false
}

func containsAny(err error, substrings ...string) bool {
	msg := err.Error()
	for _, s := range substrings {
		if strings.Contains(msg, s) {
			return true
		}
	}

	return false
}
