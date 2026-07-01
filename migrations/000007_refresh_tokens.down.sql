-- Migration 000007 (DOWN): undo 000007_refresh_tokens.up.sql.
DROP TABLE IF EXISTS refresh_tokens;
