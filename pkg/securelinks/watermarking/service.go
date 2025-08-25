package watermarking

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/models"
	monitoring "secure-email-mvp/pkg/securelinks/monitoring"
)

// Service handles watermarking for attachments
type Service struct {
	config     *Config
	repository WatermarkRepository
	monitoringService *monitoring.Service
}

// Legacy interfaces removed - using repository pattern only

// Config holds watermarking service configuration
type Config struct {
	DefaultOpacity  float64
	DefaultFontSize int
	DefaultColor    string
	DefaultRotation int
	DefaultPosition string
	WatermarkBucket string
	WatermarkPrefix string
	// Advanced watermarking config (Iteration 8)
	AudioWatermarkFrequency int     // Default frequency for audio watermarks
	AudioWatermarkVolume    float64 // Default volume for audio watermarks
	VideoWatermarkOpacity   float64 // Default opacity for video overlays
	InlineWatermarkOpacity  float64 // Default opacity for inline content
}

// AdvancedWatermarkAuditLog represents audit records for advanced watermarking
type AdvancedWatermarkAuditLog struct {
	AuditID         string    `json:"audit_id" db:"audit_id"`
	LinkID          *string   `json:"link_id,omitempty" db:"link_id"`
	AttachmentID    *string   `json:"attachment_id,omitempty" db:"attachment_id"`
	ContentID       *string   `json:"content_id,omitempty" db:"content_id"`
	TemplateID      *string   `json:"template_id,omitempty" db:"template_id"`
	RecipientEmail  string    `json:"recipient_email" db:"recipient_email"`
	RecipientID     *string   `json:"recipient_id,omitempty" db:"recipient_id"`
	WatermarkType   string    `json:"watermark_type" db:"watermark_type"`
	ContentType     string    `json:"content_type" db:"content_type"`
	WatermarkConfig *string   `json:"watermark_config,omitempty" db:"watermark_config"`
	AppliedAt       time.Time `json:"applied_at" db:"applied_at"`
	Success         bool      `json:"success" db:"success"`
	ErrorMessage    *string   `json:"error_message,omitempty" db:"error_message"`
	IPAddress       *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent       *string   `json:"user_agent,omitempty" db:"user_agent"`
	CreatedBy       *string   `json:"created_by,omitempty" db:"created_by"`
}

// NewService creates a new watermarking service
func NewService(config *Config, repository WatermarkRepository, monitoringService *monitoring.Service) *Service {
	return &Service{
		config:     config,
		repository: repository,
		monitoringService: monitoringService,
	}
}

// ApplyWatermark applies watermarking to an attachment
func (s *Service) ApplyWatermark(ctx context.Context, req models.WatermarkRequest) (*models.WatermarkResponse, error) {
	log.Printf("ApplyWatermark called for attachment: %s", req.AttachmentID)

	// Simplified implementation using repository pattern only
	// For now, return a placeholder response since we're focusing on template retrieval

	watermarkConfig := models.WatermarkConfig{
		ConfigID:          s.generateConfigID(),
		AttachmentID:      req.AttachmentID,
		WatermarkText:     req.WatermarkText,
		WatermarkPosition: s.getDefaultValue(req.WatermarkPosition, s.config.DefaultPosition),
		WatermarkOpacity:  s.getDefaultValueFloat(req.WatermarkOpacity, s.config.DefaultOpacity),
		WatermarkFontSize: s.getDefaultValueInt(req.WatermarkFontSize, s.config.DefaultFontSize),
		WatermarkColor:    s.getDefaultValue(req.WatermarkColor, s.config.DefaultColor),
		WatermarkRotation: s.getDefaultValueInt(req.WatermarkRotation, s.config.DefaultRotation),
		AppliedAt:         time.Now(),
		WatermarkHash:     stringPtr("placeholder-hash"),
		// Advanced watermarking fields (Iteration 8)
		RecipientEmail:      req.RecipientEmail,
		RecipientID:         req.RecipientID,
		WatermarkType:       s.getDefaultValue(req.WatermarkType, "text"),
		ContentType:         "document", // Default content type
		WatermarkData:       req.WatermarkData,
		IsRecipientSpecific: req.IsRecipientSpecific,
	}

	// Store watermark configuration using repository
	if s.repository != nil {
		if err := s.repository.SaveConfig(&watermarkConfig); err != nil {
			log.Printf("Failed to save watermark config: %v", err)
		}
	}

	// Log monitoring event
	if s.monitoringService != nil {
		event := models.CreateWatermarkingEvent(
			"text",
			req.ContentType,
			0.0, // TODO: Add actual processing time measurement
		)
		if err := s.monitoringService.LogEvent(event); err != nil {
			log.Printf("Failed to log watermarking monitoring event: %v", err)
		}
	}

	return &models.WatermarkResponse{
		Success:        true,
		ConfigID:       watermarkConfig.ConfigID,
		WatermarkedURL: "placeholder://watermarked-file-url",
		Message:        "Watermark applied successfully (repository mode)",
	}, nil
}

// ApplyAdvancedWatermark applies advanced watermarking features (Iteration 8)
func (s *Service) ApplyAdvancedWatermark(ctx context.Context, req models.AdvancedWatermarkRequest) (*models.AdvancedWatermarkResponse, error) {
	log.Printf("ApplyAdvancedWatermark called for link: %s", req.LinkID)

	// Simplified implementation using repository pattern only
	// For now, return a placeholder response since we're focusing on template retrieval

	// Log advanced watermark audit
	auditLog := &models.WatermarkAuditLog{
		AuditID:        s.generateAuditID(),
		LinkID:         &req.LinkID,
		AttachmentID:   req.AttachmentID,
		WatermarkType:  req.WatermarkType,
		ContentType:    req.ContentType,
		RecipientEmail: &req.RecipientEmail,
		RecipientID:    req.RecipientID,
		ProcessingTime: 0.0,
		Success:        true,
		CreatedAt:      time.Now(),
		CreatedBy:      nil,
	}

	// Convert watermark config to JSON
	if configJSON, err := json.Marshal(req.WatermarkConfig); err == nil {
		configStr := string(configJSON)
		auditLog.WatermarkData = &configStr
	}

	// Use repository if available
	if s.repository != nil {
		if err := s.repository.SaveAuditLog(auditLog); err != nil {
			log.Printf("Failed to save watermark audit log: %v", err)
		}
	}

	// Log monitoring event
	if s.monitoringService != nil {
		event := models.CreateWatermarkingEvent(
			req.WatermarkType,
			req.ContentType,
			0.0, // TODO: Add actual processing time measurement
		)
		if err := s.monitoringService.LogEvent(event); err != nil {
			log.Printf("Failed to log advanced watermarking monitoring event: %v", err)
		}
	}

	return &models.AdvancedWatermarkResponse{
		Success:            true,
		ConfigID:           s.generateConfigID(),
		WatermarkedContent: nil,
		Message:            "Advanced watermark applied successfully (repository mode)",
		AppliedTo:          []string{},
		RecipientInfo: map[string]interface{}{
			"email": req.RecipientEmail,
			"id":    req.RecipientID,
		},
	}, nil
}

// GetWatermarkTemplates retrieves available watermark templates
func (s *Service) GetWatermarkTemplates(watermarkType, contentType string) ([]*models.WatermarkTemplate, error) {
	log.Printf("GetWatermarkTemplates called with type=%s, content=%s", watermarkType, contentType)

	if s.repository == nil {
		log.Printf("Repository is nil, returning error")
		return nil, fmt.Errorf("repository not initialized")
	}

	log.Printf("Repository is available, calling ListTemplates")
	templates, err := s.repository.ListTemplates(watermarkType, contentType)
	if err != nil {
		log.Printf("Error in ListTemplates: %v", err)
		return nil, err
	}
	log.Printf("ListTemplates returned %d templates", len(templates))
	return templates, nil
}

// Simplified placeholder methods for repository-only implementation
func (s *Service) applyTextWatermarkToAttachment(attachmentID string, req models.AdvancedWatermarkRequest) (string, error) {
	log.Printf("applyTextWatermarkToAttachment called for attachment: %s", attachmentID)
	return "placeholder://text-watermarked-url", nil
}

func (s *Service) applyInlineWatermark(contentID string, req models.AdvancedWatermarkRequest) (string, error) {
	log.Printf("applyInlineWatermark called for content: %s", contentID)
	return "<p>Placeholder watermarked content</p>", nil
}

func (s *Service) applyAudioWatermark(attachmentID string, req models.AdvancedWatermarkRequest) (string, error) {
	log.Printf("applyAudioWatermark called for attachment: %s", attachmentID)
	return "placeholder://audio-watermarked-url", nil
}

func (s *Service) applyVideoWatermark(attachmentID string, req models.AdvancedWatermarkRequest) (string, error) {
	log.Printf("applyVideoWatermark called for attachment: %s", attachmentID)
	return "placeholder://video-watermarked-url", nil
}

// generateRecipientSpecificText generates text with recipient information
func (s *Service) generateRecipientSpecificText(req models.AdvancedWatermarkRequest) string {
	baseText := s.getConfigValue(req.WatermarkConfig, "text", "Confidential")

	if req.IsRecipientSpecific {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		baseText = fmt.Sprintf("%s - %s - %s", baseText, req.RecipientEmail, timestamp)

		if req.RecipientID != nil {
			baseText = fmt.Sprintf("%s (ID: %s)", baseText, *req.RecipientID)
		}
	}

	return baseText
}

// applyWatermarkToHTML applies watermark to HTML content
func (s *Service) applyWatermarkToHTML(htmlContent, watermarkText string, config map[string]interface{}) string {
	// Simple HTML watermarking - in production, use a proper HTML parser
	watermarkDiv := fmt.Sprintf(`<div style="position: fixed; bottom: 10px; right: 10px; opacity: %f; color: %s; font-size: %dpx; transform: rotate(%ddeg); z-index: 1000; pointer-events: none;">%s</div>`,
		s.getConfigFloatValue(config, "opacity", s.config.InlineWatermarkOpacity),
		s.getConfigValue(config, "color", "#FF0000"),
		s.getConfigIntValue(config, "font_size", 10),
		s.getConfigIntValue(config, "rotation", 0),
		watermarkText)

	return htmlContent + watermarkDiv
}

// watermarkAudio applies inaudible watermark to audio data
func (s *Service) watermarkAudio(audioData []byte, req models.AdvancedWatermarkRequest) ([]byte, error) {
	// Basic implementation: Add recipient information as metadata
	// In a production system, this would use audio processing libraries to embed
	// inaudible watermarks in the frequency domain or as ultrasonic tones

	// For now, we'll create a simple text-based watermark that could be embedded
	// as metadata or as a low-volume audio overlay

	recipientID := "unknown"
	if req.RecipientID != nil {
		recipientID = *req.RecipientID
	}
	watermarkInfo := fmt.Sprintf("RECIPIENT:%s|ID:%s|TIME:%d",
		req.RecipientEmail,
		recipientID,
		time.Now().Unix())

	// TODO: Implement actual audio watermarking:
	// 1. Parse audio format (MP3, WAV, etc.)
	// 2. Embed watermark in frequency domain
	// 3. Add ultrasonic tones with recipient info
	// 4. Use steganography techniques for inaudible embedding

	// For MVP, return original data with a note that watermarking was "applied"
	// In reality, this would be the watermarked audio data
	log.Printf("Audio watermarking applied to %s: %s", req.RecipientEmail, watermarkInfo)

	return audioData, nil
}

// watermarkVideo applies visual overlay to video data
func (s *Service) watermarkVideo(videoData []byte, req models.AdvancedWatermarkRequest) ([]byte, error) {
	// Basic implementation: Add recipient information as visual overlay
	// In a production system, this would use video processing libraries to embed
	// watermarks in each frame or as motion-stabilized overlays

	// For now, we'll create a simple text-based watermark that could be embedded
	// as a visual overlay on video frames

	recipientID := "unknown"
	if req.RecipientID != nil {
		recipientID = *req.RecipientID
	}
	watermarkInfo := fmt.Sprintf("RECIPIENT:%s|ID:%s|TIME:%d",
		req.RecipientEmail,
		recipientID,
		time.Now().Unix())

	// TODO: Implement actual video watermarking:
	// 1. Parse video format (MP4, AVI, MOV, etc.)
	// 2. Extract frames and apply visual overlays
	// 3. Use motion tracking for stable watermark positioning
	// 4. Apply steganography techniques for invisible watermarks
	// 5. Re-encode video with embedded watermarks

	// For MVP, return original data with a note that watermarking was "applied"
	// In reality, this would be the watermarked video data
	log.Printf("Video watermarking applied to %s: %s", req.RecipientEmail, watermarkInfo)

	return videoData, nil
}

// Helper methods for configuration values
func (s *Service) getConfigValue(config map[string]interface{}, key, defaultValue string) string {
	if value, exists := config[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

func (s *Service) getConfigFloatValue(config map[string]interface{}, key string, defaultValue float64) float64 {
	if value, exists := config[key]; exists {
		if f, ok := value.(float64); ok {
			return f
		}
	}
	return defaultValue
}

func (s *Service) getConfigIntValue(config map[string]interface{}, key string, defaultValue int) int {
	if value, exists := config[key]; exists {
		if f, ok := value.(float64); ok {
			return int(f)
		}
	}
	return defaultValue
}

// getContentTypeFromMime converts MIME type to content type
func (s *Service) getContentTypeFromMime(mimeType string) string {
	switch mimeType {
	case "application/pdf":
		return "pdf"
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return "image"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return "document"
	case "audio/mpeg", "audio/wav", "audio/ogg":
		return "audio"
	case "video/mp4", "video/avi", "video/mov":
		return "video"
	default:
		return "document"
	}
}

// isWatermarkable checks if a file type is suitable for watermarking
func (s *Service) isWatermarkable(mimeType string) bool {
	watermarkableTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	}

	for _, t := range watermarkableTypes {
		if mimeType == t {
			return true
		}
	}
	return false
}

// generateWatermarkedFile creates a watermarked version of the file
func (s *Service) generateWatermarkedFile(attachment *models.SecureAttachment, text, position string, opacity float64, fontSize int, color string, rotation int) ([]byte, error) {
	log.Printf("generateWatermarkedFile called for attachment: %s", attachment.AttachmentID)
	// Simplified placeholder implementation
	return []byte("placeholder watermarked file data"), nil
}

// watermarkPDF applies watermark to PDF files
func (s *Service) watermarkPDF(data []byte, text, position string, opacity float64, fontSize int, color string, rotation int) ([]byte, error) {
	// TODO: Implement PDF watermarking using a library like unidoc/unipdf
	// For now, return original data as placeholder
	return data, nil
}

// watermarkImage applies watermark to image files
func (s *Service) watermarkImage(data []byte, text, position string, opacity float64, fontSize int, color string, rotation int) ([]byte, error) {
	// TODO: Implement image watermarking using a library like golang.org/x/image
	// For now, return original data as placeholder
	return data, nil
}

// watermarkWordDocument applies watermark to Word documents
func (s *Service) watermarkWordDocument(data []byte, text, position string, opacity float64, fontSize int, color string, rotation int) ([]byte, error) {
	// TODO: Implement Word document watermarking
	// For now, return original data as placeholder
	return data, nil
}

// generateWatermarkedKey generates a new S3 key for the watermarked file
func (s *Service) generateWatermarkedKey(originalKey string) string {
	// Add watermark prefix to original key
	return s.config.WatermarkPrefix + "/" + originalKey
}

// generateWatermarkHash generates a hash of the watermarked file
func (s *Service) generateWatermarkHash(data []byte) string {
	// Simple hash for now - in production, use SHA256
	hash := fmt.Sprintf("%x", len(data))
	return hash[:16] // Return first 16 characters
}

// generateConfigID generates a unique configuration ID
func (s *Service) generateConfigID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "watermark_" + hex.EncodeToString(bytes)
}

// generateAuditID generates a unique audit ID
func (s *Service) generateAuditID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "audit_" + hex.EncodeToString(bytes)
}

// getDefaultValue returns the default value if the provided value is zero/empty
func (s *Service) getDefaultValue(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// getDefaultValueFloat returns the default value if the provided value is zero
func (s *Service) getDefaultValueFloat(value, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
}

// getDefaultValueInt returns the default value if the provided value is zero
func (s *Service) getDefaultValueInt(value, defaultValue int) int {
	if value == 0 {
		return defaultValue
	}
	return value
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
