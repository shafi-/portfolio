---
title: Configuration
description: Portfolio's TOML configuration file — every section and option explained.
---

# Configuration

Portfolio is configured with a single TOML file. There are no environment variables to set and no `.env` files to manage.

## File location

The config file lives in your platform's application configuration directory:

| Platform | Path |
| --- | --- |
| macOS | `~/Library/Application Support/com.portfolio.cli/config.toml` |
| Linux | `~/.config/com.portfolio.cli/config.toml` |

Use a custom file with the global `--config` flag:

```bash
portfolio --config /path/to/config.toml status
```

## Full reference

Here is a complete config file with every supported option and its default value:

```toml
[general]
# Path to the SQLite knowledge base. Defaults to an obscured file named
# ".portfoliodata" inside your platform data directory.
database_path = "/Users/you/Library/Application Support/com.portfolio.cli/.portfoliodata"
# Optional passphrase for the database. If omitted, a key file (db_key)
# is used. Prefer leaving this unset and letting the key file handle it —
# see the security docs.
# database_key = "..."

[discovery]
# Directories that are recursively scanned for Git repositories.
project_roots = []

# Directory names that are never entered during scanning.
ignored_paths = ["node_modules", ".git", "vendor", "build", "dist", "target", "bin"]

[logging]
# DEBUG | INFO | WARN | ERROR
level = "INFO"
# Log file location; empty string disables file logging.
file = "/Users/you/Library/Application Support/com.portfolio.cli/portfolio.log"

[dashboard]
host = "localhost"
port = 8090
asset_path = ""
# Origins allowed to call the API from a browser.
allowed_origins = ["http://localhost:5173", "http://localhost:3000"]
```

::: warning About `database_key`
Setting `database_key` stores the database passphrase in plain text inside the config file. It exists for headless setups; for interactive machines, prefer the automatic key file. Read the trade-offs in [Database & Security](/docs/digging-deeper/security).
:::

## Managing roots from the CLI

You rarely need to edit the file by hand. Manage scan roots with:

```bash
portfolio config set-root ~/Projects       # add a scan root
portfolio config remove-root ~/Projects    # remove a scan root
portfolio config list-roots                # show configured roots
```

## Logging behavior

- Without the `--verbose` flag, the CLI runs quiet (ERROR level) so command output stays clean.
- With `--verbose`, the configured `[logging] level` applies and diagnostics are written to the log file, not stdout.
- In MCP mode, logging is forced to stderr so the JSON-RPC stdio channel stays clean.

## Changing ports at runtime

The `[dashboard]` config section applies to configuration defaults, but the commands accept explicit flags that win at runtime:

```bash
portfolio serve -p 9000          # HTTP API on port 9000 (default 8080)
portfolio dashboard -p 4000      # Dashboard on port 4000 (default 3000)
```
