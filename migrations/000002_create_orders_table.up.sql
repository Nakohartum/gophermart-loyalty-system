CREATE TABLE IF NOT EXISTS orders (
    number TEXT PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('NEW', 'PROCESSING', 'INVALID', 'PROCESSED')),
    accrual NUMERIC(15, 2),
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_orders_user_id_uploaded_at
    ON orders (user_id, uploaded_at DESC);

CREATE INDEX IF NOT EXISTS idx_orders_status
    ON orders (status);
