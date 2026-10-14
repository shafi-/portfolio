---
title: Contribution Guide
description: How to set up a dev environment and contribute to Portfolio.
---

# Contribution Guide

Thanks for considering a contribution! Portfolio is developed in Go with a React dashboard, and this guide gets you productive quickly.

## Development setup

Requirements: **Go 1.21+** and **Node.js 18+** (only for dashboard work).

```bash
git clone https://github.com/shafi-/portfolio.git
cd portfolio

# build and run tests
go build ./cmd/portfolio
go test -race ./...

# static checks
go vet ./...
gofmt -l .
```

### Dashboard development

The dashboard is a React 19 + Vite app whose production build is embedded into the Go binary:

```bash
npm --prefix dashboard install
npm --prefix dashboard run dev      # Vite dev server with hot reload
portfolio dashboard --dev           # Go server in dev mode
```

See the [Dashboard docs](/docs/dashboard#under-the-hood) for the two-terminal workflow.

## Project layout

| Path | What lives there |
| --- | --- |
| `cmd/portfolio/` | Binary entry point |
| `internal/cli/` | Cobra commands |
| `internal/api/` | HTTP handlers |
| `internal/mcp/` | MCP tools |
| `internal/dashboard/` | Embedded asset serving |
| `internal/discovery/`, `internal/metadata/`, `internal/indexer/` | Engine internals |
| `internal/store/`, `internal/database/` | Persistence |
| `pkg/models/` | Shared domain types |
| `dashboard/` | React dashboard source |
| `docs/` | Design documents, ADRs, and specifications |

Longer-form design material lives in `docs/` — in particular `Architecture.md`, `ADR.md`, and `KnowledgeModel.md` are worth reading before large changes.

## Submitting changes

1. **Branch naming** follows the repo convention: `feature/<topic>`, `fix/<topic>`, or `epic/<topic>` for larger efforts.
2. **Tests**: add or update tests for behavior changes. The full suite must pass:

   ```bash
   go test -race ./...
   ```

3. **Formatting**: `gofmt` clean; `go vet` clean.
4. **Docs**: user-facing changes (new commands, flags, API endpoints, MCP tools) should update this documentation site's pages in `website/docs/`.
5. Open a pull request against `main` with a clear description of the behavior change.

CI runs tests, lint/format checks, the dashboard build, and security scanning on every PR.

## Reporting issues

Open a [GitHub issue](https://github.com/shafi-/portfolio/issues) with:

- Portfolio version (`portfolio --version`)
- OS and architecture
- Steps to reproduce, and relevant log output (`portfolio.log`)

## Security issues

Please do **not** open public issues for security vulnerabilities. See the security documentation in `docs/` and contact the maintainers through a private channel.

## License

By contributing, you agree that your contributions are licensed under the [GNU AGPL-3.0](https://github.com/shafi-/portfolio/blob/main/LICENSE), the project's license.
