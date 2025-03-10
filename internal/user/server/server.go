package server

import (
	"context"
	"os"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	// Update import path to use the generated code in api/gen/user
	"github.com/linkeunid/hello-go/api/gen/user"
	"github.com/linkeunid/hello-go/internal/auth/client"
	"github.com/linkeunid/hello-go/internal/user/service"
	"github.com/linkeunid/hello-go/pkg/config"
	"github.com/linkeunid/hello-go/pkg/middleware"
)

// UserServer implements the UserService gRPC service
type UserServer struct {
	user.UnimplementedUserServiceServer
	cfg          *config.Config
	service      service.UserService
	authClient   client.AuthClient
	jwtValidator *middleware.JWTValidator
	logger       *zap.Logger
	useMockMode  bool
}

// NewUserServer creates a new UserServer instance
func NewUserServer(cfg *config.Config, logger *zap.Logger) *UserServer {
	// Determine if we should use mock service
	useMock := os.Getenv("USE_MOCK_SERVICES") == "true"

	var authClient client.AuthClient
	var err error

	// Create JWT validator for bypass scenarios
	jwtValidator := middleware.NewJWTValidator(cfg, logger)

	// Only create auth client if not using bypass mode
	if !(useMock && os.Getenv("BYPASS_AUTH") == "true") {
		authClient, err = client.NewAuthClient(cfg, logger.Named("auth_client"))
		if err != nil {
			// Log error and panic as this is a critical dependency
			logger.Fatal("Failed to create auth client", zap.Error(err))
		}
	}

	var svc service.UserService
	if useMock {
		logger.Info("Using mock user service")
		svc = service.NewMockUserService(cfg, logger.Named("mock_user_service"))
	} else {
		svc = service.NewUserService(cfg, logger.Named("user_service"))
	}

	return &UserServer{
		cfg:          cfg,
		service:      svc,
		authClient:   authClient,
		jwtValidator: jwtValidator,
		logger:       logger.Named("user_server"),
		useMockMode:  useMock,
	}
}

// GetUser returns a user by ID
func (s *UserServer) GetUser(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	// Authenticate request - can be bypassed in mock mode
	userID, err := s.authenticateOrBypass(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("GetUser request",
		zap.String("requested_user_id", req.Id),
		zap.String("requester_user_id", userID))

	// Get user
	userData, err := s.service.GetUser(ctx, req.Id)
	if err != nil {
		if err == service.ErrUserNotFound {
			s.logger.Warn("User not found",
				zap.String("user_id", req.Id))
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.logger.Error("Failed to get user",
			zap.String("user_id", req.Id),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get user")
	}

	s.logger.Info("User retrieved successfully",
		zap.String("user_id", req.Id))

	// Return response
	return &user.GetUserResponse{
		User: &user.User{
			Id:        userData.ID,
			Email:     userData.Email,
			Name:      userData.Name,
			CreatedAt: userData.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: userData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

// UpdateUser updates a user's information
func (s *UserServer) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	// Authenticate request - can be bypassed in mock mode
	userID, err := s.authenticateOrBypass(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("UpdateUser request",
		zap.String("user_id", req.Id),
		zap.String("requester_user_id", userID),
		zap.String("new_name", req.Name),
		zap.String("new_email", req.Email))

	// Only allow users to update their own information
	if userID != req.Id && userID != "mock-bypass" {
		s.logger.Warn("Permission denied: user attempting to update another user",
			zap.String("requester_id", userID),
			zap.String("target_id", req.Id))
		return nil, status.Error(codes.PermissionDenied, "cannot update other users")
	}

	// Update user
	userData, err := s.service.UpdateUser(ctx, req.Id, req.Name, req.Email)
	if err != nil {
		if err == service.ErrUserNotFound {
			s.logger.Warn("User not found during update",
				zap.String("user_id", req.Id))
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.logger.Error("Failed to update user",
			zap.String("user_id", req.Id),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update user")
	}

	s.logger.Info("User updated successfully",
		zap.String("user_id", req.Id))

	// Return response
	return &user.UpdateUserResponse{
		User: &user.User{
			Id:        userData.ID,
			Email:     userData.Email,
			Name:      userData.Name,
			CreatedAt: userData.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: userData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		},
	}, nil
}

// DeleteUser deletes a user by ID
func (s *UserServer) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	// Authenticate request - can be bypassed in mock mode
	userID, err := s.authenticateOrBypass(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("DeleteUser request",
		zap.String("user_id", req.Id),
		zap.String("requester_user_id", userID))

	// Only allow users to delete their own account
	if userID != req.Id && userID != "mock-bypass" {
		s.logger.Warn("Permission denied: user attempting to delete another user",
			zap.String("requester_id", userID),
			zap.String("target_id", req.Id))
		return nil, status.Error(codes.PermissionDenied, "cannot delete other users")
	}

	// Delete user
	err = s.service.DeleteUser(ctx, req.Id)
	if err != nil {
		if err == service.ErrUserNotFound {
			s.logger.Warn("User not found during deletion",
				zap.String("user_id", req.Id))
			return nil, status.Error(codes.NotFound, "user not found")
		}
		s.logger.Error("Failed to delete user",
			zap.String("user_id", req.Id),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to delete user")
	}

	s.logger.Info("User deleted successfully",
		zap.String("user_id", req.Id))

	// Return response
	return &user.DeleteUserResponse{
		Success: true,
	}, nil
}

// ListUsers returns a list of users
func (s *UserServer) ListUsers(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	// Authenticate request - can be bypassed in mock mode
	userID, err := s.authenticateOrBypass(ctx)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("ListUsers request",
		zap.String("requester_user_id", userID),
		zap.Int32("page", req.Page),
		zap.Int32("page_size", req.PageSize))

	// List users
	users, total, err := s.service.ListUsers(ctx, int(req.Page), int(req.PageSize))
	if err != nil {
		s.logger.Error("Failed to list users", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to list users")
	}

	// Convert to proto users
	protoUsers := make([]*user.User, len(users))
	for i, userData := range users {
		protoUsers[i] = &user.User{
			Id:        userData.ID,
			Email:     userData.Email,
			Name:      userData.Name,
			CreatedAt: userData.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: userData.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	s.logger.Info("Users listed successfully",
		zap.Int("count", len(users)),
		zap.Int("total", total))

	// Return response
	return &user.ListUsersResponse{
		Users: protoUsers,
		Total: int32(total),
	}, nil
}

// authenticateOrBypass authenticates the request and returns the user ID
// If USE_MOCK_SERVICES is true and BYPASS_AUTH is true, it will bypass authentication
func (s *UserServer) authenticateOrBypass(ctx context.Context) (string, error) {
	// Check if we should bypass authentication in mock mode
	if s.useMockMode && os.Getenv("BYPASS_AUTH") == "true" {
		s.logger.Warn("Bypassing authentication in mock mode")
		return "mock-bypass", nil
	}

	return s.authenticate(ctx)
}

// authenticate authenticates the request and returns the user ID
func (s *UserServer) authenticate(ctx context.Context) (string, error) {
	// Get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		s.logger.Warn("Missing metadata in request")
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}

	// Get authorization token
	values := md.Get("authorization")
	if len(values) == 0 {
		s.logger.Warn("Missing authorization token")
		return "", status.Error(codes.Unauthenticated, "missing authorization token")
	}

	// Remove "Bearer " prefix
	token := values[0]
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Validate token using authClient or local JWT validator
	var valid bool
	var userID string
	var err error

	if s.authClient != nil {
		valid, userID, err = s.authClient.ValidateToken(ctx, token)
	} else {
		// Use local JWT validator when auth client is not available
		valid, userID, err = s.jwtValidator.ValidateToken(ctx, token)
	}

	if err != nil {
		s.logger.Error("Failed to validate token", zap.Error(err))
		return "", status.Error(codes.Internal, "failed to validate token")
	}

	if !valid {
		s.logger.Warn("Invalid token")
		return "", status.Error(codes.Unauthenticated, "invalid token")
	}

	return userID, nil
}
