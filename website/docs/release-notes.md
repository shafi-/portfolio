---
title: Release Notes
description: What's new in each Portfolio release.
---

# Release Notes

Portfolio follows a `vX.Y.Z` [semver](https://semver.org)-style release line, cut through the automated [release process](/docs/digging-deeper/release-process). Current version: **v0.3.4**.

## v0.3.4

The current release. Building on the v0.3 line:

- Dashboard milestones completed: project list, project detail, overview, statistics, and relationship explorer pages
- Structured error response format across the HTTP API
- Security hardening: input sanitization, localhost-only CORS, database race-condition fixes, health-check reliability
- Test-suite and documentation-row improvements
- Release-notes template processing in the GitHub release workflow

## v0.3.0 – v0.3.2

- **Knowledge platform expansion**: technologies registry, project relationships, features, and analyses as first-class entities
- MCP tool catalog expansion with cross-project intelligence, historical (git) tools, and code-quality tools
- Analysis workflow with prepared analyzer prompts and change tracking

## v0.2.0 – v0.2.2

- Initial public releases of the Portfolio Engine
- Project discovery across multi-root configurations with smart ignores
- Metadata extraction: git intelligence, 13+ language detection, framework/dependency detection, capability and health signals
- Documentation indexing with SQLite FTS5 full-text search
- MCP stdio server with core discovery, search, configuration, analysis, and relationship tools
- HTTP API and embedded read-only dashboard
- Install script, Homebrew tap, and GoReleaser-based cross-platform releases (darwin/linux · amd64/arm64, CGO disabled)

## Upgrading

Binary releases are self-contained; upgrade by installing the [latest release](https://github.com/shafi-/portfolio/releases) over your current binary. Database migrations are embedded in the binary and applied automatically on first connect after an upgrade — no manual steps.

If you registered an agent integration, refresh it after upgrading:

```bash
portfolio upgrade claude   # or: portfolio upgrade opencode
```
