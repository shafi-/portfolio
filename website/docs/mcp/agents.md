---
title: Connecting AI Agents
description: Register Portfolio's MCP server with Claude Code and OpenCode in one command.
---

# Connecting AI Agents

Portfolio's MCP server works with any MCP-compatible client. First-class, one-command support exists for **Claude Code** and **OpenCode**.

## Claude Code

```bash
portfolio install claude
```

This registers `portfolio mcp` as an MCP server in Claude Code's configuration. After restarting your Claude Code session, Portfolio's [42 tools](/docs/mcp/tools) are available — verify with:

```bash
portfolio doctor claude
```

Useful variations:

```bash
portfolio install claude --force   # re-register over an existing installation
portfolio upgrade claude           # refresh registration after a Portfolio upgrade
portfolio uninstall claude         # remove the registration
```

## OpenCode

```bash
portfolio install opencode
portfolio doctor opencode          # verify
```

The same `--force` / `upgrade` / `uninstall` lifecycle applies.

## What installation actually does

- Writes an MCP server entry pointing at **the installed Portfolio binary** (resolved to its absolute path) invoked with the `mcp` argument
- Records integration state under the data directory's `integrations/` folder so `doctor` can health-check it
- Touches nothing else — your agent config gets one entry, your [knowledge store](/docs/concepts/knowledge-store) is untouched

## Other MCP clients

Any client that supports MCP stdio servers can use Portfolio by adding an equivalent registration manually — a stdio server whose command is the `portfolio` binary with argument `mcp`. After registering, `portfolio doctor` and `portfolio status` still help diagnose engine-side issues.

## Verifying it works

```bash
# 1. Engine diagnostics
portfolio doctor

# 2. Integration-specific diagnostics
portfolio doctor claude

# 3. Then, inside your agent, try:
#    "List my projects and tell me which ones use Postgres."
```

If the agent sees no tools: check that the binary path is on the machine you're running the agent on, re-run `portfolio doctor <target>`, and fall back to `portfolio install <target> --force`. See [Troubleshooting](/docs/digging-deeper/troubleshooting).

## The agent-integration manual

Portfolio ships a manual aimed at agents themselves — usage patterns for the tool catalog, not just signatures:

```bash
portfolio manual          # print it
portfolio manual --write  # write it to disk for reference
```

Many users reference this manual from their agent's memory/instructions file (e.g. `CLAUDE.md`) so the agent uses the tools idiomatically.
