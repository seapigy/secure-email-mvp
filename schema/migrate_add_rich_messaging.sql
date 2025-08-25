-- Iteration 5: Rich Messaging (Attachments + Rich Text)
-- Migration to add support for rich text content and file attachments

-- Secure Attachments table
CREATE TABLE IF NOT EXISTS secure_attachments (
    attachment_id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    reply_id TEXT,
    email_id TEXT,
    filename TEXT NOT NULL,
    original_filename TEXT NOT NULL,
    file_size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    file_hash TEXT NOT NULL,
    s3_key TEXT NOT NULL,
    s3_bucket TEXT NOT NULL,
    encryption_key_id TEXT,
    virus_scan_status TEXT DEFAULT 'pending', -- pending, clean, infected, error
    virus_scan_result TEXT,
    virus_scan_timestamp DATETIME,
    download_count INTEGER DEFAULT 0,
    max_downloads INTEGER DEFAULT 10,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    status TEXT DEFAULT 'active', -- active, deleted, expired
    metadata TEXT, -- JSON metadata (watermark info, tags, etc.)
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Rich Text Content table for storing sanitized HTML
CREATE TABLE IF NOT EXISTS rich_text_content (
    content_id TEXT PRIMARY KEY,
    link_id TEXT,
    reply_id TEXT,
    email_id TEXT,
    content_type TEXT NOT NULL, -- 'email_body', 'reply_body'
    raw_content TEXT, -- Original content before sanitization
    sanitized_content TEXT NOT NULL, -- Sanitized HTML content
    content_hash TEXT NOT NULL,
    features_used TEXT, -- JSON array of rich text features used
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- File Type Whitelist for security
CREATE TABLE IF NOT EXISTS allowed_file_types (
    mime_type TEXT PRIMARY KEY,
    extension TEXT NOT NULL,
    max_size INTEGER NOT NULL, -- in bytes
    is_image BOOLEAN DEFAULT FALSE,
    is_document BOOLEAN DEFAULT FALSE,
    is_archive BOOLEAN DEFAULT FALSE,
    risk_level TEXT DEFAULT 'low', -- low, medium, high
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Insert common allowed file types
INSERT OR IGNORE INTO allowed_file_types (mime_type, extension, max_size, is_image, is_document, is_archive, risk_level) VALUES
('application/pdf', '.pdf', 26214400, FALSE, TRUE, FALSE, 'low'),
('application/msword', '.doc', 10485760, FALSE, TRUE, FALSE, 'medium'),
('application/vnd.openxmlformats-officedocument.wordprocessingml.document', '.docx', 10485760, FALSE, TRUE, FALSE, 'low'),
('application/vnd.ms-excel', '.xls', 10485760, FALSE, TRUE, FALSE, 'medium'),
('application/vnd.openxmlformats-officedocument.spreadsheetml.sheet', '.xlsx', 10485760, FALSE, TRUE, FALSE, 'low'),
('application/vnd.ms-powerpoint', '.ppt', 10485760, FALSE, TRUE, FALSE, 'medium'),
('application/vnd.openxmlformats-officedocument.presentationml.presentation', '.pptx', 10485760, FALSE, TRUE, FALSE, 'low'),
('text/plain', '.txt', 5242880, FALSE, TRUE, FALSE, 'low'),
('image/jpeg', '.jpg', 10485760, TRUE, FALSE, FALSE, 'low'),
('image/jpeg', '.jpeg', 10485760, TRUE, FALSE, FALSE, 'low'),
('image/png', '.png', 10485760, TRUE, FALSE, FALSE, 'low'),
('image/gif', '.gif', 10485760, TRUE, FALSE, FALSE, 'low'),
('image/webp', '.webp', 10485760, TRUE, FALSE, FALSE, 'low'),
('application/zip', '.zip', 52428800, FALSE, FALSE, TRUE, 'high'),
('application/x-rar-compressed', '.rar', 52428800, FALSE, FALSE, TRUE, 'high');

-- Attachment Download Tokens for secure access
CREATE TABLE IF NOT EXISTS attachment_download_tokens (
    token_id TEXT PRIMARY KEY,
    attachment_id TEXT NOT NULL,
    token_hash TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    max_downloads INTEGER DEFAULT 1,
    download_count INTEGER DEFAULT 0,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE
);

-- Enhanced audit log for rich messaging events
CREATE TABLE IF NOT EXISTS rich_messaging_audit_log (
    audit_id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL, -- 'attachment_upload', 'attachment_download', 'rich_text_edit', 'virus_scan'
    link_id TEXT,
    reply_id TEXT,
    attachment_id TEXT,
    content_id TEXT,
    user_id TEXT,
    ip_address TEXT,
    user_agent TEXT,
    event_details TEXT, -- JSON details
    file_size INTEGER,
    mime_type TEXT,
    virus_scan_result TEXT,
    download_token_id TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (reply_id) REFERENCES secure_replies(reply_id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE,
    FOREIGN KEY (content_id) REFERENCES rich_text_content(content_id) ON DELETE CASCADE,
    FOREIGN KEY (download_token_id) REFERENCES attachment_download_tokens(token_id) ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_secure_attachments_link_id ON secure_attachments(link_id);
CREATE INDEX IF NOT EXISTS idx_secure_attachments_reply_id ON secure_attachments(reply_id);
CREATE INDEX IF NOT EXISTS idx_secure_attachments_status ON secure_attachments(status);
CREATE INDEX IF NOT EXISTS idx_secure_attachments_virus_scan_status ON secure_attachments(virus_scan_status);
CREATE INDEX IF NOT EXISTS idx_rich_text_content_link_id ON rich_text_content(link_id);
CREATE INDEX IF NOT EXISTS idx_rich_text_content_reply_id ON rich_text_content(reply_id);
CREATE INDEX IF NOT EXISTS idx_attachment_download_tokens_attachment_id ON attachment_download_tokens(attachment_id);
CREATE INDEX IF NOT EXISTS idx_attachment_download_tokens_expires_at ON attachment_download_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_rich_messaging_audit_log_event_type ON rich_messaging_audit_log(event_type);
CREATE INDEX IF NOT EXISTS idx_rich_messaging_audit_log_created_at ON rich_messaging_audit_log(created_at);

-- Triggers for audit logging
CREATE TRIGGER IF NOT EXISTS trigger_attachment_audit_log
AFTER INSERT ON secure_attachments
BEGIN
    INSERT INTO rich_messaging_audit_log (
        audit_id, event_type, link_id, reply_id, attachment_id, 
        user_id, ip_address, user_agent, event_details, 
        file_size, mime_type, created_at
    ) VALUES (
        'audit_' || hex(randomblob(8)), 'attachment_upload', 
        NEW.link_id, NEW.reply_id, NEW.attachment_id,
        NEW.created_by, '127.0.0.1', 'system', 
        json_object('filename', NEW.filename, 'original_filename', NEW.original_filename),
        NEW.file_size, NEW.mime_type, CURRENT_TIMESTAMP
    );
END;

CREATE TRIGGER IF NOT EXISTS trigger_rich_text_audit_log
AFTER INSERT ON rich_text_content
BEGIN
    INSERT INTO rich_messaging_audit_log (
        audit_id, event_type, link_id, reply_id, content_id,
        user_id, ip_address, user_agent, event_details, created_at
    ) VALUES (
        'audit_' || hex(randomblob(8)), 'rich_text_edit',
        NEW.link_id, NEW.reply_id, NEW.content_id,
        NEW.created_by, '127.0.0.1', 'system',
        json_object('content_type', NEW.content_type, 'features_used', NEW.features_used),
        CURRENT_TIMESTAMP
    );
END;

-- Update schema version
PRAGMA user_version = 5;
