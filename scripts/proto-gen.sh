#!/bin/bash

# Exit on error
set -e

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Please install Protocol Buffers from https://github.com/protocolbuffers/protobuf/releases"
    exit 1
fi

# Go to the project root directory
cd "$(dirname "$0")/.."

# Define directories
PROTO_DIR="api/proto"
GEN_DIR="api/gen"
OPENAPI_DIR="api/openapi"
THIRD_PARTY_DIR="api/third_party"

# Create directories if they don't exist
mkdir -p "${GEN_DIR}"
mkdir -p "${OPENAPI_DIR}"
mkdir -p "${THIRD_PARTY_DIR}/google/api"
mkdir -p "${THIRD_PARTY_DIR}/protoc-gen-openapiv2/options"

# Download required proto files if they don't exist
if [ ! -f "${THIRD_PARTY_DIR}/google/api/annotations.proto" ]; then
    echo "Downloading required proto files..."
    
    # Google API protos
    curl -L -o "${THIRD_PARTY_DIR}/google/api/annotations.proto" \
        "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/annotations.proto"
    curl -L -o "${THIRD_PARTY_DIR}/google/api/http.proto" \
        "https://raw.githubusercontent.com/googleapis/googleapis/master/google/api/http.proto"
    
    # gRPC-Gateway OpenAPI options
    curl -L -o "${THIRD_PARTY_DIR}/protoc-gen-openapiv2/options/annotations.proto" \
        "https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/annotations.proto"
    curl -L -o "${THIRD_PARTY_DIR}/protoc-gen-openapiv2/options/openapiv2.proto" \
        "https://raw.githubusercontent.com/grpc-ecosystem/grpc-gateway/main/protoc-gen-openapiv2/options/openapiv2.proto"
fi

# Install Go plugins if not already installed
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

# Function to generate proto files
generate_proto() {
    local service=$1
    local proto_file="${PROTO_DIR}/${service}/${service}.proto"
    local output_dir="${GEN_DIR}/${service}"
    
    # Create output directory
    mkdir -p "${output_dir}"
    
    echo "Generating protobuf code for ${service}..."
    
    # Generate Go code
    protoc -I="${PROTO_DIR}" \
        -I="${THIRD_PARTY_DIR}" \
        --go_out="${GEN_DIR}" \
        --go_opt=paths=source_relative \
        --go-grpc_out="${GEN_DIR}" \
        --go-grpc_opt=paths=source_relative \
        "${proto_file}"
    
    # Generate gRPC Gateway code
    protoc -I="${PROTO_DIR}" \
        -I="${THIRD_PARTY_DIR}" \
        --grpc-gateway_out="${GEN_DIR}" \
        --grpc-gateway_opt=logtostderr=true \
        --grpc-gateway_opt=paths=source_relative \
        "${proto_file}"
    
    # Generate OpenAPI/Swagger definitions
    protoc -I="${PROTO_DIR}" \
        -I="${THIRD_PARTY_DIR}" \
        --openapiv2_out="${OPENAPI_DIR}" \
        --openapiv2_opt=logtostderr=true \
        "${proto_file}"
}

# Generate proto files for each service
generate_proto "auth"
generate_proto "user"

echo "Protocol buffer generation completed successfully!"