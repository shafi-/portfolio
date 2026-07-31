package cv

// Template represents a CV template
type Template struct {
	ID          string
	Name        string
	Description string
	Format      string // "ats", "professional", "creative"
}

// GetTemplates returns available CV templates
func GetTemplates() []Template {
	return []Template{
		{
			ID:          "ats",
			Name:        "ATS-Safe",
			Description: "Single-column, keyword-optimized for Applicant Tracking Systems",
			Format:      "ats",
		},
		{
			ID:          "professional",
			Name:        "Professional",
			Description: "Clean, modern layout with sections for experience and skills",
			Format:      "professional",
		},
		{
			ID:          "compact",
			Name:        "Compact",
			Description: "Dense format for experienced professionals with extensive history",
			Format:      "compact",
		},
	}
}

// GetTemplate returns a template by ID
func GetTemplate(id string) *Template {
	for _, t := range GetTemplates() {
		if t.ID == id {
			return &t
		}
	}
	// Default to ATS
	templates := GetTemplates()
	return &templates[0]
}
