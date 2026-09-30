---
title: Settings
order: 5
---

# Settings

The settings page has six tabs: Scan Roots, Code Standard, Author, Appearance, Plugins, and Actions.

## Scan Roots

- Add / remove scan root directories (a rescan is required after saving for changes to take effect)
- For the default roots seeded automatically on first launch, see [Getting Started](../getting-started.md)

## Code Standard

| Setting | Description | Default |
|------|------|------|
| Daily goal lines | Workday target; basis for the progress ring and card warnings | 500 (range 100-10000) |
| Max scan depth | Directory levels scanned down from scan roots | 2 (range 1-2) |

## Author

**Git author name**: the author filter for personal stats (same semantics as `git log --author`). When unset, it is read automatically from `git config user.name`.

## Appearance

Light / dark / follow system; applies instantly and is remembered.

## Plugins

- **Auto-import toggle**: whether to run all knowledge source imports automatically at startup
- **Loaded plugins**: load status and error messages per plugin directory; the **Reload** button hot-reloads
- **Knowledge import sources**: built-in (Claude memory) and plugin-registered importers, each with an individual **Import now** button

For plugin development, see the [Plugin Manual](../plugins/overview.md).

## Actions

- **Rescan all projects now**: triggers a full scan (equivalent to the dashboard button)
- **Import Claude memory**: triggers one import manually; you can jump to the knowledge base to review when it finishes
