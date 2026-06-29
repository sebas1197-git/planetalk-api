-- Migration 000004 (UP): chat messages within a match.
CREATE TABLE messages (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id        uuid NOT NULL REFERENCES matches(id) ON DELETE CASCADE,
    sender_id       uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    body            text NOT NULL,          -- original text as typed by the sender
    translated_body text,                   -- translated into the recipient's language
    src_lang        text,                   -- detected source language (e.g. "en")
    dst_lang        text,                   -- target language (recipient's, e.g. "es")
    created_at      timestamptz NOT NULL DEFAULT now()
);

-- Fetch a conversation's messages in time order, fast.
CREATE INDEX idx_messages_match_created ON messages (match_id, created_at);
