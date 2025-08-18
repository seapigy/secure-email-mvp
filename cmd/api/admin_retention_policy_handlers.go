package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"secure-email-mvp/pkg/email"

	"github.com/gorilla/mux"
)

// AdminRetentionPoliciesResponse represents the response for listing retention policies
type AdminRetentionPoliciesResponse struct {
	Policies   []*email.RetentionPolicy `json:"policies"`
	TotalCount int                      `json:"total_count"`
	Limit      int                      `json:"limit"`
	Offset     int                      `json:"offset"`
}

// CreateRetentionPolicyRequest represents the request for creating a retention policy
type CreateRetentionPolicyRequest struct {
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	Priority             int     `json:"priority"`
	Active               bool    `json:"active"`
	UserID               *string `json:"user_id,omitempty"`
	SenderDomain         *string `json:"sender_domain,omitempty"`
	RecipientDomain      *string `json:"recipient_domain,omitempty"`
	EmailStatus          *string `json:"email_status,omitempty"`
	CustomTags           *string `json:"custom_tags,omitempty"`
	MinAgeHours          *int    `json:"min_age_hours,omitempty"`
	MaxAgeHours          *int    `json:"max_age_hours,omitempty"`
	RetentionDays        int     `json:"retention_days"`
	ArchiveInstead       bool    `json:"archive_instead"`
	ArchiveRetentionDays int     `json:"archive_retention_days"`
}

// UpdateRetentionPolicyRequest represents the request for updating a retention policy
type UpdateRetentionPolicyRequest struct {
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	Priority             int     `json:"priority"`
	Active               bool    `json:"active"`
	UserID               *string `json:"user_id,omitempty"`
	SenderDomain         *string `json:"sender_domain,omitempty"`
	RecipientDomain      *string `json:"recipient_domain,omitempty"`
	EmailStatus          *string `json:"email_status,omitempty"`
	CustomTags           *string `json:"custom_tags,omitempty"`
	MinAgeHours          *int    `json:"min_age_hours,omitempty"`
	MaxAgeHours          *int    `json:"max_age_hours,omitempty"`
	RetentionDays        int     `json:"retention_days"`
	ArchiveInstead       bool    `json:"archive_instead"`
	ArchiveRetentionDays int     `json:"archive_retention_days"`
}

// adminRetentionPoliciesHandler handles GET /api/admin/email/retention-policies
func (srv *Server) adminRetentionPoliciesHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	activeStr := r.URL.Query().Get("active")
	userID := r.URL.Query().Get("user_id")
	domain := r.URL.Query().Get("domain")

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
	if activeStr != "" {
		filters["active"] = activeStr
	}
	if userID != "" {
		filters["user_id"] = userID
	}
	if domain != "" {
		filters["domain"] = domain
	}

	// Get policies
	policies, err := srv.retentionPolicyEngine.GetPolicies(r.Context(), filters, limit, offset)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get retention policies: %v", err), http.StatusInternalServerError)
		return
	}

	// Get total count (without limit/offset)
	totalPolicies, err := srv.retentionPolicyEngine.GetPolicies(r.Context(), filters, 0, 0)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get total count: %v", err), http.StatusInternalServerError)
		return
	}

	response := AdminRetentionPoliciesResponse{
		Policies:   policies,
		TotalCount: len(totalPolicies),
		Limit:      limit,
		Offset:     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminCreateRetentionPolicyHandler handles POST /api/admin/email/retention-policies
func (srv *Server) adminCreateRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	var req CreateRetentionPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Policy name is required", http.StatusBadRequest)
		return
	}

	if req.RetentionDays <= 0 {
		http.Error(w, "Retention days must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.ArchiveRetentionDays <= 0 {
		http.Error(w, "Archive retention days must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create policy
	policy := &email.RetentionPolicy{
		Name:                 req.Name,
		Description:          req.Description,
		Priority:             req.Priority,
		Active:               req.Active,
		UserID:               req.UserID,
		SenderDomain:         req.SenderDomain,
		RecipientDomain:      req.RecipientDomain,
		EmailStatus:          req.EmailStatus,
		CustomTags:           req.CustomTags,
		MinAgeHours:          req.MinAgeHours,
		MaxAgeHours:          req.MaxAgeHours,
		RetentionDays:        req.RetentionDays,
		ArchiveInstead:       req.ArchiveInstead,
		ArchiveRetentionDays: req.ArchiveRetentionDays,
		CreatedBy:            "admin", // TODO: Get from JWT token
	}

	if err := srv.retentionPolicyEngine.CreatePolicy(r.Context(), policy); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create retention policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(policy)
}

// adminUpdateRetentionPolicyHandler handles PUT /api/admin/email/retention-policies/{policy_id}
func (srv *Server) adminUpdateRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	// Extract policy ID from URL
	vars := mux.Vars(r)
	policyIDStr := vars["policy_id"]

	policyID, err := strconv.ParseInt(policyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid policy ID", http.StatusBadRequest)
		return
	}

	var req UpdateRetentionPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid request body: %v", err), http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Policy name is required", http.StatusBadRequest)
		return
	}

	if req.RetentionDays <= 0 {
		http.Error(w, "Retention days must be greater than 0", http.StatusBadRequest)
		return
	}

	if req.ArchiveRetentionDays <= 0 {
		http.Error(w, "Archive retention days must be greater than 0", http.StatusBadRequest)
		return
	}

	// Get existing policy
	existingPolicy, err := srv.retentionPolicyEngine.GetPolicyByID(r.Context(), policyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Policy not found: %v", err), http.StatusNotFound)
		return
	}

	// Update fields
	existingPolicy.Name = req.Name
	existingPolicy.Description = req.Description
	existingPolicy.Priority = req.Priority
	existingPolicy.Active = req.Active
	existingPolicy.UserID = req.UserID
	existingPolicy.SenderDomain = req.SenderDomain
	existingPolicy.RecipientDomain = req.RecipientDomain
	existingPolicy.EmailStatus = req.EmailStatus
	existingPolicy.CustomTags = req.CustomTags
	existingPolicy.MinAgeHours = req.MinAgeHours
	existingPolicy.MaxAgeHours = req.MaxAgeHours
	existingPolicy.RetentionDays = req.RetentionDays
	existingPolicy.ArchiveInstead = req.ArchiveInstead
	existingPolicy.ArchiveRetentionDays = req.ArchiveRetentionDays

	if err := srv.retentionPolicyEngine.UpdatePolicy(r.Context(), existingPolicy); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update retention policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingPolicy)
}

// adminDeleteRetentionPolicyHandler handles DELETE /api/admin/email/retention-policies/{policy_id}
func (srv *Server) adminDeleteRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	// Extract policy ID from URL
	vars := mux.Vars(r)
	policyIDStr := vars["policy_id"]

	policyID, err := strconv.ParseInt(policyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid policy ID", http.StatusBadRequest)
		return
	}

	if err := srv.retentionPolicyEngine.DeletePolicy(r.Context(), policyID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete retention policy: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// adminGetRetentionPolicyHandler handles GET /api/admin/email/retention-policies/{policy_id}
func (srv *Server) adminGetRetentionPolicyHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Add JWT authentication and admin role check

	// Extract policy ID from URL
	vars := mux.Vars(r)
	policyIDStr := vars["policy_id"]

	policyID, err := strconv.ParseInt(policyIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid policy ID", http.StatusBadRequest)
		return
	}

	policy, err := srv.retentionPolicyEngine.GetPolicyByID(r.Context(), policyID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Policy not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(policy)
}
