package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"
)

// AdminRealtimeMetricsResponse represents the response for real-time metrics
type AdminRealtimeMetricsResponse struct {
	Success bool                    `json:"success"`
	Data    *email.RealtimeMetrics `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// AdminAdaptiveChangesResponse represents the response for adaptive policy changes
type AdminAdaptiveChangesResponse struct {
	Success bool                        `json:"success"`
	Data    []*email.AdaptivePolicyChange `json:"data,omitempty"`
	Total   int                         `json:"total,omitempty"`
	Error   string                      `json:"error,omitempty"`
}

// EnableAdaptivePolicyRequest represents the request to enable adaptive policy
type EnableAdaptivePolicyRequest struct {
	PolicyID               int64   `json:"policy_id"`
	MaxChangePercentage    float64 `json:"max_change_percentage"`
	CooldownDays           int     `json:"cooldown_days"`
	RequiresAdminApproval  bool    `json:"requires_admin_approval"`
	MinRetentionDays       int     `json:"min_retention_days"`
	MaxRetentionDays       int     `json:"max_retention_days"`
	MinArchiveRetentionDays int    `json:"min_archive_retention_days"`
	MaxArchiveRetentionDays int    `json:"max_archive_retention_days"`
	MaxStorageImpactBytes  int64   `json:"max_storage_impact_bytes"`
	MaxArchivalLoadImpact  float64 `json:"max_archival_load_impact"`
}

// EnableAdaptivePolicyResponse represents the response for enabling adaptive policy
type EnableAdaptivePolicyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ApplyAdaptiveChangeRequest represents the request to apply an adaptive change
type ApplyAdaptiveChangeRequest struct {
	ChangeID  int64 `json:"change_id"`
	Preview   bool  `json:"preview"`
	AppliedBy string `json:"applied_by"`
}

// ApplyAdaptiveChangeResponse represents the response for applying an adaptive change
type ApplyAdaptiveChangeResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// adminRealtimeMetricsHandler handles GET /api/admin/email/retention-realtime
func (srv *Server) adminRealtimeMetricsHandler(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	metricType := r.URL.Query().Get("metric_type")
	metricKey := r.URL.Query().Get("metric_key")
	
	// Default to global metrics if not specified
	if metricType == "" {
		metricType = "global"
	}
	if metricKey == "" {
		metricKey = "global"
	}
	
	// Get real-time metrics
	metrics, err := srv.retentionMonitorService.GetRealtimeMetrics(r.Context(), metricType, metricKey)
	if err != nil {
		response := AdminRealtimeMetricsResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get real-time metrics: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := AdminRealtimeMetricsResponse{
		Success: true,
		Data:    metrics,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminAdaptivePolicyChangesHandler handles GET /api/admin/email/adaptive-policy-changes
func (srv *Server) adminAdaptivePolicyChangesHandler(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	filters := make(map[string]string)
	
	if policyID := r.URL.Query().Get("policy_id"); policyID != "" {
		filters["policy_id"] = policyID
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		filters["status"] = status
	}
	
	if changeType := r.URL.Query().Get("change_type"); changeType != "" {
		filters["change_type"] = changeType
	}
	
	// Get pagination parameters
	limit := 50 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	
	offset := 0
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	// Get adaptive changes
	changes, err := srv.adaptiveRetentionEngine.GetAdaptiveChanges(r.Context(), filters, limit, offset)
	if err != nil {
		response := AdminAdaptiveChangesResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to get adaptive changes: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Get total count for pagination
	total := len(changes) // For now, just use the current result count
	
	response := AdminAdaptiveChangesResponse{
		Success: true,
		Data:    changes,
		Total:   total,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminEnableAdaptivePolicyHandler handles POST /api/admin/email/adaptive-policy/enable
func (srv *Server) adminEnableAdaptivePolicyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req EnableAdaptivePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := EnableAdaptivePolicyResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Validate request
	if req.PolicyID <= 0 {
		response := EnableAdaptivePolicyResponse{
			Success: false,
			Error:   "Invalid policy ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Create adaptive config
	config := &email.AdaptivePolicyConfig{
		PolicyID:                req.PolicyID,
		AdaptiveEnabled:         true,
		MaxChangePercentage:     req.MaxChangePercentage,
		CooldownDays:            req.CooldownDays,
		RequiresAdminApproval:   req.RequiresAdminApproval,
		MinRetentionDays:        req.MinRetentionDays,
		MaxRetentionDays:        req.MaxRetentionDays,
		MinArchiveRetentionDays: req.MinArchiveRetentionDays,
		MaxArchiveRetentionDays: req.MaxArchiveRetentionDays,
		MaxStorageImpactBytes:   req.MaxStorageImpactBytes,
		MaxArchivalLoadImpact:   req.MaxArchivalLoadImpact,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}
	
	// Enable adaptive policy
	if err := srv.adaptiveRetentionEngine.EnableAdaptivePolicy(r.Context(), req.PolicyID, config); err != nil {
		response := EnableAdaptivePolicyResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to enable adaptive policy: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := EnableAdaptivePolicyResponse{
		Success: true,
		Message: fmt.Sprintf("Adaptive policy enabled for policy ID %d", req.PolicyID),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminDisableAdaptivePolicyHandler handles POST /api/admin/email/adaptive-policy/disable
func (srv *Server) adminDisableAdaptivePolicyHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		PolicyID int64 `json:"policy_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := EnableAdaptivePolicyResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Validate request
	if req.PolicyID <= 0 {
		response := EnableAdaptivePolicyResponse{
			Success: false,
			Error:   "Invalid policy ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Disable adaptive policy
	if err := srv.adaptiveRetentionEngine.DisableAdaptivePolicy(r.Context(), req.PolicyID); err != nil {
		response := EnableAdaptivePolicyResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to disable adaptive policy: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := EnableAdaptivePolicyResponse{
		Success: true,
		Message: fmt.Sprintf("Adaptive policy disabled for policy ID %d", req.PolicyID),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminApplyAdaptiveChangeHandler handles POST /api/admin/email/adaptive-policy/apply
func (srv *Server) adminApplyAdaptiveChangeHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req ApplyAdaptiveChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := ApplyAdaptiveChangeResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to parse request: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Validate request
	if req.ChangeID <= 0 {
		response := ApplyAdaptiveChangeResponse{
			Success: false,
			Error:   "Invalid change ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Get user from context (assuming it's set by JWT middleware)
	userID := "admin" // Default fallback
	if user := r.Context().Value("user"); user != nil {
		if userMap, ok := user.(map[string]interface{}); ok {
			if id, ok := userMap["id"].(string); ok {
				userID = id
			}
		}
	}
	
	// Use provided applied_by or default to user ID
	appliedBy := req.AppliedBy
	if appliedBy == "" {
		appliedBy = userID
	}
	
	// Apply adaptive change
	result, err := srv.adaptiveRetentionEngine.ApplyAdaptiveChange(r.Context(), req.ChangeID, appliedBy, req.Preview)
	if err != nil {
		response := ApplyAdaptiveChangeResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to apply adaptive change: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := ApplyAdaptiveChangeResponse{
		Success: true,
		Data:    result,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminPolicyPerformanceHandler handles GET /api/admin/email/policy-performance/{policy_id}
func (srv *Server) adminPolicyPerformanceHandler(w http.ResponseWriter, r *http.Request) {
	// Extract policy ID from URL path
	policyIDStr := r.URL.Query().Get("policy_id")
	if policyIDStr == "" {
		response := map[string]interface{}{
			"success": false,
			"error":   "Policy ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	policyID, err := strconv.ParseInt(policyIDStr, 10, 64)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   "Invalid policy ID",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	// Analyze policy performance
	metrics, err := srv.adaptiveRetentionEngine.AnalyzePolicyPerformance(r.Context(), policyID)
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to analyze policy performance: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"data":    metrics,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// adminGenerateAdaptiveRecommendationsHandler handles POST /api/admin/email/adaptive-policy/generate-recommendations
func (srv *Server) adminGenerateAdaptiveRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	// Generate adaptive recommendations
	recommendations, err := srv.adaptiveRetentionEngine.GenerateAdaptiveRecommendations(r.Context())
	if err != nil {
		response := map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to generate recommendations: %v", err),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}
	
	response := map[string]interface{}{
		"success": true,
		"data":    recommendations,
		"count":   len(recommendations),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
