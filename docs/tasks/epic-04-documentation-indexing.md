# Epic 4 — Documentation Indexing

**Milestone:** 1 — Core Engine
**Status:** todo

## Overview

Index README files, docs/ directories, ADRs, CHANGELOG files, and implement full-text search capabilities.

---

## Story 4.1: Index README

**Status:** todo
**Size:** M
**Blocked by:** 2.2, 1.5

**User Story:**
As the Portfolio Engine, I want to index README files so that users can search project documentation.

**Acceptance Criteria:**
- Finds README.md, README.rst, README.txt, README (case-insensitive in stem and extension)
- Stores in documents table: project_id, path, kind="README", content, content_hash, indexed_at
- Handles missing README (not an error)
- Handles very large READMEs (>1MB)

---

## Story 4.2: Index docs/ Directory

**Status:** todo
**Size:** M
**Blocked by:** 4.1

**User Story:**
As the Portfolio Engine, I want to index docs/ directories so that users have access to detailed documentation.

**Acceptance Criteria:**
- Recursively indexes files in docs/ directory
- Stores kind="DOC" for documentation files
- Supports: .md, .rst, .txt, .adoc formats
- Skips binary files
- Handles projects without docs/ directory
- Respects .gitignore — ignored files excluded from indexing

---

## Story 4.3: Index ADRs

**Status:** todo
**Size:** S
**Blocked by:** 4.2

**User Story:**
As the Portfolio Engine, I want to index Architecture Decision Records so that users can understand project decisions.

**Acceptance Criteria:**
- Finds ADRs in: docs/adr/, .adr/, adr/
- Recognizes common ADR naming patterns (NNN-*.md, *.md)
- Stores kind="ADR"
- Handles projects without ADRs

---

## Story 4.4: Index CHANGELOG

**Status:** todo
**Size:** S
**Blocked by:** 4.1

**User Story:**
As the Portfolio Engine, I want to index CHANGELOG files so that users can see project history.

**Acceptance Criteria:**
- Finds: CHANGELOG.md, CHANGES.md, HISTORY.md (case-insensitive)
- Stores kind="CHANGELOG"
- Handles missing CHANGELOG

---

## Story 4.5: Full-Text Search Indexing

**Status:** todo
**Size:** L
**Blocked by:** 4.1, 4.2, 4.3, 4.4

**User Story:**
As the Portfolio Engine, I want full-text search so that users can find projects by documentation content.

**Acceptance Criteria:**
- FTS index on document.content using SQLite FTS5
- Search supports: phrase queries, Boolean operators
- Returns ranked results with project context
- Fast search across all indexed documents

**Technical Context:**
- SQLite FTS5 for full-text search
- Enables searchDocumentation(query) MCP tool

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 4.1 Index README | todo | M | 2.2, 1.5 |
| 4.2 Index docs/ Directory | todo | M | 4.1 |
| 4.3 Index ADRs | todo | S | 4.2 |
| 4.4 Index CHANGELOG | todo | S | 4.1 |
| 4.5 Full-Text Search Indexing | todo | L | 4.1, 4.2, 4.3, 4.4 |

**Total Size:** 1L + 2M + 2S = ~15 days

**Can Start:** Story 4.1 (after 2.2, 1.5 complete)
