---
title: API Reference
order: 22
---

# API Reference

RepoNest's external interface is the **Wails binding surface**: Go methods are exposed to the frontend through Wails Bind (`window.go.main.App.<methodName>`), and the method names plus JSON payloads form the contract. The desktop app listens on no HTTP port by default; when an HTTP shape is needed, `cmd/server` provides a loopback-only JSON API (reusing the same service layer).

> The binding layer is thin delegation (`internal/app`); all implementation lives in `internal/service`; the CLI and MCP reuse the same implementation.

## 1.7.0 Contract Changes

- `GetHeatmapData(projectId int64)`: new parameter; `0` means global, `>0` restricts to that project's repositories (the project detail page previously used global data by mistake)
- Removed dead methods never called by the frontend: `ExportProjectStats`, `ExportHeatmapCSV`, `GetNoteVersion`, `ScanForRepositories`, `RefreshStats`, `RefreshAllStats`, `RefreshProjectStats`

---

## Headless HTTP server (`reponest server`)

A JSON API for external runtimes such as the DeepSeek Harness dsh-plugin; it shares the same `internal/service` implementation and the same SQLite database as the desktop app, CLI, and MCP.

| Endpoint | Method | Description |
|------|------|------|
| `/health` | GET | Server and database health status |
| `/api/ai_context` | GET/POST | Full knowledge base as Markdown (llms.txt style) |
| `/api/search?q=...&all=1` | GET | FTS5 full-text search (notes only by default; `all=1` includes todos) |
| `/api/project/{id}/detail` | GET | Project + repositories and their historical stats |
| `/api/project/{id}/overview` | GET | Knowledge mining results (README/tech stack/dependencies, etc.) |
| `/api/project/{id}/stats?date=` | GET | Project stats for a given day |

> **Trust boundary**: this server has **no authentication** and returns the user's complete local knowledge base. Security relies entirely on `cmd/server` binding to `127.0.0.1` only — reachability is equivalent to "another process on this machine". **Never** change it to `0.0.0.0` or expose it to the network through a reverse proxy; if remote access is needed, add an authentication scheme and submit an ADR first.

---

## Projects

| Method | Signature | Description |
|------|------|------|
| `GetProjects` | `(date string, starredOnly bool) → ProjectResponse[]` | Project list (yesterday by default; triggers an on-demand single-day refresh when today/yesterday has no data) |
| `GetProjectDetail` | `(id) → ProjectDetail` | Project + repository list and per-repository historical stats |
| `GetProjectStats` | `(id, date) → DailyStat[]` | Project stats for a given day (yesterday by default) |
| `GetProjectOverview` | `(id) → ProjectOverview` | Knowledge mining: README/tech stack/languages/dependencies/contributors/activity/latest commits (cached in repo_meta; mined async on cache miss) |
| `SearchProjects` | `(query) → ProjectResponse[]` | Fuzzy search by name/path (enriched with yesterday's stats) |
| `ToggleStar` | `(id) → bool` | Toggle star, returns the new state (atomic UPDATE) |
| `UpdateProjectLevel` | `(id, "up"\|"down") → {success, new_level}` | Merge/split projects (single transaction) |
| `RefreshProjectHistory` | `(id) → {success}` | Backfill the project's last 365 days of stats |

**ProjectResponse** (excerpt):

```json
{
  "id": 1, "name": "my-project", "root_path": "/Users/me/code/my-project",
  "is_starred": true, "repo_count": 2,
  "total_added": 1200, "total_deleted": 300,
  "my_added": 800, "my_deleted": 200, "my_files": 15,
  "is_workday": true, "below_standard": false
}
```

## Scanning

| Method | Signature | Description |
|------|------|------|
| `TriggerScan` | `() → {success, task_id}` | Async full scan, returns immediately |
| `GetScanStatus` | `() → ScanStatus` | `{running, backfilling, message, progress, total}` |

## Summary & Status

| Method | Signature | Description |
|------|------|------|
| `GetSummary` | `(date) → Summary` | Global daily summary (team/personal additions & deletions, files, repo count, workday flag) |
| `GetHeatmapData` | `(projectId) → {days: HeatmapDay[]}` | One-year heatmap; `projectId>0` restricts to a project |
| `GetStatusBar` | `() → StatusBarData` | Current time + latest commit (30s cache) |
| `GetTodoCounts` / `GetNoteCounts` | `() → counts[]` | Per-project todos (incomplete/total) and note counts |

## Search

| Method | Signature | Description |
|------|------|------|
| `SearchNotes` | `(query) → SearchHit[]` | FTS5 note search (bm25 + snippet highlighting) |
| `SearchAll` | `(query) → SearchHit[]` | Combined note + todo search |

## Notes

| Method | Description |
|------|------|
| `ListNotes(projectID)` / `ListAllNotes()` / `ListAllTags()` | Project notes / global notes (with project name) / all tags |
| `CreateNote(projectID, content)` | Create (default kind=other, source=manual) |
| `CreateNoteWithMeta(projectID, title, content, tags, kind, source)` | Create with metadata |
| `UpdateNote(noteID, content)` / `UpdateNoteMeta(noteID, title, tags, kind, pinned)` | Update content / metadata |
| `PinNote(noteID, pinned)` / `MoveNote(noteID, projectID)` / `DeleteNote(noteID)` | Pin / move / delete |
| `ListNoteVersions(noteID)` | Version list (last 50) |
| `RestoreNoteVersion(noteID, versionID)` | Restore to a historical version |
| `DiffNoteVersions(noteID, versionID)` | Line-level diff of a version vs current (`+/-/space` prefix) |

A successful create emits the plugin event `note.created`.

## Todos

`ListTodos(projectID)` / `CreateTodo(projectID, title)` / `ToggleTodo(id)` / `DeleteTodo(id)` / `ReorderTodos(ids[])` (reorder in a single transaction).

## Configuration

| Method | Description |
|------|------|
| `GetConfig()` | `{config: map, scan_roots: []}` |
| `UpdateConfig(key, value)` | Allowed keys: `daily_code_standard` / `scan_depth` / `git_author` / `auto_import` (numeric keys are validated as numbers) |
| `UpdateScanRoots(roots[])` | Atomically replace the scan root list |

## AI Export & Plugins

| Method | Description |
|------|------|
| `GenerateLLMsTxt()` | LLM-oriented knowledge base overview in Markdown |
| `ExportNoteAsMarkdown(noteID)` | Note as Markdown with YAML frontmatter |
| `GetPluginStatuses()` / `ReloadPlugins()` | Plugin load status / hot reload |
| `GetKnowledgeSources()` | Knowledge source list (built-in claude + plugin-registered) |
| `TriggerKnowledgeImport(name)` / `TriggerAllKnowledgeImports()` | Trigger imports (emits the `import.completed` frontend event) |
| `ImportClaudeMemory()` | One-click Claude memory import |
| `Health()` | `{status, version}` |

## Events (Wails → frontend)

| Event | Payload |
|------|------|
| `import.completed` | `{source, created, updated, skipped, error?}` |

---

## OpenAPI

For now the tables on this page serve as the machine-readable contract (the previous hand-written openapi.json was removed along with D10 and will return once CI generates it automatically). The desktop app exposes no public HTTP service (see [TODO](https://github.com/sky-jiangcheng/repo-nest/blob/master/TODO.md)).
