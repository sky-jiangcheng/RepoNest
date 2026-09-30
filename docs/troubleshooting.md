---
title: Troubleshooting
order: 10
---

# Troubleshooting

## Common Issues

### Dashboard shows "No repositories found yet"

**Cause**: scan roots are unconfigured or contain no Git repositories.

**Fix**: open **Settings → Scan Directories**, add directories containing Git repositories, and click **Rescan**.

> The first scan only registers repository names; it does not scan historical commit data.

---

### Repo cards show only the name, no statistics

**Cause**: the repository is not starred, or history was not backfilled after starring.

**Fix**:
1. Click the star on the card to star the repository
2. Click the **Refresh History** button and wait for the git log backfill to finish

---

### Statistics show zero

**Cause**: history has not been backfilled yet, or there were no commits on the selected dates.

**Fix**: click **Refresh History** on the repo card; the system scans that repository's git log for the past 365 days. Confirm git is installed and on PATH (when launched from Finder/Dock the app automatically patches in common PATH entries).

---

### How do I find a repository quickly?

The dashboard search box searches repositories / notes / todos together; or use the `⌘/Ctrl+K` command palette (see [Command Palette](features/command-palette.md)). In the search results you can click the star directly to toggle starring.

---

### How do I adjust project grouping?

When multiple repositories are grouped incorrectly (or the opposite), open the project detail page and click the header action buttons:

- **Split down**: split a multi-repo project into one project per repository (the original project keeps the first repository's notes and todos)
- **Merge up**: merge sibling projects under the same parent directory (repositories, notes, and todos move along)

Both operations are a single transaction and leave no half-finished state on failure.

---

### Knowledge base search misses two-character Chinese words?

The FTS5 trigram index matches a minimum of 3 characters; shorter queries automatically fall back to a `LIKE` full-table scan (hits still work, just without relevance ranking). Nothing to fix.

---

### Plugins not loading?

1. Check the plugin directory structure: `<config directory>/plugins/<plugin name>/plugin.go`, exporting `Name` and `Init`
2. **Settings → Plugins → Reload**, then check the load status and error messages
3. See the [Plugin Manual](plugins/overview.md)

---

## Log File Locations

| Platform | Log path |
|------|---------|
| macOS | `~/Library/Logs/reponest.log` |
| Linux | `$XDG_STATE_HOME/reponest/reponest.log` (default `~/.local/state/reponest/reponest.log`) |
| Windows | `%APPDATA%\reponest\logs\reponest.log` |

Logs also record the PATH environment and plugin runtime state; attach them when reporting issues.

## Data Safety Notes

- RepoNest is a **local-first** desktop app: it listens on no network ports and uploads no data; statistics are read via the local `git` CLI
- Database and configuration live in the user application data directory (see [Getting Started](getting-started.md) for locations); uninstalling does not delete them automatically; see [Data & Backup](data-management.md) for backup, migration, and full removal
- Desktop WebView responses carry security headers such as CSP (`default-src 'self'` etc.)

## Reporting Issues

Found a bug or have a feature suggestion? File it at [GitHub Issues](https://github.com/sky-jiangcheng/repo-nest/issues) (include version, platform, and log excerpts). For security vulnerabilities use the [private reporting channel](https://github.com/sky-jiangcheng/repo-nest/security/advisories/new).
