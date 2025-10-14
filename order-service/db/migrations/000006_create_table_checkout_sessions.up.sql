CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS checkout_sessions(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    customer_id UUID NOT NULL,
    selected_item_ids UUID[] NOT NULL,
    note TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'pending', -- pending, placed, canceled
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE checkout_sessions
ADD CONSTRAINT chk_checkout_session_status
CHECK (status IN ('pending', 'placed', 'canceled'));

ALTER TABLE checkout_sessions
ADD CONSTRAINT chk_checkout_session_currency
CHECK (currency ~ '^[A-Z]{3}$');

