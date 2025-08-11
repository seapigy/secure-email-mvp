-- Migration: Add simple geolocation restrictions to emails table
-- Micro-Iteration 4.10: Geolocation-Based Email Access Restrictions

-- Add allowed_city field (single city name, case-insensitive, normalized)
ALTER TABLE emails ADD COLUMN allowed_city TEXT;

-- Add allowed_country field (single ISO 3166-1 alpha-2 country code, case-insensitive)
ALTER TABLE emails ADD COLUMN allowed_country TEXT;

-- Add index for geolocation queries
CREATE INDEX IF NOT EXISTS idx_emails_simple_geolocation ON emails(allowed_city, allowed_country);

-- Add comment to document the new fields
-- allowed_city: Single city name (case-insensitive, normalized)
-- allowed_country: Single ISO 3166-1 alpha-2 country code (case-insensitive)
-- Both fields are optional. If both are set, requester must match both (AND logic).
-- If only one is set, requester must match that one.
-- If neither is set, no geolocation restriction is applied.
