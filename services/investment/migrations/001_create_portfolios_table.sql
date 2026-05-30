-- Copyright (c) 2026 RoundPenny. All rights reserved.

CREATE TABLE IF NOT EXISTS portfolios (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    strategy VARCHAR(20) NOT NULL DEFAULT 'moderate',
    balance NUMERIC(20,8) DEFAULT 0,
    external_account_id VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_portfolios_user ON portfolios(user_id);

CREATE TABLE IF NOT EXISTS investments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    portfolio_id UUID NOT NULL REFERENCES portfolios(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    amount NUMERIC(20,8) NOT NULL,
    source VARCHAR(50) NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    external_ref VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_investments_portfolio ON investments(portfolio_id);
CREATE INDEX idx_investments_user ON investments(user_id);
CREATE INDEX idx_investments_created ON investments(created_at DESC);
