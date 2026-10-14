---
title: AI Integrations
description: install, uninstall, upgrade, and doctor — wiring Portfolio into Claude Code and OpenCode.
---

# AI Integrations

Portfolio integrates with AI coding agents by registering an [MCP server](/docs/mcp/) entry in the agent's configuration. The `install`, `uninstall`, `upgrade`, and `doctor` commands manage that wiring for supported agents.

Supported targets: **`claude`** (Claude Code) and **`opencode`** (OpenCode).

## `portfolio install <target>`

Install the integration for an agent.

```bash
portfolio install claude
portfolio install opencode
```

| Flag | Description |
| --- | --- |
| `-f`, `--force` | Force reinstall even if already installed |

Installation registers Portfolio's MCP server (the running binary invoked as `portfolio mcp`) in the agent's configuration, so the agent can immediately use [Portfolio's tools](/docs/mcp/tools).

## `portfolio uninstall <target>`

Remove the integration from an agent's configuration.

```bash
portfolio uninstall claude
```

Your knowledge store and configuration are untouched — only the agent registration is removed.

## `portfolio upgrade <target>`

Upgrade an existing integration registration (for example after a Portfolio version bump).

```bash
portfolio upgrade claude
```

## `doctor`

Run diagnostics and health checks.

```bash
portfolio doctor            # general engine diagnostics
portfolio doctor claude     # check Claude Code integration health
portfolio doctor opencode   # check OpenCode integration health
```

`doctor` verifies the engine (config, database) and, with a target, that the agent's registration points at the right binary and the MCP handshake works. If something is off, it tells you — and `install --force` re-registers cleanly.

::: tip Troubleshooting agents
If your agent doesn't see Portfolio's tools, run `portfolio doctor <target>` first — it resolves the vast majority of connection issues. See also [Troubleshooting](/docs/digging-deeper/troubleshooting).
:::
