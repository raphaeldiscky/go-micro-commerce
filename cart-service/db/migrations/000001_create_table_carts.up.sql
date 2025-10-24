CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,

    CONSTRAINT chk_cart_status CHECK (status IN ('active', 'checked_out', 'archived'))
);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;


CREATE TRIGGER update_cart_updated_at
BEFORE UPDATE ON carts
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE IF NOT EXISTS cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL REFERENCES carts (id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity BIGINT NOT NULL CHECK (quantity > 0),
    selected_for_checkout BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp
);

CREATE TRIGGER update_cart_item_updated_at
BEFORE UPDATE ON cart_items
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE INDEX idx_carts_customer_status
ON carts (customer_id, status);
CREATE UNIQUE INDEX idx_unique_active_cart_per_customer
ON carts (customer_id)
WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_cart_created_at ON carts (created_at);
CREATE INDEX IF NOT EXISTS idx_cart_customer_id ON carts (customer_id);
CREATE INDEX IF NOT EXISTS idx_fk_cart_item_cart_id ON cart_items (cart_id);
CREATE INDEX IF NOT EXISTS idx_cart_items_cart_id_selected ON cart_items (cart_id, selected_for_checkout);
