CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    merchant_id UUID,
    amount NUMERIC(20,8) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    type VARCHAR(20) DEFAULT 'purchase',
    external_tx_id VARCHAR(255),
    external_provider VARCHAR(50),
    description TEXT,
    metadata JSONB DEFAULT '{}',
    settled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_transactions_user ON transactions(user_id);
CREATE INDEX idx_transactions_merchant ON transactions(merchant_id);
CREATE INDEX idx_transactions_status ON transactions(status);
CREATE INDEX idx_transactions_external ON transactions(external_tx_id);
CREATE INDEX idx_transactions_created ON transactions(created_at DESC);

CREATE TABLE IF NOT EXISTS transaction_roundups (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    original_amount NUMERIC(20,8) NOT NULL,
    rounded_amount NUMERIC(20,8) NOT NULL,
    round_up_amount NUMERIC(20,8) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    wallet_entry_id UUID,
    fee_entry_id UUID,
    investment_entry_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tx_roundups_tx ON transaction_roundups(transaction_id);
CREATE INDEX idx_tx_roundups_user ON transaction_roundups(user_id);
CREATE INDEX idx_tx_roundups_status ON transaction_roundups(status);
