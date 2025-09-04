CREATE TABLE fulfillments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    tracking_number TEXT UNIQUE NOT NULL,
    carrier TEXT,
    shipping_label_url TEXT,
    shipping_cost DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    weight_kg DECIMAL(8, 3) NOT NULL DEFAULT 0,
    dimensions JSONB, -- {width: 0, height: 0, length: 0, unit: "cm"}
    estimated_delivery_at TIMESTAMPTZ NOT NULL,
    actual_delivery_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE fulfillments
ADD CONSTRAINT chk_fulfillments_status
CHECK (status IN (
    'pending',
    'processing',
    'shipped',
    'in_transit',
    'delivered',
    'cancelled',
    'returned'
)),
ADD CONSTRAINT chk_fulfillments_cost
CHECK (shipping_cost >= 0),
ADD CONSTRAINT chk_fulfillments_currency
CHECK (currency ~ '^[A-Z]{3}$');

CREATE INDEX idx_fulfillments_order_id ON fulfillments(order_id);
CREATE INDEX idx_fulfillments_status ON fulfillments(status);
CREATE INDEX idx_fulfillments_tracking_number ON fulfillments(tracking_number);