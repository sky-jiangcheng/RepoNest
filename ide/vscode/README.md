# RepoNest for VS Code (M5 skeleton)

One VSIX covers the whole fork family — VS Code, Cursor, Windsurf (ADR-0009
step two). This is a **thin client**: commands call `reponest-mcp` over stdio
and the loopback-only headless service; every answer comes from the shared
`internal/service` layer. Zero logic lives here.

## What works in this skeleton

| Surface | What it does |
|---|---|
| Command `RepoNest: Load Session Context` | `reponest_context` via MCP; result opens in the RepoNest output channel |
| Command `RepoNest: Write Session Handoff` | prompts for summary / changes / next steps, then `reponest_handoff` |
| Command `RepoNest: Search Knowledge` | FTS5 search; hits render in the sidebar and open read-only |
| Sidebar "Knowledge Search" | tree view bound to the search command's results |
| Status bar `⌒ RepoNest` | click = load context; tooltip shows headless-service reachability |
| Command `RepoNest: Register MCP Server` | runs `scripts/reponest-init/index.mjs --with-hook` in a terminal |

## Build & try it

```bash
cd ide/vscode
npm install
npm run compile
# then: open this folder in VS Code and "Run Extension" (F5), or package:
npx @vscode/vsce package   # produces reponest-vscode-0.1.0.vsix
```

The extension needs the `reponest-mcp` binary on PATH (or installed via brew /
scoop / Releases). Point `reponest.serverUrl` at a running
`reponest server` for the health check; all commands work without it.

## Deliberately not in the skeleton

- Packaging/CI for the marketplace listing (needs publisher + icon assets)
- Auto-starting the headless server from the extension (follows the dsh plugin's
  lifecycle practice; see ADR-0009 open items)
- Status-bar "time since last handoff" — needs a service-side API for handoff
  timestamps; the current button is a visible entry point instead
