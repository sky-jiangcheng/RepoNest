package vectordb

import (
	"database/sql"
	"log"
	"strings"
)

// Open resolves the configured vector store, defaulting to (and falling back
// to) the local sqlite-vec store. storeKind is "" / "local" / "qdrant".
//
// The fallback is deliberate and never fails the caller: a misconfigured or
// unreachable Qdrant silently degrades to local so semantic search keeps
// working against whatever store IS available (ADR-0013 / ADR-0012's "never
// make search return less").
func Open(database *sql.DB, storeKind, url, apiKey, collection string) Store {
	if strings.EqualFold(strings.TrimSpace(storeKind), "qdrant") {
		q, err := NewQdrant(url, apiKey, collection)
		if err != nil {
			log.Printf("vectordb: qdrant config invalid (%v); using local sqlite-vec", err)
			return NewLocal(database)
		}
		if !q.Reachable() {
			log.Printf("vectordb: qdrant %s unreachable; using local sqlite-vec", q.Base)
			return NewLocal(database)
		}
		return q
	}
	return NewLocal(database)
}
