# Go Microservices with gRPC-Gateway

This project demonstrates a microservices architecture using Go, gRPC, and gRPC-Gateway. The project consists of two microservices:

- **Auth Service**: Handles user authentication and authorization
- **User Service**: Manages user data and operations

## Architecture

The architecture follows these key principles:

- **Microservices**: Each service is independent and can be deployed separately
- **gRPC**: Services communicate internally using gRPC for efficient, type-safe communication
- **REST API**: External clients access services via REST endpoints (provided by gRPC-Gateway)
- **Clean Architecture**: Each service follows a layered architecture (server → service → repository)
- **Mock Services**: Support for mock implementations for development and testing
- **Environment-Based Configuration**: Different settings for development, staging, and production
- **Multiple Database Support**: Works with both MySQL and PostgreSQL

## Project Structure

```
project-root/
├── .env                        # Environment variables
├── .gitignore                  # Git ignore file
├── Makefile                    # Build automation
├── docker-compose.yml          # Docker compose for local development
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
├── README.md                   # Project documentation
├── Dockerfile                  # Docker image definition
│
├── cmd/                        # Entry points for each service
│   ├── auth/                   # Auth service entry point
│   │   └── main.go
│   └── user/                   # User service entry point
│       └── main.go
│
├── pkg/                        # Shared packages
│   ├── config/                 # Configuration package
│   │   ├── config.go           # Config struct definitions
│   │   └── parser.go           # .env parsing logic
│   ├── logger/                 # Logging package
│   │   └── logger.go
│   └── middleware/             # Shared middleware
│       ├── auth.go             # Authentication middleware
│       └── logging.go          # Request logging middleware
│
├── api/                        # API definitions
│   ├── proto/                  # Protocol buffer definitions
│   │   ├── auth/               # Auth service proto files
│   │   │   ├── auth.proto
│   │   │   └── auth.swagger.json
│   │   ├── user/               # User service proto files
│   │   │   ├── user.proto
│   │   │   └── user.swagger.json
│   │   └── common/             # Shared proto definitions
│   │       └── common.proto
│   ├── gen/                    # Generated Go code from protos
│   └── openapi/                # Generated OpenAPI specs
│
├── internal/                   # Private application code
│   ├── auth/                   # Auth service implementation
│   │   ├── server/             # gRPC server implementation
│   │   │   └── server.go
│   │   ├── service/            # Business logic
│   │   │   ├── service.go
│   │   │   └── mock_service.go # Mock implementation
│   │   ├── repository/         # Data access layer
│   │   │   └── repository.go
│   │   └── client/             # Client for other services to use
│   │       ├── client.go
│   │       └── mock_client.go  # Mock implementation
│   │
│   └── user/                   # User service implementation
│       ├── server/             # gRPC server implementation
│       │   └── server.go
│       ├── service/            # Business logic
│       │   ├── service.go
│       │   └── mock_service.go # Mock implementation
│       ├── repository/         # Data access layer
│       │   └── repository.go
│       └── client/             # Client for other services to use
│           └── client.go
│
├── scripts/                    # Helper scripts
│   ├── proto-gen.sh            # Script to generate proto files
│   ├── seed/                   # Database seeders
│   │   └── users.go            # User seeder
│   └── migrations/             # Database migration scripts
│
└── k8s/                        # Kubernetes deployment configuration
    ├── base/                   # Base Kubernetes resources
    │   ├── auth-service.yaml
    │   ├── user-service.yaml
    │   ├── mysql.yaml          # MySQL database deployment
    │   ├── namespace.yaml
    │   └── kustomization.yaml
    └── overlays/               # Environment-specific overlays
        ├── development/
        │   └── kustomization.yaml
        ├── staging/
        │   └── kustomization.yaml
        └── production/
            └── kustomization.yaml
```

## Prerequisites

- Go 1.21+
- Protocol Buffers Compiler (`protoc`)
- Docker and Docker Compose (for local development)
- MySQL or PostgreSQL
- Kubernetes and kubectl (for deployment)

## Getting Started

1. **Clone the repository**

```bash
git clone https://github.com/your-username/your-project.git
cd your-project
```

2. **Install dependencies**

```bash
go mod download
```

3. **Generate Protocol Buffer code**

```bash
# Make script executable
chmod +x scripts/proto-gen.sh

# Generate proto files
make proto
```

4. **Start the services**

Using Docker:

```bash
make docker-run
```

Or locally:

```bash
# Start the database
docker-compose up -d mysql

# Start services
make run
```

## Environment Variables

Configure the application using environment variables or a `.env` file:

```
# Server settings
AUTH_SERVICE_PORT=8081
USER_SERVICE_PORT=8082
AUTH_SERVICE_GRPC_PORT=9091
USER_SERVICE_GRPC_PORT=9092

# Database settings
DB_DRIVER=mysql              # mysql or postgres
DB_HOST=localhost
DB_PORT=3306                 # 3306 for MySQL, 5432 for PostgreSQL
DB_USER=root
DB_PASSWORD=rootpassword
DB_NAME=microservices
DB_PARAMS=charset=utf8mb4&parseTime=True&loc=Local  # MySQL-specific params

# JWT settings
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Logging configuration
ENVIRONMENT=development      # development, staging, or production
LOG_LEVEL=debug             # Overrides environment-based log level

# Service discovery
SERVICE_DISCOVERY_URL=localhost:8500

# Mock services (for development and testing)
USE_MOCK_SERVICES=true      # Set to 'true' to use mock implementations
BYPASS_AUTH=false           # Set to 'true' to bypass authentication in mock mode
```

## Database Support

The application supports both MySQL and PostgreSQL databases. You can switch between them by setting the `DB_DRIVER` environment variable.

### MySQL Configuration

```
DB_DRIVER=mysql
DB_PORT=3306
DB_PARAMS=charset=utf8mb4&parseTime=True&loc=Local
```

### PostgreSQL Configuration

```
DB_DRIVER=postgres
DB_PORT=5432
DB_PARAMS=sslmode=disable
```

## Environment-Based Configuration

The application supports three environments, each with different default settings:

1. **Development**
   - Log Level: DEBUG
   - Mock Services: Available for easier testing
   - Detailed logging and debugging information

2. **Staging**
   - Log Level: INFO
   - Standard operational logging
   - Production-like configuration for testing

3. **Production**
   - Log Level: WARN or ERROR
   - Minimal logging for performance
   - No mock services or debugging features

Set the `ENVIRONMENT` variable to control environment-specific defaults.

## Development Options

### Using Mock Services

For development and testing without a database, you can use the built-in mock implementations:

1. Set `USE_MOCK_SERVICES=true` in your `.env` file
2. Optionally set `BYPASS_AUTH=true` to skip authentication checks during development
3. Start the services as usual

Mock services provide:
- In-memory data storage (no database required)
- Pre-configured test users
- Full API functionality with the same validation logic
- Simulated inter-service communication

Pre-configured mock users:
- Admin: `admin@example.com` / `admin123`
- User: `user@example.com` / `password123`
- Test: `test@example.com` / `test123`

### Seeding Data

To populate your database with initial data:

```bash
make seed-users
```

This creates test users that you can use for development.

## API Endpoints

### Auth Service

- **POST /api/v1/auth/register** - Register a new user
  ```json
  {
    "email": "user@example.com",
    "password": "password123",
    "name": "Example User"
  }
  ```

- **POST /api/v1/auth/login** - Authenticate a user and get JWT token
  ```json
  {
    "email": "user@example.com",
    "password": "password123"
  }
  ```

- **POST /api/v1/auth/validate** - Validate a JWT token
  ```json
  {
    "token": "your.jwt.token"
  }
  ```

### User Service

- **GET /api/v1/users/{id}** - Get a user by ID
- **PUT /api/v1/users/{id}** - Update a user
  ```json
  {
    "name": "New Name",
    "email": "new.email@example.com"
  }
  ```
- **DELETE /api/v1/users/{id}** - Delete a user
- **GET /api/v1/users?page=1&page_size=10** - List users (with pagination)

## Inter-Service Communication

Services communicate with each other using gRPC. The User Service calls the Auth Service to validate JWT tokens.

## Features

- **Authentication**: JWT-based authentication
- **User Management**: Create, read, update, delete users
- **Structured Logging**: Comprehensive logging with different levels
- **Configuration**: Environment-based configuration
- **Mock Services**: In-memory implementations for development
- **Database Flexibility**: Support for MySQL and PostgreSQL
- **Validation**: Input validation and error handling
- **Dockerization**: Containerized for easy deployment
- **Kubernetes Deployment**: Environment-specific K8s configurations

## Development

### Building the Services

```bash
# Build all services
make build

# Build specific service
make build-auth
make build-user
```

### Running Tests

```bash
make test
```

### Cleaning Up

```bash
make clean
```

## Docker Deployment

The project includes Docker and Docker Compose files for containerized deployment:

```bash
# Build Docker images
make docker-build

# Run containers
make docker-run
```

## Kubernetes Deployment

The project includes Kubernetes configurations for deploying to different environments:

```bash
# Deploy to development environment
kubectl apply -k k8s/overlays/development

# Deploy to staging environment
kubectl apply -k k8s/overlays/staging

# Deploy to production environment
kubectl apply -k k8s/overlays/production
```

The Kubernetes configurations include:
- Deployments for Auth and User services
- MySQL StatefulSet with persistent storage
- Ingress for external access
- ConfigMaps for environment-specific settings
- Secrets for sensitive information
- Resource limits and health checks

See the [Kubernetes Deployment Guide](k8s/README.md) for detailed instructions.

## Logging

The services use structured logging with environment-specific log levels:

```
# Development
LOG_LEVEL=debug  # Verbose, detailed logs

# Staging
LOG_LEVEL=info   # Normal operational logs

# Production
LOG_LEVEL=warn   # Only warnings and errors
```

Log output includes:
- HTTP/gRPC request details
- Database operations
- Authentication events
- Service startup/shutdown information

## License

This project is licensed under the GNU General Public License v2.0 - see the LICENSE file for details.
