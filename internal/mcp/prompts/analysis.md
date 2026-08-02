# Project Analysis Prompt

You are analyzing a software project for Portfolio. Follow this workflow to produce a comprehensive analysis.

## Investigation Order

1. `getProject(id)` — check existing state
2. `searchDocumentation("overview")` then `searchDocumentation("architecture")` — understand purpose
3. `getProjectStructure(project_id, include_content: true)` — file tree and key files
4. `getDependencies(project_id)` — technology stack
5. `searchFiles(project_id, pattern)` per suspected feature → `getFileContent` to confirm

**Rule:** Investigate only through Portfolio's MCP file tools. Never read the project directory on disk.

## Required Outputs

After investigation, call `storeAnalysis` with ALL of the following:

| Field | Description |
|-------|-------------|
| `project_id` | The project ID |
| `analyzer` | Your identifier (e.g. "claude-code", "codex") |
| `summary` | One-paragraph overview of what the project does |
| `purpose` | Primary goal and target users |
| `architecture` | High-level structure (patterns, key components, data flow) |
| `maturity` | One of: `prototype`, `alpha`, `beta`, `stable`, `mature`, `deprecated` |
| `strengths` | What the project does well (comma-separated or paragraph) |
| `weaknesses` | Areas for improvement (comma-separated or paragraph) |
| `reusable_components` | Libraries, modules, or patterns worth extracting |
| `notes` | Anything notable that doesn't fit above |

## Also Store Features

For each significant feature found during investigation, call `storeFeature`:

- `name` — feature name
- `description` — what it does
- `confidence` — 0.5 (inferred), 0.8 (confirmed via code), 1.0 (explicitly documented)

Zero features = failed analysis.

## Maturity Guidelines

- **prototype**: Proof of concept, not production-ready
- **alpha**: Early stage, API may change, limited testing
- **beta**: Feature-complete, some testing, API stabilizing
- **stable**: Production-ready, well-tested, semantic versioning
- **mature**: Battle-tested, stable API, active maintenance
- **deprecated**: No longer maintained, superseded by another project
