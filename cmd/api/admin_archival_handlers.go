package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"secure-email-mvp/pkg/email"

	"github.com/gorilla/mux"
)

// AdminArchivedEmailsResponse represents the response for listing archived emails
type AdminArchivedEmailsResponse struct {
	ArchivedEmails []*email.ArchivedEmail `json:"archived_emails"`
	TotalCount     int                    `json:"total_count"`
	Limit          int                    `json:"limit"`
	Offset         int                    `json:"offset"`
}

// ArchiveEmailRequest represents the request for archiving an email
type ArchiveEmailRequest struct {
	EmailID       string `json:"email_id"`
	ArchiveReason string `json:"archive_reason"`
	RetentionDays int    `json:"retention_days"`
}

// RestoreEmailRequest represents the request for restoring an archived email
type RestoreEmailRequest struct {
	ArchiveID int64 `json:"archive_id"`
}

// adminArchivedEmailsHandler handles GET /api/admin/email/archived
func (srv *Server) adminArchivedEmailsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	senderID := r.URL.Query().Get("sender_id")
	recipient := r.URL.Query().Get("recipient")
	reason := r.URL.Query().Get("archive_reason")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	limit := 10 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filters
	filters := make(map[string]string)
	if senderID != "" {
		filters["sender_id"] = senderID
	}
	if recipient != "" {
		filters["recipient"] = recipient
	}
	if reason != "" {
		filters["archive_reason"] = reason
	}
	if startDate != "" {
		filters["start_date"] = startDate
	}
	if endDate != "" {
		filters["end_date"] = endDate
	}

	// Get archived emails
	archivedEmails, err := srv.emailArchivalService.GetArchivedEmails(r.Context(), filters, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get archived emails: %v", err), http.StatusInternalServerError)
		return
	}

	// Get total count (without limit/offset)
	totalArchived, err := srv.emailArchivalService.GetArchivedEmails(r.Context(), filters, 0, 0)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get total count: %v", err), http.StatusInternalServerError)
		return
	}

	response := AdminArchivedEmailsResponse{
		ArchivedEmails: archivedEmails,
		TotalCount:     len(totalArchived),
		Limit:          limit,
		Offset:         offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminGetArchivedEmailHandler handles GET /api/admin/email/archived/{archive_id}
func (srv *Server) adminGetArchivedEmailHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	// Extract archive ID from URL
	vars := mux.Vars(r)
	archiveIDStr := vars["archive_id"]

	archiveID, err := strconv.ParseInt(archiveIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid archive ID", http.StatusBadRequest)
		return
	}

	archivedEmail, err := srv.emailArchivalService.GetArchivedEmailByID(r.Context(), archiveID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Archived email not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(archivedEmail)
}

// adminArchiveEmailHandler handles POST /api/admin/email/archived
func (srv *Server) adminArchiveEmailHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	var req ArchiveEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.EmailID == "" {
		http.Error(w, "Email ID is required", http.StatusBadRequest)
		return
	}

	if req.ArchiveReason == "" {
		http.Error(w, "Archive reason is required", http.StatusBadRequest)
		return
	}

	if req.RetentionDays <= 0 {
		http.Error(w, "Retention days must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create archive request
	archiveReq := &email.ArchiveRequest{
		EmailID:       req.EmailID,
		ArchiveReason: req.ArchiveReason,
		RetentionDays: req.RetentionDays,
	}

	// Archive the email
	response, err := srv.emailArchivalService.ArchiveEmail(r.Context(), archiveReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to archive email: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// adminRestoreEmailHandler handles POST /api/admin/email/archived/restore
func (srv *Server) adminRestoreEmailHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	var req RestoreEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.ArchiveID <= 0 {
		http.Error(w, "Valid archive ID is required", http.StatusBadRequest)
		return
	}

	// Create restore request
	restoreReq := &email.RestoreRequest{
		ArchiveID: req.ArchiveID,
	}

	// Restore the email
	response, err := srv.emailArchivalService.RestoreEmail(r.Context(), restoreReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to restore email: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminArchivalStatsHandler handles GET /api/admin/email/archived/stats
func (srv *Server) adminArchivalStatsHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	stats, err := srv.emailArchivalService.GetArchivalStats(r.Context())
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get archival stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// adminCleanupExpiredArchivesHandler handles POST /api/admin/email/archived/cleanup
func (srv *Server) adminCleanupExpiredArchivesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	if err := srv.emailArchivalService.CleanupExpiredArchives(r.Context()); err != nil {
		http.Error(w, fmt.Sprintf("Failed to cleanup expired archives: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"success": true,
		"message": "Expired archives cleanup completed successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}








