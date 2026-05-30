-- Copyright (c) 2026 RoundPenny. All rights reserved.

CREATE TABLE IF NOT EXISTS ledger_accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(20) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    type VARCHAR(20) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS journal_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    transaction_id UUID,
    roundup_id UUID,
    description TEXT,
    entry_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS journal_lines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    journal_entry_id UUID NOT NULL REFERENCES journal_entries(id) ON DELETE CASCADE,
    account_id UUID NOT NULL REFERENCES ledger_accounts(id),
    debit_amount NUMERIC(20,8) DEFAULT 0,
    credit_amount NUMERIC(20,8) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    user_id UUID,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_journal_lines_entry ON journal_lines(journal_entry_id);
CREATE INDEX idx_journal_lines_account ON journal_lines(account_id);
CREATE INDEX idx_journal_lines_user ON journal_lines(user_id);

INSERT INTO ledger_accounts (code, name, type) VALUES
    ('1000', 'Cash - Platform', 'asset'),
    ('1100', 'User Wallets', 'liability'),
    ('1200', 'Investment Holdings', 'asset'),
    ('2000', 'Fee Revenue', 'revenue'),
    ('2100', 'Commission Payable', 'liability'),
    ('3000', 'Retained Earnings', 'equity')
ON CONFLICT (code) DO NOTHING;
