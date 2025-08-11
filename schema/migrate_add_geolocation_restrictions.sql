-- Migration: Add geolocation restrictions to emails table
-- Micro-Iteration 4.11: Country & City-Level Geolocation Restrictions

-- Add allowed_countries field (JSON array of lowercase ISO 3166-1 alpha-2 country codes)
ALTER TABLE emails ADD COLUMN allowed_countries TEXT;

-- Add allowed_cities field (JSON array of lowercase normalized city names)
ALTER TABLE emails ADD COLUMN allowed_cities TEXT;

-- Add index for geolocation queries
CREATE INDEX IF NOT EXISTS idx_emails_geolocation ON emails(allowed_countries, allowed_cities);

-- Add comment to document the new fields
-- allowed_countries: JSON array of lowercase ISO 3166-1 alpha-2 country codes (e.g., ["us", "ca", "gb"])
-- allowed_cities: JSON array of lowercase normalized city names (e.g., ["new york", "toronto"])
-- Both fields are optional. If both are empty/null, no geolocation restriction is applied.
-- If both are populated, the requester must pass both checks (AND logic).
