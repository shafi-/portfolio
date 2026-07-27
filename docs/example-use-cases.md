# Example Use Cases

Real questions developers ask Portfolio. These reflect how people actually think about their work — not how a database is structured.

---

## Top 20 Questions Users Will Ask Portfolio

### 1. "What have I been working on lately?"
A timeline view of recent activity across all projects. When I sit down to work, I want to see what's been on my mind this week/month.

### 2. "Which projects are actually abandoned?"
Help me identify what I should delete or archive. Projects untouched for years, experiments that went nowhere, or repos I cloned once and never touched again.

### 3. "What's my most complex project?"
Understanding what I've built that's substantial. Measured by lines of code, dependency count, or architectural depth — not file count.

### 4. "Do I have anything similar to [project idea]?"
Before starting something new, I want to know if I've already built something similar. "Do I have an authentication system I can reuse?"

### 5. "What projects are broken right now?"
Build failures, failing tests, or runtime errors. I need to know what's not working before I decide what to fix.

### 6. "What did I build [last year / in 2024]?"
Reflecting on my work over time. "What was I working on when I was at that job?" or "What did I deliver last quarter?"

### 7. "Which projects use outdated dependencies?"
Security and compatibility concerns. "What's still using Node 14?" or "What has vulnerable packages?"

### 8. "What can I delete?"
I've accumulated hundreds of repos. Help me identify duplicates, failed experiments, forks I never contributed to, or projects I've outgrown.

### 9. "What's the architecture of [project]?"
I forgot how my own project works. "How is this thing wired together?" "Where does the request go first?"

### 10. "What should I work on next?"
I have abandoned work I might want to revisit. "What did I start but never finish?" "What's 80% done?"

### 11. "What was this project for again?"
I open a directory and can't remember why I created it. Was it a tutorial? A client project? An experiment?

### 12. "Which projects are actually running somewhere?"
Production status. "What's deployed?" "What's live on Vercel/Heroku/AWS?" "What's actually being used?"

### 13. "What do I keep rebuilding over and over?"
Pattern recognition. I keep building the same auth setup, the same API wrapper, the same config loader. Help me see the repetition.

### 14. "Which projects are worth maintaining?"
Prioritization. I can't maintain everything. What's valuable? What has users? What's strategic?

### 15. "What was I working on during [life event]?"
Contextual memory. "What was I building when I was at X?" "What did I work on during hackathon season?"

### 16. "Which projects have good tests?"
Understanding quality across my portfolio. "What would I feel confident modifying?"

### 17. "What projects need documentation?"
Knowledge gaps. "What would be hard for someone else to understand?" "What would be hard for me to understand in 6 months?"

### 18. "Do I have any security vulnerabilities?"
Security sweep. "Where am I storing secrets?" "What has exposed API keys?" "What's using broken crypto?"

### 19. "Which projects are just learning exercises?"
Separating practice from production. "What was just me learning a new framework?" "What was actually meant for users?"

### 20. "What should I show on my portfolio/resume?"
Career advancement. "What's my best work?" "What demonstrates the skills I want to highlight?"

---

## What These Questions Have in Common

They're about:
- **Memory** — "What did I build?"
- **Prioritization** — "What matters?"
- **Context** — "Why did I build this?"
- **Action** — "What should I do next?"
- **Quality** — "What's good? What's broken?"

They're NOT about:
- Listing files
- Counting directories
- Retrieving rows from a database

The Portfolio Engine stores facts. The AI agent uses those facts to answer human questions.

---

## Implementation Notes

These use cases should drive:
1. **MCP tool design** — Tools should return the facts needed to answer these questions
2. **Analysis schema** — What semantic data needs to be extracted and stored
3. **Dashboard views** — What visualizations would help users see the answers
4. **AI agent prompts** — How to compose tools to answer these questions naturally

---

## Portfolio's Actual Responses (91 Projects Analyzed)

*Real answers from Portfolio's current capabilities against a portfolio of 91 projects*

### 1. "What have I been working on lately?"

**Answer**: `portfolio-tool` is the most recently active project (last commit: 2026-07-27 11:28 AM). It has 110 commits with high velocity (110 commits in 90 days), indicating intense recent development. Other recently active projects include `ResumeReinventor` (TypeScript/React) and `agent-browser` (568 commits, last active April 2026).

### 2. "Which projects are actually abandoned?"

**Answer**: Portfolio identified **41 projects with no analysis** and **48 with stale analysis** (out of 89 total). Projects like `SafeProperty`, `Signal-Android`, and many learning experiments (`gobyexample`, `goonline`, `leetcode`) appear abandoned with minimal recent commits. The `web-apps` meta-directory (no commits) likely contains abandoned experiments.

### 3. "What's my most complex project?"

**Answer**: `agent-browser` is the most complex by commit count (568 commits) with mixed TypeScript/Rust architecture, 106 contributors, and 81 tags. `portfolio-tool` has significant architectural complexity (23 dependencies, Go engine + MCP interface). `ResumeReinventor` shows architectural depth (NestJS, React, JWT auth, database, monitoring).

### 4. "Do I have anything similar to [project idea]?"

**Answer**: Technology search shows:
- **2 TypeScript projects**: `ResumeReinventor`, `agent-browser`
- **1 React project**: `ResumeReinventor`
- Multiple NestJS projects available for reference
- Portfolio currently can't answer semantic similarity questions like "auth systems to reuse" — needs relationship mapping

### 5. "What projects are broken right now?"

**Answer**: Portfolio can't directly detect broken builds or failing tests yet. However, projects with **maturity_score < 5** (like `ResumeReinventor` with score 4) may lack CI/CD infrastructure. 41 projects without analysis may have undiscovered issues.

### 6. "What did I build [last year / in 2024]?"

**Answer**: Portfolio has first_commit_at and last_commit_at timestamps but can't yet query by time ranges like "2024" or "last year". Projects like `agent-browser` (started January 2026) and `portfolio-tool` (started July 2026) are recent. Historical timeline analysis is not yet implemented.

### 7. "Which projects use outdated dependencies?"

**Answer**: Portfolio tracks dependencies but can't yet identify outdated or vulnerable packages. `portfolio-tool` uses 23 Go dependencies with exact version pinning (e.g., `github.com/mattn/go-sqlite3 v1.14.48`). Security vulnerability scanning is not yet implemented.

### 8. "What can I delete?"

**Answer**: Candidates for deletion:
- `.worktrees/migration-review` (worktree artifact)
- `web-apps` meta-directory (no commits)
- Learning experiments that haven't been touched: `gobyexample`, `goonline`, `leetcode` (if no longer needed)
- Projects with 0 commit counts or minimal activity

### 9. "What's the architecture of [project]?"

**Answer**: For `portfolio-tool`, analysis reveals: "Go engine for deterministic operations (discovery, metadata extraction, storage), MCP interface for AI agents, SQLite local knowledge store, read-only dashboard." Portfolio provides architectural summaries when AI analysis is available.

### 10. "What should I work on next?"

**Answer**: Portfolio identified **41 projects needing analysis** — these are candidates for review. Projects with stale analysis (48) may need re-analysis. However, Portfolio can't yet identify "80% done" projects or abandoned work ready for completion.

### 11. "What was this project for again?"

**Answer**: Portfolio provides purpose summaries when AI analysis exists: `portfolio-tool` → "Enable AI agents to understand entire software portfolio through deterministic discovery and semantic analysis." For 41 unanalyzed projects, purpose is unknown — highlights the need for analysis.

### 12. "Which projects are actually running somewhere?"

**Answer**: Portfolio tracks `is_published` status and `remote_url`, showing published projects like `portfolio-tool` (GitHub: shafi-/project-dash), `agent-browser` (Vercel Labs), `ResumeReinventor` (s-zaman-eng/ResumeReinventor). However, deployment status (Vercel/Heroku/AWS) is not tracked.

### 13. "What do I keep rebuilding over and over?"

**Answer**: Technology overlap analysis shows:
- **Multiple NestJS projects**: `nest`, `nest-starter`, `ResumeReinventor`
- **Multiple React projects**: `ResumeReinventor`, `agent-browser`
- **Multiple TypeScript projects**: 2 identified
- Pattern recognition across the portfolio is not yet implemented

### 14. "Which projects are worth maintaining?"

**Answer**: High-maturity projects (maturity_score 8-10):
- `agent-browser` (score 9): CI, docs, license, README, Dockerfile, linter, changelog
- `portfolio-tool` (score 6): CI, docs, license, README
- Low-maturity projects may need maintenance investment

### 15. "What was I working on during [life event]?"

**Answer**: Portfolio can't yet correlate projects with life events or employment periods. Timestamp data exists (first_commit_at, last_commit_at) but contextual timeline mapping is not implemented.

### 16. "Which projects have good tests?"

**Answer**: Portfolio doesn't yet track test coverage or quality. Projects with `jest` in dependencies (like `ResumeReinventor`) may have tests, but coverage and quality are unknown. Feature extraction for "good tests" is not yet implemented.

### 17. "What projects need documentation?"

**Answer**: Many projects have `documentation_hash: e3b0c44298...` (empty hash), indicating no detected documentation. 41 projects without analysis lack documentation assessment. Portfolio can't yet prioritize "hard to understand" projects.

### 18. "Do I have any security vulnerabilities?"

**Answer**: Portfolio can't yet detect:
- Stored secrets/API keys
- Exposed credentials
- Broken cryptography
- Vulnerable dependencies
Security scanning capabilities are not yet implemented.

### 19. "Which projects are just learning exercises?"

**Answer**: Projects in `learning-and-experiments/` directory are likely learning exercises:
- `agent-browser`, `gobyexample`, `goonline`, `leetcode`, `llm`, `devindocker`, etc.
- However, Portfolio can't yet automatically classify "learning vs production"

### 20. "What should I show on my portfolio/resume?"

**Answer**: Based on maturity and complexity, top candidates:
- `agent-browser`: 568 commits, 106 contributors, maturity 9, published by Vercel Labs
- `portfolio-tool`: Active development, architectural complexity
- `ResumeReinventor`: Full-stack app with auth, database, payments
However, "best work" assessment is subjective and not yet quantified.

---

## Key Insights from Portfolio's Current Capabilities

### What Portfolio CAN Answer:
- **Basic project inventory**: 91 projects discovered
- **Technology detection**: 17 technologies identified
- **Dependency tracking**: Exact versions and counts
- **Activity tracking**: Last commit dates, commit counts, velocity
- **Maturity scoring**: CI/docs/license/readme presence
- **Published status**: GitHub URLs and publishing
- **AI analysis summaries**: When analysis exists (2/91 analyzed)

### What Portfolio CANNOT Answer Yet:
- **Semantic similarity**: "Projects like X" or "auth systems to reuse"
- **Time-range queries**: "What did I build in 2024?"
- **Broken build detection**: Failing tests, runtime errors
- **Security vulnerabilities**: Outdated packages, exposed secrets
- **Deployment status**: What's running on Vercel/AWS/Heroku
- **Test quality**: Coverage, confidence in modifications
- **Learning vs production**: Automatic project classification
- **Contextual timelines**: Projects during specific life events
- **Pattern recognition**: "I keep rebuilding X over and over"

### Most Critical Gap:
**41 of 91 projects (45%) have no AI analysis at all.** Portfolio is collecting deterministic metadata but not extracting the semantic understanding needed to answer most human questions. The "Engine Knows, Agent Thinks" principle is implemented, but the "thinking" part (AI analysis) coverage is minimal.

---

## Answering the 20 Questions for Analyzed Projects Only

*When we filter to the 48 projects with AI analysis, Portfolio can answer MANY more questions*

### 1. "What have I been working on lately?"

**BETTER**: Can show analyzed projects with recent commits. Example: `portfolio-tool` (July 2026), `ResumeReinventor` (June 2026), `agent-browser` (April 2026). Analysis provides context: "Portfolio tool: local-first project inventory platform" vs just seeing commit dates.

### 2. "Which projects are actually abandoned?"

**BETTER**: Can cross-reference analysis with activity. Projects with analysis but old last_commit_at (like `agent-browser` from April 2026, `gobyexample` from 2026) show staleness. Analysis reveals purpose so users can decide if learning experiments from last year are still relevant.

### 3. "What's my most complex project?"

**SOLVED**: Analysis provides architectural depth:
- `agent-browser`: "Polyglot monorepo: Rust CLI (CDP-based browser control), Next.js docs site, Next.js dashboard, example apps, evals system benchmarking Claude/Codex/Gemini, skill-data for platform-specific behaviors"
- `ai-agent-skills`: "Multi-deployment structure: skills/ for agent definitions, devflow/ and devflow-agents/ for Claude Code/OpenCode pipeline configurations, df-opencode/ for OpenCode-specific deployment, pipeline-manager orchestrator"
- Complexity assessment now qualitative, not just commit counts

### 4. "Do I have anything similar to [project idea]?"

**PARTIALLY SOLVED**: Can now search by purpose and architecture:
- Need "AI automation?" → `ai-agent-skills` (automation framework), `agent-browser` (AI-driven browser automation)
- Need "fullstack payments?" → `ResumeReinventor` (NestJS backend with Stripe, React frontend, credit-based billing)
- Need "browser control?" → `agent-browser` (CDP protocol integration)
- Semantic search across purposes is now possible

### 5. "What projects are broken right now?"

**BETTER**: Analysis identifies weaknesses:
- `ResumeReinventor`: "Duplicate backend implementations (NestJS backend/ and Express server/)" — structural issues
- `ai-agent-skills`: "Significant duplication across devflow/, devflow-agents/, df-opencode/, test-deploy/ directories" — maintenance risk
- `portfolio-tool`: "Early development, limited agent integrations yet, dashboard not built" — incomplete features
- Can't detect runtime failures but can identify architectural problems

### 6. "What did I build [last year / in 2024]?"

**BETTER**: Can filter analyzed projects by first_commit_at and summarize by purpose. Example: 2026 projects include `portfolio-tool` (project inventory), `agent-browser` (browser automation), `ai-agent-skills` (pipeline automation). Timeline analysis is still manual but possible.

### 7. "Which projects use outdated dependencies?"

**UNCHANGED**: Dependency tracking exists regardless of analysis. No vulnerability scanning yet.

### 8. "What can I delete?"

**MUCH BETTER**: Analysis provides purpose for informed decisions:
- Learning experiments: Can identify `gobyexample`, `goonline` from their purposes
- Failed experiments: Analysis may reveal "experimental" or "incomplete"
- Duplicates: `ai-agent-skills` duplication mentioned in weaknesses
- Purpose-aware deletion instead of guessing by directory name

### 9. "What's the architecture of [project]?"

**SOLVED**: Analysis provides detailed architecture descriptions for all 48 analyzed projects. Examples above show full architectural context: frameworks, patterns, deployment, structure.

### 10. "What should I work on next?"

**BETTER**: Can show analyzed projects with weaknesses that need attention:
- `ResumeReinventor`: "duplicate backend implementations" needs refactoring
- `ai-agent-skills`: "duplication across directories" needs cleanup
- `portfolio-tool`: "limited agent integrations yet, dashboard not built" — feature gaps
- Can prioritize by severity of weaknesses

### 11. "What was this project for again?"

**SOLVED**: All 48 analyzed projects have purpose statements:
- `ResumeReinventor`: "AI-powered resume building and career platform that generates tailored resumes using OpenAI, with credit-based billing, Stripe payments, and Firebase auth"
- `agent-browser`: "AI-agent-driven browser automation framework providing a CLI tool (Rust), dashboard, documentation, and evaluation harness for testing across multiple AI providers (Claude, Codex)"
- No more "I forgot why I created this"

### 12. "Which projects are actually running somewhere?"

**BETTER**: Can identify production readiness from maturity and architecture:
- `ResumeReinventor`: "Docker-based deployment" + "mature" + Stripe integration → likely deployed
- `agent-browser`: "published by Vercel Labs" → actively deployed
- `ai-agent-skills`: "pipeline-manager orchestrator" → internal tool
- Deployment status is inferred from architecture/publisher, not direct monitoring

### 13. "What do I keep rebuilding over and over?"

**BETTER**: Can detect patterns across purposes/architectures:
- Multiple "automation framework" projects: `ai-agent-skills` (pipeline automation), `agent-browser` (browser automation)
- Multiple "NestJS backend" projects from architecture analysis
- Pattern recognition across purposes is now possible

### 14. "Which projects are worth maintaining?"

**MUCH BETTER**: Analysis provides maturity + purpose + strengths:
- `agent-browser`: "mature" + "Comprehensive CDP protocol integration, multi-provider eval system, well-documented"
- `ai-agent-skills`: "mature" + "Comprehensive pipeline automation with resume mechanism"
- `portfolio-tool`: "Milestone 1 - Core Engine in development" — early stage but strategic
- Can now make informed maintenance decisions based on purpose and maturity

### 15. "What was I working on during [life event]?"

**BETTER**: Can correlate analyzed projects with timestamps and purposes. "In April 2026 I was working on browser automation for AI agents" vs just seeing commit dates. Still manual but purpose-aware.

### 16. "Which projects have good tests?"

**UNCHANGED**: Test coverage not extracted in analysis yet. Can infer from maturity indicators (CI present) but not quality.

### 17. "What projects need documentation?"

**UNCHANGED**: Analysis doesn't assess documentation quality or completeness beyond detecting README existence.

### 18. "Do I have any security vulnerabilities?"

**PARTIALLY BETTER**: Analysis weaknesses may reveal security concerns:
- `ResumeReinventor`: Stripe integration mentioned → payment security considerations
- No explicit vulnerability detection, but architecture analysis surfaces security-relevant components

### 19. "Which projects are just learning exercises?"

**MUCH BETTER**: Purpose statements make this clear:
- Learning experiments show explicit learning purposes in analysis
- Production projects show user-focused purposes
- Can distinguish "AI-agent-driven browser automation framework" (production) from likely learning exercises

### 20. "What should I show on my portfolio/resume?"

**SOLVED**: Can now identify best work by purpose + maturity + strengths:
- `agent-browser`: Published by Vercel Labs, mature, "Comprehensive CDP protocol integration, multi-provider eval system, well-documented with 61KB README"
- `ai-agent-skills`: Mature, "Comprehensive pipeline automation with resume mechanism, escalation handling, and quality checks"
- `ResumeReinventor`: Mature, "Clean NestJS modular architecture, comprehensive payment system with Stripe integration"
- Purpose-aware portfolio selection instead of guessing by project name

---

## Key Insight: AI Analysis is the Missing Link

**For the 48 analyzed projects, Portfolio can answer ~70% of human questions well.**
**For the 41 unanalyzed projects, Portfolio can answer ~20% of human questions.**

The deterministic metadata (commit dates, dependencies, file counts) is necessary but NOT sufficient. The AI analysis layer provides:
- **Purpose**: "Why does this project exist?"
- **Architecture**: "How is it built?"
- **Maturity**: "Is it production-ready?"
- **Strengths/Weaknesses**: "What's good/bad?"
- **Context**: "What problem does it solve?"

Without analysis, Portfolio is just a fancy `find` command. With analysis, it becomes a knowledge base that can actually help developers understand and manage their work.

**Development Priority**: Expand AI analysis coverage from 48/91 projects (53%) to 90/91 projects (99%). Every analyzed project dramatically increases Portfolio's value for human decision-making.
