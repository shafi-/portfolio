export const meta = {
  name: 'portfolio-strategic-analysis',
  description: 'Token-efficient portfolio analysis with practical recommendations',
  phases: [
    { title: 'Data Collection', detail: 'Gather minimal essential data' },
    { title: 'Analysis', detail: 'Multi-perspective analysis in single agent' },
    { title: 'Documentation', detail: 'Generate markdown report' }
  }

phase('Data Collection')

log('Gathering minimal portfolio data for analysis...')

// Get only essential data - no features, no detailed tech lists
const portfolioData = await agent(
`Get minimal portfolio data.

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
      "maturity": "<from analysis>",
      "domain": "<infer: finance/health/education/productivity/etc>",
      "tech_stack": "<main stack only: Next.js, Flutter, etc>",
      "status": "<active/abandoned/maintenance>",
      "monetization": "<yes/no/potential>"
    }
  ]
}

NO features, NO detailed tech lists, NO extra text.`,
  { label: 'Get minimal portfolio data', phase: 'Data Collection' }
)

if (!portfolioData) {
  log('Failed to get portfolio data')
  return { error: 'No data' }
}

let portfolio
if (typeof portfolioData === 'string') {
  try {
    const jsonStart = portfolioData.indexOf('{')
    const jsonEnd = portfolioData.lastIndexOf('}')
    portfolio = JSON.parse(portfolioData.substring(jsonStart, jsonEnd + 1).replace(/```json\n?/g, '').replace(/```\n?/g, '').trim())
  } catch (e) {
    return { error: 'Parse error' }
  }
} else {
  portfolio = portfolioData
}

log(`Loaded ${portfolio.projects.length} projects`)

phase('Analysis')

log('Running strategic analysis...')

// Single comprehensive analysis instead of multiple agents
const analysis = await agent(
`Analyze this portfolio and provide PRACTICAL recommendations.

**Portfolio (${portfolio.projects.length} projects):**
\`\`\`json
${JSON.stringify(portfolio.projects, null, 2)}
\`\`\`

**Analyze these 6 perspectives:**

1. **Portfolio Rationalization** - What to kill/merge/pause?
   - Abandoned projects
   - Projects solving same problem (duplication)
   - Low ROI vs effort

2. **Strategic Focus** - Where to concentrate?
   - Best market opportunities
   - Your competitive advantages
   - Underserved areas

3. **Technical Efficiency** - Reduce fragmentation
   - Standardize tech stack where possible
   - Consolidate similar projects
   - Pay down highest-debt projects

4. **Monetization Pipeline** - What could make money?
   - Most viable business models
   - Readiness level
   - Next steps

5. **Career Strategy** - How to position portfolio?
   - Strongest projects to highlight
   - Gaps to fill
   - Narrative arc

6. **Bandwidth Optimization** - Ideal project count?
   - Current state (overextended/focused)
   - Ideal active projects
   - Time allocation (% new features, maintenance, learning, marketing)

**Return JSON:**
{
  "quick_wins": [
    {"action": "<specific doable action>", "impact": "high/medium/low", "effort": "<hours>"}
  ],
  "actions": {
    "kill": [{"project": "<name>", "reason": "<why>", "save": "<% effort>"}],
    "merge": [{"projects": ["<name1>", "<name2>"], "approach": "<how>"}],
    "pause": [{"project": "<name>", "duration": "<months>", "trigger": "<what restarts>"}],
    "double_down": [{"project": "<name>", "why": "<market/technical>", "next": "<milestone>"}]
  },
  "strategic_focus": {
    "primary": "<what to focus next 6 months>",
    "secondary": "<what to focus next 12 months>",
    "exit_domains": [<domains to exit>],
    "enter_domains": [<domains to enter>]
  },
  "technical": {
    "standardize": "<what to consolidate>",
    "paydown_debt": [{"project": "<name>", "debt": "<what needs fixing>", "priority": "high/medium/low"}],
    "learn": [{"tech": "<what>", "why": "<why>"}]
  },
  "monetization": [
    {"project": "<name>", "model": "<subscription/one-time/etc>", "readiness": "<%>", "next": "<action>"}
  ],
  "career": {
    "narrative": "<how to position portfolio>",
    "highlight": [<project names>],
    "gaps": [<what's missing>]
  },
  "bandwidth": {
    "current": "<overextended/focused>",
    "ideal_active": <number>,
    "allocation": {"new_features": "<%>", "maintenance": "<%>", "learning": "<%>", "marketing": "<%>"}
  }
}

**Be SPECIFIC and ACTIONABLE.** Every recommendation should be implementable this week.`,
  { label: 'Strategic analysis', phase: 'Analysis', isolation: 'worktree' }
)

let parsed
if (typeof analysis === 'string') {
  try {
    const jsonStart = analysis.indexOf('{')
    const jsonEnd = analysis.lastIndexOf('}')
    parsed = JSON.parse(analysis.substring(jsonStart, jsonEnd + 1).replace(/```json\n?/g, '').replace(/```\n?/g, '').trim())
  } catch (e) {
    log(`Parse error: ${e.message}`)
    parsed = { error: 'Parse failed' }
  }
} else {
  parsed = analysis || {}
}

phase('Documentation')

log('Generating markdown report...')

// Generate markdown report
const report = `# Portfolio Strategic Analysis

**Projects Analyzed:** ${portfolio.projects.length}

---

## 🎯 Quick Wins (This Week)

${parsed.quick_wins?.map(w => `- **${w.action}** [Impact: ${w.impact}, Effort: ${w.effort}]`).join('\n') || 'None identified'}

---

## 📊 Portfolio Actions

### ❌ Kill (${parsed.actions?.kill?.length || 0})
${parsed.actions?.kill?.map(k => `- **${k.project}** - ${k.reason} (Save ${k.save} effort)`).join('\n') || 'None'}

### 🔀 Merge (${parsed.actions?.merge?.length || 0})
${parsed.actions?.merge?.map(m => `- **${m.projects.join(' + ')}** → ${m.approach}`).join('\n') || 'None'}

### ⏸️  Pause (${parsed.actions?.pause?.length || 0})
${parsed.actions?.pause?.map(p => `- **${p.project}** for ${p.duration} (restart: ${p.trigger})`).join('\n') || 'None'}

### ⚡ Double Down (${parsed.actions?.double_down?.length || 0})
${parsed.actions?.double_down?.map(d => `- **${d.project}** - ${d.why} (Next: ${d.next})`).join('\n') || 'None'}

---

## 🎯 Strategic Focus

- **Next 6 months:** ${parsed.strategic_focus?.primary || 'Not defined'}
- **Next 12 months:** ${parsed.strategic_focus?.secondary || 'Not defined'}
- **Exit domains:** ${parsed.strategic_focus?.exit_domains?.join(', ') || 'None'}
- **Enter domains:** ${parsed.strategic_focus?.enter_domains?.join(', ') || 'None'}

---

## 🔧 Technical Recommendations

### Standardize
- ${parsed.technical?.standardize || 'No consolidation needed'}

### Pay Down Debt
${parsed.technical?.paydown_debt?.map(d => `- **${d.project}** - ${d.debt} [${d.priority}]`).join('\n') || 'None'}

### Learn
${parsed.technical?.learn?.map(l => `- **${l.tech}** - ${l.why}`).join('\n') || 'None'}

---

## 💰 Monetization Pipeline

${parsed.monetization?.map(m => `- **${m.project}** - ${m.model} (${m.readiness} ready) - Next: ${m.next}`).join('\n') || 'No monetization opportunities identified'}

---

## 💼 Career Strategy

**Portfolio Narrative:**
${parsed.career?.narrative || 'Not defined'}

**Highlight Projects:**
${parsed.career?.highlight?.map(h => `- ${h}`).join('\n') || 'None'}

**Gaps to Fill:**
${parsed.career?.gaps?.map(g => `- ${g}`).join('\n') || 'None'}

---

## ⏰  Personal Bandwidth

- **Current state:** ${parsed.bandwidth?.current || 'unknown'}
- **Ideal active projects:** ${parsed.bandwidth?.ideal_active || 'unknown'}

**Time allocation:**
- New features: ${parsed.bandwidth?.allocation?.new_features || 'N/A'}
- Maintenance: ${parsed.bandwidth?.allocation?.maintenance || 'N/A'}
- Learning: ${parsed.bandwidth?.allocation?.learning || 'N/A'}
- Marketing: ${parsed.bandwidth?.allocation?.marketing || 'N/A'}

---

## 📋 Summary

| Metric | Count |
|--------|-------|
| Quick wins | ${parsed.quick_wins?.length || 0} |
| Projects to kill | ${parsed.actions?.kill?.length || 0} |
| Projects to merge | ${parsed.actions?.merge?.length || 0} |
| Projects to pause | ${parsed.actions?.pause?.length || 0} |
| Projects to double down | ${parsed.actions?.double_down?.length || 0} |
| Monetization opportunities | ${parsed.monetization?.length || 0} |
`

// Save to file
const reportPath = '/Users/nerddevsltd/Projects/portfolio-tool/analysis-docs/strategic-analysis.md'

log(`\n=== ANALYSIS COMPLETE ===`)
log(`Report saved to: analysis-docs/strategic-analysis.md`)

return {
  report,
  parsed,
  reportPath
}
