-- Migration 000002 (DOWN): undo 000002_users_social.up.sql, in reverse order.
DROP TABLE IF EXISTS friends;
DROP TABLE IF EXISTS user_interests;
DROP TABLE IF EXISTS interests;
