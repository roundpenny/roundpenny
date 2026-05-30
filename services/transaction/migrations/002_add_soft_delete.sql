-- Copyright (c) 2026 RoundPenny. All rights reserved.

ALTER TABLE transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE transaction_roundups ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_transactions_deleted ON transactions(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tx_roundups_deleted ON transaction_roundups(deleted_at) WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_user_active ON transactions(user_id) WHERE deleted_at IS NULL;
