# RepoNest Positioning & Narrative Reframing Brief

## One-Sentence Positioning (external)

**RepoNest is the local-first "cross-agent project memory layer"**: it helps you quickly understand what recently happened in local Git projects and what searchable knowledge has been captured, and at session boundaries it automatically hands those memories to any AI agent to continue working with.

## User Value (the promise after subtraction)

1. **Discover which projects deserve attention, faster**
2. **Understand the current state of a project, faster** (README, tech stack, dependencies, activity)
3. **Capture scattered knowledge somewhere searchable** (Markdown + search + version history)
4. **Cut the time cost of "finding context"** (global search, command palette, project jumps)
5. **Hand project knowledge to AI to keep working** (llms.txt, export, MCP)

## Product Boundary (four tiers)

### 1) Must keep investing (core)
- Local repository discovery and project grouping
- Project knowledge capture and retrieval (notes, tags, version history, full-text search)
- Project understanding (README/tech stack/dependencies/contributors/activity)
- AI context outlets (llms.txt, note export, MCP)

### 2) Keep but stay restrained (supporting)
- Commit statistics, goal ring, heatmap, status bar
- Claude memory import
- i18n, themes, and other small features that are ceremonial but serve the core

### 3) No expansion by default (experimental)
- Advanced blocks of the complex block editor (Callout/Tabs/Mermaid etc. kept as-is, no new ones)
- Plugin/SPI platform capabilities
- API documentation, contract extension, platform entry points

### 4) Stop investing or remove from the main narrative
- PWA (removed from the desktop main build, see ADR-0008) / SEO / Web-only add-on capabilities
- Wording that could be mistaken for "online platform, team dashboard, HTTP server"

## External Narrative Reframing

### Unified external messaging (authoritative short copy)

> This section is the unified external messaging delivered by PR1 "positioning governance". All Releases / promotion / homepage openings / Issue templates must follow it; all other copy (including the "full" version below and the README opening) only elaborates on it and must not conflict with it.

**Chinese (three-sentence positioning)**

> RepoNest 是本地优先的跨 agent 项目记忆层：开会话一次注入上下文（`reponest_context`）、收会话结构化交接（`reponest_handoff`），让任何 agent 都从上一次结束的地方继续。

**English**

> RepoNest is the local-first memory layer for AI coding agents: one call injects full project context at session start (`reponest_context`), one call records a structured handoff at session end (`reponest_handoff`) — any agent picks up where the last one stopped.

**Release note**

> Positioning release: RepoNest is reframed as the local-first, cross-agent project memory layer; the session loop (`reponest_scan` → `reponest_context` → `reponest_handoff`) becomes the headline capability. Dashboard/statistics stay supporting capabilities.

### Recommended (full)
> RepoNest helps you turn local Git projects from "scattered across terminals and memory" into "a searchable, reusable memory layer". It quickly discovers the projects you care about, understands each project's current state, and captures notes, dependencies, tech stacks, and activity info; then it injects all of it to AI in one call at session start (`reponest_context`) and records a structured handoff at session end (`reponest_handoff`), reusable across agents.

### Paused (easily misleading)
- "Commit dashboard / data dashboard"
- "Local development data platform / observability platform"
- "Plugin hub / API platform"

## Why the UI Currently "Feels Unresponsive / Error-Prone"

What you ran into this time is three problems stacked together, rather than a single button bug:

1. **Backend contract and frontend call rules are misaligned**  
   The desktop uses Wails bindings while browser dev mode uses the API; frontend/backend assumptions were never unified, breaking the error chain.

2. **The error chain and user perception never close the loop**  
   After a request fails, the frontend lacks unified exception handling, retry, and recovery paths, so the experience feels "stuck".

3. **The product narrative is too broad and priorities are unfocused**  
   Once too many edge capabilities exist, testing and design boundaries stretch and regression cost rises.

## Suggested Product Evaluation Criteria (before adding any feature)

A new feature must hit at least one of these steps:
1. Discover local projects
2. Understand local projects
3. Capture project knowledge
4. Retrieve project knowledge
5. Hand it to AI

If it hits none of them, the default is to keep it out.

## Suggested Release & Communication Order

1. **First make the "positioning sentence + core loop" public**
2. **Then fix frontend error consistency and user recovery paths**
3. **Finally tier the disclosure of dashboard, plugin, and AI capabilities**  
   - Core capabilities first  
   - Dashboard as a "supporting capability"  
   - Plugin platform under "currently not expanding / kept only"

## Conclusion

Next, this brief can be landed directly into:
- `README.md`
- `docs/index.md`
- `docs/getting-started.md`
- Homepage and navigation copy
- GitHub Release / Issue templates
- External short copy (Chinese + English)

That way product perception and engineering implementation stop talking past each other.
