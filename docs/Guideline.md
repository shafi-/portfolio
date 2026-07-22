# Guideline.md

# Portfolio Engineering Guidelines

Version: 1.0

---

# Purpose

This document defines the engineering principles that guide implementation.

The authoritative specifications are:

- KnowledgeModel.md
- PlatformSpecification.md

When implementation decisions are unclear, these guidelines should be applied before introducing new patterns.

---

# Core Principles

## Engine Knows, Agent Thinks

The Portfolio Engine performs deterministic work only.

Examples:

- Discover repositories
- Extract metadata
- Index documentation
- Store and retrieve knowledge
- Search

AI coding agents perform semantic work.

Examples:

- Summaries
- Feature extraction
- Architecture understanding
- Project comparisons
- Relationship discovery

Never move semantic reasoning into the engine.

---

## Deterministic by Default

Every engine operation should produce repeatable results for the same input.

Avoid heuristics that require an LLM.

---

## Store Facts, Compute Indicators

Persist immutable facts.

Examples:

- git HEAD
- documentation hash
- timestamps

Compute indicators when needed.

Examples:

- Needs analysis
- Analysis outdated
- Recently modified

Do not store derived state unless necessary.

---

## Local First

Repositories remain on the user's machine.

Knowledge remains local.

Cloud services are optional extensions, not requirements.

---

## Capabilities over Workflows

Expose small, composable capabilities.

Examples:

- discoverProjects
- listProjects
- searchProjects
- storeAnalysis

The engine should not implement high-level workflows; AI agents compose them.

---

## AI is Optional

Portfolio should provide value immediately after deterministic discovery.

AI enriches the portfolio but is not required for core functionality.

---

## Dashboard is Read-only

The dashboard visualizes knowledge.

It must not:

- invoke AI
- modify repositories
- perform analysis

---

## CLI is Administrative

The CLI exists for:

- initialization
- upgrades
- diagnostics
- integration management

Users should interact with Portfolio primarily through their AI coding agent.

---

## Agent Agnostic

The engine must never depend on a specific AI assistant.

Agent-specific behavior belongs in installable integrations.

---

## Single Knowledge Model

Every interface operates on the same canonical model.

- Database
- MCP
- HTTP API
- Dashboard
- AI agents

No interface should invent its own representation.

---

# Coding Guidelines

- Prefer composition over inheritance.
- Keep packages cohesive.
- Avoid global state.
- Design interfaces around capabilities.
- Keep public APIs stable.
- Minimize dependencies.
- Write tests for deterministic logic.

---

# Documentation

Every significant architectural decision should result in an ADR.

When the domain model changes:

1. Update KnowledgeModel.md.
2. Update PlatformSpecification.md.
3. Update implementation.

Avoid duplicating specifications across documents.

---

# Product Philosophy

Portfolio is infrastructure.

The ideal user journey is:

Install → Initialize → Forget

The AI coding agent becomes the primary interface.

The dashboard becomes the primary exploration interface.

The engine quietly maintains deterministic knowledge that enables both.
