package config

import (
	"fmt"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Environment      string
	Auth             AuthConfig
	User             UserConfig
	Database         DatabaseConfig
	Logging          LoggingConfig
	ServiceDiscovery ServiceDiscoveryConfig
}

// AuthConfig holds configuration specific to the Auth service
type AuthConfig struct {
	ServicePort   int
	GRPCPort      int
	JWTSecret     string
	JWTExpiration time.Duration
}

// UserConfig holds configuration specific to the User service
type UserConfig struct {
	ServicePort int
	GRPCPort    int
}

// DatabaseConfig holds configuration for the database connection
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	Params   string
}

// LoggingConfig holds configuration for logging
type LoggingConfig struct {
	Level string
}

// ServiceDiscoveryConfig holds configuration for service discovery
type ServiceDiscoveryConfig struct {
	URL string
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	if c.Driver == "mysql" {
		// MySQL DSN format: username:password@tcp(host:port)/dbname?params
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			c.User, c.Password, c.Host, c.Port, c.DBName, c.Params)
	} else if c.Driver == "postgres" {
		// PostgreSQL DSN format
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			c.Host, c.Port, c.User, c.Password, c.DBName)
	}

	// Default to MySQL format
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.Params)
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsStaging returns true if the environment is staging
func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
