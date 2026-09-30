---
title: Architecture Decisions (ADR)
order: 21
---

# Architecture Decision Records (ADR)

| # | Title | Status |
|------|------|------|
| [0001](0001-plugin-platform.md) | Plugin platform (M1-M4: HTTP server / RBAC / PG-ES / K8s) | Superseded (by 0002) |
| [0002](0002-c-end-repositioning.md) | Consumer-side repositioning: local-first "second brain for code projects" + in-process plugins | Accepted |
| [0003](0003-fts5-search.md) | FTS5 trigram full-text search | Accepted |
| [0004](0004-block-editor.md) | Block editor (output stays pure Markdown) | Accepted |
| [0005](0005-service-layer.md) | Service layer refactor (service / app / domain layering) | Accepted |
| [0006](0006-scope-freeze.md) | Scope freeze and feature tiering (core loop first) | Accepted |
| [0007](0007-session-memory-protocol.md) | Session memory protocol (context / handoff dual tools) | Accepted |
| [0008](0008-pwa-removal.md) | PWA moved out of the main desktop build (implements the deferred tier of ADR-0006) | Accepted |

## Conventions

- One ADR per major irreversible decision: context → decision → consequences
- Superseded ADRs keep their original text and banner; they are not deleted
- New ADRs are numbered incrementally starting from `0008`, with file names `NNNN-kebab-title.md`
