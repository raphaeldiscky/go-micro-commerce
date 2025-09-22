CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    user_type TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'participant',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    left_at TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- User type: user, admin
ALTER TABLE participants
ADD CONSTRAINT chk_participants_user_type
CHECK (user_type IN ('user', 'admin'));

-- Role: participant, moderator, observer
ALTER TABLE participants
ADD CONSTRAINT chk_participants_role
CHECK (role IN ('participant', 'moderator', 'observer'));

-- Unique constraint: one participant per user per conversation
ALTER TABLE participants
ADD CONSTRAINT unq_participants_conversation_user
UNIQUE (conversation_id, user_id, user_type);

-- Indexes for efficient querying
CREATE INDEX idx_participants_conversation_id ON participants(conversation_id);
CREATE INDEX idx_participants_user ON participants(user_id, user_type);
CREATE INDEX idx_participants_active ON participants(conversation_id, is_active) WHERE is_active = TRUE;
CREATE INDEX idx_participants_joined_at ON participants(joined_at DESC);