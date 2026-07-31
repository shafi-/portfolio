package models

// CVPortfolio represents a user's career portfolio
type CVPortfolio struct {
	ID                 string   `json:"id"`
	UserID             string   `json:"user_id"`
	Summary            string   `json:"summary,omitempty"`
	TargetRoles        []string `json:"target_roles,omitempty"`
	IndustryFocus      []string `json:"industry_focus,omitempty"`
	PreferredLocations []string `json:"preferred_locations,omitempty"`
	CreatedAt          string   `json:"created_at"`
	UpdatedAt          string   `json:"updated_at"`
}

// CVExperience represents work experience
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
	KeyResponsibilities []string `json:"key_responsibilities,omitempty"`
	TechnologiesUsed    []string `json:"technologies_used,omitempty"`
	TeamSize            *int     `json:"team_size,omitempty"`
	ReportingTo         string   `json:"reporting_to,omitempty"`
	IsCurrent           bool     `json:"is_current"`
	CreatedAt           string   `json:"created_at"`
	UpdatedAt           string   `json:"updated_at"`
}

// CVAchievement represents an achievement within an experience
type CVAchievement struct {
	ID             string            `json:"id"`
	PortfolioID    string            `json:"portfolio_id"`
	ExperienceID   *string           `json:"experience_id,omitempty"`
	Title          string            `json:"title"`
	Description    string            `json:"description"`
	Impact         string            `json:"impact,omitempty"`
	Metrics        map[string]string `json:"metrics,omitempty"`
	SkillsUsed     []string          `json:"skills_used,omitempty"`
	Category       string            `json:"category,omitempty"`
	RelevanceScore float64           `json:"relevance_score"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

// CVSkill represents a skill with proficiency
type CVSkill struct {
	ID                string `json:"id"`
	PortfolioID       string `json:"portfolio_id"`
	Name              string `json:"name"`
	Category          string `json:"category,omitempty"`
	Proficiency       string `json:"proficiency,omitempty"`
	YearsOfExperience *int   `json:"years_of_experience,omitempty"`
	LastUsed          string `json:"last_used,omitempty"`
	IsHighlight       bool   `json:"is_highlight"`
	CreatedAt         string `json:"created_at"`
}

// CVEducation represents education entry
type CVEducation struct {
	ID                 string   `json:"id"`
	PortfolioID        string   `json:"portfolio_id"`
	Institution        string   `json:"institution"`
	Degree             string   `json:"degree,omitempty"`
	FieldOfStudy       string   `json:"field_of_study,omitempty"`
	StartDate          string   `json:"start_date,omitempty"`
	EndDate            string   `json:"end_date,omitempty"`
	GPA                *float64 `json:"gpa,omitempty"`
	Honors             []string `json:"honors,omitempty"`
	RelevantCoursework []string `json:"relevant_coursework,omitempty"`
	CreatedAt          string   `json:"created_at"`
}

// CVCertification represents a certification
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

// CVGenerated represents a generated CV
type CVGenerated struct {
	ID              string  `json:"id"`
	PortfolioID     string  `json:"portfolio_id"`
	TemplateID      string  `json:"template_id,omitempty"`
	JobDescription  string  `json:"job_description,omitempty"`
	Content         string  `json:"content"`
	MarkdownContent string  `json:"markdown_content"`
	ATSScore        float64 `json:"ats_score,omitempty"`
	TailoringNotes  string  `json:"tailoring_notes,omitempty"`
	CreatedAt       string  `json:"created_at"`
}
