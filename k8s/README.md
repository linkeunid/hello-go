# Kubernetes Deployment Guide

This guide explains how to deploy the microservices application to a Kubernetes cluster, with configurations for different environments and database options.

## Prerequisites

- kubectl (v1.20+)
- kustomize (v4.0.0+)
- Access to a Kubernetes cluster
- Container registry for your images

## Project Structure

All Kubernetes configuration files are in the `k8s/` directory:

```
k8s/
├── base/                   # Base Kubernetes resources
│   ├── auth-service.yaml   # Auth service deployment and service
│   ├── user-service.yaml   # User service deployment and service
│   ├── mysql.yaml          # MySQL database StatefulSet
│   ├── namespace.yaml
│   ├── environment-config.yaml
│   └── kustomization.yaml
│
└── overlays/               # Environment-specific overlays
    ├── development/        # Development environment
    │   └── kustomization.yaml
    ├── staging/            # Staging environment
    │   └── kustomization.yaml
    └── production/         # Production environment
        └── kustomization.yaml
```

## Database Options

The deployment supports both MySQL and PostgreSQL databases:

### MySQL Configuration (Default)

The default configuration uses MySQL 8.0 with the following features:
- Persistent storage using a StatefulSet
- Proper initialization and health checks
- UTF-8 character encoding (utf8mb4)
- Secure credential management through Kubernetes Secrets

### PostgreSQL Configuration (Alternative)

To use PostgreSQL instead of MySQL:
1. Replace `mysql.yaml` with `postgres.yaml` in your kustomization.yaml
2. Update the database environment variables in the service configurations

## Environment-Specific Configurations

The deployment uses Kustomize overlays to manage environment-specific configurations:

### Development Environment
- Single replica per service
- DEBUG log level
- Limited resource requests
- Mock services can be enabled for testing
- Suitable for local Kubernetes development (minikube, kind, etc.)

### Staging Environment
- Two replicas per service
- INFO log level
- Moderate resource allocation
- Mock services disabled
- Separate domain (api.staging.your-domain.com)

### Production Environment
- Three or more replicas per service
- WARN log level (minimal logging)
- Higher resource limits and requests
- Production-grade database configuration
- Production domain (api.your-domain.com)

## Deployment Process

### Step 1: Prepare Your Cluster

Create a dedicated namespace for your microservices:

```bash
kubectl apply -f k8s/base/namespace.yaml
```

### Step 2: Build and Push Docker Images

Build your service images and push them to your container registry:

```bash
# Set your registry and version
export REGISTRY=your-registry.io
export VERSION=1.0.0

# Auth Service
docker build -t $REGISTRY/auth-service:$VERSION -f Dockerfile .
docker push $REGISTRY/auth-service:$VERSION

# User Service
docker build -t $REGISTRY/user-service:$VERSION -f Dockerfile .
docker push $REGISTRY/user-service:$VERSION
```

### Step 3: Update Image References

Update the image references in your kustomization.yaml files:

```yaml
# In k8s/overlays/[environment]/kustomization.yaml
images:
- name: your-registry/auth-service
  newName: your-registry.io/auth-service
  newTag: 1.0.0
- name: your-registry/user-service
  newName: your-registry.io/user-service
  newTag: 1.0.0
```

### Step 4: Update ConfigMaps and Secrets

For production, you should securely manage secrets. Update the secrets in each environment:

```bash
# Example for creating a secure JWT secret
kubectl create secret generic auth-secrets -n microservices-prod \
  --from-literal=JWT_SECRET=$(openssl rand -base64 32) \
  --dry-run=client -o yaml > jwt-secret.yaml

# Apply the secret
kubectl apply -f jwt-secret.yaml
```

### Step 5: Deploy to Your Environment

Use kustomize to deploy to your desired environment:

```bash
# Development
kubectl apply -k k8s/overlays/development

# Staging
kubectl apply -k k8s/overlays/staging

# Production
kubectl apply -k k8s/overlays/production
```

### Step 6: Verify Deployment

Check that all resources have been created and are running:

```bash
# List all resources in your namespace
kubectl get all -n microservices-prod

# Check pods status
kubectl get pods -n microservices-prod

# Check service endpoints
kubectl get services -n microservices-prod

# Check deployments
kubectl get deployments -n microservices-prod
```

### Step 7: Initialize the Database (First Deployment Only)

For the first deployment, you may need to initialize the database with schema and seed data:

```bash
# Deploy database initialization job
kubectl apply -f k8s/init/db-init-job.yaml

# Monitor job progress
kubectl logs -f job/db-init-job -n microservices-prod
```

## Scaling

You can scale your services based on load:

```bash
# Scale auth service to 5 replicas
kubectl scale deployment auth-service -n microservices-prod --replicas=5

# Scale user service to 5 replicas
kubectl scale deployment user-service -n microservices-prod --replicas=5
```

## Monitoring and Logging

### Monitoring with Prometheus and Grafana

The services are configured with Prometheus endpoints for monitoring:

1. Deploy Prometheus and Grafana using Helm:
   ```bash
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm install prometheus prometheus-community/kube-prometheus-stack -n monitoring
   ```

2. Configure Prometheus to scrape your services by applying service monitors:
   ```bash
   kubectl apply -f k8s/monitoring/service-monitors.yaml
   ```

3. Import the provided Grafana dashboards:
   ```bash
   kubectl apply -f k8s/monitoring/grafana-dashboards.yaml
   ```

### Centralized Logging

For centralized logging, you can use the ELK stack or Loki:

1. Deploy Elasticsearch, Logstash, and Kibana:
   ```bash
   kubectl apply -f k8s/logging/elk-stack.yaml
   ```

2. Or deploy Loki and Grafana:
   ```bash
   helm repo add grafana https://grafana.github.io/helm-charts
   helm install loki grafana/loki-stack -n logging
   ```

## Database Migration and Backup

### Database Migrations

For database schema migrations, use a Kubernetes Job:

```bash
# Create a migration job
kubectl apply -f k8s/migration/migration-job.yaml

# Check migration logs
kubectl logs -f job/migration-job -n microservices-prod
```

### Database Backups

Set up regular backups using a CronJob:

```bash
# Create a backup CronJob
kubectl apply -f k8s/backup/db-backup-cronjob.yaml

# Verify the schedule
kubectl get cronjob -n microservices-prod
```

## Troubleshooting

### Common Issues

1. **Pods stuck in Pending state**: Check for resource constraints
   ```bash
   kubectl describe pod <pod-name> -n microservices-prod
   ```

2. **Database connection issues**: Verify database secret and configuration
   ```bash
   kubectl logs deployment/auth-service -n microservices-prod
   kubectl logs deployment/user-service -n microservices-prod
   ```

3. **Services not accessible**: Check ingress configuration
   ```bash
   kubectl get ingress -n microservices-prod
   kubectl describe ingress -n microservices-prod
   ```

4. **Init container failures**: Check for database connectivity
   ```bash
   kubectl describe pod <pod-name> -n microservices-prod
   kubectl logs <pod-name> -c wait-for-mysql -n microservices-prod
   ```

### Debugging Commands

```bash
# Get detailed pod information
kubectl describe pod <pod-name> -n microservices-prod

# Check pod logs
kubectl logs <pod-name> -n microservices-prod

# Check specific container logs
kubectl logs <pod-name> -c <container-name> -n microservices-prod

# Execute commands in a pod
kubectl exec -it <pod-name> -n microservices-prod -- /bin/sh

# Port-forward to a service for direct access
kubectl port-forward service/auth-service 8081:8081 -n microservices-prod
```

## Clean Up

To remove the deployment:

```bash
# Remove development environment
kubectl delete -k k8s/overlays/development

# Remove staging environment
kubectl delete -k k8s/overlays/staging

# Remove production environment
kubectl delete -k k8s/overlays/production
```

## Environment Variables Reference

Here are all the environment variables that can be configured:

### Service Configuration
- `AUTH_SERVICE_PORT`: HTTP port for Auth service
- `AUTH_SERVICE_GRPC_PORT`: gRPC port for Auth service
- `USER_SERVICE_PORT`: HTTP port for User service
- `USER_SERVICE_GRPC_PORT`: gRPC port for User service

### Database Configuration
- `DB_DRIVER`: Database driver (mysql or postgres)
- `DB_HOST`: Database hostname
- `DB_PORT`: Database port
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `DB_NAME`: Database name
- `DB_PARAMS`: Database connection parameters

### Security
- `JWT_SECRET`: Secret key for JWT tokens
- `JWT_EXPIRATION`: JWT token expiration time

### Logging and Environment
- `ENVIRONMENT`: Application environment (development, staging, production)
- `LOG_LEVEL`: Log level (debug, info, warn, error)

### Feature Flags
- `USE_MOCK_SERVICES`: Enable/disable mock implementations
- `BYPASS_AUTH`: Bypass authentication (development only)
