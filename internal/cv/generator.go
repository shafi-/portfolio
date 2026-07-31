package cv

import (
	"fmt"
	"sort"
	"strings"

	"project-dash/pkg/models"
)

// Generator creates CVs from portfolio data
type Generator struct{}

// NewGenerator creates a new CV generator
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateInput contains all data needed to generate a CV
type GenerateInput struct {
	Portfolio      *models.CVPortfolio
	Experiences    []models.CVExperience
	Achievements   []models.CVAchievement
	Skills         []models.CVSkill
	Education      []models.CVEducation
	Certifications []models.CVCertification
	JobDescription string
	TemplateID     string
}

// GenerateResult contains the generated CV
type GenerateResult struct {
	Markdown    string
	Sections    []Section
	ATSScore    float64
	TailorNotes []string
}

// Section represents a CV section
type Section struct {
	Title   string
	Content string
}

// Generate creates a CV from portfolio data
func (g *Generator) Generate(input *GenerateInput) *GenerateResult {
	template := GetTemplate(input.TemplateID)

	// Parse JD for keywords if provided
	var jdKeywords []string
	if input.JobDescription != "" {
		jdKeywords = ExtractKeywords(input.JobDescription)
	}

	// Score and rank achievements
	scoredAchievements := g.scoreAchievements(input.Achievements, jdKeywords)

	// Filter experiences by relevance
	relevantExperiences := g.filterExperiences(input.Experiences, scoredAchievements, jdKeywords)

	// Filter skills by relevance
	relevantSkills := g.filterSkills(input.Skills, jdKeywords)

	// Build sections
	var sections []Section
	var tailorNotes []string

	// Header
	sections = append(sections, Section{
		Title:   "header",
		Content: g.buildHeader(input.Portfolio),
	})

	// Summary
	if input.Portfolio.Summary != "" {
		sections = append(sections, Section{
			Title:   "summary",
			Content: input.Portfolio.Summary,
		})
	}

	// Experience
	if len(relevantExperiences) > 0 {
		sections = append(sections, Section{
			Title:   "experience",
			Content: g.buildExperienceSection(relevantExperiences, scoredAchievements),
		})
		if input.JobDescription != "" {
			tailorNotes = append(tailorNotes, fmt.Sprintf("Filtered to %d most relevant experiences", len(relevantExperiences)))
		}
	}

	// Skills
	if len(relevantSkills) > 0 {
		sections = append(sections, Section{
			Title:   "skills",
			Content: g.buildSkillsSection(relevantSkills),
		})
		if input.JobDescription != "" {
			tailorNotes = append(tailorNotes, fmt.Sprintf("Highlighted %d relevant skills", len(relevantSkills)))
		}
	}

	// Education
	if len(input.Education) > 0 {
		sections = append(sections, Section{
			Title:   "education",
			Content: g.buildEducationSection(input.Education),
		})
	}

	// Certifications
	if len(input.Certifications) > 0 {
		sections = append(sections, Section{
			Title:   "certifications",
			Content: g.buildCertificationsSection(input.Certifications),
		})
	}

	// Build markdown based on template
	markdown := g.buildMarkdown(sections, template)

	// Calculate ATS score
	atsScore := g.calculateATSScore(markdown, jdKeywords)

	return &GenerateResult{
		Markdown:    markdown,
		Sections:    sections,
		ATSScore:    atsScore,
		TailorNotes: tailorNotes,
	}
}

func (g *Generator) buildHeader(p *models.CVPortfolio) string {
	var parts []string
	if p.TargetRoles != nil && len(p.TargetRoles) > 0 {
		parts = append(parts, strings.Join(p.TargetRoles, " | "))
	}
	return strings.Join(parts, " - ")
}

func (g *Generator) buildExperienceSection(experiences []models.CVExperience, scoredAchievements map[string]float64) string {
	var sb strings.Builder

	for _, exp := range experiences {
		// Header line
		dateRange := formatDateRange(exp.StartDate, exp.EndDate)
		sb.WriteString(fmt.Sprintf("**%s** at %s\n", exp.Position, exp.Company))
		if exp.Location != "" {
			sb.WriteString(fmt.Sprintf("*%s, %s*\n", exp.Location, dateRange))
		} else {
			sb.WriteString(fmt.Sprintf("*%s*\n", dateRange))
		}

		if exp.Description != "" {
			sb.WriteString(fmt.Sprintf("%s\n", exp.Description))
		}

		// Responsibilities
		if len(exp.KeyResponsibilities) > 0 {
			for _, resp := range exp.KeyResponsibilities {
				sb.WriteString(fmt.Sprintf("- %s\n", resp))
			}
		}

		// Achievements for this experience
		for _, ach := range scoredAchievements {
			_ = ach // Use scored achievements
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *Generator) buildSkillsSection(skills []models.CVSkill) string {
	var sb strings.Builder

	// Group by category
	byCategory := make(map[string][]models.CVSkill)
	for _, skill := range skills {
		cat := skill.Category
		if cat == "" {
			cat = "Other"
		}
		byCategory[cat] = append(byCategory[cat], skill)
	}

	// Sort categories
	categories := make([]string, 0, len(byCategory))
	for cat := range byCategory {
		categories = append(categories, cat)
	}
	sort.Strings(categories)

	for _, cat := range categories {
		catSkills := byCategory[cat]
		var names []string
		for _, s := range catSkills {
			names = append(names, s.Name)
		}
		sb.WriteString(fmt.Sprintf("**%s:** %s\n", cat, strings.Join(names, ", ")))
	}

	return sb.String()
}

func (g *Generator) buildEducationSection(educations []models.CVEducation) string {
	var sb strings.Builder

	for _, edu := range educations {
		if edu.Degree != "" && edu.FieldOfStudy != "" {
			sb.WriteString(fmt.Sprintf("**%s** in %s\n", edu.Degree, edu.FieldOfStudy))
		} else if edu.Degree != "" {
			sb.WriteString(fmt.Sprintf("**%s**\n", edu.Degree))
		}
		sb.WriteString(fmt.Sprintf("%s\n", edu.Institution))

		if edu.GPA != nil {
			sb.WriteString(fmt.Sprintf("GPA: %.2f\n", *edu.GPA))
		}

		if len(edu.Honors) > 0 {
			sb.WriteString(fmt.Sprintf("Honors: %s\n", strings.Join(edu.Honors, ", ")))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *Generator) buildCertificationsSection(certs []models.CVCertification) string {
	var sb strings.Builder

	for _, cert := range certs {
		sb.WriteString(fmt.Sprintf("- **%s**", cert.Name))
		if cert.Issuer != "" {
			sb.WriteString(fmt.Sprintf(" - %s", cert.Issuer))
		}
		if cert.IssueDate != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", cert.IssueDate))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (g *Generator) scoreAchievements(achievements []models.CVAchievement, jdKeywords []string) map[string]float64 {
	scores := make(map[string]float64)

	for _, ach := range achievements {
		score := ach.RelevanceScore

		if len(jdKeywords) > 0 {
			// Boost score based on keyword matches
			achText := strings.ToLower(ach.Title + " " + ach.Description + " " + ach.Impact)
			for _, kw := range jdKeywords {
				if strings.Contains(achText, strings.ToLower(kw)) {
					score += 0.1
				}
			}

			// Boost for skills used
			for _, skill := range ach.SkillsUsed {
				for _, kw := range jdKeywords {
					if strings.ToLower(skill) == strings.ToLower(kw) {
						score += 0.15
					}
				}
			}
		}

		scores[ach.ID] = score
	}

	return scores
}

func (g *Generator) filterExperiences(experiences []models.CVExperience, scoredAchievements map[string]float64, jdKeywords []string) []models.CVExperience {
	if len(jdKeywords) == 0 {
		return experiences
	}

	// Score each experience
	type scoredExp struct {
		exp   models.CVExperience
		score float64
	}

	var scored []scoredExp
	for _, exp := range experiences {
		score := 0.0

		// Base score for recency
		if exp.IsCurrent {
			score += 0.3
		}

		// Score for technologies matching JD
		for _, tech := range exp.TechnologiesUsed {
			for _, kw := range jdKeywords {
				if strings.ToLower(tech) == strings.ToLower(kw) {
					score += 0.2
				}
			}
		}

		// Score for position title matching JD
		posLower := strings.ToLower(exp.Position)
		for _, kw := range jdKeywords {
			if strings.Contains(posLower, strings.ToLower(kw)) {
				score += 0.3
			}
		}

		scored = append(scored, scoredExp{exp: exp, score: score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top experiences (at least 3)
	limit := len(scored)
	if limit > 5 {
		limit = 5
	}
	if limit < 3 && len(scored) > 3 {
		limit = 3
	}

	var result []models.CVExperience
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].exp)
	}

	return result
}

func (g *Generator) filterSkills(skills []models.CVSkill, jdKeywords []string) []models.CVSkill {
	if len(jdKeywords) == 0 {
		return skills
	}

	// Separate highlighted and non-highlighted
	var highlighted, other []models.CVSkill
	for _, skill := range skills {
		if skill.IsHighlight {
			highlighted = append(highlighted, skill)
		} else {
			other = append(other, skill)
		}
	}

	// Score other skills
	type scoredSkill struct {
		skill models.CVSkill
		score float64
	}

	var scored []scoredSkill
	for _, skill := range other {
		score := 0.0
		skillLower := strings.ToLower(skill.Name)

		for _, kw := range jdKeywords {
			if strings.Contains(skillLower, strings.ToLower(kw)) {
				score += 0.3
			}
		}

		// Boost by proficiency
		switch skill.Proficiency {
		case "expert":
			score += 0.2
		case "advanced":
			score += 0.15
		case "intermediate":
			score += 0.1
		}

		scored = append(scored, scoredSkill{skill: skill, score: score})
	}

	// Sort by score
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Combine: highlighted first, then top scored
	result := highlighted
	limit := len(scored)
	if limit > 10 {
		limit = 10
	}
	for i := 0; i < limit; i++ {
		result = append(result, scored[i].skill)
	}

	return result
}

func (g *Generator) buildMarkdown(sections []Section, template *Template) string {
	var sb strings.Builder

	for _, section := range sections {
		switch section.Title {
		case "header":
			sb.WriteString(section.Content)
			sb.WriteString("\n\n")
		case "summary":
			sb.WriteString("## Professional Summary\n\n")
			sb.WriteString(section.Content)
			sb.WriteString("\n\n")
		case "experience":
			sb.WriteString("## Experience\n\n")
			sb.WriteString(section.Content)
			sb.WriteString("\n")
		case "skills":
			sb.WriteString("## Skills\n\n")
			sb.WriteString(section.Content)
			sb.WriteString("\n")
		case "education":
			sb.WriteString("## Education\n\n")
			sb.WriteString(section.Content)
			sb.WriteString("\n")
		case "certifications":
			sb.WriteString("## Certifications\n\n")
			sb.WriteString(section.Content)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (g *Generator) calculateATSScore(cv string, jdKeywords []string) float64 {
	if len(jdKeywords) == 0 {
		return 80.0 // Base score without JD
	}

	score := 50.0
	cvLower := strings.ToLower(cv)

	// Check for keyword matches
	matches := 0
	for _, kw := range jdKeywords {
		if strings.Contains(cvLower, strings.ToLower(kw)) {
			matches++
		}
	}

	if len(jdKeywords) > 0 {
		score += float64(matches) / float64(len(jdKeywords)) * 30
	}

	// Check for common ATS-friendly patterns
	if strings.Contains(cvLower, "experience") {
		score += 5
	}
	if strings.Contains(cvLower, "skills") {
		score += 5
	}
	if strings.Contains(cvLower, "education") {
		score += 5
	}
	if strings.Contains(cvLower, "achievement") || strings.Contains(cvLower, "accomplished") {
		score += 5
	}

	if score > 100 {
		score = 100
	}

	return score
}

// ExtractKeywords extracts keywords from job description
func ExtractKeywords(jd string) []string {
	// Simple keyword extraction - split by common delimiters
	// and filter out common words
	commonWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "being": true, "have": true, "has": true, "had": true,
		"do": true, "does": true, "did": true, "will": true, "would": true,
		"could": true, "should": true, "may": true, "might": true, "shall": true,
		"we": true, "you": true, "they": true, "it": true, "this": true,
		"that": true, "these": true, "those": true, "i": true, "me": true,
		"my": true, "our": true, "your": true, "their": true, "its": true,
		"who": true, "which": true, "what": true, "where": true, "when": true,
		"how": true, "why": true, "all": true, "each": true, "every": true,
		"both": true, "few": true, "more": true, "most": true, "other": true,
		"some": true, "such": true, "no": true, "not": true, "only": true,
		"own": true, "same": true, "so": true, "than": true, "too": true,
		"very": true, "can": true, "just": true, "now": true,
	}

	// Split into words
	words := strings.Fields(jd)

	// Deduplicate and filter
	seen := make(map[string]bool)
	var keywords []string

	for _, word := range words {
		// Clean word
		word = strings.Trim(word, ".,;:!?\"'()[]{}")
		word = strings.ToLower(word)

		// Skip short words, common words, and duplicates
		if len(word) < 3 || commonWords[word] || seen[word] {
			continue
		}

		seen[word] = true
		keywords = append(keywords, word)
	}

	return keywords
}

// formatDateRange formats date range for display
func formatDateRange(start, end string) string {
	if end == "" || end == "present" {
		return fmt.Sprintf("%s - Present", start)
	}
	return fmt.Sprintf("%s - %s", start, end)
}
