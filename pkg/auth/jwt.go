package auth

import (
	"time"

	"github.com/hackmajoris/glad/pkg/config"
	pkgerrors "github.com/hackmajoris/glad/pkg/errors"
	"github.com/hackmajoris/glad/pkg/logger"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken Use shared authentication errors from pkg/errors
var (
	ErrInvalidToken = pkgerrors.ErrInvalidToken
)

// User interface for JWT token generation
type User interface {
	GetUsername() string
}

// JWTClaims represents the JWT claims
type JWTClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// TokenService handles JWT operations
type TokenService struct {
	secretKey []byte
	expiry    time.Duration
}

// NewTokenService creates a new TokenService
func NewTokenService(cfg *config.Config) *TokenService {
	log := logger.WithComponent("auth")

	if cfg.JWT.Secret == "default-secret-key" {
		log.Warn("Using default JWT secret - not suitable for production")
	} else {
		log.Info("JWT service initialized with custom secret")
	}

	return &TokenService{
		secretKey: []byte(cfg.JWT.Secret),
		expiry:    cfg.JWT.Expiry,
	}
}

// GenerateToken creates a new JWT token for the user
func (ts *TokenService) GenerateToken(user User) (string, error) {
	log := logger.WithComponent("auth").With("operation", "GenerateToken", "username", user.GetUsername())
	start := time.Now()

	log.Debug("Starting JWT token generation")

	expiry := time.Now().Add(ts.expiry)
	claims := JWTClaims{
		Username: user.GetUsername(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.GetUsername(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(ts.secretKey)
	if err != nil {
		log.Error("Failed to sign JWT token", "error", err.Error(), "duration", time.Since(start))
		return "", err
	}

	log.Info("JWT token generated successfully", "expires_at", expiry.Format(time.RFC3339), "duration", time.Since(start))
	return signedToken, nil
}

// ValidateToken validates and parses a JWT token
func (ts *TokenService) ValidateToken(tokenString string) (*JWTClaims, error) {
	log := logger.WithComponent("auth").With("operation", "ValidateToken")
	start := time.Now()

	log.Debug("Starting JWT token validation")

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Error("Unexpected signing method", "method", token.Header["alg"])
			return nil, pkgerrors.ErrInvalidToken
		}
		return ts.secretKey, nil
	})

	if err != nil {
		log.Error("Failed to parse JWT token", "error", err.Error(), "duration", time.Since(start))
		return nil, err
	}

	if !token.Valid {
		log.Error("Invalid JWT token", "duration", time.Since(start))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		log.Error("Invalid JWT token claims", "duration", time.Since(start))
		return nil, ErrInvalidToken
	}

	log = log.With("username", claims.Username)
	log.Info("JWT token validated successfully", "expires_at", claims.ExpiresAt.Time.Format(time.RFC3339), "duration", time.Since(start))
	return claims, nil
}
