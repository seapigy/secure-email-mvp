-- Migration to add advanced watermarking features (Iteration 8)
-- Extends existing watermarking with recipient-specific watermarks, templates, and enhanced audit logging

-- Add new columns to existing watermark_configs table
ALTER TABLE watermark_configs ADD COLUMN recipient_email TEXT;
ALTER TABLE watermark_configs ADD COLUMN recipient_id TEXT;
ALTER TABLE watermark_configs ADD COLUMN watermark_type TEXT DEFAULT 'text';
ALTER TABLE watermark_configs ADD COLUMN content_type TEXT DEFAULT 'pdf';
ALTER TABLE watermark_configs ADD COLUMN watermark_data TEXT; -- JSON data for complex watermarks
ALTER TABLE watermark_configs ADD COLUMN is_recipient_specific BOOLEAN DEFAULT FALSE;

-- Create watermark templates table
CREATE TABLE IF NOT EXISTS watermark_templates (
    template_id TEXT PRIMARY KEY,
    template_name TEXT NOT NULL,
    template_description TEXT,
    watermark_type TEXT NOT NULL, -- 'text', 'image', 'audio', 'video', 'inline'
    content_types TEXT NOT NULL, -- JSON array of supported content types
    default_config TEXT NOT NULL, -- JSON configuration
    is_recipient_specific BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    created_by TEXT
);

-- Create advanced watermark audit log table
CREATE TABLE IF NOT EXISTS advanced_watermark_audit_log (
    audit_id TEXT PRIMARY KEY,
    link_id TEXT,
    attachment_id TEXT,
    content_id TEXT, -- For inline content
    template_id TEXT,
    recipient_email TEXT NOT NULL,
    recipient_id TEXT,
    watermark_type TEXT NOT NULL,
    content_type TEXT NOT NULL,
    watermark_config TEXT, -- JSON configuration used
    applied_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success BOOLEAN NOT NULL,
    error_message TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_by TEXT,
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id) ON DELETE CASCADE,
    FOREIGN KEY (attachment_id) REFERENCES secure_attachments(attachment_id) ON DELETE CASCADE,
    FOREIGN KEY (template_id) REFERENCES watermark_templates(template_id) ON DELETE SET NULL
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_watermark_configs_recipient ON watermark_configs(recipient_email, recipient_id);
CREATE INDEX IF NOT EXISTS idx_watermark_configs_type ON watermark_configs(watermark_type, content_type);
CREATE INDEX IF NOT EXISTS idx_watermark_templates_active ON watermark_templates(is_active, watermark_type);
CREATE INDEX IF NOT EXISTS idx_advanced_watermark_audit_link ON advanced_watermark_audit_log(link_id);
CREATE INDEX IF NOT EXISTS idx_advanced_watermark_audit_recipient ON advanced_watermark_audit_log(recipient_email);
CREATE INDEX IF NOT EXISTS idx_advanced_watermark_audit_type ON advanced_watermark_audit_log(watermark_type, content_type);

-- Insert default watermark templates
INSERT OR IGNORE INTO watermark_templates (template_id, template_name, template_description, watermark_type, content_types, default_config, is_recipient_specific, is_active) VALUES
('template_basic_text', 'Basic Text Watermark', 'Simple text watermark for documents', 'text', '["pdf", "image", "document"]', '{"position": "bottom-right", "opacity": 0.7, "font_size": 12, "color": "#FF0000", "rotation": -45}', FALSE, TRUE),
('template_recipient_specific', 'Recipient-Specific Watermark', 'Watermark with recipient email and timestamp', 'text', '["pdf", "image", "document"]', '{"position": "bottom-right", "opacity": 0.8, "font_size": 14, "color": "#FF0000", "rotation": -45, "include_recipient": true, "include_timestamp": true}', TRUE, TRUE),
('template_audio_watermark', 'Audio Watermark', 'Inaudible watermark for audio files', 'audio', '["audio"]', '{"frequency": 18000, "volume": -30, "pattern": "recipient_id", "duration": 0.1}', TRUE, TRUE),
('template_video_overlay', 'Video Overlay Watermark', 'Visual overlay for video files', 'video', '["video"]', '{"position": "bottom-right", "opacity": 0.6, "font_size": 16, "color": "#FFFFFF", "background_color": "#000000", "include_recipient": true}', TRUE, TRUE),
('template_inline_image', 'Inline Image Watermark', 'Watermark for inline images in email content', 'inline', '["email_content"]', '{"position": "bottom-right", "opacity": 0.5, "font_size": 10, "color": "#FF0000", "rotation": 0}', FALSE, TRUE);

-- Update compliance_audit_log to include advanced watermarking fields
ALTER TABLE compliance_audit_log ADD COLUMN watermark_template_id TEXT;
ALTER TABLE compliance_audit_log ADD COLUMN recipient_watermark_id TEXT;
ALTER TABLE compliance_audit_log ADD COLUMN watermark_audit_id TEXT;

-- Add foreign key for watermark template reference
CREATE INDEX IF NOT EXISTS idx_compliance_audit_watermark_template ON compliance_audit_log(watermark_template_id);

-- Update database version
PRAGMA user_version = 8;
