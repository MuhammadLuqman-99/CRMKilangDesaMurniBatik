# Deployment Guide

> Production deployment guide for CRM Kilang Desa Murni Batik

---

## Prerequisites

Before deploying, ensure you have:

- ✅ **Kubernetes Cluster** (v1.25+) — EKS, GKE, AKS, or self-managed
- ✅ **Helm 3.x** — Package manager for Kubernetes
- ✅ **kubectl** — Configured with cluster access
- ✅ **Container Registry** — Docker Hub, ECR, GCR, or private registry
- ✅ **External Services**:
  - PostgreSQL 15+ (or use managed: RDS, Cloud SQL)
  - MongoDB 6.0+ (or use MongoDB Atlas)
  - Redis 7.0+ (or use ElastiCache, Memorystore)
  - RabbitMQ 3.12+ (or use CloudAMQP)
  - S3-compatible storage (for backups)

---

## Quick Start

### 1. Clone and Navigate

```bash
git clone https://github.com/MuhammadLuqman-99/CRMKilangDesaMurniBatik.git
cd CRMKilangDesaMurniBatik/deployments/helm
```

### 2. Create Namespace

```bash
kubectl create namespace crm
```

### 3. Configure Secrets

Create a `secrets.yaml` or use Kubernetes secrets directly:

```bash
# Database credentials
kubectl create secret generic db-credentials \
  --namespace crm \
  --from-literal=postgres-host=your-postgres-host \
  --from-literal=postgres-user=crm_user \
  --from-literal=postgres-password=YOUR_SECURE_PASSWORD \
  --from-literal=postgres-db=crm_db \
  --from-literal=mongo-uri=mongodb://user:pass@host:27017/crm

# Redis credentials
kubectl create secret generic redis-credentials \
  --namespace crm \
  --from-literal=redis-host=your-redis-host \
  --from-literal=redis-password=YOUR_REDIS_PASSWORD

# JWT and auth secrets
kubectl create secret generic auth-secrets \
  --namespace crm \
  --from-literal=jwt-secret=YOUR_JWT_SECRET_KEY \
  --from-literal=jwt-refresh-secret=YOUR_REFRESH_SECRET

# S3 credentials for backups
kubectl create secret generic s3-credentials \
  --namespace crm \
  --from-literal=aws-access-key=YOUR_ACCESS_KEY \
  --from-literal=aws-secret-key=YOUR_SECRET_KEY \
  --from-literal=s3-bucket=crm-backups
```

### 4. Configure Values

Edit `crm-platform/values-prod.yaml`:

```yaml
global:
  environment: production
  domain: api.your-domain.com
  
image:
  registry: your-registry.com
  tag: v1.0.0

ingress:
  enabled: true
  className: nginx
  annotations:
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: api.your-domain.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: crm-tls
      hosts:
        - api.your-domain.com

resources:
  iam:
    requests: { cpu: "100m", memory: "128Mi" }
    limits: { cpu: "500m", memory: "512Mi" }
  customer:
    requests: { cpu: "100m", memory: "128Mi" }
    limits: { cpu: "500m", memory: "512Mi" }
  sales:
    requests: { cpu: "100m", memory: "128Mi" }
    limits: { cpu: "500m", memory: "512Mi" }
  notification:
    requests: { cpu: "100m", memory: "128Mi" }
    limits: { cpu: "500m", memory: "512Mi" }
```

### 5. Deploy with Helm

```bash
helm install crm-platform ./crm-platform \
  --namespace crm \
  --values crm-platform/values-prod.yaml
```

### 6. Verify Deployment

```bash
# Check all pods are running
kubectl get pods -n crm

# Check services
kubectl get svc -n crm

# Check ingress
kubectl get ingress -n crm

# View logs
kubectl logs -n crm -l app=api-gateway --tail=100
```

---

## Configuration Reference

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DATABASE_URL` | PostgreSQL connection string | ✓ |
| `MONGO_URI` | MongoDB connection string | ✓ |
| `REDIS_URL` | Redis connection string | ✓ |
| `RABBITMQ_URL` | RabbitMQ connection string | ✓ |
| `JWT_SECRET` | JWT signing secret | ✓ |
| `JWT_REFRESH_SECRET` | Refresh token secret | ✓ |
| `AWS_ACCESS_KEY_ID` | S3 access key | For backups |
| `AWS_SECRET_ACCESS_KEY` | S3 secret key | For backups |
| `SENDGRID_API_KEY` | SendGrid API key | For emails |
| `TWILIO_ACCOUNT_SID` | Twilio account SID | For SMS |
| `TWILIO_AUTH_TOKEN` | Twilio auth token | For SMS |

---

## Monitoring Setup

### Prometheus & Grafana

The Helm chart includes monitoring components:

```bash
# Enable monitoring in values
monitoring:
  enabled: true
  prometheus:
    retention: 15d
  grafana:
    adminPassword: your-grafana-password
```

**Access Grafana:**

```bash
# Port forward
kubectl port-forward svc/grafana -n crm 3000:80

# Open http://localhost:3000
# Login: admin / your-grafana-password
```

**Pre-configured Dashboards:**
- CRM Overview — Request rates, latencies, errors
- Service Health — Per-service metrics
- Database Metrics — Connection pools, query times
- RabbitMQ — Queue depths, message rates

### Alertmanager

Configure alerting in `values.yaml`:

```yaml
alerting:
  enabled: true
  slack:
    webhookUrl: https://hooks.slack.com/services/xxx
    channel: "#crm-alerts"
  email:
    to: ops@your-domain.com
    from: alerts@your-domain.com
```

### Loki Logging

View logs via Grafana:

1. Navigate to **Explore** → Select **Loki**
2. Query examples:
   ```
   {namespace="crm", app="api-gateway"}
   {namespace="crm"} |= "error"
   {namespace="crm", app="iam-service"} | json | level="error"
   ```

---

## Database Operations

### Automated Backups

Backups are configured via CronJobs:

```yaml
backup:
  enabled: true
  schedule: "0 2 * * *"  # Daily at 2 AM
  retention: 30  # Keep 30 days
  storage:
    type: s3
    bucket: crm-backups
    region: ap-southeast-1
```

### Manual Backup

```bash
# PostgreSQL
kubectl exec -n crm deploy/postgres-backup -- \
  pg_dump -h $POSTGRES_HOST -U $POSTGRES_USER $POSTGRES_DB | \
  gzip > backup-$(date +%Y%m%d).sql.gz

# MongoDB
kubectl exec -n crm deploy/mongo-backup -- \
  mongodump --uri=$MONGO_URI --archive --gzip > mongo-backup.gz
```

### Point-in-Time Recovery (PITR)

```bash
# Restore PostgreSQL to specific time
kubectl exec -n crm deploy/postgres-restore -- \
  /scripts/pitr-restore.sh --target-time="2026-01-30 10:00:00"
```

---

## Scaling

### Horizontal Pod Autoscaling (HPA)

Pre-configured HPA for all services:

```yaml
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
  targetMemoryUtilizationPercentage: 80
```

### Manual Scaling

```bash
kubectl scale deployment -n crm api-gateway --replicas=5
```

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| Pods CrashLoopBackOff | Check logs: `kubectl logs -n crm <pod>` |
| Database connection failed | Verify secrets and network policies |
| High latency | Check HPA metrics, scale if needed |
| 5xx errors | Check service logs, verify external deps |

### Useful Commands

```bash
# Get all resources
kubectl get all -n crm

# Describe failing pod
kubectl describe pod -n crm <pod-name>

# Check events
kubectl get events -n crm --sort-by=.metadata.creationTimestamp

# Execute into container
kubectl exec -it -n crm <pod-name> -- /bin/sh

# Check resource usage
kubectl top pods -n crm
```

---

## Rollback

```bash
# View release history
helm history crm-platform -n crm

# Rollback to previous version
helm rollback crm-platform 1 -n crm

# Rollback to specific revision
helm rollback crm-platform 3 -n crm
```
