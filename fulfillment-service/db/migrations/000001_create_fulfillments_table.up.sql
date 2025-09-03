CREATE TYPE fulfillment_status AS ENUM (
    'pending',
    'processing',
    'shipped',
    'in_transit',
    'delivered',
    'cancelled',
    'returned'
);

CREATE TABLE fulfillments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    status fulfillment_status NOT NULL DEFAULT 'pending',
    tracking_number TEXT UNIQUE NOT NULL,
    carrier TEXT,
    shipping_label_url TEXT,
    shipping_cost DECIMAL(10, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'IDR',
    weight_kg DECIMAL(8, 3) NOT NULL DEFAULT 0,
    dimensions JSONB, -- {width: 0, height: 0, length: 0, unit: "cm"}
    estimated_delivery_at TIMESTAMPTZ NOT NULL,
    actual_delivery_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_fulfillments_order_id ON fulfillments(order_id);
CREATE INDEX idx_fulfillments_status ON fulfillments(status);
CREATE INDEX idx_fulfillments_tracking_number ON fulfillments(tracking_number);