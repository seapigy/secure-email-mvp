-- =============================================================================
-- MICRO-ITERATION 4.13: GEOFENCING & LOCATION-BASED ACCESS CONTROL
-- =============================================================================
-- Migration to add geofencing fields to the emails table
-- 
-- This migration adds support for:
-- - allowed_countries: JSON array of allowed country codes (ISO 3166-1 alpha-2)
-- - allowed_ip_ranges: JSON array of allowed CIDR ranges
-- - geofence_violations: Counter for geofence violation attempts
-- - geofence_last_violation: Timestamp of last geofence violation
--
-- Both fields can be NULL for unrestricted access
-- =============================================================================

-- Add geofencing fields to emails table
ALTER TABLE emails ADD COLUMN allowed_countries TEXT;
ALTER TABLE emails ADD COLUMN allowed_ip_ranges TEXT;
ALTER TABLE emails ADD COLUMN geofence_violations INTEGER DEFAULT 0;
ALTER TABLE emails ADD COLUMN geofence_last_violation DATETIME;

-- Create indexes for geofencing queries
CREATE INDEX IF NOT EXISTS idx_emails_geofence_countries ON emails(allowed_countries);
CREATE INDEX IF NOT EXISTS idx_emails_geofence_ip_ranges ON emails(allowed_ip_ranges);
CREATE INDEX IF NOT EXISTS idx_emails_geofence_violations ON emails(geofence_violations, geofence_last_violation);

-- Add comments for documentation
PRAGMA table_info(emails);

-- Verify the migration was applied successfully
SELECT 
    name, 
    type, 
    "notnull", 
    dflt_value, 
    pk 
FROM pragma_table_info('emails') 
WHERE name IN ('allowed_countries', 'allowed_ip_ranges', 'geofence_violations', 'geofence_last_violation')
ORDER BY name;
