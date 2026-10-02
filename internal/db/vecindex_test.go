package db

import (
	"reflect"
	"testing"
)

func TestVecIndexLifecycle(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	if VectorIndexReady(database) {
		t.Fatal("index should not exist initially")
	}
	if err := EnsureVectorIndex(database, 2); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if !VectorIndexReady(database) {
		t.Fatal("index should exist after ensure")
	}
	if dim, ok := storedVectorDim(database); !ok || dim != 2 {
		t.Fatalf("dim = %d ok=%v, want 2", dim, ok)
	}

	// Idempotent re-ensure keeps the table.
	if err := EnsureVectorIndex(database, 2); err != nil {
		t.Fatalf("re-ensure: %v", err)
	}

	for _, c := range []struct {
		id  int64
		vec []float32
	}{{1, []float32{0, 0}}, {2, []float32{1, 1}}, {3, []float32{5, 5}}} {
		if err := PutNoteEmbedding(database, c.id, c.vec); err != nil {
			t.Fatalf("put %d: %v", c.id, err)
		}
	}

	ids, err := KnnNoteIDs(database, []float32{0.9, 1.0}, 1)
	if err != nil {
		t.Fatalf("knn: %v", err)
	}
	if !reflect.DeepEqual(ids, []int64{2}) {
		t.Fatalf("knn top-1 = %v, want [2]", ids)
	}
	ids2, _ := KnnNoteIDs(database, []float32{0.9, 1.0}, 2)
	if len(ids2) != 2 || ids2[0] != 2 {
		t.Fatalf("knn top-2 = %v, want [2,...]", ids2)
	}

	if err := DeleteNoteEmbedding(database, 2); err != nil {
		t.Fatalf("delete: %v", err)
	}
	after, _ := KnnNoteIDs(database, []float32{0.9, 1.0}, 3)
	for _, id := range after {
		if id == 2 {
			t.Fatal("deleted note 2 still returned")
		}
	}
}

func TestVecIndexDimChangeRebuilds(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	if err := EnsureVectorIndex(database, 2); err != nil {
		t.Fatal(err)
	}
	if err := PutNoteEmbedding(database, 1, []float32{0, 0}); err != nil {
		t.Fatal(err)
	}
	// Switching dim drops + recreates (derived cache rebuilt from scratch).
	if err := EnsureVectorIndex(database, 3); err != nil {
		t.Fatalf("ensure dim3: %v", err)
	}
	if dim, ok := storedVectorDim(database); !ok || dim != 3 {
		t.Fatalf("dim after change = %d ok=%v, want 3", dim, ok)
	}
	if err := PutNoteEmbedding(database, 2, []float32{1, 2, 3}); err != nil {
		t.Fatalf("put 3-dim: %v", err)
	}
	// A 2-dim vector into a 3-dim index must error (encode ok, sqlite rejects).
	if err := PutNoteEmbedding(database, 3, []float32{1, 2}); err == nil {
		t.Error("expected error inserting wrong-dim vector")
	}
}

func TestVecIndexGuards(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()
	if err := EnsureVectorIndex(database, 0); err == nil {
		t.Error("dim 0 should error")
	}
	if err := EnsureVectorIndex(database, maxEmbeddingDim+1); err == nil {
		t.Error("oversized dim should error")
	}
	if err := PutNoteEmbedding(database, 1, nil); err == nil {
		t.Error("empty vector should error")
	}
	if _, err := KnnNoteIDs(database, []float32{1, 2}, 0); err != nil {
		t.Errorf("k=0 should be nil,err; got %v", err)
	}
}
