---
title: Workspaces
description: The portfolio workspace command group — create, add, remove, list, and show project groups.
---

# Workspaces

The `workspace` command group manages named groups of projects — typically the microservices that form one product. Workspaces are grouping-only; they never affect scanning or discovery. See [Workspaces](/docs/concepts/workspaces) for the concept.

## `portfolio workspace create <name>`

Create a workspace.

```bash
portfolio workspace create commerce -d "The commerce product services"
```

| Flag | Description |
| --- | --- |
| `-d`, `--description <text>` | Optional description |

Names are unique — creating a duplicate name fails.

## `portfolio workspace add <workspace> <project-id-or-name>`

Add a project to a workspace.

```bash
portfolio workspace add commerce orders-service
portfolio workspace add commerce 5f0c…     # by project ID works too
```

Projects can be addressed by ID, exact name, or a unique name substring; ambiguous matches are rejected with the candidates listed. Adding an already-added project is a no-op.

## `portfolio workspace remove <workspace> <project-id-or-name>`

Remove a project from a workspace. The project itself is not affected.

## `portfolio workspace list`

List all workspaces with member counts.

## `portfolio workspace show <name>`

Show a workspace's members **with their analysis freshness** — the staleness gate at a glance:

```bash
$ portfolio workspace show commerce
Workspace: commerce

PROJECT            ANALYSIS FRESHNESS     ID
orders-service     ⚠ stale (+14 commits)  5f0c…
payments-service   ✓ fresh                9a41…
```

`✓ fresh` means the stored analysis matches the repository's current HEAD; `⚠ stale` means the repository has moved on (and by how much); `– none` means no analysis is stored yet.

## Related

- [Workspaces concept](/docs/concepts/workspaces) — freshness and the staleness gate
- [`portfolio projects`](/docs/cli/projects) — the underlying project queries
