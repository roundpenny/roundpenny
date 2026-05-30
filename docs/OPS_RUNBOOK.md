# RoundPenny Operations Runbook

## 1. Architecture Overview

### Service Map

```
                          ┌─────────────┐
                          │   Kong API   │
                          │  Gateway 3.5 │
                          │ :80 / :443   │
                          └──────┬──────┘
                                 │
                    ┌────────────┼────────────┐
                    │            │            │
              ┌─────▼─────┐ ┌───▼────┐ ┌─────▼─────┐
              │   Auth    │ │  User  │ │ Merchant  │
              │ Service   │ │ Service│ │ Service   │
              │ :8081     │ │ :8088  │ │ :8092     │
              └─────┬─────┘ └───┬────┘ └─────┬─────┘
                    │            │            │
              ┌─────▼───────────▼────────────▼─────┐
              │         Transaction Service        │
              │              :8083                  │
              └─────┬───────────┬──────────────────┘
                    │           │
         ┌──────────▼──┐  ┌────▼────────┐
         │ Roundup     │  │   Wallet    │
         │ Engine      │  │   Service   │
         │ :8082       │  │   :8084     │
         └──────┬──────┘  └────┬────────┘
                │              │
         ┌──────▼──────┐ ┌────▼────────┐
         │     Fee     │ │  Ledger     │
         │   Service   │ │  Service    │
         │   :8085     │ │  :8087      │
         └──────┬──────┘ └────┬────────┘
                │              │
    ┌───────────┼──────────────┼───────────┐
    │           │              │           │
┌───▼────┐ ┌───▼────┐  ┌─────▼──┐  ┌─────▼─────┐
│Payment │ │Fraud   │  │Notific.│  │Subscription│
│Gateway │ │Service │  │Service │  │  Service   │
│:8090   │ │:8094   │  │:8091   │  │  :8096     │
└────────┘ └────────┘  └────────┘  └───────────┘

┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  PostgreSQL  │   │    Redis     │   │    Kafka     │
│     :5432    │   │    :6379     │   │    :9092     │
└──────────────┘   └──────────────┘   └──────────────┘

┌──────────────┐   ┌──────────────┐   ┌──────────────┐
│  Prometheus  │   │   Grafana    │   │    Loki      │
│     :9090    │   │    :3000     │   │    :3100     │
└──────────────┘   └──────────────┘   └──────────────┘

┌──────────────┐   ┌──────────────┐
│    Tempo     │   │  Alertmanager│
│    :3200     │   │    :9093     │
└──────────────┘   └──────────────┘
```

### Port Assignments

| Service              | Port  |
|----------------------|-------|
| Kong Proxy           | 80    |
| Kong Proxy (TLS)     | 443   |
| Kong Admin           | 8001  |
| Grafana              | 3000  |
| Loki                 | 3100  |
| Tempo                | 3200  |
| PostgreSQL           | 5432  |
| Redis                | 6379  |
| Auth Service         | 8081  |
| Roundup Engine       | 8082  |
| Transaction Service  | 8083  |
| Wallet Service       | 8084  |
| Fee Service          | 8085  |
| Investment Service   | 8086  |
| Ledger Service       | 8087  |
| User Service         | 8088  |
| Payment Gateway      | 8090  |
| Notification Service | 8091  |
| Merchant Service     | 8092  |
| Analytics Service    | 8093  |
| Fraud Service        | 8094  |
| Admin Service        | 8095  |
| Subscription Service | 8096  |
| Swagger UI           | 8080  |
| Prometheus           | 9090  |
| Alertmanager         | 9093  |
| Redis Exporter       | 9121  |

### Data Flow

```
User → Kong → Auth (JWT validation)
           → Transaction Service → Kafka (tx.settled)
           → Roundup Engine → Kafka (roundup.calculated)
           → Wallet Service → Kafka (wallet.credited)
           → Fee Service → Kafka (fee.charged)
           → Ledger Service (double-entry bookkeeping)
           → Notification Service (email/push)
```

## 2. Deployment

### 2.1 Local Development

```bash
docker compose up -d
docker compose logs -f
```

### 2.2 Production Deployment

**Helm Upgrade:**
```bash
helm upgrade --install roundup-platform ./deploy/helm/roundup-platform \
  --namespace roundup-platform \
  --create-namespace \
  --values ./deploy/helm/roundup-platform/values-production.yaml \
  --set global.imageTag=<commit-sha>
```

**Terraform Apply:**
```bash
cd deploy/terraform
terraform init
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```

**Rollback Procedure:**
```bash
# Helm rollback
helm rollback roundup-platform 1 --namespace roundup-platform

# Terraform rollback
terraform plan -destroy -var-file=terraform.tfvars
terraform apply -destroy -var-file=terraform.tfvars
```

## 3. Monitoring & Alerting

### 3.1 Dashboards

- **Grafana**: http://grafana:3000 (admin/admin)
- **Prometheus**: http://prometheus:9090
- **Alertmanager**: http://alertmanager:9093

Key Grafana panels:
- Request rate (req/s per service)
- Error rate (5xx percentage)
- Latency p95
- Circuit breaker state
- Kafka consumer queue depth
- PostgreSQL connection pool saturation

### 3.2 Alert Rules

| #  | Rule Name            | Description                                    | Severity |
|----|----------------------|------------------------------------------------|----------|
| 1  | ServiceDown          | Service not responding to health checks        | critical |
| 2  | HighErrorRate        | Error rate > 5% over 5 minutes                 | critical |
| 3  | HighLatency          | p95 latency > 1 second                         | warning  |
| 4  | CircuitBreakerOpen   | Circuit breaker triggered for upstream service | critical |
| 5  | KafkaLagHigh         | Consumer lag > 1000 messages                   | warning  |
| 6  | RedisDown            | Redis instance unreachable                     | critical |
| 7  | LowDiskSpace         | Disk usage < 10% free                          | warning  |
| 8  | HighMemoryUsage      | Memory usage > 90%                             | warning  |
| 9  | HighCPUUsage         | CPU usage > 90%                                | warning  |

### 3.3 Logging

- **Loki**: http://loki:3100
- **Log levels**: `debug`, `info`, `warn`, `error`
- **Structured format**: `key=value` pairs (e.g., `level=info service=auth-service request_id=abc123`)
- **Aggregation**: Promtail collects Docker logs, ships to Loki

## 4. Database

### 4.1 PostgreSQL

- **Connection**: `postgres://roundup:roundup@postgres:5432/roundup?sslmode=disable`
- **Migrations**: Run automatically on service startup via `MIGRATIONS_DIR` env var
- **Soft delete**: `deleted_at TIMESTAMPTZ` column with partial index `WHERE deleted_at IS NULL`
- **Production**: RDS Multi-AZ with automated backups (retention: 30 days)

### 4.2 Redis

- **Connection**: `redis:6379`
- **Use cases**: Rate limiting, caching, idempotency keys, Kafka offset commits
- **Persistence**: RDB snapshots every 5 minutes
- **Production**: ElastiCache (Redis) with automatic failover

## 5. Kafka

- **Brokers** (dev): `kafka:9092`
- **Brokers** (prod): MSK cluster, TLS on port 9094

### Topics

| Topic                 | Partitions | Description                    |
|-----------------------|-----------|--------------------------------|
| tx.settled            | 3         | Transaction settled events     |
| roundup.calculated    | 3         | Roundup amount calculated      |
| wallet.credited       | 3         | Wallet credit events           |
| fee.charged           | 3         | Fee charge events              |
| user.registered       | 1         | New user registration          |
| user.updated          | 1         | User profile updates           |
| notification.sent     | 2         | Notification dispatch events   |
| fraud.alert           | 2         | Fraud detection alerts         |
| subscription.created  | 1         | Subscription created           |
| subscription.cancelled| 1         | Subscription cancelled         |
| subscription.renewed  | 1         | Subscription auto-renewal      |
| payment.failed        | 2         | Payment failure events         |

### Consumer Groups

| Group              | Services                  |
|--------------------|---------------------------|
| fee-worker         | fee-service               |
| investment-worker  | investment-service        |
| ledger-worker      | ledger-service            |
| webhook-worker     | payment-gateway           |
| roundup-worker     | roundup-engine            |

## 6. Common Procedures

### 6.1 Scaling a Service

```bash
# Docker Compose
docker compose up -d --scale auth-service=3

# Kubernetes (Helm)
helm upgrade --install roundup-platform ./deploy/helm/roundup-platform \
  --set authService.replicaCount=5 \
  --reuse-values
```

### 6.2 Restarting a Service

```bash
docker compose restart auth-service
```

### 6.3 Viewing Logs

```bash
docker compose logs -f auth-service
docker compose logs -f --tail=100 auth-service
```

### 6.4 Running Migrations Manually

```bash
docker compose exec auth-service ./server -migrate-only
```

### 6.5 Health Check

```bash
curl http://localhost:8000/v1/health
```

## 7. Troubleshooting

### 7.1 Service Won't Start

1. Check logs: `docker compose logs <service>`
2. Verify DB is running: `docker compose exec postgres pg_isready -U roundup`
3. Check migration output: `docker compose logs <service> | Select-String -Pattern "migration"`
4. Verify environment: `docker compose config`
5. Check health check dependencies: `docker compose ps`

### 7.2 High Error Rate

1. Open Grafana dashboard → Error Rate panel
2. Check service logs for 5xx: `docker compose logs <service> | Select-String -Pattern " 5.. "`
3. Verify upstream dependencies are healthy
4. Check circuit breaker state in Grafana
5. Review recent deployments or config changes

### 7.3 Kafka Consumer Lag

1. Check consumer lag:
   ```bash
   docker compose exec kafka kafka-consumer-groups \
     --bootstrap-server localhost:9092 \
     --group <group> \
     --describe
   ```
2. Restart stale consumer: `docker compose restart <service>`
3. Increase partitions if lag is persistent

### 7.4 Database Connection Issues

1. Check pool saturation in Grafana
2. Verify connection string is correct
3. Check for long-running queries: `SELECT * FROM pg_stat_activity WHERE state != 'idle'`
4. Restart connection pool by restarting the service
5. Check `max_connections` on PostgreSQL

### 7.5 Memory/CPU Issues

1. Check resource usage: `docker stats`
2. Review Go memory metrics (heap, goroutines) in Grafana
3. Adjust `GOMEMLIMIT` environment variable
4. Scale horizontally if needed: `docker compose up -d --scale <service>=<N>`

## 8. Security Incidents

### 8.1 Suspected Data Breach

1. Rotate all secrets immediately:
   ```bash
   ./scripts/manage-secrets.sh encrypt
   ```
2. Revoke all JWT tokens by rotating `JWT_SECRET`
3. Enable maintenance mode via Kong
4. Audit logs for suspicious activity (Loki query: `{job=~".+"} |= "error"`)
5. Notify affected users within 72 hours (GDPR requirement)

### 8.2 DDoS Attack

1. Enable Kong rate limiting (already configured in kong.yml)
2. Scale up services: `docker compose up -d --scale auth-service=10`
3. Consider Cloudflare/WAF for production
4. Block offending IPs at Kong level

### 8.3 Service Compromise

1. Isolate compromised container: `docker compose stop <service>`
2. Rotate all service secrets
3. Review audit logs for lateral movement
4. Rebuild from clean image: `docker compose build --no-cache <service>`

## 9. Disaster Recovery

### 9.1 Database Failure

1. Restore from latest backup:
   ```bash
   ./scripts/backup.sh
   ./scripts/restore.sh ./backups/roundpenny_latest.sql.gz
   ```
2. Verify data integrity: `docker compose exec postgres pg_isready -U roundup`
3. Point services to restored DB (update DATABASE_URL)
4. Reprocess Kafka messages from last known good offset

### 9.2 Complete System Failure

1. Provision new infrastructure via Terraform:
   ```bash
   cd deploy/terraform
   terraform apply -var-file=terraform.tfvars
   ```
2. Restore DB from S3 backup
3. Deploy Helm chart:
   ```bash
   helm upgrade --install roundup-platform ./deploy/helm/roundup-platform \
     --values values-production.yaml
   ```
4. Verify all services healthy: `curl http://localhost:8000/v1/health`
5. Reprocess missed Kafka messages

### 9.3 Recovery Objectives

| Metric | Target  |
|--------|---------|
| RTO    | 4 hours |
| RPO    | 1 hour  |
| Critical path | DB restore + service deployment |

## 10. Backup Strategy

### 10.1 PostgreSQL

- **Full backup**: Daily via `pg_dumpall` (custom format)
- **WAL archiving**: Continuous archiving to S3
- **Retention**: 30 days daily, 12 monthly snapshots
- **S3 bucket**: `roundup-platform-backups` (encrypted at rest)
- **Automation**: Kubernetes CronJob runs every 6 hours

### 10.2 Kafka

- **Log retention**: 7 days
- **Offsets**: Stored in Redis (persistent)
- **Replication**: Critical topics replicated 3x across brokers
- **Recovery**: Reprocess from last committed offset after restore

### 10.3 Redis

- **RDB snapshots**: Every 5 minutes
- **AOF**: Enabled for durability
- **Backup**: Via SLAVEOF to a replica instance
- **Production**: ElastiCache with automatic multi-AZ failover

### 10.4 Configuration

- **Terraform state**: S3 bucket `roundup-platform-terraform-state` (DynamoDB locked)
- **Helm values**: Stored in git (`values.yaml`, `values-production.yaml`)
- **Docker Compose**: Stored in git (`docker-compose.yml`)
- **Secrets**: `.env` files encrypted via `manage-secrets.sh`
- **TLS certs**: ACM for production, local certs for development
