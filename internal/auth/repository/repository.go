package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/linkeunid/hello-go/pkg/config"
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

// AuthRepository defines the interface for auth repository operations
type AuthRepository interface {
	// GetUserByEmail gets a user by email
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	// UserExists checks if a user exists by email
	UserExists(ctx context.Context, email string) (bool, error)
	// CreateUser creates a new user
	CreateUser(ctx context.Context, email, password, name string) (string, error)
	// CheckPassword verifies a user's password
	CheckPassword(storedPassword, providedPassword string) error
}

// authRepository implements the AuthRepository interface
type authRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewAuthRepository creates a new auth repository
func NewAuthRepository(cfg *config.Config, logger *zap.Logger) AuthRepository {
	// Create custom GORM logger that uses zap
	gormLogger := logger.Named("gorm")

	zapAdapter := zapGormLogger{
		Logger: gormLogger,
		Config: gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  gormlogger.Info,
			IgnoreRecordNotFoundError: false,
			Colorful:                  false,
		},
	}

	var db *gorm.DB
	var err error

	if cfg.Database.Driver == "mysql" {
		// Connect to MySQL database
		db, err = gorm.Open(mysql.Open(cfg.Database.GetDSN()), &gorm.Config{
			Logger: zapAdapter,
		})
	} else {
		logger.Fatal("Unsupported database driver", zap.String("driver", cfg.Database.Driver))
	}

	if err != nil {
		// Log and panic
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		logger.Fatal("Failed to migrate database schema", zap.Error(err))
	}

	return &authRepository{
		db:     db,
		logger: logger,
	}
}

// GetUserByEmail gets a user by email
func (r *authRepository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	var user User

	r.logger.Debug("Getting user by email", zap.String("email", email))

	result := r.db.WithContext(ctx).Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Debug("User not found", zap.String("email", email))
		} else {
			r.logger.Error("Database error while getting user",
				zap.String("email", email),
				zap.Error(result.Error))
		}
		return nil, result.Error
	}

	r.logger.Debug("User found",
		zap.String("email", email),
		zap.String("user_id", user.ID))

	return &user, nil
}

// UserExists checks if a user exists by email
func (r *authRepository) UserExists(ctx context.Context, email string) (bool, error) {
	var count int64

	r.logger.Debug("Checking if user exists", zap.String("email", email))

	result := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email).Count(&count)
	if result.Error != nil {
		r.logger.Error("Database error while checking if user exists",
			zap.String("email", email),
			zap.Error(result.Error))
		return false, result.Error
	}

	exists := count > 0
	r.logger.Debug("User existence check result",
		zap.String("email", email),
		zap.Bool("exists", exists))

	return exists, nil
}

// CreateUser creates a new user
func (r *authRepository) CreateUser(ctx context.Context, email, password, name string) (string, error) {
	// Generate a new UUID for the user ID
	userID := uuid.New().String()

	r.logger.Debug("Creating new user",
		zap.String("email", email),
		zap.String("name", name),
		zap.String("user_id", userID))

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		r.logger.Error("Failed to hash password", zap.Error(err))
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := User{
		ID:        userID,
		Email:     email,
		Password:  string(hashedPassword),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	result := r.db.WithContext(ctx).Create(&user)
	if result.Error != nil {
		r.logger.Error("Database error while creating user",
			zap.String("email", email),
			zap.Error(result.Error))
		return "", result.Error
	}

	r.logger.Debug("User created successfully",
		zap.String("email", email),
		zap.String("user_id", userID))

	return userID, nil
}

// CheckPassword verifies a user's password
func (r *authRepository) CheckPassword(storedPassword, providedPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(providedPassword))
}

// Custom GORM logger that uses Zap
type zapGormLogger struct {
	Logger *zap.Logger
	gormlogger.Config
}

// LogMode sets log mode
func (l zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	newLogger := l
	newLogger.LogLevel = level
	return newLogger
}

// Info logs info
func (l zapGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Info {
		l.Logger.Sugar().Infof(msg, data...)
	}
}

// Warn logs warning
func (l zapGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Warn {
		l.Logger.Sugar().Warnf(msg, data...)
	}
}

// Error logs error
func (l zapGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormlogger.Error {
		l.Logger.Sugar().Errorf(msg, data...)
	}
}

// Trace logs SQL trace
func (l zapGormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && l.LogLevel >= gormlogger.Error && !errors.Is(err, gorm.ErrRecordNotFound):
		l.Logger.Error("SQL error", append(fields, zap.Error(err))...)
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormlogger.Warn:
		l.Logger.Warn("Slow SQL query", fields...)
	case l.LogLevel >= gormlogger.Info:
		l.Logger.Debug("SQL query", fields...)
	}
}
