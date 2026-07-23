# Epic 7 — MCP Server

**Milestone:** 1 — Core Engine
**Status:** done

## Overview

Implement MCP server with tools for discovery, search, analysis storage, configuration, and relationships to enable AI agent integration.

---

## Story 7.1: MCP Server Foundation

**Status:** todo
**Size:** M
**Blocked by:** 6.2

**User Story:**
As the Portfolio Engine, I want an MCP server so that AI agents can interact with Portfolio.

**Acceptance Criteria:**
- MCP server implementation per MCP specification
- Server starts on configured port/socket
- Proper tool registration
- Error handling and logging

**Technical Context:**
- Uses Go MCP SDK or stdio transport
- Exposes tools defined in PlatformSpecification.md

---

## Story 7.2: Discovery Tools

**Status:** todo
**Size:** M
**Blocked by:** 7.1

**User Story:**
As an AI agent, I want discovery tools so that I can explore the user's portfolio.

**Acceptance Criteria:**
- health() → engine status
- discoverProjects() → trigger discovery, return count
- listProjects() → return all projects
- getProject(id) → return single project

**Tool Contracts:**
- Stateless where possible
- Deterministic outputs
- Proper error responses

---

## Story 7.3: Search Tools

**Status:** todo
**Size:** M
**Blocked by:** 7.1

**User Story:**
As an AI agent, I want search tools so that I can find relevant projects and documentation.

**Acceptance Criteria:**
- searchProjects(query) → search projects by name, tech stack
- searchDocumentation(query) → search document contents via FTS
- Returns ranked results with relevance scores

---

## Story 7.4: Analysis Storage Tools

**Status:** todo
**Size:** M
**Blocked by:** 7.1

**User Story:**
As an AI agent, I want to store analyses so that semantic knowledge persists for future conversations.

**Acceptance Criteria:**
- getAnalysis(projectId) → retrieve stored analysis
- storeAnalysis(projectId, analysis) → persist analysis JSON
- listProjectsNeedingAnalysis() → return projects without or with outdated analysis
- Validates analysis schema

---

## Story 7.5: Configuration Tools

**Status:** todo
**Size:** S
**Blocked by:** 7.1

**User Story:**
As an AI agent, I want to read configuration so that I can understand Portfolio setup.

**Acceptance Criteria:**
- getConfiguration() → return current config
- updateConfiguration(config) → update with validation

---

---

## Story 7.6: Relationship Tools

**Status:** todo
**Size:** S
**Blocked by:** 7.1

**User Story:**
As an AI agent, I want to query project relationships so that I can understand connections between projects.

**Acceptance Criteria:**
- listRelationships(projectId) → return all relationships for a project
- Returns empty list if no relationships exist
- Each relationship includes: source_project, target_project, type, description, confidence

---

## Epic Status Summary

| Story | Status | Size | Blocked By |
|-------|--------|------|------------|
| 7.1 MCP Server Foundation | todo | M | 6.2 |
| 7.2 Discovery Tools | todo | M | 7.1 |
| 7.3 Search Tools | todo | M | 7.1 |
| 7.4 Analysis Storage Tools | todo | M | 7.1 |
| 7.5 Configuration Tools | todo | S | 7.1 |
| 7.6 Relationship Tools | todo | S | 7.1 |

**Total Size:** 4M + 2S = ~15 days

**Can Start:** Story 7.1 (after 6.2 complete)
