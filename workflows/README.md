# Portfolio Workflows

This directory contains workflow scripts for portfolio analysis and relationship detection.

## Available Workflows

### analyze-remaining-projects.js

Sequentially analyzes all projects that need analysis (no_analysis list).

**Phases:**
1. **Discovery** - Finds projects needing analysis via `mcp__portfolio__listProjectsNeedingAnalysis`
2. **Analysis** - Analyzes each project sequentially using agents with worktree isolation

**Returns:**
```json
{
  "total": 13,
  "analyzed": 13,
  "analyzed_projects": ["project1", "project2", ...],
  "failed": 0,
  "failed_projects": []
}
```

**Usage:**
```javascript
Workflow({
  scriptPath: '/Users/nerddevsltd/Projects/portfolio-tool/workflows/analyze-remaining-projects.js'
})
```

---

### find-project-relationships.js

Two-phase relationship detection: guess candidates from summaries, then confirm with deep analysis.

**Phases:**
1. **Candidate Discovery** - Quick pass to find 15-25 potential relationships from project summaries
2. **Confirmation** - Deep analysis of each candidate (in batches of 5) to verify relationships
3. **Storage** - Store confirmed relationships (single batch to minimize token usage)

**Relationship Types:**
- `Similar` - Projects serving similar domains/purposes
- `Evolution` - One project evolved into another (concept → implementation)
- `Shared Feature` - Frontend-backend pairs or shared functionality
- `Shared Technology` - Same tech stack/architecture patterns
- `Reuses Component` - One project uses components from another

**Returns:**
```json
{
  "candidates_found": 23,
  "confirmed": 12,
  "stored": 12,
  "errors": 0,
  "error_details": []
}
```

**Usage:**
```javascript
Workflow({
  scriptPath: '/Users/nerddevsltd/Projects/portfolio-tool/workflows/find-project-relationships.js'
})
```

**Resume capability:**
```javascript
Workflow({
  scriptPath: '/Users/nerddevsltd/Projects/portfolio-tool/workflows/find-project-relationships.js',
  resumeFromRunId: 'wf_xxx'
})
```

## Design Principles

1. **Token efficiency** - Discovery phase uses pure JS logic, only analysis/storage use agents
2. **Caching** - Workflow resumption reuses cached agent results for unchanged phases
3. **Batch operations** - Storage phase batches all relationships into single agent call
4. **Worktree isolation** - Analysis phases use isolation to prevent conflicts
5. **Error handling** - Graceful degradation with detailed error reporting

## Notes

- Workflow scripts use JavaScript (not TypeScript) per workflow engine requirements
- Cannot call MCP tools directly from workflow JavaScript - must use agents
- Agent results are cached by prompt hash - unchanged phases replay instantly on resume
