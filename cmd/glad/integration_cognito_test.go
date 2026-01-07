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

	"github.com/hackmajoris/glad-stack/cmd/glad/internal/database"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/dto"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/handler"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/models"
	"github.com/hackmajoris/glad-stack/cmd/glad/internal/service"
	"github.com/hackmajoris/glad-stack/pkg/auth"

	"github.com/aws/aws-lambda-go/events"
)

// CognitoIntegrationTestSuite represents the test environment for Cognito integration tests
type CognitoIntegrationTestSuite struct {
	userRepo   database.UserRepository
	skillsRepo database.SkillRepository
	apiHandler *handler.Handler
	server     *httptest.Server
}

// SetupCognitoIntegrationTest creates a test environment for Cognito integration
func SetupCognitoIntegrationTest() *CognitoIntegrationTestSuite {
	// Create a single mock repository that implements all interfaces
	mockRepo := database.NewMockRepository()

	// Note: TokenService is no longer used for authentication, but UserService still needs it for backwards compatibility
	tokenService := auth.NewTokenService(testConfig())

	userService := service.NewUserService(mockRepo, tokenService)
	skillsService := service.NewSkillService(mockRepo, mockRepo, mockRepo)
	apiHandler := handler.New(userService, skillsService)

	// Create HTTP server with Cognito-authenticated routes
	mux := http.NewServeMux()

	// All routes now assume Cognito authentication via API Gateway
	// No /register or /login endpoints - those are handled by Cognito
	mux.HandleFunc("/users/me", handleCognitoRequest(apiHandler.GetCurrentUser))
	mux.HandleFunc("/users/me/update", handleCognitoRequest(apiHandler.UpdateUser))
	mux.HandleFunc("/users", handleCognitoRequest(apiHandler.ListUsers))
	mux.HandleFunc("/protected", handleCognitoRequest(apiHandler.Protected))

	server := httptest.NewServer(mux)

	return &CognitoIntegrationTestSuite{
		userRepo:   mockRepo,
		skillsRepo: mockRepo,
		apiHandler: apiHandler,
		server:     server,
	}
}

// TearDown cleans up the test environment
func (suite *CognitoIntegrationTestSuite) TearDown() {
	suite.server.Close()
}

// TestCognitoUserJourney tests the user journey with Cognito authentication
func TestCognitoUserJourney(t *testing.T) {
	suite := SetupCognitoIntegrationTest()
	defer suite.TearDown()

	baseURL := suite.server.URL

	t.Log("=== Testing Cognito-Authenticated User Journey ===")

	// Step 1: Create mock users in DynamoDB (simulating Post Confirmation Lambda trigger)
	t.Log("1. Setting up test users (simulating Cognito Post Confirmation trigger)...")
	testUser1 := createMockCognitoUser("testuser1", "test1@example.com", "Test User 1")
	testUser2 := createMockCognitoUser("testuser2", "test2@example.com", "Test User 2")

	err := suite.userRepo.CreateUser(testUser1)
	if err != nil {
		t.Fatalf("Failed to create test user 1: %v", err)
	}

	err = suite.userRepo.CreateUser(testUser2)
	if err != nil {
		t.Fatalf("Failed to create test user 2: %v", err)
	}
	t.Logf("✅ Test users created in DynamoDB")

	// Step 2: Test protected route with Cognito claims
	t.Log("2. Testing protected route with Cognito authentication...")
	cognitoContext := createCognitoAuthorizerContext("testuser1", "test1@example.com", "cognito-sub-123")
	protectedResp := makeCognitoHTTPRequest(t, "GET", baseURL+"/protected", nil, cognitoContext)

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
	t.Logf("✅ Protected route access successful with Cognito claims")

	// Step 3: Get current user profile
	t.Log("3. Getting current user profile...")
	meResp := makeCognitoHTTPRequest(t, "GET", baseURL+"/users/me", nil, cognitoContext)

	if meResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for /users/me, got %d. Response: %s", meResp.StatusCode, meResp.Body)
	}

	var currentUser dto.CurrentUserResponse
	err = json.Unmarshal([]byte(meResp.Body), &currentUser)
	if err != nil {
		t.Fatalf("Failed to parse current user response: %v", err)
	}

	if currentUser.Username != "testuser1" {
		t.Errorf("Expected username testuser1, got %s", currentUser.Username)
	}
	if currentUser.Name != "Test User 1" {
		t.Errorf("Expected name 'Test User 1', got '%s'", currentUser.Name)
	}
	t.Logf("✅ Current user profile retrieved successfully")

	// Step 4: List all users
	t.Log("4. Listing all users...")
	listResp := makeCognitoHTTPRequest(t, "GET", baseURL+"/users", nil, cognitoContext)

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

	if usernames["testuser1"] != "Test User 1" {
		t.Errorf("Expected testuser1 name 'Test User 1', got '%s'", usernames["testuser1"])
	}
	if usernames["testuser2"] != "Test User 2" {
		t.Errorf("Expected testuser2 name 'Test User 2', got '%s'", usernames["testuser2"])
	}
	t.Logf("✅ User listing successful - found %d users", len(users))

	// Step 5: Update user profile
	t.Log("5. Updating user profile...")
	updatePayload := dto.UpdateUserRequest{
		Name: stringPtr("Updated Test User"),
	}
	updateResp := makeCognitoHTTPRequest(t, "PUT", baseURL+"/users/me/update", updatePayload, cognitoContext)

	if updateResp.StatusCode != 200 {
		t.Fatalf("Expected status 200 for profile update, got %d. Response: %s", updateResp.StatusCode, updateResp.Body)
	}
	t.Logf("✅ User profile updated successfully")

	// Step 6: Verify update by getting profile again
	t.Log("6. Verifying profile update...")
	verifyResp := makeCognitoHTTPRequest(t, "GET", baseURL+"/users/me", nil, cognitoContext)

	var updatedUser dto.CurrentUserResponse
	err = json.Unmarshal([]byte(verifyResp.Body), &updatedUser)
	if err != nil {
		t.Fatalf("Failed to parse updated user response: %v", err)
	}

	if updatedUser.Name != "Updated Test User" {
		t.Errorf("Expected updated name 'Updated Test User', got '%s'", updatedUser.Name)
	}
	t.Logf("✅ Profile update verified")

	t.Log("=== Cognito Integration Test Complete - All Steps Passed! ===")
}

// TestCognitoClaimsExtraction tests the Cognito claims extraction
func TestCognitoClaimsExtraction(t *testing.T) {
	suite := SetupCognitoIntegrationTest()
	defer suite.TearDown()

	baseURL := suite.server.URL

	t.Log("=== Testing Cognito Claims Extraction ===")

	// Create test user
	testUser := createMockCognitoUser("claimstest", "claims@example.com", "Claims Test User")
	err := suite.userRepo.CreateUser(testUser)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test with complete Cognito claims
	t.Run("Complete_Claims", func(t *testing.T) {
		cognitoContext := createCognitoAuthorizerContext("claimstest", "claims@example.com", "cognito-sub-456")
		resp := makeCognitoHTTPRequest(t, "GET", baseURL+"/protected", nil, cognitoContext)

		if resp.StatusCode != 200 {
			t.Errorf("Expected status 200 with complete claims, got %d. Response: %s", resp.StatusCode, resp.Body)
		}

		var response map[string]interface{}
		json.Unmarshal([]byte(resp.Body), &response)

		if response["username"] != "claimstest" {
			t.Errorf("Expected username claimstest from claims, got %v", response["username"])
		}
	})

	// Test with missing claims (should fail)
	t.Run("Missing_Claims", func(t *testing.T) {
		emptyContext := make(map[string]interface{})
		resp := makeCognitoHTTPRequest(t, "GET", baseURL+"/protected", nil, emptyContext)

		if resp.StatusCode != 401 {
			t.Errorf("Expected status 401 with missing claims, got %d. Response: %s", resp.StatusCode, resp.Body)
		}
	})

	t.Log("✅ Cognito claims extraction tests passed")
}

// TestUnauthorizedAccessCognito tests unauthorized access with Cognito
func TestUnauthorizedAccessCognito(t *testing.T) {
	suite := SetupCognitoIntegrationTest()
	defer suite.TearDown()

	baseURL := suite.server.URL

	t.Log("=== Testing Unauthorized Access with Cognito ===")

	// Test accessing protected routes without Cognito claims
	// Note: /users is public and doesn't require authentication
	protectedEndpoints := []string{"/protected", "/users/me"}

	for _, endpoint := range protectedEndpoints {
		t.Run("Unauthorized_"+endpoint, func(t *testing.T) {
			// Empty authorizer context simulates missing Cognito authentication
			emptyContext := make(map[string]interface{})
			resp := makeCognitoHTTPRequest(t, "GET", baseURL+endpoint, nil, emptyContext)

			if resp.StatusCode != 401 {
				t.Errorf("Expected status 401 for %s without Cognito claims, got %d", endpoint, resp.StatusCode)
			}
		})
	}

	t.Log("✅ All unauthorized access tests passed")
}

// Helper Functions

// createMockCognitoUser creates a mock user as if created by Post Confirmation trigger
func createMockCognitoUser(username, email, name string) *models.User {
	now := time.Now()
	user := &models.User{
		Username:     username,
		Email:        email,
		Name:         name,
		PasswordHash: "", // Empty for Cognito users
		CreatedAt:    now,
		UpdatedAt:    now,
		EntityType:   "User",
	}
	user.SetKeys()
	return user
}

// createCognitoAuthorizerContext creates a mock Cognito authorizer context
func createCognitoAuthorizerContext(username, email, sub string) map[string]interface{} {
	return map[string]interface{}{
		"sub":              sub,
		"cognito:username": username,
		"username":         username, // Some contexts use this format
		"email":            email,
	}
}

// makeCognitoHTTPRequest makes an HTTP request with Cognito authorizer context
func makeCognitoHTTPRequest(t *testing.T, method, url string, payload interface{}, cognitoContext map[string]interface{}) *HTTPResponse {
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

	// Add Cognito context as a custom header (will be extracted by handleCognitoRequest)
	if cognitoContext != nil {
		contextJSON, _ := json.Marshal(cognitoContext)
		req.Header.Set("X-Cognito-Context", string(contextJSON))
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

// handleCognitoRequest converts HTTP requests to Lambda events with Cognito authorizer context
func handleCognitoRequest(handler func(events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Convert HTTP request to Lambda event
		body := make([]byte, r.ContentLength)
		if r.ContentLength > 0 {
			r.Body.Read(body)
		}

		headers := make(map[string]string)
		for k, v := range r.Header {
			headers[k] = strings.Join(v, ",")
		}

		// Extract Cognito context from custom header (for testing)
		authorizerContext := make(map[string]interface{})
		if cognitoContextJSON := r.Header.Get("X-Cognito-Context"); cognitoContextJSON != "" {
			json.Unmarshal([]byte(cognitoContextJSON), &authorizerContext)
		}

		event := events.APIGatewayProxyRequest{
			HTTPMethod: r.Method,
			Path:       r.URL.Path,
			Headers:    headers,
			Body:       string(body),
			RequestContext: events.APIGatewayProxyRequestContext{
				Authorizer: authorizerContext,
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

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
