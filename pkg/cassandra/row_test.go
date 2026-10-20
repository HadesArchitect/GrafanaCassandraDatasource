package cassandra

import (
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
)

func TestRow_normalize(t *testing.T) {
	testCases := []struct {
		name    string
		input   *Row
		want    *Row
		wantErr error
	}{
		{
			name:    "empty",
			input:   &Row{},
			want:    &Row{},
			wantErr: nil,
		},
		{
			name: "normal row",
			input: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			wantErr: nil,
		},
		{
			name: "normal row with nil in the beginning",
			input: &Row{
				Columns: []string{"field1", "id", "time", "value"},
				Fields:  map[string]interface{}{"field1": nil, "id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"field1", "id", "time", "value"},
				Fields:  map[string]interface{}{"field1": nil, "id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", nil),
		},
		{
			name: "normal row with nil in the end",
			input: &Row{
				Columns: []string{"id", "time", "value", "field1"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1, "field1": nil},
			},
			want: &Row{
				Columns: []string{"id", "time", "value", "field1"},
				Fields:  map[string]interface{}{"id": "id", "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1, "field1": nil},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", nil),
		},
		{
			name: "normal row with unsupported",
			input: &Row{
				Columns: []string{"id", "field1", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "field1": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			want: &Row{
				Columns: []string{"id", "field1", "time", "value"},
				Fields:  map[string]interface{}{"id": "id", "field1": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "value": 0.1},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", struct{}{}),
		},
		{
			name: "normal row with multiple unsupported fields",
			input: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": struct{}{}, "id": "id", "field2": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "field3": struct{}{}, "value": 0.1},
			},
			want: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value"},
				Fields:  map[string]interface{}{"field1": struct{}{}, "id": "id", "field2": struct{}{}, "time": time.UnixMilli(1257894000000).UTC(), "field3": struct{}{}, "value": 0.1},
			},
			wantErr: fmt.Errorf("field %s has unsupported type %T", "field1", struct{}{}),
		},
		{
			name: "normal row with conversion",
			input: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value", "field4"},
				Fields:  map[string]interface{}{"field1": gocql.UUID{}, "id": "id", "field2": net.ParseIP("127.0.0.1"), "time": time.UnixMilli(1257894000000).UTC(), "field3": []byte("some string"), "value": 0.1, "field4": big.NewInt(12345)},
			},
			want: &Row{
				Columns: []string{"field1", "id", "field2", "time", "field3", "value", "field4"},
				Fields:  map[string]interface{}{"field1": "00000000-0000-0000-0000-000000000000", "id": "id", "field2": "127.0.0.1", "time": time.UnixMilli(1257894000000).UTC(), "field3": "some string", "value": 0.1, "field4": "12345"},
			},
			wantErr: nil,
		},
		{
			name: "row with large varint that overflows int64",
			input: &Row{
				Columns: []string{"id", "large_number"},
				Fields:  map[string]interface{}{"id": "test", "large_number": new(big.Int).SetBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF})},
			},
			want: &Row{
				Columns: []string{"id", "large_number"},
				Fields:  map[string]interface{}{"id": "test", "large_number": "4722366482869645213695"},
			},
			wantErr: nil,
		},
		{
			name: "row with nil big.Int",
			input: &Row{
				Columns: []string{"id", "varint_field"},
				Fields:  map[string]interface{}{"id": "test", "varint_field": (*big.Int)(nil)},
			},
			want: &Row{
				Columns: []string{"id", "varint_field"},
				Fields:  map[string]interface{}{"id": "test", "varint_field": ""},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.input.normalize()
			if tc.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.wantErr.Error())
			}
			assert.Equal(t, tc.want, tc.input)
		})
	}
}

// mapScanValue decodes a serialized column value exactly as Iter.MapScan does
// in gocql v2: into an interface{} that holds the zero value of the column type.
func mapScanValue(t *testing.T, typ gocql.Type, data []byte) interface{} {
	t.Helper()

	info := gocql.NewNativeType(4, typ, "")
	value := info.Zero()
	if err := gocql.Unmarshal(info, data, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", typeName(typ), err)
	}

	return value
}

func marshal(t *testing.T, typ gocql.Type, value interface{}) []byte {
	t.Helper()

	data, err := gocql.Marshal(gocql.NewNativeType(4, typ, ""), value)
	if err != nil {
		t.Fatalf("marshal %s: %v", typeName(typ), err)
	}

	return data
}

// TestRow_normalize_driverValues checks every CQL type as the driver hands it
// over from MapScan, so that changes in the Go types the driver returns are
// caught here rather than in a dashboard.
func TestRow_normalize_driverValues(t *testing.T) {
	ts := time.UnixMilli(1257894000123).UTC()
	day := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	id := gocql.MustRandomUUID()
	timeID := gocql.UUIDFromTime(ts)
	bigVarint, _ := new(big.Int).SetString("123456789012345678901234567890", 10)

	testCases := []struct {
		name    string
		typ     gocql.Type
		data    []byte
		want    interface{}
		wantErr string
	}{
		{name: "ascii", typ: gocql.TypeAscii, data: marshal(t, gocql.TypeAscii, "abc"), want: "abc"},
		{name: "text", typ: gocql.TypeText, data: marshal(t, gocql.TypeText, "héllo"), want: "héllo"},
		{name: "varchar", typ: gocql.TypeVarchar, data: marshal(t, gocql.TypeVarchar, "x"), want: "x"},
		{name: "blob", typ: gocql.TypeBlob, data: marshal(t, gocql.TypeBlob, []byte("raw")), want: "raw"},
		{name: "int", typ: gocql.TypeInt, data: marshal(t, gocql.TypeInt, 42), want: int64(42)},
		{name: "bigint", typ: gocql.TypeBigInt, data: marshal(t, gocql.TypeBigInt, int64(1)<<40), want: int64(1) << 40},
		{name: "counter", typ: gocql.TypeCounter, data: marshal(t, gocql.TypeCounter, int64(7)), want: int64(7)},
		{name: "smallint", typ: gocql.TypeSmallInt, data: marshal(t, gocql.TypeSmallInt, int16(-300)), want: int16(-300)},
		{name: "tinyint", typ: gocql.TypeTinyInt, data: marshal(t, gocql.TypeTinyInt, int8(-5)), want: int8(-5)},
		{name: "varint", typ: gocql.TypeVarint, data: marshal(t, gocql.TypeVarint, bigVarint), want: "123456789012345678901234567890"},
		{name: "float", typ: gocql.TypeFloat, data: marshal(t, gocql.TypeFloat, float32(1.5)), want: float32(1.5)},
		{name: "double", typ: gocql.TypeDouble, data: marshal(t, gocql.TypeDouble, 2.25), want: 2.25},
		{name: "boolean", typ: gocql.TypeBoolean, data: marshal(t, gocql.TypeBoolean, true), want: true},
		{name: "timestamp", typ: gocql.TypeTimestamp, data: marshal(t, gocql.TypeTimestamp, ts), want: ts},
		{name: "date", typ: gocql.TypeDate, data: marshal(t, gocql.TypeDate, day), want: day},
		{name: "uuid", typ: gocql.TypeUUID, data: marshal(t, gocql.TypeUUID, id), want: id.String()},
		{name: "timeuuid", typ: gocql.TypeTimeUUID, data: marshal(t, gocql.TypeTimeUUID, timeID), want: timeID.String()},
		{name: "inet IPv4", typ: gocql.TypeInet, data: marshal(t, gocql.TypeInet, net.ParseIP("10.0.0.1")), want: "10.0.0.1"},
		{name: "inet IPv6", typ: gocql.TypeInet, data: marshal(t, gocql.TypeInet, net.ParseIP("2001:db8::1")), want: "2001:db8::1"},

		// NULL values come back as the zero value of the Go type.
		{name: "NULL text", typ: gocql.TypeText, data: nil, want: ""},
		{name: "NULL int", typ: gocql.TypeInt, data: nil, want: int64(0)},
		{name: "NULL double", typ: gocql.TypeDouble, data: nil, want: float64(0)},
		{name: "NULL varint", typ: gocql.TypeVarint, data: nil, want: ""},
		{name: "NULL uuid", typ: gocql.TypeUUID, data: nil, want: "00000000-0000-0000-0000-000000000000"},
		{name: "NULL inet", typ: gocql.TypeInet, data: nil, want: ""},
		{name: "NULL timestamp", typ: gocql.TypeTimestamp, data: nil, want: time.Time{}},

		// Types Grafana fields cannot hold; unsupported before the driver upgrade too.
		{name: "decimal", typ: gocql.TypeDecimal, data: []byte{0, 0, 0, 2, 0x30, 0x39}, wantErr: "field value has unsupported type *inf.Dec"},
		{name: "time", typ: gocql.TypeTime, data: marshal(t, gocql.TypeTime, time.Hour), wantErr: "field value has unsupported type time.Duration"},
		{name: "duration", typ: gocql.TypeDuration, data: marshal(t, gocql.TypeDuration, gocql.Duration{Days: 1}), wantErr: "field value has unsupported type gocql.Duration"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			row := &Row{
				Columns: []string{"value"},
				Fields:  map[string]interface{}{"value": mapScanValue(t, tc.typ, tc.data)},
			}

			err := row.normalize()
			if tc.wantErr != "" {
				assert.EqualError(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.want, row.Fields["value"])
		})
	}
}
