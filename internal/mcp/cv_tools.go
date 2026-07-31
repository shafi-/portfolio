package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"project-dash/internal/cv"
	"project-dash/pkg/models"
)

func (s *Server) cvTools() []serverTool {
	return []serverTool{
		// Portfolio
		{
			Tool:    mcp.NewTool("getCVPortfolio", mcp.WithString("user_id", mcp.Description("User ID (default: default)"))),
			Handler: s.handleGetCVPortfolio,
		},
		// Experience
		{
			Tool: mcp.NewTool("addExperience",
				mcp.WithString("company", mcp.Required(), mcp.Description("Company name")),
				mcp.WithString("position", mcp.Required(), mcp.Description("Job title/position")),
				mcp.WithString("start_date", mcp.Required(), mcp.Description("Start date (YYYY-MM-DD)")),
				mcp.WithString("location", mcp.Description("Work location")),
				mcp.WithString("end_date", mcp.Description("End date (YYYY-MM-DD, empty for current)")),
				mcp.WithString("employment_type", mcp.Description("Employment type: full-time, part-time, contract, internship, freelance")),
				mcp.WithString("description", mcp.Description("Role description")),
				mcp.WithString("key_responsibilities", mcp.Description("JSON array of key responsibilities")),
				mcp.WithString("technologies_used", mcp.Description("JSON array of technologies used")),
				mcp.WithString("team_size", mcp.Description("Team size")),
				mcp.WithString("reporting_to", mcp.Description("Who they reported to")),
			),
			Handler: s.handleAddExperience,
		},
		{
			Tool: mcp.NewTool("updateExperience",
				mcp.WithString("id", mcp.Required(), mcp.Description("Experience ID")),
				mcp.WithString("company", mcp.Description("Company name")),
				mcp.WithString("position", mcp.Description("Job title/position")),
				mcp.WithString("start_date", mcp.Description("Start date (YYYY-MM-DD)")),
				mcp.WithString("end_date", mcp.Description("End date (YYYY-MM-DD)")),
				mcp.WithString("location", mcp.Description("Work location")),
				mcp.WithString("employment_type", mcp.Description("Employment type")),
				mcp.WithString("description", mcp.Description("Role description")),
				mcp.WithString("key_responsibilities", mcp.Description("JSON array of key responsibilities")),
				mcp.WithString("technologies_used", mcp.Description("JSON array of technologies used")),
				mcp.WithString("team_size", mcp.Description("Team size")),
				mcp.WithString("reporting_to", mcp.Description("Who they reported to")),
				mcp.WithString("is_current", mcp.Description("Is this the current role (true/false)")),
			),
			Handler: s.handleUpdateExperience,
		},
		{
			Tool: mcp.NewTool("deleteExperience",
				mcp.WithString("id", mcp.Required(), mcp.Description("Experience ID")),
			),
			Handler: s.handleDeleteExperience,
		},
		{
			Tool:    mcp.NewTool("listExperiences"),
			Handler: s.handleListExperiences,
		},
		// Achievements
		{
			Tool: mcp.NewTool("addAchievement",
				mcp.WithString("title", mcp.Required(), mcp.Description("Achievement title")),
				mcp.WithString("description", mcp.Required(), mcp.Description("What was achieved")),
				mcp.WithString("experience_id", mcp.Description("Related experience ID")),
				mcp.WithString("impact", mcp.Description("Impact description")),
				mcp.WithString("metrics", mcp.Description("JSON object of metrics: {\"revenue\": \"$1M\", \"reduction\": \"30%\"}")),
				mcp.WithString("skills_used", mcp.Description("JSON array of skills demonstrated")),
				mcp.WithString("category", mcp.Description("Category: project, leadership, technical, process, revenue, other")),
			),
			Handler: s.handleAddAchievement,
		},
		{
			Tool: mcp.NewTool("updateAchievement",
				mcp.WithString("id", mcp.Required(), mcp.Description("Achievement ID")),
				mcp.WithString("title", mcp.Description("Achievement title")),
				mcp.WithString("description", mcp.Description("What was achieved")),
				mcp.WithString("experience_id", mcp.Description("Related experience ID")),
				mcp.WithString("impact", mcp.Description("Impact description")),
				mcp.WithString("metrics", mcp.Description("JSON object of metrics")),
				mcp.WithString("skills_used", mcp.Description("JSON array of skills demonstrated")),
				mcp.WithString("category", mcp.Description("Category")),
			),
			Handler: s.handleUpdateAchievement,
		},
		{
			Tool: mcp.NewTool("deleteAchievement",
				mcp.WithString("id", mcp.Required(), mcp.Description("Achievement ID")),
			),
			Handler: s.handleDeleteAchievement,
		},
		{
			Tool:    mcp.NewTool("listAchievements"),
			Handler: s.handleListAchievements,
		},
		{
			Tool: mcp.NewTool("getAchievementsForExperience",
				mcp.WithString("experience_id", mcp.Required(), mcp.Description("Experience ID")),
			),
			Handler: s.handleGetAchievementsForExperience,
		},
		// Skills
		{
			Tool: mcp.NewTool("addSkill",
				mcp.WithString("name", mcp.Required(), mcp.Description("Skill name")),
				mcp.WithString("category", mcp.Description("Category: technical, soft, language, tool, framework, other")),
				mcp.WithString("proficiency", mcp.Description("Proficiency: beginner, intermediate, advanced, expert")),
				mcp.WithString("years_of_experience", mcp.Description("Years of experience")),
				mcp.WithString("last_used", mcp.Description("Last used date (YYYY-MM-DD)")),
				mcp.WithString("is_highlight", mcp.Description("Highlight on CV (true/false)")),
			),
			Handler: s.handleAddSkill,
		},
		{
			Tool: mcp.NewTool("updateSkill",
				mcp.WithString("id", mcp.Required(), mcp.Description("Skill ID")),
				mcp.WithString("name", mcp.Description("Skill name")),
				mcp.WithString("category", mcp.Description("Category")),
				mcp.WithString("proficiency", mcp.Description("Proficiency")),
				mcp.WithString("years_of_experience", mcp.Description("Years of experience")),
				mcp.WithString("last_used", mcp.Description("Last used date")),
				mcp.WithString("is_highlight", mcp.Description("Highlight on CV")),
			),
			Handler: s.handleUpdateSkill,
		},
		{
			Tool: mcp.NewTool("deleteSkill",
				mcp.WithString("id", mcp.Required(), mcp.Description("Skill ID")),
			),
			Handler: s.handleDeleteSkill,
		},
		{
			Tool:    mcp.NewTool("listSkills"),
			Handler: s.handleListSkills,
		},
		// Education
		{
			Tool: mcp.NewTool("addEducation",
				mcp.WithString("institution", mcp.Required(), mcp.Description("School/university name")),
				mcp.WithString("degree", mcp.Description("Degree type")),
				mcp.WithString("field_of_study", mcp.Description("Field of study")),
				mcp.WithString("start_date", mcp.Description("Start date (YYYY-MM-DD)")),
				mcp.WithString("end_date", mcp.Description("End date (YYYY-MM-DD)")),
				mcp.WithString("gpa", mcp.Description("GPA")),
				mcp.WithString("honors", mcp.Description("JSON array of honors")),
				mcp.WithString("relevant_coursework", mcp.Description("JSON array of relevant coursework")),
			),
			Handler: s.handleAddEducation,
		},
		{
			Tool: mcp.NewTool("updateEducation",
				mcp.WithString("id", mcp.Required(), mcp.Description("Education ID")),
				mcp.WithString("institution", mcp.Description("School/university name")),
				mcp.WithString("degree", mcp.Description("Degree type")),
				mcp.WithString("field_of_study", mcp.Description("Field of study")),
				mcp.WithString("start_date", mcp.Description("Start date")),
				mcp.WithString("end_date", mcp.Description("End date")),
				mcp.WithString("gpa", mcp.Description("GPA")),
				mcp.WithString("honors", mcp.Description("JSON array of honors")),
				mcp.WithString("relevant_coursework", mcp.Description("JSON array of relevant coursework")),
			),
			Handler: s.handleUpdateEducation,
		},
		{
			Tool: mcp.NewTool("deleteEducation",
				mcp.WithString("id", mcp.Required(), mcp.Description("Education ID")),
			),
			Handler: s.handleDeleteEducation,
		},
		{
			Tool:    mcp.NewTool("listEducation"),
			Handler: s.handleListEducation,
		},
		// Certifications
		{
			Tool: mcp.NewTool("addCertification",
				mcp.WithString("name", mcp.Required(), mcp.Description("Certification name")),
				mcp.WithString("issuer", mcp.Description("Issuing organization")),
				mcp.WithString("issue_date", mcp.Description("Issue date (YYYY-MM-DD)")),
				mcp.WithString("expiry_date", mcp.Description("Expiry date (YYYY-MM-DD)")),
				mcp.WithString("credential_id", mcp.Description("Credential ID")),
				mcp.WithString("credential_url", mcp.Description("Credential verification URL")),
			),
			Handler: s.handleAddCertification,
		},
		{
			Tool: mcp.NewTool("updateCertification",
				mcp.WithString("id", mcp.Required(), mcp.Description("Certification ID")),
				mcp.WithString("name", mcp.Description("Certification name")),
				mcp.WithString("issuer", mcp.Description("Issuing organization")),
				mcp.WithString("issue_date", mcp.Description("Issue date")),
				mcp.WithString("expiry_date", mcp.Description("Expiry date")),
				mcp.WithString("credential_id", mcp.Description("Credential ID")),
				mcp.WithString("credential_url", mcp.Description("Credential verification URL")),
			),
			Handler: s.handleUpdateCertification,
		},
		{
			Tool: mcp.NewTool("deleteCertification",
				mcp.WithString("id", mcp.Required(), mcp.Description("Certification ID")),
			),
			Handler: s.handleDeleteCertification,
		},
		{
			Tool:    mcp.NewTool("listCertifications"),
			Handler: s.handleListCertifications,
		},
		// Search
		{
			Tool: mcp.NewTool("searchCVPortfolio",
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
			),
			Handler: s.handleSearchCVPortfolio,
		},
		// Generated CVs
		{
			Tool:    mcp.NewTool("listGeneratedCVs"),
			Handler: s.handleListGeneratedCVs,
		},
		// CV Generation
		{
			Tool: mcp.NewTool("generateCV",
				mcp.WithString("job_description", mcp.Description("Job description to tailor CV to")),
				mcp.WithString("template", mcp.Description("Template ID: ats, professional, compact")),
			),
			Handler: s.handleGenerateCV,
		},
		{
			Tool:    mcp.NewTool("listCVTemplates"),
			Handler: s.handleListCVTemplates,
		},
	}
}

// Handlers

func (s *Server) handleGetCVPortfolio(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	experiences, _ := s.cv.GetExperiences(portfolio.ID)
	achievements, _ := s.cv.GetAchievements(portfolio.ID)
	skills, _ := s.cv.GetSkills(portfolio.ID)
	educations, _ := s.cv.GetEducation(portfolio.ID)
	certs, _ := s.cv.GetCertifications(portfolio.ID)

	result := map[string]interface{}{
		"portfolio":      portfolio,
		"experiences":    experiences,
		"achievements":   achievements,
		"skills":         skills,
		"education":      educations,
		"certifications": certs,
	}

	return mcp.NewToolResultJSON(result)
}

func (s *Server) handleAddExperience(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	userID := "default"
	if uid, ok := args["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	company, _ := args["company"].(string)
	position, _ := args["position"].(string)
	startDate, _ := args["start_date"].(string)

	if company == "" || position == "" || startDate == "" {
		return mcp.NewToolResultError("company, position, and start_date are required"), nil
	}

	exp := &models.CVExperience{
		PortfolioID:    portfolio.ID,
		Company:        company,
		Position:       position,
		StartDate:      startDate,
		Location:       getStringArg(args, "location"),
		EndDate:        getStringArg(args, "end_date"),
		EmploymentType: getStringArg(args, "employment_type"),
		Description:    getStringArg(args, "description"),
		ReportingTo:    getStringArg(args, "reporting_to"),
	}

	if isCurrent, ok := args["is_current"].(bool); ok {
		exp.IsCurrent = isCurrent
	}

	if err := s.cv.AddExperience(exp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add experience: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success":    true,
		"experience": exp,
		"message":    fmt.Sprintf("Added experience at %s as %s", company, position),
	})
}

func (s *Server) handleUpdateExperience(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	exp := &models.CVExperience{
		ID:             id,
		Company:        getStringArg(args, "company"),
		Position:       getStringArg(args, "position"),
		StartDate:      getStringArg(args, "start_date"),
		EndDate:        getStringArg(args, "end_date"),
		Location:       getStringArg(args, "location"),
		EmploymentType: getStringArg(args, "employment_type"),
		Description:    getStringArg(args, "description"),
		ReportingTo:    getStringArg(args, "reporting_to"),
	}

	if isCurrent, ok := args["is_current"].(string); ok {
		exp.IsCurrent = isCurrent == "true"
	}

	if err := s.cv.UpdateExperience(exp); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update experience: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Experience updated",
	})
}

func (s *Server) handleDeleteExperience(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := req.GetArguments()["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	if err := s.cv.DeleteExperience(id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete experience: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Experience deleted",
	})
}

func (s *Server) handleListExperiences(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	experiences, err := s.cv.GetExperiences(portfolio.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get experiences: %v", err)), nil
	}

	return mcp.NewToolResultJSON(experiences)
}

func (s *Server) handleAddAchievement(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	userID := "default"
	if uid, ok := args["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	title, _ := args["title"].(string)
	description, _ := args["description"].(string)

	if title == "" || description == "" {
		return mcp.NewToolResultError("title and description are required"), nil
	}

	ach := &models.CVAchievement{
		PortfolioID: portfolio.ID,
		Title:       title,
		Description: description,
		Impact:      getStringArg(args, "impact"),
		Category:    getStringArg(args, "category"),
	}

	if expID, ok := args["experience_id"].(string); ok && expID != "" {
		ach.ExperienceID = &expID
	}

	if err := s.cv.AddAchievement(ach); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add achievement: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success":     true,
		"achievement": ach,
		"message":     fmt.Sprintf("Added achievement: %s", title),
	})
}

func (s *Server) handleUpdateAchievement(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	ach := &models.CVAchievement{
		ID:          id,
		Title:       getStringArg(args, "title"),
		Description: getStringArg(args, "description"),
		Impact:      getStringArg(args, "impact"),
		Category:    getStringArg(args, "category"),
	}

	if expID, ok := args["experience_id"].(string); ok && expID != "" {
		ach.ExperienceID = &expID
	}

	if err := s.cv.UpdateAchievement(ach); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update achievement: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Achievement updated",
	})
}

func (s *Server) handleDeleteAchievement(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := req.GetArguments()["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	if err := s.cv.DeleteAchievement(id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete achievement: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Achievement deleted",
	})
}

func (s *Server) handleListAchievements(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	achievements, err := s.cv.GetAchievements(portfolio.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get achievements: %v", err)), nil
	}

	return mcp.NewToolResultJSON(achievements)
}

func (s *Server) handleGetAchievementsForExperience(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	expID, _ := req.GetArguments()["experience_id"].(string)
	if expID == "" {
		return mcp.NewToolResultError("experience_id is required"), nil
	}

	achievements, err := s.cv.GetAchievementsForExperience(expID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get achievements: %v", err)), nil
	}

	return mcp.NewToolResultJSON(achievements)
}

func (s *Server) handleAddSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	userID := "default"
	if uid, ok := args["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	name, _ := args["name"].(string)
	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	skill := &models.CVSkill{
		PortfolioID: portfolio.ID,
		Name:        name,
		Category:    getStringArg(args, "category"),
		Proficiency: getStringArg(args, "proficiency"),
		LastUsed:    getStringArg(args, "last_used"),
	}

	if isHighlight, ok := args["is_highlight"].(bool); ok {
		skill.IsHighlight = isHighlight
	}

	if err := s.cv.AddSkill(skill); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add skill: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"skill":   skill,
		"message": fmt.Sprintf("Added skill: %s", name),
	})
}

func (s *Server) handleUpdateSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	skill := &models.CVSkill{
		ID:          id,
		Name:        getStringArg(args, "name"),
		Category:    getStringArg(args, "category"),
		Proficiency: getStringArg(args, "proficiency"),
		LastUsed:    getStringArg(args, "last_used"),
	}

	if isHighlight, ok := args["is_highlight"].(bool); ok {
		skill.IsHighlight = isHighlight
	}

	if err := s.cv.UpdateSkill(skill); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update skill: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Skill updated",
	})
}

func (s *Server) handleDeleteSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := req.GetArguments()["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	if err := s.cv.DeleteSkill(id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete skill: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Skill deleted",
	})
}

func (s *Server) handleListSkills(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	skills, err := s.cv.GetSkills(portfolio.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get skills: %v", err)), nil
	}

	return mcp.NewToolResultJSON(skills)
}

func (s *Server) handleAddEducation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	userID := "default"
	if uid, ok := args["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	institution, _ := args["institution"].(string)
	if institution == "" {
		return mcp.NewToolResultError("institution is required"), nil
	}

	edu := &models.CVEducation{
		PortfolioID:  portfolio.ID,
		Institution:  institution,
		Degree:       getStringArg(args, "degree"),
		FieldOfStudy: getStringArg(args, "field_of_study"),
		StartDate:    getStringArg(args, "start_date"),
		EndDate:      getStringArg(args, "end_date"),
	}

	if err := s.cv.AddEducation(edu); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add education: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success":   true,
		"education": edu,
		"message":   fmt.Sprintf("Added education at %s", institution),
	})
}

func (s *Server) handleUpdateEducation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	edu := &models.CVEducation{
		ID:           id,
		Institution:  getStringArg(args, "institution"),
		Degree:       getStringArg(args, "degree"),
		FieldOfStudy: getStringArg(args, "field_of_study"),
		StartDate:    getStringArg(args, "start_date"),
		EndDate:      getStringArg(args, "end_date"),
	}

	if err := s.cv.UpdateEducation(edu); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update education: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Education updated",
	})
}

func (s *Server) handleDeleteEducation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := req.GetArguments()["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	if err := s.cv.DeleteEducation(id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete education: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Education deleted",
	})
}

func (s *Server) handleListEducation(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	educations, err := s.cv.GetEducation(portfolio.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get education: %v", err)), nil
	}

	return mcp.NewToolResultJSON(educations)
}

func (s *Server) handleAddCertification(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	userID := "default"
	if uid, ok := args["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	name, _ := args["name"].(string)
	if name == "" {
		return mcp.NewToolResultError("name is required"), nil
	}

	cert := &models.CVCertification{
		PortfolioID:   portfolio.ID,
		Name:          name,
		Issuer:        getStringArg(args, "issuer"),
		IssueDate:     getStringArg(args, "issue_date"),
		ExpiryDate:    getStringArg(args, "expiry_date"),
		CredentialID:  getStringArg(args, "credential_id"),
		CredentialURL: getStringArg(args, "credential_url"),
	}

	if err := s.cv.AddCertification(cert); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to add certification: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success":       true,
		"certification": cert,
		"message":       fmt.Sprintf("Added certification: %s", name),
	})
}

func (s *Server) handleUpdateCertification(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()
	id, _ := args["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	cert := &models.CVCertification{
		ID:            id,
		Name:          getStringArg(args, "name"),
		Issuer:        getStringArg(args, "issuer"),
		IssueDate:     getStringArg(args, "issue_date"),
		ExpiryDate:    getStringArg(args, "expiry_date"),
		CredentialID:  getStringArg(args, "credential_id"),
		CredentialURL: getStringArg(args, "credential_url"),
	}

	if err := s.cv.UpdateCertification(cert); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to update certification: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Certification updated",
	})
}

func (s *Server) handleDeleteCertification(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, _ := req.GetArguments()["id"].(string)
	if id == "" {
		return mcp.NewToolResultError("id is required"), nil
	}

	if err := s.cv.DeleteCertification(id); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete certification: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"success": true,
		"message": "Certification deleted",
	})
}

func (s *Server) handleListCertifications(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	certs, err := s.cv.GetCertifications(portfolio.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get certifications: %v", err)), nil
	}

	return mcp.NewToolResultJSON(certs)
}

func (s *Server) handleSearchCVPortfolio(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, _ := req.GetArguments()["query"].(string)
	if query == "" {
		return mcp.NewToolResultError("query is required"), nil
	}

	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	results, err := s.cv.SearchPortfolio(portfolio.ID, query)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to search portfolio: %v", err)), nil
	}

	return mcp.NewToolResultJSON(results)
}

func (s *Server) handleListGeneratedCVs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	userID := "default"
	if uid, ok := req.GetArguments()["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	cvs, err := s.cv.GetGeneratedCVs(portfolio.ID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get generated CVs: %v", err)), nil
	}

	return mcp.NewToolResultJSON(cvs)
}

func (s *Server) handleGenerateCV(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := req.GetArguments()

	userID := "default"
	if uid, ok := args["user_id"].(string); ok && uid != "" {
		userID = uid
	}

	portfolio, err := s.cv.GetOrCreatePortfolio(userID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get portfolio: %v", err)), nil
	}

	experiences, _ := s.cv.GetExperiences(portfolio.ID)
	achievements, _ := s.cv.GetAchievements(portfolio.ID)
	skills, _ := s.cv.GetSkills(portfolio.ID)
	educations, _ := s.cv.GetEducation(portfolio.ID)
	certs, _ := s.cv.GetCertifications(portfolio.ID)

	jobDesc, _ := args["job_description"].(string)
	templateID, _ := args["template"].(string)
	if templateID == "" {
		templateID = "ats"
	}

	generator := cv.NewGenerator()
	result := generator.Generate(&cv.GenerateInput{
		Portfolio:      portfolio,
		Experiences:    experiences,
		Achievements:   achievements,
		Skills:         skills,
		Education:      educations,
		Certifications: certs,
		JobDescription: jobDesc,
		TemplateID:     templateID,
	})

	// Save generated CV
	generated := &models.CVGenerated{
		PortfolioID:     portfolio.ID,
		TemplateID:      templateID,
		JobDescription:  jobDesc,
		MarkdownContent: result.Markdown,
		ATSScore:        result.ATSScore,
		TailoringNotes:  strings.Join(result.TailorNotes, "; "),
	}

	if err := s.cv.SaveGeneratedCV(generated); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to save generated CV: %v", err)), nil
	}

	return mcp.NewToolResultJSON(map[string]interface{}{
		"cv_id":        generated.ID,
		"markdown":     result.Markdown,
		"ats_score":    result.ATSScore,
		"tailor_notes": result.TailorNotes,
		"template":     templateID,
	})
}

func (s *Server) handleListCVTemplates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	templates := cv.GetTemplates()
	return mcp.NewToolResultJSON(templates)
}
