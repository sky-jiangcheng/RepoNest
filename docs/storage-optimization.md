---
title: Storage Optimization & AI Value
order: 22
---

# Storage Optimization & AI Value

What RepoNest does to a git repository is not "compression / dedup"-style storage optimization; it **models unstructured, massive raw git information into a structured, indexable, materializable, incrementally maintainable local knowledge base**. It solves the problem of AI reading git repositories expensively, messily, incompletely, and slowly.

This document targets two kinds of readers:

- Developers who want to understand "where RepoNest actually stores data, and how"
- Decision makers who want to understand "why hand the repository to RepoNest first instead of straight to AI"

## 1. Storage Structure Optimization for Git Repositories

| # | Optimization | Implementation (`internal/db/migrate.go` / `daily_stats.go` / `repo_meta.go`) |
|---|------|------|
| 1 | **Unstructured → relational model** | `projects(1) — repositories(N)` normalized; `project_notes` / `project_todos` separate "knowledge" from the code repo; `repo_meta` caches README / tech stack / languages / dependencies / contributors / activity as **JSON columns**, avoiding a filesystem scan every time |
| 2 | **FTS5 full-text index** | `project_notes_fts` / `project_todos_fts` virtual tables, `tokenize='trigram'`. Trigram is friendly to Chinese / CJK; short queries automatically fall back to `LIKE` (see [ADR-0003](adr/0003-fts5-search.md)) |
| 3 | **Trigger-based dual writes** | Note insert / update / delete automatically syncs FTS; update automatically writes a version snapshot to `note_versions_snap` (the `WHEN` condition has been narrowed so only content-field changes produce snapshots), zero maintenance at the application layer |
| 4 | **Secondary indexes** | `idx_projects_starred / collected / collected_at` speed up starred filtering; `idx_note_versions(note_id, created_at DESC)` supports version traversal |
| 5 | **Materialized statistics (core)** | `daily_stats` pre-aggregates raw `git log` commits by `(repository_id, stat_date, author)` (`ON CONFLICT DO UPDATE`, idempotent). heatmap / summary read straight from this table with `GROUP BY`, **never re-running `git log`** |
| 6 | **Mining result cache** | `repo_meta` uses `ON CONFLICT DO UPDATE` + `updated_at` for invalidation and re-mines incrementally, avoiding repeated filesystem scans |
| 7 | **Engineering-layer tuning** | SQLite WAL + single-connection tuning; concurrent reads and writes do not block; 8 versioned migrations run automatically, immutable and append-only |

> Item 5 is the fundamental source of the "AI value": a git repository's raw information (commit history, file paths, file contents) is **volatile and expensive**; every AI-direct `git log` / per-file read is a one-off consumption. RepoNest **parses it once at scan time**, after which queries never touch the `git` command.

## 2. Advantages over AI Reading Git Repositories Directly

| Dimension | AI reads the git repository directly | RepoNest structured storage |
|------|--------------------|----------------------|
| **Retrieval** | Per-file scanning / whole-repo dumps, context explosion | FTS5 trigram exact retrieval, Chinese-friendly, millisecond-level |
| **Statistics** | `git log` recomputed every time, slow and token-hungry | `daily_stats` already materialized; one `GROUP BY` away, zero recomputation |
| **Cross-repository** | Single-repo view, hard to aggregate | `projects / repositories` normalized; one line of SQL aggregates across repos |
| **Context window** | Large repos exceed the limit and get truncated | MCP tools fetch data on demand, feeding only relevant fragments |
| **History** | Only the current snapshot | `note_versions` snapshots make knowledge traceable |
| **Determinism** | The model may hallucinate / miss binary files, `.gitignore` | Deterministic SQL, reproducible |
| **Cost** | Repeated reading = repeated tokens | Parse once, query unlimited times |
| **Offline / latency** | Depends on network and model | Local SQLite, zero latency, works offline |

## 3. Core Argument

The AI bottleneck is not "whether git can be read" but **reading expensively, messily, incompletely, and slowly**. In essence, RepoNest gives AI a layer of **structured memory**:

- Volatile raw git data (commits, paths, files) is parsed once at scan time → materialized into `daily_stats` / `repo_meta`, after which queries never touch the `git` command;
- Knowledge notes live in tables with FTS5 indexes + version snapshots, letting AI hit precisely with `reponest_notes_search` / `reponest_ask` instead of semantic guessing;
- MCP interfaces let AI **fetch data on demand** instead of dumping the whole repository into the context window every time.

In one sentence: **AI reading git directly = re-understanding the world every time; RepoNest = pre-organizing the world into an index AI can query precisely**. The former burns tokens and misses things; the latter organizes once, reuses repeatedly, and is deterministic and reliable.

## 4. This Layer Itself Calls No LLM

To emphasize: the storage structure, indexes, materialized statistics, MCP interfaces — this data foundation **calls no large language model at all** (no API key / model / endpoint configuration). It is purely a "data foundation"; AI consumes it through the 13 tools of [`reponest-mcp`](features/ai-integration.md). In other words, RepoNest is the **data source for AI**, rather than AI itself.

Related pages:

- [Architecture](architecture.md) — layering, scan pipeline, database table overview
- [AI Integration](features/ai-integration.md) — MCP / llms.txt integration and value positioning
- [Knowledge Base & Notes](features/knowledge.md) — FTS5 search and version history for notes
