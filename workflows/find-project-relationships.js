export const meta = {
  name: 'find-project-relationships',
  description: 'Two-phase relationship detection: guess candidates, then confirm',
  phases: [
    { title: 'Candidate Discovery', detail: 'Quick pass to find potential relationships' },
    { title: 'Confirmation', detail: 'Deep analysis to confirm candidates' },
    { title: 'Storage', detail: 'Store confirmed relationships' }
  ]
}

phase('Candidate Discovery')

log('Finding potential relationship candidates...')

const candidates = await agent(
`You are finding POTENTIAL relationship candidates between projects using their analysis summaries.

**Your Task:**

1. Call mcp__portfolio__listProjects to get all projects
2. For each project, call mcp__portfolio__getAnalysis to get summary/overview
3. **SKIP projects where analysis contains "not a real project" or similar**
4. Analyze summaries to guess potential relationships

**Return JSON array of CANDIDATE relationships only:**
[
  {
    "source_project": "<id>",
    "target_project": "<id>",
    "suspected_type": "<Similar|Evolution|Shared Feature|Shared Technology|Reuses Component>",
    "reason": "<brief reason based on summaries>"
  }
]

**Important:**
- Skip projects where summary/purpose is "This is not a real project. So we won't analyze it"
- Skip projects with no/empty analysis
- Only include real projects with actual analysis

**Look for signals in summaries:**
- Similar purpose/domain (from "purpose" field)
- Same architecture patterns
- Shared technologies mentioned
- Related feature sets
- Frontend-backend pairs
- Evolution/successor relationships

**Be selective** - only return 15-25 strong candidates max.
Quality over quantity - only relationships with strong signals.`,
{ label: 'Guess candidates from summaries', phase: 'Candidate Discovery' }
)

if (!candidates || candidates.length === 0) {
  log('No candidates found')
  return { relationships_stored: 0 }
}

// Parse candidates - handle string or array
let parsedCandidates = candidates
if (typeof candidates === 'string') {
  const clean = candidates.replace(/```json\n?/g, '').replace(/```\n?/g, '').trim()
  parsedCandidates = JSON.parse(clean)
}

log(`Found ${parsedCandidates.length} candidate relationships to confirm`)

phase('Confirmation')

log('Confirming candidates with deeper analysis...')

const batchSize = 5
const confirmedRelationships = []

for (let i = 0; i < parsedCandidates.length; i += batchSize) {
  const batch = parsedCandidates.slice(i, i + batchSize)
  log(`Confirming batch ${Math.floor(i/batchSize) + 1}/${Math.ceil(parsedCandidates.length/batchSize)}...`)

  const batchResults = await parallel(
    batch.map(candidate => () => agent(
`Confirm this relationship candidate:

**Source:** ${candidate.source_project}
**Target:** ${candidate.target_project}
**Suspected Type:** ${candidate.suspected_type}
**Reason:** ${candidate.reason}

**Your Task:**

1. Call mcp__portfolio__getProject for both projects
2. Call mcp__portfolio__listProjectTechnologies for both projects
3. Call mcp__portfolio__listFeatures for both projects
4. Analyze ACTUAL code/files to confirm if relationship exists

**Return JSON if confirmed:**
{
  "source_project": "${candidate.source_project}",
  "target_project": "${candidate.target_project}",
  "type": "${candidate.suspected_type}",
  "description": "<detailed description of confirmed relationship>",
  "confidence": 0.0-1.0
}

**Return null if relationship is NOT confirmed.**

Only confirm if:
- Clear evidence of relationship
- Not just coincidence
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

// Parse confirmed relationships - agents return JSON strings
const parsedRelationships = confirmedRelationships
  .map(rel => {
    if (typeof rel === 'string') {
      try {
        // Handle markdown code blocks
        const clean = rel.replace(/```json\n?/g, '').replace(/```\n?/g, '').trim()
        return JSON.parse(clean)
      } catch (e) {
        return null
      }
    }
    // Skip null values and already-parsed objects (shouldn't happen)
    return rel === null ? null : rel
  })
  .filter(Boolean)

log(`\nConfirmed ${confirmedRelationships.length}/${parsedCandidates.length} relationships`)

phase('Storage')

if (parsedRelationships.length === 0) {
  log('No confirmed relationships to store')
  return { relationships_stored: 0, candidates_found: parsedCandidates.length }
}

log(`Storing ${parsedRelationships.length} confirmed relationships in single batch...`)

// Store all relationships in ONE agent call to avoid wasting tokens
const relationshipsToStore = parsedRelationships.map(rel => ({
  source_project: rel.source_project,
  target_project: rel.target_project,
  type: rel.type,
  description: rel.description,
  confidence: rel.confidence
}))

const storeResult = await agent(
`Store these ${parsedRelationships.length} relationships using mcp__portfolio__storeRelationship.

**IMPORTANT:** Call mcp__portfolio__storeRelationship EXACTLY ${parsedRelationships.length} times with these EXACT parameters. DO NOT paraphrase or change any values. DO NOT combine relationships. Store EACH relationship INDIVIDUALLY.

${parsedRelationships.map((rel, i) => `
**Relationship ${i + 1}:**
- source_project: "${rel.source_project}"
- target_project: "${rel.target_project}"
- type: "${rel.type}"
- description: "${rel.description}"
- confidence: ${rel.confidence}
`).join('\n')}

**Your task:**
1. For EACH relationship above, call mcp__portfolio__storeRelationship with the EXACT parameters listed
2. Return a JSON array of booleans: [true, true, false, ...] indicating success/failure for each
3. DO NOT skip any relationships
4. DO NOT change any parameter values
5. DO NOT paraphrase descriptions

Return ONLY the JSON array of ${parsedRelationships.length} boolean results.`,
  { label: 'Store all relationships', phase: 'Storage' }
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

log(`\n=== Summary ===`)
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
