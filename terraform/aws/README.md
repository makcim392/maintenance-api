# AWS EKS Infrastructure for Maintenance API

This directory contains Terraform configuration for deploying the Maintenance API on AWS EKS. This is provided as a **documentation-only example** for portfolio purposes.

## ⚠️ Cost Warning
Running this infrastructure will incur AWS charges. Estimated costs:
- **EKS Cluster**: ~$73/month
- **EC2 Instances**: ~$30-60/month (t3.medium, 2 instances)
- **Load Balancer**: ~$20/month
- **Total**: ~$120-150/month

## Prerequisites

1. AWS CLI configured with appropriate credentials
2. Terraform >= 1.0
3. kubectl configured
4. Helm 3.x

## Quick Start (Documentation Only)

```bash
# Initialize Terraform
terraform init

# Plan the deployment
terraform plan

# Apply the configuration
terraform apply

# Configure kubectl
aws eks update-kubeconfig --region us-west-2 --name maintenance-api-cluster

# Deploy the application
helm upgrade --install maintenance-api ../../helm/maintenance-api \
  --namespace maintenance-api \
  --create-namespace \
  --values values-production.yaml
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                        AWS Cloud                              │
├─────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │   ALB (Ingress)     │    │   EKS Cluster               │ │
│  │   Port 80/443       │    │   Kubernetes 1.28          │ │
│  └──────────┬──────────┘    └──────────┬──────────────────┘ │
│             │                          │                      │
│  ┌──────────┴──────────┐    ┌──────────┴──────────────────┐ │
│  │   Route 53          │    │   Node Group (t3.medium)    │ │
│  │   DNS               │    │   Min: 1, Max: 3            │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Security Features

- **IRSA (IAM Roles for Service Accounts)**: Fine-grained IAM permissions
- **Security Groups**: Restricted network access
- **Private Subnets**: Database and application isolation
- **Encryption**: EBS encryption at rest

## Monitoring Setup

After cluster creation, install monitoring stack:

```bash
# Install Prometheus and Grafana
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace monitoring \
  --create-namespace \
  --values ../prometheus-values.yaml
```

## Scaling Configuration

The cluster supports:
- **Horizontal Pod Autoscaling**: Based on CPU/memory metrics
- **Cluster Autoscaling**: Node group auto-scaling
- **Vertical Pod Autoscaling**: Resource optimization

## Cleanup

To avoid ongoing charges:

```bash
# Delete all resources
terraform destroy

# Clean up ECR images
aws ecr delete-repository --repository-name maintenance-api --force
```

## Alternative: Local Development

For cost-free development, use:
- **kind** (Kubernetes in Docker)
- **minikube** (local Kubernetes)
- **k3d** (lightweight k3s in Docker)

See `../../k8s/` directory for local deployment manifests.