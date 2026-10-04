---
title: Available Commands
description: Complete reference of every Portfolio CLI command.
---

# Available Commands

The CLI is the administrative interface to the Portfolio Engine. (The intended daily driver is your [AI agent via MCP](/docs/mcp/); the CLI is for setup and direct querying.)

Global flags available on every command:

| Flag | Description |
| --- | --- |
| `--config <path>` | Use a specific config file instead of the default location |
| `--verbose`, `-v` | Show diagnostic logging (honors the configured `[logging]` level) |

## Command list

| Command | Description |
| --- | --- |
| [`portfolio init`](/docs/cli/scanning#portfolio-init) | Interactive setup wizard |
| [`portfolio discover`](/docs/cli/scanning#portfolio-discover) | Discover projects in configured roots |
| [`portfolio scan`](/docs/cli/scanning#portfolio-scan) | Extract metadata and index documentation |
| [`portfolio status`](/docs/cli/scanning#portfolio-status) | Show engine system status |
| [`portfolio projects`](/docs/cli/projects) | List, search, and inspect projects |
| [`portfolio workspace`](/docs/cli/workspaces) | Manage workspaces — named groups of projects |
| [`portfolio config`](/docs/cli/configuration) | Manage configuration and scan roots |
| [`portfolio install`](/docs/cli/ai-integrations) | Install an AI agent integration |
| [`portfolio uninstall`](/docs/cli/ai-integrations) | Remove an AI agent integration |
| [`portfolio upgrade`](/docs/cli/ai-integrations) | Upgrade an AI agent integration |
| [`portfolio doctor`](/docs/cli/ai-integrations#doctor) | Run diagnostics and health checks |
| [`portfolio serve`](/docs/cli/servers#portfolio-serve) | Start the HTTP API server |
| [`portfolio dashboard`](/docs/cli/servers#portfolio-dashboard) | Start the dashboard web server |
| [`portfolio mcp`](/docs/cli/servers#portfolio-mcp) | Start the MCP stdio server |
| [`portfolio manual`](/docs/cli/servers#portfolio-manual) | Print the agent-integration manual |

## Getting help

Every command supports `-h` / `--help` with full usage, flags, and examples:

```bash
portfolio --help
portfolio scan --help
```

Check what version you're running with:

```bash
portfolio --version
```
