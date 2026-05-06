package router

import (
	"net/http"

	"github.com/hackmajoris/glad-stack/pkg/middleware"

	"github.com/aws/aws-lambda-go/events"
)

// HandlerFunc is the function signature for route handlers
type HandlerFunc = middleware.HandlerFunc

// Middleware wraps a handler with additional functionality
type Middleware func(HandlerFunc) HandlerFunc

// Route represents a single route
type Route struct {
	Method     string
	Path       string
	Handler    HandlerFunc
	Middleware []Middleware
}

// Router handles HTTP routing for Lambda
type Router struct {
	routes           map[string]map[string]Route // path -> method -> route
	globalMiddleware []Middleware
}

// New creates a new Router
func New() *Router {
	return &Router{
		routes:           make(map[string]map[string]Route),
		globalMiddleware: make([]Middleware, 0),
	}
}

// Use adds global middleware that applies to all routes
func (r *Router) Use(middleware Middleware) {
	r.globalMiddleware = append(r.globalMiddleware, middleware)
}

// Handle registers a route with optional middleware
func (r *Router) Handle(method, path string, handler HandlerFunc, middleware ...Middleware) {
	if r.routes[path] == nil {
		r.routes[path] = make(map[string]Route)
	}

	r.routes[path][method] = Route{
		Method:     method,
		Path:       path,
		Handler:    handler,
		Middleware: middleware,
	}
}

// GET registers a GET route
func (r *Router) GET(path string, handler HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodGet, path, handler, middleware...)
}

// POST registers a POST route
func (r *Router) POST(path string, handler HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPost, path, handler, middleware...)
}

// PUT registers a PUT route
func (r *Router) PUT(path string, handler HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodPut, path, handler, middleware...)
}

// DELETE registers a DELETE route
func (r *Router) DELETE(path string, handler HandlerFunc, middleware ...Middleware) {
	r.Handle(http.MethodDelete, path, handler, middleware...)
}

// Route handles an incoming request
func (r *Router) Route(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Use Resource instead of Path to match route patterns (handles stage prefix)
	pathRoutes, exists := r.routes[request.Resource]
	if !exists {
		return NotFoundResponse(), nil
	}

	route, exists := pathRoutes[request.HTTPMethod]
	if !exists {
		return MethodNotAllowedResponse(), nil
	}

	// Apply middleware in reverse order (last registered runs first around handler)
	handler := route.Handler

	// First apply route-specific middleware
	for i := len(route.Middleware) - 1; i >= 0; i-- {
		handler = route.Middleware[i](handler)
	}

	// Then apply global middleware (outermost)
	for i := len(r.globalMiddleware) - 1; i >= 0; i-- {
		handler = r.globalMiddleware[i](handler)
	}

	return handler(request)
}

// NotFoundResponse returns a 404 response
func NotFoundResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusNotFound,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"error": "Not Found"}`,
	}
}

// MethodNotAllowedResponse returns a 405 response
func MethodNotAllowedResponse() events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusMethodNotAllowed,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: `{"error": "Method Not Allowed"}`,
	}
}
