package auth

import (
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

// CognitoClaims represents user claims from Cognito JWT token
// These claims are extracted from the API Gateway authorizer context
// after the Cognito User Pools Authorizer validates the token
type CognitoClaims struct {
	Sub      string // Cognito user ID (UUID) - unique identifier
	Username string // Cognito username or email used for login
	Email    string // User's email address
}

// ExtractCognitoClaimsFromRequest extracts Cognito user claims from API Gateway request context
// The Cognito authorizer populates the request.RequestContext.Authorizer map with claims
// from the validated JWT token.
//
// Example authorizer context structure:
//
//	{
//	  "claims": {
//	    "sub": "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111",
//	    "cognito:username": "johndoe",
//	    "email": "john@example.com"
//	  }
//	}
func ExtractCognitoClaimsFromRequest(request events.APIGatewayProxyRequest) (*CognitoClaims, error) {
	authorizer := request.RequestContext.Authorizer

	if authorizer == nil || len(authorizer) == 0 {
		return nil, errors.New("missing authorizer context - request may not be authenticated")
	}

	claims := &CognitoClaims{}

	// Extract 'sub' claim (Cognito user ID - UUID)
	// This is the unique, immutable identifier for the user
	if sub, ok := authorizer["sub"].(string); ok {
		claims.Sub = sub
	} else {
		return nil, errors.New("missing 'sub' claim in authorizer context")
	}

	// Extract username
	// Cognito can use either "cognito:username" or "username" depending on configuration
	// Try "cognito:username" first (most common), fallback to "username"
	if username, ok := authorizer["cognito:username"].(string); ok {
		claims.Username = username
	} else if username, ok := authorizer["username"].(string); ok {
		claims.Username = username
	} else {
		return nil, errors.New("missing username claim in authorizer context")
	}

	// Extract email (optional but typically present)
	if email, ok := authorizer["email"].(string); ok {
		claims.Email = email
	}

	return claims, nil
}

// String returns a string representation of the Cognito claims
// Useful for logging and debugging
func (c *CognitoClaims) String() string {
	return fmt.Sprintf("CognitoClaims{Sub: %s, Username: %s, Email: %s}", c.Sub, c.Username, c.Email)
}

// Validate checks if the claims contain the minimum required information
func (c *CognitoClaims) Validate() error {
	if c.Sub == "" {
		return errors.New("sub (user ID) is required")
	}
	if c.Username == "" {
		return errors.New("username is required")
	}
	return nil
}
