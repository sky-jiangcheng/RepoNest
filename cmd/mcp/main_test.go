package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"reponest/internal/db"
	"reponest/internal/service"
)

// testServer bundles the pieces a test needs: a real MCPServer with every
// tool registered, the service behind it, and the raw *sql.DB so tests can
// seed fixtures. The same *sql.DB backs both the service and the seed
// (matching internal/app's setup), so there is no cross-handle visibility
// problem with :memory: databases.
type testServer struct {
	t   *testing.T
	s   *server.MCPServer
	svc *service.Service
	db  *sql.DB
}

// newTestServer wires the real tool registration over an in-memory database,
// so tests exercise the same code path a client does (registerTools →
// server.GetTool → handler) rather than a reimplementation of it.
func newTestServer(t *testing.T) *testServer {
	t.Helper()
	database, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("init db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	svc := service.New(database, "tester")
	s := server.NewMCPServer("reponest-mcp", "test")
	registerTools(s, svc)
	registerContextTools(s, svc)
	registerScanTool(s, svc)
	return &testServer{t: t, s: s, svc: svc, db: database}
}

// seedProject creates a project the same way the desktop app does and returns
// its ID.
func (ts *testServer) seedProject(name, root string) int64 {
	ts.t.Helper()
	tx, err := ts.db.Begin()
	if err != nil {
		ts.t.Fatal(err)
	}
	pid, err := db.SyncProjectTx(tx, name, root, 0, true)
	if err != nil {
		ts.t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		ts.t.Fatal(err)
	}
	return pid
}

// call invokes a registered tool by name the way a client would: through the
// server's tool table, not a captured handler reference.
func (ts *testServer) call(name string, args map[string]any) *mcp.CallToolResult {
	ts.t.Helper()
	st := ts.s.GetTool(name)
	if st == nil {
		ts.t.Fatalf("tool %q not registered", name)
	}
	res, err := st.Handler(context.Background(), mcp.CallToolRequest{
		Params: mcp.CallToolParams{Name: name, Arguments: args},
	})
	if err != nil {
		ts.t.Fatalf("handler %q returned error: %v", name, err)
	}
	if res == nil {
		ts.t.Fatalf("handler %q returned nil result", name)
	}
	return res
}

// text flattens a result's content to a single string for assertions.
func (ts *testServer) text(res *mcp.CallToolResult) string {
	ts.t.Helper()
	if len(res.Content) != 1 {
		ts.t.Fatalf("expected 1 content block, got %d", len(res.Content))
	}
	tc, ok := res.Content[0].(mcp.TextContent)
	if !ok {
		ts.t.Fatalf("expected TextContent, got %T", res.Content[0])
	}
	return tc.Text
}

// jsonPayload decodes a result's JSON text into a map for key assertions.
func (ts *testServer) jsonPayload(res *mcp.CallToolResult) map[string]any {
	ts.t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(ts.text(res)), &out); err != nil {
		ts.t.Fatalf("result is not JSON: %v\n%s", err, ts.text(res))
	}
	return out
}

func TestToolsRegistered(t *testing.T) {
	ts := newTestServer(t)
	want := []string{
		"reponest_notes_list",
		"reponest_notes_search",
		"reponest_notes_read",
		"reponest_projects_list",
		"reponest_projects_stats",
		"reponest_ask",
		"reponest_notes_create",
		"reponest_notes_update",
		"reponest_agent_score",
		"reponest_integrity",
		"reponest_context",
		"reponest_handoff",
		"reponest_scan",
	}
	for _, name := range want {
		st := ts.s.GetTool(name)
		if st == nil {
			t.Errorf("tool %q missing", name)
			continue
		}
		if st.Tool.Description == "" {
			t.Errorf("tool %q has empty description — agents pick tools by description", name)
		}
		if st.Tool.InputSchema.Type != "object" {
			t.Errorf("tool %q schema type = %q, want object", name, st.Tool.InputSchema.Type)
		}
	}
	if got := len(ts.s.ListTools()); got != len(want) {
		t.Errorf("registered %d tools, want %d", got, len(want))
	}
}

func TestProjectsListEmpty(t *testing.T) {
	ts := newTestServer(t)
	res := ts.call("reponest_projects_list", nil)
	// Fresh DB: JSON `[]`, not `null` — an agent testing for an empty array
	// should not have to handle null.
	if got := strings.TrimSpace(ts.text(res)); got != "[]" {
		t.Errorf("empty projects = %q, want []", got)
	}
}

func TestProjectsListAndStats(t *testing.T) {
	ts := newTestServer(t)
	pid := ts.seedProject("demo", "/tmp/demo")

	list := ts.call("reponest_projects_list", nil)
	var projects []map[string]any
	if err := json.Unmarshal([]byte(ts.text(list)), &projects); err != nil {
		t.Fatalf("projects list is not a JSON array: %v\n%s", err, ts.text(list))
	}
	if len(projects) != 1 {
		t.Fatalf("got %d projects, want 1", len(projects))
	}
	if id, _ := projects[0]["id"].(float64); int64(id) != pid {
		t.Errorf("project id = %v, want %d", projects[0]["id"], pid)
	}

	stats := ts.call("reponest_projects_stats", map[string]any{"id": float64(pid)})
	payload := ts.jsonPayload(stats)
	if _, ok := payload["project"]; !ok {
		t.Errorf("project_stats missing project key: %s", ts.text(stats))
	}
}

func TestNotesSearchRequiresQuery(t *testing.T) {
	ts := newTestServer(t)
	for _, args := range []map[string]any{nil, {"query": ""}} {
		res := ts.call("reponest_notes_search", args)
		if res.IsError {
			t.Errorf("args %v: expected non-error result, got error", args)
		}
		if got := ts.text(res); !strings.Contains(got, "query is required") {
			t.Errorf("args %v: message = %q, want it to mention query is required", args, got)
		}
	}
}

func TestAskRequiresQuery(t *testing.T) {
	ts := newTestServer(t)
	res := ts.call("reponest_ask", nil)
	if res.IsError {
		t.Error("expected non-error result for missing query")
	}
	if got := ts.text(res); !strings.Contains(got, "query is required") {
		t.Errorf("message = %q, want it to mention query is required", got)
	}
}

func TestNotesCreateAndReadRoundTrip(t *testing.T) {
	ts := newTestServer(t)
	if got := len(ts.svc.ListProjects()); got != 0 {
		t.Fatalf("fresh install should have 0 projects, got %d", got)
	}
	pid := ts.seedProject("demo", "/tmp/demo")

	created := ts.call("reponest_notes_create", map[string]any{
		"project_id": float64(pid),
		"title":      "FTS5 quirk",
		"content":    "external-content FTS5 tables answer content queries from the content table",
		"category":   "knowledge",
		"tags":       "sqlite,search",
	})
	createdPayload := ts.jsonPayload(created)
	noteID, ok := createdPayload["id"].(float64)
	if !ok {
		t.Fatalf("create result has no numeric id: %s", ts.text(created))
	}

	read := ts.call("reponest_notes_read", map[string]any{"id": noteID})
	readPayload := ts.jsonPayload(read)
	if title, _ := readPayload["title"].(string); title != "FTS5 quirk" {
		t.Errorf("read title = %q, want %q", title, "FTS5 quirk")
	}

	// Update content only, then confirm the new content is readable.
	ts.call("reponest_notes_update", map[string]any{
		"id":      noteID,
		"content": "updated: v12 rebuilds the index",
	})
	updated := ts.jsonPayload(ts.call("reponest_notes_read", map[string]any{"id": noteID}))
	if content, _ := updated["content"].(string); !strings.Contains(content, "v12 rebuilds") {
		t.Errorf("updated content = %q, want it to contain the new text", content)
	}

	// The new content must be findable by search — this is the round trip
	// that silently broke when the FTS backfill was a no-op (CHANGELOG 1.8.0).
	hits := ts.call("reponest_notes_search", map[string]any{"query": "rebuilds"})
	if got := ts.text(hits); !strings.Contains(got, "FTS5 quirk") {
		t.Errorf("search for indexed content returned %s, want a hit on the note", got)
	}
}

func TestNotesCreateMissingProject(t *testing.T) {
	ts := newTestServer(t)
	// Nonexistent project: the handler must surface an error-bearing result,
	// not panic and not return a success-shaped payload.
	res := ts.call("reponest_notes_create", map[string]any{
		"project_id": float64(999999),
		"title":      "orphan",
		"content":    "attached to nothing",
	})
	if res.IsError {
		t.Fatal("expected a text result with an error message, not a protocol error")
	}
	if got := ts.text(res); !strings.Contains(got, "error") {
		t.Errorf("expected error indication, got: %s", got)
	}
}

func TestNotesReadMissingID(t *testing.T) {
	ts := newTestServer(t)
	res := ts.call("reponest_notes_read", map[string]any{"id": float64(424242)})
	if res.IsError {
		t.Fatal("not-found should be a text result, not a protocol error")
	}
	if got := ts.text(res); !strings.Contains(got, "note not found") {
		t.Errorf("message = %q, want it to say note not found", got)
	}
}

func TestProjectsStatsMissingProject(t *testing.T) {
	ts := newTestServer(t)
	res := ts.call("reponest_projects_stats", map[string]any{"id": float64(424242)})
	if res.IsError {
		t.Fatal("not-found should be a text result, not a protocol error")
	}
	if got := ts.text(res); !strings.Contains(got, "project not found") {
		t.Errorf("message = %q, want it to say project not found", got)
	}
}

func TestAgentScoreOnFreshInstall(t *testing.T) {
	ts := newTestServer(t)
	out := ts.text(ts.call("reponest_agent_score", nil))
	if !strings.Contains(out, "Agent Score") {
		t.Errorf("output missing header:\n%s", out)
	}
	if !strings.Contains(out, "Score: ") {
		t.Errorf("output missing score line:\n%s", out)
	}
	// Fresh install has no notes: check 1 (DB reachable) must still pass,
	// because it now pings the handle instead of reusing noteCount.
	if !strings.Contains(out, "Database reachable") {
		t.Errorf("output missing database check:\n%s", out)
	}
	// The two checks must not be the same signal — "No notes" should appear
	// as its own warning, which is only possible if check 1 didn't also hinge
	// on noteCount (the double-count bug fixed in 1.8.0).
	if !strings.Contains(out, "No notes") {
		t.Errorf("expected a distinct no-notes warning:\n%s", out)
	}
}

func TestAgentScoreScoreLineIsFractionOfTotal(t *testing.T) {
	ts := newTestServer(t)
	out := ts.text(ts.call("reponest_agent_score", nil))
	// Parse "Score: N/M" and N <= M; a renderer bug that dropped the total
	// would silently make every install look "ready".
	i := strings.Index(out, "Score: ")
	if i < 0 {
		t.Fatalf("no score line:\n%s", out)
	}
	rest := out[i+len("Score: "):]
	j := strings.IndexAny(rest, "\n")
	if j < 0 {
		t.Fatalf("score line not terminated:\n%s", out)
	}
	var earned, total int
	e, tt, err := sscanfScore(rest[:j])
	if err != nil {
		t.Fatalf("cannot parse score %q: %v", rest[:j], err)
	}
	earned, total = e, tt
	if total <= 0 || earned < 0 || earned > total {
		t.Errorf("score %d/%d out of range", earned, total)
	}
}

// sscanfScore parses "N/M" without pulling fmt's verb handling into the test.
func sscanfScore(s string) (int, int, error) {
	parts := strings.SplitN(s, "/", 2)
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("want N/M, got %q", s)
	}
	earned, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, err
	}
	total, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, err
	}
	return earned, total, nil
}

func TestIntegrityReportOnFreshInstall(t *testing.T) {
	ts := newTestServer(t)
	out := ts.text(ts.call("reponest_integrity", nil))
	if !strings.Contains(out, "Data Integrity") {
		t.Errorf("output missing header:\n%s", out)
	}
	if !strings.Contains(out, "Trust score:") {
		t.Errorf("output missing trust score:\n%s", out)
	}
	if !strings.Contains(out, "checks:") {
		t.Errorf("output missing check summary:\n%s", out)
	}
	// On an empty database nothing should be reported as FAILED.
	if strings.Contains(out, "FAILED") {
		t.Errorf("fresh install should not report failures:\n%s", out)
	}
}

func TestIntegrityReportFindsSeededDrift(t *testing.T) {
	ts := newTestServer(t)
	pid := ts.seedProject("drift", "/tmp/drift")

	// Create a note through MCP (so it lands in both content table and FTS),
	// then delete the FTS row directly to simulate index drift — the exact
	// failure mode the integrity tool exists to detect.
	created := ts.jsonPayload(ts.call("reponest_notes_create", map[string]any{
		"project_id": float64(pid),
		"title":      "will be orphaned",
		"content":    "search should still find this",
	}))
	noteID, _ := created["id"].(float64)

	if _, err := ts.db.Exec("DELETE FROM project_notes_fts WHERE rowid = ?", int64(noteID)); err != nil {
		t.Fatalf("simulating drift: %v", err)
	}

	out := ts.text(ts.call("reponest_integrity", nil))
	// The report must not come back clean when the index is provably wrong.
	if !strings.Contains(out, "FTS") {
		t.Errorf("expected an FTS-related finding:\n%s", out)
	}
}

// TestScanToolDiscoversRepositories covers the headless cold start: with no
// desktop app and an empty database, reponest_scan alone must discover repos
// and make reponest_context usable in the very next call.
func TestScanToolDiscoversRepositories(t *testing.T) {
	ts := newTestServer(t)
	tmp := t.TempDir()
	// Two repositories under one parent: the grouper collapses them into a
	// single (monorepo) project.
	for _, name := range []string{"alpha", "beta"} {
		if err := os.MkdirAll(filepath.Join(tmp, name, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.ReplaceScanRoots(ts.db, []string{tmp}); err != nil {
		t.Fatalf("seed scan root: %v", err)
	}

	res := ts.call("reponest_scan", nil)
	payload := ts.jsonPayload(res)
	if got, _ := payload["repos_found"].(float64); int(got) != 2 {
		t.Fatalf("repos_found = %v, want 2\n%s", payload["repos_found"], ts.text(res))
	}
	if got, _ := payload["projects"].(float64); int(got) != 1 {
		t.Fatalf("projects = %v, want 1 (both repos share a parent)", payload["projects"])
	}

	// The point of the scan: the next context call works with zero parameters.
	if text := ts.text(ts.call("reponest_context", nil)); !strings.Contains(text, "# Project Context:") {
		t.Errorf("context after scan should render a project\n%s", text)
	}
}
