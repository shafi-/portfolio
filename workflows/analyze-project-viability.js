export const meta = {
  name: 'analyze-project-viability',
  description: 'Token-efficient project viability analysis with market research',
  phases: [
    { title: 'Data Collection', detail: 'Get essential project data' },
    { title: 'Viability Analysis', detail: 'Market research and scoring' },
    { title: 'Documentation', detail: 'Generate markdown report' }
  ]
}

phase('Data Collection')

log('Gathering projects for viability analysis...')

const projectsData = await agent(
`Get projects with analysis.

**Your Task:**
1. Call mcp__portfolio__listProjects
2. Call mcp__portfolio__getAnalysis for each
3. **SKIP** "not a real project" entries

**Return ONLY:**
{
  "projects": [
    {
      "id": "<id>",
      "name": "<name>",
      "purpose": "<from analysis>",
      "maturity": "<from analysis>"
    }
  ]
}

NO extra text, NO markdown.`,
  { label: 'Get projects', phase: 'Data Collection' }
)

if (!projectsData) {
  log('No projects found')
  return { error: 'No projects' }
}

let projects
if (typeof projectsData === 'string') {
  try {
    const jsonStart = projectsData.indexOf('{')
    const jsonEnd = projectsData.lastIndexOf('}')
    projects = JSON.parse(projectsData.substring(jsonStart, jsonEnd + 1).replace(/```json\n?/g, '').replace(/```\n?/g, '').trim())
  } catch (e) {
    return { error: 'Parse error' }
  }
} else {
  projects = projectsData
}

log(`Loaded ${projects.projects.length} projects`)

phase('Viability Analysis')

log('Analyzing viability with market research...')

// Analyze in batches of 5 to balance speed and cost
const batchSize = 5
const viabilityScores = []

for (let i = 0; i < projects.projects.length; i += batchSize) {
  const batch = projects.projects.slice(i, i + batchSize)
  log(`Analyzing batch ${Math.floor(i/batchSize) + 1}/${Math.ceil(projects.projects.length/batchSize)}...`)

  const batchResults = await parallel(
    batch.map(project => () => agent(
`Analyze MARKET VIABILITY:

**Project:** ${project.name}
**Purpose:** ${project.purpose}
**Maturity:** ${project.maturity}

**Research (use web search):**
1. Market size (tiny/small/medium/large/huge)
2. Competition level (none/low/medium/high/saturated)
3. Uniqueness (none/low/medium/high/very-high)
4. Monetization potential (none/low/medium/high)

**Top 3 competitors:**

**Opportunity score (0-10):**

**Return ONLY this JSON:**
{
  "project_id": "${project.id}",
  "market_size": "<size>",
  "competition": "<level>",
  "uniqueness": "<level>",
  "monetization": "<level>",
  "competitors": ["<name1>", "<name2>", "<name3>"],
  "opportunity_score": <0-10>,
  "viability_score": <0-100 calculated as: market(25) + low_competition(20) + uniqueness(25) + monetization(20) + maturity(10)>
}`,
      { label: `Viability: ${project.name}`, phase: 'Viability Analysis', isolation: 'worktree' }
    ))
  )

  // Parse results
  batchResults.forEach(result => {
    if (result && typeof result === 'string') {
      try {
        const jsonStart = result.indexOf('{')
        const jsonEnd = result.lastIndexOf('}')
        if (jsonStart !== -1 && jsonEnd !== -1) {
          const parsed = JSON.parse(result.substring(jsonStart, jsonEnd + 1).replace(/```json\n?/g, '').replace(/```\n?/g, '').trim())
          viabilityScores.push(parsed)
        }
      } catch (e) {
        log(`  ⚠ Parse error for result`)
      }
    } else if (result && typeof result === 'object') {
      viabilityScores.push(result)
    }
  })

  log(`  ✓ Batch complete`)
}

// Generate recommendations
const recommendations = viabilityScores.map(score => {
  const project = projects.projects.find(p => p.id === score.project_id)
  let recommendation, priority, rationale

  if (score.viability_score >= 70) {
    recommendation = 'CONTINUE'
    priority = 'HIGH'
    rationale = `Strong potential (${score.market_size} market, ${score.competition} competition, ${score.uniqueness} uniqueness)`
  } else if (score.viability_score >= 50) {
    recommendation = 'CONTINUE'
    priority = 'MEDIUM'
    rationale = `Moderate potential, needs validation (${score.market_size} market, ${score.competition} competition)`
  } else if (score.viability_score >= 30) {
    recommendation = 'PAUSE'
    priority = 'LOW'
    rationale = `Weak signals - consider pivot (${score.competition} competition, ${score.uniqueness} uniqueness)`
  } else {
    recommendation = 'DROP'
    priority = 'NONE'
    rationale = `Poor market fit (${score.market_size} market, ${score.competition} competition, ${score.uniqueness} uniqueness)`
  }

  return {
    ...project,
    ...score,
    recommendation,
    priority,
    rationale
  }
})

// Sort by viability score
recommendations.sort((a, b) => b.viability_score - a.viability_score)

phase('Documentation')

log('Generating viability report...')

const toContinue = recommendations.filter(r => r.recommendation === 'CONTINUE')
const toPause = recommendations.filter(r => r.recommendation === 'PAUSE')
const toDrop = recommendations.filter(r => r.recommendation === 'DROP')

const report = `# Project Viability Analysis

**Projects Analyzed:** ${recommendations.length}

---

## 📊 Summary

| Recommendation | Count | Avg Viability |
|----------------|-------|---------------|
| **CONTINUE (HIGH)** | ${toContinue.filter(r => r.priority === 'HIGH').length} | - |
| **CONTINUE (MEDIUM)** | ${toContinue.filter(r => r.priority === 'MEDIUM').length} | - |
| **PAUSE** | ${toPause.length} | - |
| **DROP** | ${toDrop.length} | - |

---

## ⚡ High Priority - CONTINUE

${toContinue.filter(r => r.priority === 'HIGH').map(r => `
### ${r.name} [${r.viability_score}/100]

**Purpose:** ${r.purpose}
**Rationale:** ${r.rationale}

**Market Analysis:**
- Market Size: ${r.market_size}
- Competition: ${r.competition}
- Uniqueness: ${r.uniqueness}
- Monetization: ${r.monetization}
- Opportunity Score: ${r.opportunity_score}/10

**Top Competitors:**
${r.competitors.map(c => `- ${c}`).join('\n')}

**Status:** ${r.maturity}
`).join('\n---\n') || 'No high-priority projects identified.'}

---

## 🔄 Medium Priority - CONTINUE

${toContinue.filter(r => r.priority === 'MEDIUM').map(r => `
### ${r.name} [${r.viability_score}/100]

**Purpose:** ${r.purpose}
**Rationale:** ${r.rationale}

**Market:** ${r.market_size} | ${r.competition} competition | ${r.uniqueness} uniqueness
**Monetization:** ${r.monetization} | Opportunity: ${r.opportunity_score}/10
**Status:** ${r.maturity}
`).join('\n---\n') || 'No medium-priority projects.'}

---

## ⏸️  PAUSE for Validation

${toPause.map(r => `
### ${r.name} [${r.viability_score}/100]

**Purpose:** ${r.purpose}
**Rationale:** ${r.rationale}

**Market:** ${r.market_size} | ${r.competition} competition | ${r.uniqueness} uniqueness
**Status:** ${r.maturity}
`).join('\n---\n') || 'No projects to pause.'}

---

## ❌ DROP - Poor Market Fit

${toDrop.map(r => `
### ${r.name} [${r.viability_score}/100]

**Purpose:** ${r.purpose}
**Rationale:** ${r.rationale}

**Market:** ${r.market_size} | ${r.competition} competition | ${r.uniqueness} uniqueness
**Status:** ${r.maturity}
`).join('\n---\n') || 'No projects to drop.'}

---

## 📋 All Projects Ranked

| Rank | Project | Viability | Market | Competition | Uniqueness | Recommendation |
|------|---------|-----------|--------|-------------|------------|----------------|
${recommendations.map((r, i) => `| ${i + 1} | ${r.name} | ${r.viability_score}/100 | ${r.market_size} | ${r.competition} | ${r.uniqueness} | ${r.recommendation} |`).join('\n')}

---

## 🔍 Insights

### Strongest Opportunities
${recommendations.slice(0, 3).map((r, i) => `${i + 1}. **${r.name}** - ${r.market_size} market, ${r.competition} competition, ${r.opportunity_score}/10 opportunity`).join('\n')}

### Most Competitive Markets
${[...new Set(recommendations.map(r => r.competition))].filter(c => c === 'high' || c === 'saturated').map(c => {
  const projects = recommendations.filter(r => r.competition === c)
  return `- **${c}**: ${projects.map(p => p.name).join(', ')}`
}).join('\n') || 'No highly competitive markets identified.'}

### Highest Uniqueness
${recommendations.filter(r => r.uniqueness === 'high' || r.uniqueness === 'very-high').map(r => `- **${r.name}** - ${r.uniqueness} uniqueness in ${r.market_size} market`).join('\n') || 'No highly unique projects identified.'}

### Monetization Ready
${recommendations.filter(r => r.monetization === 'high' || r.monetization === 'medium').map(r => `- **${r.name}** - ${r.monetization} monetization potential`).join('\n') || 'No projects with strong monetization identified.'}

---

## 🎯 Recommendations Summary

**Immediate Actions:**
- Focus resources on ${toContinue.filter(r => r.priority === 'HIGH').length} high-priority projects
- Pause ${toPause.length} projects for validation
- Drop ${toDrop.length} projects with poor market fit

**Strategic Focus:**
- Next 6 months: Top ${Math.min(3, toContinue.filter(r => r.priority === 'HIGH').length)} high-viability projects
- Re-evaluate paused projects in 3-6 months based on market changes
`

log(`\n=== VIABILITY ANALYSIS COMPLETE ===`)
log(`Report saved to: analysis-docs/viability-analysis.md`)

return {
  recommendations,
  report,
  summary: {
    high_priority: toContinue.filter(r => r.priority === 'HIGH').length,
    medium_priority: toContinue.filter(r => r.priority === 'MEDIUM').length,
    pause: toPause.length,
    drop: toDrop.length
  }
}
