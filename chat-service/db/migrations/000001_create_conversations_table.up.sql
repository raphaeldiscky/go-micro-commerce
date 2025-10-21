CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status TEXT NOT NULL DEFAULT 'waiting',
    subject TEXT,
    priority INT NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at TIMESTAMPTZ
);

-- Status: waiting, active, ended, transferred
ALTER TABLE conversations
ADD CONSTRAINT chk_conversations_status
CHECK (status IN ('waiting', 'active', 'ended', 'transferred'));

-- Priority: 1=low, 2=normal, 3=high, 4=urgent
ALTER TABLE conversations
ADD CONSTRAINT chk_conversations_priority
CHECK (priority BETWEEN 1 AND 4);

-- Indexes for common queries
CREATE INDEX idx_conversations_status ON conversations (status);
CREATE INDEX idx_conversations_priority ON conversations (priority DESC);
CREATE INDEX idx_conversations_created_at ON conversations (created_at DESC);
CREATE INDEX idx_conversations_status_priority ON conversations (status, priority DESC) WHERE status IN ('waiting', 'active');
