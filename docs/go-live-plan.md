# Go-Live Plan — Portfolio v0.2.0

**Status:** Draft  
**Target version:** v0.2.0  
**Release repo:** `shafi-/project-dash`  
**Last updated:** 2026-07-26

---

## Phase 0: Pre-flight Cleanup

Tasks before any release infrastructure — fixes that affect source quality and user trust.

| # | Task | File | Detail | Time |
|---|------|------|--------|------|
| 0.1 | Fix README clone URL | `README.md:16` | `nerddevsltd/portfolio` → `shafi-/project-dash` | 2m |
| 0.2 | Bump binary version | `cmd/portfolio/main.go:37` | `0.1.0` → `0.2.0` | 2m |
| 0.3 | Fix manual version label | `USER_MANUAL.md:3` | `Version: 1.0` → `Version: 0.2.0` | 1m |
| 0.4 | Fix manual version history | `USER_MANUAL.md:834` | `v1.0` → `v0.2.0` | 1m |
| 0.5 | Add `/portfolio` to `.gitignore` | `.gitignore` | Prevent future binary tracking | 1m |
| 0.6 | Audit MCP tools in USER_MANUAL | `USER_MANUAL.md:382-404` | Add 15 missing tools, remove phantom `deleteRelationship` | 30m |
| 0.7 | Add non-Claude agent integration docs | `USER_MANUAL.md:340` | Document `portfolio install opencode`, generic MCP usage | 15m |
| 0.8 | Explain skill system | `USER_MANUAL.md:359-363` | Expand item 2 to describe skill template + purpose | 10m |
| 0.9 | Full test/lint pass | — | `go test -race ./... && go vet ./... && gofmt -l .` | 5m |

**Total Phase 0: ~1h**

---

## Phase 1: Release Infrastructure

Automated builds, releases, and distribution.

| # | Task | Detail | Time |
|---|------|--------|------|
| 1.1 | Install goreleaser | `brew install goreleaser` | 2m |
| 1.2 | Create `.goreleaser.yaml` | Build matrix: `darwin/amd64`, `darwin/arm64`, `linux/amd64`, `linux/arm64`. Single binary `portfolio`. No extra files. | 20m |
| 1.3 | Create release workflow | `.github/workflows/release.yml` — trigger on `v*` tag push. Steps: checkout, setup Go, test, vet, goreleaser release. Upload binaries. | 30m |
| 1.4 | Push tag `v0.2.0` | `git tag v0.2.0 && git push origin v0.2.0` | 1m |
| 1.5 | Verify release assets | `curl -L https://github.com/shafi-/project-dash/releases/latest/download/portfolio-darwin-arm64 -o portfolio` — binary downloads and works | 5m |
| 1.6 | Verify `portfolio --version` | Reports `0.2.0` | 1m |

**Total Phase 1: ~1h**

---

## Phase 2: USER_MANUAL MCP Tools Overhaul

Rewrite the `### Available MCP Tools` section (lines 382-404) to cover all 26 tools in accurate groups.

### Current state

Only 12 of 26 tools documented. 15 missing entirely. 1 phantom (`deleteRelationship`).

### Required groups

```
Discovery (4)     — health, discoverProjects, listProjects, getProject
Search (2)        — searchProjects, searchDocumentation
Config (2)        — getConfiguration, updateConfiguration                    ← NEW
Code Content (5)  — listProjectFiles, getFileContent, searchFiles,          ← NEW
                     getProjectStructure, getDependencies
Analysis (3)      — getAnalysis, storeAnalysis, listProjectsNeedingAnalysis
Feature (3)       — storeFeature, listFeatures, searchFeatures              ← NEW
Technology (5)    — storeTechnology, tagProjectWithTechnology,              ← NEW
                     listTechnologies, listProjectTechnologies,
                     searchByTechnology
Relationship (2)  — listRelationships, storeRelationship                    ← fixed (drop deleteRelationship)
```

### Also update

- Version history date/version (line 834)
- Claude Code section → rename to "AI Agent Integration" with per-agent subsections
- Add CLI commands for opencode integration if available

**Time:** ~30m

---

## Phase 3: Homebrew Tap

Optional but documented in manual as Method 3.

| # | Task | Detail | Time |
|---|------|--------|------|
| 3.1 | Create `shafi-/homebrew-project-dash` | New public GitHub repo | 2m |
| 3.2 | Create Formula | `Formula/portfolio.rb` — points to GitHub release tarball, SHA256 checksum | 15m |
| 3.3 | Test install | `brew tap shafi-/project-dash && brew install portfolio` | 5m |
| 3.4 | Verify version | `portfolio --version` → `0.2.0` | 1m |

**Total Phase 3: ~25m**

---

## Phase 4: Post-Launch Hardening

Deferrable improvements for the first week after release.

| # | Task | Detail | Priority | Time |
|---|------|--------|----------|------|
| 4.1 | Boost test coverage | Focus on weak packages: `logging` (16%), `fs` (23%), `cli` (8%), `dashboard` (40%) | Medium | 3h |
| 4.2 | Clean stale AI agent config dirs | `.agents/`, `.claude/`, `.continue/`, `.forge/`, `.kilocode/`, `.windsurf/` — root clutter for source browsers | Low | 10m |
| 4.3 | Remove stale test binary | `git rm cli.test` — already gitignored, safe to delete | Low | 1m |
| 4.4 | Add `version` subcommand | `portfolio version` currently fails. `cli/app.go` needs a version command. | Low | 15m |

**Total Phase 4: ~3.5h**

---

## Phase 5: Public Presence

| # | Task | Detail | Time |
|---|------|--------|------|
| 5.1 | Create CONTRIBUTING.md | Bug reports, PR process, dev setup, Code of Conduct link | 20m |
| 5.2 | Add GitHub issue templates | `.github/ISSUE_TEMPLATE/bug.yml`, `feature.yml` | 15m |
| 5.3 | Add security policy | `.github/SECURITY.md` — how to report vulnerabilities privately | 10m |
| 5.4 | Verify all external links | README, USER_MANUAL, docs — no broken URLs | 10m |

**Total Phase 5: ~55m**

---

## Effort Summary

| Phase | Hours | Deferrable? |
|-------|-------|-------------|
| 0: Pre-flight | 1h | No |
| 1: Release infra | 1h | No |
| 2: MCP docs fix | 0.5h | No |
| 3: Homebrew | 0.5h | Yes (soft launch without) |
| 4: Hardening | 3.5h | Yes (week after) |
| 5: Public presence | 1h | Yes (week after) |
| **Total** | **~7.5h** | **Core: 2.5h** |

### Minimum viable launch (2.5h)

Phases 0 + 1 + 2 — fixes all ship-blocking docs issues, creates release pipeline, ships v0.2.0 tag.

---

## Files Changed

| File | Phase | Change |
|------|-------|--------|
| `README.md` | 0.1 | Fix org URL |
| `cmd/portfolio/main.go` | 0.2 | Bump version string |
| `USER_MANUAL.md` | 0.3, 0.4, 2 | Version label, version history, MCP tools list, agent integration |
| `.gitignore` | 0.5 | Add `/portfolio` |
| `.goreleaser.yaml` | 1.2 | **New** — release build config |
| `.github/workflows/release.yml` | 1.3 | **New** — tag-triggered release workflow |
| `Formula/portfolio.rb` | 3.2 | **New** — Homebrew formula (in separate repo) |
| `CONTRIBUTING.md` | 5.1 | **New** |
| `.github/ISSUE_TEMPLATE/*.yml` | 5.2 | **New** |
| `.github/SECURITY.md` | 5.3 | **New** |

---

## Rollback Plan

| Scenario | Action |
|----------|--------|
| Release assets broken | Delete GitHub release, fix, push new tag (v0.2.1) |
| Test regression after v0.2.0 tag | Revert tagged commit, force-push corrected tag (requires `GIT_FORCE_PUSH` — use with care) |
| Homebrew formula wrong checksum | Update formula, push to tap repo |
| Manual error discovered post-release | Fix in next PR, update release notes |
