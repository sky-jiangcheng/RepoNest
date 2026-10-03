package vectordb

import (
	"database/sql"
	"reflect"
	"testing"

	_ "modernc.org/sqlite" // driver; vec auto-loads via the db package import in local.go
)

func openMem(t *testing.T) *sql.DB {
	t.Helper()
	d, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return d
}

func TestLocalStore(t *testing.T) {
	st := NewLocal(openMem(t))
	if err := st.Ensure(2); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		id  int64
		vec []float32
	}{{1, []float32{0, 0}}, {2, []float32{1, 1}}, {3, []float32{5, 5}}} {
		if err := st.Upsert(c.id, c.vec); err != nil {
			t.Fatalf("upsert %d: %v", c.id, err)
		}
	}
	ids, err := st.Search([]float32{0.9, 1.0}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(ids, []int64{2}) {
		t.Fatalf("local search = %v, want [2]", ids)
	}
	// Clear empties but keeps the index.
	if err := st.Clear(2); err != nil {
		t.Fatal(err)
	}
	if ids, _ := st.Search([]float32{0, 0}, 5); len(ids) != 0 {
		t.Fatalf("after clear expected none, got %v", ids)
	}
}
