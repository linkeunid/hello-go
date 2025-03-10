package service

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/linkeunid/hello-go/pkg/config"
)

// MockAuthService implements the AuthService interface with mock data
type mockAuthService struct {
	cfg    *config.Config
	logger *zap.Logger
	users  map[string]*mockUser // email -> user
}

// mockUser represents a mock user
type mockUser struct {
	ID        string
	Email     string
	Password  string
	Name      string
	CreatedAt time.Time
}

// NewMockAuthService creates a new mock auth service
func NewMockAuthService(cfg *config.Config, logger *zap.Logger) AuthService {
	// Create some mock users
	users := map[string]*mockUser{
		"admin@example.com": {
			ID:        "00000000-0000-0000-0000-000000000001",
			Email:     "admin@example.com",
			Password:  "admin123", // In a real app, this would be hashed
			Name:      "Admin User",
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
		},
		"user@example.com": {
			ID:        "00000000-0000-0000-0000-000000000002",
			Email:     "user@example.com",
			Password:  "password123", // In a real app, this would be hashed
			Name:      "Regular User",
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
		},
		"test@example.com": {
			ID:        "00000000-0000-0000-0000-000000000003",
			Email:     "test@example.com",
			Password:  "test123", // In a real app, this would be hashed
			Name:      "Test User",
			CreatedAt: time.Now().Add(-1 * 24 * time.Hour),
		},
	}

	return &mockAuthService{
		cfg:    cfg,
		logger: logger,
		users:  users,
	}
}

// Authenticate authenticates a user with email and password
func (s *mockAuthService) Authenticate(ctx context.Context, email, password string) (string, error) {
	s.logger.Debug("Mock: Authenticating user", zap.String("email", email))

	// Find user by email
	user, exists := s.users[email]
	if !exists {
		return "", ErrInvalidCredentials
	}

	// Verify password (in real app, would use bcrypt.CompareHashAndPassword)
	if user.Password != password {
		return "", ErrInvalidCredentials
	}

	return user.ID, nil
}

// Register creates a new user
func (s *mockAuthService) Register(ctx context.Context, email, password, name string) (string, error) {
	s.logger.Debug("Mock: Registering new user", zap.String("email", email), zap.String("name", name))

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return "", ErrInvalidCredentials
	}

	// Check if user already exists
	if _, exists := s.users[email]; exists {
		return "", ErrUserAlreadyExists
	}

	// Check password strength (simple example)
	if len(password) < 6 {
		return "", ErrInvalidCredentials
	}

	// Create user
	userID := "mock-" + strings.ReplaceAll(email, "@", "-at-")
	s.users[email] = &mockUser{
		ID:        userID,
		Email:     email,
		Password:  password, // In a real app, this would be hashed
		Name:      name,
		CreatedAt: time.Now(),
	}

	return userID, nil
}

// ValidateToken validates a token and returns the user ID
func (s *mockAuthService) ValidateToken(ctx context.Context, tokenString string) (string, error) {
	s.logger.Debug("Mock: Validating token")

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Auth.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return "", ErrInvalidCredentials
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidCredentials
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidCredentials
	}

	// For mock service, let's allow any token with a user ID
	return userID, nil
}
