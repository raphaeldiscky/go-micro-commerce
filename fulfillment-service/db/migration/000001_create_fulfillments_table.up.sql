CREATE TABLE fulfillments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    tracking_number VARCHAR(100) UNIQUE NOT NULL,
    carrier VARCHAR(100),
    shipping_label_url TEXT,
    estimated_delivery TIMESTAMPTZ NOT NULL,
    actual_delivery TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
);

CREATE INDEX idx_fulfillments_order_id ON fulfillments(order_id);
CREATE INDEX idx_fulfillments_status ON fulfillments(status);
CREATE INDEX idx_fulfillments_tracking_number ON fulfillments(tracking_number);