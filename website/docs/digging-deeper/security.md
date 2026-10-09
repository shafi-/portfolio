---
title: Database & Security
description: How Portfolio protects your data — obscured filenames, key handling, and localhost-only networking.
---

# Database & Security

Portfolio holds a detailed map of your development life — every repo, its git activity, its dependencies. The security model is built around one goal: **that data never leaves your machine and never becomes readable to others on it.**

## Threat model

Portfolio is a *personal, local-first* tool. It assumes:

- The machine you run it on is your own (single-user).
- The network is untrusted — so nothing listens beyond localhost.
- Other local users and whatever reads your home directory are a real risk — hence the obscured database and key handling below.

## Localhost-only networking

- The HTTP API ([`portfolio serve`](/docs/cli/servers#portfolio-serve)) binds to `127.0.0.1` — not reachable from other machines.
- The API's CORS layer only allows **localhost origins** (`localhost`, `127.0.0.1`, `[::1]`); requests with any other `Origin` are not granted cross-origin access.
- The MCP server communicates over stdio — it opens no network port at all.
- The [dashboard](/docs/dashboard) is intended for local browsing.

There is deliberately no authentication layer: nothing remote can connect, so there is nothing to authenticate.

## Obscured database filename

The knowledge store does not live at an obvious path like `portfolio.db`. The default is a dot-file named **`.portfoliodata`** inside the platform data directory (`~/Library/Application Support/com.portfolio.cli/` on macOS). Nothing about the filename reveals its contents or format.

## Key handling

Database content is protected with a key managed by the engine:

- By default, Portfolio generates and stores a **key file (`db_key`)** next to the database, with filesystem permissions locked to your user.
- The `[general] database_key` config option exists for headless setups where a key file is impractical — but storing a passphrase in a plain-text config trades one file for another. Prefer the default.
- Security design notes live in the repository: `docs/DATABASE-PASSWORD-SECURITY.md`, `docs/SECURE-DATABASE-LOCATION.md`, and `docs/OBSCURED-DATABASE-FILENAME.md`.

## What this model does *not* protect against

- **Root/admin users** — anyone with root can read your files, key included.
- **Compromised machines** — Portfolio encrypts/obscures data at rest locally; it is not a defense against malware running as your user.
- **Sync services** — if you sync your data directory to cloud storage, the database (and key) leave the machine. Don't sync it.

## Reporting vulnerabilities

Do not open public issues for security problems. Reach the maintainers privately; see the [Contribution Guide](/docs/contribution-guide#security-issues).
