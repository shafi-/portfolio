---
title: Servers & Tools
description: serve, dashboard, mcp, and manual — running Portfolio's long-lived surfaces.
---

# Servers & Tools

Portfolio ships several long-running and utility surfaces beyond the plain CLI commands.

## `portfolio serve`

Start the HTTP API server.

```bash
portfolio serve            # listens on 127.0.0.1:8080
portfolio serve -p 9000    # custom port
```

| Flag | Description |
| --- | --- |
| `-p`, `--port <port>` | HTTP server port (default `8080`) |

The server binds to `127.0.0.1` only and exposes the [HTTP API](/docs/api/). Read and write timeouts are 15s, idle timeout 60s. Graceful shutdown on `SIGINT`/`SIGTERM`.

## `portfolio dashboard`

Start the dashboard web server.

```bash
portfolio dashboard            # listens on :3000
portfolio dashboard -p 4000    # custom port
```

| Flag | Description |
| --- | --- |
| `-p`, `--port <port>` | Dashboard server port (default `3000`) |
| `--dev` | Development mode — serve assets from the filesystem for hot reload |
| `--dist <path>` | Dist folder used in dev mode (default `./dashboard/dist`) |

In production mode the dashboard's compiled React assets are served from inside the binary. The dashboard also proxies the HTTP API on the same origin — see the [Dashboard page](/docs/dashboard).

## `portfolio mcp`

Start the MCP stdio server for AI agent integration.

```bash
portfolio mcp
```

This speaks JSON-RPC over stdio per the [Model Context Protocol](https://modelcontextprotocol.io). You normally never run it by hand — `portfolio install claude` registers it with your agent, which launches and manages the process itself. All logging in this mode is forced to stderr to keep the stdio channel clean.

See [MCP Integration](/docs/mcp/) for the full tool catalog.

## `portfolio manual`

Print the agent-integration manual — a self-contained guide for AI agents (and humans) on using Portfolio's MCP tools effectively.

```bash
portfolio manual              # print to stdout
portfolio manual --write      # write to disk instead
portfolio manual --path <p>   # custom output path when --write is set
```

| Flag | Description |
| --- | --- |
| `--write` | Write the manual to disk instead of printing |
| `--path <path>` | Output path when `--write` is set |

Agents that support it benefit from having this manual referenced in their context — it documents tool usage patterns, not just signatures.
