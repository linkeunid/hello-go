package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/linkeunid/hello-go/pkg/config"
	"github.com/linkeunid/hello-go/pkg/logger"
	"github.com/linkeunid/hello-go/pkg/middleware"

	// Update import path to use the generated code in api/gen/auth
	authpb "github.com/linkeunid/hello-go/api/gen/auth"
	"github.com/linkeunid/hello-go/internal/auth/server"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := logger.NewLogger(cfg)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	log.Info("Starting auth service",
		zap.Int("http_port", cfg.Auth.ServicePort),
		zap.Int("grpc_port", cfg.Auth.GRPCPort))

	// Initialize gRPC server
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Auth.GRPCPort))
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	// Create gRPC server with logging interceptor
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GrpcLoggingInterceptor(log)),
	)

	// Initialize auth server with logger
	authServer := server.NewAuthServer(cfg, log)
	authpb.RegisterAuthServiceServer(grpcServer, authServer)

	// Start gRPC server in a goroutine
	go func() {
		log.Info("Starting gRPC server", zap.Int("port", cfg.Auth.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve gRPC", zap.Error(err))
		}
	}()

	// Initialize REST gateway
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	if err := authpb.RegisterAuthServiceHandlerFromEndpoint(
		ctx,
		mux,
		fmt.Sprintf("localhost:%d", cfg.Auth.GRPCPort),
		opts,
	); err != nil {
		log.Fatal("Failed to register gateway", zap.Error(err))
	}

	// Add logging middleware
	httpHandler := middleware.LoggingMiddleware(log)(mux)

	// Start HTTP server
	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Auth.ServicePort),
		Handler: httpHandler,
	}

	// Start HTTP server in a goroutine
	go func() {
		log.Info("Starting HTTP server", zap.Int("port", cfg.Auth.ServicePort))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to serve HTTP", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shut down the servers
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	s := <-quit
	log.Info("Shutting down servers", zap.String("signal", s.String()))

	// Gracefully stop the gRPC server
	grpcServer.GracefulStop()
	log.Info("gRPC server stopped")

	// Gracefully shut down the HTTP server
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctxShutdown); err != nil {
		log.Fatal("Server shutdown failed", zap.Error(err))
	}

	log.Info("Auth service exited properly")
}
