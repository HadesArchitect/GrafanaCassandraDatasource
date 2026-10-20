package cassandra

import (
	"testing"

	gocql "github.com/apache/cassandra-gocql-driver/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeIter is a rowScanner over fixed rows that counts how many rows were read.
type fakeIter struct {
	columns []gocql.ColumnInfo
	rows    []map[string]interface{}
	read    int
}

func (f *fakeIter) Columns() []gocql.ColumnInfo { return f.columns }

func (f *fakeIter) MapScan(m map[string]interface{}) bool {
	if f.read >= len(f.rows) {
		return false
	}
	for k, v := range f.rows[f.read] {
		m[k] = v
	}
	f.read++
	return true
}

// sensorIter returns an iterator over one row per ID, in order.
func sensorIter(ids ...string) *fakeIter {
	iter := &fakeIter{columns: []gocql.ColumnInfo{{Name: "sensor_id"}, {Name: "value"}}}
	for i, id := range ids {
		iter.rows = append(iter.rows, map[string]interface{}{"sensor_id": id, "value": float64(i)})
	}
	return iter
}

func rowsPerID(result SelectResult) map[string]int {
	counts := map[string]int{}
	for id, rows := range result.Rows {
		counts[id] = len(rows)
	}
	return counts
}

func Test_scanRows(t *testing.T) {
	testCases := map[string]struct {
		ids           []string
		maxRows       int
		wantCount     int
		wantTruncated bool
		wantPerID     map[string]int
		// wantRead is how many rows were taken from the iterator, including
		// the one read to detect truncation.
		wantRead int
	}{
		"limit counts rows across all IDs": {
			ids: []string{"a", "b", "a", "c", "b", "c"}, maxRows: 3,
			wantCount: 3, wantTruncated: true, wantPerID: map[string]int{"a": 2, "b": 1}, wantRead: 4,
		},
		"exactly at the limit is not truncated": {
			ids: []string{"a", "b", "a"}, maxRows: 3,
			wantCount: 3, wantTruncated: false, wantPerID: map[string]int{"a": 2, "b": 1}, wantRead: 3,
		},
		"one row over the limit is truncated": {
			ids: []string{"a", "b", "a", "b"}, maxRows: 3,
			wantCount: 3, wantTruncated: true, wantPerID: map[string]int{"a": 2, "b": 1}, wantRead: 4,
		},
		"under the limit": {
			ids: []string{"a", "b"}, maxRows: 3,
			wantCount: 2, wantPerID: map[string]int{"a": 1, "b": 1}, wantRead: 2,
		},
		"limit of one": {
			ids: []string{"a", "b"}, maxRows: 1,
			wantCount: 1, wantTruncated: true, wantPerID: map[string]int{"a": 1}, wantRead: 2,
		},
		"zero is unlimited": {
			ids: []string{"a", "b", "a", "c", "b", "c"}, maxRows: 0,
			wantCount: 6, wantPerID: map[string]int{"a": 2, "b": 2, "c": 2}, wantRead: 6,
		},
		"negative is unlimited": {
			ids: []string{"a", "b"}, maxRows: -1,
			wantCount: 2, wantPerID: map[string]int{"a": 1, "b": 1}, wantRead: 2,
		},
		"no rows": {
			ids: nil, maxRows: 3,
			wantCount: 0, wantPerID: map[string]int{}, wantRead: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			iter := sensorIter(tc.ids...)

			result, err := scanRows(iter, tc.maxRows)
			require.NoError(t, err)

			assert.Equal(t, tc.wantCount, result.RowCount)
			assert.Equal(t, tc.wantTruncated, result.Truncated)
			assert.Equal(t, tc.wantPerID, rowsPerID(result))
			assert.Equal(t, tc.wantRead, iter.read, "rows beyond the limit must not be read")
		})
	}
}

func Test_scanRows_keepsRowContent(t *testing.T) {
	iter := sensorIter("a", "b", "a")

	result, err := scanRows(iter, 2)
	require.NoError(t, err)

	assert.Equal(t, map[string][]Row{
		"a": {{Columns: []string{"sensor_id", "value"}, Fields: map[string]interface{}{"sensor_id": "a", "value": 0.0}}},
		"b": {{Columns: []string{"sensor_id", "value"}, Fields: map[string]interface{}{"sensor_id": "b", "value": 1.0}}},
	}, result.Rows)
}

func Test_scanRows_errors(t *testing.T) {
	t.Run("unsupported ID type", func(t *testing.T) {
		iter := &fakeIter{
			columns: []gocql.ColumnInfo{{Name: "id"}},
			rows:    []map[string]interface{}{{"id": []int{1}}},
		}

		_, err := scanRows(iter, 10)
		assert.ErrorContains(t, err, "row processing")
	})

	t.Run("unsupported column type", func(t *testing.T) {
		iter := &fakeIter{
			columns: []gocql.ColumnInfo{{Name: "id"}, {Name: "v"}},
			rows:    []map[string]interface{}{{"id": "a", "v": struct{}{}}},
		}

		_, err := scanRows(iter, 10)
		assert.ErrorContains(t, err, "row.normalize")
	})
}
