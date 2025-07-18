# Deployment Guide

This guide covers various deployment strategies for the Maintenance API, from local development to production environments.

## 🚀 Quick Start (Local Development)

### Prerequisites
- Docker & Docker Compose
- Go 1.22+
- Make (optional)

### Local Development with Hot Reload
```bash
# Start the application with hot reload
docker-compose up -d

# View logs
docker-compose logs -f app

# Stop the application
docker-compose down
```

### Local Development with Air (Hot Reload)
```bash
# Install Air (hot reload tool)
go install github.com/cosmtrek/air@latest

# Start with hot reload
air -c .air.toml
```

## 🐳 Docker Deployment

### Build and Run
```bash
# Build the image
docker build -t maintenance-api:latest .

# Run the container
docker run -p 8080:8080 \
  -e DB_HOST=localhost \
  -e DB_USER=user \
  -e DB_PASSWORD=password \
  -e DB_NAME=tasks_db \
  maintenance-api:latest
```

### Multi-stage Build
```bash
# Production build
docker build --target prod -t maintenance-api:prod .

# Development build
docker build --target dev -t maintenance-api:dev .
```

## ☸️ Kubernetes Deployment

### Local Kubernetes (kind/minikube)

#### 1. Create Local Cluster
```bash
# Using kind
kind create cluster --name maintenance-api

# Using minikube
minikube start --memory=4096 --cpus=2
```

#### 2. Deploy Application
```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Deploy MySQL
kubectl apply -f k8s/mysql/

# Deploy application
kubectl apply -f k8s/app/

# Check status
kubectl get pods -n maintenance-api
kubectl get services -n maintenance-api
```

#### 3. Access Application
```bash
# Port forward to access locally
kubectl port-forward svc/maintenance-api 8080:8080 -n maintenance-api

# Access the API
curl http://localhost:8080/health/ready
```

### Production Kubernetes (EKS/GKE/AKS)

#### Using Helm
```bash
# Add repository (if using custom registry)
helm repo add maintenance-api ./helm/maintenance-api

# Install with custom values
helm install maintenance-api ./helm/maintenance-api \
  --namespace maintenance-api \
  --create-namespace \
  --values helm/maintenance-api/values-production.yaml

# Upgrade deployment
helm upgrade maintenance-api ./helm/maintenance-api \
  --namespace maintenance-api \
  --values helm/maintenance-api/values-production.yaml
```

## 📊 Monitoring Setup

### Prometheus & Grafana
```bash
# Install monitoring stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values monitoring/prometheus-values.yaml

# Access Grafana
kubectl port-forward svc/prometheus-grafana 3000:80 -n monitoring
# Default credentials: admin/prom-operator
```

### Application Metrics
The application exposes metrics at `/metrics` endpoint:
- HTTP request duration
- Request count
- Error rates
- Database connection status

## 🔒 Security Configuration

### Environment Variables
```bash
# Required environment variables
DB_HOST=mysql
DB_USER=user
DB_PASSWORD=secure_password
DB_NAME=tasks_db
JWT_SECRET=your_jwt_secret_key
LOG_LEVEL=info
```

### Secrets Management
```bash
# Create Kubernetes secrets
kubectl create secret generic maintenance-api-secrets \
  --from-literal=db-password=secure_password \
  --from-literal=jwt-secret=your_jwt_secret_key \
  --namespace maintenance-api
```

## 🔄 CI/CD Pipeline

### GitHub Actions
The repository includes a complete CI/CD pipeline:
- **Build**: Multi-stage Docker build
- **Test**: Unit and integration tests
- **Security**: Container scanning with Trivy
- **Deploy**: Automated deployment to staging/production

### Manual Deployment
```bash
# Build and push image
docker build -t ghcr.io/yourusername/maintenance-api:latest .
docker push ghcr.io/yourusername/maintenance-api:latest

# Deploy to Kubernetes
kubectl set image deployment/maintenance-api app=ghcr.io/yourusername/maintenance-api:latest -n maintenance-api
```

## 🌐 Cloud Deployment

### AWS EKS
```bash
# Configure AWS CLI
aws configure

# Create EKS cluster (see terraform/aws/)
cd terraform/aws/
terraform init
terraform apply

# Deploy application
kubectl apply -f k8s/
```

### Google GKE
```bash
# Configure gcloud
gcloud auth login
gcloud config set project your-project-id

# Create cluster
gcloud container clusters create maintenance-api \
  --num-nodes=3 \
  --machine-type=e2-medium \
  --region=us-central1

# Deploy application
kubectl apply -f k8s/
```

### Azure AKS
```bash
# Configure Azure CLI
az login
az account set --subscription your-subscription-id

# Create cluster
az aks create \
  --resource-group maintenance-api-rg \
  --name maintenance-api-cluster \
  --node-count 3 \
  --generate-ssh-keys

# Get credentials
az aks get-credentials --resource-group maintenance-api-rg --name maintenance-api-cluster

# Deploy application
kubectl apply -f k8s/
```

## 📈 Scaling Configuration

### Horizontal Pod Autoscaler
```yaml
# HPA configuration
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: maintenance-api-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: maintenance-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Vertical Pod Autoscaler
```bash
# Install VPA
kubectl apply -f https://github.com/kubernetes/autoscaler/releases/latest/download/vertical-pod-autoscaler.yaml

# Create VPA for application
kubectl apply -f k8s/vpa/maintenance-api-vpa.yaml
```

## 🧪 Testing Deployment

### Smoke Tests
```bash
# Run smoke tests
k6 run load-testing/k6-smoke-test.js

# Run load tests
k6 run load-testing/k6-load-test.js
```

### Health Checks
```bash
# Check application health
kubectl get pods -n maintenance-api
kubectl describe pod <pod-name> -n maintenance-api

# Check services
kubectl get svc -n maintenance-api
kubectl describe svc maintenance-api -n maintenance-api
```

## 🗄️ Database Migration

### Initial Setup
```bash
# Run database migrations
kubectl create configmap db-init-script --from-file=db/init.sql -n maintenance-api

# Apply database configuration
kubectl apply -f k8s/mysql/
```

### Backup Strategy
```bash
# Create database backup
kubectl exec -it <mysql-pod> -- mysqldump -u user -p tasks_db > backup.sql

# Restore from backup
kubectl exec -i <mysql-pod> -- mysql -u user -p tasks_db < backup.sql
```

## 🚨 Troubleshooting

### Common Issues

1. **Pod CrashLoopBackOff**
   ```bash
   kubectl logs <pod-name> -n maintenance-api
   kubectl describe pod <pod-name> -n maintenance-api
   ```

2. **Service Not Accessible**
   ```bash
   kubectl get endpoints -n maintenance-api
   kubectl port-forward svc/maintenance-api 8080:8080 -n maintenance-api
   ```

3. **Database Connection Issues**
   ```bash
   kubectl get pods -n maintenance-api
   kubectl logs <mysql-pod> -n maintenance-api
   ```

### Debug Commands
```bash
# Get cluster info
kubectl cluster-info

# Check node status
kubectl get nodes

# Check all resources
kubectl get all -n maintenance-api

# View logs
kubectl logs -f deployment/maintenance-api -n maintenance-api