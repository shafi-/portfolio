
# PlatformSpecification.md

Version: 0.1

> This specification complements KnowledgeModel.md and defines the implementation
> contracts for the Portfolio platform.

==================================================================
1. DATABASE SCHEMA
==================================================================

Core Tables

projects
- id (UUID, PK)
- name
- root_path
- repository_type
- discovered_at
- updated_at

metadata
- project_id (FK)
- git_head
- default_branch
- last_commit_at
- last_modified_at
- language_summary
- framework_summary
- dependency_summary
- documentation_hash
- last_scan_at

documents
- id
- project_id
- path
- kind (README, ADR, DOC, CHANGELOG, ...)
- content
- content_hash
- indexed_at

analyses
- id
- project_id
- analyzer
- analyzed_git_head
- analyzed_at
- summary
- purpose
- architecture
- maturity
- strengths
- weaknesses
- reusable_components
- notes
- raw_json

features
- id
- analysis_id
- name
- description
- confidence

technologies
- id
- name
- category

project_technologies
- project_id
- technology_id

relationships
- id
- source_project
- target_project
- type
- description
- confidence

configuration
- key
- value

Design Principles

- Store facts.
- Compute indicators.
- Never duplicate deterministic metadata.
- Analyses are versionable.

==================================================================
2. MCP TOOL SPECIFICATION
==================================================================

Discovery

- health()
- discoverProjects()
- listProjects()
- getProject(id)

Search

- searchProjects(query)
- searchDocumentation(query)

Analysis

- getAnalysis(projectId)
- storeAnalysis(projectId, analysis)
- listProjectsNeedingAnalysis()

Configuration

- getConfiguration()
- updateConfiguration()

Relationships

- listRelationships(projectId)

Rules

- Small composable tools
- Stateless where possible
- Deterministic outputs

==================================================================
3. HTTP API
==================================================================

GET /health

GET /projects
GET /projects/{id}

GET /projects/{id}/documents
GET /projects/{id}/analysis

GET /search?q=

GET /relationships
GET /relationships/{projectId}

GET /statistics

GET /configuration

PATCH /configuration

Dashboard consumes only HTTP.

Agents consume MCP.

==================================================================
4. AI AGENT SPECIFICATION
==================================================================

Responsibilities

- Discover projects
- Detect changes
- Decide when analysis is needed
- Investigate repositories
- Produce semantic knowledge
- Persist analyses

Recommended workflow

User
↓

Portfolio question
↓

health()

↓

discoverProjects()

↓

Search metadata

↓

If semantic knowledge missing or outdated:
    Investigate repository
    Produce analysis
    storeAnalysis()

↓

Answer user

Never edit repositories.

Prefer existing knowledge before re-analysis.

==================================================================
5. DASHBOARD SPECIFICATION
==================================================================

Read-only.

Pages

1. Portfolio Overview
- counts
- activity
- technologies

2. Project List
- search
- filters
- sorting

3. Project Detail
- metadata
- documentation
- analysis
- relationships

4. Relationship Explorer

5. Statistics

Never invokes AI.

Never modifies knowledge.

==================================================================
6. SYNCHRONIZATION WITH KNOWLEDGEMODEL
==================================================================

KnowledgeModel.md defines concepts.

This document defines implementation.

Mapping

Project -> projects
Metadata -> metadata
Documentation -> documents
Analysis -> analyses
Feature -> features
Technology -> technologies
Relationship -> relationships

Every API and MCP tool exchanges these canonical entities.

==================================================================
7. IMPLEMENTATION ORDER
==================================================================

1. Database
2. Discovery
3. Metadata extraction
4. Documentation indexing
5. Search
6. HTTP API
7. MCP
8. Agent integration
9. Dashboard
10. Portfolio intelligence

==================================================================
Guiding Principle

The Engine owns deterministic knowledge.

AI Agents own semantic understanding.

The Dashboard visualizes.

The CLI bootstraps.

Everything is built around the Knowledge Model.
