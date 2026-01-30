# Database Operations Guide

Comprehensive guide for database operations in the CRM Platform including automated backups, PITR, replication, connection pooling, and blue-green deployments.

## Table of Contents

- [Automated Backups](#automated-backups)
- [Point-in-Time Recovery (PITR)](#point-in-time-recovery-pitr)
- [Replication](#replication)
- [Connection Pooling](#connection-pooling)
- [Blue-Green Deployments](#blue-green-deployments)
- [Monitoring](#monitoring)

---

## Automated Backups

### Overview

Backups are performed automatically via Kubernetes CronJobs:
- **PostgreSQL**: Daily at 2:00 AM UTC using `pg_dump`
- **MongoDB**: Daily at 2:30 AM UTC using `mongodump`

Backups are stored in S3-compatible storage (MinIO for local/dev, AWS S3 for production).

### Retention Policy

| Type    | Retention |
|---------|-----------|
| Daily   | 7 days    |
| Weekly  | 4 weeks   |
| Monthly | 12 months |

### Manual Backup

```bash
# PostgreSQL - backup all databases
./scripts/backup-postgres.sh

# PostgreSQL - backup specific database
./scripts/backup-postgres.sh --database crm_iam

# MongoDB
./scripts/backup-mongodb.sh

# Dry run (show what would be done)
./scripts/backup-postgres.sh --dry-run
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PG_HOST` | PostgreSQL host | localhost |
| `PG_PORT` | PostgreSQL port | 5432 |
| `PG_USER` | PostgreSQL user | crm_admin |
| `PG_PASSWORD` | PostgreSQL password | - |
| `S3_ENABLED` | Enable S3 upload | false |
| `S3_BUCKET` | S3 bucket name | crm-backups |
| `S3_ENDPOINT` | S3 endpoint URL | - |

### Kubernetes CronJob

```bash
# Check CronJob status
kubectl get cronjobs -n crm

# View recent backup jobs
kubectl get jobs -n crm | grep backup

# View backup logs
kubectl logs job/postgres-backup-<timestamp> -n crm
```

---

## Point-in-Time Recovery (PITR)

### Prerequisites

1. WAL archiving enabled in `postgresql.conf`:
   ```
   archive_mode = on
   archive_command = 'aws s3 cp %p s3://crm-backups/wal/%f'
   ```

2. Base backup available

### Recovery Steps

1. **List available backups**:
   ```bash
   ./scripts/restore-postgres.sh --list-backups
   ```

2. **Restore to specific point in time**:
   ```bash
   ./scripts/restore-postgres.sh \
     --backup /backups/crm_iam_20240115_020000.dump \
     --pitr \
     --target-time "2024-01-15 14:30:00"
   ```

3. **Full database restore**:
   ```bash
   ./scripts/restore-postgres.sh \
     --backup /backups/crm_iam_20240115_020000.dump \
     --database crm_iam \
     --target crm_iam_restored
   ```

> [!CAUTION]
> PITR requires stopping PostgreSQL and replacing the data directory. Always test recovery procedures in a non-production environment first.

---

## Replication

### PostgreSQL Streaming Replication

Architecture: 1 Primary + 2 Read Replicas

```
┌─────────────┐     Streaming     ┌─────────────┐
│   Primary   │ ───────────────▶ │  Replica 1  │
│ (postgres-0)│                   │ (postgres-1)│
└──────┬──────┘                   └─────────────┘
       │
       │ Streaming
       ▼
┌─────────────┐
│  Replica 2  │
│ (postgres-2)│
└─────────────┘
```

#### Services

| Service | Purpose | Endpoint |
|---------|---------|----------|
| `postgres-primary` | Read-Write (Primary only) | postgres-primary:5432 |
| `postgres-read` | Read-Only (All replicas) | postgres-read:5432 |
| `postgres-headless` | StatefulSet DNS | postgres-{0,1,2}.postgres-headless:5432 |

#### Checking Replication Status

```sql
-- On primary: view replication slots
SELECT * FROM pg_replication_slots;

-- On primary: view connected replicas
SELECT client_addr, state, sent_lsn, write_lsn 
FROM pg_stat_replication;

-- On replica: check replication lag
SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;
```

### MongoDB Replica Set

Architecture: 1 Primary + 2 Secondaries with automatic failover.

```bash
# Check replica set status
mongosh --eval "rs.status()"

# Check who is primary
mongosh --eval "rs.isMaster()"
```

---

## Connection Pooling

### PgBouncer Configuration

PgBouncer provides connection pooling to reduce database connection overhead.

| Setting | Value | Description |
|---------|-------|-------------|
| Pool Mode | Transaction | Connection per transaction |
| Max Clients | 1000 | Maximum client connections |
| Default Pool | 20 | Connections per database/user |
| Min Pool | 5 | Minimum connections |

#### Connection Strings

```bash
# Through PgBouncer (recommended for applications)
postgresql://crm_admin:password@pgbouncer:6432/crm_iam

# Direct to PostgreSQL (for admin tasks only)
postgresql://crm_admin:password@postgres-primary:5432/crm_iam
```

#### Monitoring PgBouncer

```bash
# Connect to admin console
psql -h pgbouncer -p 6432 -U pgbouncer_admin pgbouncer

# Show pools
SHOW POOLS;

# Show stats
SHOW STATS;

# Show active connections
SHOW CLIENTS;
```

---

## Blue-Green Deployments

### Overview

Blue-green deployment enables zero-downtime updates by maintaining two identical environments.

```
                    ┌─────────────┐
                    │   Ingress   │
                    └──────┬──────┘
                           │
              ┌────────────┴────────────┐
              ▼                         ▼
    ┌─────────────────┐       ┌─────────────────┐
    │  Blue (Active)  │       │ Green (Standby) │
    │    v1.0.0       │       │    v1.1.0       │
    └─────────────────┘       └─────────────────┘
```

### Commands

```bash
# Deploy new version to green
./scripts/blue-green-deploy.sh deploy iam-service v1.1.0

# Check deployment status
./scripts/blue-green-deploy.sh status iam-service

# Switch traffic to green
./scripts/blue-green-deploy.sh switch iam-service

# Rollback to blue (if issues found)
./scripts/blue-green-deploy.sh rollback iam-service
```

### Options

| Flag | Description |
|------|-------------|
| `--namespace <ns>` | Kubernetes namespace (default: crm) |
| `--skip-migration` | Skip database migration check |
| `--dry-run` | Show what would be done |

---

## Monitoring

### Key Metrics

#### PostgreSQL
- Connection count (`pg_stat_activity`)
- Replication lag (`pg_stat_replication`)
- Transaction rate
- Lock waits

#### MongoDB
- Replica set health (`rs.status()`)
- Oplog window
- Connection pool usage

#### PgBouncer
- Active/waiting clients (`SHOW POOLS`)
- Connection wait time
- Query count

### Prometheus Endpoints

| Service | Endpoint |
|---------|----------|
| PostgreSQL | `:9187/metrics` (postgres_exporter) |
| MongoDB | `:9216/metrics` (mongodb_exporter) |
| PgBouncer | `:9127/metrics` (pgbouncer_exporter) |

### Alerts (Recommended)

| Alert | Threshold |
|-------|-----------|
| Replication Lag | > 30 seconds |
| Connection Pool Full | > 90% |
| Backup Failed | Any failure |
| Disk Usage | > 80% |

---

## Quick Reference

### Backup Commands

```bash
# Manual backup
./scripts/backup-postgres.sh
./scripts/backup-mongodb.sh

# Restore
./scripts/restore-postgres.sh --backup <file> --database <db>

# PITR
./scripts/restore-postgres.sh --backup <file> --pitr --target-time "YYYY-MM-DD HH:MM:SS"
```

### Deployment Commands

```bash
# Blue-green deploy
./scripts/blue-green-deploy.sh deploy <service> <version>
./scripts/blue-green-deploy.sh switch <service>
./scripts/blue-green-deploy.sh rollback <service>
```

### Docker Compose (Production)

```bash
# Start production stack
docker-compose -f docker-compose.prod.yml up -d

# Check status
docker-compose -f docker-compose.prod.yml ps

# View logs
docker-compose -f docker-compose.prod.yml logs -f postgres-primary
```

### Kubernetes

```bash
# Apply all resources
kubectl apply -k deployments/kubernetes/base

# Check CronJobs
kubectl get cronjobs -n crm

# Scale replicas
kubectl scale statefulset postgres --replicas=3 -n crm
```
