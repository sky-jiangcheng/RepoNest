---
title: Knowledge Base & Notes
order: 3
---

# Knowledge Base & Notes

The knowledge base is the app's home page and the cross-project notes hub: Markdown / block editor, tag organization, FTS5 full-text search, version history, and AI memory imports.

## Note Management

- **Create a note**: pick the owning project in "Quick create note" on the home page; or create one from the notes panel on the project detail page
- **Metadata**: title (falls back to the first line when empty), tags (comma-separated), kind (knowledge / log / idea / other), pinned
- **Move across projects**: one click on the "Linked project" dropdown while editing
- **Draft autosave**: edits persist locally in real time, so an unexpected close loses nothing

## Block Editor

Type `/` to open the block panel and insert structured blocks:

| Block | Description |
|----|------|
| Callout | TIP / WARNING / NOTE callout blocks |
| Code block | With language highlighting |
| Mermaid diagram | Flowcharts / sequence diagrams, etc. |
| Math formula | Rendered by KaTeX |
| Todo list / table / divider | Common structures |
| Collapse block / Tabs | `<details>` and `{% tabs %}` |

- Blocks can be drag-sorted, moved up/down, and deleted individually
- **Markdown ↔ blocks, switchable anytime**: the storage format is always plain Markdown, openable by any editor

## Rich Rendering

highlight.js code highlighting, Mermaid diagrams, KaTeX math formulas, GFM callouts and task lists.

## Search

- The home search box auto-focuses; "Ask the knowledge base" mode returns answer snippets ranked by relevance
- FTS5 trigram + bm25 ranking, snippet highlighting for matched terms; short CJK queries automatically fall back to LIKE
- Covers notes and todos; the `⌘/Ctrl+K` command palette is available anytime

## Version History

Every save automatically creates a snapshot (the last 50 are kept). The **History** button on a note card:

- Version list (time, title)
- Line-level LCS diff of any version against the current one (+/- markers)
- One-click restore to any historical version

## Import Claude Memory

One click idempotently imports `~/.claude/projects/*/memory/*.md` as knowledge notes (re-imports update the existing notes). **Settings → Plugins** lists the import sources and allows manual triggering.

## Export

The note card's **Export .md** copies Markdown with YAML frontmatter (title / tags / project / kind / updated time) to the clipboard. For batch AI consumption, see llms.txt under [AI Integration](ai-integration.md).
