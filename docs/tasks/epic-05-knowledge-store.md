# Epic 5 — Knowledge Store

**Milestone:** 1 — Core Engine
**Status:** todo

## Overview

Implement repository layer, migrations system, and search indexes for the SQLite knowledge store.

---

## Story 5.1: Repository Layer

**Status:** todo
**Size:** M
**Blocked by:** 1.5

**User Story:**
As the Portfolio Engine, I want a repository abstraction so that database operations are maintainable.

**Acceptance Criteria:**
- Repository interface for: projects, metadata, documents, analyses, technologies, relationships
- SQL generation or ORM appropriate for Go
- Transaction support for multi-table operations
- Proper connection pooling

---

## Story 5.2: Migrations System

**Status:** todo
**Size:** M
**Blocked by:** 1.5, 5.1

**User Story:**
As the Portfolio Engine, I want schema migrations so that database can evolve.

**Acceptance Criteria:**
- Migration table tracking applied versions
- Up and down migrations
- Migration files in structured location
- Automatic migration on startup

---

## Story 5.3: Search Indexes

**Status:** todo
**Size:** M
**Blocked by:** 4.5

**User Story:**
As the Portfolio Engine, I want optimized search indexes so that queries are fast.

**Acceptance Criteria:**
- Indexes on: project.name, metadata.language_summary, metadata.framework_summary
- FTS index on document.content
- Query performance: <100ms for typical searches
- Index maintenance on document updates

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 5.1 Repository Layer | todo | M | 1.5 |
| 5.2 Migrations System | todo | M | 1.5, 5.1 |
| 5.3 Search Indexes | todo | M | 4.5 |

**Total Size:** 3M = ~9-15 days

**Can Start:** Story 5.1 (after 1.5 complete)
