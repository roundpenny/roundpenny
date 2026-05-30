-- Copyright (c) 2026 RoundPenny. All rights reserved.

ALTER TABLE portfolios ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE investments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_portfolios_deleted ON portfolios(deleted_at) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_investments_deleted ON investments(deleted_at) WHERE deleted_at IS NULL;
