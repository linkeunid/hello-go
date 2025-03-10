package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcLoggingInterceptor is a gRPC interceptor for logging requests
func GrpcLoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Create a logger for this request
		reqLogger := logger.With(
			zap.String("grpc_method", info.FullMethod),
		)

		reqLogger.Debug("gRPC request received", zap.Any("request", req))

		// Process the request
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		code := codes.OK
		if err != nil {
			st, ok := status.FromError(err)
			if ok {
				code = st.Code()
			} else {
				code = codes.Internal
			}
		}

		// Log the result
		if err != nil {
			reqLogger.Error("gRPC request failed",
				zap.Error(err),
				zap.String("code", code.String()),
				zap.Duration("duration", duration),
			)
		} else {
			reqLogger.Info("gRPC request completed",
				zap.String("code", code.String()),
				zap.Duration("duration", duration),
			)

			reqLogger.Debug("gRPC response", zap.Any("response", resp))
		}

		return resp, err
	}
}
