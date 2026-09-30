---
title: Knowledge Source Import
order: 8
---

# Knowledge Source Import

> ⚠️ **Experimental**: the plugin system interface may change and is not an extension direction for the platform (see [ADR-0006](../adr/0006-scope-freeze.md)).

RepoNest supports idempotent imports of documents into the knowledge base via Go scripts interpreted by yaegi.

## Built-in Knowledge Sources

| Source | Description |
|----|------|
| `claude` | Imports `~/.claude/projects/*/memory/*.md`, matching ownership by project name / repository path |

Startup auto-import can be toggled under **Settings → Plugins** (the `auto_import` config key). For manual triggering, see the Settings page.

## Idempotent Import Semantics

The runtime upserts by `(project_id, source, title)`: importing the same content again **updates** the existing note instead of creating a duplicate.
