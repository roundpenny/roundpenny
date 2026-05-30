# Regulatory Plan — RoundPenny

## 1. Business Model Classification

RoundPenny operates two distinct regulated activities:
- **Broker-dealer services** (executing round-up investments into securities)
- **Asset management / investment advice** (automatic portfolio allocation)

## 2. US Regulatory Path

### 2.1 SEC Registration — RIA (Recommended First Step)

| Item | Detail |
|------|--------|
| Entity | RoundPenny Advisors LLC |
| Form | ADV (via IARD system) |
| Threshold | <$100M AUM → state-registered (SEC if >$100M or multi-state) |
| Timeline | 4–6 weeks for state approval |
| Cost | ~$5k–$15k (filing + legal) |

**Requirements:**
- CCO designation (Chief Compliance Officer)
- Written compliance manual (SEC Rule 206(4)-7)
- Annual audit (if >$100M AUM) or surprise exam (if discretionary)
- Form ADV Parts 1 (filing) + 2A (brochure) + 2B (brochure supplement)
- Recordkeeping (Rule 204-2): 5-year retention
- Custody rule (if holding funds/securities directly)

### 2.2 Broker-Dealer Registration

| Item | Detail |
|------|--------|
| Entity | RoundPenny Securities LLC |
| Form | BD (via CRD system) |
| SRO | FINRA membership required |
| Timeline | 6–12 months |
| Cost | ~$50k–$100k (filing + legal + FINRA fees) |

**Consider alternatives:**
- Partner with an existing BD (e.g., Apex Clearing, DriveWealth) as an introducing broker
- Use a turnkey RIA platform (e.g., Alto, iCapital) to avoid direct BD registration

### 2.3 State-Level (Blue Sky) Compliance

Register offerings in each state where customers reside, or rely on:
- Rule 506(b) — SEC preempts state registration for accredited investors
- Limit launch to 10–15 pilot states via Coordinated Equity Review (CER)

## 3. Operational Requirements

### 3.1 AML Program (USA PATRIOT Act)

- BSA/AML policy
- FinCEN registration (if acting as a broker-dealer)
- SAR filing process
- CIP (Customer Identification Program)
- OFAC screening

### 3.2 KYC/CIP Integration

Already built: Onfido for identity verification. Enhance to include:
- SSN/TIN collection (W-9 form)
- Beneficial ownership (if entity accounts)
- Politically Exposed Person (PEP) screening: add World-Check or Onfido Watchlist

### 3.3 Data Privacy

- **CCPA** (California): privacy policy, opt-out mechanism, data deletion API
- **GLBA**: financial privacy notice, opt-out for non-affiliate sharing
- **Regulation S-P**: annual privacy notice (opt-out for opt-in states)

### 3.4 Cybersecurity

- Written Information Security Plan (WISP)
- Annual penetration test + vulnerability scan
- SOC 2 Type II (target: 6–12 months post-launch)
- Incident response plan (aligned with FTC Safeguards Rule)

## 4. Timeline & Budget

| Phase | Tasks | Est. Cost | Timeline |
|-------|-------|-----------|----------|
| 1. RIA Setup | ADV filing, compliance manual, CCO | $10k–$20k | 4–8 weeks |
| 2. State Filings | 10 pilot states via CER | $5k–$10k | 4–6 weeks |
| 3. AML/CIP | Policy, FinCEN, OFAC | $5k–$10k | 2–4 weeks |
| 4. Privacy | CCPA/GLBA policies, disclosures | $3k–$5k | 2 weeks |
| 5. Security | Pen test, WISP, SOC 2 prep | $15k–$30k | 8–12 weeks |
| **Total** | | **$38k–$75k** | **20–32 weeks** |

## 5. Recommended Path (MVP)

1. **Register as RIA** (state-level) — enables discretionary investment management
2. **Use Apex Clearing or DriveWealth** for custody + execution (avoids BD registration)
3. **Launch with 10 pilot states** — limit regulatory exposure
4. **Achieve SOC 2 Type II** within 12 months — required for enterprise partners
5. **Re-evaluate BD registration** at $1B+ AUM or when offering direct securities origination

## 6. Key Legal Partners

| Specialty | Recommended Firms |
|-----------|-------------------|
| Securities law | Seward & Kissel, Katten Muchin Rosenman |
| Fintech regulatory | Goodwin Procter, Cooley |
| RIA compliance | RIA in a Box, Foreside Compliance |
| Cyber/SOC 2 | A-LIGN, KirkpatrickPrice |
