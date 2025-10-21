CREATE TABLE connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    connection_id TEXT NOT NULL UNIQUE,
    socket_id TEXT NOT NULL,
    user_agent TEXT,
    ip_address INET,
    connected_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT now(),
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Indexes for connection management
CREATE INDEX idx_connections_user ON connections (user_id);
CREATE INDEX idx_connections_active ON connections (is_active) WHERE is_active
= TRUE;
CREATE INDEX idx_connections_heartbeat ON connections (
    last_heartbeat
) WHERE is_active
= TRUE;
CREATE INDEX idx_connections_connection_id ON connections (connection_id);

-- Cleanup old inactive connections periodically
CREATE INDEX idx_connections_cleanup ON connections (
    last_heartbeat
) WHERE is_active
= FALSE;
