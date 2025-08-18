-- Migration: Add failed attempts counter for self-destruct enforcement (Micro-Iteration 4.8)
-- This migration adds a failed_attempts counter to track failed access attempts
-- and enable self-destruct functionality when threshold is reached

-- Add failed attempts counter column to emails table
-- This field tracks the number of failed access attempts for self-destruct enforcement
ALTER TABLE emails ADD COLUMN failed_attempts INTEGER DEFAULT 0;

-- Create index for performance on failed attempts counter
CREATE INDEX IF NOT EXISTS idx_emails_failed_attempts ON emails(failed_attempts);

-- Add comments to document the new field
/*
Failed Attempts Counter (Micro-Iteration 4.8):

failed_attempts (INTEGER, default 0):
  - Tracks the number of failed access attempts for each email
  - Incremented on any failed access (wrong password, failed MFA, invalid decoy, unauthorized)
  - Reset to 0 on successful access
  - When failed_attempts >= self_destruct_threshold, email is securely deleted
  - Provides brute force protection and automatic cleanup of compromised emails
  - Counter persists across sessions and server restarts
  - Never exposed to clients to prevent information leakage
*/
