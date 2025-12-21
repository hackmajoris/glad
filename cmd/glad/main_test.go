//go:build lambda
// +build lambda

package main

import (
	"testing"

	"go-on-aws/cmd/glad/internal/api"
	"go-on-aws/cmd/glad/internal/database"
	"go-on-aws/pkg/auth"
	"go-on-aws/pkg/middleware"

	"github.com/aws/aws-lambda-go/events"
)

func TestNotFoundHandler(t *testing.T) {
	response := notFound()

	if response.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", response.StatusCode)
	}

	if response.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", response.Headers["Content-Type"])
	}

	expected := `{"error": "Not Found"}`
	if response.Body != expected {
		t.Errorf("Expected body %s, got %s", expected, response.Body)
	}
}

func TestMethodNotAllowedHandler(t *testing.T) {
	response := methodNotAllowed()

	if response.StatusCode != 405 {
		t.Errorf("Expected status 405, got %d", response.StatusCode)
	}

	expected := `{"error": "Method Not Allowed"}`
	if response.Body != expected {
		t.Errorf("Expected body %s, got %s", expected, response.Body)
	}
}

func TestRegisterHandler(t *testing.T) {
	// Setup dependencies with mock repository
	userRepo := database.NewMockRepository()
	tokenService := auth.NewTokenService()
	apiHandler := api.NewHandler(userRepo, tokenService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// Test valid registration request
	request := events.APIGatewayProxyRequest{
		HTTPMethod: "POST",
		Path:       "/register",
		Body:       `{"username":"testuser","name":"Test User","password":"password123"}`,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}

	response, err := handleRequest(request, apiHandler, authMiddleware)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response.StatusCode != 201 {
		t.Errorf("Expected status 201, got %d", response.StatusCode)
	}
}

func TestInvalidPath(t *testing.T) {
	userRepo := database.NewMockRepository()
	tokenService := auth.NewTokenService()
	apiHandler := api.NewHandler(userRepo, tokenService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	request := events.APIGatewayProxyRequest{
		HTTPMethod: "GET",
		Path:       "/invalid",
	}

	response, err := handleRequest(request, apiHandler, authMiddleware)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", response.StatusCode)
	}
}
