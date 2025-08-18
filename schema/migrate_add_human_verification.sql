-- Micro-Iteration 4.16: Anti-Automation & Human Verification Layer
-- Add human verification logging table

CREATE TABLE IF NOT EXISTS human_verification_logs (
    id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    verification_type TEXT NOT NULL, -- 'captcha' or 'proof_of_work'
    challenge_id TEXT,
    result TEXT NOT NULL, -- 'success', 'failure', 'timeout'
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    details TEXT, -- JSON field for additional verification details
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Indexes for abuse detection and audit queries
CREATE INDEX IF NOT EXISTS idx_human_verification_logs_email_ip ON human_verification_logs(email_id, ip_address);
CREATE INDEX IF NOT EXISTS idx_human_verification_logs_timestamp ON human_verification_logs(timestamp);
CREATE INDEX IF NOT EXISTS idx_human_verification_logs_ip_timestamp ON human_verification_logs(ip_address, timestamp);
CREATE INDEX IF NOT EXISTS idx_human_verification_logs_result ON human_verification_logs(result, timestamp);
