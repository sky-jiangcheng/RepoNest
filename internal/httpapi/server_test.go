package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"repo-nest/internal/db"
	"repo-nest/internal/service"
)

// Regression: the headless server is a third JSON boundary (besides Wails and
// MCP) over internal/service. Empty collections must serialise as []/{} here
// too, never null, or dsh-plugin tools that forward the payload to a model
// (and any JSON consumer) get a null where an array is expected.
func TestSearchHitsNotEmptyNull(t *testing.T) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	svc := service.New(database, "me")
	h := New(svc)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/search?q=anything", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	var body struct {
		Query string          `json:"query"`
		Hits  json.RawMessage `json:"hits"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if body.Hits == nil || string(body.Hits) == "null" {
		t.Errorf("hits serialised as null, want []: %s", body.Hits)
	}
}

// Regression: the MCP layer bounds its query argument (v1.9.2); the HTTP
// boundary serves the same service, so it must bound ?q= too. An unbounded
// value from any local process turns each FTS/LIKE match into a full-table
// scan, and the endpoint answers repeatedly.
func TestSearchQueryTooLong(t *testing.T) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	h := New(service.New(database, "me"))

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/search?q="+strings.Repeat("a", maxSearchQueryLen+1), nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "query too long") {
		t.Errorf("body %q missing the length rejection", rec.Body.String())
	}

	// A query at the boundary is still served, not rejected.
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/search?q="+strings.Repeat("a", maxSearchQueryLen), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("boundary query: status = %d, want 200", rec.Code)
	}
}

func TestEndpoints(t *testing.T) {
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })

	h := New(service.New(database, "me"))

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{"health", http.MethodGet, "/health", http.StatusOK, `"status":"ok"`},
		{"search missing q", http.MethodGet, "/api/search", http.StatusBadRequest, "missing query"},
		{"project not found", http.MethodGet, "/api/project/999999/overview", http.StatusNotFound, "error"},
		{"project bad id", http.MethodGet, "/api/project/abc/overview", http.StatusBadRequest, "invalid project id"},
		{"ai_context has markdown", http.MethodPost, "/api/ai_context", http.StatusOK, `"markdown"`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
			if rec.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d (body %s)", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.wantBody) {
				t.Errorf("body %q missing %q", rec.Body.String(), tc.wantBody)
			}
		})
	}
}
