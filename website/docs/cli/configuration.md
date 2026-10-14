---
title: Configuration Commands
description: The portfolio config command group — manage scan roots and configuration.
---

# Configuration Commands

The `config` command group edits Portfolio's TOML configuration, with scan-root management as the day-to-day workhorse. For the underlying file format, see [Configuration](/docs/getting-started/configuration).

## `portfolio config set-root <path>`

Add a project root directory to discovery.

```bash
portfolio config set-root ~/Projects
portfolio config set-root /srv/work
```

Every directory passed here is recursively scanned by `portfolio discover` and `portfolio scan`.

## `portfolio config remove-root <path>`

Remove a project root directory.

```bash
portfolio config remove-root ~/Projects
```

Removing a root stops *future* scans from entering it; already-stored projects remain in the knowledge store until you remove the data separately.

## `portfolio config list-roots`

List all configured project root directories.

```bash
portfolio config list-roots
```

## Editing the file directly

The config file is plain TOML at the [platform config path](/docs/getting-started/configuration#file-location) — open it with any editor. The `config` commands exist because root management is the one part worth automating; the rest (logging, dashboard origins, database path) changes rarely.
