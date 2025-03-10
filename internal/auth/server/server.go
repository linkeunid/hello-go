package server

import (
	"context"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// Update import path to use the generated code in api/gen/auth
	"github.com/linkeunid/hello-go/api/gen/auth"
	"github.com/linkeunid/hello-go/internal/auth/service"
	"github.com/linkeunid/hello-go/pkg/config"
)

// AuthServer implements the AuthService gRPC service
type AuthServer struct {
	auth.UnimplementedAuthServiceServer
	cfg     *config.Config
	service service.AuthService
	logger  *zap.Logger
}

// NewAuthServer creates a new AuthServer instance
func NewAuthServer(cfg *config.Config, logger *zap.Logger) *AuthServer {
	// Determine if we should use mock service
	useMock := os.Getenv("USE_MOCK_SERVICES") == "true"

	var svc service.AuthService
	if useMock {
		logger.Info("Using mock auth service")
		svc = service.NewMockAuthService(cfg, logger.Named("mock_auth_service"))
	} else {
		svc = service.NewAuthService(cfg, logger.Named("auth_service"))
	}

	return &AuthServer{
		cfg:     cfg,
		service: svc,
		logger:  logger.Named("auth_server"),
	}
}

// Login authenticates a user and returns a JWT token
func (s *AuthServer) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	// Check email and password (simplified for example)
	if req.Email == "" || req.Password == "" {
		s.logger.Warn("Login attempt with missing credentials",
			zap.String("email", req.Email))
		return nil, status.Error(codes.InvalidArgument, "email and password are required")
	}

	s.logger.Debug("Login attempt",
		zap.String("email", req.Email))

	// Authenticate user
	userID, err := s.service.Authenticate(ctx, req.Email, req.Password)
	if err != nil {
		s.logger.Warn("Authentication failed",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Generate JWT token
	token, err := s.generateToken(userID)
	if err != nil {
		s.logger.Error("Failed to generate token",
			zap.String("user_id", userID),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	s.logger.Info("User logged in successfully",
		zap.String("user_id", userID),
		zap.String("email", req.Email))

	return &auth.LoginResponse{
		Token:  token,
		UserId: userID,
	}, nil
}

// Register creates a new user account
func (s *AuthServer) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	// Validate request
	if req.Email == "" || req.Password == "" || req.Name == "" {
		s.logger.Warn("Registration attempt with missing fields",
			zap.String("email", req.Email),
			zap.String("name", req.Name))
		return nil, status.Error(codes.InvalidArgument, "email, password, and name are required")
	}

	s.logger.Debug("Registration attempt",
		zap.String("email", req.Email),
		zap.String("name", req.Name))

	// Register user
	userID, err := s.service.Register(ctx, req.Email, req.Password, req.Name)
	if err != nil {
		if err == service.ErrUserAlreadyExists {
			s.logger.Warn("User already exists during registration",
				zap.String("email", req.Email))
			return nil, status.Error(codes.AlreadyExists, "user already exists")
		}
		s.logger.Error("Failed to register user",
			zap.String("email", req.Email),
			zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to register user")
	}

	s.logger.Info("User registered successfully",
		zap.String("user_id", userID),
		zap.String("email", req.Email))

	return &auth.RegisterResponse{
		UserId: userID,
	}, nil
}

// ValidateToken validates a JWT token
func (s *AuthServer) ValidateToken(ctx context.Context, req *auth.ValidateTokenRequest) (*auth.ValidateTokenResponse, error) {
	// Validate token
	if req.Token == "" {
		s.logger.Warn("Token validation attempt with empty token")
		return nil, status.Error(codes.InvalidArgument, "token is required")
	}

	s.logger.Debug("Token validation attempt")

	// Parse token
	token, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			s.logger.Warn("Token with invalid signing method",
				zap.String("method", token.Method.Alg()))
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}
		return []byte(s.cfg.Auth.JWTSecret), nil
	})

	// Check for parsing errors
	if err != nil {
		s.logger.Debug("Invalid token during validation",
			zap.Error(err))
		return &auth.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Check if token is valid
	if !token.Valid {
		s.logger.Debug("Token validation failed")
		return &auth.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		s.logger.Warn("Failed to extract claims from token")
		return &auth.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		s.logger.Warn("Token missing user ID claim")
		return &auth.ValidateTokenResponse{
			Valid:  false,
			UserId: "",
		}, nil
	}

	s.logger.Debug("Token validated successfully",
		zap.String("user_id", userID))

	return &auth.ValidateTokenResponse{
		Valid:  true,
		UserId: userID,
	}, nil
}

// generateToken generates a JWT token for the given user ID
func (s *AuthServer) generateToken(userID string) (string, error) {
	// Create JWT claims
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(s.cfg.Auth.JWTExpiration).Unix(),
		"iat": time.Now().Unix(),
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(s.cfg.Auth.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
