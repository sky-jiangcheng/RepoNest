---
title: RepoNest Documentation
---

# RepoNest Documentation

<p class="subtitle"><strong>The local-first memory layer for AI coding agents</strong> — discover your Git repos, understand what changed, capture knowledge as searchable Markdown, and hand it to any agent in one tool call.<br><strong>A local-first, cross-agent project memory layer</strong>: it automatically discovers local Git projects, understands what is happening in each project right now and what knowledge has been captured, and injects and withdraws that context at session boundaries, so any agent can reuse the same project context.</p>

<!--NAV_LINKS-->

## Product Positioning

The core value of RepoNest: **turning local Git projects from things scattered across terminals and memory into searchable, reusable context**.

Current priorities:

1. **Local project understanding and knowledge context** comes first
2. The dashboard and statistics are supporting capabilities, not the main product story
3. Plugin/web-only capabilities are no longer expanded by default (the PWA has been removed from the main desktop build; see [ADR-0008](adr/0008-pwa-removal.md))

The frontend landing page is the **knowledge base**, and the navigation order is **Knowledge base → Dashboard → Settings** (see [ADR-0006](adr/0006-scope-freeze.md)).

The core loop:

```
Discover local projects → Understand projects → Capture knowledge → Retrieve knowledge → Hand it to AI
```

The "hand it to AI" step happens at **session boundaries**: `reponest_scan` builds the knowledge base, `reponest_context` injects the full context once when a session starts, and `reponest_handoff` captures a structured handoff when it ends — reusable across agents.

- **Discover**: automatically scans local Git repositories and intelligently groups them (monorepo / single repo)
- **Understand**: project details automatically surface a README summary, tech stack, dependencies, contributors, and activity
- **Capture**: Markdown notes (categories / tags / version history), with Claude memory import
- **Retrieve**: FTS5 full-text search (with a fallback for short CJK queries); hits can be located and explained
- **AI consumption**: the MCP session memory protocol (`reponest_scan` / `reponest_context` / `reponest_handoff`) plus llms.txt, ready for direct consumption by Claude Code, Cursor, and others

Feature tiers (core / supporting / experimental / deferred) and the scope-freeze rules are described in [ADR-0006](adr/0006-scope-freeze.md).

## About These Docs

This manual covers RepoNest's core features and use cases. The docs are written in Markdown (the single source of content, stored in the repository's `docs/` directory), rendered to HTML by `scripts/build-docs.mjs`, and deployed to GitHub Pages.

- **Read online**: <https://sky-jiangcheng.github.io/repo-nest/> — updated automatically with the master branch
- **Build locally**: `node scripts/build-docs.mjs` (uses marked from `web/node_modules`)
- **Report issues**: <https://github.com/sky-jiangcheng/repo-nest/issues>

## Quick Navigation

| I want to… | Go to |
|------|------|
| Install RepoNest and run the first scan | [Getting Started](getting-started.md) |
| Back up data / migrate to a new machine / uninstall | [Data and Backup](data-management.md) |
| View commits / goals / heatmap | [Dashboard](features/dashboard.md) |
| Write notes, search notes, recover old versions | [Knowledge Base and Notes](features/knowledge.md) |
| See a project's tech stack and dependencies | [Project Detail](features/project-detail.md) |
| Let Claude / Cursor read and write my knowledge base | [AI Integration](features/ai-integration.md) |
| Write a plugin or a knowledge source importer | [Plugin Guide](plugins/overview.md) |
| Understand the code layering and key decisions | [Architecture](architecture.md), [ADRs](adr/index.md) |
| Understand the storage layout and "why not let the AI read git directly" | [Storage Optimization and AI Value](storage-optimization.md) |
