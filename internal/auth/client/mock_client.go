package client

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/linkeunid/hello-go/pkg/config"
)

// MockAuthClient implements the AuthClient interface with mock data
type mockAuthClient struct {
	cfg    *config.Config
	logger *zap.Logger
}

// NewMockAuthClient creates a new mock auth client
func NewMockAuthClient(cfg *config.Config, logger *zap.Logger) AuthClient {
	return &mockAuthClient{
		cfg:    cfg,
		logger: logger.Named("mock_auth_client"),
	}
}

// ValidateToken validates a token and returns the user ID
func (c *mockAuthClient) ValidateToken(ctx context.Context, token string) (bool, string, error) {
	// For testing, allow bypassing validation
	if os.Getenv("BYPASS_AUTH") == "true" {
		c.logger.Warn("Authentication bypassed in mock mode")
		return true, "mock-user-id", nil
	}

	// Don't log the actual token, just the first few characters
	tokenPreview := ""
	if len(token) > 8 {
		tokenPreview = token[:8] + "..."
	}

	c.logger.Debug("Mock: Validating token",
		zap.String("token_preview", tokenPreview))

	// Basic token validation logic
	if token == "" {
		return false, "", nil
	}

	// Parse token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(c.cfg.Auth.JWTSecret), nil
	})

	if err != nil {
		c.logger.Debug("Token validation failed", zap.Error(err))
		return false, "", nil
	}

	if !parsedToken.Valid {
		return false, "", nil
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return false, "", nil
	}

	// Simple token for testing
	if token == "mock-token" {
		return true, "mock-user-id", nil
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return false, "", nil
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			return false, "", nil
		}
	}

	c.logger.Debug("Token validated successfully",
		zap.String("user_id", userID))

	return true, userID, nil
}

// Close closes the mock auth client
func (c *mockAuthClient) Close() error {
	c.logger.Debug("Closing mock auth client")
	return nil
}
