-- Copyright (c) 2026 RoundPenny. All rights reserved.

CREATE TABLE IF NOT EXISTS user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    avatar_url VARCHAR(500),
    date_of_birth DATE,
    address_line1 VARCHAR(255),
    address_line2 VARCHAR(255),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(3),
    tax_id VARCHAR(50),
    occupation VARCHAR(100),
    income_range VARCHAR(20),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kyc_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    document_type VARCHAR(20) NOT NULL,
    document_number VARCHAR(100),
    issued_country VARCHAR(3),
    status VARCHAR(20) DEFAULT 'pending',
    front_image_url VARCHAR(500),
    back_image_url VARCHAR(500),
    verified_at TIMESTAMPTZ,
    verified_by UUID,
    rejection_reason TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_kyc_docs_user ON kyc_documents(user_id);
CREATE INDEX idx_kyc_docs_status ON kyc_documents(status);

CREATE TABLE IF NOT EXISTS user_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    round_to_nearest NUMERIC(10,2) DEFAULT 1.00,
    max_daily_roundup NUMERIC(10,2) DEFAULT 5.00,
    multiplier INTEGER DEFAULT 1,
    auto_invest BOOLEAN DEFAULT TRUE,
    investment_strategy VARCHAR(20) DEFAULT 'moderate',
    notifications_email BOOLEAN DEFAULT TRUE,
    notifications_push BOOLEAN DEFAULT TRUE,
    notifications_sms BOOLEAN DEFAULT FALSE,
    language VARCHAR(5) DEFAULT 'en',
    timezone VARCHAR(50) DEFAULT 'UTC',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
