---
title: RepoNest Documentation
---

# RepoNest Documentation

<p class="subtitle"><strong>The local-first memory layer for AI coding agents</strong> — discover your Git repos, understand what changed, capture knowledge as searchable Markdown, and hand it to any agent in one tool call.<br><strong>The local-first cross-agent project memory layer</strong>: automatically discovers local Git projects, understands what each project is doing now and what knowledge it has accumulated, and injects and retracts context automatically at session boundaries, so any agent reuses the same project context.</p>

<!--NAV_LINKS-->

## Product Positioning

RepoNest's core value: **turning local Git projects from "scattered across terminals and memory" into "searchable, reusable context"**.

Current priority statement:
1. **Local project understanding and knowledge context** is the first priority
2. Dashboard and statistics are supporting capabilities, not the main product narrative
3. Plugin/Web-only capabilities are no longer extended by default (PWA has been removed from the desktop main build, see [ADR-0008](adr/0008-pwa-removal.md))

The frontend default page is the **Knowledge Base**, with navigation order **Knowledge Base → Dashboard → Settings** (see [ADR-0006](adr/0006-scope-freeze.md)).

Core loop:

```
Discover local projects → Understand projects → Capture knowledge → Retrieve knowledge → Hand to AI
```

The "hand to AI" step lands at **session boundaries**: `reponest_scan` builds the knowledge base, `reponest_context` injects the full context in one call at session start, `reponest_handoff` records a structured handoff at session end, reusable across agents.

- **Discover**: automatically scans local Git repositories and intelligently groups them by Monorepo/single-repo
- **Understand**: project details automatically extract README summaries, tech stacks, dependencies, contributors, and activity
- **Record**: Markdown notes (categories/tags/version history), importable into Claude memory
- **Retrieve**: FTS5 full-text search (with short-CJK fallback); hits are locatable and explainable
- **AI use**: MCP session memory protocol (`reponest_scan` / `reponest_context` / `reponest_handoff`) + llms.txt, ready for direct consumption by Claude Code / Cursor and others

Feature tiering (core / supporting / experimental / deferred) and scope-freeze rules: see [ADR-0006](adr/0006-scope-freeze.md).

## About This Documentation

This manual covers RepoNest's core features and use cases. The docs are written in Markdown (the single content source, stored in the repository's `docs/` directory), built into HTML by `scripts/build-docs.mjs` and deployed to GitHub Pages.

- **Browse online**: <https://sky-jiangcheng.github.io/repo-nest/>, auto-updated with the master branch
- **Build locally**: `node scripts/build-docs.mjs` (depends on marked in `web/node_modules`)
- **Report issues**: <https://github.com/sky-jiangcheng/repo-nest/issues>

## Quick Navigation

| I want to… | Go to |
|------|------|
| Install and run the first scan | [Getting Started](getting-started.md) |
| Back up data / migrate to a new machine / uninstall | [Data & Backup](data-management.md) |
| View commits / goals / heatmap | [Dashboard](features/dashboard.md) |
| Write notes, search notes, recover old versions | [Knowledge Base & Notes](features/knowledge.md) |
| View a project's tech stack and dependencies | [Project Detail](features/project-detail.md) |
| Let Claude / Cursor read and write my knowledge base | [AI Integration](features/ai-integration.md) |
| Write a plugin or a knowledge-source importer | [Plugin Manual](plugins/overview.md) |
| Understand code layering and key decisions | [Architecture](architecture.md), [ADR](adr/index.md) |
| Understand storage structure and "why not let AI read git directly" | [Storage Optimization & AI Value](storage-optimization.md) |
