CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations (
        id
    ) ON DELETE CASCADE,
    sender_id UUID, -- NULL for system messages
    content TEXT NOT NULL,  -- noqa: RF04
    message_type TEXT NOT NULL DEFAULT 'text',
    metadata JSONB DEFAULT '{}',
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Message type: text, image, file, system
ALTER TABLE messages
ADD CONSTRAINT chk_messages_type
CHECK (message_type IN ('text', 'image', 'file', 'system'));

-- System messages have no sender
ALTER TABLE messages
ADD CONSTRAINT chk_messages_system_sender
CHECK (
    (is_system = TRUE AND sender_id IS NULL)
    OR (is_system = FALSE AND sender_id IS NOT NULL)
);

-- Indexes for efficient querying
CREATE INDEX idx_messages_conversation_id ON messages (conversation_id);
CREATE INDEX idx_messages_conversation_created ON messages (
    conversation_id, created_at DESC
);
CREATE INDEX idx_messages_sender ON messages (
    sender_id
) WHERE sender_id IS NOT NULL;
CREATE INDEX idx_messages_created_at ON messages (created_at DESC);
