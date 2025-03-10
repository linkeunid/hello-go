package config

import (
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Auth             AuthConfig
	User             UserConfig
	Database         DatabaseConfig
	Logging          LoggingConfig
	ServiceDiscovery ServiceDiscoveryConfig
	Environment      string
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
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
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
	return "host=" + c.Host +
		" port=" + strconv.Itoa(c.Port) +
		" user=" + c.User +
		" password=" + c.Password +
		" dbname=" + c.DBName +
		" sslmode=" + c.SSLMode
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsStaging() bool {
	return c.Environment == "staging"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}
