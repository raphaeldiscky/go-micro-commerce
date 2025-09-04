-- Create saga_states table
CREATE TABLE IF NOT EXISTS saga_states (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    current_step BIGINT NOT NULL DEFAULT 0,
    version BIGINT NOT NULL DEFAULT 1,
    executed_steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    compensated_steps JSONB NOT NULL DEFAULT '[]'::jsonb,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    error TEXT,
    retry_count BIGINT NOT NULL DEFAULT 0,
    last_retry_at TIMESTAMPTZ,
    timeout_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    
    -- Foreign key constraint
    CONSTRAINT fk_saga_states_order_id 
        FOREIGN KEY (order_id) 
        REFERENCES orders(id) 
        ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX idx_saga_states_order_id ON saga_states(order_id);
CREATE INDEX idx_saga_states_status ON saga_states(status);
CREATE INDEX idx_saga_states_updated_at ON saga_states(updated_at);
CREATE INDEX idx_saga_states_created_at ON saga_states(created_at);
CREATE INDEX idx_saga_states_version ON saga_states(version);
CREATE INDEX idx_saga_states_retry ON saga_states(retry_count, last_retry_at);
CREATE INDEX idx_saga_states_timeout ON saga_states(timeout_at);


-- Composite index for status and updated_at (useful for recovery queries)
CREATE INDEX idx_saga_states_status_updated 
    ON saga_states(status, updated_at);

-- Partial index for finding sagas that need recovery
CREATE INDEX idx_saga_states_recovery 
    ON saga_states(status, updated_at, timeout_at) 
    WHERE status IN ('pending', 'executing', 'failed', 'compensating');

-- Index for finding completed sagas for cleanup
CREATE INDEX idx_saga_states_cleanup 
    ON saga_states(status, completed_at) 
    WHERE status IN ('completed', 'compensated');


COMMENT ON TABLE saga_states IS 'Stores the execution state of order processing sagas';
COMMENT ON COLUMN saga_states.id IS 'Unique identifier for the saga instance';
COMMENT ON COLUMN saga_states.order_id IS 'Reference to the order being processed';
COMMENT ON COLUMN saga_states.status IS 'Current status of the saga (pending, executing, compensating, completed, failed, compensated)';
COMMENT ON COLUMN saga_states.current_step IS 'Index of the current/last executed step';
COMMENT ON COLUMN saga_states.executed_steps IS 'Array of step names that have been successfully executed';
COMMENT ON COLUMN saga_states.compensated_steps IS 'Array of step names that have been compensated';
COMMENT ON COLUMN saga_states.data IS 'Shared data between saga steps stored as JSON';
COMMENT ON COLUMN saga_states.error IS 'Error message if the saga failed';
COMMENT ON COLUMN saga_states.created_at IS 'Timestamp when the saga was created';
COMMENT ON COLUMN saga_states.updated_at IS 'Timestamp of the last update';
COMMENT ON COLUMN saga_states.completed_at IS 'Timestamp when the saga completed or failed';
COMMENT ON COLUMN saga_states.version IS 'Version number for optimistic locking to prevent race conditions';
COMMENT ON COLUMN saga_states.retry_count IS 'Number of times this saga has been retried';
COMMENT ON COLUMN saga_states.last_retry_at IS 'Timestamp of the last retry attempt';
COMMENT ON COLUMN saga_states.timeout_at IS 'Timestamp when this saga should timeout';

-- Create trigger to automatically update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_saga_states_updated_at 
    BEFORE UPDATE ON saga_states 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Create enum type for saga status (optional but recommended for type safety)
DO $$ BEGIN
    CREATE TYPE saga_status AS ENUM (
        'pending',
        'executing',
        'compensating',
        'completed',
        'failed',
        'compensated'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;