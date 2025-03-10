.PHONY: all proto clean build run docker-build docker-run test seed

# Default target
all: proto build

# Generate proto files
proto:
	@echo "Generating proto files..."
	@./scripts/proto-gen.sh

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/*

# Build both services
build: build-auth build-user

# Build auth service
build-auth:
	@echo "Building auth service..."
	@go build -o bin/auth cmd/auth/main.go

# Build user service
build-user:
	@echo "Building user service..."
	@go build -o bin/user cmd/user/main.go

# Run both services
run: run-auth run-user

# Run auth service
run-auth:
	@echo "Running auth service..."
	@go run cmd/auth/main.go

# Run user service
run-user:
	@echo "Running user service..."
	@go run cmd/user/main.go

# Build Docker images
docker-build:
	@echo "Building Docker images..."
	@docker-compose build

# Run Docker containers
docker-run:
	@echo "Running Docker containers..."
	@docker-compose up

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run database seeders
seed: seed-users

# Seed users
seed-users:
	@echo "Seeding users..."
	@go run scripts/seed/users.go
