-- Copyright (c) 2026 RoundPenny. All rights reserved.

ALTER TABLE fee_configs ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE fee_transactions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_fee_configs_deleted ON fee_configs(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_fee_tx_deleted ON fee_transactions(deleted_at) WHERE deleted_at IS NULL;
