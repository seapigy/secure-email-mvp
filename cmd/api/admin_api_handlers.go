package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"secure-email-mvp/pkg/middleware"
	"secure-email-mvp/pkg/securelinks/admin"
	"secure-email-mvp/pkg/securelinks/audit"
	"secure-email-mvp/pkg/models"
)

type AdminAPIContainer struct {
	db           *sql.DB
	adminService *admin.Service
	adminRepo    admin.Repository
	auditService *audit.Service
}

func NewAdminAPIContainer(db *sql.DB) *AdminAPIContainer {
	repo := admin.NewSQLiteAdminRepository(db)
	service := admin.NewService(db, repo)
	
	auditRepo := audit.NewSQLiteAuditRepository(db)
	auditService := audit.NewService(auditRepo)
	
	return &AdminAPIContainer{
		db:           db,
		adminService: service,
		adminRepo:    repo,
		auditService: auditService,
	}
}

// APIAdminLoginRequest represents the request for admin login via API
type APIAdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// APIAdminLoginResponse represents the response for admin login via API
type APIAdminLoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Admin   *models.AdminUser `json:"admin,omitempty"`
}

// adminLoginHandler handles POST /api/admin/login
func (c *AdminAPIContainer) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req APIAdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Authenticate admin
	token, adminUser, err := c.adminService.LoginAdmin(req.Email, req.Password)
	if err != nil {
		log.Printf("Admin login failed for %s: %v", req.Email, err)
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Log admin login
	if adminUser != nil {
		c.auditService.LogAdminLogin(adminUser.Email)
	}

	// Return success response
	response := APIAdminLoginResponse{
		Success: true,
		Message: "Login successful",
		Token:   token,
		Admin:   adminUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminDLPLogsHandler handles GET /api/admin/dlp/logs
func (c *AdminAPIContainer) adminDLPLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Get admin info from context
	adminEmail := middleware.GetAdminEmailFromContext(r.Context())
	
	// For now, return mock DLP logs
	// In a real implementation, this would fetch from the database
	dlpLogs := []map[string]interface{}{
		{
			"id":        "1",
			"timestamp": "2024-01-15T10:30:00Z",
			"email":     "user@example.com",
			"scan_type": "content_scan",
			"findings":  []string{"credit_card", "ssn"},
			"action":    "blocked",
			"admin":     adminEmail,
		},
		{
			"id":        "2",
			"timestamp": "2024-01-15T11:15:00Z",
			"email":     "user2@example.com",
			"scan_type": "content_scan",
			"findings":  []string{},
			"action":    "allowed",
			"admin":     adminEmail,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"logs":    dlpLogs,
	})
}

// adminAuditLogsHandler handles GET /api/admin/audit/logs
func (c *AdminAPIContainer) adminAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Get admin info from context
	adminEmail := middleware.GetAdminEmailFromContext(r.Context())
	
	// Parse query parameters for filtering and pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}
	
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100 // Max page size
	}
	
	offset := (page - 1) * limit
	
	// Build filters
	filters := models.AuditLogFilters{
		UserID:   r.URL.Query().Get("user_id"),
		Action:   r.URL.Query().Get("action"),
		Entity:   r.URL.Query().Get("entity"),
		Severity: r.URL.Query().Get("severity"),
		Limit:    limit,
		Offset:   offset,
	}
	
	// Get audit logs
	logs, total, err := c.auditService.GetAuditLogs(filters)
	if err != nil {
		log.Printf("Failed to get audit logs: %v", err)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}
	
	// Log audit log view
	c.auditService.LogAuditLogView(adminEmail)
	
	// Return response
	response := models.AuditLogResponse{
		Success: true,
		Logs:    logs,
		Total:   total,
		Page:    page,
		Limit:   limit,
		Filters: filters,
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
