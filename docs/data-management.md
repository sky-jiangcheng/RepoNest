---
title: Data & Backup
order: 8
---

# Data & Backup

RepoNest is a local-first app: no cloud services and no account system. All knowledge data (notes, todos, project metadata, statistics) lives only in a single SQLite database on your machine. What AI (e.g. Claude Code) reads via MCP is this same local data. Therefore **backup = copying one directory, migration = moving one directory**.

## Where the Data Lives

| Item | Location |
|------|------|
| Database (notes, todos, version history, statistics, scan root config) | `dashboard.db` in the data directory (see the table below) |
| Plugins | `plugins/` under the data directory |
| Logs | Platform log directory (see [Troubleshooting](troubleshooting.md#log-file-locations)) |

Data directory:

| Platform | Path |
|------|------|
| macOS | `~/Library/Application Support/reponest/` |
| Windows | `%APPDATA%\reponest\` |
| Linux | `~/.config/reponest/` |

## Backup

1. **Quit RepoNest** (the database runs in WAL mode; copying it while running may lose WAL content not yet flushed)
2. Copy the whole data directory `reponest/` (or at least `dashboard.db` plus `dashboard.db-wal` / `dashboard.db-shm`)

A cold backup before a disk replacement or a major upgrade is recommended; day to day you can also include the data directory in your Time Machine / File History style full-disk backups.

## Machine Migration

1. Old machine: quit the app and copy the data directory
2. New machine: [install RepoNest](getting-started.md#download-and-install), **start it once and then quit** (so the app creates the directory structure and schema)
3. Overwrite the `reponest/` directory on the new machine with the backed-up one and restart

Note: scan roots are stored in the database as **absolute paths**. If the new machine's username or directory layout differs, go to **Settings → Scan Directories** after launch, change them to local paths, and click **Rescan**. If repository locations changed, project grouping (Monorepo split/merge records) is re-identified against the new paths; notes and todos are unaffected.

## Reset

There is no "factory reset" button in the app. Reset = delete the data directory and restart:

```bash
# macOS example: quit the app first
rm -rf ~/Library/Application\ Support/reponest
```

After the restart the app re-seeds the default scan roots and rebuilds an empty database. **This wipes all notes and todos**; confirm you have a backup before doing it.

There is no built-in switch for "clear statistics, keep notes": if you need that, [export your notes](features/knowledge.md#export) from the knowledge base before deleting the database, then import them after the reset.

## Uninstall

1. Delete the binary (the install script places it at `/usr/local/bin/reponest`; on Windows: `%LOCALAPPDATA%\RepoNest\reponest.exe`, or wherever you placed it manually)
2. Delete the data directory (see the table above) and logs (`~/Library/Logs/reponest.log` and the other platform log paths)
3. If you installed the PWA: remove RepoNest from the system/browser app list

## Related Pages

- [Getting Started](getting-started.md): quick overview of data and log locations
- [Troubleshooting](troubleshooting.md): locating logs and common issues
- [Storage Optimization & AI Value](storage-optimization.md): why the knowledge base beats letting AI read git directly
