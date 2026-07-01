-- Migration 000006 (UP): moderation — blocks and reports.

-- One row per "A has blocked B". Blocking is one-directional in storage, but
-- enforced both ways (if either blocked the other, they can't interact).
CREATE TABLE blocks (
    blocker_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (blocker_id, blocked_id)
);
CREATE INDEX idx_blocks_blocked ON blocks (blocked_id);

-- User reports (for the safety / review queue).
CREATE TABLE reports (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reported_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason      text NOT NULL,
    context     text,           -- optional free-text / where it happened
    status      text NOT NULL DEFAULT 'open'
                 CHECK (status IN ('open', 'reviewed', 'actioned', 'dismissed')),
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_reports_reported ON reports (reported_id);
