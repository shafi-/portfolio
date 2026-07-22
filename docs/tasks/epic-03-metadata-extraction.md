# Epic 3 — Metadata Extraction

**Milestone:** 1 — Core Engine
**Status:** todo

## Overview

Extract Git metadata, detect languages and frameworks, analyze dependencies, compute project statistics, and generate documentation hashes.

---

## Story 3.1: Extract Git Metadata

**Status:** todo
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

---

## Story 3.2: Detect Languages

**Status:** todo
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

**Status:** todo
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

**Status:** todo
**Size:** M
**Blocked by:** 3.3

**User Story:**
As the Portfolio Engine, I want to extract dependency names so that users can see project relationships.

**Acceptance Criteria:**
- Parses package manager files (package.json, go.mod, requirements.txt, Cargo.toml)
- Produces dependency_summary (top 10 direct dependencies)
- Handles different package managers per project type
- Stores raw dependency list for future relationship analysis

---

## Story 3.5: Compute Project Statistics

**Status:** todo
**Size:** S
**Blocked by:** 3.2

**User Story:**
As the Portfolio Engine, I want to compute basic statistics so that users can assess project size.

**Acceptance Criteria:**
- Counts: total files, code files, documentation files
- Estimates lines of code (excluding vendor, node_modules)
- Stores in metadata table
- Fast enough for large repositories (>100k files)

---

## Story 3.6: Compute Documentation Hashes

**Status:** todo
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
| 3.1 Extract Git Metadata | todo | L | 2.2, 1.5 |
| 3.2 Detect Languages | todo | M | 3.1 |
| 3.3 Detect Frameworks | todo | M | 3.2 |
| 3.4 Detect Dependencies | todo | M | 3.3 |
| 3.5 Compute Project Statistics | todo | S | 3.2 |
| 3.6 Compute Documentation Hashes | todo | M | Epic 4 |

**Total Size:** 1L + 4M + 1S = ~18 days

**Can Start:** Story 3.1 (after 2.2, 1.5 complete)
