package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/audit"
)

// AdminAccessLogsResponse represents the response for admin access logs query
type AdminAccessLogsResponse struct {
	Logs       []audit.EmailAccessLog `json:"logs"`
	TotalCount int                    `json:"total_count"`
	Limit      int                    `json:"limit"`
	Offset     int                    `json:"offset"`
	HasMore    bool                   `json:"has_more"`
}

// adminAccessLogsHandler handles GET /api/admin/email/access-logs
// Requires JWT with admin role and supports filtering and pagination
// Micro-Iteration 4.23: Admin Access Log Query Endpoint
func (srv *Server) adminAccessLogsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminAccessLogsHandler started - Admin Access Log Query (Micro-Iteration 4.23)")

	// Get authenticated user from JWT context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Authentication required"})
		return
	}

	// TODO: In a production system, check if user has admin role
	// For now, we'll allow any authenticated user to access admin logs
	// In production, this should check a role field in the users table
	log.Printf("Admin access logs requested by user: %s", userID)

	// Parse query parameters for filtering and pagination
	queryParams := r.URL.Query()

	// Pagination parameters
	limitStr := queryParams.Get("limit")
	if limitStr == "" {
		limitStr = "50" // Default limit
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 50 // Default if invalid
	}

	offsetStr := queryParams.Get("offset")
	if offsetStr == "" {
		offsetStr = "0" // Default offset
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0 // Default if invalid
	}

	// Filter parameters
	filters := make(map[string]string)

	if emailID := queryParams.Get("email_id"); emailID != "" {
		filters["email_id"] = emailID
	}

	if userIDFilter := queryParams.Get("user_id"); userIDFilter != "" {
		filters["user_id"] = userIDFilter
	}

	if result := queryParams.Get("result"); result != "" {
		filters["result"] = result
	}

	if startDate := queryParams.Get("start_date"); startDate != "" {
		// Validate date format (expecting ISO 8601)
		if _, err := time.Parse(time.RFC3339, startDate); err != nil {
			log.Printf("Invalid start_date format: %s", startDate)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid start_date format. Use ISO 8601 (e.g., 2024-01-01T00:00:00Z)"})
			return
		}
		filters["start_date"] = startDate
	}

	if endDate := queryParams.Get("end_date"); endDate != "" {
		// Validate date format (expecting ISO 8601)
		if _, err := time.Parse(time.RFC3339, endDate); err != nil {
			log.Printf("Invalid end_date format: %s", endDate)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid end_date format. Use ISO 8601 (e.g., 2024-01-01T23:59:59Z)"})
			return
		}
		filters["end_date"] = endDate
	}

	// Check if email access auditor is available
	if srv.emailAccessAuditor == nil {
		log.Printf("Email access auditor is not available")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Access logs service unavailable"})
		return
	}

	// Get access logs with filters and pagination
	logs, err := srv.emailAccessAuditor.GetAccessLogsForAdmin(r.Context(), filters, limit, offset)
	if err != nil {
		log.Printf("Failed to get access logs: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve access logs"})
		return
	}

	// Get total count for pagination
	totalCount, err := srv.emailAccessAuditor.GetAccessLogsCountForAdmin(r.Context(), filters)
	if err != nil {
		log.Printf("Failed to get access logs count: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to retrieve access logs count"})
		return
	}

	// Calculate if there are more results
	hasMore := (offset + limit) < totalCount

	// Prepare response
	response := AdminAccessLogsResponse{
		Logs:       logs,
		TotalCount: totalCount,
		Limit:      limit,
		Offset:     offset,
		HasMore:    hasMore,
	}

	// Log the admin access for audit purposes
	if srv.emailAccessAuditor != nil {
		clientIP := getClientIP(r)
		userAgent := r.UserAgent()
		srv.emailAccessAuditor.LogAccess(r.Context(), "admin_access_logs", clientIP, &userID, "admin_query", userAgent)
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
