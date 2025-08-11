-- Migration: Add enhanced city and country verification fields to emails table
-- Micro-Iteration 4.15: Enhanced Geolocation Verification with City and Country Support

-- Add geolocation verification type field
ALTER TABLE emails ADD COLUMN geo_verification_type TEXT CHECK (geo_verification_type IN ('none', 'country', 'city', 'city_country')) DEFAULT 'none';

-- Add geolocation verification city and country fields
ALTER TABLE emails ADD COLUMN geo_city TEXT;
ALTER TABLE emails ADD COLUMN geo_country TEXT;

-- Add index for geolocation verification queries
CREATE INDEX IF NOT EXISTS idx_emails_geo_verification ON emails(geo_verification_type, geo_city, geo_country);

-- Add comment to document the new fields
-- geo_verification_type: Type of geolocation verification required
--   - 'none': No geolocation verification required
--   - 'country': Only country verification required
--   - 'city': Only city verification required
--   - 'city_country': Both city and country verification required
-- geo_city: City name for verification (case-insensitive, normalized)
-- geo_country: ISO 3166-1 alpha-2 country code for verification (case-insensitive)
--
-- This enhancement provides more flexible geolocation verification options
-- compared to the previous single city/country restriction system.
-- The verification type determines which checks are performed during access.
