export const meta = {
  name: 'analyze-remaining-projects',
  description: 'Sequentially analyze all projects that need analysis',
  phases: [
    { title: 'Discovery', detail: 'Find projects needing analysis' },
    { title: 'Analysis', detail: 'Analyze each project' }
  ]
}

phase('Discovery')

// Use direct tool to get projects needing analysis
log('Finding projects that need analysis...')

const result = await agent(
`Call mcp__portfolio__listProjectsNeedingAnalysis.
Return ONLY the "no_analysis" array: [{"id":"...","name":"..."},...]`,
{ label: 'List projects needing analysis', phase: 'Discovery' }
)

// Parse result - handle string or array, extract no_analysis only
let projectsToAnalyze = result
if (typeof result === 'string') {
  const clean = result.replace(/```json\n?/g, '').replace(/```\n?/g, '').trim()
  const parsed = JSON.parse(clean)
  // If result has no_analysis field, use that; otherwise use parsed as-is
  projectsToAnalyze = parsed.no_analysis || parsed
} else if (result && result.no_analysis) {
  projectsToAnalyze = result.no_analysis
}

if (!projectsToAnalyze || projectsToAnalyze.length === 0) {
  log('All projects already have analysis!')
  return { projects_analyzed: 0, message: 'All projects complete' }
}

log(`Found ${projectsToAnalyze.length} projects needing analysis: ${projectsToAnalyze.map(p => p.name).join(', ')}`)

phase('Analysis')

const analyzed = []
const skipped = []
const failures = []

// Process each project sequentially
for (const [index, project] of projectsToAnalyze.entries()) {
  log(`[${index + 1}/${projectsToAnalyze.length}] ${project.name}...`)

  try {
    const result = await agent(
`Analyze project "${project.name}" (ID: ${project.id}).
Use portfolio MCP tools.
Store features + analysis.
Return: {name, maturity, features_stored}`,
      {
        label: project.name,
        phase: 'Analysis',
        isolation: 'worktree'
      }
    )

    if (result) {
      analyzed.push(project.name)
      log(`  ✓ Done`)
    } else {
      failures.push(project.name)
      log(`  ✗ Failed`)
    }

  } catch (e) {
    failures.push(project.name)
    log(`  ✗ Error: ${e.message}`)
  }
}

log(`\n=== Summary ===`)
log(`Analyzed: ${analyzed.length}`)
log(`Failed: ${failures.length}`)

if (failures.length > 0) {
  log(`Failed projects: ${failures.join(', ')}`)
}

return {
  total: projectsToAnalyze.length,
  analyzed: analyzed.length,
  analyzed_projects: analyzed,
  failed: failures.length,
  failed_projects: failures
}
