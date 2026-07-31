# AI CV Builder — Go-Native Implementation Plan

Generated: 2026-07-31
Status: DRAFT

---

## Architecture Decision

**Original plan:** Two separate TypeScript MCP servers + Supabase
**Actual implementation:** Go-native, integrated into existing MCP server + SQLite

Rationale:
- Existing codebase is 100% Go with `mcp-go` library
- SQLite already in use with PRAGMA key encryption
- Store layer pattern already established
- Single MCP server architecture (not multi-server)
- No external dependencies (Supabase, TypeScript toolchain)

---

## What We're Building

New domain within the existing Portfolio Engine:

1. **CV Portfolio** — structured career data (experiences, achievements, skills, education, certifications)
2. **CV Generator** — takes portfolio data + job description, produces tailored CV

All exposed as MCP tools on the existing server.

---

## Database Schema (SQLite)

### Migration v4: cv_portfolio

```sql
-- CV Portfolio tables

CREATE TABLE cv_portfolios (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  user_id TEXT NOT NULL DEFAULT 'default',
  summary TEXT,
  target_roles TEXT,          -- JSON array
  industry_focus TEXT,        -- JSON array
  preferred_locations TEXT,   -- JSON array
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cv_experiences (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  company TEXT NOT NULL,
  position TEXT NOT NULL,
  location TEXT,
  start_date TEXT NOT NULL,
  end_date TEXT,              -- NULL = current
  employment_type TEXT CHECK (employment_type IN ('full-time', 'part-time', 'contract', 'internship', 'freelance')),
  description TEXT,
  key_responsibilities TEXT,  -- JSON array
  technologies_used TEXT,     -- JSON array
  team_size INTEGER,
  reporting_to TEXT,
  is_current INTEGER DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cv_achievements (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  experience_id TEXT REFERENCES cv_experiences(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  impact TEXT,
  metrics TEXT,               -- JSON object: {"revenue": "$1M", "reduction": "30%"}
  skills_used TEXT,           -- JSON array
  category TEXT CHECK (category IN ('project', 'leadership', 'technical', 'process', 'revenue', 'other')),
  relevance_score REAL DEFAULT 0.5,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cv_skills (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  category TEXT CHECK (category IN ('technical', 'soft', 'language', 'tool', 'framework', 'other')),
  proficiency TEXT CHECK (proficiency IN ('beginner', 'intermediate', 'advanced', 'expert')),
  years_of_experience INTEGER,
  last_used TEXT,
  is_highlight INTEGER DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cv_education (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  institution TEXT NOT NULL,
  degree TEXT,
  field_of_study TEXT,
  start_date TEXT,
  end_date TEXT,
  gpa REAL,
  honors TEXT,                -- JSON array
  relevant_coursework TEXT,   -- JSON array
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cv_certifications (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  issuer TEXT,
  issue_date TEXT,
  expiry_date TEXT,
  credential_id TEXT,
  credential_url TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE cv_generated (
  id TEXT PRIMARY KEY DEFAULT (lower(hex(randomblob(16)))),
  portfolio_id TEXT NOT NULL REFERENCES cv_portfolios(id) ON DELETE CASCADE,
  template_id TEXT,
  job_description TEXT,
  content TEXT NOT NULL,      -- JSON structured CV
  markdown_content TEXT,      -- Rendered markdown
  ats_score REAL,
  tailoring_notes TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Indexes
CREATE INDEX idx_cv_portfolios_user ON cv_portfolios(user_id);
CREATE INDEX idx_cv_experiences_portfolio ON cv_experiences(portfolio_id);
CREATE INDEX idx_cv_achievements_portfolio ON cv_achievements(portfolio_id);
CREATE INDEX idx_cv_achievements_experience ON cv_achievements(experience_id);
CREATE INDEX idx_cv_skills_portfolio ON cv_skills(portfolio_id);
CREATE INDEX idx_cv_education_portfolio ON cv_education(portfolio_id);
CREATE INDEX idx_cv_certifications_portfolio ON cv_certifications(portfolio_id);
CREATE INDEX idx_cv_generated_portfolio ON cv_generated(portfolio_id);
```

---

## Go Package Structure

```
internal/
├── cv/                          # NEW: CV Portfolio domain
│   ├── portfolio.go             # Portfolio CRUD
│   ├── experience.go            # Experience CRUD
│   ├── achievement.go           # Achievement CRUD
│   ├── skill.go                 # Skill CRUD
│   ├── education.go             # Education CRUD
│   ├── certification.go         # Certification CRUD
│   ├── generated.go             # Generated CV storage
│   └── generator.go             # CV generation logic
├── mcp/
│   ├── cv_tools.go              # NEW: MCP tools for CV portfolio
│   └── server.go                # Modified: register cvTools()
├── database/
│   └── migrations.go            # Modified: add v4 migration
├── store/
│   └── cv.go                    # NEW: CV data access layer
└── pdf/                         # NEW: PDF generation
    └── renderer.go
```

---

## MCP Tools

### Portfolio Memory Tools

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `get_cv_portfolio` | Get user's CV portfolio | `user_id?` | Portfolio JSON |
| `add_experience` | Add work experience | Company, role, dates, etc. | Confirmation |
| `update_experience` | Update experience | ID + fields | Confirmation |
| `delete_experience` | Remove experience | ID | Confirmation |
| `add_achievement` | Add achievement to experience | Experience ID, data | Confirmation |
| `update_achievement` | Update achievement | ID + fields | Confirmation |
| `delete_achievement` | Remove achievement | ID | Confirmation |
| `add_skill` | Add skill with proficiency | Name, level, category | Confirmation |
| `update_skill` | Update skill | ID + fields | Confirmation |
| `delete_skill` | Remove skill | ID | Confirmation |
| `add_education` | Add education entry | Institution, degree, dates | Confirmation |
| `update_education` | Update education | ID + fields | Confirmation |
| `delete_education` | Remove education | ID | Confirmation |
| `add_certification` | Add certification | Name, issuer, dates | Confirmation |
| `update_certification` | Update certification | ID + fields | Confirmation |
| `delete_certification` | Remove certification | ID | Confirmation |
| `search_cv_portfolio` | Search across portfolio | Query string | Matching entries |
| `get_achievements_for_role` | Get relevant achievements for job | Job description | Ranked achievements |

### CV Generator Tools

| Tool | Description | Input | Output |
|------|-------------|-------|--------|
| `generate_cv` | Generate CV from portfolio | `user_id`, `job_description?`, `template?` | CV markdown |
| `tailor_cv` | Adjust existing CV | `cv_id`, `adjustments` | Updated CV |
| `preview_cv` | Preview without saving | Portfolio + JD | CV markdown |
| `list_cv_templates` | List available templates | None | Template list |
| `get_cv_ats_score` | Score CV for ATS | `cv_id` or content | Score + suggestions |

---

## Models

```go
// pkg/models/cv.go

type CVPortfolio struct {
    ID                string   `json:"id"`
    UserID            string   `json:"user_id"`
    Summary           string   `json:"summary,omitempty"`
    TargetRoles       []string `json:"target_roles,omitempty"`
    IndustryFocus     []string `json:"industry_focus,omitempty"`
    PreferredLocations []string `json:"preferred_locations,omitempty"`
    CreatedAt         string   `json:"created_at"`
    UpdatedAt         string   `json:"updated_at"`
}

type CVExperience struct {
    ID                  string   `json:"id"`
    PortfolioID         string   `json:"portfolio_id"`
    Company             string   `json:"company"`
    Position            string   `json:"position"`
    Location            string   `json:"location,omitempty"`
    StartDate           string   `json:"start_date"`
    EndDate             string   `json:"end_date,omitempty"`
    EmploymentType      string   `json:"employment_type,omitempty"`
    Description         string   `json:"description,omitempty"`
    KeyResponsibilities  []string `json:"key_responsibilities,omitempty"`
    TechnologiesUsed    []string `json:"technologies_used,omitempty"`
    TeamSize            *int     `json:"team_size,omitempty"`
    ReportingTo         string   `json:"reporting_to,omitempty"`
    IsCurrent           bool     `json:"is_current"`
    CreatedAt           string   `json:"created_at"`
    UpdatedAt           string   `json:"updated_at"`
}

type CVAchievement struct {
    ID              string             `json:"id"`
    PortfolioID     string             `json:"portfolio_id"`
    ExperienceID    *string            `json:"experience_id,omitempty"`
    Title           string             `json:"title"`
    Description     string             `json:"description"`
    Impact          string             `json:"impact,omitempty"`
    Metrics         map[string]string  `json:"metrics,omitempty"`
    SkillsUsed      []string           `json:"skills_used,omitempty"`
    Category        string             `json:"category,omitempty"`
    RelevanceScore  float64            `json:"relevance_score"`
    CreatedAt       string             `json:"created_at"`
    UpdatedAt       string             `json:"updated_at"`
}

type CVSkill struct {
    ID                string  `json:"id"`
    PortfolioID       string  `json:"portfolio_id"`
    Name              string  `json:"name"`
    Category          string  `json:"category,omitempty"`
    Proficiency       string  `json:"proficiency,omitempty"`
    YearsOfExperience *int    `json:"years_of_experience,omitempty"`
    LastUsed          string  `json:"last_used,omitempty"`
    IsHighlight       bool    `json:"is_highlight"`
    CreatedAt         string  `json:"created_at"`
}

type CVEducation struct {
    ID                  string   `json:"id"`
    PortfolioID         string   `json:"portfolio_id"`
    Institution         string   `json:"institution"`
    Degree              string   `json:"degree,omitempty"`
    FieldOfStudy        string   `json:"field_of_study,omitempty"`
    StartDate           string   `json:"start_date,omitempty"`
    EndDate             string   `json:"end_date,omitempty"`
    GPA                 *float64 `json:"gpa,omitempty"`
    Honors              []string `json:"honors,omitempty"`
    RelevantCoursework  []string `json:"relevant_coursework,omitempty"`
    CreatedAt           string   `json:"created_at"`
}

type CVCertification struct {
    ID            string `json:"id"`
    PortfolioID   string `json:"portfolio_id"`
    Name          string `json:"name"`
    Issuer        string `json:"issuer,omitempty"`
    IssueDate     string `json:"issue_date,omitempty"`
    ExpiryDate    string `json:"expiry_date,omitempty"`
    CredentialID  string `json:"credential_id,omitempty"`
    CredentialURL string `json:"credential_url,omitempty"`
    CreatedAt     string `json:"created_at"`
}

type CVGenerated struct {
    ID              string  `json:"id"`
    PortfolioID     string  `json:"portfolio_id"`
    TemplateID      string  `json:"template_id,omitempty"`
    JobDescription  string  `json:"job_description,omitempty"`
    Content         string  `json:"content"`           // JSON structured
    MarkdownContent string  `json:"markdown_content"`
    ATSScore        float64 `json:"ats_score,omitempty"`
    TailoringNotes  string  `json:"tailoring_notes,omitempty"`
    CreatedAt       string  `json:"created_at"`
}
```

---

## Implementation Phases

### Phase 1: CV Portfolio Storage (Days 1-3)

1. **Database**
   - Add v4 migration to `internal/database/migrations.go`
   - Create `internal/store/cv.go` with all CRUD operations
   - Follow existing `Querier` interface pattern

2. **Models**
   - Create `pkg/models/cv.go` with all CV types

3. **MCP Tools**
   - Create `internal/mcp/cv_tools.go`
   - Register in `server.go` via `cvTools()` method
   - Implement: `get_cv_portfolio`, `add_experience`, `add_achievement`, `add_skill`, `add_education`, `add_certification`
   - Implement update/delete variants

4. **Testing**
   - Unit tests for store layer
   - MCP tool tests

### Phase 2: CV Generation (Days 4-6)

1. **Generator**
   - Create `internal/cv/generator.go`
   - JD parsing (extract requirements, keywords)
   - Portfolio matching (score achievements against JD)
   - Markdown CV generation

2. **Templates**
   - Create `internal/cv/templates.go`
   - ATS-safe single-column
   - Professional
   - Creative

3. **MCP Tools**
   - `generate_cv`, `tailor_cv`, `preview_cv`
   - `list_cv_templates`, `get_cv_ats_score`

4. **ATS Scoring**
   - Keyword matching
   - Format validation
   - Structure analysis

### Phase 3: Polish + Testing (Days 7-8)

1. Integration tests
2. Error handling
3. Documentation
4. CLI command `portfolio cv` for direct access

---

## Key Differences from Original Plan

| Aspect | Original Plan | Actual Implementation |
|--------|---------------|----------------------|
| Language | TypeScript | Go |
| Database | Supabase (PostgreSQL) | SQLite (existing) |
| MCP Server | Two separate servers | One integrated server |
| MCP SDK | `@modelcontextprotocol/sdk` | `github.com/mark3labs/mcp-go` |
| Auth | Supabase Auth + RLS | Local single-user |
| PDF Library | Puppeteer / react-pdf | gopdf (pure Go) |
| Deployment | Cloudflare Workers | Local binary |

---

## Success Criteria

1. User can build portfolio via MCP conversation
2. CV generated from portfolio + JD scores 80+ ATS
3. PDF export works
4. All tools accessible via existing MCP server
5. Tests pass
