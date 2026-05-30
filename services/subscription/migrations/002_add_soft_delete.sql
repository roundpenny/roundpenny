-- Copyright (c) 2026 RoundPenny. All rights reserved.

ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
ALTER TABLE billing_history ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;
