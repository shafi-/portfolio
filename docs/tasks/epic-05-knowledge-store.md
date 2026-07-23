# Epic 5 — Knowledge Store

**Milestone:** 1 — Core Engine
**Status:** completed

## Overview

Implement repository layer, migrations system, and search indexes for the SQLite knowledge store.

---

## Story 5.1: Repository Layer

**Status:** completed
**Size:** M
**Blocked by:** 1.5

**User Story:**
As the Portfolio Engine, I want a repository abstraction so that database operations are maintainable.

**Acceptance Criteria:**
- Repository interface for: projects, metadata, documents, analyses, features, technologies, relationships, configuration
- SQL generation or ORM appropriate for Go
- Transaction support for multi-table operations
- Proper connection pooling

---

## Story 5.2: Migrations System

**Status:** completed
**Size:** M
**Blocked by:** 1.5, 5.1

**User Story:**
As the Portfolio Engine, I want schema migrations so that database can evolve.

**Acceptance Criteria:**
- Migration table tracking applied versions with checksum of migration content
- Up and down migrations
- Migration files in structured location
- Automatic migration on startup
- Duplicate version detection causes fail-fast startup

---

## Story 5.3: Search Indexes

**Status:** completed
**Size:** M
**Blocked by:** 4.5

**User Story:**
As the Portfolio Engine, I want optimized search indexes so that queries are fast.

**Acceptance Criteria:**
- Indexes on: project.name, metadata.language_summary, metadata.framework_summary, documents.kind
- FTS index on document.content
- Query performance: <100ms for typical searches
- Index maintenance on document updates

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 5.1 Repository Layer | completed | M | 1.5 |
| 5.2 Migrations System | completed | M | 1.5, 5.1 |
| 5.3 Search Indexes | completed | M | 4.5 |

**Total Size:** 3M = ~9-15 days

**Completed Stories:**
- Story 5.1 — Repository Layer: `internal/store/projects.go`, `internal/store/analyses.go`, `internal/store/features.go`, `internal/store/technologies.go`, `internal/store/relationships.go`, `internal/store/configuration.go`; Tx variants added to `internal/store/metadata.go`, `internal/store/dependencies.go`
- Story 5.2 — Migrations System: SHA-256 checksum verification, `MigrateDown()`, file-based migration loading, `ListAppliedMigrations()`
- Story 5.3 — Search Indexes: migration 5 (search indexes on project.name, metadata.language_summary, metadata.framework_summary, documents.kind), query benchmarks (<15μs)
