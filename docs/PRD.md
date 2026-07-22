# PRD.md

# Portfolio — Product Requirements Document

Version: 1.0

## Vision

Portfolio is a local-first project inventory and knowledge platform that enables developers and AI coding agents to understand an entire software portfolio.

It is not a project management tool. It is a portfolio awareness system.

---

## Problem

Developers accumulate dozens or hundreds of repositories over time and lose visibility into:

- what they have built
- what each project does
- which projects are active
- what can be reused
- how projects relate to one another

Current tools either require manual maintenance or only represent a subset of a developer's work.

---

## Goals

Portfolio should allow developers to:

- Discover every project.
- Search projects instantly.
- Rediscover forgotten work.
- Understand relationships across projects.
- Ask AI natural-language questions about their portfolio.
- Keep knowledge current with minimal effort.

---

## Non-Goals

Portfolio is not:

- Project management software
- Issue tracking
- Documentation authoring
- Source control
- CI/CD
- Deployment tooling

---

## Product Principles

- Local-first
- Deterministic engine
- AI-native integrations
- Zero ongoing maintenance
- Read-only dashboard
- Existing folder structure is respected

---

## User Experience

Typical lifecycle:

```text
Install
    ↓
Initialize
    ↓
Forget
```

After initialization:

- The AI coding agent is the primary interface.
- The dashboard is the primary exploration interface.
- The CLI is reserved for administration.

---

## System Components

- Portfolio Engine
- MCP Interface
- HTTP API
- Agent Integrations
- Dashboard
- Local Knowledge Store

---

## Success Criteria

A successful implementation enables a developer to:

- View every project in one place.
- Find projects by technology, feature, or documentation.
- Recover context without rereading source code.
- Explore portfolio-wide relationships.
- Ask AI meaningful questions about the portfolio.
- Operate without manually maintaining an inventory.

---

## Scope

This document intentionally remains concise.

The authoritative technical specifications are maintained in:

- **KnowledgeModel.md** — canonical domain model.
- **PlatformSpecification.md** — implementation contracts, APIs, database schema, dashboard, MCP tools, and agent workflows.

All implementation decisions should derive from those documents rather than duplicating technical details here.
