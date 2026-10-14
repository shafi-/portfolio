---
title: Release Process
description: How Portfolio versions are cut — branch flow, scripts, and automation.
---

# Release Process

Releases are cut from dedicated release branches with a single script that handles version bumping, testing, and pull-request creation. Cross-platform binaries are published by GitHub Actions via GoReleaser.

## Cutting a release

```bash
./scripts/make-release.sh vX.Y.Z
```

The script:

1. Validates the version format and that it is greater than the current version
2. Bumps the version in `internal/version/`
3. Runs the full test suite (`go test -race ./...`) — a failing suite aborts the release
4. Creates a `release-vX.Y.Z` branch and opens a pull request
5. On merge, the `release.yml` workflow builds and publishes binaries with GoReleaser

Release notes follow the repository's release template and are processed by the workflow.

## What the release pipeline produces

Per release ([`.goreleaser.yaml`](https://github.com/shafi-/portfolio/blob/main/.goreleaser.yaml)):

- Binaries for **darwin** and **linux** on **amd64** and **arm64**
- Built with `CGO_ENABLED=0` — fully static, pure-Go SQLite included
- Version, commit, and build date injected via ldflags into `internal/version` (visible in `portfolio --version`)

The dashboard build workflow compiles the React dashboard separately; CI keeps both halves verified on every change.

## Versioning conventions

- **Patch** (`v0.3.3 → v0.3.4`): bug fixes, polish, security hardening
- **Minor** (`v0.2.x → v0.3.x`): new capabilities — e.g. the v0.3 knowledge-platform expansion (technologies, relationships, features, analyses)
- **Major**: reserved; not yet used pre-1.0

Historical releases are listed in [Release Notes](/docs/release-notes).
