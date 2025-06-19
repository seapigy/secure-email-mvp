-- schema/temp_totp.sql
CREATE TABLE IF NOT EXISTS temp_totp (
    temp_id TEXT PRIMARY KEY,
    email TEXT NOT NULL,
    password_hash BLOB NOT NULL,
    totp_secret TEXT NOT NULL,
    expires_at INTEGER NOT NULL
); 