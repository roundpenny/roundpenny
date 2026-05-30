CREATE TABLE IF NOT EXISTS admin_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL DEFAULT 'Admin',
    role VARCHAR(20) NOT NULL DEFAULT 'admin',
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO admin_users (email, password_hash, full_name) VALUES
    ('admin@roundpenny.com', '$2a$10$dummyhash', 'Admin')
ON CONFLICT (email) DO NOTHING;
