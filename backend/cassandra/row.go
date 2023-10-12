package cassandra

import (
	"fmt"
	"net"
	"time"

	"github.com/gocql/gocql"
)

type Row struct {
	Columns []string
	Fields  map[string]interface{}
}

// normalize checks the type of returned field and in case if
// it is not supported by grafana tries to convert it to a supported type.
// If some field has type that cannot be converted then error is returned.
// Type mappings are based on these:
// Cassandra gocql types: https://github.com/gocql/gocql/blob/master/marshal.go#L164
// Grafana field types: https://github.com/grafana/grafana-plugin-sdk-go/blob/main/data/field.go#L39
func (r *Row) normalize() error {
	for _, colName := range r.Columns {
		switch v := r.Fields[colName].(type) {
		case int8, int16, int32, int64, float32, float64, string, bool, time.Time:
		case int:
			r.Fields[colName] = int64(v)
		case []byte:
			r.Fields[colName] = string(v)
		case net.IP:
			r.Fields[colName] = v.String()
		case gocql.UUID:
			r.Fields[colName] = v.String()
		default:
			return fmt.Errorf("field %s has unsupported type %T", colName, v)
		}
	}

	return nil
}
