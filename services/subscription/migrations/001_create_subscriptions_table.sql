-- Copyright (c) 2026 RoundPenny. All rights reserved.

CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    plan_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    current_period_start TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    current_period_end TIMESTAMPTZ NOT NULL,
    cancelled_at TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    stripe_subscription_id VARCHAR(255),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    amount NUMERIC(20,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    interval VARCHAR(20) NOT NULL DEFAULT 'month',
    features JSONB DEFAULT '[]',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS billing_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    subscription_id UUID REFERENCES subscriptions(id),
    amount NUMERIC(20,2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payment_method VARCHAR(50),
    stripe_invoice_id VARCHAR(255),
    period_start TIMESTAMPTZ NOT NULL,
    period_end TIMESTAMPTZ NOT NULL,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

INSERT INTO subscription_plans (name, description, amount, interval, features)
VALUES 
    ('Free', 'Basic round-up features', 0.00, 'month', '["5 transactions/day", "Basic analytics", "Email support"]'),
    ('Premium', 'Unlimited round-ups + investments', 2.99, 'month', '["Unlimited transactions", "Automated investing", "Priority support", "Advanced analytics"]'),
    ('Pro', 'Everything + merchant discounts', 9.99, 'month', '["All Premium features", "Merchant discounts up to 5%", "Tax reporting", "Dedicated support"]')
ON CONFLICT DO NOTHING;

CREATE INDEX IF NOT EXISTS idx_subs_user ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subs_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subs_stripe ON subscriptions(stripe_subscription_id);
CREATE INDEX IF NOT EXISTS idx_billing_user ON billing_history(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_sub ON billing_history(subscription_id);
