-- Copyright (c) 2026 RoundPenny. All rights reserved.

ALTER TABLE payments ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_payments_deleted ON payments(deleted_at) WHERE deleted_at IS NULL;
