# AI CV Builder — Portfolio Memory + CV Generator MCP Server

Generated: 2026-07-31
Status: DRAFT

---

## Problem Statement

Job seekers forget their own accomplishments. When applying to a new role, they manually reconstruct their history, missing relevant achievements. Existing AI resume builders (Rezi, Teal, Kickresume) are one-off tools — they don't remember you. Every session starts from scratch.

**The insight:** The value isn't "AI builds your CV." The value is "the system remembers everything you ever told it, and deploys the right details at the right time."

---

## What We're Building

Two standalone MCP servers, decoupled from the job application management app:

1. **Portfolio Memory MCP Server** — conversational AI that collects, stores, and retrieves career data
2. **CV Generator MCP Server** — takes portfolio data + job description, produces tailored CV

Any AI agent (Claude, GPT, Gemini, custom agents) can connect to these servers via MCP protocol.

---

## Architecture

### High-Level

```
┌─────────────────────────────────────────────────────┐
│                  AI Agent (any)                      │
│         Claude Desktop / Cursor / Custom Agent       │
└──────────┬──────────────────┬───────────────────────┘
           │ MCP              │ MCP
           ▼                  ▼
┌──────────────────┐  ┌──────────────────┐
│  Portfolio MCP   │  │  CV Generator    │
│  Server          │  │  MCP Server      │
│                  │  │                  │
│  Tools:          │  │  Tools:          │
│  - add_profile   │  │  - generate_cv   │
│  - add_experience│  │  - tailor_to_jd  │
│  - add_achievement│ │  - export_pdf    │
│  - add_skill     │  │  - preview_cv    │
│  - get_portfolio │  │                  │
│  - search_portfolio│ │  Resources:      │
│  - start_interview│ │  - cv_templates  │
│                  │  │  - style_guide   │
│  Resources:      │  │                  │
│  - portfolio     │  └────────┬─────────┘
│  - achievements  │           │
└────────┬─────────┘           │
         │                     │
         ▼                     ▼
┌─────────────────────────────────────┐
│           Supabase (PostgreSQL)      │
│                                     │
│  Tables:                            │
│  - user_portfolios                  │
│  - portfolio_experiences            │
│  - portfolio_achievements           │
│  - portfolio_skills                 │
│  - portfolio_education              │
│  - portfolio_certifications         │
│  - ai_conversations                 │
│  - cv_generated                     │
│  - cv_templates                     │
└─────────────────────────────────────┘
```

### MCP Server 1: Portfolio Memory

**Purpose:** Collect comprehensive career data through guided conversation. Store it structured. Retrieve it intelligently.

#### Tools

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `start_interview` | Begin guided conversation for a section | `section: "experience" \| "education" \| "skills" \| "achievements"` | First question + conversation state |
| `add_experience` | Add work experience | Company, role, dates, description | Confirmation + follow-up questions |
| `add_achievement` | Add achievement to an experience | Experience ID, description, impact, metrics | Confirmation |
| `add_skill` | Add skill with proficiency | Name, level, category, years | Confirmation |
| `add_education` | Add education entry | Institution, degree, dates, details | Confirmation |
| `add_certification` | Add certification | Name, issuer, date, expiry | Confirmation |
| `get_portfolio` | Retrieve full portfolio | User ID | Structured portfolio JSON |
| `search_portfolio` | Search across portfolio | Query string | Matching entries |
| `get_achievements_for_role` | Get achievements relevant to a job role | Job description or role title | Ranked achievements |
| `update_entry` | Update any portfolio entry | Entry ID + fields | Confirmation |
| `delete_entry` | Soft-delete an entry | Entry ID | Confirmation |

#### Resources

| Resource | URI Pattern | Description |
|----------|-------------|-------------|
| `portfolio` | `portfolio://{user_id}` | Full portfolio as structured JSON |
| `achievements` | `portfolio://{user_id}/achievements` | All achievements, ranked by recency |
| `skills` | `portfolio://{user_id}/skills` | Skills matrix with proficiency |
| `conversation_history` | `portfolio://{user_id}/conversations/{session_id}` | Interview session transcript |

#### Prompts

| Prompt | Description |
|--------|-------------|
| `interview_experience` | Guided interview for work experience |
| `interview_achievements` | Deep-dive on achievements for a specific role |
| `interview_skills` | Skills assessment conversation |
| `portfolio_review` | Review and refine existing portfolio |

### MCP Server 2: CV Generator

**Purpose:** Take portfolio data + job description → produce tailored, ATS-optimized CV.

#### Tools

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `generate_cv` | Generate CV from portfolio | `user_id`, `job_description?`, `template?`, `style?` | CV content (markdown) |
| `tailor_to_jd` | Tailor existing CV to specific JD | `cv_id`, `job_description` | Updated CV content |
| `export_pdf` | Convert CV to PDF | `cv_id` or `cv_content` | PDF URL or base64 |
| `preview_cv` | Preview CV without saving | Portfolio data + JD | Rendered preview |
| `get_templates` | List available CV templates | None | Template list |
| `get_style_guide` | Get ATS optimization tips | None | Style guidelines |

#### Resources

| Resource | URI Pattern | Description |
|----------|-------------|-------------|
| `cv_templates` | `cv://templates` | Available CV templates |
| `style_guide` | `cv://style-guide` | ATS optimization rules |
| `generated_cvs` | `cv://{user_id}/cvs` | User's generated CVs |

---

## Database Schema

### user_portfolios
```sql
CREATE TABLE user_portfolios (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  summary TEXT,                          -- AI-generated career summary
  target_roles TEXT[],                   -- Roles user is targeting
  industry_focus TEXT[],                 -- Industries of interest
  preferred_locations TEXT[],            -- Where user wants to work
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_user_portfolios_user_id (user_id)
);

ALTER TABLE user_portfolios ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own portfolio"
  ON user_portfolios FOR ALL
  USING (auth.uid() = user_id);
```

### portfolio_experiences
```sql
CREATE TABLE portfolio_experiences (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  company VARCHAR(255) NOT NULL,
  position VARCHAR(255) NOT NULL,
  location VARCHAR(255),
  start_date DATE NOT NULL,
  end_date DATE,                         -- NULL = current
  employment_type VARCHAR(50) CHECK (employment_type IN ('full-time', 'part-time', 'contract', 'internship', 'freelance')),
  description TEXT,                      -- Role description
  key_responsibilities TEXT[],           -- Main responsibilities
  technologies_used TEXT[],              -- Tech stack at this role
  team_size INT,                         -- Team size managed/worked in
  reporting_to VARCHAR(255),            -- Who they reported to
  is_current BOOLEAN DEFAULT false,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_portfolio_experiences_user_id (user_id),
  INDEX idx_portfolio_experiences_dates (start_date, end_date)
);

ALTER TABLE portfolio_experiences ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own experiences"
  ON portfolio_experiences FOR ALL
  USING (auth.uid() = user_id);
```

### portfolio_achievements
```sql
CREATE TABLE portfolio_achievements (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  experience_id UUID REFERENCES portfolio_experiences(id) ON DELETE CASCADE,
  title VARCHAR(255) NOT NULL,          -- Short title
  description TEXT NOT NULL,             -- What was achieved
  impact TEXT,                           -- Impact description
  metrics JSONB,                         -- Quantified results: {"revenue": "$1M", "reduction": "30%", "users": "10K"}
  skills_used TEXT[],                    -- Skills demonstrated
  category VARCHAR(50) CHECK (category IN ('project', 'leadership', 'technical', 'process', 'revenue', 'other')),
  relevance_score FLOAT DEFAULT 0.5,     -- AI-calculated relevance for common roles
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_portfolio_achievements_user_id (user_id),
  INDEX idx_portfolio_achievements_experience (experience_id),
  INDEX idx_portfolio_achievements_category (category)
);

ALTER TABLE portfolio_achievements ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own achievements"
  ON portfolio_achievements FOR ALL
  USING (auth.uid() = user_id);
```

### portfolio_skills
```sql
CREATE TABLE portfolio_skills (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  name VARCHAR(100) NOT NULL,
  category VARCHAR(50) CHECK (category IN ('technical', 'soft', 'language', 'tool', 'framework', 'other')),
  proficiency VARCHAR(50) CHECK (proficiency IN ('beginner', 'intermediate', 'advanced', 'expert')),
  years_of_experience INT,
  last_used DATE,
  is_highlight BOOLEAN DEFAULT false,    -- Show prominently on CV
  created_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_portfolio_skills_user_id (user_id),
  INDEX idx_portfolio_skills_category (category)
);

ALTER TABLE portfolio_skills ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own skills"
  ON portfolio_skills FOR ALL
  USING (auth.uid() = user_id);
```

### portfolio_education
```sql
CREATE TABLE portfolio_education (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  institution VARCHAR(255) NOT NULL,
  degree VARCHAR(255),
  field_of_study VARCHAR(255),
  start_date DATE,
  end_date DATE,
  gpa DECIMAL(3, 2),
  honors TEXT[],
  relevant_coursework TEXT[],
  created_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_portfolio_education_user_id (user_id)
);

ALTER TABLE portfolio_education ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own education"
  ON portfolio_education FOR ALL
  USING (auth.uid() = user_id);
```

### portfolio_certifications
```sql
CREATE TABLE portfolio_certifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  name VARCHAR(255) NOT NULL,
  issuer VARCHAR(255),
  issue_date DATE,
  expiry_date DATE,
  credential_id VARCHAR(255),
  credential_url TEXT,
  created_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_portfolio_certifications_user_id (user_id)
);

ALTER TABLE portfolio_certifications ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own certifications"
  ON portfolio_certifications FOR ALL
  USING (auth.uid() = user_id);
```

### ai_conversations
```sql
CREATE TABLE ai_conversations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  session_type VARCHAR(50) CHECK (session_type IN ('experience', 'achievements', 'skills', 'review')),
  messages JSONB NOT NULL,               -- [{role, content, timestamp}]
  extracted_data JSONB,                  -- Data extracted from conversation
  status VARCHAR(50) DEFAULT 'active',   -- active, completed, archived
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_ai_conversations_user_id (user_id),
  INDEX idx_ai_conversations_type (session_type)
);

ALTER TABLE ai_conversations ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own conversations"
  ON ai_conversations FOR ALL
  USING (auth.uid() = user_id);
```

### cv_generated
```sql
CREATE TABLE cv_generated (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES auth.users NOT NULL,
  template_id VARCHAR(50),
  job_description TEXT,                  -- JD it was tailored to (if any)
  content JSONB NOT NULL,                -- Structured CV content
  pdf_url TEXT,                          -- Generated PDF URL
  ats_score FLOAT,                       -- ATS optimization score
  tailoring_notes TEXT,                  -- What was emphasized/omitted
  created_at TIMESTAMP DEFAULT NOW(),
  
  INDEX idx_cv_generated_user_id (user_id)
);

ALTER TABLE cv_generated ENABLE ROW LEVEL SECURITY;
CREATE POLICY "Users can only access their own CVs"
  ON cv_generated FOR ALL
  USING (auth.uid() = user_id);
```

---

## Conversation Flow

### Initial Portfolio Build (New User)

```
Agent connects to Portfolio MCP Server
  ↓
Calls start_interview(section: "experience")
  ↓
Server returns: "Tell me about your most recent job. What company, what was your role?"
  ↓
User responds
  ↓
Agent calls add_experience(data)
  ↓
Server stores + returns: "Great. What were your top 2-3 achievements there?"
  ↓
Agent calls add_achievement(data) for each
  ↓
Server asks follow-ups: "How did that impact the team? What was the measurable outcome?"
  ↓
Process repeats for each role
  ↓
Agent calls start_interview(section: "skills")
  ↓
Similar flow for skills, education, certifications
  ↓
Portfolio complete
```

### CV Generation Flow

```
User: "I want to apply to Senior Engineer at Google"
  ↓
Agent calls get_portfolio(user_id)
  ↓
Server returns full portfolio
  ↓
Agent calls generate_cv(user_id, job_description: "Senior Engineer at Google...")
  ↓
CV Generator:
  1. Analyzes JD for requirements
  2. Matches against portfolio skills
  3. Ranks achievements by relevance
  4. Selects most relevant experiences
  5. Generates ATS-optimized content
  6. Returns structured CV
  ↓
Agent presents preview to user
  ↓
User: "Include my Python experience more"
  ↓
Agent calls tailor_to_jd(cv_id, adjustments)
  ↓
Updated CV returned
  ↓
Agent calls export_pdf(cv_id)
  ↓
PDF ready for download
```

---

## Implementation Plan

### Phase 1: Portfolio Memory MCP Server (Week 1-2)

**Goal:** Working MCP server that stores and retrieves portfolio data.

1. **Setup**
   - Create new project directory: `mcp-portfolio-server/`
   - Initialize with `@modelcontextprotocol/sdk`
   - Configure TypeScript, ESM, Zod validation
   - Setup Supabase client

2. **Database**
   - Create all tables via migration
   - Enable RLS on all tables
   - Create indexes

3. **Core Tools**
   - `start_interview` — returns first question + conversation state
   - `add_experience` — validates + stores experience
   - `add_achievement` — validates + stores achievement
   - `add_skill` — validates + stores skill
   - `get_portfolio` — returns full portfolio JSON
   - `search_portfolio` — full-text search across portfolio

4. **Resources**
   - `portfolio://{user_id}` — full portfolio
   - `portfolio://{user_id}/achievements` — achievements only

5. **Testing**
   - Unit tests for each tool
   - Integration test with Supabase
   - MCP Inspector for manual testing

### Phase 2: CV Generator MCP Server (Week 3-4)

**Goal:** Working MCP server that generates tailored CVs.

1. **Setup**
   - Create new project directory: `mcp-cv-server/`
   - Initialize with MCP SDK
   - Setup PDF generation library (Puppeteer or similar)

2. **CV Generation Engine**
   - JD parsing (extract requirements, keywords, skills)
   - Portfolio matching (score achievements against JD)
   - Content generation (LLM-powered bullet points)
   - ATS optimization (formatting, keywords, structure)

3. **Core Tools**
   - `generate_cv` — full generation pipeline
   - `tailor_to_jd` — adjust existing CV
   - `export_pdf` — render to PDF
   - `preview_cv` — preview without saving

4. **Templates**
   - ATS-safe single-column template
   - Professional template
   - Creative template (for non-ATS submission)

5. **Testing**
   - CV generation with sample portfolios
   - ATS score validation
   - PDF rendering tests

### Phase 3: Integration + Polish (Week 5-6)

1. **Job Search App Integration**
   - Connect job search app to MCP servers
   - "Build CV" button in application flow
   - Auto-pull portfolio data when applying

2. **Conversation UI**
   - Chat interface for portfolio building
   - Progress tracking (% complete)
   - Resume interrupted sessions

3. **BYOK Support**
   - Settings page for API key entry
   - Encrypted storage
   - Per-session override option

4. **Free Tier Limits**
   - Profile updates: 2/month
   - CV generations: 5/month (more generous than profile edits)
   - Usage tracking in Supabase

---

## Tech Stack

| Component | Technology | Reason |
|-----------|------------|--------|
| MCP Server | TypeScript + `@modelcontextprotocol/sdk` | Official SDK, type-safe |
| Validation | Zod | Schema validation, MCP-compatible |
| Database | Supabase (PostgreSQL) | Already in stack, RLS, free tier |
| Auth | Supabase Auth | JWT tokens, user isolation |
| PDF Generation | Puppeteer or `@react-pdf/renderer` | Professional PDF output |
| AI (CV generation) | Cloudflare Workers AI + BYOK | Edge, cheap, user's own keys |
| Deployment | Cloudflare Workers or Supabase Edge Functions | Matches existing stack |
| Testing | Vitest + MCP Inspector | Unit + manual testing |

---

## Competitive Advantage

| Feature | Rezi | Teal | Kickresume | **Ours** |
|---------|------|------|------------|----------|
| Portfolio memory | ❌ | ❌ | ❌ | ✅ |
| Conversational build | ❌ | ❌ | ❌ | ✅ |
| MCP protocol | ❌ | ❌ | ❌ | ✅ |
| BYOK | ❌ | ❌ | ❌ | ✅ |
| Agent-agnostic | ❌ | ❌ | ❌ | ✅ |
| ATS optimization | ✅ | ✅ | ✅ | ✅ |
| Templates | ✅ | ✅ | ✅ | ✅ |
| Job tracking | ❌ | ✅ | ❌ | ✅ (separate app) |

**The moat:** Portfolio memory + MCP protocol = reusable by any AI agent. Not just a product, it's infrastructure.

---

## Open Questions

1. **PDF library choice:** Puppeteer (heavy, accurate) vs `@react-pdf/renderer` (lighter, React-native)? Need to test both.
2. **CV generation quality:** How to ensure generated CVs are actually good? Need human evaluation loop.
3. **Rate limiting:** How to handle BYOK users vs platform users? Different limits?
4. **Portfolio portability:** Should users be able to export their portfolio data? MCP resource makes this natural.
5. **Multi-tenant MCP:** How to handle auth for remote MCP server deployment? OAuth 2.1?

---

## Success Criteria

1. **Portfolio building:** User can complete full portfolio in < 30 minutes via conversation
2. **CV generation:** Generated CV scores 80+ on ATS checks
3. **Tailoring time:** From JD paste to tailored CV < 2 minutes
4. **MCP compatibility:** Works with Claude Desktop, Cursor, and custom agents
5. **Memory accuracy:** System correctly recalls 95%+ of user-provided data

---

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| AI hallucination in CV | Low quality output | Human review step, confidence scores |
| MCP adoption slower than expected | Less utility | Also expose as REST API |
| PDF rendering inconsistencies | Unprofessional output | Standardized templates, testing |
| BYOK security concerns | Key exposure | Encrypted storage, never log keys |
| Portfolio data gets stale | Irrelevant CVs | Regular prompts to update |

---

**Next Step:** Review this plan, then start Phase 1 implementation.
