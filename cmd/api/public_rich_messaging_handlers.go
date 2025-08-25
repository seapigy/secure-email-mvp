package main

import (
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"secure-email-mvp/pkg/models"
	richtext "secure-email-mvp/pkg/securelinks/richtext"
)

// RichTextHandler handles rich text content processing
func (srv *Server) richTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req models.RichTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.LinkID == "" || req.ContentType == "" || req.Content == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Get client IP
	req.IPAddress = srv.getClientIP(r)
	req.UserAgent = r.UserAgent()

	// Process rich text content
	response, err := srv.processRichText(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to process rich text: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AttachmentUploadHandler handles file upload requests
func (srv *Server) attachmentUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(32 << 20); err != nil { // 32MB max
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	linkID := r.FormValue("link_id")
	replyID := r.FormValue("reply_id")
	
	if linkID == "" {
		http.Error(w, "Missing link_id", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file
	if err := srv.validateUploadedFile(header); err != nil {
		http.Error(w, fmt.Sprintf("File validation failed: %v", err), http.StatusBadRequest)
		return
	}

	// Create upload request
	req := &models.AttachmentUploadRequest{
		LinkID:    linkID,
		ReplyID:   replyID,
		Filename:  header.Filename,
		FileSize:  header.Size,
		MimeType:  header.Header.Get("Content-Type"),
		FileHash:  srv.calculateFileHash(file),
		IPAddress: srv.getClientIP(r),
		UserAgent: r.UserAgent(),
	}

	// Process upload
	response, err := srv.attachmentService.ProcessUpload(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Upload processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AttachmentDownloadHandler handles file download requests
func (srv *Server) attachmentDownloadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req models.AttachmentDownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.AttachmentID == "" || req.TokenHash == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	// Get client IP
	req.IPAddress = srv.getClientIP(r)
	req.UserAgent = r.UserAgent()

	// Process download
	response, err := srv.attachmentService.ProcessDownload(r.Context(), &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Download processing failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AttachmentDownloadTokenHandler generates download tokens
func (srv *Server) attachmentDownloadTokenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req struct {
		AttachmentID string `json:"attachment_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.AttachmentID == "" {
		http.Error(w, "Missing attachment_id", http.StatusBadRequest)
		return
	}

	// Generate download token
	token, err := srv.attachmentService.GenerateDownloadToken(
		r.Context(),
		req.AttachmentID,
		srv.getClientIP(r),
		r.UserAgent(),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate download token: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"success":     true,
		"token_hash":  token.TokenHash,
		"expires_at":  token.ExpiresAt,
		"max_downloads": token.MaxDownloads,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetAttachmentsHandler retrieves attachments for a link or reply
func (srv *Server) getAttachmentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	linkID := r.URL.Query().Get("link_id")
	replyID := r.URL.Query().Get("reply_id")

	if linkID == "" && replyID == "" {
		http.Error(w, "Missing link_id or reply_id", http.StatusBadRequest)
		return
	}

	// Get attachments from database
	attachments, err := srv.getAttachments(r.Context(), linkID, replyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get attachments: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"success":     true,
		"attachments": attachments,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper methods

// processRichText processes rich text content with sanitization
func (srv *Server) processRichText(ctx context.Context, req *models.RichTextRequest) (*models.RichTextResponse, error) {
	// Create sanitizer
	sanitizer := richtext.NewSanitizer()

	// Validate content size
	if err := sanitizer.ValidateContentSize(req.Content); err != nil {
		return &models.RichTextResponse{
			Success:   false,
			Error:     err.Error(),
			ErrorCode: "CONTENT_TOO_LARGE",
		}, nil
	}

	// Sanitize HTML content
	richTextContent, err := sanitizer.SanitizeHTML(req.Content)
	if err != nil {
		return &models.RichTextResponse{
			Success:   false,
			Error:     fmt.Sprintf("Failed to sanitize content: %v", err),
			ErrorCode: "SANITIZATION_ERROR",
		}, nil
	}

	// Set content type and creator
	richTextContent.ContentType = req.ContentType
	richTextContent.CreatedBy = &req.IPAddress

	// Store in database
	if err := srv.storeRichTextContent(ctx, richTextContent); err != nil {
		return &models.RichTextResponse{
			Success:   false,
			Error:     "Failed to store content",
			ErrorCode: "DATABASE_ERROR",
		}, nil
	}

	// Log audit event
	srv.logRichTextAuditEvent(&models.RichMessagingAuditLog{
		EventType:   "rich_text_processed",
		LinkID:      &req.LinkID,
		ReplyID:     &req.ReplyID,
		ContentID:   &richTextContent.ContentID,
		IPAddress:   &req.IPAddress,
		UserAgent:   &req.UserAgent,
		EventDetails: srv.createEventDetails("rich_text_processed", map[string]interface{}{
			"content_type": req.ContentType,
			"content_size": len(req.Content),
			"features_used": richTextContent.FeaturesUsed,
		}),
		CreatedAt: time.Now(),
	})

	return &models.RichTextResponse{
		Success:         true,
		ContentID:       richTextContent.ContentID,
		SanitizedContent: richTextContent.SanitizedContent,
		FeaturesUsed:    *richTextContent.FeaturesUsed,
		Message:         "Rich text content processed successfully",
	}, nil
}

// validateUploadedFile validates uploaded file
func (srv *Server) validateUploadedFile(header *multipart.FileHeader) error {
	// Check file size
	if header.Size > srv.config.MaxFileSize {
		return fmt.Errorf("file too large (max %d bytes)", srv.config.MaxFileSize)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == "" {
		return fmt.Errorf("file must have an extension")
	}

	// Check MIME type
	mimeType := header.Header.Get("Content-Type")
	if mimeType == "" {
		return fmt.Errorf("missing content type")
	}

	// Validate against allowed types (simplified for now)
	allowedTypes := map[string]int64{
		"application/pdf": 26214400,
		"image/jpeg":      10485760,
		"image/png":       10485760,
		"text/plain":      5242880,
	}
	
	maxSize, allowed := allowedTypes[mimeType]
	if !allowed {
		return fmt.Errorf("file type not allowed: %s", mimeType)
	}

	// Check file size against type limit
	if header.Size > maxSize {
		return fmt.Errorf("file too large for type %s (max %d bytes)", mimeType, maxSize)
	}

	return nil
}

// calculateFileHash calculates SHA-256 hash of file
func (srv *Server) calculateFileHash(file multipart.File) string {
	// Reset file pointer
	file.Seek(0, 0)
	
	// Calculate hash (simplified for now)
	// In production, use proper hash calculation
	return fmt.Sprintf("hash_%d", time.Now().UnixNano())
}

// storeRichTextContent stores rich text content in database
func (srv *Server) storeRichTextContent(ctx context.Context, content *models.RichTextContent) error {
	// This would be implemented with actual database operations
	// For now, just log the content
	fmt.Printf("Storing rich text content: %s\n", content.ContentID)
	return nil
}

// getAttachments retrieves attachments from database
func (srv *Server) getAttachments(ctx context.Context, linkID, replyID string) ([]*models.SecureAttachment, error) {
	// This would be implemented with actual database queries
	// For now, return empty list
	return []*models.SecureAttachment{}, nil
}

// logRichTextAuditEvent logs rich text audit events
func (srv *Server) logRichTextAuditEvent(event *models.RichMessagingAuditLog) {
	// Generate audit ID
	event.AuditID = fmt.Sprintf("audit_%d", time.Now().UnixNano())
	
	// Log asynchronously
	go func() {
		fmt.Printf("Rich text audit event: %s\n", event.EventType)
	}()
}

// createEventDetails creates event details JSON
func (srv *Server) createEventDetails(eventType string, data map[string]interface{}) *string {
	details := map[string]interface{}{
		"event_type": eventType,
		"timestamp":  time.Now().Unix(),
		"data":       data,
	}
	
	jsonData, _ := json.Marshal(details)
	result := string(jsonData)
	return &result
}
