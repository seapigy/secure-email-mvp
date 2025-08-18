-- Migration to add email access logs table for security hardening (Micro-Iteration 4.22)
-- This table tracks all decryption attempts for audit and rate limiting purposes
-- Enhanced for Micro-Iteration 4.22: Log every email retrieval attempt with timestamp, 
-- requesting IP, user agent, email_id, and result (success, failed password, expired, 
-- burn_after_read triggered, etc.)

CREATE TABLE IF NOT EXISTS email_access_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email_id TEXT NOT NULL,                    -- The email being accessed
    user_id TEXT,                              -- User ID if authenticated, NULL if anonymous
    ip_address TEXT NOT NULL,                  -- IP address of the request
    user_agent TEXT,                           -- User agent string from the request (Micro-Iteration 4.22)
    status TEXT NOT NULL,                      -- 'success' or 'fail'
    attempt_count INTEGER DEFAULT 1,           -- Current attempt count for this IP/email combination
    result TEXT NOT NULL,                      -- Detailed result: success, failed_password, expired, burn_after_read, rate_limited, etc. (Micro-Iteration 4.22)
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    -- Indexes for efficient querying
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Index for rate limiting queries (IP + email_id + time window)
CREATE INDEX IF NOT EXISTS idx_email_access_logs_rate_limit 
ON email_access_logs(ip_address, email_id, created_at);

-- Index for audit queries by email
CREATE INDEX IF NOT EXISTS idx_email_access_logs_email 
ON email_access_logs(email_id, created_at);

-- Index for audit queries by user
CREATE INDEX IF NOT EXISTS idx_email_access_logs_user 
ON email_access_logs(user_id, created_at);

-- Index for cleanup of old logs
CREATE INDEX IF NOT EXISTS idx_email_access_logs_created_at 
ON email_access_logs(created_at);

-- Index for result-based queries (Micro-Iteration 4.22)
CREATE INDEX IF NOT EXISTS idx_email_access_logs_result 
ON email_access_logs(result, created_at);

-- Index for user agent analysis (Micro-Iteration 4.22)
CREATE INDEX IF NOT EXISTS idx_email_access_logs_user_agent 
ON email_access_logs(user_agent, created_at);

-- Add comments to document the enhanced audit logging system
-- email_access_logs: Enhanced table for tracking all email access attempts with detailed metadata
-- 
-- Fields added for Micro-Iteration 4.22:
-- - user_agent: Captures browser/client information for security analysis
-- - result: Detailed outcome of access attempt for better audit trail
-- 
-- Result values include:
-- - success: Successful email retrieval
-- - failed_password: Password protection failed
-- - expired: Email has expired
-- - burn_after_read: Email was consumed by burn-after-read
-- - rate_limited: Access blocked by rate limiting
-- - concurrent_blocked: Access blocked by concurrent access protection
-- - unauthorized: User not authorized to access
-- - not_found: Email not found
-- - system_error: System error during access
