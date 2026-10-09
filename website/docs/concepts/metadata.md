---
title: Metadata Extraction
description: What Portfolio extracts from each project — git intelligence, languages, frameworks, dependencies, and capability signals.
---

# Metadata Extraction

Once a project is [discovered](/docs/concepts/discovery), `portfolio scan` extracts deterministic metadata from it. No AI is involved — the same repo always produces the same metadata.

```bash
portfolio scan                 # all projects
portfolio scan --project <id>  # a single project
```

## Git intelligence

Read straight from the local repository:

- Current `HEAD` and default branch
- Last commit timestamp and total commit count
- Contributors
- Remotes (origin and others)

This gives you a live picture of where each project stands — without contacting any remote service.

## Languages, frameworks, and dependencies

Portfolio detects **13+ languages** by inspecting source trees and manifests, and identifies frameworks and package managers from lockfiles and manifests (`package.json`, `go.mod`, `composer.json`, `requirements.txt`, and friends). Detected dependencies feed the [technologies](/docs/api/endpoints#get-technologies) index used for cross-project queries like *"which of my projects use Postgres?"*

Each detection is summarized per project in a language summary and framework summary stored alongside the metadata.

## Capability signals

Beyond names and versions, Portfolio categorizes what a project *does* — database access, authentication, payments, background jobs, and similar capabilities — by recognizing well-known libraries. These capability signals power the dashboard's project cards and make portfolio-wide questions answerable ("show me everything that touches Stripe").

## Health and maturity indicators

Presence of key project files is recorded as lightweight health signals:

| Signal | File |
| --- | --- |
| README | `README.md` |
| License | `LICENSE` |
| CI | `.github/workflows/` |
| Docs | `docs/` directory |
| ADRs | Architecture decision records |

## Documentation indexing

The other half of the scan is [documentation indexing](/docs/concepts/knowledge-store): READMEs, ADRs, architecture documents, OpenAPI specs, and changelogs are parsed and pushed into the full-text store, with change tracking so unchanged documents aren't re-indexed.

## Why deterministic matters

Because extraction is rule-based, results are stable across runs and machines. AI agents can rely on Portfolio's answers as *ground truth* about your portfolio and spend their reasoning on interpretation instead of discovery.
