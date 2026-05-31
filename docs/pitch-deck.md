---
marp: true
theme: uncover
class: invert
paginate: true
---

<!-- Copyright (c) 2026 RoundPenny. All rights reserved. -->

# **RoundPenny**

## White-Label Round-Up Micro-Investment Platform

**Investor Pitch Deck · 2026**

---

## **The Problem**

### $1.2 trillion

Global micro-investment market opportunity (TAM)

### 1.7 billion

Unbanked & underbanked adults worldwide

### 3 in 4

Millennials say they'd invest more if the process were automatic

**Barriers to entry:**
- Complex regulatory requirements (FINRA, SEC, MiFID II)
- High infrastructure costs (brokerage, custody, compliance)
- Limited white-label solutions for fintechs & neobanks

---

## **The Solution**

### RoundPenny — Micro-Investment-as-a-Service

A complete, white-label round-up investment platform that any fintech, neobank, or HCM platform can launch in **weeks, not years**.

```
Purchase → Round-Up → Invest → Grow
  $4.50      $5.00     $0.50    Portfolio
```

**Key value proposition:** Turn spare change into investments automatically, under your own brand.

---

## **Market Size**

| Segment | Value | Source |
|---------|-------|--------|
| **TAM** — Global micro-investment | $1.2T | Allied Market Research |
| **SAM** — US + EU digital investment | $84B | Statista |
| **SOM** — Round-up segment (targetable) | $4.2B | Internal analysis |
| **CAGR** (2025–2030) | 22% | Grand View Research |

**Target markets:** Neobanks (500+ globally), HCM/Payroll platforms (200+), Traditional banks (100+)

---

## **Product Architecture**

```
┌─────────────┐  ┌─────────────┐  ┌─────────────┐
│  API Gateway │  │ Auth/User   │  │ Merchant    │
│  (Kong)      │  │ Service     │  │ Service     │
├─────────────┤  ├─────────────┤  ├─────────────┤
│ Transaction │  │ Round-Up    │  │ Wallet      │
│ Service     │──│ Engine      │──│ Service     │
├─────────────┤  ├─────────────┤  ├─────────────┤
│ Payment     │  │ Investment  │  │ Ledger      │
│ Gateway     │  │ Service     │  │ Service     │
├─────────────┤  ├─────────────┤  ├─────────────┤
│ Fee Service │  │ Fraud       │  │ Analytics   │
│             │  │ Detection   │  │ Service     │
├─────────────┤  ├─────────────┤  ├─────────────┤
│ Notification│  │ Admin       │  │ Subscription│
│ Service     │  │ Dashboard   │  │ Billing     │
└─────────────┘  └─────────────┘  └─────────────┘
```

**15 decoupled microservices · Kafka event-driven · PostgreSQL · Redis · Kubernetes**

---

## **Business Model**

| Revenue Stream | Model | Projected (Year 3) |
|----------------|-------|-------------------|
| **Subscription** | Free / $2.99 / $9.99 per user/mo | $3.6M ARR |
| **Merchant Commission** | 0.25% per transaction | $4.2M ARR |
| **White-Label Licensing** | $50k–$500k per partner/year | $3.0M ARR |
| **Spread Income** | Interest on uninvested cash | $1.2M ARR |
| **Total Projected ARR** | | **$12.0M** |

**Unit economics:** CAC ~$15, LTV ~$180, Payback period ~3 months

---

## **Traction & Roadmap**

```
Q1 2026  ● Platform launch, 15 microservices live
Q2 2026  ● Partner onboarding, SOC 2 audit (in progress)
Q3 2026  ● Target: 50k users, $500k GMV
Q4 2026  ● Target: 200k users, $5M GMV
Year 2   ● Target: 1M users, $50M GMV, FINRA licensed
Year 3   ● Target: 5M users, $250M GMV, profitable
```

---

## **Competitive Advantage**

| Feature | RoundPenny | Acorns | Stash | Betterment |
|---------|:----------:|:------:|:-----:|:----------:|
| White-Label API | ✓ | ✗ | ✗ | ✗ |
| Merchant Commission | ✓ | ✗ | ✗ | ✗ |
| Real-Time Kafka Engine | ✓ | ✗ | ✗ | ✗ |
| Multi-Tenant | ✓ | ✗ | ✗ | ✗ |
| Self-Hostable | ✓ | ✗ | ✗ | ✗ |
| Full Observability | ✓ | Limited | Limited | ✓ |
| No Vendor Lock-in | ✓ | ✗ | ✗ | ✗ |

**Moat:** Infrastructure that competitors cannot replicate without years of development.

---

## **Why Now?**

### Market timing is perfect

1. **Neobank explosion** — 500+ digital banks globally need investment products
2. **Regulatory tailwind** — FINRA/SEC modernizing rules for digital investing
3. **Behavioral shift** — Gen Z & Millennials prefer automatic, app-based investing
4. **Acorns IPO** — Validated the round-up model at $2.2B valuation
5. **Embedded finance** — Every platform wants to offer financial products

---

## **Technology Moat**

- **15 microservices** in Go — independently scalable, deployable
- **Apache Kafka** — real-time event streaming, exactly-once semantics
- **PostgreSQL + Redis** — reliable persistence + sub-millisecond caching
- **Kong API Gateway** — enterprise-grade rate limiting, auth, routing
- **Prometheus + Grafana + Loki + Tempo** — full observability stack
- **AES-256 + TLS 1.3** — bank-grade encryption
- **Docker + Kubernetes** — multi-cloud, zero-downtime deployments

---

## **Security & Compliance**

| Certification | Status | Timeline |
|--------------|--------|----------|
| SOC 2 Type II | In progress | Q3 2026 |
| GDPR | Compliant | ✓ |
| CCPA | Compliant | ✓ |
| AES-256 Encryption | Implemented | ✓ |
| TLS 1.3 | Implemented | ✓ |
| FINRA / RIA | Planned | 2027 |
| UK EMI License | Planned | 2027 |

---

## **Target Customers**

### Who will buy RoundPenny?

1. **Neobanks & Challenger Banks** (N26, Revolut, Chime, Monese)
   → Add investment products to existing accounts

2. **HCM & Payroll Platforms** (Gusto, Paychex, ADP, Workday)
   → Offer micro-investing as employee benefit

3. **Traditional Banks** (Banks with 1M+ customers)
   → Modernize their investment offerings

4. **Super Apps** (Gojek, Grab, Paytm)
   → Add micro-investing to their ecosystem

**Pilot partners:** 3 LOIs in negotiation (neobank, HCM platform, regional bank)

---

## **Go-To-Market Strategy**

### Phase 1: Direct Sales (Q2-Q3 2026)
- Target 10 pilot partners in US + EU
- Dedicated solutions engineering team
- 60-day implementation guarantee

### Phase 2: Self-Service (Q4 2026)
- Developer portal with API docs & SDKs
- Automated onboarding & sandbox
- Tiered pricing based on volume

### Phase 3: Platform (2027)
- Partner marketplace & integrations
- Referral program for fintech consultants
- Regional data centers for compliance

---

## **Financial Projections**

| Metric | Year 1 | Year 2 | Year 3 |
|--------|--------|--------|--------|
| Active Users | 200k | 1M | 5M |
| GMV (Annual) | $5M | $50M | $250M |
| Revenue | $500k | $4.2M | $12M |
| Gross Margin | 65% | 78% | 85% |
| CAC | $15 | $12 | $8 |
| LTV | $60 | $120 | $180 |
| Team Size | 8 | 18 | 35 |

**Path to profitability:** Month 18 (breakeven)

---

## **The Team**

### [Name, placeholder]
**Founder & CEO**
- Background in fintech & distributed systems
- Full-stack development (Go, Kafka, K8s)
- Product vision & execution

### Key Hires (Next 12 Months)
- CTO / Lead Engineer
- Head of Partnerships & Sales
- Compliance Officer
- Head of Product

---

## **Investment Ask**

| Item | Detail |
|------|--------|
| **Round** | Seed |
| **Amount** | $750k – $1.5M |
| **Use of Funds** | Engineering (40%), Sales & Marketing (30%), Compliance (20%), Operations (10%) |
| **Milestones** | SOC 2 completion, 10 pilot partners, 50k users, $500k GMV |

**Previous investment:** Bootstrapped (self-funded MVP)

---

## **Use of Funds**

```
Engineering (40%)
├── Platform scaling & performance optimization
├── Mobile SDK development (iOS/Android)
└── AI-powered fraud detection v2

Sales & Marketing (30%)
├── Partner onboarding & solutions engineering
├── Developer portal & documentation
└── Industry event presence & content marketing

Compliance (20%)
├── SOC 2 Type II certification costs
├── FINRA/RIA application preparation
└── Legal & regulatory consulting

Operations (10%)
├── Cloud infrastructure scaling
├── Customer success & support
└── Working capital
```

---

## **Risks & Mitigation**

| Risk | Mitigation |
|------|-----------|
| Regulatory delay | Started SOC 2 early; legal counsel retained |
| Partner acquisition | 3 LOIs in pipeline; flexible white-label pricing |
| Technology scaling | Kafka + K8s proven at 10x projected scale |
| Market competition | Technical moat (15 microservices, real-time engine) |
| Team capacity | Key hires planned; contractor network available |

---

## **Why RoundPenny Will Win**

1. **First-mover advantage** in white-label micro-investment infrastructure
2. **Technical moat** that takes years to replicate
3. **Multiple revenue streams** (subscription + commission + licensing + spread)
4. **Massive addressable market** growing at 22% CAGR
5. **Perfect market timing** — neobanks need investment products now
6. **Capital efficient** — built MVP by a small team

---

## **Thank You**

# **RoundPenny**

**Contact:** roundpennny@outlook.com

**GitHub:** [github.com/tmuhammet36-cmd/roundpenny](https://github.com/tmuhammet36-cmd/roundpenny) *(private repo available on request)*

**Site:** [tmuhammet36-cmd.github.io/roundpenny](https://tmuhammet36-cmd.github.io/roundpenny/)

---

*This presentation is confidential and intended solely for the use of the individual or entity to whom it is addressed. © 2026 RoundPenny. All rights reserved.*
