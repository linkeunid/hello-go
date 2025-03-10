package client

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Update import path to use the generated code in api/gen/auth
	"github.com/linkeunid/hello-go/api/gen/auth"
	"github.com/linkeunid/hello-go/pkg/config"
	"github.com/linkeunid/hello-go/pkg/middleware"
)

// AuthClient is a client for the auth service
type AuthClient interface {
	// ValidateToken validates a token and returns the user ID
	ValidateToken(ctx context.Context, token string) (bool, string, error)
	// Close closes the gRPC connection
	Close() error
}

// authClient implements the AuthClient interface
type authClient struct {
	cfg    *config.Config
	client auth.AuthServiceClient
	conn   *grpc.ClientConn
	logger *zap.Logger
}

// NewAuthClient creates a new auth client
func NewAuthClient(cfg *config.Config, logger *zap.Logger) (AuthClient, error) {
	logger = logger.Named("auth_client")

	// Check if we should use mock client
	if os.Getenv("USE_MOCK_SERVICES") == "true" {
		logger.Info("Using mock auth client")
		return NewMockAuthClient(cfg, logger), nil
	}

	logger.Debug("Creating auth client",
		zap.Int("grpc_port", cfg.Auth.GRPCPort))

	// Set up a connection to the gRPC server with logging interceptor
	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", cfg.Auth.GRPCPort),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(middleware.GrpcClientLoggingInterceptor(logger)),
	)
	if err != nil {
		logger.Error("Failed to connect to auth service", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to auth service: %w", err)
	}

	logger.Debug("Connection to auth service established")

	// Create gRPC client
	client := auth.NewAuthServiceClient(conn)

	return &authClient{
		cfg:    cfg,
		client: client,
		conn:   conn,
		logger: logger,
	}, nil
}

// ValidateToken validates a token and returns the user ID
func (c *authClient) ValidateToken(ctx context.Context, token string) (bool, string, error) {
	// Don't log the actual token, just the first few characters
	tokenPreview := ""
	if len(token) > 8 {
		tokenPreview = token[:8] + "..."
	}

	c.logger.Debug("Validating token",
		zap.String("token_preview", tokenPreview))

	// Set timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Call gRPC method
	res, err := c.client.ValidateToken(ctx, &auth.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		c.logger.Error("Failed to validate token", zap.Error(err))
		return false, "", fmt.Errorf("failed to validate token: %w", err)
	}

	c.logger.Debug("Token validation result",
		zap.Bool("valid", res.Valid),
		zap.String("user_id", res.UserId))

	return res.Valid, res.UserId, nil
}

// Close closes the gRPC connection
func (c *authClient) Close() error {
	c.logger.Debug("Closing auth client connection")
	return c.conn.Close()
}
