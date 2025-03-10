package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	// "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"

	"github.com/linkeunid/hello-go/pkg/config"
)

// AuthTokenValidator defines the interface for auth token validation
type AuthTokenValidator interface {
	ValidateToken(ctx context.Context, token string) (bool, string, error)
}

// AuthMiddleware is a middleware for authenticating HTTP requests
func AuthMiddleware(validator AuthTokenValidator, logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")

			// Validate token
			valid, userID, err := validator.ValidateToken(r.Context(), token)
			if err != nil {
				logger.Error("Error validating token", zap.Error(err))
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			if !valid {
				logger.Warn("Invalid token")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), "userID", userID)
			r = r.WithContext(ctx)

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// JWTValidator implements simple JWT validation without requiring auth client
type JWTValidator struct {
	JWTSecret string
	Logger    *zap.Logger
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator(cfg *config.Config, logger *zap.Logger) *JWTValidator {
	return &JWTValidator{
		JWTSecret: cfg.Auth.JWTSecret,
		Logger:    logger.Named("jwt_validator"),
	}
}

// ValidateToken validates a JWT token
func (v *JWTValidator) ValidateToken(ctx context.Context, tokenString string) (bool, string, error) {
	if tokenString == "" {
		return false, "", nil
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(v.JWTSecret), nil
	})

	if err != nil {
		v.Logger.Debug("Token validation failed", zap.Error(err))
		return false, "", nil
	}

	if !token.Valid {
		return false, "", nil
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, "", nil
	}

	// Get user ID from claims
	userID, ok := claims["sub"].(string)
	if !ok {
		return false, "", nil
	}

	return true, userID, nil
}

// ForwardAuthToken forwards the Authorization header from HTTP to gRPC metadata
func ForwardAuthToken(ctx context.Context, r *http.Request) metadata.MD {
	md := make(metadata.MD)
	if auth := r.Header.Get("Authorization"); auth != "" {
		md.Set("authorization", auth)
	}
	return md
}

// RegisterAuthMetadataAnnotator registers a metadata annotator for the gRPC gateway
// func RegisterAuthMetadataAnnotator(mux *runtime.ServeMux) {
// 	mux.SetMetadata(ForwardAuthToken)
// }
