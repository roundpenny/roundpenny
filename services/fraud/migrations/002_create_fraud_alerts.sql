-- Copyright (c) 2026 RoundPenny. All rights reserved.

CREATE TABLE IF NOT EXISTS fraud_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    rule_id UUID NOT NULL REFERENCES fraud_rules(id),
    transaction_id UUID NOT NULL,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    status VARCHAR(30) NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'investigated', 'resolved', 'false_positive')),
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ
);

CREATE INDEX idx_fraud_alerts_user ON fraud_alerts (user_id);
CREATE INDEX idx_fraud_alerts_rule ON fraud_alerts (rule_id);
CREATE INDEX idx_fraud_alerts_status ON fraud_alerts (status);
CREATE INDEX idx_fraud_alerts_severity ON fraud_alerts (severity);
