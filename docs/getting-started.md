---
title: Getting Started
order: 1
---

# Getting Started

RepoNest is a local-first desktop app (Wails v2, single file, zero runtime dependencies). Its core value is a **cross-agent project memory layer**: it automatically discovers local Git repositories and quickly captures notes, dependencies, tech stacks, and activity info for you and any AI agent to retrieve and reuse. Dashboard and statistics are supporting capabilities, not the main product entry.

## Download and Install

The install script installs both the desktop app and `reponest-mcp` (the MCP server used by AI clients).

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/sky-jiangcheng/repo-nest/master/scripts/install.sh | bash
```

macOS installs to `/Applications/RepoNest.app`, Linux installs to `/usr/local/bin/reponest`; both place `reponest-mcp` into `/usr/local/bin`.

### Windows

```powershell
iwr -useb https://raw.githubusercontent.com/sky-jiangcheng/repo-nest/master/scripts/install.ps1 | iex
```

Or download the binary for your platform from [GitHub Releases](https://github.com/sky-jiangcheng/repo-nest/releases).

## First Launch

The desktop window opens directly on launch (Wails app, no browser needed).

### 1. Configure Scan Directories

The first launch automatically seeds the default scan roots:

| Platform | Default scan scope |
|------|-------------|
| macOS | Current user HOME directory |
| Linux | Current user HOME directory |
| Windows | All disks except C: |

Change them under **Settings → Scan Directories** (add/remove supported, scan depth 1-2 levels).

### 2. Run a Scan

Click **Rescan** on the dashboard; the app recursively discovers all Git repositories under the scan roots and intelligently groups them into projects by directory.

> The first scan only registers repositories and projects; it does not prescan historical commit data.
>
> **At this point the knowledge base is already usable**: you can write notes, search, and hand off to AI (`reponest_context` / `reponest_handoff`). The starring and backfill below only affect dashboard statistics, not the knowledge base.

### 3. Star Repositories (optional, dashboard statistics)

Click the star on the dashboard to star the repositories you care about. Once starred, cards show full statistics (today's additions / files / repo count / net growth / team total).

### 4. Backfill Historical Data (optional, dashboard statistics)

Click the **Refresh History** button on a starred repo card to backfill that repository's daily statistics for the past 365 days (the progress ring and heatmap fill in immediately).

## Core Features at a Glance

| Feature | Tier | Description |
|------|------|------|
| Knowledge Base | Core | Cross-project note hub: Markdown, tags, pinning, FTS5 full-text search, version history |
| Repository knowledge mining | Core | Automatically extracts README, tech stack, language breakdown, dependencies, contributors, activity |
| Project understanding & retrieval | Core | Command palette `⌘/Ctrl+K`, global search, project context jumps |
| AI integration | Core | llms.txt, note export, MCP server (with agent-score self-check, see [AI Integration](features/ai-integration.md)) |
| Dashboard | Supporting | Daily goal progress ring, project cards, trend line chart, commit heatmap |
| Plugin system | Experimental | yaegi in-process Go scripts + knowledge-source import (see [Knowledge Source Import](plugins/overview.md)) |

## Data & Log Locations

| Item | macOS | Windows | Linux |
|------|-------|---------|-------|
| Database | `~/Library/Application Support/reponest/dashboard.db` | `%APPDATA%\reponest\dashboard.db` | `~/.config/reponest/dashboard.db` |
| Plugin directory | `…/reponest/plugins/` | `…/reponest\plugins\` | `…/reponest/plugins/` |
| Logs | `~/Library/Logs/reponest.log` | `%APPDATA%\reponest\logs\reponest.log` | `$XDG_STATE_HOME/reponest/reponest.log` (default `~/.local/state/reponest/`) |

The schema migrates automatically on upgrade; data needs no manual handling. For detailed troubleshooting see [Troubleshooting](troubleshooting.md).

## Build from Source

**Prerequisites**: Go 1.25+, Node.js 20+; [Wails CLI](https://wails.io) v2.13+ optional for development debugging.

```bash
# Frontend deps & build (output is go:embed-ed into the binary)
cd web && npm install && npm run build && cd ..

# Desktop app
go build -ldflags="-s -w" -o reponest .

# MCP server
go build -o reponest-mcp ./cmd/mcp/
```

Dev mode: `wails dev` (frontend hot reload + Wails binding injection). Tests: `go test ./...`; frontend `npm test` / `npm run build` (strict tsc checks).

## Next Steps

- [Data & Backup](data-management.md): backup, machine migration, reset, and uninstall
- [Dashboard](features/dashboard.md): goal progress, heatmap, and sorting
- [Knowledge Base & Notes](features/knowledge.md): block editor and full-text search
- [Settings](features/settings.md): scanning, standards, appearance, plugins
