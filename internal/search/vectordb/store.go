package vectordb

import (
	"database/sql"
	"fmt"
	"log"
	"sort"
	"strings"
)

// remoteConfig carries what a remote-store factory needs (ignored by local).
type remoteConfig struct {
	url        string
	apiKey     string
	collection string
}

// factory builds a store from config; a non-nil error (or an unreachable
// remote) makes Open fall back to local. Adding a backend = write a Store +
// Register("kind", factory) — no change to callers.
type factory func(database *sql.DB, rc remoteConfig) (Store, error)

var registry = map[string]factory{
	"qdrant": func(_ *sql.DB, rc remoteConfig) (Store, error) {
		q, err := NewQdrant(rc.url, rc.apiKey, rc.collection)
		if err != nil {
			return nil, err
		}
		if !q.Reachable() {
			return nil, fmt.Errorf("qdrant %s unreachable", q.Base)
		}
		return q, nil
	},
	// Future backends register here: "lancedb" (NOTE: CGO → breaks the
	// zero-CGO build constraint), "bleve" (pure-Go, also subsumes text).
	// See ADR-0013 candidate matrix.
	"weaviate": func(_ *sql.DB, rc remoteConfig) (Store, error) {
		w, err := NewWeaviate(rc.url, rc.apiKey, rc.collection)
		if err != nil {
			return nil, err
		}
		if !w.Reachable() {
			return nil, fmt.Errorf("weaviate %s unreachable", w.Base)
		}
		return w, nil
	},
}

// Register adds a named store backend to the selectable set (used by
// future adapters; local is always implicit).
func Register(kind string, f factory) { registry[strings.ToLower(kind)] = f }

// Kinds lists selectable vector-store backends (for help / validation / UI).
func Kinds() []string {
	out := make([]string, 0, len(registry)+1)
	out = append(out, "local")
	for k := range registry {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Open resolves the configured store. local is the default; an empty/"local",
// an unknown kind, or a remote that errors/is-unreachable all fall back to the
// local sqlite-vec store (ADR-0013: never make search return less than before).
func Open(database *sql.DB, kind, url, apiKey, collection string) Store {
	kind = strings.ToLower(strings.TrimSpace(kind))
	if kind == "" || kind == "local" {
		return NewLocal(database)
	}
	f, ok := registry[kind]
	if !ok {
		log.Printf("vectordb: unknown vector_store %q (available: %v); using local sqlite-vec", kind, Kinds())
		return NewLocal(database)
	}
	st, err := f(database, remoteConfig{url, apiKey, collection})
	if err != nil {
		log.Printf("vectordb: vector_store %q unavailable (%v); using local sqlite-vec", kind, err)
		return NewLocal(database)
	}
	return st
}
