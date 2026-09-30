# RepoNest: Local Git Knowledge Base

English | [简体中文](./README.zh-CN.md)

**The local-first memory layer for AI coding agents.** Your agents (Claude Code, Cursor, OpenCode...) read code brilliantly and forget everything the moment the session ends — why a decision was made, what gotcha was discovered, what to do next. RepoNest keeps that knowledge on your machine, searchable, and hands it back to *any* agent in one tool call.

## Table of Contents

- [Features](#features)
- [Quick Start](#quick-start)
- [Build from Source](#build-from-source)
- [Project Grouping](#project-grouping)
- [Project Structure](#project-structure)
- [Naming Layers](#naming-layers)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

```
First run     →  reponest_scan      discover local repos in one call (pure MCP, no desktop app needed)
Session start →  reponest_context   load full project context in one call (tech stack/README/todos/notes/last handoff)
   ...
Session end   →  reponest_handoff   structured record: what was done, why, gotchas, next steps
   ↓
The next session — from any agent — picks up exactly where this one ended
```

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![React](https://img.shields.io/badge/React-19-61DAFB?logo=react)](https://react.dev)
[![TypeScript](https://img.shields.io/badge/TypeScript-7-3178C6?logo=typescript)](https://www.typescriptlang.org)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

> Single-binary Wails v2 desktop app (Go + React, embedded SQLite, zero CGO), cross-platform **macOS / Windows / Linux**.
> Works offline with no cloud dependency; AI reaches the same local database through the separately distributed [`reponest-mcp`](#install-the-ai-execution-interface-reponest-mcp) MCP server.
> Switching agents never loses context: a handoff written by Claude Code is read directly by Cursor.
> See [ADR-0006](docs/adr/0006-scope-freeze.md) and the [positioning brief](docs/positioning-brief.md) for positioning priorities, feature tiers and the scope-freeze rules.

**Why RepoNest?** Today's coding agents read code well but forget **the judgement you accumulated in these repos**: why it was designed this way, which traps you hit last time, what the next todo is. Each agent's private memory format is incompatible with the others, so switching tools means starting from zero. RepoNest keeps all of that in one local, searchable memory layer any agent can read and write — `reponest_context` injects it at session start, `reponest_handoff` captures a structured record at session end, and everything in between is retrieved on demand.

## Features

> Tier legend: **Core** (forms the discover → understand → capture → retrieve → AI loop) | **Support** (keeps the loop understandable) | **Experimental** (kept, not expanded) | **Deferred** (no further investment). See [ADR-0006](docs/adr/0006-scope-freeze.md).

### Knowledge Base (Core)

| Feature | Description |
|---------|-------------|
| Markdown notes | Title / tags / category (knowledge · log · idea · other) / pinning / cross-project move; drafts auto-saved |
| Block editor | Type `/` to open the block panel and insert Callout / Tabs / collapsible / code / Mermaid / math / table blocks, drag to reorder — output stays pure Markdown (**Experimental**: complex blocks frozen, see ADR-0006) |
| Rich rendering | highlight.js code highlighting, Mermaid diagrams, KaTeX math, GFM callouts and task lists |
| FTS5 full-text search | trigram + bm25 ranking, snippet highlighting, across notes and todos; short CJK queries fall back to LIKE |
| Version history | Auto snapshot on every save, line-level diff against any version, one-click restore |
| Global search | ⌘/Ctrl+K command palette + unified dashboard search (repos / notes / todos) |

### Repo Knowledge Mining (Core)

The project detail page auto-extracts: README excerpt, tech stack (20+ manifest detectors), language LOC share, dependency lists (npm / go.mod including block require / cargo), top contributors, activity stats and the recent commit stream. Results are cached in `repo_meta` to avoid repeated scans.

### AI-Ready Interface (Core)

| Channel | Description |
|---------|-------------|
| MCP Server | `reponest-mcp` stdio server with 13 tools (repo scan + context injection + session handoff + notes CRUD + project queries + search + two self-checks), plugs into Claude Code / Cursor etc. (the single AI execution interface) |
| `reponest_scan` | One-shot cold start: seeds default scan roots and scans synchronously to discover local Git repos. A pure-MCP install (no desktop app) can build the knowledge base too |
| `reponest_context` | Session start: load a project's full context in one call — tech stack / README excerpt / dependencies / recent commits / open todos / relevant notes (handoffs first) — replacing 3-4 chained queries |
| `reponest_handoff` | Session end: structured handoff — summary / changes / decisions / gotchas / next_steps rendered into a unified template, picked up automatically by the next session (from any agent) |
| llms.txt | `GenerateLLMsTxt` produces an LLM-facing overview of the knowledge base |
| Note export | Export any note as `.md` with YAML frontmatter |
| Claude memory import | One-click idempotent import of `~/.claude/projects/*/memory/*.md` as knowledge notes (**Support**) |
| Data trust audit | `reponest_integrity` runs 6 read-only checks: FTS index drift, orphan rows, schema shape vs version stamp, scan coverage, knowledge-cache freshness, version-snapshot orphans. Index drift makes search **silently miss results** with no other mechanism to detect it — this is the only way to find it |

### Dashboard & Stats (Support)

> Dashboard and stats serve the understandability of the core loop and stay out of the product's main path; the frontend default page is Knowledge, with navigation ordered Knowledge → Dashboard → Settings (see [ADR-0006](docs/adr/0006-scope-freeze.md)).

| Feature | Description |
|---------|-------------|
| Repo auto-discovery | Configure scan roots and all Git repos are found recursively; platform-adaptive defaults |
| Visual dashboard | Daily goal ring, project cards, trend line charts (7d / 30d / all), commit heatmap |
| Repo favorites | Favorited repos show full stat cards; the rest show names only, expand on demand |
| On-demand history | "Refresh history" backfills up to 365 days of daily stats per repo |
| Smart grouping | Auto-detects Monorepo vs single repo; manual split/merge runs in a single transaction (notes and todos move with the project) |
| Workday check | Custom daily code-volume standard with alerts when unmet |
| Status bar | Live latest-commit display (repo / branch / time, 30s cache) |

### Other

| Feature | Description |
|---------|-------------|
| Plugin system | yaegi in-process Go scripts + knowledge source importers ([docs](docs/plugins/overview.md); **Experimental**, platform infrastructure frozen) |
| i18n | Chinese / English one-click switch (react-i18next, zh-CN + en) |
| Single binary | One Go binary per platform, no runtime dependencies |

## Quick Start

### Download and Install

Grab the latest build for your platform from [Releases](https://github.com/sky-jiangcheng/repo-nest/releases).

**Option 1: direct download**

Download the archive for your platform from Releases, extract and run:

| Platform | Asset |
|----------|-------|
| macOS | `reponest-darwin-arm64.dmg` / `reponest-darwin-amd64.dmg` |
| Linux | `reponest-linux-amd64.tar.gz` |
| Windows | `reponest-windows-amd64.zip` |

**Option 2: one-line install script** (desktop app + `reponest-mcp` together)

| Platform | Command | Installs to |
|----------|---------|-------------|
| macOS | `curl -fsSL https://raw.githubusercontent.com/sky-jiangcheng/repo-nest/master/scripts/install.sh \| bash` | `/Applications/RepoNest.app` + `/usr/local/bin/reponest-mcp` |
| Linux | same as macOS | `/usr/local/bin/reponest` + `/usr/local/bin/reponest-mcp` |
| Windows | `iwr -useb https://raw.githubusercontent.com/sky-jiangcheng/repo-nest/master/scripts/install.ps1 \| iex` | `%LOCALAPPDATA%\RepoNest` (added to user PATH) |

macOS also ships via Homebrew (add the tap first, see [`packaging/`](packaging/README.md)):

```bash
brew tap sky-jiangcheng/repo
brew install --cask sky-jiangcheng/repo/reponest
```

On first launch the desktop window opens directly (Wails app, no browser):

1. Default scan roots are seeded automatically (HOME on macOS/Linux, a non-system drive on Windows)
2. Hit **Rescan** on the dashboard to discover repos — **the knowledge base is usable at this point**: write/search notes and hand context to AI right away
3. (Optional, affects dashboard stats only) favorite repos → **Refresh history** to backfill 365 days of stats

> **Just want the AI side?** Skip the desktop app: install `reponest-mcp` and have your agent call `reponest_scan` once.

See [Getting Started](docs/getting-started.md) for details.

### Install the AI Execution Interface (`reponest-mcp`)

AI clients use the separately distributed `reponest-mcp` (MCP stdio server) — **no desktop app required**; it reads the same local SQLite database. The install script above includes it; it can also be installed standalone:

| Method | Platform | Command |
|--------|----------|---------|
| Manual (no prerequisites) | All | Download `reponest-mcp-<target>.tar.gz` / `.zip` from [Releases](https://github.com/sky-jiangcheng/repo-nest/releases) |
| Homebrew | macOS | `brew tap sky-jiangcheng/repo && brew install --cask sky-jiangcheng/repo/reponest-mcp` |
| Homebrew | Linux | `brew tap sky-jiangcheng/repo && brew install sky-jiangcheng/repo/reponest-mcp` |
| Scoop | Windows | `scoop bucket add repo https://github.com/sky-jiangcheng/scoop-repo && scoop install repo/reponest-mcp` |

Register it with your AI client:

```bash
claude mcp add reponest -- "$(which reponest-mcp)"
```

#### See It Work in 30 Seconds

After registering, the first step is one call:

> **First use** (build the knowledge base, no desktop app needed):
> The agent calls `reponest_scan()` — seeds default scan roots, scans, and returns your local repo list.

Then every working session follows the same rhythm:

> **Session start** (a new agent takes over a project):
> "Continue working on the auth project."
>
> The agent calls `reponest_context({ project_name: "auth" })` — tech stack, todos and the previous session's handoff arrive in one call, and work begins.

> **Session end** (knowledge doesn't evaporate):
> The agent calls `reponest_handoff({ project_id: 1, summary: "finished the OAuth migration", gotchas: ["prod cookie keys need rotation"], next_steps: ["run the regression suite"] })` — the next session picks up from here, even from a different agent.

Ad-hoc questions work anytime:

> "What projects do I have locally? What did I note about auth?"

The agent chains `reponest_projects_list` → `reponest_notes_search` → `reponest_notes_read`, and writes new conclusions back with `reponest_notes_create`. Notes can be exported as `.md` with YAML frontmatter, or aggregated into an LLM-facing `llms.txt`. See [SKILL.md](SKILL.md) for the full tool list and workflows.

Manifests live in [`packaging/`](packaging/README.md); versions derive from `wails.json` and `sha256` values come from the actual release assets (Homebrew / Scoop rejecting a checksum mismatch is expected behaviour). Published to:

- **Homebrew**: [sky-jiangcheng/homebrew-repo](https://github.com/sky-jiangcheng/homebrew-repo)
- **Scoop**: [sky-jiangcheng/scoop-repo](https://github.com/sky-jiangcheng/scoop-repo)

> Desktop app likewise: `brew install --cask sky-jiangcheng/repo/reponest` (macOS), `scoop install repo/reponest` (Windows). Linux desktop ships as a tarball only.

### Data Directory

Config and the database live in the per-user app-data directory (schema migrates automatically on upgrade):

- **macOS**: `~/Library/Application Support/reponest/dashboard.db`
- **Windows**: `%APPDATA%/reponest/dashboard.db`
- **Linux**: `~/.config/reponest/dashboard.db`

Log locations are listed in [Troubleshooting](docs/troubleshooting.md).

## Build from Source

Requirements: **Go 1.25+**, **Node.js 20+** (frontend build), optional [Wails CLI](https://wails.io) v2.13+.

```bash
# Frontend deps and build (web/dist is go:embed'ed into the binary)
cd web && npm install && npm run build && cd ..

# Desktop app
go build -ldflags "-s -w" -o reponest .

# MCP server
go build -o reponest-mcp ./cmd/mcp/

# Or use the script
./scripts/build.sh
```

Dev mode: `wails dev` (frontend hot reload + Wails binding injection).

Test and check:

```bash
go test ./...            # full Go test suite (service/db/knowledge/scanner/diff...)
cd web && npm test       # vitest
cd web && npm run build  # strict tsc + build (ESLint status: see TODO.md)
```

## Project Grouping

| Scenario | Grouping rule |
|----------|---------------|
| Single repo | Parent directory contains one repo → the parent directory is the project |
| Monorepo | Parent directory contains multiple sub-repos → grouped as one project |
| Nested repos | Parent is itself a Git repo and subdirectories contain repos → split into separate projects |

The project detail page can **merge up** / **split down** to adjust the grouping level (single transaction; notes and todos move along).

## Project Structure

```
main.go                  # Wails entry: DB init, scan-root seeding, window & security headers
internal/
  app/                   # Wails binding layer: each method delegates to service in 1-3 lines
  service/               # Business core: scan pipeline, stats refresh, project/note/search/export
                         # (shared by the Wails desktop, CLI and MCP entrypoints)
  domain/                # Row types shared across layers
  db/                    # SQLite: schema/migrations + domain-split queries (projects/notes/...)
  core/git/              # Git provider abstraction (local CLI implementation)
  core/plugin/           # Plugin SPI + yaegi runtime
  stats/ knowledge/      # git log stats, repo knowledge mining
  scanner/ grouper/      # filesystem scanning, project grouping
  platform/              # OS differences: data dir, log path, default scan roots
  version/ diff/         # single version source, note line-level diff
cmd/
  mcp/                   # MCP stdio server (AI execution interface + self-check tools)
web/src/
  api/                   # types + transport (Wails/HTTP dual mode) + endpoints
  hooks/                 # useApiData (cache) / useDebouncedCallback / useScanPolling...
  pages/ components/     # pages and components (large pages split by domain)
  locales/ styles/       # zh-CN + en; design-system CSS
```

Architecture decisions live in the [ADRs](docs/adr/) (especially [ADR-0005 service layer](docs/adr/0005-service-layer.md)); layering and data flows are detailed in [Architecture](docs/architecture.md). The frontend-backend contract (Wails binding surface) is in [API Reference](docs/api/reference.md).

## Naming Layers

The brand name and machine identifiers are **intentionally different**: the display layer exists to be remembered; the identifier layer exists to stay stable (URLs, upgrade paths, data migration and external contracts never follow brand wording).

| Layer | Value | Used for |
|-------|-------|----------|
| Brand (display) | `RepoNest` | `productName`, in-app logo, docs and UI copy |
| Full display name | `RepoNest: Local Git Knowledge Base` | Window title, HTML `<title>`, README title |
| Repo & package identity | `repo-nest` | GitHub repo name, Go module name, npm package, docs site URL |
| Frozen identity (never follows brand) | `reponest` | Command / binary name (`outputfilename`), user data directory, MCP server name `reponest-mcp` and the `reponest_*` tool prefix |

Why the frozen identity stays: the `reponest` data directory has already survived two automatic migrations (gitboard → gitbuddy → reponest); another rename means a third data migration, and MCP tool names are an external contract with AI clients — renaming would invalidate existing configs and allowlists. Please do not "helpfully unify" these names.

## Documentation

| Doc | Contents |
|-----|----------|
| [Getting Started](docs/getting-started.md) | Install, first-run setup, scanning |
| [Feature Manual](docs/features/dashboard.md) | Dashboard / Knowledge / Project detail / Settings / command palette |
| [Knowledge Sources](docs/plugins/overview.md) | Plugin SPI, events, knowledge importers |
| [AI Integration](docs/features/ai-integration.md) | CLI, MCP, llms.txt |
| [API Reference](docs/api/reference.md) | Wails binding surface contract |
| [Architecture](docs/architecture.md) | Layers, data flows, key decisions |
| [Troubleshooting](docs/troubleshooting.md) | FAQ and log paths |
| [SKILL.md](SKILL.md) | Capability card for AI agents |
| [TODO.md](TODO.md) | Known issues and roadmap |

[Online docs](https://sky-jiangcheng.github.io/repo-nest/) (GitHub Pages, auto-deployed from master).

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for the dev environment and commit conventions. Quick start: **Go 1.25+**, **Node.js 20+**, Git.

```bash
cd web && npm install && npm run build && cd ..  # frontend build
go test ./...                                    # Go tests
cd web && npm test                               # frontend tests
wails dev                                        # dev mode (optional)
```

Architecture conventions: [docs/architecture.md](docs/architecture.md) and [docs/adr/](docs/adr/). Commits follow [Conventional Commits](https://www.conventionalcommits.org/). Report security issues via the [private disclosure channel](SECURITY.md).

## License

[MIT](LICENSE)
