# Deployment Guide

## Kubernetes with Helm

### Prerequisites

- Kubernetes cluster 1.28+
- Helm 3.12+
- kubectl configured with cluster access
- Cert-Manager (for TLS)
- NGINX Ingress Controller or equivalent

### Install

```bash
# Add helm repository (if using external registry)
helm repo add roundup https://charts.roundup-platform.com

# Install with default values
helm upgrade --install roundup-platform deploy/helm/roundup-platform \
  --namespace roundup-platform --create-namespace

# Install with custom values
helm upgrade --install roundup-platform deploy/helm/roundup-platform \
  --namespace roundup-platform --create-namespace \
  -f deploy/helm/roundup-platform/values-custom.yaml
```

### Uninstall

```bash
helm uninstall roundup-platform -n roundup-platform
kubectl delete namespace roundup-platform
```

## Production Checklist

- [ ] JWT secrets stored in Kubernetes Secret / Vault
- [ ] TLS enabled with valid certificate
- [ ] Database credentials rotated from defaults
- [ ] Kafka configured with SASL/SSL authentication
- [ ] Resource limits set for all services
- [ ] Horizontal Pod Autoscaler configured
- [ ] PodDisruptionBudgets configured
- [ ] PrometheusServiceMonitor CRDs installed
- [ ] Backups configured for PostgreSQL
- [ ] Network policies applied
- [ ] Audit log shipping configured (S3/GCS/Elasticsearch)
- [ ] Rate limiting thresholds tuned for production
- [ ] Health check and readiness probes configured

## TLS Setup

### With Cert-Manager

```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@roundup-platform.com
    privateKeySecretRef:
      name: letsencrypt-prod-key
    solvers:
    - http01:
        ingress:
          class: nginx

# Ingress annotation:
# cert-manager.io/cluster-issuer: letsencrypt-prod
```

### Manual TLS

Place TLS certificate and key in a Kubernetes secret:

```bash
kubectl create secret tls roundup-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  -n roundup-platform
```

## Scaling Considerations

### Database

- PostgreSQL should use read replicas for analytics workloads
- Connection pooling via PgBouncer recommended for high concurrency
- Migrations run as init containers to avoid race conditions

### Kafka

- Minimum 3 brokers for production deployments
- Audit topic (`audit-logs`) configured with `min.insync.replicas=2`
- Consumer groups configured with `auto.offset.reset=earliest` for critical consumers

### Services

- Stateless services (auth, user, merchant, analytics, notification) scale horizontally
- Stateful services (wallet, ledger) require careful sharding strategy
- Roundup-engine is CPU-bound; scale based on transaction volume
- Investment-service batch processing tunable via `BATCH_SIZE` and `BATCH_INTERVAL`

### Resource Recommendations

| Service | CPU | Memory | Replicas |
|---------|-----|--------|----------|
| auth-service | 250m-500m | 256Mi-512Mi | 2-4 |
| roundup-engine | 500m-1000m | 512Mi-1Gi | 2-6 |
| transaction-service | 250m-500m | 256Mi-512Mi | 2-4 |
| wallet-service | 250m-500m | 256Mi-512Mi | 2-4 |
| fee-service | 100m-250m | 128Mi-256Mi | 2-3 |
| investment-service | 500m-1000m | 512Mi-1Gi | 2-4 |
| ledger-service | 250m-500m | 256Mi-512Mi | 2-4 |
| user-service | 100m-250m | 128Mi-256Mi | 2-3 |
| payment-gateway | 250m-500m | 256Mi-512Mi | 2-4 |
| fraud-service | 500m-1000m | 512Mi-1Gi | 2-4 |
| merchant-service | 100m-250m | 128Mi-256Mi | 2-3 |
| analytics-service | 250m-500m | 256Mi-512Mi | 2-3 |
| notification-service | 100m-250m | 128Mi-256Mi | 2-3 |

### Monitoring Alerts

Configure Prometheus alerting rules for:

- `HighErrorRate` - HTTP 5xx rate > 5% for 5 minutes
- `HighLatency` - p99 latency > 5s for 5 minutes
- `ServiceDown` - Service endpoint down for 1 minute
- `KafkaLagHigh` - Consumer lag > 1000 messages
- `PodCrashLooping` - Pod restarting frequently
- `DiskSpaceLow` - Persistent volume usage > 80%

## Secret Management

### Kubernetes Secrets

```bash
# JWT signing key (min 32 chars)
kubectl create secret generic jwt-secret \
  --from-literal=jwt-secret=$(openssl rand -base64 48) \
  -n roundup-platform

# Stripe API keys (only needed for payment-gateway)
kubectl create secret generic stripe-secret \
  --from-literal=stripe-api-key=sk_live_... \
  --from-literal=stripe-webhook-secret=whsec_... \
  -n roundup-platform

# Database credentials
kubectl create secret generic db-secret \
  --from-literal=postgres-user=roundup \
  --from-literal=postgres-password=$(openssl rand -base64 32) \
  -n roundup-platform
```

Mount secrets as env vars (`secretKeyRef`) or volumes (`JWT_SECRET_FILE`) (preferred for zero-copy in env dumps). All services use `JWT_SECRET_FILE` env var to read JWT secret from a mounted file (compatible with Docker secrets, Vault agent sidecar, or K8s projected volumes).

### AWS Secrets Manager (via CSI Driver)

For production on EKS:

1. Install [Secrets Store CSI Driver](https://secrets-store-csi-driver.sigs.k8s.io/):
   ```bash
   helm repo add secrets-store-csi-driver https://kubernetes-sigs.github.io/secrets-store-csi-driver/charts
   helm install csi-secrets-store secrets-store-csi-driver/csi-secrets-store \
     --namespace kube-system
   ```

2. Create Provider:
   ```bash
   helm install -n kube-system secrets-provider-aws \
     https://aws.github.io/secrets-store-csi-driver-provider-aws/deploy/provider-aws-installer.yaml
   ```

3. Create IAM policy + service account (IRSA) for the CSI driver.

4. Mount secrets as a volume (each service reads via `JWT_SECRET_FILE`):
   ```yaml
   volumes:
   - name: secrets-store
     csi:
       driver: secrets-store.csi.k8s.io
       readOnly: true
       volumeAttributes:
         secretProviderClass: "roundup-platform-secrets"
   ```

## Backup & Disaster Recovery

### PostgreSQL

Two strategies (use both for defense-in-depth):

**Strategy A: pg_dump (logical backup)**
```bash
# Kubernetes CronJob: weekly full dump
PGPASSWORD=$POSTGRES_PASSWORD pg_dump \
  -h $POSTGRES_HOST \
  -U $POSTGRES_USER \
  -d $POSTGRES_DB \
  -F c -f /backups/roundup-$(date +%Y%m%d).dump

# Restore
pg_restore -h $POSTGRES_HOST -U $POSTGRES_USER -d $POSTGRES_DB \
  -c /backups/roundup-YYYYMMDD.dump
```

**Strategy B: WAL archiving (point-in-time recovery)**
- Enable `archive_mode=on` and `archive_command` in postgresql.conf
- Archive WALs to S3 using `pg_receivewal` + `wal-g` or `pgBackRest`
- Retain WALs for 7+ days (adjust `wal_keep_size`)

**Kubernetes CronJob example** (included in Helm chart at `templates/backup-cronjob.yaml`):
- Schedule: `0 2 * * 0` (weekly Sunday 02:00)
- Retention: 30 days (S3 lifecycle policy)
- Backup format: `pg_dump` compressed, uploaded to S3

### Kafka

- Replication factor: 3 (set at topic creation via `--replication-factor 3`)
- `min.insync.replicas: 2` for critical topics (transactions, audit-logs)
- MSK auto-minor-version upgrade enabled
- Cross-region replication via MirrorMaker 2 if multi-region required

### Terraform State

Remote state in S3 with DynamoDB locking:
```hcl
terraform {
  backend "s3" {
    bucket         = "roundup-platform-terraform-state"
    key            = "prod/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}
```

### Chaos Engineering

Use `chaos-mesh` or `litmus` to periodically test resilience:
- Random pod kills (validate PDB and HPA)
- Network latency injection (validate circuit breakers)
- Kafka broker failure (validate consumer rebalancing)

### Recovery Time Objective (RTO) / Recovery Point Objective (RPO)

| Component | RTO | RPO | Method |
|-----------|-----|-----|--------|
| PostgreSQL | 1h | 5min | WAL archiving + hot standby |
| Kafka | 15min | 0 (MSK HA) | Auto-rebalancing |
| Services | 5min | 0 (stateless) | HPA + rolling update |
| TLS certs | 1h | — | cert-manager auto-renew |
| Terraform state | 15min | 24h | S3 versioning |
