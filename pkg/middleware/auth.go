package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/hackmajoris/glad-stack/pkg/auth"
	pkgerrors "github.com/hackmajoris/glad-stack/pkg/errors"
	"github.com/hackmajoris/glad-stack/pkg/logger"

	"github.com/aws/aws-lambda-go/events"
)

// HandlerFunc is the function signature for route handlers
type HandlerFunc func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	tokenService *auth.TokenService
}

// NewAuthMiddleware creates a new AuthMiddleware
func NewAuthMiddleware(tokenService *auth.TokenService) *AuthMiddleware {
	log := logger.WithComponent("middleware")
	log.Info("Auth middleware initialized")

	return &AuthMiddleware{
		tokenService: tokenService,
	}
}

// ValidateJWT wraps a handler with JWT validation
func (m *AuthMiddleware) ValidateJWT(next HandlerFunc) HandlerFunc {
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		log := logger.WithComponent("middleware").With("operation", "ValidateJWT", "path", request.Path, "method", request.HTTPMethod)
		start := time.Now()

		log.Debug("Starting JWT validation for request")

		token := extractTokenFromHeader(request.Headers)
		if token == "" {
			log.Warn("Missing authorization token in request", "duration", time.Since(start))
			return unauthorizedResponse("Missing authorization token"), nil
		}

		log.Debug("JWT token extracted from headers")

		claims, err := m.tokenService.ValidateToken(token)
		if err != nil {
			switch {
			case pkgerrors.Is(err, pkgerrors.ErrInvalidToken):
				log.Warn("Invalid token provided", "duration", time.Since(start))
				return unauthorizedResponse("Invalid or expired token"), nil
			case pkgerrors.Is(err, pkgerrors.ErrTokenExpired):
				log.Warn("Expired token provided", "duration", time.Since(start))
				return unauthorizedResponse("Invalid or expired token"), nil
			default:
				log.Error("JWT token validation failed", "error", err.Error(), "duration", time.Since(start))
				return unauthorizedResponse("Invalid or expired token"), nil
			}
		}

		log = log.With("username", claims.Username)
		log.Debug("JWT validation successful, adding claims to context")

		// Add claims to request context
		if request.RequestContext.Authorizer == nil {
			request.RequestContext.Authorizer = make(map[string]interface{})
		}
		request.RequestContext.Authorizer["claims"] = claims

		log.Info("JWT middleware validation completed, calling handler", "duration", time.Since(start))
		return next(request)
	}
}

// RequireAuth returns a middleware function for use with router
func (m *AuthMiddleware) RequireAuth() func(HandlerFunc) HandlerFunc {
	return m.ValidateJWT
}

// extractTokenFromHeader extracts the JWT token from the Authorization header
func extractTokenFromHeader(headers map[string]string) string {
	log := logger.WithComponent("middleware").With("operation", "extractToken")

	authHeader := headers["Authorization"]
	if authHeader == "" {
		// Try lowercase (API Gateway sometimes normalizes headers)
		authHeader = headers["authorization"]
		if authHeader != "" {
			log.Debug("Found authorization header (lowercase)")
		}
	} else {
		log.Debug("Found Authorization header (capitalized)")
	}

	if authHeader == "" {
		log.Debug("No authorization header found")
		return ""
	}

	parts := strings.Split(authHeader, "Bearer ")
	if len(parts) != 2 {
		log.Debug("Invalid authorization header format", "header", authHeader)
		return ""
	}

	log.Debug("JWT token extracted successfully")
	return parts[1]
}

// unauthorizedResponse creates a standardized unauthorized response
func unauthorizedResponse(message string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusUnauthorized,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"error": "` + message + `"}`,
	}
}
