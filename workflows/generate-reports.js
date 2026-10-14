#!/usr/bin/env node

const fs = require('fs');
const path = require('path');

// Read the viability result
const viabilityPath = '/Users/nerddevsltd/.claude/projects/-Users-nerddevsltd-Projects-portfolio-tool/6f57df14-e706-4f24-948a-dfd5aff5bdff/viability-result.json';
const viabilityData = JSON.parse(fs.readFileSync(viabilityPath, 'utf8'));

const recommendations = viabilityData.recommendations || [];

// Group by recommendation
const toContinue = recommendations.filter(r => r.recommendation === 'CONTINUE');
const toPause = recommendations.filter(r => r.recommendation === 'PAUSE');
const toDrop = recommendations.filter(r => r.recommendation === 'DROP');

// Sort by viability score
toContinue.sort((a, b) => b.viability_score - a.viability_score);

// Generate markdown report
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

${toContinue.filter(r => r.priority === 'HIGH').slice(0, 5).map(r => `
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
`).join('\n---\n')}

---

## 🔄 Medium Priority - CONTINUE

${toContinue.filter(r => r.priority === 'MEDIUM').slice(0, 5).map(r => `
### ${r.name} [${r.viability_score}/100]

**Purpose:** ${r.purpose}
**Rationale:** ${r.rationale}

**Market:** ${r.market_size} | ${r.competition} competition | ${r.uniqueness} uniqueness
**Monetization:** ${r.monetization} | Opportunity: ${r.opportunity_score}/10
**Status:** ${r.maturity}
`).join('\n---\n')}

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

## 📋 Top 20 Projects Ranked

| Rank | Project | Viability | Market | Competition | Uniqueness | Recommendation |
|------|---------|-----------|--------|-------------|------------|----------------|
${recommendations.slice(0, 20).map((r, i) => `| ${i + 1} | ${r.name} | ${r.viability_score}/100 | ${r.market_size} | ${r.competition} | ${r.uniqueness} | ${r.recommendation} |`).join('\n')}

---

## 🔍 Insights

### Strongest Opportunities (Top 5)
${recommendations.slice(0, 5).map((r, i) => `${i + 1}. **${r.name}** - ${r.market_size} market, ${r.competition} competition, ${r.opportunity_score}/10 opportunity (${r.viability_score}/100 viability)`).join('\n')}

### High Uniqueness Projects
${recommendations.filter(r => r.uniqueness === 'high' || r.uniqueness === 'very-high' || r.uniqueness === 'medium-high').slice(0, 10).map(r => `- **${r.name}** - ${r.uniqueness} uniqueness in ${r.market_size} market (${r.viability_score}/100)`).join('\n')}

### Monetization Ready
${recommendations.filter(r => r.monetization === 'high').slice(0, 10).map(r => `- **${r.name}** - ${r.monetization} monetization potential (${r.viability_score}/100)`).join('\n')}

### Low Competition Gems
${recommendations.filter(r => r.competition === 'none' || r.competition === 'low').slice(0, 10).map(r => `- **${r.name}** - ${r.competition} competition in ${r.market_size} market (${r.viability_score}/100)`).join('\n')}

---

## 🎯 Strategic Recommendations

### Immediate Actions (This Week)
- Focus resources on top ${Math.min(5, toContinue.filter(r => r.priority === 'HIGH').length)} high-viability projects
- Pause or re-evaluate ${toPause.length} projects for validation
- Consider dropping ${toDrop.length} projects with poor market fit

### Next 6 Months Focus
- Top 3 high-viability projects:
${recommendations.slice(0, 3).map((r, i) => `${i + 1}. **${r.name}** (${r.viability_score}/100)`).join('\n')}

### Portfolio Optimization
- **Consolidate duplicate solutions** - Multiple projects solving same problem
- **Standardize tech stack** - Reduce fragmentation
- **Double down on winners** - High-viability + high-uniqueness projects
`;

// Save report
const reportPath = '/Users/nerddevsltd/Projects/portfolio-tool/analysis-docs/viability-analysis.md';
fs.writeFileSync(reportPath, report);
console.log(`✅ Viability report saved to: ${reportPath}`);

// Summary
console.log('\n=== SUMMARY ===');
console.log(`High Priority: ${toContinue.filter(r => r.priority === 'HIGH').length}`);
console.log(`Medium Priority: ${toContinue.filter(r => r.priority === 'MEDIUM').length}`);
console.log(`Pause: ${toPause.length}`);
console.log(`Drop: ${toDrop.length}`);
