---
title: Discovery
description: How Portfolio finds every Git repository across your configured project roots.
---

# Discovery

Discovery is the first half of every scan: Portfolio walks your configured [project roots](/docs/getting-started/configuration#managing-roots-from-the-cli) and finds every Git repository beneath them.

## How it works

1. **Multi-root traversal** — Each configured root is walked recursively. You can point Portfolio at `~/Projects`, `~/work`, or anywhere else repos accumulate; there is no limit on the number of roots.
2. **Git detection** — A directory counts as a project when it contains a `.git` entry. Nested repos are discovered too (a monorepo containing checked-out sub-repos yields each one).
3. **Smart ignoring** — Directories named `node_modules`, `.git`, `vendor`, `build`, `dist`, `target`, and `bin` are never entered. Configure the list via `ignored_paths`.
4. **Incremental re-discovery** — Each project is keyed by its `root_path`, so re-running discovery updates existing records instead of duplicating them.

Run discovery on demand:

```bash
portfolio discover        # shortcut for `portfolio projects discover`
portfolio projects discover
```

Discovery only registers *that a project exists*. To fill in languages, git stats, and documentation, run a [scan](/docs/concepts/metadata) — or just use `portfolio scan`, which does both.

## What gets recorded

For every discovered repository, Portfolio stores:

| Field | Description |
| --- | --- |
| `id` | Stable unique identifier (UUID) |
| `name` | Repository directory name |
| `root_path` | Absolute path — unique key for the project |
| `repository_type` | How the repo was identified (e.g. Git) |
| `discovered_at` / `updated_at` | Timestamps for incremental behavior |

## Choosing good roots

- Prefer *root directories that contain projects*, not your home directory — discovery walks everything below a root.
- Add narrow roots for big or unusual trees to keep scans fast.
- Re-run discovery after cloning new repositories; it is incremental and cheap.
