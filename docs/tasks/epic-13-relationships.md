# Epic 13 — Relationships

**Milestone:** 4 — Portfolio Intelligence
**Status:** todo

## Overview

Implement relationship model, queries, and visualization support to enable exploration of cross-project connections.

---

## Story 13.1: Relationship Model

**Status:** todo
**Size:** M
**Blocked by:** 10.4

**User Story:**
As the Portfolio Engine, I want a relationship model so that AI agents can persist cross-project connections.

**Acceptance Criteria:**
- Database table for relationships (per PlatformSpecification.md)
- Support for directed relationships (source → target)
- Relationship type enum
- Confidence score (0-1)
- Unique constraint to prevent duplicates

---

## Story 13.2: Relationship Queries

**Status:** todo
**Size:** M
**Blocked by:** 13.1

**User Story:**
As an AI agent, I want to query relationships so that I can understand project connections.

**Acceptance Criteria:**
- listRelationships(projectId) MCP tool
- Returns all relationships for a project
- Filters by relationship type
- Includes target project details
- Orders by confidence

---

## Story 13.3: Visualization Support

**Status:** todo
**Size:** M
**Blocked by:** 13.2

**User Story:**
As the dashboard, I want relationship data so that I can visualize connections.

**Acceptance Criteria:**
- HTTP API endpoint for relationships
- Returns format suitable for graph visualization
- Includes node and edge data
- Handles missing relationships gracefully

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 13.1 Relationship Model | todo | M | 10.4 |
| 13.2 Relationship Queries | todo | M | 13.1 |
| 13.3 Visualization Support | todo | M | 13.2 |

**Total Size:** 3M = ~9-15 days

**Can Start:** Story 13.1 (after 10.4 complete)
