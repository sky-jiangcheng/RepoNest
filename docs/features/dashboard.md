---
title: Dashboard
order: 2
---

# Dashboard

The dashboard shows daily commit statistics for all projects, goal progress, and a year-round commit heatmap.

## Page Structure (top to bottom)

### Goal Progress Area

- **Progress ring**: today's (or the selected date's) personal added lines vs the daily goal; the goal is hidden on non-workdays
- **Subtitle**: gap to goal, personal additions / file count / repositories touched
- **Summary bar**: team additions / deletions, personal additions / files, total todos

### Commit Heatmap

Daily commit density over the last 52 weeks (GitHub style); clicking any cell jumps to that date to view per-project data for the day. The stats scope follows the current git author (Settings → Author).

### Controls

- **Date switcher**: yesterday / today / any date
- **Combined search box**: real-time search across repositories, notes, and todos (300ms debounce; results can be starred or opened directly)
- **Filter**: all / starred only
- **Sort**: name / personal additions / file count / repository count
- **Rescan**: triggers an async scan after a two-click confirmation, shows live progress, and refreshes data automatically on completion

### Project Grid

- **Starred repositories**: full stat cards (today's additions, goal progress bar, repos/files/added/deleted, net change, team totals, todo & note badges)
- **Other repositories**: minimal cards (name + star button); star them to expand when needed

## Feature Details

### Starring Repositories

Click the star to star a repository: the card expands into full stats and gains a **Refresh history** button (on-demand backfill of the last 365 days). Unstarring immediately collapses it back to a minimal card.

### Goal Progress

Configure the daily goal in lines under **Settings → Code Standard** (default 500, range 100-10000).

- The progress ring and the card's goal bar show completion; reaching the goal shows 🎉
- When a workday is below the goal, cards show a "below standard" warning; weekends never warn

### Project Grouping

Projects group automatically by parent directory (Monorepo detection); for manual splitting/merging see [Project Detail](project-detail.md).
