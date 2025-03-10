# Kubernetes Deployment Guide

This guide explains how to deploy the microservices application to a Kubernetes cluster, with configurations for different environments.

## Prerequisites

- kubectl (configured to connect to your cluster)
- kustomize (v3.8.0+)
- Access to a container registry for your images

## Project Structure

All Kubernetes configuration files are in the `k8s/` directory:

```
k8s/
├── auth-service.yaml    # Auth service deployment, service, and ingress
├── user-service.yaml    # User service deployment, service, and ingress
├── postgres.yaml        # PostgreSQL database deployment and service
├── namespace.yaml       # Namespace definition
├── environment-config.yaml    # Environment-specific configurations
└── kustomization.yaml   # Kustomize configuration
```

## Available Environments

The deployment supports three environments:

1. **Development** - For local or dev clusters
2. **Staging** - Pre-production environment
3. **Production** - Production environment

Each environment uses different configuration settings like log levels and feature flags.

## Deployment Instructions

### 1. Build and Push Docker Images

First, build your service images and push them to your container registry:

```bash
# Build images
docker build -t your-registry/auth-service:latest -f Dockerfile .
docker build -t your-registry/user-service:latest -f Dockerfile .

# Push to registry
docker push your-registry/auth-service:latest
docker push your-registry/user-service:latest
```

### 2. Update Image References

Before deploying, update the image references in the YAML files to match your registry:

```yaml
# In auth-service.yaml and user-service.yaml
image: your-registry/auth-service:latest
image: your-registry/user-service:latest
```

### 3. Deploy to a Specific Environment

Use kustomize to deploy to your desired environment:

#### Development Environment

```bash
kubectl apply -k k8s/overlays/development
```

#### Staging Environment

```bash
kubectl apply -k k8s/overlays/staging
```

#### Production Environment

```bash
kubectl apply -k k8s/overlays/production
```

### 4. Verify Deployment

Check that all resources have been created:

```bash
# Check namespace
kubectl get namespace microservices

# Check pods
kubectl get pods -n microservices

# Check services
kubectl get services -n microservices

# Check deployments
kubectl get deployments -n microservices
```

## Environment-Specific Configurations

Different environments have different configurations, managed through ConfigMaps:

### Development
- LOG_LEVEL: debug
- USE_MOCK_SERVICES: true (can be enabled for development)
- BYPASS_AUTH: false

### Staging
- LOG_LEVEL: info
- USE_MOCK_SERVICES: false
- BYPASS_AUTH: false

### Production
- LOG_LEVEL: warn
- USE_MOCK_SERVICES: false
- BYPASS_AUTH: false

## Managing Secrets

Sensitive information is stored in Kubernetes Secrets:

- PostgreSQL password
- JWT secret key

In a real production environment, you should:
1. Use a secrets management solution (e.g., Vault, AWS Secrets Manager)
2. Replace the base64-encoded values with proper secrets
3. Consider using a solution like Sealed Secrets for git-ops

## Resource Management

The deployment includes resource requests and limits:

- Microservices:
  - Requests: 100m CPU, 128Mi memory
  - Limits: 200m CPU, 256Mi memory
  
- PostgreSQL:
  - Requests: 500m CPU, 512Mi memory
  - Limits: 1000m CPU, 1Gi memory

Adjust these values based on your workload requirements.

## Health Checks

All services include:
- Readiness probes - to determine when a container is ready to accept traffic
- Liveness probes - to determine when a container needs to be restarted

## Ingress Configuration

The services are exposed through an Ingress resource:

- Auth Service: https://api.your-domain.com/auth/...
- User Service: https://api.your-domain.com/users/...

Update the host and paths in the Ingress resources as needed.

## Storage

PostgreSQL uses a PersistentVolumeClaim for data storage. The default configuration:

- 10Gi storage
- Standard storage class
- ReadWriteOnce access mode

## Troubleshooting

If you encounter issues:

1. Check pod logs:
   ```bash
   kubectl logs -n microservices <pod-name>
   ```

2. Check pod description:
   ```bash
   kubectl describe pod -n microservices <pod-name>
   ```

3. Check service endpoints:
   ```bash
   kubectl get endpoints -n microservices
   ```

4. Check ingress configuration:
   ```bash
   kubectl describe ingress -n microservices
   ```
