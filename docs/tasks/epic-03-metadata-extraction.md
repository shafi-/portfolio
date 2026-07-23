# Epic 3 — Metadata Extraction

**Milestone:** 1 — Core Engine
**Status:** completed

## Overview

Extract Git metadata, detect languages and frameworks, analyze dependencies, compute project statistics, and generate documentation hashes.

---

## Story 3.1: Extract Git Metadata

**Status:** completed
**Size:** L
**Blocked by:** 2.2, 1.5

**User Story:**
As the Portfolio Engine, I want to extract Git repository metadata so that users can see project activity.

**Acceptance Criteria:**
- Extracts: default_branch, HEAD (git_head), last_commit_at, last_modified_at, commit_count
- Handles repositories with no commits (empty repo)
- Handles bare repositories
- Handles detached HEAD state
- Stores in metadata table per PlatformSpecification.md

**Edge Cases:**
- No commits → git_head is NULL, commit_count is 0
- Bare repo → limited metadata available
- Detached HEAD → still record commit SHA
- Local-only repo (no remote) → default_branch from local HEAD ref

---

## Story 3.2: Detect Languages

**Status:** completed
**Size:** M
**Blocked by:** 3.1

**User Story:**
As the Portfolio Engine, I want to detect programming languages so that users can filter projects by technology.

**Acceptance Criteria:**
- Analyzes file extensions in repository
- Produces language_summary (e.g., "Go, TypeScript, Shell")
- Handles polyglot projects (multiple languages)
- Configurable extension to language mapping
- Ignores vendor/, node_modules/, generated files

---

## Story 3.3: Detect Frameworks

**Status:** completed
**Size:** M
**Blocked by:** 3.2

**User Story:**
As the Portfolio Engine, I want to detect frameworks so that users can identify projects by capabilities.

**Acceptance Criteria:**
- Scans dependency files for framework markers
- Examples: React, Vue, Django, Rails, Gin, Spring
- Produces framework_summary field
- Supports multiple frameworks per project
- Framework list extensible

---

## Story 3.4: Detect Dependencies

**Status:** completed
**Size:** M
**Blocked by:** 3.3

**User Story:**
As the Portfolio Engine, I want to extract dependency names so that users can see project relationships.

**Acceptance Criteria:**
- Parses package manager files: package.json, go.mod, requirements.txt/pyproject.toml, Cargo.toml, Gemfile, pom.xml/build.gradle
- Produces dependency_summary (top 10 direct dependencies)
- Handles different package managers per project type
- Stores raw dependency list for future relationship analysis

---

## Story 3.6: Compute Documentation Hashes

**Status:** completed
**Size:** M
**Blocked by:** Epic 4

**User Story:**
As the Portfolio Engine, I want to hash documentation files so that I can detect changes requiring re-indexing.

**Acceptance Criteria:**
- Computes SHA-256 hash of documentation file contents
- Stores in documentation_hash field in metadata table
- Hashes all files discovered in Epic 4
- Used to detect documentation_changed indicator

**Technical Context:**
- Store facts (hashes), compute indicators (documentation_changed)

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 3.1 Extract Git Metadata | completed | L | 2.2, 1.5 |
| 3.2 Detect Languages | completed | M | 3.1 |
| 3.3 Detect Frameworks | completed | M | 3.2 |
| 3.4 Detect Dependencies | completed | M | 3.3 |
| 3.5 Service Assembly | completed | M | 3.4 |
| 3.6 Compute Documentation Hashes | completed | M | Epic 4 |

**Total Size:** 1L + 4M = ~18 days

**Completed:** all 6 stories

**Completed Stories:**
- Story 3.1 — Extract Git Metadata: `internal/metadata/git.go`, `walk.go`, `internal/store/metadata.go`
- Story 3.2 — Detect Languages: `internal/metadata/languages.go`, `config.go`, `languages_data.go`
- Story 3.3 — Detect Frameworks: `internal/metadata/frameworks.go`, `frameworks_data.go`
- Story 3.4 — Detect Dependencies: `internal/metadata/dependencies.go`, `internal/store/dependencies.go`
- Story 3.5 — Service Assembly: `internal/metadata/service.go`
- Story 3.6 — Compute Documentation Hashes: `internal/indexer/runner.go` (hash computation + storage), `internal/indexer/types.go` (DocsChanged indicator)
