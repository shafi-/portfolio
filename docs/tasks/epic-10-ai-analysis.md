# Epic 10 — AI Analysis

**Milestone:** 2 — Agent Integration
**Status:** todo

## Overview

Implement analysis schema, analysis persistence, stale analysis detection, and relationship persistence to support AI-generated semantic knowledge.

---

## Story 10.1: Analysis Schema

**Status:** todo
**Size:** M
**Blocked by:** 5.1

**User Story:**
As the Portfolio Engine, I want an analysis schema so that AI agents can provide consistent semantic knowledge.

**Acceptance Criteria:**
- JSON schema for analysis object
- Fields: summary, purpose, architecture, maturity, strengths, weaknesses, reusable_components, notes, analyzed_at, analyzed_git_head, analyzer
- Schema validation on storeAnalysis()
- Stores raw_json for flexibility

**Technical Context:**
- Per KnowledgeModel.md, Analysis entity

---

## Story 10.2: Persist Analyses

**Status:** todo
**Size:** M
**Blocked by:** 10.1, 7.4

**User Story:**
As an AI agent, I want to store analyses so that semantic knowledge persists across sessions.

**Acceptance Criteria:**
- storeAnalysis() creates/updates analysis record
- Stores analyzer (which agent created it)
- Links to project and git_head
- Overwrites previous analysis by same analyzer

---

## Story 10.3: Detect Stale Analyses

**Status:** todo
**Size:** M
**Blocked by:** 3.1, 10.2

**User Story:**
As the Portfolio Engine, I want to detect outdated analyses so that agents know when to re-analyze.

**Acceptance Criteria:**
- Compares analyzed_git_head vs current git_head
- Computes analysis_outdated indicator
- listProjectsNeedingAnalysis() returns stale projects
- Projects without analysis are also returned

**Technical Context:**
- Store facts (git_head), compute indicators (analysis_outdated)

---

## Story 10.4: Relationship Persistence

**Status:** todo
**Size:** M
**Blocked by:** 10.1

**User Story:**
As an AI agent, I want to store relationships so that cross-project connections are preserved.

**Acceptance Criteria:**
- Relationships API in MCP tools
- Stores: source_project, target_project, type, description, confidence
- Relationship types: Similar, Evolution, Shared Feature, Shared Technology, Reuses Component
- Queryable via listRelationships(projectId)

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 10.1 Analysis Schema | todo | M | 5.1 |
| 10.2 Persist Analyses | todo | M | 10.1, 7.4 |
| 10.3 Detect Stale Analyses | todo | M | 3.1, 10.2 |
| 10.4 Relationship Persistence | todo | M | 10.1 |

**Total Size:** 4M = ~12-20 days

**Can Start:** Story 10.1 (after 5.1 complete)
