# RoundPenny

> Round-up micro-transaction platform — spare change investing, built for scale.

[![Go](https://img.shields.io/badge/Go-1.26-blue?logo=go)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![CI](https://img.shields.io/badge/CI-GitHub%20Actions-brightgreen?logo=github)](.github/workflows/ci.yml)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker)](docker-compose.yml)
[![K8s](https://img.shields.io/badge/K8s-Helm-326CE5?logo=kubernetes)](deploy/helm/)
[![Terraform](https://img.shields.io/badge/Terraform-AWS-7B42BC?logo=terraform)](deploy/terraform/)
[![Kong](https://img.shields.io/badge/API%20Gateway-Kong-003459?logo=kong)](deploy/kong/kong.yml)
[![k6](https://img.shields.io/badge/Load%20Test-k6-7D64FF?logo=k6)](scripts/load-test.js)

---

## Features

| Capability | Details |
|-----------|---------|
| **Round-Up Engine** | Auto-calculates spare change on every transaction |
| **Investment Portfolios** | Micro-investment accounts for users |
| **Multi-Tenant Merchants** | Onboard merchants with fee tiers |
| **Fraud Detection** | Rule-based risk scoring engine |
| **KYC & MFA** | Identity verification + TOTP multi-factor auth |
| **Double-Entry Ledger** | Full accounting audit trail |
| **Stripe Integration** | Payment processing (mock mode for dev) |
| **API Gateway** | Kong-powered rate limiting, auth, routing |
| **Observability** | Prometheus + Grafana + Loki + Tempo |
| **CI/CD** | GitHub Actions → ghcr.io → Helm → EKS |
| **Infra as Code** | Terraform (VPC, EKS, RDS, MSK, ECR, ALB) |

## US Market

RoundPenny is positioned to disrupt the US micro-investing market:

| Metric | Value |
|--------|-------|
| TAM | $1.2T |
| Leader (Acorns) revenue | $60M/mo |
| Gen Z wanting to invest | 73% |
| Market growth (2020→2025) | 340% |

**Key differentiation:** API-first architecture, merchant commission revenue (Acorns has none), open source, multi-tenant by design. Built for B2B2C — neobanks, payroll platforms, and enterprise partners.

### US Pricing

| Tier | Price | Audience |
|------|-------|----------|
| **Direct** | $2/mo + 0.25% round-up commission | B2C |
| **White-Label API** | $5k/mo (unlimited) | Neobanks, fintech partners |
| **Enterprise** | Custom | On-prem, source code, 24/7 |

See [`docs/US_PRICING.md`](docs/US_PRICING.md) for full details.

## Architecture

```
                                    +-----------+
                                    |   Kong    |
                                    |  Gateway  |
                                    +-----+-----+
                                          |
          +-------------------------------+-------------------------------+
          |               |               |               |               |
    +-----v-----+   +-----v-----+   +-----v-----+   +-----v-----+   +-----v-----+
    |   Auth    |   |   User    |   | Merchant  |   |  Payment  |   | Analytics |
    |  Service  |   |  Service  |   |  Service  |   |  Gateway  |   |  Service  |
    +-----------+   +-----------+   +-----------+   +-----------+   +-----------+
          |               |               |               |               |
    +-----v-----+   +-----v-----+   +-----v-----+   +-----v-----+   +-----v-----+
    |   Fraud   |   |  Roundup  |   |Trans-     |   |   Wallet  |   |  Ledger   |
    |  Service  |   |  Engine   |   |action     |   |  Service  |   |  Service  |
    +-----------+   +-----------+   +-----------+   +-----------+   +-----------+
                                          |
    +-----------+   +-----------+   +-----v-----+
    |   Fee     |   |Invest-   |   |Notifi-    |
    |  Service  |   |ment Svc  |   |cation Svc |
    +-----------+   +-----------+   +-----------+
                                          |
                               +---------v---------+
                               |   Message Queue   |
                               |     (Kafka)       |
                               +-------------------+
                                          |
                               +---------v---------+
                               |  PostgreSQL        |
                               +-------------------+
```

**13 microservices** · **11 shared packages** · **20 Docker containers**

---

## Quick Start (Demo)

```bash
# 1. Start everything
docker compose up -d --build

# 2. Verify health
curl http://localhost/v1/health

# 3. Run the demo
./scripts/demo.sh
```

### Demo Walkthrough

| Step | What happens |
|------|-------------|
| 1️⃣ | User registers (`POST /v1/auth/register`) |
| 2️⃣ | User logs in (`POST /v1/auth/login`), gets JWT |
| 3️⃣ | MFA setup (`POST /v1/auth/mfa/setup`) |
| 4️⃣ | Merchant onboarded (`POST /v1/merchants`) |
| 5️⃣ | Payment created (`POST /v1/payments`) |
| 6️⃣ | Webhook registered (`POST /v1/webhooks`) |
| 7️⃣ | Analytics event tracked (`POST /v1/analytics/events`) |
| 8️⃣ | Profile fetched (`GET /v1/auth/me`) |
| 9️⃣ | Token refreshed (`POST /v1/auth/refresh`) |
| 🔟 | Logout (`POST /v1/auth/logout`) |

> **Load test:** 20 VU, p95 = 67ms, 0% error rate

---

## Services

| # | Service | Port | Description |
|---|---------|------|-------------|
| 1 | **auth-service** | 8081 | Auth, JWT, MFA, OAuth, KYC |
| 2 | **user-service** | 8088 | User profiles, preferences |
| 3 | **merchant-service** | 8092 | Merchant CRUD, fee tiers |
| 4 | **transaction-service** | 8083 | Transaction processing |
| 5 | **roundup-engine** | 8082 | Round-up calculation (Kafka consumer) |
| 6 | **wallet-service** | 8084 | Wallet management, balances |
| 7 | **fee-service** | 8085 | Fee calculation |
| 8 | **investment-service** | 8086 | Portfolio management |
| 9 | **ledger-service** | 8087 | Double-entry accounting |
| 10 | **payment-gateway** | 8090 | Stripe payment processing |
| 11 | **fraud-service** | 8094 | Fraud detection |
| 12 | **analytics-service** | 8093 | Event analytics |
| 13 | **notification-service** | 8091 | Webhooks, email, push |

---

## API Documentation

Interactive API docs: [Swagger UI](http://localhost:8080/swagger)

Base URL: `http://localhost/v1`

### Key Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/v1/health` | Health check |
| POST | `/v1/auth/register` | User registration |
| POST | `/v1/auth/login` | User login |
| POST | `/v1/auth/refresh` | Refresh token |
| POST | `/v1/auth/logout` | Logout |
| GET | `/v1/auth/me` | Current user profile |
| POST | `/v1/auth/mfa/setup` | Setup MFA |
| POST | `/v1/auth/mfa/verify` | Verify MFA code |
| POST | `/v1/auth/oauth/{provider}` | OAuth login (Google/GitHub mock) |
| POST | `/v1/auth/kyc` | Submit KYC |
| POST | `/v1/payments` | Create payment |
| GET | `/v1/payments/{id}` | Get payment status |
| POST | `/v1/payments/{id}/confirm` | Confirm payment |
| POST | `/v1/webhooks/stripe` | Stripe webhook |
| GET/POST | `/v1/merchants` | List/Create merchants |
| GET/PUT/DELETE | `/v1/merchants/{id}` | Manage merchant |
| POST | `/v1/analytics/events` | Track event |
| GET | `/v1/analytics/events` | Query analytics |
| POST | `/v1/webhooks` | Register webhook |
| GET/PUT/DELETE | `/v1/webhooks/{id}` | Manage webhook |

Full spec: [`docs/openapi.yaml`](docs/openapi.yaml)

---

## Monitoring

| Tool | URL | Credentials |
|------|-----|-------------|
| **Grafana** | http://localhost:3000 | admin / admin |
| **Prometheus** | http://localhost:9090 | — |
| **Kong Admin** | http://localhost:8001 | — |

Dashboards: service metrics, Kafka lag, DB pools, business KPIs.

---

## Deployment

| Method | Docs |
|--------|------|
| Docker Compose | `docker compose up -d` |
| Kubernetes (Helm) | `deploy/helm/` |
| Kubernetes (raw) | `deploy/k8s/` |
| Terraform (AWS) | `deploy/terraform/` |

See [`deploy/README.md`](deploy/README.md) for production checklist.

---

## Project Structure

```
├── services/          # 13 Go microservices
├── pkg/               # 11 shared Go packages
├── deploy/            # Docker, K8s, Helm, Terraform
│   ├── kong/          # Kong API Gateway config
│   ├── helm/          # Helm chart
│   ├── terraform/     # AWS infra (VPC, EKS, RDS, MSK, ECR, ALB)
│   ├── k8s/           # Raw Kubernetes manifests
│   ├── prometheus/    # Alert rules + scrape config
│   ├── grafana/       # Dashboards + datasources
│   ├── loki/          # Log aggregation config
│   ├── tempo/         # Tracing config
│   └── alertmanager/  # Alert routing + notifications
├── docs/              # API spec
└── scripts/           # Integration + load tests
```

---

## Load Test Results

```
✓ http_req_duration......: avg=12ms   p(95)=67ms
✓ http_req_failed........: 0.00%
✓ iterations.............: 1,234
```

---

## Tech Stack

**Backend:** Go 1.26, PostgreSQL 16, Kafka 3.6  
**Infrastructure:** Docker, Kubernetes (EKS), Terraform  
**API Gateway:** Kong  
**Observability:** Prometheus, Grafana, Loki, Tempo, Alertmanager  
**CI/CD:** GitHub Actions, Trivy, Helm, ghcr.io  
**Payments:** Stripe (mock mode for dev)  
**Load Testing:** k6

---

## License

MIT
