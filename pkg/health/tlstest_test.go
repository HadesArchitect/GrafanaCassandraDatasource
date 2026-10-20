package health

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"net"
	"reflect"
	"testing"
	"time"
)

// testPKI is a throwaway certificate authority for TLS handshake tests.
type testPKI struct {
	caCert *x509.Certificate
	caKey  *ecdsa.PrivateKey
	pool   *x509.CertPool
}

func newTestPKI(t *testing.T) *testPKI {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}

	pool := x509.NewCertPool()
	pool.AddCert(cert)

	return &testPKI{caCert: cert, caKey: key, pool: pool}
}

// leaf issues a certificate for 127.0.0.1 valid between notBefore and notAfter.
func (p *testPKI) leaf(t *testing.T, notBefore, notAfter time.Time) tls.Certificate {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, p.caCert, &key.PublicKey, p.caKey)
	if err != nil {
		t.Fatal(err)
	}

	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

func (p *testPKI) validLeaf(t *testing.T) tls.Certificate {
	return p.leaf(t, time.Now().Add(-time.Hour), time.Now().Add(time.Hour))
}

// serve accepts one connection on a loopback listener and hands it to handle.
func serve(t *testing.T, handle func(net.Conn)) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		handle(conn)
	}()

	return ln.Addr().String()
}

// serveTLS runs a TLS server for one connection and completes its handshake.
func serveTLS(t *testing.T, cfg *tls.Config) string {
	return serve(t, func(conn net.Conn) {
		server := tls.Server(conn, cfg)
		_ = server.SetDeadline(time.Now().Add(5 * time.Second))
		if err := server.Handshake(); err != nil {
			return
		}
		// Keep the connection open until the client is done reading.
		buf := make([]byte, 1)
		_, _ = server.Read(buf)
	})
}

// clientHandshake connects to addr with cfg, completes the handshake and
// reads once, which is when a TLS 1.3 client learns that the server rejected
// its certificate. It returns the first error met.
func clientHandshake(t *testing.T, addr string, cfg *tls.Config) error {
	t.Helper()

	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	if cfg.ServerName == "" {
		// gocql sets the server name to the contact point address the same way.
		cfg = cfg.Clone()
		cfg.ServerName = "127.0.0.1"
	}
	client := tls.Client(conn, cfg)
	_ = client.SetDeadline(time.Now().Add(5 * time.Second))
	if err := client.Handshake(); err != nil {
		return err
	}

	buf := make([]byte, 1)
	_, err = client.Read(buf)
	return err
}

// alertWithCode returns a value of the same type as alert, crypto/tls's
// unexported alert type, holding code.
func alertWithCode(alert error, code uint8) error {
	v := reflect.New(reflect.TypeOf(alert)).Elem()
	v.SetUint(uint64(code))
	return v.Interface().(error)
}
