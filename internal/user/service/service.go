package service

import (
	"context"
	"errors"
	"time"

	"go.uber.org/zap"

	"github.com/linkeunid/hello-go/internal/user/repository"
	"github.com/linkeunid/hello-go/pkg/config"
)

// Common errors
var (
	ErrUserNotFound = errors.New("user not found")
)

// User represents a user in the service layer
type User struct {
	ID        string
	Email     string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserService defines the interface for user service operations
type UserService interface {
	// GetUser gets a user by ID
	GetUser(ctx context.Context, id string) (*User, error)
	// UpdateUser updates a user's information
	UpdateUser(ctx context.Context, id, name, email string) (*User, error)
	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id string) error
	// ListUsers returns a list of users
	ListUsers(ctx context.Context, page, pageSize int) ([]*User, int, error)
}

// userService implements the UserService interface
type userService struct {
	cfg    *config.Config
	repo   repository.UserRepository
	logger *zap.Logger
}

// NewUserService creates a new user service
func NewUserService(cfg *config.Config, logger *zap.Logger) UserService {
	return &userService{
		cfg:    cfg,
		repo:   repository.NewUserRepository(cfg, logger.Named("user_repository")),
		logger: logger,
	}
}

// GetUser gets a user by ID
func (s *userService) GetUser(ctx context.Context, id string) (*User, error) {
	s.logger.Debug("Getting user by ID", zap.String("user_id", id))

	// Get user
	user, err := s.repo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Debug("User not found", zap.String("user_id", id))
			return nil, ErrUserNotFound
		}
		s.logger.Error("Error getting user",
			zap.String("user_id", id),
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("User found", zap.String("user_id", id))

	// Map to service layer user
	return &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUser updates a user's information
func (s *userService) UpdateUser(ctx context.Context, id, name, email string) (*User, error) {
	s.logger.Debug("Updating user",
		zap.String("user_id", id),
		zap.String("name", name),
		zap.String("email", email))

	// Update user
	user, err := s.repo.UpdateUser(ctx, id, name, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Debug("User not found during update", zap.String("user_id", id))
			return nil, ErrUserNotFound
		}
		s.logger.Error("Error updating user",
			zap.String("user_id", id),
			zap.Error(err))
		return nil, err
	}

	s.logger.Debug("User updated successfully", zap.String("user_id", id))

	// Map to service layer user
	return &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// DeleteUser deletes a user by ID
func (s *userService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Debug("Deleting user", zap.String("user_id", id))

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			s.logger.Debug("User not found during delete", zap.String("user_id", id))
			return ErrUserNotFound
		}
		s.logger.Error("Error deleting user",
			zap.String("user_id", id),
			zap.Error(err))
		return err
	}

	s.logger.Debug("User deleted successfully", zap.String("user_id", id))
	return nil
}

// ListUsers returns a list of users
func (s *userService) ListUsers(ctx context.Context, page, pageSize int) ([]*User, int, error) {
	// Validate page and pageSize
	if page < 1 {
		page = 1
		s.logger.Debug("Adjusted page to 1", zap.Int("requested_page", page))
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
		s.logger.Debug("Adjusted page size to 10", zap.Int("requested_page_size", pageSize))
	}

	s.logger.Debug("Listing users",
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	// Get users
	users, total, err := s.repo.ListUsers(ctx, page, pageSize)
	if err != nil {
		s.logger.Error("Error listing users", zap.Error(err))
		return nil, 0, err
	}

	// Map to service layer users
	result := make([]*User, len(users))
	for i, user := range users {
		result[i] = &User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
	}

	s.logger.Debug("Listed users successfully",
		zap.Int("count", len(result)),
		zap.Int("total", total))

	return result, total, nil
}
