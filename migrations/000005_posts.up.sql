-- Migration 000005 (UP): posts (a user's "personal homepage" content).
CREATE TABLE posts (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    media_url  text,          -- image/video URL (uploaded to storage by the client)
    media_type text,          -- 'image' | 'video' | null (text-only)
    caption    text,
    -- who can see this post:
    --   public  -> anyone
    --   friends -> only accepted friends (V-friends)
    --   private -> only the author
    visibility text NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'friends', 'private')),
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Fetch a user's posts newest-first, fast.
CREATE INDEX idx_posts_user_created ON posts (user_id, created_at DESC);
