package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hackmajoris/glad/cmd/glad/internal/dto"
	"github.com/hackmajoris/glad/cmd/glad/internal/service"

	"github.com/aws/aws-lambda-go/events"
)

// MasterSkillHandler handles master skill HTTP requests
type MasterSkillHandler struct {
	service     *service.MasterSkillService
	errorMapper *ErrorMapper
}

// NewMasterSkillHandler creates a new MasterSkillHandler
func NewMasterSkillHandler(service *service.MasterSkillService) *MasterSkillHandler {
	return &MasterSkillHandler{
		service:     service,
		errorMapper: NewErrorMapper(),
	}
}

// CreateMasterSkill handles creating a new master skill
// POST /skills
func (h *MasterSkillHandler) CreateMasterSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var req dto.CreateMasterSkillRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Create master skill
	skill, err := h.service.CreateMasterSkill(req.SkillID, req.SkillName, req.Description, req.Category, req.Tags)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusCreated, dto.MasterSkillResponse{
		SkillID:     skill.SkillID,
		SkillName:   skill.SkillName,
		Description: skill.Description,
		Category:    skill.Category,
		Tags:        skill.Tags,
		CreatedAt:   skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// GetMasterSkill handles retrieving a master skill by ID
// GET /skills/{skillID}
func (h *MasterSkillHandler) GetMasterSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get skill ID from path parameter
	skillID, ok := request.PathParameters["skillID"]
	if !ok || skillID == "" {
		return errorResponse(http.StatusBadRequest, "Skill ID is required"), nil
	}

	// Get master skill
	skill, err := h.service.GetMasterSkill(skillID)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.MasterSkillResponse{
		SkillID:     skill.SkillID,
		SkillName:   skill.SkillName,
		Description: skill.Description,
		Category:    skill.Category,
		Tags:        skill.Tags,
		CreatedAt:   skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// UpdateMasterSkill handles updating an existing master skill
// PUT /skills/{skillID}
func (h *MasterSkillHandler) UpdateMasterSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get skill ID from path parameter
	skillID, ok := request.PathParameters["skillID"]
	if !ok || skillID == "" {
		return errorResponse(http.StatusBadRequest, "Skill ID is required"), nil
	}

	// Parse request body
	var req dto.UpdateMasterSkillRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return errorResponse(http.StatusBadRequest, "Invalid request body"), nil
	}

	// Update master skill
	skill, err := h.service.UpdateMasterSkill(skillID, req.SkillName, req.Description, req.Category, req.Tags)
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.MasterSkillResponse{
		SkillID:     skill.SkillID,
		SkillName:   skill.SkillName,
		Description: skill.Description,
		Category:    skill.Category,
		Tags:        skill.Tags,
		CreatedAt:   skill.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   skill.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}), nil
}

// DeleteMasterSkill handles deleting a master skill
// DELETE /skills/{skillID}
func (h *MasterSkillHandler) DeleteMasterSkill(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Get skill ID from path parameter
	skillID, ok := request.PathParameters["skillID"]
	if !ok || skillID == "" {
		return errorResponse(http.StatusBadRequest, "Skill ID is required"), nil
	}

	// Delete master skill
	if err := h.service.DeleteMasterSkill(skillID); err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, dto.MessageResponse{
		Message: "Master skill deleted successfully",
	}), nil
}

// ListMasterSkills handles listing all master skills
// GET /skills
func (h *MasterSkillHandler) ListMasterSkills(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// List all master skills
	skills, err := h.service.ListMasterSkills()
	if err != nil {
		return h.handleServiceError(err), nil
	}

	return successResponse(http.StatusOK, skills), nil
}

// handleServiceError converts service errors to HTTP responses using the error mapper
func (h *MasterSkillHandler) handleServiceError(err error) events.APIGatewayProxyResponse {
	statusCode, message := h.errorMapper.MapToHTTP(err)
	return errorResponse(statusCode, message)
}
