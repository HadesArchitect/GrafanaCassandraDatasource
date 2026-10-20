package cassandra

import (
	"fmt"
	"math/big"
	"net"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
)

type Row struct {
	Columns []string
	Fields  map[string]interface{}
}

// normalize checks the type of returned field and in case if
// it is not supported by grafana tries to convert it to a supported type.
// If some field has type that cannot be converted then error is returned.
// Type mappings are based on these:
// Cassandra gocql types: https://pkg.go.dev/github.com/apache/cassandra-gocql-driver/v2#Iter.SliceMap
// Grafana field types: https://github.com/grafana/grafana-plugin-sdk-go/blob/main/data/field.go#L39
func (r *Row) normalize() error {
	for _, colName := range r.Columns {
		switch v := r.Fields[colName].(type) {
		case int8, int16, int32, int64, float32, float64, string, bool, time.Time:
		// nulls are not supported yet, so just in case they will`
		// https://github.com/apache/cassandra-gocql-driver/issues/1248
		case *int, *int8, *int16, *int32, *int64, *float32, *float64, *string, *bool:
		case int:
			r.Fields[colName] = int64(v)
		case []byte:
			r.Fields[colName] = string(v)
		case net.IP:
			// gocql v2 returns inet columns as net.IP, v1 returned a string.
			r.Fields[colName] = ipString(v)
		case gocql.UUID:
			r.Fields[colName] = v.String()
		case *big.Int:
			if v != nil {
				r.Fields[colName] = v.String()
			} else {
				r.Fields[colName] = ""
			}
		default:
			return fmt.Errorf("field %s has unsupported type %T", colName, v)
		}
	}

	return nil
}

// ipString formats an inet value. The driver returns a NULL inet as an empty
// net.IP, which is rendered as an empty string, as it was with gocql v1, rather
// than as net.IP's "<nil>".
func ipString(ip net.IP) string {
	if len(ip) == 0 {
		return ""
	}

	return ip.String()
}
