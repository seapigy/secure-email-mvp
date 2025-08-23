-- =============================================================================
-- PHASE 3 DATABASE SCHEMA - VIEWING & REPLY FLOW
-- =============================================================================

-- View sessions for external users
CREATE TABLE IF NOT EXISTS link_view_sessions (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT 1,
    email_viewed BOOLEAN NOT NULL DEFAULT 0,
    viewed_at DATETIME,
    session_token TEXT NOT NULL UNIQUE,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- Secure replies from external users
CREATE TABLE IF NOT EXISTS secure_replies (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    chain_id TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    recipient_email TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at DATETIME,
    status TEXT NOT NULL DEFAULT 'pending',
    internal_email_id TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- Email chains for conversation management
CREATE TABLE IF NOT EXISTS email_chains (
    id TEXT PRIMARY KEY,
    initial_link_id TEXT NOT NULL,
    internal_user_id TEXT NOT NULL,
    external_email TEXT NOT NULL,
    subject TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_activity DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    message_count INTEGER NOT NULL DEFAULT 1,
    expires_at DATETIME,
    FOREIGN KEY (initial_link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (internal_user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Chain messages for conversation history
CREATE TABLE IF NOT EXISTS chain_messages (
    id TEXT PRIMARY KEY,
    chain_id TEXT NOT NULL,
    message_type TEXT NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    sender_email TEXT NOT NULL,
    sender_type TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    link_id TEXT,
    email_id TEXT,
    FOREIGN KEY (chain_id) REFERENCES email_chains(id) ON DELETE CASCADE,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE SET NULL,
    FOREIGN KEY (email_id) REFERENCES emails(id) ON DELETE SET NULL
);

-- Secure attachments for external access
CREATE TABLE IF NOT EXISTS secure_attachments (
    id TEXT PRIMARY KEY,
    link_id TEXT NOT NULL,
    original_name TEXT NOT NULL,
    secure_name TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    storage_path TEXT NOT NULL,
    encryption_key TEXT NOT NULL,
    checksum TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    download_count INTEGER NOT NULL DEFAULT 0,
    max_downloads INTEGER NOT NULL DEFAULT -1,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE
);

-- Attachment downloads tracking
CREATE TABLE IF NOT EXISTS attachment_downloads (
    id TEXT PRIMARY KEY,
    attachment_id TEXT NOT NULL,
    session_token TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    downloaded_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_view_sessions_link_id ON link_view_sessions(link_id);
CREATE INDEX IF NOT EXISTS idx_view_sessions_token ON link_view_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_view_sessions_expires ON link_view_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_secure_replies_link_id ON secure_replies(link_id);
CREATE INDEX IF NOT EXISTS idx_secure_replies_chain_id ON secure_replies(chain_id);
CREATE INDEX IF NOT EXISTS idx_secure_replies_status ON secure_replies(status);

CREATE INDEX IF NOT EXISTS idx_email_chains_initial_link ON email_chains(initial_link_id);
CREATE INDEX IF NOT EXISTS idx_email_chains_internal_user ON email_chains(internal_user_id);
CREATE INDEX IF NOT EXISTS idx_email_chains_external_email ON email_chains(external_email);
CREATE INDEX IF NOT EXISTS idx_email_chains_status ON email_chains(status);

CREATE INDEX IF NOT EXISTS idx_chain_messages_chain_id ON chain_messages(chain_id);
CREATE INDEX IF NOT EXISTS idx_chain_messages_type ON chain_messages(message_type);

CREATE INDEX IF NOT EXISTS idx_secure_attachments_link_id ON secure_attachments(link_id);
CREATE INDEX IF NOT EXISTS idx_attachment_downloads_attachment_id ON attachment_downloads(attachment_id);
