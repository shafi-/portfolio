---
layout: home

hero:
  name: Portfolio
  text: The engine knows. The agent thinks.
  tagline: Local-first project inventory and knowledge platform for developers and AI coding agents. Discover every repository on your machine, extract its metadata, index its documentation — 100% local, offline, and ready for your AI agents.
  actions:
    - theme: brand
      text: Get Started
      link: /docs/getting-started/introduction
    - theme: alt
      text: Installation
      link: /docs/getting-started/installation
    - theme: alt
      text: GitHub
      link: https://github.com/shafi-/portfolio

features:
  - icon: 🧭
    title: Automatic Discovery
    details: Recursively scans your configured project roots to find every Git repository, with smart ignore patterns and incremental re-discovery.
    link: /docs/concepts/discovery
    linkText: How discovery works
  - icon: 🧬
    title: Rich Metadata
    details: Extracts languages, frameworks, dependencies, git intelligence, and capability signals — deterministically, without calling any AI.
    link: /docs/concepts/metadata
    linkText: What gets extracted
  - icon: 🔎
    title: Documentation Search
    details: Indexes READMEs, ADRs, architecture docs, OpenAPI specs, and changelogs into a full-text SQLite knowledge store.
    link: /docs/concepts/knowledge-store
    linkText: The knowledge store
  - icon: 🤖
    title: Built for AI Agents
    details: Ships with an MCP stdio server exposing dozens of tools, plus one-command installation for Claude Code and OpenCode.
    link: /docs/mcp/
    linkText: MCP integration
  - icon: 📊
    title: HTTP API & Dashboard
    details: A localhost REST API and an embedded, read-only dashboard for browsing projects, statistics, and cross-project relationships.
    link: /docs/api/
    linkText: HTTP API reference
  - icon: 🔒
    title: Local-First & Private
    details: Everything lives in an encrypted, obscured SQLite database on your machine. No telemetry, no cloud, no accounts.
    link: /docs/digging-deeper/security
    linkText: Security model
---
