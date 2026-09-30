---
title: Project Detail
order: 4
---

# Project Detail

The project detail page shows a single project's stats, repository knowledge mining, heatmap, and trends, plus its notes and todos.

## Page Structure

### Header Card

- Project name and path, auto/manual grouping marker
- **Split down / merge up**: adjust the grouping level (single transaction; notes and todos move with it)
- Aggregate stats: sub-repository count, active days, file changes, added/deleted lines

### Project Overview (repository knowledge mining)

Extracted automatically from the repository working tree and cached in `repo_meta` (mined live the first time, "from cache" afterwards):

| Content | Description |
|------|------|
| README excerpt | Up to 200 lines / 8KB |
| Tech stack | 20+ manifest recognition (package.json / go.mod / Cargo.toml / pom.xml / Dockerfile …) |
| Language share | Bar chart of top languages by lines of code |
| Dependencies | npm / go.mod (including block require) / Cargo direct dependencies |
| Top contributors | Top 5 from `git shortlog -sn` |
| Activity | Total commits / last 30 days / active days within 90 / active months / latest commit |
| Recent commits | Latest 8 (time / branch / author / message) |

### Commit Heatmap (project scope)

Year-round commit density for **this project's repositories** (independent of the dashboard's global heatmap).

### Trend Chart

Three-line comparison of additions / deletions / file changes, switchable across **last 7 days / last 30 days / all**.

### Sub-repository List

Each repository's path, cumulative additions/deletions, and latest stats details.

### Notes & Todos Panel

Project-scoped knowledge notes (same source as the [Knowledge Base](knowledge.md)) and todos (add/delete / complete / reorder).
