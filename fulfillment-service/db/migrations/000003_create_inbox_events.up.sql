BEGIN;

-- Create inbox_events table for consuming events from other services
CREATE TABLE inbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    -- Unique identifier from Kafka message metadata
    message_id UUID NOT NULL UNIQUE,
    -- 'order', 'product', 'user', etc. from source service
    aggregate_type TEXT NOT NULL,
    aggregate_id UUID NOT NULL, -- ID of the aggregate from source service
    -- 'OrderCreated', 'OrderUpdated', 'ProductUpdated', etc.
    event_type TEXT NOT NULL,
    -- 'order.lifecycle', 'product.lifecycle', 'user.verification', etc.
    topic TEXT NOT NULL,
    -- 'order-service', 'product-service', 'user-service', etc.
    source_service TEXT NOT NULL,
    payload JSONB NOT NULL, -- Complete event payload from the source service
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ DEFAULT now(),
    processed_at TIMESTAMPTZ,
    scheduled_for TIMESTAMPTZ NOT NULL DEFAULT now(),
    attempts INTEGER DEFAULT 0,
    last_error TEXT,
    correlation_id UUID, -- For tracing requests across services
    causation_id UUID -- For linking cause-and-effect events
);


CREATE UNIQUE INDEX idx_inbox_message_id ON inbox_events (message_id);
CREATE INDEX idx_inbox_status_scheduled ON inbox_events (status, scheduled_for);
CREATE INDEX idx_inbox_aggregate_type_id ON inbox_events (
    aggregate_type, aggregate_id
);
CREATE INDEX idx_inbox_event_type ON inbox_events (event_type);
CREATE INDEX idx_inbox_source_service ON inbox_events (source_service);
CREATE INDEX idx_inbox_created_at ON inbox_events (created_at);
CREATE INDEX idx_inbox_correlation_id ON inbox_events (
    correlation_id
) WHERE correlation_id IS NOT NULL;


ALTER TABLE inbox_events
ADD CONSTRAINT chk_inbox_status
CHECK (
    status IN (
        'pending', 'processing', 'processed', 'failed', 'retry', 'duplicate'
    )
);

ALTER TABLE inbox_events
ADD CONSTRAINT chk_inbox_attempts
CHECK (attempts >= 0);

ALTER TABLE inbox_events
ADD CONSTRAINT chk_inbox_scheduled_for
CHECK (scheduled_for >= created_at);


COMMENT ON TABLE inbox_events IS 'Stores events consumed from message brokers using the inbox pattern for idempotent processing';
COMMENT ON COLUMN inbox_events.id IS 'Unique identifier for the inbox event record';
COMMENT ON COLUMN inbox_events.message_id IS 'Unique identifier from Kafka message metadata for deduplication';
COMMENT ON COLUMN inbox_events.aggregate_type IS 'Type of aggregate from the source service that generated the event';
COMMENT ON COLUMN inbox_events.aggregate_id IS 'ID of the aggregate from the source service';
COMMENT ON COLUMN inbox_events.event_type IS 'Type of event received from source service';
COMMENT ON COLUMN inbox_events.topic IS 'Kafka topic from which the event was consumed';
COMMENT ON COLUMN inbox_events.source_service IS 'Name of the microservice that published the event';
COMMENT ON COLUMN inbox_events.payload IS 'Complete event payload received from the source service';
COMMENT ON COLUMN inbox_events.status IS 'Current processing status of the consumed event';
COMMENT ON COLUMN inbox_events.created_at IS 'When the event was first received and stored';
COMMENT ON COLUMN inbox_events.processed_at IS 'When the event was successfully processed';
COMMENT ON COLUMN inbox_events.scheduled_for IS 'When the event should be processed (for delayed processing or retries)';
COMMENT ON COLUMN inbox_events.attempts IS 'Number of processing attempts made for this event';
COMMENT ON COLUMN inbox_events.last_error IS 'Last error message if event processing failed';
COMMENT ON COLUMN inbox_events.correlation_id IS 'For tracing requests across multiple services';
COMMENT ON COLUMN inbox_events.causation_id IS 'For linking cause-and-effect relationships between events';

COMMIT;
