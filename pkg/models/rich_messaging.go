package models

import (
	"encoding/json"
	"time"
)

// SecureAttachment represents a file attachment for secure links and replies
type SecureAttachment struct {
	AttachmentID        string    `json:"attachment_id" db:"attachment_id"`
	LinkID              string    `json:"link_id" db:"link_id"`
	ReplyID             *string   `json:"reply_id,omitempty" db:"reply_id"`
	EmailID             *string   `json:"email_id,omitempty" db:"email_id"`
	Filename            string    `json:"filename" db:"filename"`
	OriginalFilename    string    `json:"original_filename" db:"original_filename"`
	FileSize            int64     `json:"file_size" db:"file_size"`
	MimeType            string    `json:"mime_type" db:"mime_type"`
	FileHash            string    `json:"file_hash" db:"file_hash"`
	S3Key               string    `json:"s3_key" db:"s3_key"`
	S3Bucket            string    `json:"s3_bucket" db:"s3_bucket"`
	EncryptionKeyID     *string   `json:"encryption_key_id,omitempty" db:"encryption_key_id"`
	VirusScanStatus     string    `json:"virus_scan_status" db:"virus_scan_status"`
	VirusScanResult     *string   `json:"virus_scan_result,omitempty" db:"virus_scan_result"`
	VirusScanTimestamp  *time.Time `json:"virus_scan_timestamp,omitempty" db:"virus_scan_timestamp"`
	DownloadCount       int       `json:"download_count" db:"download_count"`
	MaxDownloads        int       `json:"max_downloads" db:"max_downloads"`
	ExpiresAt           *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy           *string   `json:"created_by,omitempty" db:"created_by"`
	Status              string    `json:"status" db:"status"`
	Metadata            *string   `json:"metadata,omitempty" db:"metadata"`
}

// RichTextContent represents sanitized rich text content
type RichTextContent struct {
	ContentID         string    `json:"content_id" db:"content_id"`
	LinkID            *string   `json:"link_id,omitempty" db:"link_id"`
	ReplyID           *string   `json:"reply_id,omitempty" db:"reply_id"`
	EmailID           *string   `json:"email_id,omitempty" db:"email_id"`
	ContentType       string    `json:"content_type" db:"content_type"`
	RawContent        *string   `json:"raw_content,omitempty" db:"raw_content"`
	SanitizedContent  string    `json:"sanitized_content" db:"sanitized_content"`
	ContentHash       string    `json:"content_hash" db:"content_hash"`
	FeaturesUsed      *string   `json:"features_used,omitempty" db:"features_used"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
	CreatedBy         *string   `json:"created_by,omitempty" db:"created_by"`
}

// AllowedFileType represents allowed file types for uploads
type AllowedFileType struct {
	MimeType   string `json:"mime_type" db:"mime_type"`
	Extension  string `json:"extension" db:"extension"`
	MaxSize    int64  `json:"max_size" db:"max_size"`
	IsImage    bool   `json:"is_image" db:"is_image"`
	IsDocument bool   `json:"is_document" db:"is_document"`
	IsArchive  bool   `json:"is_archive" db:"is_archive"`
	RiskLevel  string `json:"risk_level" db:"risk_level"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// AttachmentDownloadToken represents a secure download token
type AttachmentDownloadToken struct {
	TokenID        string    `json:"token_id" db:"token_id"`
	AttachmentID   string    `json:"attachment_id" db:"attachment_id"`
	TokenHash      string    `json:"token_hash" db:"token_hash"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	MaxDownloads   int       `json:"max_downloads" db:"max_downloads"`
	DownloadCount  int       `json:"download_count" db:"download_count"`
	IPAddress      *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent      *string   `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	CreatedBy      *string   `json:"created_by,omitempty" db:"created_by"`
}

// RichMessagingAuditLog represents audit events for rich messaging
type RichMessagingAuditLog struct {
	AuditID         string    `json:"audit_id" db:"audit_id"`
	EventType       string    `json:"event_type" db:"event_type"`
	LinkID          *string   `json:"link_id,omitempty" db:"link_id"`
	ReplyID         *string   `json:"reply_id,omitempty" db:"reply_id"`
	AttachmentID    *string   `json:"attachment_id,omitempty" db:"attachment_id"`
	ContentID       *string   `json:"content_id,omitempty" db:"content_id"`
	UserID          *string   `json:"user_id,omitempty" db:"user_id"`
	IPAddress       *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent       *string   `json:"user_agent,omitempty" db:"user_agent"`
	EventDetails    *string   `json:"event_details,omitempty" db:"event_details"`
	FileSize        *int64    `json:"file_size,omitempty" db:"file_size"`
	MimeType        *string   `json:"mime_type,omitempty" db:"mime_type"`
	VirusScanResult *string   `json:"virus_scan_result,omitempty" db:"virus_scan_result"`
	DownloadTokenID *string   `json:"download_token_id,omitempty" db:"download_token_id"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// AttachmentUploadRequest represents a file upload request
type AttachmentUploadRequest struct {
	LinkID          string `json:"link_id"`
	ReplyID         string `json:"reply_id,omitempty"`
	Filename        string `json:"filename"`
	FileSize        int64  `json:"file_size"`
	MimeType        string `json:"mime_type"`
	FileHash        string `json:"file_hash"`
	IPAddress       string `json:"ip_address,omitempty"`
	UserAgent       string `json:"user_agent,omitempty"`
}

// AttachmentUploadResponse represents the response to a file upload
type AttachmentUploadResponse struct {
	Success         bool   `json:"success"`
	AttachmentID    string `json:"attachment_id,omitempty"`
	UploadURL       string `json:"upload_url,omitempty"`
	Message         string `json:"message,omitempty"`
	Error           string `json:"error,omitempty"`
	ErrorCode       string `json:"error_code,omitempty"`
}

// AttachmentDownloadRequest represents a file download request
type AttachmentDownloadRequest struct {
	AttachmentID string `json:"attachment_id"`
	TokenHash    string `json:"token_hash"`
	IPAddress    string `json:"ip_address,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
}

// AttachmentDownloadResponse represents the response to a file download request
type AttachmentDownloadResponse struct {
	Success       bool   `json:"success"`
	DownloadURL   string `json:"download_url,omitempty"`
	Filename      string `json:"filename,omitempty"`
	FileSize      int64  `json:"file_size,omitempty"`
	MimeType      string `json:"mime_type,omitempty"`
	Message       string `json:"message,omitempty"`
	Error         string `json:"error,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
}

// RichTextRequest represents a rich text content request
type RichTextRequest struct {
	LinkID      string `json:"link_id"`
	ReplyID     string `json:"reply_id,omitempty"`
	ContentType string `json:"content_type"` // 'email_body', 'reply_body'
	Content     string `json:"content"`
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// RichTextResponse represents the response to a rich text request
type RichTextResponse struct {
	Success      bool   `json:"success"`
	ContentID    string `json:"content_id,omitempty"`
	SanitizedContent string `json:"sanitized_content,omitempty"`
	FeaturesUsed string `json:"features_used,omitempty"`
	Message      string `json:"message,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorCode    string `json:"error_code,omitempty"`
}

// RichTextFeatures represents features used in rich text content
type RichTextFeatures struct {
	Bold          bool     `json:"bold"`
	Italic        bool     `json:"italic"`
	Underline     bool     `json:"underline"`
	Links         []string `json:"links,omitempty"`
	Lists         bool     `json:"lists"`
	Images        bool     `json:"images"`
	Tables        bool     `json:"tables"`
	CodeBlocks    bool     `json:"code_blocks"`
	Quotes        bool     `json:"quotes"`
	Headings      bool     `json:"headings"`
	Colors        bool     `json:"colors"`
	FontSizes     bool     `json:"font_sizes"`
}

// AttachmentMetadata represents metadata for attachments
type AttachmentMetadata struct {
	Watermark     string            `json:"watermark,omitempty"`
	Tags          []string          `json:"tags,omitempty"`
	Description   string            `json:"description,omitempty"`
	Category      string            `json:"category,omitempty"`
	Confidential  bool              `json:"confidential"`
	CustomFields  map[string]string `json:"custom_fields,omitempty"`
}

// VirusScanResult represents the result of a virus scan
type VirusScanResult struct {
	Status       string `json:"status"` // pending, clean, infected, error
	Result       string `json:"result,omitempty"`
	ScanEngine   string `json:"scan_engine,omitempty"`
	ScanVersion  string `json:"scan_version,omitempty"`
	Threats      []string `json:"threats,omitempty"`
	ScanTime     int64   `json:"scan_time,omitempty"` // milliseconds
}

// Helper methods for JSON serialization
func (f *RichTextFeatures) ToJSON() (string, error) {
	data, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (f *RichTextFeatures) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), f)
}

func (m *AttachmentMetadata) ToJSON() (string, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (m *AttachmentMetadata) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), m)
}

func (v *VirusScanResult) ToJSON() (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (v *VirusScanResult) FromJSON(data string) error {
	return json.Unmarshal([]byte(data), v)
}
