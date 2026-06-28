-- Migration 000001 (DOWN): undo 000001_init.up.sql, in reverse order.
DROP TABLE IF EXISTS otp_codes;
DROP TABLE IF EXISTS users;
-- Leave the pgcrypto extension; later tables use it too.


