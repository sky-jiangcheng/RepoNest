// Package httpapi exposes RepoNest's shared business logic (internal/service)
// over a small HTTP/JSON surface so that external runtimes — most importantly
// the DeepSeek Harness dsh-plugin — can call the same analysis code the
// desktop App uses, with zero logic duplication.
//
// The desktop Wails App, the (future) CLI and this headless server are all thin
// adapters over internal/service. They share one implementation and one SQLite
// database.
//
// Trust boundary: this API has NO authentication — it serves the complete
// local knowledge base (notes, todos, mined repo knowledge). It is safe only
// because cmd/server binds it to 127.0.0.1, so reachability is equivalent to
// "another process on this machine". Anyone embedding this handler (tests,
// plugins, future transports) must preserve that property: loopback-only,
// never 0.0.0.0, and never documented as a network service. Exposing it
// beyond loopback would publish the user's private code knowledge to the
// network.
package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"reponest/internal/service"
)

// New returns an http.Handler (ServeMux) that serves RepoNest's capabilities
// as JSON endpoints. The supplied service must already be constructed with a
// valid database; callers do not need to invoke service.Startup for read-only
// endpoints.
func New(svc *service.Service) http.Handler {
	mux := http.NewServeMux()
	h := &handler{svc: svc}

	mux.HandleFunc("/health", h.health)
	mux.HandleFunc("/api/ai_context", h.aiContext)
	mux.HandleFunc("/api/search", h.search)
	mux.HandleFunc("/api/project/", h.project)
	return mux
}

type handler struct {
	svc *service.Service
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *handler) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, h.svc.Health())
}

// aiContext returns the AI-readable knowledge-base markdown (llms.txt style).
// POST with an optional JSON body {"project_id": N} — project_id is accepted
// for forward compatibility but the current implementation returns the full
// local knowledge base, matching the desktop "Copy AI Context" action.
func (h *handler) aiContext(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	markdown := h.svc.GenerateLLMsTxt()
	writeJSON(w, http.StatusOK, map[string]string{"markdown": markdown})
}

// search runs a full-text knowledge search across notes and todos.
// Query param: q (required). Use ?all=1 to also include todos (default notes).
func (h *handler) search(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing query param 'q'"})
		return
	}
	var hits any
	if r.URL.Query().Get("all") == "1" {
		hits = h.svc.SearchAll(q)
	} else {
		hits = h.svc.SearchNotes(q)
	}
	writeJSON(w, http.StatusOK, map[string]any{"query": q, "hits": hits})
}

// project dispatches /api/project/{id}/detail and /api/project/{id}/stats.
func (h *handler) project(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	// Strip prefix "/api/project/" → "{id}/detail" or "{id}/stats".
	rest := strings.TrimPrefix(r.URL.Path, "/api/project/")
	rest = strings.Trim(rest, "/")
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing project id"})
		return
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid project id"})
		return
	}

	switch {
	case len(parts) == 1 || parts[1] == "detail" || parts[1] == "":
		detail, err := h.svc.GetProjectDetail(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, detail)
	case parts[1] == "overview":
		ov, err := h.svc.GetProjectOverview(id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, ov)
	case parts[1] == "stats":
		date := r.URL.Query().Get("date")
		stats := h.svc.GetProjectStats(id, date)
		writeJSON(w, http.StatusOK, map[string]any{"project_id": id, "date": date, "stats": stats})
	default:
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "unknown project subroute: " + parts[1]})
	}
}
