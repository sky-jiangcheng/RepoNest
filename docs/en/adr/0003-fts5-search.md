# ADR-0003: FTS5 trigram Full-text Search

- Status: Accepted
- Date: 2026-08-10 (backfilled; implemented in 1.6.x / issue #18)
- Related: [ADR-0002](0002-c-end-repositioning.md)

## Background

The knowledge base needs full-text search over note titles / content / tags and todo titles, and it must support Chinese (segmentation without spaces). The previous `LIKE %q%` scan offered: no relevance ranking, no snippet, full table scans.

## Decision

Adopt SQLite **FTS5 + trigram tokenizer**:

1. `project_notes_fts` / `project_todos_fts` virtual tables + sync triggers; writes update the index automatically
2. Each query term is wrapped as a phrase (double-quote escaped) and combined into an implicit AND, preventing FTS query syntax injection
3. `bm25()` ranking with top 20 results; snippet windows with `<mark>` highlighting
4. Trigram requires a minimum of 3 characters: any query term shorter than 3 characters (e.g., 2-character Chinese words) automatically degrades to a `LIKE` scan (with `%`/`_` escaped)
5. When the FTS index is missing or a query errors, fall back to LIKE — search never breaks

## Consequences

- Positive: O(index) queries, relevance ranking, highlighted snippets; 2-character CJK queries lose no results
- Negative: index size and trigger maintenance cost; trigram cannot handle queries shorter than 3 characters (LIKE serves as the fallback)
- Implementation location: `internal/db/search.go` (split out of queries.go in 1.7.0)
