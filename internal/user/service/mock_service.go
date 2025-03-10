package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/linkeunid/hello-go/pkg/config"
)

// MockUserService implements the UserService interface with mock data
type mockUserService struct {
	cfg    *config.Config
	logger *zap.Logger
	users  map[string]*User // id -> user
}

// NewMockUserService creates a new mock user service
func NewMockUserService(cfg *config.Config, logger *zap.Logger) UserService {
	// Create some mock users
	mockUsers := map[string]*User{
		"00000000-0000-0000-0000-000000000001": {
			ID:        "00000000-0000-0000-0000-000000000001",
			Email:     "admin@example.com",
			Name:      "Admin User",
			CreatedAt: time.Now().Add(-30 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * 24 * time.Hour),
		},
		"00000000-0000-0000-0000-000000000002": {
			ID:        "00000000-0000-0000-0000-000000000002",
			Email:     "user@example.com",
			Name:      "Regular User",
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * 24 * time.Hour),
		},
		"00000000-0000-0000-0000-000000000003": {
			ID:        "00000000-0000-0000-0000-000000000003",
			Email:     "test@example.com",
			Name:      "Test User",
			CreatedAt: time.Now().Add(-1 * 24 * time.Hour),
			UpdatedAt: time.Now(),
		},
	}

	// Add more mock users
	for i := 4; i <= 20; i++ {
		id := "00000000-0000-0000-0000-00000000000" + string(rune('0'+i))
		if i >= 10 {
			id = "00000000-0000-0000-0000-0000000000" + string(rune('0'+i-10))
		}
		mockUsers[id] = &User{
			ID:        id,
			Email:     "user" + string(rune('0'+i)) + "@example.com",
			Name:      "User " + string(rune('0'+i)),
			CreatedAt: time.Now().Add(-time.Duration(i) * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-time.Duration(i/2) * 24 * time.Hour),
		}
	}

	return &mockUserService{
		cfg:    cfg,
		logger: logger,
		users:  mockUsers,
	}
}

// GetUser gets a user by ID
func (s *mockUserService) GetUser(ctx context.Context, id string) (*User, error) {
	s.logger.Debug("Mock: Getting user by ID", zap.String("user_id", id))

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Return a copy to prevent modification of internal state
	return &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// UpdateUser updates a user's information
func (s *mockUserService) UpdateUser(ctx context.Context, id, name, email string) (*User, error) {
	s.logger.Debug("Mock: Updating user",
		zap.String("user_id", id),
		zap.String("name", name),
		zap.String("email", email))

	user, exists := s.users[id]
	if !exists {
		return nil, ErrUserNotFound
	}

	// Check if email is already taken by another user
	for _, u := range s.users {
		if u.Email == email && u.ID != id {
			return nil, ErrUserAlreadyExists
		}
	}

	// Update user
	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()

	// Return a copy to prevent modification of internal state
	return &User{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

// DeleteUser deletes a user by ID
func (s *mockUserService) DeleteUser(ctx context.Context, id string) error {
	s.logger.Debug("Mock: Deleting user", zap.String("user_id", id))

	if _, exists := s.users[id]; !exists {
		return ErrUserNotFound
	}

	delete(s.users, id)
	return nil
}

// ListUsers returns a list of users
func (s *mockUserService) ListUsers(ctx context.Context, page, pageSize int) ([]*User, int, error) {
	s.logger.Debug("Mock: Listing users",
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	// Validate page and pageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// Convert map to slice
	var allUsers []*User
	for _, user := range s.users {
		// Create a copy to prevent modification of internal state
		allUsers = append(allUsers, &User{
			ID:        user.ID,
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		})
	}

	// Sort by creation date (newest first) - simplified for mock
	// In a real implementation, you'd use sort.Slice

	// Calculate total
	total := len(allUsers)

	// Calculate pagination
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		return []*User{}, total, nil
	}
	if end > total {
		end = total
	}

	return allUsers[start:end], total, nil
}

// Add error for email already taken
var ErrUserAlreadyExists = ErrUserNotFound
