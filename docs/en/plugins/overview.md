---
title: Knowledge Source Import
order: 8
---

# Knowledge Source Import

> ⚠️ **Experimental**: the plugin system interface may change and is not a platform extension direction (see [ADR-0006](../adr/0006-scope-freeze.md)).

RepoNest supports idempotently importing documents into the knowledge base through Go scripts interpreted by yaegi.

## Built-in knowledge sources

| Source | Description |
|----|------|
| `claude` | Imports `~/.claude/projects/*/memory/*.md`, matching ownership by project name / repository path |

Startup auto-import can be toggled in **Settings → Plugins** (the `auto_import` config item). Manual triggering is available on the settings page.

## Idempotent import semantics

The runtime upserts by `(project_id, source, title)`: importing the same content again **updates** the existing note instead of creating a duplicate.
