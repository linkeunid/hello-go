package service

import (
	"context"
	"errors"

	"go.uber.org/zap"

	"github.com/linkeunid/hello-go/internal/auth/repository"
	"github.com/linkeunid/hello-go/pkg/config"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
)

// AuthService defines the interface for auth service operations
type AuthService interface {
	// Authenticate authenticates a user with email and password
	Authenticate(ctx context.Context, email, password string) (string, error)
	// Register creates a new user
	Register(ctx context.Context, email, password, name string) (string, error)
	// ValidateToken validates a token and returns the user ID
	ValidateToken(ctx context.Context, token string) (string, error)
}

// authService implements the AuthService interface
type authService struct {
	cfg    *config.Config
	repo   repository.AuthRepository
	logger *zap.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(cfg *config.Config, logger *zap.Logger) AuthService {
	return &authService{
		cfg:    cfg,
		repo:   repository.NewAuthRepository(cfg, logger.Named("auth_repository")),
		logger: logger,
	}
}

// Authenticate authenticates a user with email and password
func (s *authService) Authenticate(ctx context.Context, email, password string) (string, error) {
	s.logger.Debug("Authenticating user", zap.String("email", email))

	// Get user by email
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Debug("User not found during authentication",
			zap.String("email", email),
			zap.Error(err))
		return "", ErrInvalidCredentials
	}

	// Verify password
	if err := s.repo.CheckPassword(user.Password, password); err != nil {
		s.logger.Debug("Password verification failed",
			zap.String("email", email),
			zap.Error(err))
		return "", ErrInvalidCredentials
	}

	s.logger.Debug("User authenticated successfully",
		zap.String("email", email),
		zap.String("user_id", user.ID))

	return user.ID, nil
}

// Register creates a new user
func (s *authService) Register(ctx context.Context, email, password, name string) (string, error) {
	s.logger.Debug("Registering new user",
		zap.String("email", email),
		zap.String("name", name))

	// Check if user already exists
	exists, err := s.repo.UserExists(ctx, email)
	if err != nil {
		s.logger.Error("Error checking if user exists",
			zap.String("email", email),
			zap.Error(err))
		return "", err
	}

	if exists {
		s.logger.Debug("User already exists during registration",
			zap.String("email", email))
		return "", ErrUserAlreadyExists
	}

	// Create user (password hashing is handled in the repository)
	userID, err := s.repo.CreateUser(ctx, email, password, name)
	if err != nil {
		s.logger.Error("Error creating user",
			zap.String("email", email),
			zap.Error(err))
		return "", err
	}

	s.logger.Debug("User registered successfully",
		zap.String("email", email),
		zap.String("user_id", userID))

	return userID, nil
}

// ValidateToken validates a token and returns the user ID
func (s *authService) ValidateToken(ctx context.Context, token string) (string, error) {
	// This is handled in the server layer already, but we could add more logic here
	return "", nil
}
