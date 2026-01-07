# Integration Tests for GLAD Stack

This directory contains integration tests for the GLAD Stack application.

## Test Files

### `integration_cognito_test.go` ✅ **CURRENT**
Tests for Cognito-based authentication (current implementation).

**What it tests:**
- Cognito claims extraction from API Gateway authorizer context
- Protected routes with Cognito authentication
- User profile operations with Cognito users
- Unauthorized access scenarios
- Mock Post Confirmation Lambda trigger behavior

**Key Test Cases:**
1. **TestCognitoUserJourney** - Complete user journey with Cognito authentication
   - Simulates users created by Post Confirmation trigger
   - Tests protected route access with Cognito claims
   - Tests profile retrieval and updates
   - Tests user listing

2. **TestCognitoClaimsExtraction** - Cognito claims extraction
   - Tests extraction of username, email, and sub from authorizer context
   - Tests behavior with missing or incomplete claims

3. **TestUnauthorizedAccessCognito** - Security tests
   - Verifies unauthorized requests are rejected
   - Tests missing Cognito claims handling

### `integration_test.go` ⚠️ **DEPRECATED**
Tests for the old custom JWT authentication system.

**Status:** Deprecated but kept for reference. The custom JWT authentication has been replaced by Amazon Cognito.

## Running Integration Tests

### Run all integration tests:
```bash
cd cmd/glad
go test -v -tags=integration ./...
```

### Run specific test:
```bash
# Run Cognito tests
go test -v -tags=integration -run TestCognito

# Run specific test case
go test -v -tags=integration -run TestCognitoUserJourney
```

### Run with coverage:
```bash
go test -v -tags=integration -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Architecture

### Mock Infrastructure
The tests use **mock repositories** (`database.NewMockRepository()`) to simulate DynamoDB without requiring actual AWS infrastructure.

### Cognito Simulation
Since Cognito authentication happens at the **API Gateway level**, the tests simulate this by:
1. Creating a mock HTTP server
2. Converting HTTP requests to Lambda `APIGatewayProxyRequest` events
3. Populating the `RequestContext.Authorizer` with Cognito claims
4. Calling the Lambda handlers directly

### Key Helper Functions

**`createMockCognitoUser(username, email, name)`**
- Creates a user as if created by the Cognito Post Confirmation trigger
- Sets empty `PasswordHash` (Cognito manages passwords)
- Properly initializes DynamoDB keys

**`createCognitoAuthorizerContext(username, email, sub)`**
- Creates a mock Cognito authorizer context
- Includes all required Cognito claims: `sub`, `cognito:username`, `email`

**`makeCognitoHTTPRequest(method, url, payload, cognitoContext)`**
- Makes HTTP request with Cognito context
- Passes Cognito claims via custom header (extracted by test handler)

**`handleCognitoRequest(handler)`**
- Converts HTTP requests to Lambda events
- Populates `RequestContext.Authorizer` from test header
- Simulates API Gateway Cognito authorizer behavior

## Test Coverage

The Cognito integration tests cover:
- ✅ Cognito claims extraction (`pkg/auth/cognito.go`)
- ✅ Protected route access with valid Cognito authentication
- ✅ User profile CRUD operations
- ✅ Unauthorized access rejection
- ✅ Missing claims error handling
- ✅ Handler integration with Cognito users

## Differences from Production

### In Tests:
- Cognito claims passed via custom HTTP header (`X-Cognito-Context`)
- Mock DynamoDB repository (in-memory)
- Users manually created to simulate Post Confirmation trigger

### In Production:
- API Gateway validates Cognito tokens and populates authorizer context
- Real DynamoDB tables
- Post Confirmation Lambda automatically creates users after email verification

## Future Enhancements

Potential additions to integration tests:
- [ ] Skills CRUD operations with Cognito authentication
- [ ] User directory search with Cognito users
- [ ] Multi-user skill endorsement scenarios
- [ ] Comprehensive error handling tests
- [ ] Performance/load testing scenarios

## Notes

- **No actual Cognito calls**: Tests don't call Cognito APIs (signup, login, etc.)
- **No API Gateway**: Tests directly invoke Lambda handlers with mocked events
- **Fast execution**: All tests run in-memory without external dependencies
- **Isolation**: Each test creates its own mock repository

## Related Documentation

- [Cognito Integration Plan](../../docs/cognito-angular-integration-plan.md)
- [Main README](../../README.md)
- Backend API documentation (to be added)
