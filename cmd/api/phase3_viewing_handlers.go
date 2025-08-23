package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/securelinks/attachments"
	"secure-email-mvp/pkg/securelinks/chains"
	"secure-email-mvp/pkg/securelinks/reply"
	"secure-email-mvp/pkg/securelinks/viewer"
)

// =============================================================================
// PHASE 3 VIEWING & REPLY API HANDLERS
// =============================================================================

// SecureViewerHandler handles secure email viewing for external users
func (srv *Server) SecureViewerHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("SecureViewerHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodPost:
		srv.handleCreateViewSession(w, r)
	case http.MethodGet:
		srv.handleGetEmailView(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleCreateViewSession handles POST requests to create a new viewing session
func (srv *Server) handleCreateViewSession(w http.ResponseWriter, r *http.Request) {
	var req viewer.CreateViewSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get client IP address
	req.IPAddress = srv.getClientIP(r)
	if req.IPAddress == "" {
		req.IPAddress = "unknown"
	}

	// Get user agent
	req.UserAgent = r.UserAgent()

	// Validate required fields
	if req.LinkID == "" {
		http.Error(w, "Link ID is required", http.StatusBadRequest)
		return
	}

	// Create viewer service
	viewerService := viewer.NewViewerService(srv.db)

	// Create view session
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resp, err := viewerService.CreateViewSession(ctx, req)
	if err != nil {
		log.Printf("Error creating view session: %v", err)
		http.Error(w, "Failed to create viewing session", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleGetEmailView handles GET requests to retrieve email content
func (srv *Server) handleGetEmailView(w http.ResponseWriter, r *http.Request) {
	// Get session token from query parameter or header
	sessionToken := r.URL.Query().Get("session_token")
	if sessionToken == "" {
		sessionToken = r.Header.Get("X-Session-Token")
	}

	if sessionToken == "" {
		http.Error(w, "Session token is required", http.StatusBadRequest)
		return
	}

	// Create viewer service
	viewerService := viewer.NewViewerService(srv.db)

	// Get email view
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	req := viewer.GetEmailViewRequest{
		SessionToken: sessionToken,
	}

	resp, err := viewerService.GetEmailView(ctx, req)
	if err != nil {
		log.Printf("Error getting email view: %v", err)
		http.Error(w, "Failed to retrieve email content", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// SecureReplyHandler handles replies from external users
func (srv *Server) SecureReplyHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("SecureReplyHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodPost:
		srv.handleProcessReply(w, r)
	case http.MethodGet:
		srv.handleGetReplyHistory(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleProcessReply handles POST requests to process a reply
func (srv *Server) handleProcessReply(w http.ResponseWriter, r *http.Request) {
	var req reply.ReplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding reply request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get client IP address
	req.IPAddress = srv.getClientIP(r)
	if req.IPAddress == "" {
		req.IPAddress = "unknown"
	}

	// Get user agent
	req.UserAgent = r.UserAgent()

	// Validate required fields
	if req.LinkID == "" || req.Subject == "" || req.Body == "" || req.SenderEmail == "" {
		http.Error(w, "Link ID, subject, body, and sender email are required", http.StatusBadRequest)
		return
	}

	// Create reply service
	replyService := reply.NewReplyService(srv.db)

	// Process reply
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resp, err := replyService.ProcessReply(ctx, req)
	if err != nil {
		log.Printf("Error processing reply: %v", err)
		http.Error(w, "Failed to process reply", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleGetReplyHistory handles GET requests to retrieve reply history
func (srv *Server) handleGetReplyHistory(w http.ResponseWriter, r *http.Request) {
	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		http.Error(w, "Chain ID is required", http.StatusBadRequest)
		return
	}

	// Create reply service
	replyService := reply.NewReplyService(srv.db)

	// Get reply history
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	replies, err := replyService.GetReplyHistory(ctx, chainID)
	if err != nil {
		log.Printf("Error getting reply history: %v", err)
		http.Error(w, "Failed to retrieve reply history", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"replies": replies,
	})
}

// EmailChainHandler handles email chain management
func (srv *Server) EmailChainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("EmailChainHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodGet:
		srv.handleGetChain(w, r)
	case http.MethodPost:
		srv.handleCreateChain(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetChain handles GET requests to retrieve chain information
func (srv *Server) handleGetChain(w http.ResponseWriter, r *http.Request) {
	chainID := r.URL.Query().Get("chain_id")
	if chainID == "" {
		http.Error(w, "Chain ID is required", http.StatusBadRequest)
		return
	}

	// Create chain service
	chainService := chains.NewChainsService(srv.db)

	// Get chain messages
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	messages, err := chainService.GetChainMessages(ctx, chainID)
	if err != nil {
		log.Printf("Error getting chain messages: %v", err)
		http.Error(w, "Failed to retrieve chain messages", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"chain_id": chainID,
		"messages": messages,
	})
}

// handleCreateChain handles POST requests to create a new chain
func (srv *Server) handleCreateChain(w http.ResponseWriter, r *http.Request) {
	var req struct {
		LinkID         string `json:"link_id"`
		InternalUserID string `json:"internal_user_id"`
		ExternalEmail  string `json:"external_email"`
		Subject        string `json:"subject"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding chain request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.LinkID == "" || req.InternalUserID == "" || req.ExternalEmail == "" || req.Subject == "" {
		http.Error(w, "Link ID, internal user ID, external email, and subject are required", http.StatusBadRequest)
		return
	}

	// Create chain service
	chainService := chains.NewChainsService(srv.db)

	// Create chain
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	chain, err := chainService.CreateChain(ctx, req.LinkID, req.InternalUserID, req.ExternalEmail, req.Subject)
	if err != nil {
		log.Printf("Error creating chain: %v", err)
		http.Error(w, "Failed to create chain", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"chain":   chain,
	})
}

// SecureAttachmentHandler handles secure attachment downloads
func (srv *Server) SecureAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("SecureAttachmentHandler called - Method: %s", r.Method)

	switch r.Method {
	case http.MethodPost:
		srv.handleUploadAttachment(w, r)
	case http.MethodGet:
		srv.handleDownloadAttachment(w, r)
	case http.MethodDelete:
		srv.handleDeleteAttachment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleUploadAttachment handles POST requests to upload attachments
func (srv *Server) handleUploadAttachment(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max
		log.Printf("Error parsing multipart form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get form values
	linkID := r.FormValue("link_id")
	if linkID == "" {
		http.Error(w, "Link ID is required", http.StatusBadRequest)
		return
	}

	// Get uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Printf("Error getting uploaded file: %v", err)
		http.Error(w, "No file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create attachment service
	attachmentService := attachments.NewAttachmentService(srv.db, "./attachments")

	// Create upload request
	req := attachments.UploadAttachmentRequest{
		LinkID: linkID,
		File:   header,
	}

	// Upload attachment
	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	resp, err := attachmentService.UploadAttachment(ctx, req)
	if err != nil {
		log.Printf("Error uploading attachment: %v", err)
		http.Error(w, "Failed to upload attachment", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleDownloadAttachment handles GET requests to download attachments
func (srv *Server) handleDownloadAttachment(w http.ResponseWriter, r *http.Request) {
	attachmentID := r.URL.Query().Get("attachment_id")
	sessionToken := r.URL.Query().Get("session_token")

	if attachmentID == "" || sessionToken == "" {
		http.Error(w, "Attachment ID and session token are required", http.StatusBadRequest)
		return
	}

	// Create attachment service
	attachmentService := attachments.NewAttachmentService(srv.db, "./attachments")

	// Create download request
	req := attachments.DownloadAttachmentRequest{
		AttachmentID: attachmentID,
		SessionToken: sessionToken,
		IPAddress:    srv.getClientIP(r),
		UserAgent:    r.UserAgent(),
	}

	// Download attachment
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	resp, err := attachmentService.DownloadAttachment(ctx, req)
	if err != nil {
		log.Printf("Error downloading attachment: %v", err)
		http.Error(w, "Failed to download attachment", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if !resp.Success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(resp)
}

// handleDeleteAttachment handles DELETE requests to delete attachments
func (srv *Server) handleDeleteAttachment(w http.ResponseWriter, r *http.Request) {
	attachmentID := r.URL.Query().Get("attachment_id")
	if attachmentID == "" {
		http.Error(w, "Attachment ID is required", http.StatusBadRequest)
		return
	}

	// Create attachment service
	attachmentService := attachments.NewAttachmentService(srv.db, "./attachments")

	// Delete attachment
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	err := attachmentService.DeleteAttachment(ctx, attachmentID)
	if err != nil {
		log.Printf("Error deleting attachment: %v", err)
		http.Error(w, "Failed to delete attachment", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Attachment deleted successfully",
	})
}

// Helper method to get client IP address
func (srv *Server) getClientIP(r *http.Request) string {
	// Check for forwarded headers first (for proxy/load balancer scenarios)
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}

	// Fall back to remote address
	return r.RemoteAddr
}
