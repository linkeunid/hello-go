# Build stage
FROM golang:1.23.2-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build auth service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/auth ./cmd/auth/main.go

# Build user service
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/user ./cmd/user/main.go

# Final stage
FROM alpine:latest

# Install certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy binaries from builder stage
COPY --from=builder /app/bin/ .

# Copy .env file
COPY .env .

# Expose ports
EXPOSE 8081 8082 9091 9092

# Command is overridden in docker-compose.yml
CMD ["./auth"]
