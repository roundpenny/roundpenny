# Security

## Security Practices

Roundup Platform follows security best practices to protect user data and system integrity.

## TLS Configuration

All inter-service communication must use TLS in production. The Kong API gateway terminates TLS at the edge. Service-to-service communication within the Kubernetes cluster uses mTLS via service mesh (Istio/Linkerd).

Configuration in Helm values:
```yaml
global:
  tls:
    enabled: true
    certManager:
      clusterIssuer: letsencrypt-prod
```

## JWT Secret Management

- JWT secrets are never hardcoded in source code
- In production, secrets are provisioned via Kubernetes Secrets or Vault
- Secret rotation is supported via Helm values
- Use strong secrets: `openssl rand -base64 48`

```yaml
auth:
  jwtSecret:
    valueFrom:
      secretKeyRef:
        name: jwt-secret
        key: secret
```

## MFA Implementation

- TOTP-based multi-factor authentication (RFC 6238)
- Uses time-based one-time passwords with 30-second window
- Backup codes provided on MFA enable (single-use)
- Rate limiting on MFA verification endpoint (5 attempts per minute per user)

## Rate Limiting

- Per-IP rate limiting on auth endpoints
- Login: 10 requests/minute
- Register: 5 requests/minute
- MFA verify: 5 requests/minute
- Global rate limiting at Kong gateway layer

## Input Validation

All API inputs are validated before processing:
- Email addresses validated via regex
- Password minimum length: 8 characters
- String length limits enforced
- JSON payload size limited at gateway level (1MB default)
- SQL injection prevented via parameterized queries

## SQL Injection Prevention

All database queries use parameterized queries via `pgx`:

```go
// Safe - parameterized query
row := db.QueryRow(ctx, "SELECT id, email FROM users WHERE email = $1", email)

// Unsafe - DO NOT use string interpolation
// row := db.QueryRow(ctx, "SELECT id, email FROM users WHERE email = '"+email+"'")
```

All Go database packages (`pkg/db`) enforce parameterized queries. Raw SQL string building is prohibited.

## Audit Logging

All security-relevant events are logged via the audit logging system:
- Authentication events (login, logout, register, refresh)
- MFA lifecycle (setup, enable, disable, verify)
- KYC submissions and status changes
- Account changes (password change, email verification)
- Merchant CRUD operations
- Payment lifecycle events

Audit logs are:
- Structured JSON format
- Tamper-evident (append-only in production)
- Sent to dedicated Kafka topic (`audit-logs`) in production
- Written to stdout in development

## Environment Variables

Sensitive environment variables that must be kept secret:
- `JWT_SECRET`
- `DATABASE_URL` (contains credentials)
- `MFA_ISSUER`
- `KAFKA_PASSWORD` (if SASL enabled)
- `SLACK_WEBHOOK`
- `GITHUB_TOKEN`

## Vulnerability Scanning

- All Docker images are scanned with Trivy in CI/CD pipeline
- Critical and High severity vulnerabilities fail the build
- Go modules are scanned via `golangci-lint` and `govulncheck`

## Reporting Vulnerabilities

Report security vulnerabilities to security@roundup-platform.com.
