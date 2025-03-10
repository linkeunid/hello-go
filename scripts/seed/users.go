package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/linkeunid/hello-go/pkg/config"
	"github.com/linkeunid/hello-go/pkg/logger"
)

// User represents a user in the database
type User struct {
	ID        string `gorm:"primaryKey;type:varchar(36)"`
	Email     string `gorm:"uniqueIndex;type:varchar(100)"`
	Password  string `gorm:"type:varchar(255)"`
	Name      string `gorm:"type:varchar(100)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Generate hashed password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

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

	log.Info("Starting user seeder")

	// Connect to database
	log.Debug("Connecting to MySQL database",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("database", cfg.Database.DBName),
		zap.String("driver", cfg.Database.Driver))

	db, err := gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Migrate the schema
	log.Debug("Migrating database schema")
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	// Define users to seed
	users := []struct {
		Email    string
		Password string
		Name     string
	}{
		{
			Email:    "admin@example.com",
			Password: "admin123",
			Name:     "Admin User",
		},
		{
			Email:    "user1@example.com",
			Password: "user123",
			Name:     "Regular User 1",
		},
		{
			Email:    "user2@example.com",
			Password: "user456",
			Name:     "Regular User 2",
		},
		{
			Email:    "test@example.com",
			Password: "test123",
			Name:     "Test User",
		},
	}

	// Seed users
	ctx := context.Background()
	log.Info("Seeding users", zap.Int("count", len(users)))

	for _, u := range users {
		// Check if user already exists
		var count int64
		db.WithContext(ctx).Model(&User{}).Where("email = ?", u.Email).Count(&count)
		if count > 0 {
			log.Info("User already exists, skipping", zap.String("email", u.Email))
			continue
		}

		// Hash password
		log.Debug("Hashing password", zap.String("email", u.Email))
		hashedPassword, err := hashPassword(u.Password)
		if err != nil {
			log.Error("Failed to hash password",
				zap.String("email", u.Email),
				zap.Error(err))
			continue
		}

		// Create user
		userID := uuid.New().String()
		log.Debug("Creating user",
			zap.String("email", u.Email),
			zap.String("user_id", userID))

		user := User{
			ID:        userID,
			Email:     u.Email,
			Password:  hashedPassword,
			Name:      u.Name,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		result := db.WithContext(ctx).Create(&user)
		if result.Error != nil {
			log.Error("Failed to create user",
				zap.String("email", u.Email),
				zap.Error(result.Error))
			continue
		}

		log.Info("User created successfully",
			zap.String("email", u.Email),
			zap.String("name", u.Name),
			zap.String("user_id", userID))
	}

	log.Info("User seeding completed successfully!")
}
