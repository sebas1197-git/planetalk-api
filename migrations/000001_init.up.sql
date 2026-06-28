-- Migration 000001 (UP): create the first tables (users + otp_codes).

-- gen_random_uuid() lives in the pgcrypto extension; enable it once.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    phone text UNIQUE NOT NULL,
    display_name text,
    bio text,
    avatar_url text,
    gender text,
    birthdate date,
    country text,
    language text,
    is_banned boolean NOT NULL DEFAULT false,
    last_seen timestamptz,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE otp_codes (
    id  uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    phone text NOT NULL ,
    code_hash text NOT NULL ,
    expires_at timestamptz NOT NULL ,
    attempts int NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- We look up OTPs by phone often, so index it for speed.
CREATE INDEX idx_otp_codes_phone ON otp_codes (phone);