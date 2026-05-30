-- Copyright (c) 2026 RoundPenny. All rights reserved.

ALTER TABLE wallets ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE wallet_entries ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE withdrawal_requests ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_wallets_deleted ON wallets(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_wallet_entries_deleted ON wallet_entries(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_withdrawals_deleted ON withdrawal_requests(deleted_at) WHERE deleted_at IS NULL;
