# Architecture.md

# Portfolio Architecture

Version: 1.0

---

# Purpose

This document describes the high-level architecture of Portfolio.

It intentionally avoids implementation details.

The canonical definitions are maintained in:

- KnowledgeModel.md
- PlatformSpecification.md

---

# System Philosophy

Portfolio follows a simple separation of responsibilities.

> **The Engine knows. The AI Agent thinks.**

The engine is deterministic.

AI agents provide reasoning.

The dashboard visualizes knowledge.

---

# System Overview

```text
                     Human
                        │
        ┌───────────────┴───────────────┐
        │                               │
        ▼                               ▼
 AI Coding Agent                  Dashboard
 (Primary Interface)          (Read-only Interface)
        │                               │
        └───────────────┬───────────────┘
                        │
               Agent Integration
                        │
                 MCP / HTTP APIs
                        │
                        ▼
               Portfolio Engine
                        │
        ┌───────────────┼────────────────┐
        │               │                │
        ▼               ▼                ▼
   Discovery      Knowledge Store     Search
                        │
                        ▼
              Local Project Repositories
```

---

# Components

## Portfolio Engine

The engine is responsible for deterministic operations.

Responsibilities include:

- Project discovery
- Metadata extraction
- Documentation extraction
- Knowledge storage
- Search
- MCP server
- HTTP API

The engine never performs AI reasoning.

---

## AI Coding Agent

The AI coding agent is the primary user interface.

Responsibilities include:

- Understanding user intent
- Investigating repositories
- Producing semantic knowledge
- Comparing projects
- Maintaining analyses
- Answering natural language questions

The AI agent never owns project data.

---

## Agent Integration

The integration connects an AI coding agent to the Portfolio Engine.

Responsibilities include:

- MCP registration
- Agent instructions / skills
- Installation
- Validation
- Upgrades

The first supported integration is Claude Code.

Future integrations may include:

- Codex CLI
- OpenCode
- Cursor
- Other MCP-compatible agents

---

## Dashboard

The dashboard provides portfolio exploration.

Capabilities include:

- Browse projects
- Search
- Filter
- View metadata
- View analyses
- Explore relationships
- Portfolio statistics

The dashboard never invokes AI.

The dashboard never modifies repositories.

---

## CLI

The CLI exists for administration.

Typical commands include:

- init
- config
- doctor
- upgrade
- integration

The CLI is not intended for day-to-day portfolio interaction.

---

# Knowledge Flow

```text
Repository
      │
      ▼
Discovery
      │
      ▼
Deterministic Metadata
      │
      ▼
Knowledge Store
      │
      ▼
AI Analysis (optional)
      │
      ▼
Semantic Knowledge
      │
      ▼
Dashboard / AI Responses
```

Projects remain useful immediately after discovery and become richer after AI analysis.

---

# Interfaces

Portfolio exposes two interfaces over the same service layer.

```text
              Portfolio Engine
                     │
          ┌──────────┴──────────┐
          ▼                     ▼
      HTTP API              MCP Server
          │                     │
          ▼                     ▼
     Dashboard          AI Coding Agent
```

Both interfaces operate on the same knowledge model.

---

# Deployment

Portfolio is local-first.

Typical installation:

```text
Developer Machine

├── Portfolio Engine
├── SQLite Database
├── Local Repositories
├── Dashboard
└── AI Coding Agent
```

No cloud services are required.

---

# Initialization

Typical lifecycle:

```text
Install
    │
    ▼
portfolio init
    │
    ▼
Configure Project Roots
    │
    ▼
Initial Discovery
    │
    ▼
Install Agent Integration
    │
    ▼
Ready
```

After initialization, Portfolio should require little or no direct maintenance.

---

# Architectural Principles

- Local-first
- Deterministic engine
- AI-native
- Agent-agnostic
- Read-only dashboard
- Capabilities over workflows
- Store facts, compute indicators
- Existing project structures are respected

---

# Source of Truth

Responsibilities are divided across the project documentation:

| Document | Responsibility |
|----------|----------------|
| PRD.md | Product vision and goals |
| KnowledgeModel.md | Canonical domain model |
| PlatformSpecification.md | Database, APIs, MCP tools, dashboard, agent contracts |
| Architecture.md | High-level system design |
| ADR.md | Architectural decisions |
| Guideline.md | Engineering principles |
| Tasks.md | Roadmap and implementation |
