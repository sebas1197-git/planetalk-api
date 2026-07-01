-- Migration 000006 (DOWN): undo 000006_moderation.up.sql, reverse order.
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS blocks;
