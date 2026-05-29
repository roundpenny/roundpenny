CREATE TABLE IF NOT EXISTS wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID UNIQUE NOT NULL,
    balance NUMERIC(20,8) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    version INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_wallets_user ON wallets(user_id);

CREATE TABLE IF NOT EXISTS wallet_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    amount NUMERIC(20,8) NOT NULL,
    balance_before NUMERIC(20,8) NOT NULL,
    balance_after NUMERIC(20,8) NOT NULL,
    entry_type VARCHAR(20) NOT NULL,
    reference_type VARCHAR(50),
    reference_id UUID,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_wallet_entries_wallet ON wallet_entries(wallet_id);
CREATE INDEX idx_wallet_entries_user ON wallet_entries(user_id);
CREATE INDEX idx_wallet_entries_created ON wallet_entries(created_at DESC);

CREATE TABLE IF NOT EXISTS withdrawal_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    amount NUMERIC(20,8) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    destination_type VARCHAR(20) NOT NULL,
    destination_details JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'pending',
    external_ref VARCHAR(255),
    processed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_withdrawals_user ON withdrawal_requests(user_id);
CREATE INDEX idx_withdrawals_status ON withdrawal_requests(status);
