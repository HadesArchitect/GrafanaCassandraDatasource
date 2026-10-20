package cassandra

import (
	"context"
	"errors"
	"fmt"
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/stretchr/testify/assert"
)

func Test_logArgs(t *testing.T) {
	err := errors.New("connection refused")

	args := logArgs([]gocql.LogField{
		gocql.NewLogFieldString("host_id", "abc"),
		gocql.NewLogFieldInt("protocol_version", 5),
		gocql.NewLogFieldBool("up", true),
		gocql.NewLogFieldError("err", err),
	})

	assert.Equal(t, []interface{}{
		"host_id", "abc",
		"protocol_version", int64(5),
		"up", true,
		"err", "connection refused",
	}, args)
}

func Test_logArgs_empty(t *testing.T) {
	assert.Empty(t, logArgs(nil))
}

// recordingLogger is a log.Logger that records the level and message of
// every call.
type recordingLogger struct {
	with  []interface{}
	calls *[]string
}

func (l recordingLogger) record(level, msg string, args ...interface{}) {
	*l.calls = append(*l.calls, fmt.Sprint(level, " ", msg, " ", append(append([]interface{}{}, l.with...), args...)))
}

func (l recordingLogger) Debug(msg string, args ...interface{}) { l.record("debug", msg, args...) }
func (l recordingLogger) Info(msg string, args ...interface{})  { l.record("info", msg, args...) }
func (l recordingLogger) Warn(msg string, args ...interface{})  { l.record("warn", msg, args...) }
func (l recordingLogger) Error(msg string, args ...interface{}) { l.record("error", msg, args...) }
func (l recordingLogger) Level() log.Level                      { return log.Debug }
func (l recordingLogger) FromContext(context.Context) log.Logger {
	return l
}
func (l recordingLogger) With(args ...interface{}) log.Logger {
	return recordingLogger{with: append(append([]interface{}{}, l.with...), args...), calls: l.calls}
}

func Test_driverLogger_levels(t *testing.T) {
	var calls []string
	logger := newDriverLogger(recordingLogger{calls: &calls})

	logger.Debug("d", gocql.NewLogFieldInt("n", 1))
	logger.Info("i")
	logger.Warning("w")
	logger.Error("e", gocql.NewLogFieldString("host", "10.0.0.1"))

	assert.Equal(t, []string{
		"debug d [component cassandra-driver n 1]",
		"debug i [component cassandra-driver]",
		"warn w [component cassandra-driver]",
		"error e [component cassandra-driver host 10.0.0.1]",
	}, calls)
}
