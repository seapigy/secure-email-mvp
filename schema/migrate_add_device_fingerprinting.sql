-- =============================================================================
-- MICRO-ITERATION 4.14: DEVICE FINGERPRINTING & TRUSTED DEVICES
-- =============================================================================
-- Migration to add device fingerprinting and trusted devices functionality
--
-- This migration adds support for:
-- - trusted_devices_only: Boolean flag to require trusted devices only
-- - trusted_devices table: Stores device fingerprints for each email
-- - device fingerprinting: Hash-based device identification
--
-- Device fingerprinting allows emails to be restricted to specific devices
-- that have been previously authorized through MFA verification.
-- =============================================================================

-- Add trusted_devices_only column to emails table
ALTER TABLE emails ADD COLUMN trusted_devices_only BOOLEAN DEFAULT FALSE;

-- Create trusted_devices table
CREATE TABLE IF NOT EXISTS trusted_devices (
    id TEXT PRIMARY KEY,  -- UUID for the trusted device record
    email_id TEXT NOT NULL,  -- Foreign key to emails table
    device_hash TEXT NOT NULL,  -- Argon2id hash of device fingerprint
    device_fingerprint TEXT NOT NULL,  -- Original fingerprint (for debugging)
    user_agent TEXT,  -- User-Agent string
    ip_address TEXT,  -- IP address (for audit purposes)
    added_at DATETIME DEFAULT CURRENT_TIMESTAMP,  -- When device was trusted
    last_used_at DATETIME,  -- Last successful access from this device
    access_count INTEGER DEFAULT 0,  -- Number of successful accesses
    
    -- Foreign key constraint
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_id ON trusted_devices(email_id);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_device_hash ON trusted_devices(device_hash);
CREATE INDEX IF NOT EXISTS idx_trusted_devices_email_hash ON trusted_devices(email_id, device_hash);

-- Add index for trusted_devices_only queries
CREATE INDEX IF NOT EXISTS idx_emails_trusted_devices_only ON emails(trusted_devices_only);

-- Add comments for documentation
PRAGMA table_info(emails);
PRAGMA table_info(trusted_devices);

-- Verify the migration was applied successfully
SELECT
    name,
    type,
    "notnull",
    dflt_value,
    pk
FROM pragma_table_info('emails')
WHERE name = 'trusted_devices_only'
ORDER BY name;

-- Verify trusted_devices table structure
SELECT
    name,
    type,
    "notnull",
    dflt_value,
    pk
FROM pragma_table_info('trusted_devices')
ORDER BY name;
