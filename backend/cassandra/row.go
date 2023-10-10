package cassandra

import (
	"net"
	"time"

	"github.com/gocql/gocql"
)

type Row struct {
	Columns []string
	Fields  map[string]interface{}
}

// filterUnsupportedTypes checks the type of returned field and in case if
// it is not supported by grafana tries to convert it to supported type.
// Fields with types that cannot be converted are removed from row. Filtration
// is not stable, so elements could change their position.
// Type mappings are based on these:
// Cassandra gocql types: https://github.com/gocql/gocql/blob/master/marshal.go#L164
// Grafana field types: https://github.com/grafana/grafana-plugin-sdk-go/blob/main/data/field.go#L39
func (r *Row) filterUnsupportedTypes() {
	i, j := 0, len(r.Columns)-1
	for i <= j {
		colName := r.Columns[i]
		switch v := r.Fields[colName].(type) {
		case int8, int16, int32, int64, float32, float64, string, bool, time.Time:
			i++
		case int:
			r.Fields[colName] = int64(v)
			i++
		case []byte:
			r.Fields[colName] = string(v)
			i++
		case net.IP:
			r.Fields[colName] = v.String()
			i++
		case gocql.UUID:
			r.Fields[colName] = v.String()
			i++
		default:
			delete(r.Fields, colName)
			r.Columns[i], r.Columns[j] = r.Columns[j], r.Columns[i]
			j--
		}
	}
	r.Columns = r.Columns[0:i]
}
