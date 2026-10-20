package cassandra

import (
	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
)

// driverLogger forwards the Cassandra driver's structured log messages to the
// Grafana plugin logger, so they follow the plugin's log level and format.
// gocql v2 discards its log output unless a logger is configured.
//
// The driver logs routine connection lifecycle events (protocol negotiation,
// control connection, session ready) at info level every time a data source
// connects; those are logged at debug level here to keep Grafana's log quiet.
// Warnings and errors are passed through unchanged.
type driverLogger struct {
	logger log.Logger
}

var _ gocql.StructuredLogger = driverLogger{}

func newDriverLogger(logger log.Logger) driverLogger {
	return driverLogger{logger: logger.With("component", "cassandra-driver")}
}

func (l driverLogger) Error(msg string, fields ...gocql.LogField) {
	l.logger.Error(msg, logArgs(fields)...)
}

func (l driverLogger) Warning(msg string, fields ...gocql.LogField) {
	l.logger.Warn(msg, logArgs(fields)...)
}

func (l driverLogger) Info(msg string, fields ...gocql.LogField) {
	l.logger.Debug(msg, logArgs(fields)...)
}

func (l driverLogger) Debug(msg string, fields ...gocql.LogField) {
	l.logger.Debug(msg, logArgs(fields)...)
}

// logArgs converts driver log fields into the key/value pairs the plugin SDK
// logger expects.
func logArgs(fields []gocql.LogField) []interface{} {
	args := make([]interface{}, 0, 2*len(fields))
	for _, field := range fields {
		args = append(args, field.Name, field.Value.Any())
	}

	return args
}
