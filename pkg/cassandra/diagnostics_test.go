package cassandra

import (
	"crypto/tls"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/HadesArchitect/GrafanaCassandraDatasource/pkg/health"
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_newClusterConfig_settingsErrorsAreCategorised(t *testing.T) {
	negative := -1
	zero := 0

	testCases := map[string]struct {
		settings Settings
		want     health.Category
	}{
		"invalid consistency": {Settings{Hosts: []string{"127.0.0.1"}, Consistency: "QORUM"}, health.InvalidConsistency},
		"empty consistency":   {Settings{Hosts: []string{"127.0.0.1"}}, health.InvalidConsistency},
		"negative timeout":    {Settings{Hosts: []string{"127.0.0.1"}, Consistency: "ONE", Timeout: &negative}, health.InvalidTimeout},
		"no hosts":            {Settings{Consistency: "ONE"}, health.InvalidContactPoints},
		"blank hosts":         {Settings{Hosts: []string{"", "  "}, Consistency: "ONE"}, health.InvalidContactPoints},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			cluster, err := newClusterConfig(tc.settings)
			assert.Nil(t, cluster)
			assert.Equal(t, tc.want, health.Classify(err), "got %v", err)
		})
	}

	t.Run("zero timeout is still accepted", func(t *testing.T) {
		_, err := newClusterConfig(Settings{Hosts: []string{"127.0.0.1"}, Consistency: "ONE", Timeout: &zero})
		assert.NoError(t, err)
	})
}

func Test_newClusterConfig_trimsContactPoints(t *testing.T) {
	cluster, err := newClusterConfig(Settings{
		Hosts:       []string{" 10.0.0.1:9042", "", "10.0.0.2:9042 "},
		Consistency: "ONE",
	})

	require.NoError(t, err)
	assert.Equal(t, []string{"10.0.0.1:9042", "10.0.0.2:9042"}, cluster.Hosts)
}

func Test_connectRecorder(t *testing.T) {
	var nilRecorder *connectRecorder
	assert.NoError(t, nilRecorder.lastError())

	r := &connectRecorder{}
	assert.NoError(t, r.lastError())

	refused := errors.New("refused")
	r.ObserveConnect(gocql.ObservedConnect{Err: refused})
	assert.Same(t, refused, r.lastError())

	r.ObserveConnect(gocql.ObservedConnect{})
	assert.NoError(t, r.lastError(), "a successful connection clears the previous failure")
}

// testCluster returns a cluster config for addr that gives up quickly.
func testCluster(t *testing.T, addr string, tlsConfig *tls.Config) *gocql.ClusterConfig {
	t.Helper()

	cluster, err := newClusterConfig(Settings{Hosts: []string{addr}, Consistency: "ONE", TLSConfig: tlsConfig})
	require.NoError(t, err)
	cluster.ConnectTimeout = 300 * time.Millisecond
	cluster.Timeout = 300 * time.Millisecond

	return cluster
}

// listen accepts connections on a loopback port and hands each to handle.
func listen(t *testing.T, handle func(net.Conn)) string {
	t.Helper()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func() {
				defer conn.Close()
				handle(conn)
			}()
		}
	}()

	return ln.Addr().String()
}

// Test_createSession_failuresAreClassified runs the real driver against
// servers that fail in different ways, and checks that the error returned by
// session creation, where the driver flattens the cause into text, still
// classifies correctly thanks to the recorded connection attempt.
func Test_createSession_failuresAreClassified(t *testing.T) {
	t.Run("connection refused", func(t *testing.T) {
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := ln.Addr().String()
		ln.Close()

		session, err := createSession(testCluster(t, addr, nil))
		require.Error(t, err)
		assert.Nil(t, session)
		assert.Equal(t, health.ConnectionRefused, health.Classify(err), "got %v", err)
	})

	t.Run("server never answers", func(t *testing.T) {
		addr := listen(t, func(conn net.Conn) {
			_, _ = conn.Read(make([]byte, 1024))
			time.Sleep(2 * time.Second)
		})

		_, err := createSession(testCluster(t, addr, nil))
		require.Error(t, err)
		assert.Equal(t, health.Timeout, health.Classify(err), "got %v", err)
	})

	t.Run("TLS enabled but the server does not speak TLS", func(t *testing.T) {
		addr := listen(t, func(conn net.Conn) {
			_, _ = conn.Write([]byte("this is not a TLS server\n"))
		})

		_, err := createSession(testCluster(t, addr, &tls.Config{InsecureSkipVerify: true}))
		require.Error(t, err)
		assert.Equal(t, health.TLSNotSupportedByServer, health.Classify(err), "got %v", err)
	})

	t.Run("server closes the connection", func(t *testing.T) {
		addr := listen(t, func(net.Conn) {})

		_, err := createSession(testCluster(t, addr, nil))
		require.Error(t, err)
		assert.Equal(t, health.ConnectionClosed, health.Classify(err), "got %v", err)
	})
}
