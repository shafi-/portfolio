---
title: Introduction
description: What Portfolio is, the philosophy behind it, and how its pieces fit together.
---

# Introduction

Portfolio is a **local-first project inventory and knowledge platform** for developers and AI coding agents. It discovers every Git repository on your machine, extracts deterministic metadata from each one, and stores everything — along with a searchable index of your project documentation — in a local SQLite knowledge base.

Install it once, run a scan, and forget about it. Your entire software portfolio becomes queryable from the CLI, an HTTP API, a web dashboard, and your AI coding assistant.

## Why Portfolio?

Most developers work across dozens of repositories and can't answer simple questions about them: *Which projects use Express? What did I touch last month? Where are the architecture docs for that service?* AI coding assistants are even more blind — they only see the project currently open.

Portfolio solves this with a simple split of responsibilities:

> **Engine knows, agent thinks.**

The Portfolio Engine does deterministic work — discovering repositories, parsing manifests, extracting git history, indexing documents. Your AI agent does the semantic reasoning on top, querying the engine instead of guessing.

## Key principles

- **Local-first** — Everything is stored in a SQLite database on your machine. No cloud, no telemetry, no accounts. It works fully offline.
- **Deterministic by default** — Discovery, metadata extraction, and indexing are repeatable and stable. No AI in the critical path.
- **Zero maintenance** — Set it up once. Incremental scans keep it fresh; there is nothing to babysit.
- **AI-ready** — A built-in [MCP server](/docs/mcp/) exposes the knowledge base as tools for Claude Code, OpenCode, and any MCP-compatible client.

## What you get

Portfolio ships as a single static binary with four surfaces:

| Surface | Command | What it does |
| --- | --- | --- |
| CLI | `portfolio …` | Administration and querying from the terminal |
| HTTP API | `portfolio serve` | Localhost REST API for integrations |
| Dashboard | `portfolio dashboard` | Read-only web UI for browsing your portfolio |
| MCP server | `portfolio mcp` | Tools for AI agents over stdio |

## Where to go next

- [Installation](/docs/getting-started/installation) — get the binary on your machine
- [Quick Start](/docs/getting-started/quick-start) — from zero to a scanned portfolio in five minutes
- [Core Concepts](/docs/concepts/discovery) — how discovery, metadata, and the knowledge store work
