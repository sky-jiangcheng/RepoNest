package service

import (
	"strings"
	"testing"

	"reponest/internal/db"
)

func TestResolveProjectByIDAndName(t *testing.T) {
	s, _ := setupService(t)
	pid := seedProject(t, s.db, "auth-service", "/home/me/code/auth-service")
	seedProject(t, s.db, "web", "/home/me/code/web")

	if got := s.ResolveProject(""); got.Project != nil || len(got.Candidates) != 2 {
		t.Fatalf("empty query with 2 projects should return candidates, got %+v", got)
	}

	byID := s.ResolveProject("1")
	if byID.Project == nil || byID.Project.ID != pid {
		t.Fatalf("numeric query should resolve by ID, got %+v", byID)
	}

	byName := s.ResolveProject("auth")
	if byName.Project == nil || byName.Project.Name != "auth-service" {
		t.Fatalf("fuzzy name should resolve, got %+v", byName)
	}

	// Prefix "web" matches one project exactly; "e" matches both, which must
	// stay ambiguous instead of silently picking the first.
	if got := s.ResolveProject("e"); got.Project != nil || len(got.Candidates) != 2 {
		t.Fatalf("ambiguous match should return candidates, got %+v", got)
	}

	if got := s.ResolveProject("999"); got.Project != nil {
		t.Fatalf("unknown ID should not resolve, got %+v", got)
	}
}

func TestResolveProjectSingleProjectNoQuery(t *testing.T) {
	s, _ := setupService(t)
	seedProject(t, s.db, "only", "/home/me/only")

	got := s.ResolveProject("")
	if got.Project == nil || got.Project.Name != "only" {
		t.Fatalf("single project should auto-resolve without a query, got %+v", got)
	}
}

func TestBuildProjectContextRendersNotesAndTodos(t *testing.T) {
	s, _ := setupService(t)
	pid := seedProject(t, s.db, "demo", "/home/me/demo")

	if _, err := s.CreateNoteWithMeta(pid, "Auth flow design", "uses signed cookies", "auth", "knowledge", "manual"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateTodo(pid, "add rate limiting"); err != nil {
		t.Fatal(err)
	}

	doc := s.BuildProjectContext(s.ResolveProject("demo"))
	for _, want := range []string{
		"# Project Context: demo",
		"`/home/me/demo`",
		"Open Todos",
		"- [ ] add rate limiting",
		"### Auth flow design",
		"uses signed cookies",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("context doc missing %q\n%s", want, doc)
		}
	}
}

func TestBuildProjectContextHandoffNotesFirst(t *testing.T) {
	s, _ := setupService(t)
	pid := seedProject(t, s.db, "demo", "/home/me/demo")

	if _, err := s.CreateNoteWithMeta(pid, "Older design note", "design detail", "design", "knowledge", "manual"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateHandoffNote(HandoffInput{
		ProjectID: pid,
		Agent:     "claude-code",
		Summary:   "migrated auth to signed cookies",
		NextSteps: []string{"rotate cookie keys in prod"},
	}); err != nil {
		t.Fatal(err)
	}

	doc := s.BuildProjectContext(s.ResolveProject("demo"))
	handoffIdx := strings.Index(doc, "Session Handoff")
	designIdx := strings.Index(doc, "Older design note")
	if handoffIdx == -1 || designIdx == -1 {
		t.Fatalf("both notes should be present\n%s", doc)
	}
	if handoffIdx > designIdx {
		t.Errorf("handoff note should render before plain notes\n%s", doc)
	}
}

func TestBuildProjectContextAmbiguousReturnsCatalog(t *testing.T) {
	s, _ := setupService(t)
	seedProject(t, s.db, "alpha", "/home/me/alpha")
	seedProject(t, s.db, "beta", "/home/me/beta")

	doc := s.BuildProjectContext(s.ResolveProject(""))
	if !strings.Contains(doc, "| 1 | alpha |") || !strings.Contains(doc, "| 2 | beta |") {
		t.Errorf("ambiguous request should render a project catalog\n%s", doc)
	}
}

func TestCreateHandoffNoteValidationAndTemplate(t *testing.T) {
	s, _ := setupService(t)
	pid := seedProject(t, s.db, "demo", "/home/me/demo")

	if _, err := s.CreateHandoffNote(HandoffInput{ProjectID: pid, Summary: "  "}); err == nil {
		t.Error("empty summary must be rejected")
	}
	if _, err := s.CreateHandoffNote(HandoffInput{ProjectID: pid, Summary: "s"}); err == nil {
		t.Error("summary without any section must be rejected")
	}

	res, err := s.CreateHandoffNote(HandoffInput{
		ProjectID: pid,
		Agent:     "cursor",
		Summary:   "shipped the settings page",
		Changes:   []string{"", "added settings route"},
		Decisions: []string{"kept global CSS for design tokens"},
		Gotchas:   []string{"vite needs allowedHosts for preview"},
		NextSteps: []string{"run smoke test on windows"},
		Tags:      "frontend",
	})
	if err != nil {
		t.Fatal(err)
	}

	note, err := db.GetNoteByID(s.db, res.NoteID)
	if err != nil {
		t.Fatal(err)
	}
	if note.Kind != "knowledge" || note.Source != "mcp" {
		t.Errorf("handoff should persist as knowledge/mcp, got kind=%s source=%s", note.Kind, note.Source)
	}
	if !noteIsHandoff(*note) {
		t.Errorf("handoff note must carry the %q tag, tags=%q", handoffTag, note.Tags)
	}
	if !strings.Contains(note.Tags, "frontend") {
		t.Errorf("extra tags should be preserved, tags=%q", note.Tags)
	}
	for _, want := range []string{
		"## Summary\n\nshipped the settings page",
		"- added settings route",
		"- kept global CSS for design tokens",
		"- vite needs allowedHosts for preview",
		"- run smoke test on windows",
		"> Written by `cursor`",
	} {
		if !strings.Contains(note.Content, want) {
			t.Errorf("handoff template missing %q\n%s", want, note.Content)
		}
	}
	if strings.Contains(note.Content, "- \n") {
		t.Errorf("blank items must be dropped\n%s", note.Content)
	}
}
