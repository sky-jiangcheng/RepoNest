---
title: Command Palette & Shortcuts
order: 6
---

# Command Palette & Shortcuts

## Keyboard Shortcuts

| Shortcut | Action | Description |
|--------|------|------|
| `⌘/Ctrl + K` | Open / close the command palette | Works globally |
| `↑` / `↓` | Move the selection | In-panel navigation |
| `↵ Enter` | Open the selected item | Jumps to the corresponding project |
| `Esc` | Close the panel | Focus returns to the trigger element |

> Also: typing `/` inside the editor opens the block panel (see [Knowledge Base](knowledge.md)); deleting todos and notes both use two-click confirmation.

## Command Palette

Opened with `⌘/Ctrl + K`; one input box combines three kinds of results:

- **Notes & todos**: FTS5 full-text search (title / content / todo titles); Enter jumps to the owning project
- **Projects**: filtered by project name; Enter opens the project detail page
- **Quick access**: recent projects shown when the input is empty

Accessibility implementation: ARIA combobox-in-dialog, focus trap + focus restore on close, keyboard selection driven by `aria-activedescendant`.

## Dashboard Combined Search

The dashboard search box matches the command palette's capabilities, plus:

- Star / unstar repositories directly in the results
- Real-time search with a 300ms debounce
- Auto-collapse on outside click
