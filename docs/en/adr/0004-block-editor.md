# ADR-0004: Block Editor

- Status: Accepted
- Date: 2026-08-10 (backfilled; implemented in 1.5.7 / issue #19)
- Related: [ADR-0002](0002-c-end-repositioning.md)

## Background

Knowledge notes need an editing experience beyond a plain textarea (structured blocks benchmarked against GitBook), but **the storage format must remain pure Markdown** — the data belongs to the user and must stay free of editor lock-in.

## Decision

Implement a lightweight block editor on the React side (`web/src/components/BlockEditor.tsx`) rather than introducing a heavyweight rich-text framework:

1. Content is split into blocks by blank lines; `detectType()` identifies the block type (heading/code/callout/list/table/formula/divider…) from the first line's syntax
2. Each block is edited independently inline; `joinBlocks()` reassembles with `\n\n` — **the output is fully equivalent to hand-written Markdown**
3. Typing `/` opens the block palette, inserting templates for callout / tabs / details / code / Mermaid / formula / table, etc.
4. Block-level drag-and-drop reordering plus move up/down; the Markdown source ↔ blocks dual view switches at any time
5. Preview uses `renderMarkdownAsync` (Mermaid/KaTeX async rendering)

## Consequences

- Positive: zero new runtime dependencies; data stays readable and writable by any Markdown tool at any time; naturally compatible with version history diffs
- Negative: block detection is heuristic; extremely nested Markdown may merge into a single block (degrades to ordinary editing, lossless)
- 1.7.0: palette labels/descriptions/insert templates are i18n-ized (`buildPaletteItems(t)`)
