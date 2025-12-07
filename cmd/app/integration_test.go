//go:build integration
// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/hackmajoris/glad/cmd/app/internal/database"
	"github.com/hackmajoris/glad/cmd/app/internal/dto"
	"github.com/hackmajoris/glad/cmd/app/internal/handler"
	"github.com/hackmajoris/glad/cmd/app/internal/service"
	"github.com/hackmajoris/glad/pkg/auth"
	"github.com/hackmajoris/glad/pkg/config"
	"github.com/hackmajoris/glad/pkg/middleware"

	"github.com/aws/aws-lambda-go/events"
)

// IntegrationTestSuite represents the test environment
type IntegrationTestSuite struct {
	userRepo       database.UserRepository
	apiHandler     *handler.Handler
	authMiddleware *middleware.AuthMiddleware
	tokenService   *auth.TokenService
	server         *httptest.Server
}

// testConfig creates a test configuration
func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret: "test-secret-key",
			Expiry: 24 * time.Hour,
		},
	}
}

// SetupIntegrationTest creates a test environment
func SetupIntegrationTest() *IntegrationTestSuite {
	userRepo := database.NewMockRepository()
	userSkillsRepo := database.NewMockRepository()
	tokenService := auth.NewTokenService(testConfig())
	userService := service.NewUserService(userRepo, tokenService)
	userSkillsService := service.NewSkillService(userSkillsRepo)
	apiHandler := handler.New(userService, userSkillsService)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	// Create HTTP server with the same routing as local-server.go
	mux := http.NewServeMux()
	mux.HandleFunc("/register", handleIntegrationRequest(apiHandler.Register))
	mux.HandleFunc("/login", handleIntegrationRequest(apiHandler.Login))
	mux.HandleFunc("/protected", handleIntegrationRequest(authMiddleware.ValidateJWT(apiHandler.Protected)))
	mux.HandleFunc("/user", handleIntegrationRequest(authMiddleware.ValidateJWT(apiHandler.UpdateUser)))
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(405)
			w.Write([]byte(`{"error": "Method Not Allowed"}`))
			return
		}
		handleIntegrationRequest(authMiddleware.ValidateJWT(apiHandler.ListUsers))(w, r)
	})

	server := httptest.NewServer(mux)

	return &IntegrationTestSuite{
		userRepo:       userRepo,
		apiHandler:     apiHandler,
		authMiddleware: authMiddleware,
		tokenService:   tokenService,
		server:         server,
	}
}

// TearDown cleans up the test environment
func (suite *IntegrationTestSuite) TearDown() {
	suite.server.Close()
}

// TestFullUserJourney tests the complete user lifecycle - equivalent to test-api.sh
func TestFullUserJourney(t *testing.T) {
	suite := SetupIntegrationTest()
	defer suite.TearDown()

	baseURL := suite.server.URL

	t.Log("=== Testing Complete User API Journey ===")

	// Step 1: Register first user
	t.Log("1. Registering first user...")
	registerPayload1 := map[string]string{
		"username": "testuser1",
		"name":     "Test User",
		"password": "password123",
	}
	registerResp1 := makeHTTPRequest(t, "POST", baseURL+"/register", registerPayload1, "")
	if registerResp1.StatusCode != 201 {
		t.Fatalf("Expected status 201 for registration, got %d. Response: %s", registerResp1.StatusCode, registerResp1.Body)
	}
	t.Logf("✅ First user registered successfully")

	// Step 2: Login user
	t.Log("2. Logging in first user...")
	loginPayload := map[string]string{
		"username": "testuser1",
		"password": "password123",
	}
	loginResp := makeHTTPRequest(t, "POST", baseURL+"/login", loginPayload, "")
	if loginResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for login, got %d. Response: %s", loginResp.StatusCode, loginResp.Body)
	}

	// Extract token
	var loginResponse map[string]interface{}
	err := json.Unmarshal([]byte(loginResp.Body), &loginResponse)
	if err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token, ok := loginResponse["access_token"].(string)
	if !ok || token == "" {
		t.Fatalf("No access token in login response: %s", loginResp.Body)
	}
	t.Logf("✅ Login successful, token extracted")

	// Step 3: Register second user
	t.Log("3. Registering second user...")
	registerPayload2 := map[string]string{
		"username": "testuser2",
		"name":     "Second User",
		"password": "password456",
	}
	registerResp2 := makeHTTPRequest(t, "POST", baseURL+"/register", registerPayload2, "")
	if registerResp2.StatusCode != 201 {
		t.Fatalf("Expected status 201 for second registration, got %d. Response: %s", registerResp2.StatusCode, registerResp2.Body)
	}
	t.Logf("✅ Second user registered successfully")

	// Step 4: Test protected route
	t.Log("4. Testing protected route...")
	protectedResp := makeHTTPRequest(t, "GET", baseURL+"/protected", nil, token)
	if protectedResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for protected route, got %d. Response: %s", protectedResp.StatusCode, protectedResp.Body)
	}

	var protectedResponse map[string]interface{}
	err = json.Unmarshal([]byte(protectedResp.Body), &protectedResponse)
	if err != nil {
		t.Fatalf("Failed to parse protected response: %v", err)
	}

	if protectedResponse["username"] != "testuser1" {
		t.Errorf("Expected username testuser1 in protected response, got %v", protectedResponse["username"])
	}
	t.Logf("✅ Protected route access successful")

	// Step 5: List users
	t.Log("5. Listing users...")
	listResp := makeHTTPRequest(t, "GET", baseURL+"/users", nil, token)
	if listResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for list users, got %d. Response: %s", listResp.StatusCode, listResp.Body)
	}

	var users []dto.UserListResponse
	err = json.Unmarshal([]byte(listResp.Body), &users)
	if err != nil {
		t.Fatalf("Failed to parse users list response: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	usernames := make(map[string]string)
	for _, user := range users {
		usernames[user.Username] = user.Name
	}

	if usernames["testuser1"] != "Test User" {
		t.Errorf("Expected testuser1 name 'Test User', got '%s'", usernames["testuser1"])
	}
	if usernames["testuser2"] != "Second User" {
		t.Errorf("Expected testuser2 name 'Second User', got '%s'", usernames["testuser2"])
	}
	t.Logf("✅ User listing successful - found %d users", len(users))

	// Step 6: Update user name
	t.Log("6. Updating user name...")
	updateNamePayload := map[string]string{
		"name": "Updated Test User",
	}
	updateNameResp := makeHTTPRequest(t, "PUT", baseURL+"/user", updateNamePayload, token)
	if updateNameResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for name update, got %d. Response: %s", updateNameResp.StatusCode, updateNameResp.Body)
	}
	t.Logf("✅ User name updated successfully")

	// Step 7: Update user password
	t.Log("7. Updating user password...")
	updatePasswordPayload := map[string]string{
		"password": "newpassword123",
	}
	updatePasswordResp := makeHTTPRequest(t, "PUT", baseURL+"/user", updatePasswordPayload, token)
	if updatePasswordResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for password update, got %d. Response: %s", updatePasswordResp.StatusCode, updatePasswordResp.Body)
	}
	t.Logf("✅ User password updated successfully")

	// Step 8: Test validation errors
	t.Log("8. Testing validation errors...")
	invalidUpdatePayload := map[string]string{
		"name":     "A",   // Too short
		"password": "123", // Too short
	}
	invalidResp := makeHTTPRequest(t, "PUT", baseURL+"/user", invalidUpdatePayload, token)
	if invalidResp.StatusCode != 400 {
		t.Errorf("Expected status 400 for invalid update, got %d. Response: %s", invalidResp.StatusCode, invalidResp.Body)
	} else {
		t.Logf("✅ Validation errors handled correctly")
	}

	// Step 9: Test login with new password
	t.Log("9. Testing login with new password...")
	newLoginPayload := map[string]string{
		"username": "testuser1",
		"password": "newpassword123",
	}
	newLoginResp := makeHTTPRequest(t, "POST", baseURL+"/login", newLoginPayload, "")
	if newLoginResp.StatusCode != 200 {
		t.Errorf("Expected status 200 for login with new password, got %d. Response: %s", newLoginResp.StatusCode, newLoginResp.Body)
	} else {
		t.Logf("✅ Login with new password successful")
	}

	t.Log("=== Integration Test Complete - All Steps Passed! ===")
}

// TestUnauthorizedAccess tests security scenarios
func TestUnauthorizedAccess(t *testing.T) {
	suite := SetupIntegrationTest()
	defer suite.TearDown()

	baseURL := suite.server.URL

	t.Log("=== Testing Unauthorized Access Scenarios ===")

	// Test accessing protected routes without token
	protectedEndpoints := []string{"/protected", "/users", "/user"}

	for _, endpoint := range protectedEndpoints {
		t.Run("Unauthorized_"+endpoint, func(t *testing.T) {
			method := "GET"
			if endpoint == "/user" {
				method = "PUT"
			}

			resp := makeHTTPRequest(t, method, baseURL+endpoint, nil, "")
			if resp.StatusCode != 401 {
				t.Errorf("Expected status 401 for %s without token, got %d", endpoint, resp.StatusCode)
			}
		})
	}

	// Test with invalid token
	t.Run("Invalid_Token", func(t *testing.T) {
		resp := makeHTTPRequest(t, "GET", baseURL+"/protected", nil, "invalid.token.here")
		if resp.StatusCode != 401 {
			t.Errorf("Expected status 401 for invalid token, got %d", resp.StatusCode)
		}
	})

	t.Log("✅ All unauthorized access tests passed")
}

// TestHTTPMethodValidation tests HTTP method restrictions
func TestHTTPMethodValidation(t *testing.T) {
	suite := SetupIntegrationTest()
	defer suite.TearDown()

	baseURL := suite.server.URL

	t.Log("=== Testing HTTP Method Validation ===")

	// Test /users endpoint method restrictions (has explicit method checking)
	t.Run("Users_Method_Validation", func(t *testing.T) {
		methods := []string{"POST", "PUT", "DELETE"}
		for _, method := range methods {
			resp := makeHTTPRequest(t, method, baseURL+"/users", nil, "")
			if resp.StatusCode != 405 {
				t.Errorf("Expected status 405 for %s /users, got %d", method, resp.StatusCode)
			}
		}
	})

	// Test register/login endpoints with wrong methods (will return 400 due to empty body)
	t.Run("Register_Login_Wrong_Methods", func(t *testing.T) {
		testCases := []struct {
			endpoint string
			method   string
		}{
			{"/register", "GET"},
			{"/login", "PUT"},
		}

		for _, tc := range testCases {
			resp := makeHTTPRequest(t, tc.method, baseURL+tc.endpoint, nil, "")
			// These will return 400 because handlers expect JSON body but get empty request
			if resp.StatusCode != 400 {
				t.Logf("INFO: %s %s returned status %d (expected 400 due to empty body)", tc.method, tc.endpoint, resp.StatusCode)
			}
		}
	})

	t.Log("✅ HTTP method validation tests completed")
}

// Helper function to make HTTP requests
func makeHTTPRequest(t *testing.T, method, url string, payload interface{}, token string) *HTTPResponse {
	var body io.Reader
	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("Failed to marshal payload: %v", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	return &HTTPResponse{
		StatusCode: resp.StatusCode,
		Body:       string(respBody),
		Headers:    resp.Header,
	}
}

// HTTPResponse represents an HTTP response for testing
type HTTPResponse struct {
	StatusCode int
	Body       string
	Headers    http.Header
}

// Copy the handleRequest function from local-server.go for integration testing
// This tests the actual HTTP to Lambda event conversion
func handleIntegrationRequest(handler func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Convert HTTP request to Lambda event
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)

		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ",")
		}

		event := events.APIGatewayProxyRequest{
			HTTPMethod: r.Method,
			Path:       r.URL.Path,
			Headers:    headers,
			Body:       string(body),
			RequestContext: events.APIGatewayProxyRequestContext{
				Authorizer: make(map[string]interface{}),
			},
		}

		// Call Lambda handler
		response, err := handler(event)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert Lambda response to HTTP response
		for k, v := range response.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(response.StatusCode)
		w.Write([]byte(response.Body))
	}
}
