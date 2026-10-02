package vecprobe

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
	_ "modernc.org/sqlite/vec" // side-effect: installs sqlite-vec (bundled, CGO-free)
)

// TestVec0KNNZeroCGO proves the vec0 virtual table and KNN search work through
// the app's modernc.org/sqlite driver. Run with CGO_ENABLED=0 to assert the
// zero-CGO property the product depends on.
func TestVec0KNNZeroCGO(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	var ver string
	if err := db.QueryRow("SELECT vec_version()").Scan(&ver); err != nil {
		t.Fatalf("vec_version() failed (vec extension not active?): %v", err)
	}
	t.Logf("sqlite-vec version = %s", ver)

	if _, err := db.Exec(`CREATE VIRTUAL TABLE vec_items USING vec0(embedding float[4])`); err != nil {
		t.Fatalf("create vec0: %v", err)
	}
	seed := []struct {
		id  int
		vec string
	}{
		{1, "[0,0,0,0]"},
		{2, "[1,1,1,1]"},
		{3, "[5,5,5,5]"},
	}
	for _, s := range seed {
		if _, err := db.Exec(`INSERT INTO vec_items(rowid, embedding) VALUES (?, ?)`, s.id, s.vec); err != nil {
			t.Fatalf("insert %d: %v", s.id, err)
		}
	}

	// Query closest to row 2 ([1,1,1,1]).
	var nearest int
	var dist float64
	err = db.QueryRow(
		`SELECT rowid, distance FROM vec_items WHERE embedding MATCH ? AND k = 1 ORDER BY distance`,
		"[0.9,1.0,1.1,1.0]",
	).Scan(&nearest, &dist)
	if err != nil {
		t.Fatalf("knn query: %v", err)
	}
	if nearest != 2 {
		t.Fatalf("expected nearest rowid 2, got %d (distance %v)", nearest, dist)
	}
	t.Logf("KNN nearest rowid = %d, distance = %.4f", nearest, dist)
}

// TestVecDistanceFunction proves the vec_* SQL scalar functions are callable.
func TestVecDistanceFunction(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	var d float64
	err = db.QueryRow(`SELECT vec_distance_l2('[0,0]', '[3,4]')`).Scan(&d)
	if err != nil {
		t.Fatalf("vec_distance_l2: %v", err)
	}
	if d < 4.99 || d > 5.01 {
		t.Fatalf("expected L2 distance ~5.0, got %v", d)
	}
}
