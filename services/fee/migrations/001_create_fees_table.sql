CREATE TABLE IF NOT EXISTS fee_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    fee_type VARCHAR(20) NOT NULL,
    percentage NUMERIC(5,2),
    flat_amount NUMERIC(20,8),
    min_amount NUMERIC(20,8),
    max_amount NUMERIC(20,8),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS fee_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID,
    roundup_id UUID,
    user_id UUID NOT NULL,
    amount NUMERIC(20,8) NOT NULL,
    fee_type VARCHAR(20) NOT NULL,
    percentage_applied NUMERIC(5,2),
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_fee_tx_transaction ON fee_transactions(transaction_id);
CREATE INDEX idx_fee_tx_user ON fee_transactions(user_id);
CREATE INDEX idx_fee_tx_created ON fee_transactions(created_at DESC);

INSERT INTO fee_configs (name, fee_type, percentage, flat_amount, min_amount, max_amount)
VALUES 
    ('Round-Up Default', 'roundup', 10.00, NULL, 0.01, NULL),
    ('Premium Monthly', 'subscription', NULL, 2.99, NULL, NULL),
    ('Merchant Standard', 'merchant', 0.50, NULL, NULL, NULL),
    ('Investment Mgmt', 'investment', 0.25, NULL, NULL, NULL)
ON CONFLICT DO NOTHING;
