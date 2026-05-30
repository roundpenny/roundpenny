# Legal Operations — RoundPenny

**Last Updated: May 30, 2026**

This directory contains legal templates and operational documentation for the RoundPenny platform. All documents herein are **templates** and **must be reviewed by qualified legal counsel** before being made available to users or the public.

---

## Document Inventory

| File | Status | Description |
|------|--------|-------------|
| `docs/TERMS_OF_SERVICE.md` | Template | Terms of Service — governs user access and use of the platform |
| `docs/PRIVACY_POLICY.md` | Template | Privacy Policy — CCPA/GLBA/GDPR-aligned data practices disclosure |
| `docs/COOKIE_POLICY.md` | Template | Cookie Policy — describes essential and analytics cookie usage |
| `README.md` (this file) | Internal | Legal operations and compliance procedures |

---

## Jurisdiction

All legal documents are governed by the laws of the **State of Delaware, USA**. Disputes are resolved through binding individual arbitration in **Wilmington, Delaware**.

---

## Required Registrations (Pre-Launch)

Before making the platform available to users, the following regulatory registrations must be completed:

### SEC / State Securities

- [ ] **RIA Registration** — RoundPenny Advisors LLC: Form ADV filed via IARD
  - State-registered if AUM < $100M; SEC-registered if AUM > $100M or multi-state
  - Estimated timeline: 4–8 weeks
  - Estimated cost: $5k–$15k
- [ ] **RIA Compliance Manual** — Written policies and procedures (SEC Rule 206(4)-7)
- [ ] **CCO Designation** — Chief Compliance Officer appointed
- [ ] **Form ADV Parts 1, 2A, and 2B** — Filed and delivered to clients

### Broker-Dealer (Alternative Path)

- [ ] **BD Registration** — RoundPenny Securities LLC via CRD/FINRA (6–12 months, $50k–$100k)
  - *Recommended alternative:* Partner with Apex Clearing or DriveWealth as introducing broker

### State-Level

- [ ] **Blue Sky Filings** — 10–15 pilot states via Coordinated Equity Review (CER)
- [ ] **Money Transmitter License** — Verify applicability based on business model

### AML / BSA

- [ ] **AML Program** — Written BSA/AML policy adopted
- [ ] **FinCEN Registration** — If acting as broker-dealer or money services business
- [ ] **CIP Procedures** — Customer Identification Program implemented
- [ ] **OFAC Screening** — Sanctions screening integrated into KYC workflow
- [ ] **SAR Filing Process** — Suspicious Activity Report procedures documented

### Privacy

- [ ] **CCPA Compliance** — Privacy policy, data deletion API, consumer request process
- [ ] **GLBA Compliance** — Financial privacy notice, opt-out mechanisms
- [ ] **Regulation S-P** — Annual privacy notice delivery process
- [ ] **Cookie Consent Banner** — Implemented on website

### Cybersecurity

- [ ] **WISP** — Written Information Security Plan adopted
- [ ] **Incident Response Plan** — Documented and tested (see below)
- [ ] **Penetration Test** — Annual third-party penetration test completed
- [ ] **SOC 2 Type II** — Audit completed (target: 6–12 months post-launch)

---

## Data Protection Officer (DPO) Contact Process

The DPO is responsible for overseeing data protection compliance, handling data subject requests, and serving as the point of contact for regulatory authorities.

### Contact Information

- **Email:** dpo@roundpenny.com
- **Response SLA:** 48 hours for privacy-related inquiries
- **Escalation:** legal@roundpenny.com (if DPO unreachable)

### DPO Responsibilities

1. Monitor compliance with CCPA, GLBA, and GDPR requirements
2. Advise on data protection impact assessments (DPIAs)
3. Maintain records of processing activities
4. Serve as contact point for data subjects and regulators
5. Conduct annual privacy training for employees
6. Review vendor data processing agreements (DPAs)

### Data Subject Request Process

1. Request received at privacy@roundpenny.com
2. Acknowledge receipt within 10 business days (CCPA) or 5 business days (GDPR)
3. Verify identity of requester
4. Respond substantively within 45 days (CCPA) or 30 days (GDPR)
5. If extension needed, notify requester with reason

---

## Incident Response Process — Data Breach

### 72-Hour Notification Policy

In accordance with applicable law (including state breach notification laws and GDPR Article 33), RoundPenny commits to notifying affected users and relevant authorities **within 72 hours** of confirmed discovery of a data breach involving personal information.

### Incident Response Team

| Role | Responsibility | Contact |
|------|---------------|---------|
| Incident Commander | Leads response, coordinates team | security@roundpenny.com |
| Legal Counsel | Assesses legal obligations, interfaces with regulators | legal@roundpenny.com |
| DPO | Assesses data protection impact | dpo@roundpenny.com |
| Communications Lead | Drafts user notifications, PR statements | comms@roundpenny.com |
| Engineering Lead | Contains/eradicates threat, preserves forensics | infra@roundpenny.com |

### Response Steps

#### Phase 1: Detection & Assessment (0–4 hours)
- [ ] Confirm and classify the incident (type, scope, systems affected)
- [ ] Assemble incident response team
- [ ] Preserve evidence (logs, disk images, network captures)
- [ ] Determine if personal information is involved
- [ ] Classify severity (Low / Medium / High / Critical)

#### Phase 2: Containment (1–8 hours)
- [ ] Isolate affected systems (network segmentation, service takedown)
- [ ] Revoke compromised credentials and access tokens
- [ ] Engage AWS support for account-level containment if needed
- [ ] Implement emergency patches or configuration changes

#### Phase 3: Investigation (4–48 hours)
- [ ] Determine root cause and attack vector
- [ ] Identify affected data categories and data subjects
- [ ] Determine if data was exfiltrated, altered, or destroyed
- [ ] Engage external forensics firm if critical severity
- [ ] Document chain of custody for evidence

#### Phase 4: Notification (24–72 hours)
- [ ] **Notify affected users** — email describing incident, data involved, steps taken, and user remediation steps
- [ ] **Notify regulatory authorities** — state attorneys general, SEC/FINRA (if applicable), and state-specific breach notification agencies
- [ ] **Notify law enforcement** — FBI/Secretary of State if financial fraud or identity theft suspected
- [ ] **Website posting** — public notice if email unavailable
- [ ] **File SAR with FinCEN** if suspicious activity related to money laundering

#### Phase 5: Remediation (post-72 hours)
- [ ] Implement permanent fixes and security improvements
- [ ] Conduct post-mortem and document lessons learned
- [ ] Update incident response plan
- [ ] Provide affected users with credit monitoring if sensitive data exposed (SSN, financial accounts)
- [ ] Cooperate with regulatory investigations

### Notification Template — User

```
Subject: Security Notice — RoundPenny Account [Ref: INC-XXXXX]

Dear [User Name],

We are writing to notify you of a security incident involving your
RoundPenny account.

What happened: [Brief description of the incident]
What information was involved: [Categories of data affected]
What we are doing: [Actions taken by RoundPenny]
What you can do: [Steps user should take, e.g., change password, monitor accounts]
Contact: security@roundpenny.com | Incident Ref: INC-XXXXX

We apologize for any concern this may cause. We take the security of
your information seriously and are taking all appropriate steps to
address this incident.

Sincerely,
RoundPenny Security Team
```

### Regulatory Notification Requirements (US)

| Jurisdiction | Authority | Notification Trigger | Timeline |
|-------------|-----------|---------------------|----------|
| All states | State Attorney General | Breach of SSN, DL, or financial account info | Varies (30–60 days typical) |
| California | CA AG (oag.ca.gov) | Breach of 500+ CA residents | Most expedient time, no later than 30 days |
| New York | NY AG + DFS | Breach of private information | Most expedient time, no later than 30 days |
| SEC | SEC EDGAR | Material breach affecting public company | 4 business days |
| FinCEN | FinCEN BSA E-Filing | Suspicious activity related to breach | 30 calendar days |

---

## Third-Party Data Processors

| Processor | Service | Data Processing Agreement (DPA) | SOC 2 |
|-----------|---------|-------------------------------|-------|
| Amazon Web Services | Cloud infrastructure | [ ] Signed | [ ] Reviewed |
| Stripe, Inc. | Payment processing | [ ] Signed | [ ] Reviewed |
| Onfido | Identity verification | [ ] Signed | [ ] Reviewed |
| Twilio (SendGrid) | Email delivery | [ ] Signed | [ ] Reviewed |
| Plaid / Finicity | Bank account linking | [ ] Signed | [ ] Reviewed |
| Apex Clearing / DriveWealth | Custody / execution | [ ] Signed | [ ] Reviewed |

**Action Required:** Obtain and review SOC 2 reports and signed DPAs from all data processors before processing live user data.

---

## Links

- [Terms of Service](../../docs/TERMS_OF_SERVICE.md)
- [Privacy Policy](../../docs/PRIVACY_POLICY.md)
- [Cookie Policy](../../docs/COOKIE_POLICY.md)
- [Regulatory Plan](../../docs/REGULATORY_PLAN.md)
- [Pricing Model](../../docs/US_PRICING.md)

---

*This document is for internal operational use only. It contains procedures and checklists intended for RoundPenny personnel and legal advisors. Do not distribute externally without legal review.*
