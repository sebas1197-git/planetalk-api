-- Migration 000003 (UP): matches (a paired 1v1 video session).
CREATE TABLE matches (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_a       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel_name text NOT NULL,                 -- the Agora video channel for this match
    status       text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'ended')),
    started_at   timestamptz NOT NULL DEFAULT now(),
    ended_at     timestamptz
);

-- Look up a user's matches quickly (history, "am I in this match?").
CREATE INDEX idx_matches_user_a ON matches (user_a);
CREATE INDEX idx_matches_user_b ON matches (user_b);
