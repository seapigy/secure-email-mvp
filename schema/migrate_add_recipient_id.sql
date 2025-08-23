-- =============================================================================
-- SECURE EMAIL MVP - RECIPIENT ID MIGRATION
-- =============================================================================
-- This migration adds recipient_id field to support Micro-Iteration 4.18:
-- Secure Email Forwarding Prevention
-- =============================================================================

-- Add recipient_id field to emails table
-- This field stores the user_id of the intended recipient for access control
ALTER TABLE emails ADD COLUMN recipient_id TEXT;

-- Add index for recipient_id lookups
CREATE INDEX IF NOT EXISTS idx_emails_recipient_id ON emails(recipient_id);

-- Add foreign key constraint to ensure recipient_id references valid users
-- Note: This is optional since recipient might not be a registered user
-- ALTER TABLE emails ADD CONSTRAINT fk_emails_recipient_id 
--     FOREIGN KEY (recipient_id) REFERENCES users(user_id);

-- Add comment to document the new field
-- recipient_id: User ID of the intended recipient (for access control)
--              NULL if recipient is not a registered user
--              Used to prevent unauthorized access and forwarding











