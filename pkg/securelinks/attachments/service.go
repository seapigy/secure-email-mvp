package attachments

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"time"

	"secure-email-mvp/pkg/models"
)

// Service handles secure file attachments
type Service struct {
	db            Database
	s3Client      S3Client
	virusScanner  VirusScanner
	config        *Config
}

// Database interface for attachment operations
type Database interface {
	GetAllowedFileType(mimeType string) (*models.AllowedFileType, error)
	CreateAttachment(attachment *models.SecureAttachment) error
	GetAttachment(attachmentID string) (*models.SecureAttachment, error)
	CreateDownloadToken(token *models.AttachmentDownloadToken) error
	GetDownloadToken(tokenHash string) (*models.AttachmentDownloadToken, error)
	UpdateDownloadCount(attachmentID string) error
	UpdateDownloadTokenCount(tokenID string) error
	LogAuditEvent(event *models.RichMessagingAuditLog) error
}

// S3Client interface for S3 operations
type S3Client interface {
	GeneratePresignedUploadURL(bucket, key string, expires time.Duration) (string, error)
	GeneratePresignedDownloadURL(bucket, key string, expires time.Duration) (string, error)
	UploadFile(bucket, key string, file multipart.File) error
}

// VirusScanner interface for virus scanning
type VirusScanner interface {
	ScanFile(file multipart.File) (*models.VirusScanResult, error)
}

// Config holds service configuration
type Config struct {
	MaxFileSize     int64
	DefaultBucket   string
	TokenExpiry     time.Duration
	MaxDownloads    int
	AllowedDomains  []string
}

// NewService creates a new attachment service
func NewService(db Database, s3Client S3Client, virusScanner VirusScanner, config *Config) *Service {
	return &Service{
		db:           db,
		s3Client:     s3Client,
		virusScanner: virusScanner,
		config:       config,
	}
}

// ProcessUpload handles file upload validation and preparation
func (s *Service) ProcessUpload(ctx context.Context, req *models.AttachmentUploadRequest) (*models.AttachmentUploadResponse, error) {
	// Validate file type
	allowedType, err := s.db.GetAllowedFileType(req.MimeType)
	if err != nil {
		return &models.AttachmentUploadResponse{
			Success:   false,
			Error:     "File type not allowed",
			ErrorCode: "INVALID_FILE_TYPE",
		}, nil
	}

	// Check file size
	if req.FileSize > allowedType.MaxSize {
		return &models.AttachmentUploadResponse{
			Success:   false,
			Error:     fmt.Sprintf("File too large. Max size: %d bytes", allowedType.MaxSize),
			ErrorCode: "FILE_TOO_LARGE",
		}, nil
	}

	// Generate attachment ID and S3 key
	attachmentID := s.generateAttachmentID()
	s3Key := s.generateS3Key(attachmentID, req.Filename)

	// Create attachment record
	attachment := &models.SecureAttachment{
		AttachmentID:     attachmentID,
		LinkID:           req.LinkID,
		ReplyID:          &req.ReplyID,
		Filename:         req.Filename,
		OriginalFilename: req.Filename,
		FileSize:         req.FileSize,
		MimeType:         req.MimeType,
		FileHash:         req.FileHash,
		S3Key:            s3Key,
		S3Bucket:         s.config.DefaultBucket,
		VirusScanStatus:  "pending",
		MaxDownloads:     s.config.MaxDownloads,
		Status:           "active",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
		CreatedBy:        &req.IPAddress, // For external uploads, use IP as creator
	}

	// Store attachment record
	if err := s.db.CreateAttachment(attachment); err != nil {
		return &models.AttachmentUploadResponse{
			Success:   false,
			Error:     "Failed to create attachment record",
			ErrorCode: "DATABASE_ERROR",
		}, nil
	}

	// Generate presigned upload URL
	uploadURL, err := s.s3Client.GeneratePresignedUploadURL(
		s.config.DefaultBucket,
		s3Key,
		time.Hour, // 1 hour expiry
	)
	if err != nil {
		return &models.AttachmentUploadResponse{
			Success:   false,
			Error:     "Failed to generate upload URL",
			ErrorCode: "S3_ERROR",
		}, nil
	}

	// Log audit event
	s.logAuditEvent(&models.RichMessagingAuditLog{
		EventType:    "attachment_upload_initiated",
		LinkID:       &req.LinkID,
		ReplyID:      &req.ReplyID,
		AttachmentID: &attachmentID,
		IPAddress:    &req.IPAddress,
		UserAgent:    &req.UserAgent,
		FileSize:     &req.FileSize,
		MimeType:     &req.MimeType,
		EventDetails: s.createEventDetails("upload_initiated", map[string]interface{}{
			"filename": req.Filename,
			"s3_key":   s3Key,
		}),
		CreatedAt: time.Now(),
	})

	return &models.AttachmentUploadResponse{
		Success:      true,
		AttachmentID: attachmentID,
		UploadURL:    uploadURL,
		Message:      "Upload URL generated successfully",
	}, nil
}

// ProcessFileUpload handles the actual file upload and virus scanning
func (s *Service) ProcessFileUpload(ctx context.Context, attachmentID string, file multipart.File, header *multipart.FileHeader) error {
	// Get attachment record
	attachment, err := s.db.GetAttachment(attachmentID)
	if err != nil {
		return fmt.Errorf("attachment not found: %w", err)
	}

	// Upload file to S3
	if err := s.s3Client.UploadFile(attachment.S3Bucket, attachment.S3Key, file); err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Reset file pointer for virus scanning
	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset file pointer: %w", err)
	}

	// Virus scan the file
	scanResult, err := s.virusScanner.ScanFile(file)
	if err != nil {
		// Log scan error but don't fail the upload
		scanResult = &models.VirusScanResult{
			Status: "error",
			Result: fmt.Sprintf("Scan failed: %v", err),
		}
	}

	// Update attachment with scan results
	scanResultJSON, _ := scanResult.ToJSON()
	attachment.VirusScanStatus = scanResult.Status
	attachment.VirusScanResult = &scanResultJSON
	attachment.VirusScanTimestamp = &time.Time{}
	*attachment.VirusScanTimestamp = time.Now()

	// Update attachment record
	if err := s.db.CreateAttachment(attachment); err != nil {
		return fmt.Errorf("failed to update attachment: %w", err)
	}

	// Log audit event
	linkID := attachment.LinkID
	s.logAuditEvent(&models.RichMessagingAuditLog{
		EventType:    "attachment_upload_completed",
		LinkID:       &linkID,
		ReplyID:      attachment.ReplyID,
		AttachmentID: &attachmentID,
		FileSize:     &attachment.FileSize,
		MimeType:     &attachment.MimeType,
		VirusScanResult: &scanResultJSON,
		EventDetails: s.createEventDetails("upload_completed", map[string]interface{}{
			"filename":        attachment.Filename,
			"virus_scan_status": scanResult.Status,
		}),
		CreatedAt: time.Now(),
	})

	return nil
}

// GenerateDownloadToken creates a secure download token for an attachment
func (s *Service) GenerateDownloadToken(ctx context.Context, attachmentID, ipAddress, userAgent string) (*models.AttachmentDownloadToken, error) {
	// Get attachment
	attachment, err := s.db.GetAttachment(attachmentID)
	if err != nil {
		return nil, fmt.Errorf("attachment not found: %w", err)
	}

	// Check if attachment is active
	if attachment.Status != "active" {
		return nil, fmt.Errorf("attachment is not available")
	}

	// Check virus scan status
	if attachment.VirusScanStatus == "infected" {
		return nil, fmt.Errorf("attachment contains malware")
	}

	// Check download limit
	if attachment.DownloadCount >= attachment.MaxDownloads {
		return nil, fmt.Errorf("download limit exceeded")
	}

	// Generate token
	tokenHash := s.generateTokenHash()
	token := &models.AttachmentDownloadToken{
		TokenID:       s.generateTokenID(),
		AttachmentID:  attachmentID,
		TokenHash:     tokenHash,
		ExpiresAt:     time.Now().Add(s.config.TokenExpiry),
		MaxDownloads:  1, // One-time use token
		DownloadCount: 0,
		IPAddress:     &ipAddress,
		UserAgent:     &userAgent,
		CreatedAt:     time.Now(),
	}

	// Store token
	if err := s.db.CreateDownloadToken(token); err != nil {
		return nil, fmt.Errorf("failed to create download token: %w", err)
	}

	// Log audit event
	linkID := attachment.LinkID
	s.logAuditEvent(&models.RichMessagingAuditLog{
		EventType:       "download_token_created",
		LinkID:          &linkID,
		ReplyID:         attachment.ReplyID,
		AttachmentID:    &attachmentID,
		IPAddress:       &ipAddress,
		UserAgent:       &userAgent,
		DownloadTokenID: &token.TokenID,
		EventDetails:    s.createEventDetails("token_created", map[string]interface{}{
			"filename": attachment.Filename,
			"expires_at": token.ExpiresAt,
		}),
		CreatedAt: time.Now(),
	})

	return token, nil
}

// ProcessDownload handles file download with token validation
func (s *Service) ProcessDownload(ctx context.Context, req *models.AttachmentDownloadRequest) (*models.AttachmentDownloadResponse, error) {
	// Validate token
	token, err := s.db.GetDownloadToken(req.TokenHash)
	if err != nil {
		return &models.AttachmentDownloadResponse{
			Success:   false,
			Error:     "Invalid download token",
			ErrorCode: "INVALID_TOKEN",
		}, nil
	}

	// Check token expiry
	if time.Now().After(token.ExpiresAt) {
		return &models.AttachmentDownloadResponse{
			Success:   false,
			Error:     "Download token expired",
			ErrorCode: "TOKEN_EXPIRED",
		}, nil
	}

	// Check download count
	if token.DownloadCount >= token.MaxDownloads {
		return &models.AttachmentDownloadResponse{
			Success:   false,
			Error:     "Download limit exceeded",
			ErrorCode: "DOWNLOAD_LIMIT_EXCEEDED",
		}, nil
	}

	// Get attachment
	attachment, err := s.db.GetAttachment(token.AttachmentID)
	if err != nil {
		return &models.AttachmentDownloadResponse{
			Success:   false,
			Error:     "Attachment not found",
			ErrorCode: "ATTACHMENT_NOT_FOUND",
		}, nil
	}

	// Check attachment status
	if attachment.Status != "active" {
		return &models.AttachmentDownloadResponse{
			Success:   false,
			Error:     "Attachment not available",
			ErrorCode: "ATTACHMENT_UNAVAILABLE",
		}, nil
	}

	// Generate download URL
	downloadURL, err := s.s3Client.GeneratePresignedDownloadURL(
		attachment.S3Bucket,
		attachment.S3Key,
		time.Hour, // 1 hour expiry
	)
	if err != nil {
		return &models.AttachmentDownloadResponse{
			Success:   false,
			Error:     "Failed to generate download URL",
			ErrorCode: "S3_ERROR",
		}, nil
	}

	// Update download counts
	s.db.UpdateDownloadCount(token.AttachmentID)
	s.db.UpdateDownloadTokenCount(token.TokenID)

	// Log audit event
	linkID := attachment.LinkID
	s.logAuditEvent(&models.RichMessagingAuditLog{
		EventType:       "attachment_download",
		LinkID:          &linkID,
		ReplyID:         attachment.ReplyID,
		AttachmentID:    &token.AttachmentID,
		IPAddress:       &req.IPAddress,
		UserAgent:       &req.UserAgent,
		DownloadTokenID: &token.TokenID,
		FileSize:        &attachment.FileSize,
		MimeType:        &attachment.MimeType,
		EventDetails:    s.createEventDetails("download", map[string]interface{}{
			"filename": attachment.Filename,
		}),
		CreatedAt: time.Now(),
	})

	return &models.AttachmentDownloadResponse{
		Success:     true,
		DownloadURL: downloadURL,
		Filename:    attachment.OriginalFilename,
		FileSize:    attachment.FileSize,
		MimeType:    attachment.MimeType,
		Message:     "Download URL generated successfully",
	}, nil
}

// Helper methods

func (s *Service) generateAttachmentID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("att_%s", hex.EncodeToString(bytes))
}

func (s *Service) generateS3Key(attachmentID, filename string) string {
	ext := filepath.Ext(filename)
	return fmt.Sprintf("attachments/%s/%s%s", time.Now().Format("2006/01/02"), attachmentID, ext)
}

func (s *Service) generateTokenHash() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (s *Service) generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("token_%s", hex.EncodeToString(bytes))
}

func (s *Service) createEventDetails(eventType string, data map[string]interface{}) *string {
	details := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Unix(),
		"data":       data,
	}
	
	jsonData, _ := json.Marshal(details)
	result := string(jsonData)
	return &result
}

func (s *Service) logAuditEvent(event *models.RichMessagingAuditLog) {
	// Generate audit ID
	bytes := make([]byte, 16)
	rand.Read(bytes)
	event.AuditID = fmt.Sprintf("audit_%s", hex.EncodeToString(bytes))
	
	// Log asynchronously to avoid blocking
	go func() {
		if err := s.db.LogAuditEvent(event); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Failed to log audit event: %v\n", err)
		}
	}()
}
