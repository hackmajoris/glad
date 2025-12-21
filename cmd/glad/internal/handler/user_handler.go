package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/dto"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/models"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/service"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/validation"
	"github.com/hackmajoris/glad-stack/pkg/auth"
	_ "github.com/hackmajoris/glad-stack/pkg/errors"
	"github.com/hackmajoris/glad-stack/pkg/logger"
)

// Handler handles HTTP requests
type Handler struct {
	userService  *service.UserService
	skillService *service.SkillService
	errorMapper  *ErrorMapper
	validator    *validation.Validator
}

// New creates a new Handler
func New(userService *service.UserService, skillService *service.SkillService) *Handler {
	return &Handler{
		userService:  userService,
		skillService: skillService,
		errorMapper:  NewErrorMapper(),
		validator:    validation.New(),
	}
}

// Register handles user registration
func (h *Handler) Register(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req dto.RegisterRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Validate input at handler layer
	if err := h.validator.ValidateRegisterInput(req.Username, req.Name, req.Password); err != nil {
		return h.handleServiceError(err), nil
	}

	_, err := h.userService.Register(req.Username, req.Name, req.Password)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusCreated, dto.MessageResponse{
		Message: "User created successfully",
	}), nil
}

// Login handles user authentication
func (h *Handler) Login(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var req dto.LoginRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Validate input at handler layer
	if err := h.validator.ValidateLoginInput(req.Username, req.Password); err != nil {
		return h.handleServiceError(err), nil
	}

	result, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.TokenResponse{
		AccessToken: result.AccessToken,
		TokenType:   result.TokenType,
	}), nil
}

// Protected handles protected resource access
func (h *Handler) Protected(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, ok := request.RequestContext.Authorizer["claims"].(*auth.JWTClaims)
	if !ok {
		return errorResponse(http.StatusUnauthorized, "Invalid token claims"), nil
	}

	return successResponse(http.StatusOK, dto.ProtectedResponse{
		Message:  "Access granted to protected resource",
		Username: claims.Username,
	}), nil
}

// UpdateUser handles user profile updates
func (h *Handler) UpdateUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, ok := request.RequestContext.Authorizer["claims"].(*auth.JWTClaims)
	if !ok {
		return errorResponse(http.StatusUnauthorized, "Invalid token claims"), nil
	}

	var req dto.UpdateUserRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Validate optional inputs at handler layer
	if err := h.validator.ValidateOptionalName(req.Name); err != nil {
		return h.handleServiceError(err), nil
	}
	if err := h.validator.ValidateOptionalPassword(req.Password); err != nil {
		return h.handleServiceError(err), nil
	}

	err := h.userService.UpdateUser(claims.Username, req.Name, req.Password)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.MessageResponse{
		Message: "User updated successfully",
	}), nil
}

// ListUsers handles listing all users
func (h *Handler) ListUsers(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	users, err := h.userService.ListUsers()
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, users), nil
}

// GetCurrentUser handles retrieving the current authenticated user's information
func (h *Handler) GetCurrentUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	claims, ok := request.RequestContext.Authorizer["claims"].(*auth.JWTClaims)
	if !ok {
		return errorResponse(http.StatusUnauthorized, "Invalid token claims"), nil
	}

	log := logger.WithComponent("handler").With("operation", "GetCurrentUser", "username", claims.Username)
	log.Debug("Fetching current user")

	user, err := h.userService.GetUser(claims.Username)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.CurrentUserResponse{
		Username:  user.Username,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// ============================================================================
// SKILL HANDLERS
// ============================================================================

// AddSkill handles adding a new skill to a user
// POST /users/{username}/skills
func (h *Handler) AddSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get username from path parameter
	username, ok := request.PathParameters["username"]
	if !ok || username == "" {
		return errorResponse(http.StatusBadRequest, "Username is required"), nil
	}

	// Parse request body
	var req dto.CreateSkillRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Convert proficiency level string to type
	proficiencyLevel := models.ProficiencyLevel(req.ProficiencyLevel)

	// Add skill
	skill, err := h.skillService.AddSkill(username, req.SkillName, proficiencyLevel, req.YearsOfExperience, req.Notes)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusCreated, dto.SkillResponse{
		SkillName:         skill.SkillName,
		ProficiencyLevel:  string(skill.ProficiencyLevel),
		YearsOfExperience: skill.YearsOfExperience,
		Endorsements:      skill.Endorsements,
		LastUsedDate:      skill.LastUsedDate,
		Notes:             skill.Notes,
		CreatedAt:         skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// GetSkill handles retrieving a specific skill for a user
// GET /users/{username}/skills/{skillName}
func (h *Handler) GetSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get path parameters
	username, ok := request.PathParameters["username"]
	if !ok || username == "" {
		return errorResponse(http.StatusBadRequest, "Username is required"), nil
	}

	skillName, ok := request.PathParameters["skillName"]
	if !ok || skillName == "" {
		return errorResponse(http.StatusBadRequest, "Skill name is required"), nil
	}

	// Get skill
	skill, err := h.skillService.GetSkill(username, skillName)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.SkillResponse{
		SkillName:         skill.SkillName,
		ProficiencyLevel:  string(skill.ProficiencyLevel),
		YearsOfExperience: skill.YearsOfExperience,
		Endorsements:      skill.Endorsements,
		LastUsedDate:      skill.LastUsedDate,
		Notes:             skill.Notes,
		CreatedAt:         skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// ListSkillsForUser handles listing all skills for a user
// GET /users/{username}/skills
func (h *Handler) ListSkillsForUser(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get username from path parameter
	username, ok := request.PathParameters["username"]
	if !ok || username == "" {
		return errorResponse(http.StatusBadRequest, "Username is required"), nil
	}

	// Get skills
	skills, err := h.skillService.ListSkillsForUser(username)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, skills), nil
}

// UpdateSkill handles updating an existing skill
// PUT /users/{username}/skills/{skillName}
func (h *Handler) UpdateSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get path parameters
	username, ok := request.PathParameters["username"]
	if !ok || username == "" {
		return errorResponse(http.StatusBadRequest, "Username is required"), nil
	}

	skillName, ok := request.PathParameters["skillName"]
	if !ok || skillName == "" {
		return errorResponse(http.StatusBadRequest, "Skill name is required"), nil
	}

	// Parse request body
	var req dto.UpdateSkillRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Convert proficiency level if provided
	var proficiencyLevel *models.ProficiencyLevel
	if req.ProficiencyLevel != nil {
		level := models.ProficiencyLevel(*req.ProficiencyLevel)
		proficiencyLevel = &level
	}

	// Update skill
	skill, err := h.skillService.UpdateSkill(username, skillName, proficiencyLevel, req.YearsOfExperience, req.Notes)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.SkillResponse{
		SkillName:         skill.SkillName,
		ProficiencyLevel:  string(skill.ProficiencyLevel),
		YearsOfExperience: skill.YearsOfExperience,
		Endorsements:      skill.Endorsements,
		LastUsedDate:      skill.LastUsedDate,
		Notes:             skill.Notes,
		CreatedAt:         skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// DeleteSkill handles deleting a skill from a user
// DELETE /users/{username}/skills/{skillName}
func (h *Handler) DeleteSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get path parameters
	username, ok := request.PathParameters["username"]
	if !ok || username == "" {
		return errorResponse(http.StatusBadRequest, "Username is required"), nil
	}

	skillName, ok := request.PathParameters["skillName"]
	if !ok || skillName == "" {
		return errorResponse(http.StatusBadRequest, "Skill name is required"), nil
	}

	// Delete skill
	if err := h.skillService.DeleteSkill(username, skillName); err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.MessageResponse{
		Message: "Skill deleted successfully",
	}), nil
}

// ListUsersBySkill handles finding all users with a specific skill
// GET /skills/{skillName}/users?category=<category>&level=<level>
func (h *Handler) ListUsersBySkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get skill name from path parameter
	skillName, ok := request.PathParameters["skillName"]
	if !ok || skillName == "" {
		return errorResponse(http.StatusBadRequest, "Skill name is required"), nil
	}

	// Get category from query parameters (required for multi-key GSI)
	category, ok := request.QueryStringParameters["category"]
	if !ok || category == "" {
		return errorResponse(http.StatusBadRequest, "Category is required"), nil
	}

	// Check for proficiency level filter in query parameters
	proficiencyLevel, ok := request.QueryStringParameters["level"]
	if ok && proficiencyLevel != "" {
		// Query with level filter
		level := models.ProficiencyLevel(proficiencyLevel)
		users, err := h.skillService.ListUsersBySkillAndLevel(category, skillName, level)
		if err != nil {
			return h.handleServiceError(err), nil
		}
		return successResponse(http.StatusOK, users), nil
	}

	// Query all users with skill
	users, err := h.skillService.ListUsersBySkill(category, skillName)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, users), nil
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// handleServiceError converts service errors to HTTP responses using the error mapper
func (h *Handler) handleServiceError(err error) events.APIGatewayProxyResponse {
	statusCode, message := h.errorMapper.MapToHTTP(err)
	return errorResponse(statusCode, message)
}

func successResponse(statusCode int, data interface{}) events.APIGatewayProxyResponse {
	body, err := json.Marshal(data)
	if err != nil {
		// If marshaling fails, return an error response
		return errorResponse(http.StatusInternalServerError, "Internal server error")
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}

func errorResponse(statusCode int, message string) events.APIGatewayProxyResponse {
	body, err := json.Marshal(dto.ErrorResponse{Error: message})
	if err != nil {
		// Fallback to plain text if JSON marshaling fails
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Content-Type": "text/plain",
			},
			Body: "Internal server error",
		}
	}
	return events.APIGatewayProxyResponse{
		StatusCode: statusCode,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(body),
	}
}
