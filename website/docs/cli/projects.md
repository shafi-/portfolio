---
title: Projects
description: The portfolio projects command group — list, search, get, and discover.
---

# Projects

The `projects` command group queries the [knowledge store](/docs/concepts/knowledge-store) directly — everything here is read-only.

## `portfolio projects list`

List all discovered projects.

```bash
portfolio projects list
```

## `portfolio projects search <query>`

Search projects by name.

```bash
portfolio projects search billing
```

::: tip Searching beyond names
`projects search` matches project names. To search *inside* indexed documentation, use the [search endpoint](/docs/api/endpoints#get-search) or ask your AI agent to use the `searchDocumentation` MCP tool.
:::

## `portfolio projects get <project-id>`

Print detailed information about one project: metadata, git intelligence, languages, frameworks, capabilities, and health signals.

```bash
portfolio projects get <project-id>
```

Project IDs are shown by `projects list`.

## `portfolio projects discover`

Run discovery over all configured roots (the same operation as the `discover` shortcut):

```bash
portfolio projects discover
```

## The `discover` shortcut

For convenience, discovery is also exposed at the top level:

```bash
portfolio discover
```

## Related

- [HTTP API: projects endpoints](/docs/api/endpoints#projects) — the same data over HTTP
- [Discovery concepts](/docs/concepts/discovery) — how projects are found
