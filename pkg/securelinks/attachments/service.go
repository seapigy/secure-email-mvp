package attachments

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"secure-email-mvp/pkg/securelinks"
)

// =============================================================================
// SECURE ATTACHMENT SERVICE
// =============================================================================

// AttachmentService handles secure file attachments for external users
type AttachmentService struct {
	db           *sql.DB
	storagePath  string
	maxFileSize  int64 // in bytes
	allowedTypes []string
}

// SecureAttachment represents a secure attachment in the database
type SecureAttachment struct {
	ID             string     `json:"id" db:"id"`
	LinkID         string     `json:"link_id" db:"link_id"`
	FileName       string     `json:"file_name" db:"file_name"`
	OriginalName   string     `json:"original_name" db:"original_name"`
	FileSize       int64      `json:"file_size" db:"file_size"`
	MimeType       string     `json:"mime_type" db:"mime_type"`
	StoragePath    string     `json:"storage_path" db:"storage_path"`
	EncryptionKey  string     `json:"encryption_key" db:"encryption_key"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	Status         string     `json:"status" db:"status"` // "active", "expired", "deleted"
	AccessCount    int        `json:"access_count" db:"access_count"`
	MaxDownloads   *int       `json:"max_downloads,omitempty" db:"max_downloads"`
	DownloadExpiry *time.Time `json:"download_expiry,omitempty" db:"download_expiry"`
}

// AttachmentDownload represents a download record
type AttachmentDownload struct {
	ID           string    `json:"id" db:"id"`
	AttachmentID string    `json:"attachment_id" db:"attachment_id"`
	IPAddress    string    `json:"ip_address" db:"ip_address"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	DownloadedAt time.Time `json:"downloaded_at" db:"downloaded_at"`
	Success      bool      `json:"success" db:"success"`
	ErrorMessage *string   `json:"error_message,omitempty" db:"error_message"`
}

// UploadAttachmentRequest represents a request to upload an attachment
type UploadAttachmentRequest struct {
	LinkID       string                `json:"link_id" validate:"required"`
	File         *multipart.FileHeader `json:"file" validate:"required"`
	MaxDownloads *int                  `json:"max_downloads,omitempty"`
	ExpiresAt    *time.Time            `json:"expires_at,omitempty"`
}

// UploadAttachmentResponse represents the response to an upload request
type UploadAttachmentResponse struct {
	Success      bool   `json:"success"`
	AttachmentID string `json:"attachment_id,omitempty"`
	FileName     string `json:"file_name,omitempty"`
	FileSize     int64  `json:"file_size,omitempty"`
	SecureURL    string `json:"secure_url,omitempty"`
	Error        string `json:"error,omitempty"`
}

// DownloadAttachmentRequest represents a request to download an attachment
type DownloadAttachmentRequest struct {
	AttachmentID string `json:"attachment_id" validate:"required"`
	SessionToken string `json:"session_token" validate:"required"`
	IPAddress    string `json:"ip_address"`
	UserAgent    string `json:"user_agent"`
}

// DownloadAttachmentResponse represents the response to a download request
type DownloadAttachmentResponse struct {
	Success     bool   `json:"success"`
	FileName    string `json:"file_name,omitempty"`
	MimeType    string `json:"mime_type,omitempty"`
	FileSize    int64  `json:"file_size,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	Error       string `json:"error,omitempty"`
}

// NewAttachmentService creates a new attachment service
func NewAttachmentService(db *sql.DB, storagePath string) *AttachmentService {
	return &AttachmentService{
		db:          db,
		storagePath: storagePath,
		maxFileSize: 50 * 1024 * 1024, // 50MB
		allowedTypes: []string{
			"application/pdf",
			"application/msword",
			"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			"application/vnd.ms-excel",
			"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
			"text/plain",
			"image/jpeg",
			"image/png",
			"image/gif",
		},
	}
}

// UploadAttachment uploads and secures a file attachment
func (a *AttachmentService) UploadAttachment(ctx context.Context, req UploadAttachmentRequest) (*UploadAttachmentResponse, error) {
	// Validate that the link exists and is active
	link, err := a.getSecureLink(ctx, req.LinkID)
	if err != nil {
		return &UploadAttachmentResponse{
			Success: false,
			Error:   "Invalid or expired secure link",
		}, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Check if link is expired
	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		return &UploadAttachmentResponse{
			Success: false,
			Error:   "Secure link has expired",
		}, fmt.Errorf("secure link expired")
	}

	// Validate file
	if err := a.validateFile(req.File); err != nil {
		return &UploadAttachmentResponse{
			Success: false,
			Error:   "Invalid file",
		}, fmt.Errorf("file validation failed: %w", err)
	}

	// Generate attachment ID and storage path
	attachmentID := a.generateAttachmentID()
	storagePath := filepath.Join(a.storagePath, attachmentID)

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		return &UploadAttachmentResponse{
			Success: false,
			Error:   "Failed to create storage directory",
		}, fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Save file to storage
	if err := a.saveFile(req.File, storagePath); err != nil {
		return &UploadAttachmentResponse{
			Success: false,
			Error:   "Failed to save file",
		}, fmt.Errorf("failed to save file: %w", err)
	}

	// Generate encryption key (in production, use proper encryption)
	encryptionKey := a.generateEncryptionKey()

	// Create attachment record
	attachment := &SecureAttachment{
		ID:            attachmentID,
		LinkID:        req.LinkID,
		FileName:      attachmentID,
		OriginalName:  req.File.Filename,
		FileSize:      req.File.Size,
		MimeType:      req.File.Header.Get("Content-Type"),
		StoragePath:   storagePath,
		EncryptionKey: encryptionKey,
		CreatedAt:     time.Now(),
		ExpiresAt:     req.ExpiresAt,
		Status:        "active",
		AccessCount:   0,
		MaxDownloads:  req.MaxDownloads,
	}

	// Store attachment in database
	if err := a.storeAttachment(ctx, attachment); err != nil {
		// Clean up file if database storage fails
		os.Remove(storagePath)
		return &UploadAttachmentResponse{
			Success: false,
			Error:   "Failed to store attachment metadata",
		}, fmt.Errorf("failed to store attachment: %w", err)
	}

	// Generate secure download URL
	downloadURL := fmt.Sprintf("https://securemail.yourdomain.com/attachments/%s", attachmentID)

	return &UploadAttachmentResponse{
		Success:      true,
		AttachmentID: attachmentID,
		FileName:     req.File.Filename,
		FileSize:     req.File.Size,
		SecureURL:    downloadURL,
	}, nil
}

// DownloadAttachment handles secure attachment downloads
func (a *AttachmentService) DownloadAttachment(ctx context.Context, req DownloadAttachmentRequest) (*DownloadAttachmentResponse, error) {
	// Validate session token
	if err := a.validateSessionToken(ctx, req.SessionToken); err != nil {
		return &DownloadAttachmentResponse{
			Success: false,
			Error:   "Invalid session token",
		}, fmt.Errorf("invalid session token: %w", err)
	}

	// Get attachment
	attachment, err := a.getAttachment(ctx, req.AttachmentID)
	if err != nil {
		return &DownloadAttachmentResponse{
			Success: false,
			Error:   "Attachment not found",
		}, fmt.Errorf("failed to get attachment: %w", err)
	}

	// Check if attachment is active
	if attachment.Status != "active" {
		return &DownloadAttachmentResponse{
			Success: false,
			Error:   "Attachment is no longer available",
		}, fmt.Errorf("attachment not active: %s", attachment.Status)
	}

	// Check if attachment is expired
	if attachment.ExpiresAt != nil && time.Now().After(*attachment.ExpiresAt) {
		return &DownloadAttachmentResponse{
			Success: false,
			Error:   "Attachment has expired",
		}, fmt.Errorf("attachment expired")
	}

	// Check download limits
	if attachment.MaxDownloads != nil && attachment.AccessCount >= *attachment.MaxDownloads {
		return &DownloadAttachmentResponse{
			Success: false,
			Error:   "Download limit exceeded",
		}, fmt.Errorf("download limit exceeded")
	}

	// Check if file exists
	if _, err := os.Stat(attachment.StoragePath); os.IsNotExist(err) {
		return &DownloadAttachmentResponse{
			Success: false,
			Error:   "File not found",
		}, fmt.Errorf("file not found: %s", attachment.StoragePath)
	}

	// Record download attempt
	download := &AttachmentDownload{
		ID:           a.generateDownloadID(),
		AttachmentID: req.AttachmentID,
		IPAddress:    req.IPAddress,
		UserAgent:    req.UserAgent,
		DownloadedAt: time.Now(),
		Success:      true,
	}

	if err := a.recordDownload(ctx, download); err != nil {
		log.Printf("Warning: Failed to record download: %v", err)
	}

	// Update access count
	if err := a.updateAccessCount(ctx, req.AttachmentID); err != nil {
		log.Printf("Warning: Failed to update access count: %v", err)
	}

	// Generate temporary download URL
	downloadURL := fmt.Sprintf("https://securemail.yourdomain.com/attachments/download/%s?token=%s",
		req.AttachmentID, req.SessionToken)

	return &DownloadAttachmentResponse{
		Success:     true,
		FileName:    attachment.OriginalName,
		MimeType:    attachment.MimeType,
		FileSize:    attachment.FileSize,
		DownloadURL: downloadURL,
	}, nil
}

// GetAttachmentInfo gets attachment information without downloading
func (a *AttachmentService) GetAttachmentInfo(ctx context.Context, attachmentID string) (*SecureAttachment, error) {
	return a.getAttachment(ctx, attachmentID)
}

// DeleteAttachment marks an attachment as deleted
func (a *AttachmentService) DeleteAttachment(ctx context.Context, attachmentID string) error {
	query := `UPDATE secure_attachments SET status = 'deleted' WHERE id = ?`
	_, err := a.db.ExecContext(ctx, query, attachmentID)
	return err
}

// Helper methods

// validateFile validates the uploaded file
func (a *AttachmentService) validateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > a.maxFileSize {
		return fmt.Errorf("file too large (max %d bytes)", a.maxFileSize)
	}

	// Check file type
	mimeType := file.Header.Get("Content-Type")
	allowed := false
	for _, allowedType := range a.allowedTypes {
		if mimeType == allowedType {
			allowed = true
			break
		}
	}
	if !allowed {
		return fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	return nil
}

// saveFile saves the uploaded file to storage
func (a *AttachmentService) saveFile(file *multipart.FileHeader, storagePath string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(storagePath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

// getSecureLink retrieves a secure link from the database
func (a *AttachmentService) getSecureLink(ctx context.Context, linkID string) (*securelinks.SecureLink, error) {
	query := `
		SELECT link_id, email_id, recipient_email, created_at, expires_at, status
		FROM secure_links
		WHERE link_id = ?
	`

	var link securelinks.SecureLink
	err := a.db.QueryRowContext(ctx, query, linkID).Scan(
		&link.LinkID, &link.EmailID, &link.RecipientEmail, &link.CreatedAt,
		&link.ExpiresAt, &link.Status,
	)
	if err != nil {
		return nil, err
	}

	return &link, nil
}

// validateSessionToken validates a session token
func (a *AttachmentService) validateSessionToken(ctx context.Context, sessionToken string) error {
	query := `
		SELECT id FROM link_view_sessions 
		WHERE session_token = ? AND is_active = 1 AND expires_at > CURRENT_TIMESTAMP
	`

	var sessionID string
	err := a.db.QueryRowContext(ctx, query, sessionToken).Scan(&sessionID)
	return err
}

// getAttachment retrieves an attachment from the database
func (a *AttachmentService) getAttachment(ctx context.Context, attachmentID string) (*SecureAttachment, error) {
	query := `
		SELECT id, link_id, file_name, original_name, file_size, mime_type,
		       storage_path, encryption_key, created_at, expires_at, status,
		       access_count, max_downloads, download_expiry
		FROM secure_attachments
		WHERE id = ?
	`

	var attachment SecureAttachment
	err := a.db.QueryRowContext(ctx, query, attachmentID).Scan(
		&attachment.ID, &attachment.LinkID, &attachment.FileName, &attachment.OriginalName,
		&attachment.FileSize, &attachment.MimeType, &attachment.StoragePath, &attachment.EncryptionKey,
		&attachment.CreatedAt, &attachment.ExpiresAt, &attachment.Status, &attachment.AccessCount,
		&attachment.MaxDownloads, &attachment.DownloadExpiry,
	)
	if err != nil {
		return nil, err
	}

	return &attachment, nil
}

// storeAttachment stores an attachment in the database
func (a *AttachmentService) storeAttachment(ctx context.Context, attachment *SecureAttachment) error {
	query := `
		INSERT INTO secure_attachments (
			id, link_id, file_name, original_name, file_size, mime_type,
			storage_path, encryption_key, created_at, expires_at, status,
			access_count, max_downloads, download_expiry
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := a.db.ExecContext(ctx, query,
		attachment.ID, attachment.LinkID, attachment.FileName, attachment.OriginalName,
		attachment.FileSize, attachment.MimeType, attachment.StoragePath, attachment.EncryptionKey,
		attachment.CreatedAt, attachment.ExpiresAt, attachment.Status, attachment.AccessCount,
		attachment.MaxDownloads, attachment.DownloadExpiry,
	)

	return err
}

// recordDownload records a download attempt
func (a *AttachmentService) recordDownload(ctx context.Context, download *AttachmentDownload) error {
	query := `
		INSERT INTO attachment_downloads (
			id, attachment_id, ip_address, user_agent, downloaded_at, success, error_message
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := a.db.ExecContext(ctx, query,
		download.ID, download.AttachmentID, download.IPAddress, download.UserAgent,
		download.DownloadedAt, download.Success, download.ErrorMessage,
	)

	return err
}

// updateAccessCount updates the access count for an attachment
func (a *AttachmentService) updateAccessCount(ctx context.Context, attachmentID string) error {
	query := `UPDATE secure_attachments SET access_count = access_count + 1 WHERE id = ?`
	_, err := a.db.ExecContext(ctx, query, attachmentID)
	return err
}

// generateAttachmentID generates a unique attachment ID
func (a *AttachmentService) generateAttachmentID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateDownloadID generates a unique download ID
func (a *AttachmentService) generateDownloadID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}

// generateEncryptionKey generates an encryption key
func (a *AttachmentService) generateEncryptionKey() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)
}
