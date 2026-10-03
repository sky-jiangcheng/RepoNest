package vectordb

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type wStub struct {
	t           *testing.T
	created     map[string]bool
	seenIDs     map[string]bool
	lastCreate  string
	lastPut     string
	lastGraphQL string
}

func (s *wStub) serve(w http.ResponseWriter, r *http.Request) {
	b, _ := io.ReadAll(r.Body)
	switch {
	case r.URL.Path == "/v1/.well-known/ready":
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/schema":
		s.lastCreate = string(b)
		var body struct {
			Class string `json:"class"`
		}
		_ = json.Unmarshal(b, &body)
		s.created[body.Class] = true
		writeJSON(w, map[string]any{"class": body.Class})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/v1/schema/"):
		if s.created[strings.TrimPrefix(r.URL.Path, "/v1/schema/")] {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	case r.Method == http.MethodPost && r.URL.Path == "/v1/objects":
		var o struct {
			ID string `json:"id"`
		}
		_ = json.Unmarshal(b, &o)
		if s.seenIDs[o.ID] {
			w.WriteHeader(http.StatusUnprocessableEntity) // dup id -> client PUTs
			return
		}
		s.seenIDs[o.ID] = true
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/v1/objects/"):
		s.lastPut = string(b)
		w.WriteHeader(http.StatusOK)
	case r.Method == http.MethodPost && r.URL.Path == "/v1/graphql":
		s.lastGraphQL = string(b)
		writeJSON(w, map[string]any{"data": map[string]any{"Get": map[string]any{
			"Repo_x": []any{map[string]any{"note_id": 5}, map[string]any{"note_id": 6}},
		}}})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func newWStub(t *testing.T) (*wStub, *httptest.Server) {
	s := &wStub{t: t, created: map[string]bool{}, seenIDs: map[string]bool{}}
	return s, httptest.NewServer(http.HandlerFunc(s.serve))
}

func TestWeaviateConfig(t *testing.T) {
	if _, err := NewWeaviate("localhost:8080", "", "x"); err == nil {
		t.Error("non-http base should error")
	}
	w, err := NewWeaviate("http://h:8080/", "", "repo_smoke-test")
	if err != nil {
		t.Fatal(err)
	}
	if w.Base != "http://h:8080" {
		t.Errorf("base not trimmed: %q", w.Base)
	}
	if w.Class != "Repo_smoke_test" { // - -> _, first rune capitalised (Weaviate does this too)
		t.Errorf("class = %q, want Repo_smoke_test", w.Class)
	}
}

func TestWeaviateRoundTrip(t *testing.T) {
	stub, srv := newWStub(t)
	defer srv.Close()
	w, _ := NewWeaviate(srv.URL, "k", "repo_x")
	if !w.Reachable() {
		t.Fatal("expected reachable")
	}
	if err := w.Ensure(3); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if !strings.Contains(stub.lastCreate, `"class":"Repo_x"`) || !strings.Contains(stub.lastCreate, "note_id") {
		t.Errorf("create body wrong: %s", stub.lastCreate)
	}
	// Second Ensure is a no-op (GET schema returns 200).
	if err := w.Ensure(3); err != nil {
		t.Fatalf("ensure again: %v", err)
	}
	if err := w.Upsert(1, []float32{1, 0, 0}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	// Duplicate id -> POST 422 -> PUT fallback recorded.
	if err := w.Upsert(1, []float32{1, 0, 0}); err != nil {
		t.Fatalf("upsert dup: %v", err)
	}
	if !strings.Contains(stub.lastPut, `"id":"00000000-0000-0000-0000-000000000001"`) {
		t.Errorf("PUT fallback not used; lastPut=%q", stub.lastPut)
	}
	ids, err := w.Search([]float32{1, 0, 0}, 2)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !reflect.DeepEqual(ids, []int64{5, 6}) {
		t.Fatalf("search ids = %v, want [5 6]", ids)
	}
	if !strings.Contains(stub.lastGraphQL, "nearVector") || !strings.Contains(stub.lastGraphQL, "Repo_x") {
		t.Errorf("graphql query wrong: %s", stub.lastGraphQL)
	}
}
