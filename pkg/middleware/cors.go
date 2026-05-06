package middleware

import (
	"github.com/aws/aws-lambda-go/events"
)

// CORS middleware adds CORS headers to all responses
type CORSMiddleware struct {
	allowOrigins []string
	allowMethods string
	allowHeaders string
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(allowOrigins []string) *CORSMiddleware {
	return &CORSMiddleware{
		allowOrigins: allowOrigins,
		allowMethods: "GET, POST, PUT, DELETE, OPTIONS",
		allowHeaders: "Content-Type, Authorization, X-Requested-With",
	}
}

// AddCORSHeaders adds CORS headers to the response
func (m *CORSMiddleware) AddCORSHeaders() func(HandlerFunc) HandlerFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
			// Handle preflight OPTIONS request
			if request.HTTPMethod == "OPTIONS" {
				return m.preflightResponse(request), nil
			}

			// Call the next handler
			response, err := next(request)
			if err != nil {
				return response, err
			}

			// Add CORS headers to response
			if response.Headers == nil {
				response.Headers = make(map[string]string)
			}

			origin := request.Headers["origin"]
			if origin == "" {
				origin = request.Headers["Origin"]
			}

			// Check if origin is allowed
			if m.isOriginAllowed(origin) {
				response.Headers["Access-Control-Allow-Origin"] = origin
			} else if len(m.allowOrigins) == 1 && m.allowOrigins[0] == "*" {
				response.Headers["Access-Control-Allow-Origin"] = "*"
			}

			response.Headers["Access-Control-Allow-Methods"] = m.allowMethods
			response.Headers["Access-Control-Allow-Headers"] = m.allowHeaders
			response.Headers["Access-Control-Allow-Credentials"] = "true"
			response.Headers["Access-Control-Max-Age"] = "86400"

			return response, nil
		}
	}
}

// preflightResponse returns a response for OPTIONS preflight requests
func (m *CORSMiddleware) preflightResponse(request events.APIGatewayProxyRequest) events.APIGatewayProxyResponse {
	origin := request.Headers["origin"]
	if origin == "" {
		origin = request.Headers["Origin"]
	}

	headers := map[string]string{
		"Access-Control-Allow-Methods":     m.allowMethods,
		"Access-Control-Allow-Headers":     m.allowHeaders,
		"Access-Control-Max-Age":           "86400",
		"Access-Control-Allow-Credentials": "true",
	}

	// Check if origin is allowed
	if m.isOriginAllowed(origin) {
		headers["Access-Control-Allow-Origin"] = origin
	} else if len(m.allowOrigins) == 1 && m.allowOrigins[0] == "*" {
		headers["Access-Control-Allow-Origin"] = "*"
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    headers,
		Body:       "",
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func (m *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range m.allowOrigins {
		if allowed == origin || allowed == "*" {
			return true
		}
	}
	return false
}
