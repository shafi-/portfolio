# Epic 10 — AI Analysis

**Milestone:** 2 — Agent Integration
**Status:** Completed

## Overview

Implement analysis schema, analysis persistence, stale analysis detection, and relationship persistence to support AI-generated semantic knowledge.

---

## Story 10.1: Analysis Schema

**Status:** Completed
**Size:** M
**Blocked by:** 5.1

**User Story:**
As the Portfolio Engine, I want an analysis schema so that AI agents can provide consistent semantic knowledge.

**Acceptance Criteria:**
- ✅ JSON schema for analysis object
- ✅ Fields: summary, purpose, architecture, maturity, strengths, weaknesses, reusable_components, notes, analyzed_at, analyzed_git_head, analyzer
- ✅ Schema validation on storeAnalysis()
- ✅ Stores raw_json for flexibility

**Technical Context:**
- Per KnowledgeModel.md, Analysis entity
- Added validation in pkg/models/validation.go
- Database migration (version 6) adds new fields

---

## Story 10.2: Persist Analyses

**Status:** Completed
**Size:** M
**Blocked by:** 10.1, 7.4

**User Story:**
As an AI agent, I want to store analyses so that semantic knowledge persists across sessions.

**Acceptance Criteria:**
- ✅ storeAnalysis() creates/updates analysis record
- ✅ Stores analyzer (which agent created it)
- ✅ Links to project and git_head
- ✅ Overwrites previous analysis by same analyzer (via INSERT OR REPLACE)

**Technical Context:**
- Unique constraint on (project_id, analyzer)
- GetAnalysisByProjectAndAnalyzer() method for querying by analyzer

---

## Story 10.3: Detect Stale Analyses

**Status:** Completed
**Size:** M
**Blocked by:** 3.1, 10.2

**User Story:**
As the Portfolio Engine, I want to detect outdated analyses so that agents know when to re-analyze.

**Acceptance Criteria:**
- ✅ Compares analyzed_git_head vs current git_head
- ✅ Computes analysis_outdated indicator
- ✅ listProjectsNeedingAnalysis() returns structured response:
  - `no_analysis`: Projects with no analysis
  - `stale_analysis`: Projects with outdated analysis (includes analyzed_at, analyzed_git_head, current_git_head)
  - `counts`: Breakdown by category
- ✅ Projects without analysis are in `no_analysis` array
- ✅ Projects with stale analysis are in `stale_analysis` array
- ✅ Projects with up-to-date analysis excluded

**Technical Context:**
- Store facts (git_head), compute indicators (analysis_outdated)
- IsAnalysisOutdated() helper function in pkg/models/validation.go
- MCP tool listProjectsNeedingAnalysis returns structured data

---

## Story 10.4: Relationship Persistence

**Status:** Completed
**Size:** M
**Blocked by:** 10.1

**User Story:**
As an AI agent, I want to store relationships so that cross-project connections are preserved.

**Acceptance Criteria:**
- ✅ Relationships API in MCP tools (listRelationships, storeRelationship)
- ✅ Stores: source_project, target_project, type, description, confidence
- ✅ Relationship types: Similar, Evolution, Shared Feature, Shared Technology, Reuses Component
- ✅ Queryable via listRelationships(projectId)
- ✅ HTTP API: GET/POST /relationships/{id}

**Technical Context:**
- Validation in ValidateRelationship() in pkg/models/validation.go
- Confidence score validation (0-1)
- Prevents self-relationships

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 10.1 Analysis Schema | Completed | M | 5.1 |
| 10.2 Persist Analyses | Completed | M | 10.1, 7.4 |
| 10.3 Detect Stale Analyses | Completed | M | 3.1, 10.2 |
| 10.4 Relationship Persistence | Completed | M | 10.1 |

**Total Size:** 4M = ~12-20 days

**Completed:** All stories completed

---

## Implementation Summary

### Files Added:
- `pkg/models/validation.go` - Analysis and relationship validation
- `pkg/models/validation_test.go` - Comprehensive validation tests
- `internal/store/analyses_test.go` - Analysis store tests
- `internal/store/relationships_test.go` - Relationship store tests
- `internal/store/test_setup.go` - Test infrastructure
- `internal/api/analysis.go` - HTTP API for analysis retrieval

### Files Modified:
- `pkg/models/database.go` - Updated Analysis model with new fields
- `internal/database/migrations.go` - Migration 6 for AI analysis schema
- `internal/store/analyses.go` - Added new fields, overwrite logic, GetAnalysisByProjectAndAnalyzer
- `internal/mcp/tools.go` - Enhanced storeAnalysis, added storeRelationship
- `internal/api/relationships.go` - Added POST endpoint for relationships
- `internal/api/server.go` - Added new endpoints
- `internal/api/projects.go` - Added analyses to project response

### Key Features:
1. **Enhanced Analysis Schema**: maturity, strengths, weaknesses, reusable_components
2. **Analyzer Overwrite**: INSERT OR REPLACE for (project_id, analyzer) uniqueness
3. **Comprehensive Validation**: Analysis, relationship, JSON validation
4. **Stale Detection**: IsAnalysisOutdated() helper for git HEAD comparison
5. **Relationship Management**: Full CRUD with validation
6. **HTTP API**: Analysis and relationship endpoints
7. **MCP Tools**: Enhanced analysis and relationship tools
8. **Production Tests**: Full test coverage for deterministic logic

### Test Coverage:
- All store operations (analyses, relationships)
- Validation logic (analyses, relationships, JSON)
- Deterministic behavior (overwrite by analyzer)
- Edge cases (nil analysis, empty git head, confidence bounds)

### Quality Gates:
- ✅ All tests passing (100%)
- ✅ Build succeeds
- ✅ Schema migration runs cleanly
- ✅ Validation prevents invalid data
- ✅ Overwrite behavior correct
- ✅ Stale detection works as specified

---

## Usage Examples

### MCP - Store Analysis:
```
{
  "project_id": "proj-123",
  "analyzer": "claude-code",
  "analyzed_git_head": "abc123",
  "summary": "E-commerce platform",
  "purpose": "Online retail system",
  "architecture": "Microservices",
  "maturity": "Production",
  "strengths": "Scalable, well-tested",
  "weaknesses": "Legacy payment integration",
  "reusable_components": "Auth service, payment gateway",
  "notes": "Consider modernizing payment system"
}
```

### MCP - Store Relationship:
```
{
  "source_project": "proj-1",
  "target_project": "proj-2",
  "type": "Shared Technology",
  "description": "Both projects use Go and PostgreSQL",
  "confidence": 0.9
}
```

### HTTP API:
- `GET /projects/{id}/analysis` - Get all analyses for a project
- `POST /relationships/{id}` - Create a new relationship
- `GET /relationships/{id}` - Get relationships for a project
