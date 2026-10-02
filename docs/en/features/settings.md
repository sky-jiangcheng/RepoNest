---
title: Settings
order: 5
---

# Settings

The Settings page has six tabs: Scan Directories, Code Target, Author, Appearance, Plugins, and Actions.

## Scan Directories

- Add / remove scan roots (a rescan is required after saving for changes to take effect)
- For the default roots seeded automatically on first launch, see [Getting Started](../getting-started.md)

Only directories added here get scanned — which is exactly what step "① Configure scan directories" in [Getting Started](../getting-started.md#2-run-a-scan) is for.

## Code Target

| Setting | Description | Default |
|---------|-------------|---------|
| Daily goal (lines) | The workday check threshold; drives the progress ring and card alerts | 500 (range 100-10000) |
| Maximum scan depth | Directory levels to descend below each scan root | 2 (range 1-2) |

## Author

**Git author name**: the author filter for personal statistics (same semantics as `git log --author`). When unset, it is read automatically from `git config user.name`.

## Appearance

Light / Dark / Follow system — pick one; changes apply immediately and are remembered.

## Plugins

- **Auto-import toggle**: whether to run all knowledge source imports automatically at startup
- **Loaded plugins**: load status and error messages for each plugin directory; the **Reload** button hot-reloads
- **Knowledge import sources**: built-in (Claude memory) and plugin-registered importers, each with an **Import Now** action

For plugin development, see the [Plugin Handbook](../plugins/overview.md).

## Actions

- **Rescan all projects now**: triggers a full scan (equivalent to the dashboard button)
- **Import Claude memory**: triggers one import manually; once it finishes, jump to the knowledge base to review
