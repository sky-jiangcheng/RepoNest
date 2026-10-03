package vectordb

import (
	"database/sql"

	"repo-nest/internal/db"
)

// Local is the default store: sqlite-vec `note_embeddings` living inside the
// app's dashboard.db (wraps the internal/db vec helpers, ADR-0013).
type Local struct{ database *sql.DB }

// NewLocal binds the sqlite-vec store to an opened database.
func NewLocal(database *sql.DB) *Local { return &Local{database: database} }

func (l *Local) Name() string { return "local-sqlite-vec" }

func (l *Local) Ensure(dim int) error { return db.EnsureVectorIndex(l.database, dim) }

func (l *Local) Clear(_ int) error { return db.ClearVectorIndex(l.database) }

func (l *Local) Upsert(id int64, vec []float32) error {
	return db.PutNoteEmbedding(l.database, id, vec)
}

func (l *Local) Search(vec []float32, limit int) ([]int64, error) {
	return db.KnnNoteIDs(l.database, vec, limit)
}
