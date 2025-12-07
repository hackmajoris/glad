package handler

import (
	"encoding/json"
	"net/http"

	"github.com/hackmajoris/glad/cmd/app/internal/dto"
	"github.com/hackmajoris/glad/cmd/app/internal/service"
	"github.com/hackmajoris/glad/cmd/app/internal/validation"
	"github.com/hackmajoris/glad/pkg/auth"
	_ "github.com/hackmajoris/glad/pkg/errors"

	"github.com/aws/aws-lambda-go/events"
)

// Handler handles HTTP requests
type Handler struct {
	userService *service.UserService
	errorMapper *ErrorMapper
	validator   *validation.Validator
}

// New creates a new Handler
func New(userService *service.UserService) *Handler {
	return &Handler{
		userService: userService,
		errorMapper: NewErrorMapper(),
		validator:   validation.New(),
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
