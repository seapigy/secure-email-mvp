-- Migration: Add enhanced geo-restriction features to emails table
-- Micro-Iteration 4.7: GeoIP Country Restriction

-- Add enhanced geo-restriction rules (JSON array of rules)
ALTER TABLE emails ADD COLUMN geo_restriction_rules TEXT;

-- Add geo-restriction configuration (JSON object)
ALTER TABLE emails ADD COLUMN geo_restriction_config TEXT;

-- Add geo-restriction enforcement status
ALTER TABLE emails ADD COLUMN geo_restriction_enabled INTEGER DEFAULT 1;

-- Add geo-restriction violation tracking
ALTER TABLE emails ADD COLUMN geo_restriction_violations INTEGER DEFAULT 0;

-- Add geo-restriction last violation timestamp
ALTER TABLE emails ADD COLUMN geo_restriction_last_violation DATETIME;

-- Add index for geo-restriction queries
CREATE INDEX IF NOT EXISTS idx_emails_geo_restriction ON emails(geo_restriction_enabled, geo_restriction_violations);

-- Add comment to document the new fields
-- geo_restriction_rules: JSON array of geo-restriction rules with allow/block logic
-- geo_restriction_config: JSON object with configuration settings (strict mode, default action, etc.)
-- geo_restriction_enabled: Boolean flag to enable/disable geo-restrictions for this email
-- geo_restriction_violations: Counter for tracking violation attempts
-- geo_restriction_last_violation: Timestamp of the last violation attempt
-- 
-- Example geo_restriction_rules:
-- [
--   {
--     "id": "rule1",
--     "type": "allow",
--     "countries": ["us", "ca"],
--     "cities": ["new york", "toronto"],
--     "description": "Allow access from US and Canada"
--   },
--   {
--     "id": "rule2", 
--     "type": "block",
--     "countries": ["ru", "cn"],
--     "description": "Block access from Russia and China"
--   }
-- ]
--
-- Example geo_restriction_config:
-- {
--   "enabled": true,
--   "default_action": "allow",
--   "strict_mode": false,
--   "log_violations": true,
--   "block_on_geolocation_failure": true
-- }





















