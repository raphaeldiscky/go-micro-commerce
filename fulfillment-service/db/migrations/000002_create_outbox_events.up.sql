BEGIN;

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- 'order', 'product', 'payment', etc. base on table name
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL,
    event_type TEXT NOT NULL, -- 'PaymentCreated', 'PaymentUpdated', etc.
    topic TEXT NOT NULL, -- 'payment.lifecycle', etc.
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now(),
    processed_at TIMESTAMPTZ,
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts INTEGER DEFAULT 0,
    last_error TEXT
);


CREATE INDEX idx_outbox_status_scheduled ON outbox_events (
    status, scheduled_for
);
CREATE INDEX idx_outbox_aggregate_type_id ON outbox_events (
    aggregate_type, aggregate_id
);
CREATE INDEX idx_outbox_created_at ON outbox_events (created_at);


ALTER TABLE outbox_events
ADD CONSTRAINT chk_outbox_status
CHECK (status IN ('pending', 'processing', 'processed', 'failed', 'retry'));

ALTER TABLE outbox_events
ADD CONSTRAINT chk_outbox_attempts
CHECK (attempts >= 0);

ALTER TABLE outbox_events
ADD CONSTRAINT chk_outbox_scheduled_for
CHECK (scheduled_for >= created_at);


COMMENT ON TABLE outbox_events IS 'Stores events to be published to message brokers using the outbox pattern';
COMMENT ON COLUMN outbox_events.id IS 'Unique identifier for the outbox event';
COMMENT ON COLUMN outbox_events.aggregate_type IS 'Type of aggregate that generated the event (order, product, payment, etc.)';
COMMENT ON COLUMN outbox_events.aggregate_id IS 'ID of the aggregate that generated the event';
COMMENT ON COLUMN outbox_events.event_type IS 'Type of event (PaymentCreated, PaymentUpdated, etc.)';
COMMENT ON COLUMN outbox_events.topic IS 'Kafka topic (payment.lifecycle, etc.) where the event should be published';
COMMENT ON COLUMN outbox_events.payload IS 'Complete event payload in JSON format';
COMMENT ON COLUMN outbox_events.status IS 'Current processing status of the event';
COMMENT ON COLUMN outbox_events.created_at IS 'When the event was first created';
COMMENT ON COLUMN outbox_events.processed_at IS 'When the event was successfully processed';
COMMENT ON COLUMN outbox_events.scheduled_for IS 'When the event should be processed (for delayed events)';
COMMENT ON COLUMN outbox_events.attempts IS 'Number of processing attempts made';
COMMENT ON COLUMN outbox_events.last_error IS 'Last error message if processing failed';

COMMIT;
