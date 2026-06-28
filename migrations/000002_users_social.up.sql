-- Migration 000002 (UP): interests, user_interests, and friends.
-- The `users` table already exists from 000001.

-- A fixed list of interests users can pick from (for matching later).
CREATE TABLE interests (
    id   serial PRIMARY KEY,
    name text UNIQUE NOT NULL
);

-- Seed a starter set so the app has options on day one.
INSERT INTO interests (name) VALUES
    ('Music'), ('Travel'), ('Sports'), ('Movies'), ('Gaming'),
    ('Cooking'), ('Reading'), ('Art'), ('Technology'), ('Fitness'),
    ('Photography'), ('Languages'), ('Nature'), ('Dancing'), ('Fashion');

-- Which interests each user picked (many-to-many link table).
CREATE TABLE user_interests (
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest_id int  NOT NULL REFERENCES interests(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, interest_id)
);

-- Friendships ("V-friends"). One row per relationship.
--   status 'pending'  -> user_id sent a request to friend_id
--   status 'accepted' -> they are friends (treated as bidirectional)
CREATE TABLE friends (
    user_id    uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id  uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status     text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, friend_id)
);

-- Speed up "show me this user's friends/requests".
CREATE INDEX idx_friends_friend_id ON friends (friend_id);
