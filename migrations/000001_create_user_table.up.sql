CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    login TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    current_balance NUMERIC(15, 2) NOT NULL DEFAULT 0 CHECK (current_balance >= 0),
    withdrawn_total NUMERIC(15, 2) NOT NULL DEFAULT 0 CHECK (withdrawn_total >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
