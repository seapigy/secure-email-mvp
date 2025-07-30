-- Migration: Add encryption fields for AES-256-GCM
-- Adds nonce and auth tag fields to the emails table

-- Add nonce field (12 bytes for AES-256-GCM)
ALTER TABLE emails ADD COLUMN encryption_nonce TEXT;

-- Add auth tag field (16 bytes for AES-256-GCM)
ALTER TABLE emails ADD COLUMN encryption_auth_tag TEXT;

-- Update existing records to have placeholder values
-- (Note: Existing records won't be decryptable without these fields)
UPDATE emails SET 
    encryption_nonce = 'placeholder_nonce',
    encryption_auth_tag = 'placeholder_auth_tag'
WHERE encryption_nonce IS NULL OR encryption_auth_tag IS NULL; 