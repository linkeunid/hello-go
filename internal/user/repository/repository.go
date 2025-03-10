package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/linkeunid/hello-go/pkg/config"
)

// Common errors
var (
	ErrUserNotFound = errors.New("user not found")
)

// User represents a user in the database
type User struct {
	ID        string `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex"`
	Password  string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// UserRepository defines the interface for user repository operations
type UserRepository interface {
	// GetUserByID gets a user by ID
	GetUserByID(ctx context.Context, id string) (*User, error)
	// UpdateUser updates a user's information
	UpdateUser(ctx context.Context, id, name, email string) (*User, error)
	// DeleteUser deletes a user by ID
	DeleteUser(ctx context.Context, id string) error
	// ListUsers returns a list of users
	ListUsers(ctx context.Context, page, pageSize int) ([]*User, int, error)
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db     *gorm.DB
	logger *zap.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(cfg *config.Config, logger *zap.Logger) UserRepository {
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

	// Connect to database
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: zapAdapter,
	})
	if err != nil {
		// Log and panic
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Migrate the schema
	if err := db.AutoMigrate(&User{}); err != nil {
		logger.Fatal("Failed to migrate database schema", zap.Error(err))
	}

	return &userRepository{
		db:     db,
		logger: logger,
	}
}

// GetUserByID gets a user by ID
func (r *userRepository) GetUserByID(ctx context.Context, id string) (*User, error) {
	var user User

	r.logger.Debug("Getting user by ID", zap.String("user_id", id))

	result := r.db.WithContext(ctx).Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			r.logger.Debug("User not found", zap.String("user_id", id))
			return nil, ErrUserNotFound
		}
		r.logger.Error("Database error while getting user",
			zap.String("user_id", id),
			zap.Error(result.Error))
		return nil, result.Error
	}

	r.logger.Debug("User found",
		zap.String("user_id", id),
		zap.String("email", user.Email))

	return &user, nil
}

// UpdateUser updates a user's information
func (r *userRepository) UpdateUser(ctx context.Context, id, name, email string) (*User, error) {
	r.logger.Debug("Updating user",
		zap.String("user_id", id),
		zap.String("name", name),
		zap.String("email", email))

	// Get user
	user, err := r.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	user.Name = name
	user.Email = email
	user.UpdatedAt = time.Now()

	// Save to database
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		r.logger.Error("Database error while updating user",
			zap.String("user_id", id),
			zap.Error(result.Error))
		return nil, result.Error
	}

	r.logger.Debug("User updated successfully",
		zap.String("user_id", id),
		zap.String("email", email))

	return user, nil
}

// DeleteUser deletes a user by ID
func (r *userRepository) DeleteUser(ctx context.Context, id string) error {
	r.logger.Debug("Deleting user", zap.String("user_id", id))

	// Check if user exists
	if _, err := r.GetUserByID(ctx, id); err != nil {
		return err
	}

	result := r.db.WithContext(ctx).Delete(&User{}, "id = ?", id)
	if result.Error != nil {
		r.logger.Error("Database error while deleting user",
			zap.String("user_id", id),
			zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected == 0 {
		r.logger.Warn("No rows affected when deleting user",
			zap.String("user_id", id))
		return fmt.Errorf("no rows affected: %w", ErrUserNotFound)
	}

	r.logger.Debug("User deleted successfully", zap.String("user_id", id))
	return nil
}

// ListUsers returns a list of users
func (r *userRepository) ListUsers(ctx context.Context, page, pageSize int) ([]*User, int, error) {
	var users []*User
	var total int64

	r.logger.Debug("Listing users",
		zap.Int("page", page),
		zap.Int("page_size", pageSize))

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	result := r.db.WithContext(ctx).Model(&User{}).Count(&total)
	if result.Error != nil {
		r.logger.Error("Database error counting users", zap.Error(result.Error))
		return nil, 0, result.Error
	}

	// Get users
	result = r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&users)
	if result.Error != nil {
		r.logger.Error("Database error listing users", zap.Error(result.Error))
		return nil, 0, result.Error
	}

	r.logger.Debug("Listed users successfully",
		zap.Int("count", len(users)),
		zap.Int64("total", total))

	return users, int(total), nil
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
