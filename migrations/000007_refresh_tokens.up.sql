-- Migration 000007 (UP): DB-backed refresh tokens (for rotation + revocation).
-- We store only a HASH of each token, never the raw value.
CREATE TABLE refresh_tokens (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash text NOT NULL UNIQUE,     -- sha256 of the opaque token
    expires_at timestamptz NOT NULL,
    revoked    boolean NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);
