package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
	"project-dash/pkg/models"
)

type CVStore struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewCVStore(db *sql.DB, logger *zap.Logger) *CVStore {
	return &CVStore{db: db, logger: logger}
}

// Portfolio operations

func (s *CVStore) GetOrCreatePortfolio(userID string) (*models.CVPortfolio, error) {
	query := `
		SELECT id, user_id, summary, target_roles, industry_focus, preferred_locations, created_at, updated_at
		FROM cv_portfolios WHERE user_id = ?
	`
	p := &models.CVPortfolio{}
	var targetRoles, industryFocus, preferredLocations sql.NullString

	err := s.db.QueryRow(query, userID).Scan(
		&p.ID, &p.UserID, &p.Summary, &targetRoles, &industryFocus, &preferredLocations,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return s.createPortfolio(userID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	p.TargetRoles = parseStringSlice(targetRoles.String)
	p.IndustryFocus = parseStringSlice(industryFocus.String)
	p.PreferredLocations = parseStringSlice(preferredLocations.String)
	return p, nil
}

func (s *CVStore) createPortfolio(userID string) (*models.CVPortfolio, error) {
	query := `
		INSERT INTO cv_portfolios (user_id) VALUES (?)
		RETURNING id, user_id, summary, created_at, updated_at
	`
	p := &models.CVPortfolio{UserID: userID}
	err := s.db.QueryRow(query, userID).Scan(
		&p.ID, &p.UserID, &p.Summary, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create portfolio: %w", err)
	}
	return p, nil
}

func (s *CVStore) UpdatePortfolio(p *models.CVPortfolio) error {
	query := `
		UPDATE cv_portfolios
		SET summary = ?, target_roles = ?, industry_focus = ?, preferred_locations = ?, updated_at = datetime('now')
		WHERE id = ?
	`
	_, err := s.db.Exec(query, p.Summary, stringSliceToJSON(p.TargetRoles),
		stringSliceToJSON(p.IndustryFocus), stringSliceToJSON(p.PreferredLocations), p.ID)
	if err != nil {
		return fmt.Errorf("failed to update portfolio: %w", err)
	}
	return nil
}

// Experience operations

func (s *CVStore) AddExperience(e *models.CVExperience) error {
	query := `
		INSERT INTO cv_experiences (portfolio_id, company, position, location, start_date, end_date,
			employment_type, description, key_responsibilities, technologies_used, team_size, reporting_to, is_current)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRow(query, e.PortfolioID, e.Company, e.Position, e.Location,
		e.StartDate, e.EndDate, e.EmploymentType, e.Description,
		stringSliceToJSON(e.KeyResponsibilities), stringSliceToJSON(e.TechnologiesUsed),
		e.TeamSize, e.ReportingTo, boolToInt(e.IsCurrent)).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to add experience: %w", err)
	}
	return nil
}

func (s *CVStore) GetExperiences(portfolioID string) ([]models.CVExperience, error) {
	query := `
		SELECT id, portfolio_id, company, position, location, start_date, end_date,
			employment_type, description, key_responsibilities, technologies_used,
			team_size, reporting_to, is_current, created_at, updated_at
		FROM cv_experiences WHERE portfolio_id = ? ORDER BY start_date DESC
	`
	rows, err := s.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get experiences: %w", err)
	}
	defer rows.Close()

	var experiences []models.CVExperience
	for rows.Next() {
		var e models.CVExperience
		var keyResp, techUsed sql.NullString
		var isCurrent int
		err := rows.Scan(&e.ID, &e.PortfolioID, &e.Company, &e.Position, &e.Location,
			&e.StartDate, &e.EndDate, &e.EmploymentType, &e.Description,
			&keyResp, &techUsed, &e.TeamSize, &e.ReportingTo, &isCurrent,
			&e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan experience: %w", err)
		}
		e.IsCurrent = isCurrent == 1
		e.KeyResponsibilities = parseStringSlice(keyResp.String)
		e.TechnologiesUsed = parseStringSlice(techUsed.String)
		experiences = append(experiences, e)
	}
	return experiences, nil
}

func (s *CVStore) UpdateExperience(e *models.CVExperience) error {
	query := `
		UPDATE cv_experiences
		SET company = ?, position = ?, location = ?, start_date = ?, end_date = ?,
			employment_type = ?, description = ?, key_responsibilities = ?, technologies_used = ?,
			team_size = ?, reporting_to = ?, is_current = ?, updated_at = datetime('now')
		WHERE id = ?
	`
	_, err := s.db.Exec(query, e.Company, e.Position, e.Location, e.StartDate, e.EndDate,
		e.EmploymentType, e.Description, stringSliceToJSON(e.KeyResponsibilities),
		stringSliceToJSON(e.TechnologiesUsed), e.TeamSize, e.ReportingTo, boolToInt(e.IsCurrent), e.ID)
	if err != nil {
		return fmt.Errorf("failed to update experience: %w", err)
	}
	return nil
}

func (s *CVStore) DeleteExperience(id string) error {
	_, err := s.db.Exec("DELETE FROM cv_experiences WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete experience: %w", err)
	}
	return nil
}

// Achievement operations

func (s *CVStore) AddAchievement(a *models.CVAchievement) error {
	metricsJSON, _ := json.Marshal(a.Metrics)
	query := `
		INSERT INTO cv_achievements (portfolio_id, experience_id, title, description, impact,
			metrics, skills_used, category, relevance_score)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at, updated_at
	`
	err := s.db.QueryRow(query, a.PortfolioID, a.ExperienceID, a.Title, a.Description,
		a.Impact, string(metricsJSON), stringSliceToJSON(a.SkillsUsed),
		a.Category, a.RelevanceScore).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to add achievement: %w", err)
	}
	return nil
}

func (s *CVStore) GetAchievements(portfolioID string) ([]models.CVAchievement, error) {
	query := `
		SELECT id, portfolio_id, experience_id, title, description, impact, metrics,
			skills_used, category, relevance_score, created_at, updated_at
		FROM cv_achievements WHERE portfolio_id = ? ORDER BY relevance_score DESC
	`
	rows, err := s.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}
	defer rows.Close()

	var achievements []models.CVAchievement
	for rows.Next() {
		var a models.CVAchievement
		var experienceID, metrics, skillsUsed sql.NullString
		err := rows.Scan(&a.ID, &a.PortfolioID, &experienceID, &a.Title, &a.Description,
			&a.Impact, &metrics, &skillsUsed, &a.Category, &a.RelevanceScore,
			&a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		if experienceID.Valid {
			a.ExperienceID = &experienceID.String
		}
		if metrics.Valid {
			json.Unmarshal([]byte(metrics.String), &a.Metrics)
		}
		a.SkillsUsed = parseStringSlice(skillsUsed.String)
		achievements = append(achievements, a)
	}
	return achievements, nil
}

func (s *CVStore) GetAchievementsForExperience(experienceID string) ([]models.CVAchievement, error) {
	query := `
		SELECT id, portfolio_id, experience_id, title, description, impact, metrics,
			skills_used, category, relevance_score, created_at, updated_at
		FROM cv_achievements WHERE experience_id = ? ORDER BY relevance_score DESC
	`
	rows, err := s.db.Query(query, experienceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievements: %w", err)
	}
	defer rows.Close()

	var achievements []models.CVAchievement
	for rows.Next() {
		var a models.CVAchievement
		var experienceID, metrics, skillsUsed sql.NullString
		err := rows.Scan(&a.ID, &a.PortfolioID, &experienceID, &a.Title, &a.Description,
			&a.Impact, &metrics, &skillsUsed, &a.Category, &a.RelevanceScore,
			&a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan achievement: %w", err)
		}
		if experienceID.Valid {
			a.ExperienceID = &experienceID.String
		}
		if metrics.Valid {
			json.Unmarshal([]byte(metrics.String), &a.Metrics)
		}
		a.SkillsUsed = parseStringSlice(skillsUsed.String)
		achievements = append(achievements, a)
	}
	return achievements, nil
}

func (s *CVStore) UpdateAchievement(a *models.CVAchievement) error {
	metricsJSON, _ := json.Marshal(a.Metrics)
	query := `
		UPDATE cv_achievements
		SET experience_id = ?, title = ?, description = ?, impact = ?, metrics = ?,
			skills_used = ?, category = ?, relevance_score = ?, updated_at = datetime('now')
		WHERE id = ?
	`
	_, err := s.db.Exec(query, a.ExperienceID, a.Title, a.Description, a.Impact,
		string(metricsJSON), stringSliceToJSON(a.SkillsUsed), a.Category,
		a.RelevanceScore, a.ID)
	if err != nil {
		return fmt.Errorf("failed to update achievement: %w", err)
	}
	return nil
}

func (s *CVStore) DeleteAchievement(id string) error {
	_, err := s.db.Exec("DELETE FROM cv_achievements WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete achievement: %w", err)
	}
	return nil
}

// Skill operations

func (s *CVStore) AddSkill(skill *models.CVSkill) error {
	query := `
		INSERT INTO cv_skills (portfolio_id, name, category, proficiency, years_of_experience, last_used, is_highlight)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(query, skill.PortfolioID, skill.Name, skill.Category,
		skill.Proficiency, skill.YearsOfExperience, skill.LastUsed, boolToInt(skill.IsHighlight)).
		Scan(&skill.ID, &skill.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to add skill: %w", err)
	}
	return nil
}

func (s *CVStore) GetSkills(portfolioID string) ([]models.CVSkill, error) {
	query := `
		SELECT id, portfolio_id, name, category, proficiency, years_of_experience, last_used, is_highlight, created_at
		FROM cv_skills WHERE portfolio_id = ? ORDER BY is_highlight DESC, name
	`
	rows, err := s.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skills: %w", err)
	}
	defer rows.Close()

	var skills []models.CVSkill
	for rows.Next() {
		var sk models.CVSkill
		var isHighlight int
		err := rows.Scan(&sk.ID, &sk.PortfolioID, &sk.Name, &sk.Category, &sk.Proficiency,
			&sk.YearsOfExperience, &sk.LastUsed, &isHighlight, &sk.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan skill: %w", err)
		}
		sk.IsHighlight = isHighlight == 1
		skills = append(skills, sk)
	}
	return skills, nil
}

func (s *CVStore) UpdateSkill(skill *models.CVSkill) error {
	query := `
		UPDATE cv_skills
		SET name = ?, category = ?, proficiency = ?, years_of_experience = ?, last_used = ?, is_highlight = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, skill.Name, skill.Category, skill.Proficiency,
		skill.YearsOfExperience, skill.LastUsed, boolToInt(skill.IsHighlight), skill.ID)
	if err != nil {
		return fmt.Errorf("failed to update skill: %w", err)
	}
	return nil
}

func (s *CVStore) DeleteSkill(id string) error {
	_, err := s.db.Exec("DELETE FROM cv_skills WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete skill: %w", err)
	}
	return nil
}

// Education operations

func (s *CVStore) AddEducation(edu *models.CVEducation) error {
	query := `
		INSERT INTO cv_education (portfolio_id, institution, degree, field_of_study, start_date, end_date, gpa, honors, relevant_coursework)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(query, edu.PortfolioID, edu.Institution, edu.Degree, edu.FieldOfStudy,
		edu.StartDate, edu.EndDate, edu.GPA, stringSliceToJSON(edu.Honors),
		stringSliceToJSON(edu.RelevantCoursework)).Scan(&edu.ID, &edu.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to add education: %w", err)
	}
	return nil
}

func (s *CVStore) GetEducation(portfolioID string) ([]models.CVEducation, error) {
	query := `
		SELECT id, portfolio_id, institution, degree, field_of_study, start_date, end_date, gpa, honors, relevant_coursework, created_at
		FROM cv_education WHERE portfolio_id = ? ORDER BY start_date DESC
	`
	rows, err := s.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get education: %w", err)
	}
	defer rows.Close()

	var educations []models.CVEducation
	for rows.Next() {
		var edu models.CVEducation
		var honors, relevantCoursework sql.NullString
		err := rows.Scan(&edu.ID, &edu.PortfolioID, &edu.Institution, &edu.Degree, &edu.FieldOfStudy,
			&edu.StartDate, &edu.EndDate, &edu.GPA, &honors, &relevantCoursework, &edu.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan education: %w", err)
		}
		edu.Honors = parseStringSlice(honors.String)
		edu.RelevantCoursework = parseStringSlice(relevantCoursework.String)
		educations = append(educations, edu)
	}
	return educations, nil
}

func (s *CVStore) UpdateEducation(edu *models.CVEducation) error {
	query := `
		UPDATE cv_education
		SET institution = ?, degree = ?, field_of_study = ?, start_date = ?, end_date = ?,
			gpa = ?, honors = ?, relevant_coursework = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, edu.Institution, edu.Degree, edu.FieldOfStudy, edu.StartDate,
		edu.EndDate, edu.GPA, stringSliceToJSON(edu.Honors), stringSliceToJSON(edu.RelevantCoursework), edu.ID)
	if err != nil {
		return fmt.Errorf("failed to update education: %w", err)
	}
	return nil
}

func (s *CVStore) DeleteEducation(id string) error {
	_, err := s.db.Exec("DELETE FROM cv_education WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete education: %w", err)
	}
	return nil
}

// Certification operations

func (s *CVStore) AddCertification(cert *models.CVCertification) error {
	query := `
		INSERT INTO cv_certifications (portfolio_id, name, issuer, issue_date, expiry_date, credential_id, credential_url)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(query, cert.PortfolioID, cert.Name, cert.Issuer, cert.IssueDate,
		cert.ExpiryDate, cert.CredentialID, cert.CredentialURL).Scan(&cert.ID, &cert.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to add certification: %w", err)
	}
	return nil
}

func (s *CVStore) GetCertifications(portfolioID string) ([]models.CVCertification, error) {
	query := `
		SELECT id, portfolio_id, name, issuer, issue_date, expiry_date, credential_id, credential_url, created_at
		FROM cv_certifications WHERE portfolio_id = ? ORDER BY issue_date DESC
	`
	rows, err := s.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get certifications: %w", err)
	}
	defer rows.Close()

	var certs []models.CVCertification
	for rows.Next() {
		var cert models.CVCertification
		err := rows.Scan(&cert.ID, &cert.PortfolioID, &cert.Name, &cert.Issuer, &cert.IssueDate,
			&cert.ExpiryDate, &cert.CredentialID, &cert.CredentialURL, &cert.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan certification: %w", err)
		}
		certs = append(certs, cert)
	}
	return certs, nil
}

func (s *CVStore) UpdateCertification(cert *models.CVCertification) error {
	query := `
		UPDATE cv_certifications
		SET name = ?, issuer = ?, issue_date = ?, expiry_date = ?, credential_id = ?, credential_url = ?
		WHERE id = ?
	`
	_, err := s.db.Exec(query, cert.Name, cert.Issuer, cert.IssueDate, cert.ExpiryDate,
		cert.CredentialID, cert.CredentialURL, cert.ID)
	if err != nil {
		return fmt.Errorf("failed to update certification: %w", err)
	}
	return nil
}

func (s *CVStore) DeleteCertification(id string) error {
	_, err := s.db.Exec("DELETE FROM cv_certifications WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete certification: %w", err)
	}
	return nil
}

// Generated CV operations

func (s *CVStore) SaveGeneratedCV(cv *models.CVGenerated) error {
	query := `
		INSERT INTO cv_generated (portfolio_id, template_id, job_description, content, markdown_content, ats_score, tailoring_notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id, created_at
	`
	err := s.db.QueryRow(query, cv.PortfolioID, cv.TemplateID, cv.JobDescription,
		cv.Content, cv.MarkdownContent, cv.ATSScore, cv.TailoringNotes).Scan(&cv.ID, &cv.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save generated CV: %w", err)
	}
	return nil
}

func (s *CVStore) GetGeneratedCVs(portfolioID string) ([]models.CVGenerated, error) {
	query := `
		SELECT id, portfolio_id, template_id, job_description, content, markdown_content, ats_score, tailoring_notes, created_at
		FROM cv_generated WHERE portfolio_id = ? ORDER BY created_at DESC
	`
	rows, err := s.db.Query(query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get generated CVs: %w", err)
	}
	defer rows.Close()

	var cvs []models.CVGenerated
	for rows.Next() {
		var cv models.CVGenerated
		err := rows.Scan(&cv.ID, &cv.PortfolioID, &cv.TemplateID, &cv.JobDescription,
			&cv.Content, &cv.MarkdownContent, &cv.ATSScore, &cv.TailoringNotes, &cv.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan generated CV: %w", err)
		}
		cvs = append(cvs, cv)
	}
	return cvs, nil
}

// Search

func (s *CVStore) SearchPortfolio(portfolioID, query string) (map[string]interface{}, error) {
	results := make(map[string]interface{})

	// Search experiences
	expQuery := `
		SELECT id, company, position, description FROM cv_experiences
		WHERE portfolio_id = ? AND (company LIKE ? OR position LIKE ? OR description LIKE ?)
	`
	expRows, err := s.db.Query(expQuery, portfolioID, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err == nil {
		defer expRows.Close()
		var experiences []map[string]interface{}
		for expRows.Next() {
			var id, company, position, description string
			expRows.Scan(&id, &company, &position, &description)
			experiences = append(experiences, map[string]interface{}{
				"type": "experience", "id": id, "company": company,
				"position": position, "description": description,
			})
		}
		results["experiences"] = experiences
	}

	// Search achievements
	achQuery := `
		SELECT id, title, description, impact FROM cv_achievements
		WHERE portfolio_id = ? AND (title LIKE ? OR description LIKE ? OR impact LIKE ?)
	`
	achRows, err := s.db.Query(achQuery, portfolioID, "%"+query+"%", "%"+query+"%", "%"+query+"%")
	if err == nil {
		defer achRows.Close()
		var achievements []map[string]interface{}
		for achRows.Next() {
			var id, title, description, impact string
			achRows.Scan(&id, &title, &description, &impact)
			achievements = append(achievements, map[string]interface{}{
				"type": "achievement", "id": id, "title": title,
				"description": description, "impact": impact,
			})
		}
		results["achievements"] = achievements
	}

	// Search skills
	skillQuery := `
		SELECT id, name, category FROM cv_skills
		WHERE portfolio_id = ? AND (name LIKE ? OR category LIKE ?)
	`
	skillRows, err := s.db.Query(skillQuery, portfolioID, "%"+query+"%", "%"+query+"%")
	if err == nil {
		defer skillRows.Close()
		var skills []map[string]interface{}
		for skillRows.Next() {
			var id, name, category string
			skillRows.Scan(&id, &name, &category)
			skills = append(skills, map[string]interface{}{
				"type": "skill", "id": id, "name": name, "category": category,
			})
		}
		results["skills"] = skills
	}

	return results, nil
}

// Helpers

func parseStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	json.Unmarshal([]byte(s), &result)
	return result
}

func stringSliceToJSON(arr []string) string {
	if arr == nil {
		return ""
	}
	b, _ := json.Marshal(arr)
	return string(b)
}

// Ensure context is used
var _ = context.Background
