package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// GrpcClientLoggingInterceptor is a gRPC client interceptor for logging requests
func GrpcClientLoggingInterceptor(logger *zap.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()

		// Create a logger for this request
		reqLogger := logger.With(
			zap.String("grpc_method", method),
		)

		reqLogger.Debug("gRPC client request", zap.Any("request", req))

		// Process the request
		err := invoker(ctx, method, req, reply, cc, opts...)

		// Calculate duration
		duration := time.Since(start)

		// Get status code
		st, _ := status.FromError(err)

		// Log the result
		if err != nil {
			reqLogger.Error("gRPC client request failed",
				zap.Error(err),
				zap.String("code", st.Code().String()),
				zap.Duration("duration", duration),
			)
		} else {
			reqLogger.Debug("gRPC client request completed",
				zap.String("code", st.Code().String()),
				zap.Duration("duration", duration),
				zap.Any("response", reply),
			)
		}

		return err
	}
}
