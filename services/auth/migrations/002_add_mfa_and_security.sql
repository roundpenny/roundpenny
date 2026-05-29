CREATE EXTENSION IF NOT EXISTS "pgcrypto";

ALTER TABLE users ADD COLUMN IF NOT EXISTS mfa_backup_codes TEXT[] DEFAULT '{}';
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMPTZ;

CREATE TABLE IF NOT EXISTS mfa_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_mfa_sessions_user ON mfa_sessions(user_id);
CREATE INDEX idx_mfa_sessions_token ON mfa_sessions(token_hash);
