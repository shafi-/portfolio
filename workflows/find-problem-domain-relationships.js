export const meta = {
  name: 'find-problem-domain-relationships',
  description: 'Find relationships based on solving the same problem, not shared technology',
  phases: [
    { title: 'Problem Discovery', detail: 'Find projects solving the same problem' },
    { title: 'Confirmation', detail: 'Verify with analysis' },
    { title: 'Storage', detail: 'Store confirmed relationships' }
  ]
}

phase('Problem Discovery')

log('Finding projects that solve the SAME PROBLEM...')

const candidates = await agent(
`You are finding projects that SOLVE THE SAME PROBLEM.

**Your Task:**

1. Call mcp__portfolio__listProjects to get all projects
2. For each project, call mcp__portfolio__getAnalysis to get analysis
3. **SKIP projects where analysis contains "not a real project" or is empty**

**Focus ONLY on:**
- SAME CORE PROBLEM being solved
- SAME TARGET AUDIENCE
- SAME USER INTENT/GOAL
- Frontend-backend pairs for SAME system
- Evolution/successor projects solving SAME base problem

**DO NOT focus on:**
- Shared technology (Next.js, Supabase, etc.)
- Similar architecture patterns
- Same country/region unless same problem
- Same framework unless solving same problem

**Valid relationship types:**
- "Shared Problem" - Same core problem, potentially different approaches
- "Frontend-Backend" - UI + API for same system
- "Evolution" - Version 2 solving same core problem
- "Complementary" - Solve different parts of same larger problem

**Return JSON array of CANDIDATES only:**
[
  {
    "source_project": "<id>",
    "target_project": "<id>",
    "suspected_type": "<Shared Problem|Frontend-Backend|Evolution|Complementary>",
    "reason": "<brief explanation of what SHARED PROBLEM they both solve>"
  }
]

**Be selective** - only 10-20 strong candidates.
Quality over quantity.`,
  { label: 'Find same-problem relationships', phase: 'Problem Discovery' }
)

if (!candidates || candidates.length === 0) {
  log('No candidates found')
  return { relationships_stored: 0 }
}

// Parse candidates - handle text before JSON, markdown code blocks
let parsedCandidates = candidates
if (typeof candidates === 'string') {
  // Find the JSON array - look for '[' and take everything from there to the last ']'
  const jsonStart = candidates.indexOf('[')
  const jsonEnd = candidates.lastIndexOf(']')
  if (jsonStart !== -1 && jsonEnd !== -1) {
    const jsonStr = candidates.substring(jsonStart, jsonEnd + 1)
    // Also remove any markdown code blocks
    const clean = jsonStr.replace(/```json\n?/g, '').replace(/```\n?/g, '').trim()
    parsedCandidates = JSON.parse(clean)
  } else {
    throw new Error('Could not find JSON array in response')
  }
}

log(`Found ${parsedCandidates.length} same-problem candidates to confirm`)

phase('Confirmation')

log('Confirming same-problem relationships with analysis...')

const batchSize = 5
const confirmedRelationships = []

for (let i = 0; i < parsedCandidates.length; i += batchSize) {
  const batch = parsedCandidates.slice(i, i + batchSize)
  log(`Confirming batch ${Math.floor(i/batchSize) + 1}/${Math.ceil(parsedCandidates.length/batchSize)}...`)

  const batchResults = await parallel(
    batch.map(candidate => () => agent(
`Confirm this SAME-PROBLEM relationship:

**Source:** ${candidate.source_project}
**Target:** ${candidate.target_project}
**Suspected Type:** ${candidate.suspected_type}
**Reason:** ${candidate.reason}

**Your Task:**

1. Call mcp__portfolio__getProject for both projects
2. Call mcp__portfolio__listFeatures for both projects
3. Call mcp__portfolio__getAnalysis for both projects
4. Confirm they SOLVE THE SAME CORE PROBLEM

**What constitutes SAME PROBLEM:**
- Same user need/intent
- Same target audience
- Same use case/job-to-be-done
- Core functionality overlap > 50%
- Direct competitive or complementary relationship

**What is NOT same problem:**
- Same technology stack
- Same architecture pattern
- Same country/region only
- Same framework only

**Return JSON if SAME PROBLEM:**
{
  "source_project": "${candidate.source_project}",
  "target_project": "${candidate.target_project}",
  "type": "${candidate.suspected_type}",
  "description": "<clear explanation of WHAT SHARED PROBLEM they both solve>",
  "confidence": 0.0-1.0
}

**Return null if NOT same problem.**

Only confirm if:
- Clear evidence of same core problem
- Not just technology overlap
- Confidence ≥ 0.7`,
      {
        label: `Confirm ${candidate.source_project} → ${candidate.target_project}`,
        phase: 'Confirmation',
        isolation: 'worktree'
      }
    ))
  )

  const validResults = batchResults.filter(Boolean)
  confirmedRelationships.push(...validResults)
  log(`  ✓ Batch confirmed: ${validResults.length}/${batch.length}`)
}

log(`\nConfirmed ${confirmedRelationships.length}/${parsedCandidates.length} same-problem relationships`)

// Parse confirmed relationships
const parsedRelationships = confirmedRelationships
  .map(rel => {
    if (typeof rel === 'string') {
      try {
        const clean = rel.replace(/```json\n?/g, '').replace(/```\n?/g, '').trim()
        return JSON.parse(clean)
      } catch (e) {
        return null
      }
    }
    return rel === null ? null : rel
  })
  .filter(Boolean)

phase('Storage')

if (parsedRelationships.length === 0) {
  log('No same-problem relationships to store')
  return { relationships_stored: 0, candidates_found: parsedCandidates.length }
}

log(`Storing ${parsedRelationships.length} same-problem relationships in single batch...`)

const storeResult = await agent(
`Store these ${parsedRelationships.length} SAME-PROBLEM relationships using mcp__portfolio__storeRelationship.

**IMPORTANT:** Call mcp__portfolio__storeRelationship EXACTLY ${parsedRelationships.length} times with these EXACT parameters. DO NOT paraphrase or change any values.

${parsedRelationships.map((rel, i) => `
**Relationship ${i + 1}:**
- source_project: "${rel.source_project}"
- target_project: "${rel.target_project}"
- type: "${rel.type}"
- description: "${rel.description}"
- confidence: ${rel.confidence}
`).join('\n')}

**Your task:**
1. For EACH relationship, call mcp__portfolio__storeRelationship with EXACT parameters
2. Return JSON array of booleans: [true, true, false, ...]
3. DO NOT skip any relationships
4. DO NOT change any parameter values

Return ONLY the JSON array of ${parsedRelationships.length} boolean results.`,
  { label: 'Store all same-problem relationships', phase: 'Storage' }
)

// Parse the result
let storedCount = 0
const errors = []

let resultsArray
if (typeof storeResult === 'string') {
  try {
    const clean = storeResult.replace(/```json\n?/g, '').replace(/```\n?/g, '').trim()
    resultsArray = JSON.parse(clean)
  } catch (e) {
    log(`  ✗ Failed to parse storage results: ${e.message}`)
    errors.push({ error: `Parse error: ${e.message}` })
  }
} else if (Array.isArray(storeResult)) {
  resultsArray = storeResult
}

if (resultsArray) {
  resultsArray.forEach((success, index) => {
    if (success === true || success === 'true') {
      storedCount++
      log(`  ✓ ${parsedRelationships[index].type}: ${parsedRelationships[index].source_project} → ${parsedRelationships[index].target_project}`)
    } else {
      errors.push({ rel: parsedRelationships[index], error: 'Storage failed' })
      log(`  ✗ Failed: ${parsedRelationships[index].source_project} → ${parsedRelationships[index].target_project}`)
    }
  })
}

log(`\n=== Same-Problem Relationships Summary ===`)
log(`Candidates: ${parsedCandidates.length}`)
log(`Confirmed: ${parsedRelationships.length}`)
log(`Stored: ${storedCount}`)
log(`Errors: ${errors.length}`)

if (errors.length > 0) {
  log('\nFailed to store:')
  errors.forEach(({ rel, error }) => {
    log(`  - ${rel.source_project} → ${rel.target_project}: ${error}`)
  })
}

return {
  candidates_found: parsedCandidates.length,
  confirmed: parsedRelationships.length,
  stored: storedCount,
  errors: errors.length,
  error_details: errors
}
