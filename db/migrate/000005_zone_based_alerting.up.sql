CREATE TABLE IF NOT EXISTS zones (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    camera VARCHAR(255) NOT NULL,
    polygon JSONB NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    active_from TIME,
    active_until TIME,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);