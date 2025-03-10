package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// LoadConfig loads configuration from .env file and environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		// It's okay if .env doesn't exist in production
		fmt.Printf("Warning: .env file not found: %v\n", err)
	}

	// Get environment
	environment := getEnv("ENVIRONMENT", "development")

	// Determine log level based on environment
	var logLevel string
	switch environment {
	case "development":
		logLevel = getEnv("LOG_LEVEL", "debug")
	case "staging":
		logLevel = getEnv("LOG_LEVEL", "info")
	case "production":
		logLevel = getEnv("LOG_LEVEL", "warn")
	default:
		logLevel = getEnv("LOG_LEVEL", "info")
	}

	config := &Config{
		Environment: environment,
		Auth: AuthConfig{
			ServicePort:   getEnvAsInt("AUTH_SERVICE_PORT", 8081),
			GRPCPort:      getEnvAsInt("AUTH_SERVICE_GRPC_PORT", 9091),
			JWTSecret:     getEnv("JWT_SECRET", "default-secret-key"),
			JWTExpiration: getEnvAsDuration("JWT_EXPIRATION", 24*time.Hour),
		},
		User: UserConfig{
			ServicePort: getEnvAsInt("USER_SERVICE_PORT", 8082),
			GRPCPort:    getEnvAsInt("USER_SERVICE_GRPC_PORT", 9092),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "microservices"),
			SSLMode:  getEnv("DB_SSL_MODE", "disable"),
		},
		Logging: LoggingConfig{
			Level: logLevel,
		},
		ServiceDiscovery: ServiceDiscoveryConfig{
			URL: getEnv("SERVICE_DISCOVERY_URL", "localhost:8500"),
		},
	}

	return config, nil
}

// Helper functions to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
