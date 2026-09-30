---
title: AI Integration (MCP/llms.txt)
order: 7
---

# AI Integration

RepoNest provides AI agents with read channels and self-check tools, all reusing the same `internal/service` implementation (identical behavior to the desktop app).

## Value Positioning: RepoNest Is a Data Source for AI, Not AI Itself

**RepoNest calls no large language model** — there is no OpenAI / Anthropic / any API key configuration in the code, no model selection, no endpoint settings. All of its AI features amount to "exporting project knowledge for AI tools to consume", rather than built-in chat or generation. More precisely:

- RepoNest is the **data foundation**: it models raw git information into a structured, indexable, materialized local knowledge base (see [Storage Optimization & AI Value](../storage-optimization.md));
- AI tools (Claude Code / Cursor, etc.) **consume** this data layer through the 13 MCP tools, fetching on demand and searching precisely.

For the argument of "why this beats letting AI read git directly", see [Storage Optimization & AI Value](../storage-optimization.md).

### Configuration Keys (none tied to models / APIs)

The whitelist in `internal/service/config.go` contains only 4 keys, **none of them LLM-related**:

| Key | Type | Default | Purpose |
|----|------|------|------|
| `auto_import` | 0 / 1 | `1` | Whether to import Claude memory automatically (**the only AI-related setting**) |
| `daily_code_standard` | integer | `500` | Daily code line goal, used for dashboard goal display. The name says "code standard" but it is not an AI spec; easy to misread |
| `scan_depth` | integer | `2` | Scan directory depth |
| `git_author` | string | System git user | Affects attribution of "my" stats / heatmap |

## MCP Server (`reponest-mcp`)

MCP is the only AI execution interface (the `reponest` CLI is not shipped with releases). stdio protocol, single in-process DB open, 13 tools (including 4 write operations: scan + note create/update + session handoff):

| Tool | Description | R/W |
|------|------|------|
| `reponest_scan` | Cold start: seed the default scan roots and scan synchronously to discover local Git repositories (usable from a pure MCP install, no desktop app needed) | Write |
| `reponest_context` | Inject the full project context in one call at session start (tech stack / README / todos / highly relevant notes, handoff notes pinned first) | Read |
| `reponest_handoff` | Structured handoff at session end (summary/changes/decisions/gotchas/next_steps), persisted and pinned for the next `reponest_context` read; the persisted note carries a `handoff` tag and `reponest_notes_update` refuses to overwrite it | Write |
| `reponest_notes_list` | All notes | Read |
| `reponest_notes_search` | FTS5 search (query) | Read |
| `reponest_notes_read` | Read a note by ID | Read |
| `reponest_notes_create` | Create a knowledge note | Write |
| `reponest_notes_update` | Update note content and metadata (partial update; refuses to overwrite `handoff` protocol notes) | Write |
| `reponest_projects_list` | All projects | Read |
| `reponest_projects_stats` | Project stats (by id) | Read |
| `reponest_ask` | Question-style retrieval, top-5 text | Read |
| `reponest_agent_score` | Check local AI readiness (DB/notes/search/MCP/llms.txt/SKILL.md/i18n) | Read |
| `reponest_integrity` | Audit data trustworthiness (FTS index drift / orphan rows / cache freshness / coverage) | Read |

### Readiness vs Data Trustworthiness

The two self-check tools answer **different questions** and must not be conflated:

| Tool | Answers | Cannot answer |
|------|-----------|----------------|
| `reponest_agent_score` | Is this installation configured correctly? (any notes? does MCP work? is i18n complete?) | Whether the data itself is correct |
| `reponest_integrity` | Can the data still be trusted? (has the index drifted? any orphan rows? is the cache fresh?) | Whether configuration is complete |

**Why the second one exists**: once the FTS5 index gets out of sync with `project_notes`, search **silently returns fewer results**, while `agent_score` still reports "Search operational" — it tests the pathway, not the content. `reponest_integrity` compares the index's actual document count against the FTS5 `_docsize` shadow table (for external-content tables, `SELECT rowid` reads the content table and cannot detect the drift), and checks that the sync triggers are all in place. See the [checklist in `internal/integrity`](../../internal/integrity/integrity.go).

When an AI suspects "the search results seem incomplete", it should run `reponest_integrity` first before concluding that the knowledge base simply holds little content.

### Connecting Claude Code

```bash
claude mcp add reponest -- /path/to/reponest-mcp
```

Or write it into `.mcp.json` (project level) / `~/.claude.json` (user level):

```json
{
  "mcpServers": {
    "reponest": { "command": "/usr/local/bin/reponest-mcp", "args": [] }
  }
}
```

### Automatic Handoff at Session End (SessionEnd hook)

The real meaning of "zero-cost accumulation" here goes beyond "the agent remembering to call it on its own" — relying on goodwill equals having no guarantee. Claude Code's **SessionEnd hook** makes the handoff part of the session lifecycle: as soon as a session ends, the handoff is guaranteed to happen, independent of the agent's memory. This is also the path that lets a new user feel "the next session starts with full context" from day one.

Write to `.claude/settings.json` (project level, shared with collaborators through the repository):

```json
{
  "hooks": {
    "SessionEnd": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/reponest-handoff.sh"
          }
        ]
      }
    ]
  }
}
```

The companion script `.claude/hooks/reponest-handoff.sh` (remember `chmod +x`):

```sh
#!/bin/sh
# SessionEnd hook: make the handoff automatic instead of relying on the
# agent's goodwill. One bounded headless turn costs a fraction of the
# session it preserves; "the agent will remember to call handoff" costs
# the entire exit record whenever it forgets.
set -u
cd "${CLAUDE_PROJECT_DIR:-$(pwd)}" || exit 0
command -v claude >/dev/null 2>&1 || exit 0
claude -p --mcp-config .mcp.json \
  'This session ended. Call reponest_handoff for the current project: a concise summary plus next_steps. If the project cannot be resolved, call reponest_context once first. Do nothing else.' \
  >/dev/null 2>&1 || true
```

Design points and costs (spelled out, so the narrative stays honest):

- **Trigger timing**: fired by Claude Code at session end; the reason (`clear` / `logout` / `prompt_input_exit` / `other`) is written to stdin as JSON; the script can read stdin and skip as needed (e.g. "just cleared the context" does not need a handoff).
- **Cost**: one bounded headless turn (`claude -p`), a fraction of the full session context it preserves; the hook has a default timeout (60s) and the script ends silently with `|| true`, without affecting session exit.
- **Degradation path**: skips immediately when the `claude` CLI is missing; when MCP is not registered, the headless call fails and is swallowed — the protocol requires the handoff to be written by `reponest_handoff`, and the hook only guarantees "it always fires", never bypassing the protocol to write the database directly.
- **Protocol protection**: handoff notes carry a `handoff` tag and `reponest_notes_update` refuses to overwrite them (against accidental clobbering); the next session renders them in full, pinned by `reponest_context`.
- The hook's event names and config fields evolve with Claude Code versions; verify against `claude --help` and the official hooks documentation before integrating.

### Connecting Cursor / Other MCP Clients

Add the same `command` pointing to the `reponest-mcp` binary in the client's MCP configuration (Cursor: `Settings → MCP → Add Server`).

## llms.txt and Markdown Export (in-app)

- **llms.txt**: `GenerateLLMsTxt` generates a knowledge base overview in Markdown (project directory + tech stack + the 20 most recent knowledge notes), suitable for feeding an LLM to establish context. **llms.txt is an export format** (generated in-app, not distributed with the repository) and is not a standalone product direction (see ADR-0006)
- **Note export**: any note can be exported as `.md` with YAML frontmatter (`ExportNoteAsMarkdown`)
- **Claude memory import**: idempotent import of `~/.claude/projects/*/memory/*.md` (see [Knowledge Base](knowledge.md))

## agent-score Self-Check

agent-score has been merged into the MCP tool `reponest_agent_score`; no separate build is needed. Call it from any MCP client to get the 7-point AI readiness score.

## Why Not Just Let AI Read the Git Repository

A common question: since Claude Code / Cursor can run `git log` and read files directly, why go through RepoNest? The core answer is **cost and determinism**:

- **Reading is expensive**: every direct `git log` or file-by-file scan is one-off consumption; repeated reads = repeated tokens;
- **Reading is messy / incomplete**: large repositories inevitably exceed the context window and get truncated, and the model may hallucinate or skip binaries / `.gitignore`;
- **Reading is slow**: recomputing stats every time turns second-scale work into minute-scale work.

RepoNest parses and materializes raw git data into local SQLite **once, at scan time** (`daily_stats` pre-aggregates stats, `repo_meta` caches mining results, `project_notes_fts` builds the full-text index); afterwards AI only fetches on demand and hits precisely through the MCP tools. For the full comparison and storage details, see [Storage Optimization & AI Value](../storage-optimization.md).

## Skill Card for AI Agents

[SKILL.md](https://github.com/sky-jiangcheng/repo-nest/blob/master/SKILL.md) at the repository root is a capability card for agents to read (commands, tool table, paths) and can be fed in directly.
