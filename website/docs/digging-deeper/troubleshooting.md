---
title: Troubleshooting
description: Common Portfolio problems and how to fix them.
---

# Troubleshooting

Quick fixes for the issues that come up most often.

## `command not found: portfolio`

The binary isn't on your `PATH`.

```bash
# find where it was installed
which portfolio || ls /usr/local/bin/portfolio ~/go/bin/portfolio

# add the directory to PATH (example for ~/.zshrc)
export PATH="$PATH:/usr/local/bin"
```

## `portfolio` runs but scans find nothing

1. Check your roots: `portfolio config list-roots`
2. Add the directory that actually contains your repos: `portfolio config set-root ~/Projects`
3. Roots must contain **Git repositories** — plain directories without `.git` are skipped by design
4. Remember the default ignores: `node_modules`, `.git`, `vendor`, `build`, `dist`, `target`, `bin` are never entered

## Database errors on start

- **"failed to connect/open database"** — check that the path in `[general] database_path` exists and is writable; a partially-copied database file can also cause this.
- **Lockups** — SQLite allows one writer; make sure another `portfolio serve`/`dashboard`/`scan` process isn't holding a write transaction, then retry.
- If a scan was interrupted, simply re-run `portfolio scan` — indexing is incremental and idempotent.

## Port already in use

```bash
portfolio serve -p 8081       # API default is 8080
portfolio dashboard -p 3001   # dashboard default is 3000
```

## Agent doesn't show Portfolio's tools

1. Run `portfolio doctor claude` (or `opencode`) — it checks the registration and the MCP handshake
2. Confirm the binary path registered in the agent config exists (move/rename the binary? re-register)
3. Re-register cleanly: `portfolio install claude --force`
4. Restart the agent session — integrations load at startup
5. Verify the engine itself is healthy: `portfolio status` and `portfolio doctor`

## Logs

The engine writes to the configured log files in the data directory:

- `portfolio.log` — application log (`level` configurable via `[logging]`)
- `error.log` — errors

Run any command with `-v` to surface diagnostic logging (it goes to the log file, keeping command output clean).

## Still stuck?

- Re-run [`portfolio init`](/docs/cli/scanning#portfolio-init) — it's safe and repairs configuration/state mismatches.
- Search [existing issues](https://github.com/shafi-/portfolio/issues) or open a new one with `portfolio --version`, your OS, and the relevant log excerpt.
