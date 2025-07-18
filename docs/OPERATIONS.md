# Operations Guide

This guide covers operational procedures, monitoring, and maintenance for the Maintenance API in production environments.

## 📊 Service Level Objectives (SLOs)

### Availability SLO
- **Target**: 99.9% uptime
- **Measurement**: HTTP 200 responses / Total requests
- **Alert**: < 99.9% for 5 minutes

### Latency SLO
- **Target**: 95th percentile < 500ms
- **Measurement**: HTTP request duration
- **Alert**: > 500ms for 5 minutes

### Error Rate SLO
- **Target**: < 1% error rate
- **Measurement**: 5xx responses / Total requests
- **Alert**: > 1% for 5 minutes

## 🔍 Monitoring & Alerting

### Key Metrics Dashboards

#### Application Metrics
- **Request Rate**: Requests per second
- **Error Rate**: Percentage of 4xx/5xx responses
- **Latency**: P50, P95, P99 response times
- **Active Connections**: Current database connections

#### Infrastructure Metrics
- **CPU Usage**: Container and node CPU utilization
- **Memory Usage**: Container and node memory usage
- **Disk I/O**: Read/write operations
- **Network I/O**: Ingress/egress traffic

#### Database Metrics
- **Connection Pool**: Active/idle connections
- **Query Performance**: Slow query log
- **Replication Lag**: If using read replicas
- **Storage**: Database size and growth

### Alerting Rules

#### Critical Alerts (PagerDuty/Slack)
- Service down for > 1 minute
- Database connection failure
- High error rate (> 5%)
- Memory usage > 90%

#### Warning Alerts (Slack/Email)
- High CPU usage (> 80%)
- High latency (> 500ms)
- Low disk space (< 10%)
- Certificate expiration (< 30 days)

## 🚨 Incident Response

### Severity Levels

#### P0 - Critical
- **Definition**: Complete service outage
- **Response Time**: < 5 minutes
- **Resolution Time**: < 1 hour
- **Examples**: Database down, all pods crashing

#### P1 - High
- **Definition**: Major functionality impaired
- **Response Time**: < 15 minutes
- **Resolution Time**: < 4 hours
- **Examples**: High error rate, slow responses

#### P2 - Medium
- **Definition**: Minor functionality issues
- **Response Time**: < 1 hour
- **Resolution Time**: < 1 day
- **Examples**: Single endpoint failing, intermittent issues

#### P3 - Low
- **Definition**: Cosmetic issues
- **Response Time**: < 1 day
- **Resolution Time**: < 1 week
- **Examples**: Log formatting, documentation issues

### Incident Response Playbook

#### 1. Detection
- Monitor alerts from Prometheus/Grafana
- Check application logs
- Verify user reports

#### 2. Assessment
```bash
# Check pod status
kubectl get pods -n maintenance-api

# Check service endpoints
kubectl get endpoints -n maintenance-api

# Check recent logs
kubectl logs -f deployment/maintenance-api -n maintenance-api --tail=100
```

#### 3. Communication
- Create incident in status page
- Notify stakeholders via Slack
- Update incident timeline

#### 4. Mitigation
```bash
# Scale up if needed
kubectl scale deployment maintenance-api --replicas=5 -n maintenance-api

# Restart pods
kubectl rollout restart deployment/maintenance-api -n maintenance-api

# Check database
kubectl exec -it <mysql-pod> -n maintenance-api -- mysql -u user -p -e "SHOW PROCESSLIST;"
```

#### 5. Resolution
- Verify fix with smoke tests
- Update status page
- Post-incident review

## 🔧 Maintenance Procedures

### Daily Tasks
- [ ] Check overnight alerts
- [ ] Review error logs
- [ ] Verify backup completion
- [ ] Monitor resource usage

### Weekly Tasks
- [ ] Review SLO performance
- [ ] Update dependencies
- [ ] Security scan
- [ ] Performance analysis

### Monthly Tasks
- [ ] Capacity planning review
- [ ] Disaster recovery test
- [ ] Security audit
- [ ] Documentation updates

### Quarterly Tasks
- [ ] Architecture review
- [ ] Cost optimization
- [ ] Team training
- [ ] Tool evaluation

## 💾 Backup & Recovery

### Database Backup Strategy

#### Automated Backups
- **Frequency**: Daily at 2 AM UTC
- **Retention**: 30 days
- **Storage**: AWS S3 / Google Cloud Storage
- **Encryption**: AES-256

#### Manual Backup
```bash
# Create backup
kubectl exec -it <mysql-pod> -n maintenance-api -- \
  mysqldump -u user -p tasks_db > backup-$(date +%Y%m%d-%H%M%S).sql

# Upload to cloud storage
aws s3 cp backup-*.sql s3://maintenance-api-backups/
```

### Recovery Procedures

#### Database Recovery
```bash
# Restore from backup
kubectl exec -i <mysql-pod> -n maintenance-api -- \
  mysql -u user -p tasks_db < backup-20240101-020000.sql

# Verify data
kubectl exec -it <mysql-pod> -n maintenance-api -- \
  mysql -u user -p -e "SELECT COUNT(*) FROM tasks;"
```

#### Application Recovery
```bash
# Rollback deployment
kubectl rollout undo deployment/maintenance-api -n maintenance-api

# Check rollback status
kubectl rollout status deployment/maintenance-api -n maintenance-api
```

## 📈 Capacity Planning

### Resource Requirements

#### Current Baseline
- **CPU**: 100m per pod
- **Memory**: 128Mi per pod
- **Storage**: 1Gi for database
- **Network**: 1Mbps per pod

#### Growth Projections
- **Users**: 10% monthly growth
- **Requests**: 15% monthly growth
- **Data**: 5% monthly growth

### Scaling Triggers

#### Horizontal Pod Autoscaler
- **Scale up**: CPU > 70% for 2 minutes
- **Scale down**: CPU < 30% for 5 minutes
- **Min replicas**: 2
- **Max replicas**: 20

#### Cluster Autoscaler
- **Scale up**: Node CPU > 80%
- **Scale down**: Node CPU < 20% for 10 minutes
- **Min nodes**: 2
- **Max nodes**: 10

## 🔐 Security Operations

### Security Monitoring
- **Container scanning**: Trivy on every build
- **Dependency scanning**: Dependabot alerts
- **Runtime monitoring**: Falco for container security
- **Network policies**: Restrict pod-to-pod communication

### Security Response
```bash
# Check for vulnerabilities
trivy image ghcr.io/yourusername/maintenance-api:latest

# Update base image
# Edit Dockerfile to use newer base image
docker build -t maintenance-api:security-patch .
```

## 🧪 Testing Procedures

### Smoke Tests
```bash
# Run smoke tests
k6 run load-testing/k6-smoke-test.js

# Check endpoints
curl -f http://localhost:8080/health/ready
curl -f http://localhost:8080/health/live
```

### Load Tests
```bash
# Run load tests
k6 run load-testing/k6-load-test.js

# Monitor during test
kubectl top pods -n maintenance-api
kubectl top nodes
```

### Chaos Engineering
```bash
# Install chaos mesh
helm install chaos-mesh chaos-mesh/chaos-mesh --namespace chaos-testing

# Create pod failure experiment
kubectl apply -f chaos/pod-failure.yaml
```

## 📋 Runbooks

### High CPU Usage
1. **Symptoms**: High CPU alerts, slow response times
2. **Diagnosis**:
   ```bash
   kubectl top pods -n maintenance-api
   kubectl logs -f deployment/maintenance-api -n maintenance-api
   ```
3. **Resolution**:
   - Scale up pods: `kubectl scale deployment maintenance-api --replicas=5`
   - Check for infinite loops in code
   - Review database queries

### Database Connection Issues
1. **Symptoms**: Database connection errors
2. **Diagnosis**:
   ```bash
   kubectl get pods -n maintenance-api
   kubectl logs <mysql-pod> -n maintenance-api
   ```
3. **Resolution**:
   - Restart MySQL pod
   - Check database credentials
   - Verify network connectivity

### Memory Leaks
1. **Symptoms**: High memory usage, OOM kills
2. **Diagnosis**:
   ```bash
   kubectl top pods -n maintenance-api
   kubectl describe pod <pod-name> -n maintenance-api
   ```
3. **Resolution**:
   - Restart affected pods
   - Review memory usage patterns
   - Increase memory limits

## 📞 Contact Information

### On-call Rotation
- **Primary**: Platform Team
- **Secondary**: Backend Team
- **Escalation**: Engineering Manager

### Communication Channels
- **Slack**: #maintenance-api-alerts
- **Email**: platform-team@company.com
- **PagerDuty**: Maintenance API Service

### Documentation Links
- [Architecture Diagram](architecture/README.md)
- [API Documentation](docs/API.md)
- [Runbooks](docs/RUNBOOKS.md)
- [Post-mortems](docs/POSTMORTEMS.md)