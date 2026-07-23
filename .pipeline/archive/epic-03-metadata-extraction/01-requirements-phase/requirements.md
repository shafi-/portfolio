# Epic 3 — Metadata Extraction: Requirements

**Milestone:** 1 — Core Engine
**Status:** Draft
**Author:** DevFlow Phase 1 — Requirement Collector
**Date:** 2026-07-22

---

## Feature Overview

Epic 3 enables the Portfolio Engine to extract deterministic metadata from discovered Git repositories. After Epic 2 inserts project rows via `discoverProjects()`, Epic 3 fills in the `metadata` table with facts about each project's Git state, programming languages, frameworks, dependencies, and documentation content hashes.

All extraction is deterministic: the same repository at the same Git HEAD always produces the same metadata. No heuristic or AI reasoning is involved.

---

## Requirements

### Functional Requirements

#### F-3.1: Git Metadata Extraction
| ID | Requirement |
|----|-------------|
| F-3.1.1 | Extract `default_branch` from the repository's remote origin HEAD reference (e.g., `refs/remotes/origin/HEAD`) or local Git config. |
| F-3.1.2 | Extract `git_head` — the full SHA of the current HEAD commit. |
| F-3.1.3 | Extract `last_commit_at` — the author timestamp of the most recent commit. |
| F-3.1.4 | Extract `last_modified_at` — the most recent file modification timestamp in the working tree (different from last commit, as it may include uncommitted changes). |
| F-3.1.5 | Extract `commit_count` — total number of commits reachable from HEAD. |
| F-3.1.6 | Store extracted values in the `metadata` table under the project's `project_id`. |
| F-3.1.7 | Set `last_scan_at` to the current timestamp after successful extraction. |

#### F-3.2: Language Detection
| ID | Requirement |
|----|-------------|
| F-3.2.1 | Walk project files (excluding vendor/, node_modules/, .git/, build/, dist/, generated code directories) and map file extensions to language names. |
| F-3.2.2 | Produce a `language_summary` string of comma-separated unique language names sorted by prevalence (most used first), e.g., "Go, TypeScript, Shell". |
| F-3.2.3 | Support a configurable extension-to-language mapping that can be extended via configuration file (not hardcoded to a single list). |
| F-3.2.4 | For a project with zero analyzable files, set `language_summary` to NULL. |

#### F-3.3: Framework Detection
| ID | Requirement |
|----|-------------|
| F-3.3.1 | Scan dependency/configuration files for known framework markers (e.g., `react` in `package.json`, `django` in `requirements.txt`, `gin` in `go.mod`, `spring-boot` in `pom.xml`). |
| F-3.3.2 | Produce a `framework_summary` string of comma-separated framework names. |
| F-3.3.3 | Support multiple frameworks per project (e.g., a project using both React and Express). |
| F-3.3.4 | Framework detection rules MUST be defined in an extensible mapping, not hardcoded. New frameworks should be addable via configuration. |

#### F-3.4: Dependency Detection
| ID | Requirement |
|----|-------------|
| F-3.4.1 | Parse project dependency manifest files per project language/ecosystem: `package.json` (npm/yarn/pnpm), `go.mod` (Go), `requirements.txt`/`pyproject.toml` (Python), `Cargo.toml` (Rust), `Gemfile` (Ruby), `pom.xml`/`build.gradle` (Java). |
| F-3.4.2 | Produce `dependency_summary` as a comma-separated string of the top 10 direct dependency names (not version numbers). |
| F-3.4.3 | For projects with no recognized manifest files, set `dependency_summary` to NULL. |
| F-3.4.4 | Do NOT include transitive dependencies in the summary — only direct dependencies declared by the project. |
| F-3.4.5 | Store all direct dependency names (untruncated) in a separate `dependencies` table or JSON column for downstream relationship analysis (Epic 13). |



#### F-3.6: Documentation Hashes
| ID | Requirement |
|----|-------------|
| F-3.6.1 | Compute SHA-256 hash of each documentation file's content discovered by Epic 4. |
| F-3.6.2 | Produce a single `documentation_hash` for the project that covers all documentation files (e.g., a combined hash or hash of hashes). |
| F-3.6.3 | Store `documentation_hash` in the `metadata` table. |
| F-3.6.4 | The hash calculation MUST be deterministic: same file contents always produce the same hash. |

### Non-Functional Requirements

| ID | Requirement | Target |
|----|-------------|--------|
| NFR-1 | **Determinism** — Same input (repository at same HEAD) must always produce identical metadata output. | 100% repeatability |
| NFR-2 | **Performance** — Metadata extraction for a single project must complete within reasonable time. | <5s for repos <10k files; <30s for repos up to 100k files |
| NFR-3 | **Isolation** — Extraction for one project must not affect or be affected by extraction for another project. | No cross-project coupling |
| NFR-4 | **Idempotency** — Running extraction multiple times on the same project must produce identical metadata values for all fields except `last_scan_at`, which MUST always be updated to the current timestamp. | Full idempotency |
| NFR-5 | **Disk respect** — Extraction must avoid reading unnecessary files. | Skip ignored dirs at walk level, not post-filter |
| NFR-6 | **Graceful degradation** — Partial failure in one extraction story must not block other stories from succeeding. | Each story writes independently |

---

## Edge Cases & Error Handling

| # | Edge Case | Expected Behavior |
|---|-----------|-------------------|
| EC-1 | **Empty repository** (no commits) | `git_head` = NULL, `commit_count` = 0, `last_commit_at` = NULL, `last_modified_at` = current time or NULL. Other metadata stories still run. |
| EC-2 | **Bare repository** | Extract available metadata only (no working tree). `last_modified_at` = NULL. Language/framework extraction is skipped. |
| EC-3 | **Detached HEAD** | `git_head` still records the current commit SHA. `default_branch` remains valid (from remote HEAD ref). All other metadata proceeds normally. |
| EC-4 | **Repository with no recognized language files** | `language_summary` = NULL. |
| EC-5 | **Repository with unknown/obscure languages** | Files with unrecognized extensions are not included in `language_summary`. |
| EC-6 | **Polyglot project** (20+ languages) | `language_summary` lists all detected languages sorted by prevalence. No artificial truncation. |
| EC-7 | **Multi-framework project** | `framework_summary` lists all detected frameworks. No artificial limit. |
| EC-8 | **Monorepo with multiple manifest files** | Parse each manifest independently. `dependency_summary` aggregates top 10 across all manifests. |
| EC-9 | **Corrupted manifest file** (invalid JSON/YAML/TOML) | Gracefully skip the corrupted file, log a warning, continue with other manifests. |
| EC-10 | **Repository with >100k files** | Use file tree walking with early ignore-directory skipping. LOC estimation may use sampling. |
| EC-11 | **Repository with no documentation files** | `documentation_hash` = NULL. Story 3.6 produces no hash but does not error. |
| EC-12 | **Repository deleted or moved between discovery and extraction** | Skip the project, log a warning, and continue. The orphaned project row remains in the `projects` table with its existing metadata; cleanup of orphaned rows is delegated to a future Epic or garbage-collection story. |
| EC-13 | **Symlinks pointing outside repository** | Do not follow symlinks outside the repository root. |
| EC-14 | **Non-UTF-8 file content in documentation** | Hash raw bytes, not decoded content. SHA-256 operates on byte sequences. |
| EC-15 | **Repository with no remote** (local-only) | `default_branch` is determined from local `HEAD` ref (e.g., `refs/heads/main`) rather than remote HEAD. |
| EC-16 | **Permissions errors on specific files or directories** | Skip inaccessible files, log warning, continue. Do not abort. |

---

## Acceptance Criteria

Derived from the Epic 3 stories. Each AC maps to one or more stories.

### Story 3.1 — Extract Git Metadata

| AC ID | Criteria |
|-------|----------|
| AC-3.1.1 | Engine extracts `default_branch`, `git_head`, `last_commit_at`, `last_modified_at`, `commit_count` from a Git repository. |
| AC-3.1.2 | Empty repository (no commits): `git_head` is NULL, `commit_count` is 0, timestamps are NULL. |
| AC-3.1.3 | Bare repository: extraction produces available metadata, `last_modified_at` is NULL. |
| AC-3.1.4 | Detached HEAD state: `git_head` still records the commit SHA; `default_branch` is still resolved. |
| AC-3.1.5 | All extracted values are stored in the `metadata` table per the schema in PlatformSpecification.md. |

### Story 3.2 — Detect Languages

| AC ID | Criteria |
|-------|----------|
| AC-3.2.1 | Engine analyzes file extensions and produces a `language_summary` (e.g., "Go, TypeScript, Shell"). |
| AC-3.2.2 | Polyglot projects: all detected languages are listed, sorted by prevalence. |
| AC-3.2.3 | Extension-to-language mapping is configurable (not hardcoded atomically). |
| AC-3.2.4 | vendor/, node_modules/, .git/, build/, dist/ directories are excluded from analysis. |
| AC-3.2.5 | Project with no code files: `language_summary` is NULL. |

### Story 3.3 — Detect Frameworks

| AC ID | Criteria |
|-------|----------|
| AC-3.3.1 | Engine scans dependency/configuration files for framework markers (React, Vue, Django, Rails, Gin, Spring, etc.). |
| AC-3.3.2 | `framework_summary` is produced as a comma-separated string. |
| AC-3.3.3 | Multiple frameworks per project are supported. |
| AC-3.3.4 | Framework mapping is extensible via configuration. |

### Story 3.4 — Detect Dependencies

| AC ID | Criteria |
|-------|----------|
| AC-3.4.1 | Engine parses `package.json`, `go.mod`, `requirements.txt`, `Cargo.toml` (and additional per-language manifests). |
| AC-3.4.2 | `dependency_summary` contains the top 10 direct dependency names. |
| AC-3.4.3 | Different package managers are handled per project type. |
| AC-3.4.4 | No transitive dependencies are included. |

### Story 3.6 — Compute Documentation Hashes

| AC ID | Criteria |
|-------|----------|
| AC-3.6.1 | Engine computes SHA-256 hashes for all documentation files discovered in Epic 4. |
| AC-3.6.2 | Combined `documentation_hash` is stored in the `metadata` table. |
| AC-3.6.3 | Hash is deterministic: same file contents produce the same hash. |
| AC-3.6.4 | No documentation files: `documentation_hash` is NULL. |

---

## Data Flow

```
┌─────────────────────────────────────────────────────────────┐
│                    Epic 2 — Discovery                        │
│  discoverProjects() → inserts rows into `projects` table    │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Epic 3 — Metadata Extraction              │
│                                                             │
│  ┌─────────────┐                                            │
│  │ 3.1 Git     │──► git_head, default_branch,               │
│  │ Metadata    │    last_commit_at, last_modified_at,        │
│  │             │    commit_count                             │
│  └──────┬──────┘                                            │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────┐                                            │
│  │ 3.2 Langs   │──► language_summary                        │
│  └──────┬──────┘                                            │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────┐                                            │
│  │ 3.3 Frameworks          │──► framework_summary           │
│  └──────┬──────┘                                            │
│         │                                                   │
│         ▼                                                   │
│  ┌─────────────┐                                            │
│  │ 3.4 Deps    │                                            │
│  │             │──► dependency_summary                      │
│  └─────────────┘                                            │
│                                                             │
│  ┌─────────────┐  (blocked by Epic 4)                       │
│  │ 3.6 Doc     │──► documentation_hash                      │
│  │ Hashes      │                                            │
│  └─────────────┘                                            │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │               Write to metadata table                 │   │
│  │  UPDATE metadata SET git_head=?, language_summary=?,  │   │
│  │  ..., last_scan_at=NOW() WHERE project_id = ?         │   │
│  └──────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Downstream Consumers                     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌───────────┐  │
│  │ Epic 5   │  │  MCP     │  │ HTTP API │  │ Dashboard │  │
│  │ Search   │  │  Tools   │  │          │  │           │  │
│  └──────────┘  └──────────┘  └──────────┘  └───────────┘  │
└─────────────────────────────────────────────────────────────┘
```

### Execution Model

Each story runs as an independent capability (per the "Capabilities over Workflows" principle). The caller (CLI command or MCP tool) may orchestrate them in sequence, but the engine does not mandate a pipeline. This enables:

- **Partial re-scanning**: If only languages change, re-run only Story 3.2.
- **Incremental adoption**: Stories 3.6 can be enabled only after Epic 4 is ready.
- **Graceful partial failure**: If 3.3 fails, 3.2's output is still persisted.

---

## Dependencies

### Prerequisites (must be complete before Epic 3 can start)

| Dependency | Epic | Description |
|------------|------|-------------|
| Epic 1 — Database Schema | 1 | `projects` and `metadata` tables must exist in SQLite. |
| Epic 2 — Project Discovery | 2 | `discoverProjects()` must be inserting project rows so Epic 3 has projects to analyze. |

### Internal Story Dependencies

| Story | Depends On | Rationale |
|-------|-----------|-----------|
| 3.1 Git Metadata | — | No dependency — can start first when Epics 1 and 2 are done. |
| 3.2 Detect Languages | 3.1 | Needs project root from discovered project. |
| 3.3 Detect Frameworks | 3.2 | Framework detection reads language-specific dependency files. |
| 3.4 Detect Dependencies | 3.3 | Dependency detection uses the same manifest files as framework detection. |
| 3.6 Doc Hashes | Epic 4 | Must know which files are "documentation files" — Epic 4 defines that. |

### External Dependencies

| Dependency | Notes |
|------------|-------|
| `git` CLI or libgit2 | For Git metadata extraction (Story 3.1). |
| Go standard library `crypto/sha256` | For documentation hashing (Story 3.6). |
| JSON/TOML/YAML parsers | For manifest file parsing (Stories 3.3, 3.4). Language: Go's `encoding/json`, third-party TOML/YAML libs. |
| File system walk | Go's `filepath.Walk` or `fs.WalkDir`. Must support early directory skipping for ignored dirs. |

### Downstream Consumers

| Consumer | Epic | What It Uses |
|----------|------|-------------|
| Search Indexing | Epic 5 | `language_summary`, `framework_summary`, `dependency_summary` for search filtering. |
| Change Detection | (ongoing) | `git_head`, `documentation_hash`, `last_scan_at` to detect when re-analysis is needed. |
| Dashboard | Epic 3+ | All metadata fields for display in Project Detail and Portfolio Overview pages. |
| AI Agent Analysis | (ongoing) | `language_summary`, `framework_summary` inform agent about project stack before deep analysis. |

---

## Principles Applied

From Guideline.md:

1. **Engine Knows, Agent Thinks** — All six stories are purely deterministic. No LLM calls, no heuristic reasoning, no AI.
2. **Deterministic by Default** — Same git HEAD → same metadata. Story 3.6's SHA-256 hash is the purest expression of this.
3. **Store Facts, Compute Indicators** — Store `documentation_hash`, `git_head`, `last_scan_at`; compute `needs_analysis`, `documentation_changed`, `analysis_outdated` at query time.
4. **Capabilities over Workflows** — Each story is a standalone capability (`extractGitMetadata`, `detectLanguages`, etc.). No engine-level metadata pipeline.
5. **Ignore Generated** — vendor/, node_modules/, .git/, build/, dist/ are universally excluded.
6. **Single Knowledge Model** — All outputs map to the `metadata` table defined in PlatformSpecification.md. No synthetic fields.

---

## Open Questions

1. Should `dependency_summary` be stored as a raw JSON array instead of a comma-separated string for easier querying? (The schema says "top 10 names" as string — confirm if structured storage is preferred.)
2. Should Story 3.6 hash each doc file individually and combine via hash-of-hashes, or compute a single rolling hash over all doc file content? The former is more deterministic and debuggable.
4. Should language detection use a built-in mapping or load from a config file path? A built-in default with optional per-user overrides is recommended.
