package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"secure-email-mvp/pkg/admin"
)

// AdminAuthContainer holds dependencies for admin authentication handlers
type AdminAuthContainer struct {
	db           *sql.DB
	adminAuthSvc *admin.AdminAuthService
}

// NewAdminAuthContainer creates a new admin authentication container
func NewAdminAuthContainer(db *sql.DB) *AdminAuthContainer {
	return &AdminAuthContainer{
		db:           db,
		adminAuthSvc: admin.NewAdminAuthService(db),
	}
}

// AdminSetupRequest represents the request for initial admin setup
type AdminSetupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminSetupResponse represents the response for admin setup
type AdminSetupResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	AdminID string `json:"admin_id,omitempty"`
}

// AdminLoginRequest represents the request for admin login
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

// AdminLoginResponse represents the response for admin login
type AdminLoginResponse struct {
	Success      bool             `json:"success"`
	Message      string           `json:"message"`
	SessionToken string           `json:"session_token,omitempty"`
	Admin        *admin.AdminUser `json:"admin,omitempty"`
}

// AdminLogoutRequest represents the request for admin logout
type AdminLogoutRequest struct {
	SessionToken string `json:"session_token"`
}

// AdminLogoutResponse represents the response for admin logout
type AdminLogoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AdminSessionResponse represents the response for session validation
type AdminSessionResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Admin   *admin.AdminUser `json:"admin,omitempty"`
}

// adminSetupHandler handles the initial admin setup (first-time bootstrap)
func (c *AdminAuthContainer) adminSetupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AdminSetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Check if admin already exists
	exists, err := c.adminAuthSvc.CheckAdminExists()
	if err != nil {
		log.Printf("Failed to check admin existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, "Admin already exists", http.StatusConflict)
		return
	}

	// Validate root admin email
	rootAdminEmail := os.Getenv("ROOT_ADMIN_EMAIL")
	if rootAdminEmail == "" {
		rootAdminEmail = "cpigusch@gmail.com" // Default fallback
	}

	if req.Email != rootAdminEmail {
		http.Error(w, "Invalid root admin email", http.StatusForbidden)
		return
	}

	// Create root admin
	adminUser, err := c.adminAuthSvc.CreateRootAdmin(req.Email, req.Password)
	if err != nil {
		log.Printf("Failed to create root admin: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create admin: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := AdminSetupResponse{
		Success: true,
		Message: "Root admin created successfully",
		AdminID: adminUser.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// adminLoginHandler handles admin login
func (c *AdminAuthContainer) adminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// Get client IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Authenticate admin
	adminUser, session, err := c.adminAuthSvc.AuthenticateAdmin(req.Email, req.Password, req.TOTPCode, ipAddress, userAgent)
	if err != nil {
		log.Printf("Admin login failed for %s: %v", req.Email, err)
		http.Error(w, fmt.Sprintf("Authentication failed: %v", err), http.StatusUnauthorized)
		return
	}

	// Return success response
	response := AdminLoginResponse{
		Success:      true,
		Message:      "Login successful",
		SessionToken: session.SessionToken,
		Admin:        adminUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminLogoutHandler handles admin logout
func (c *AdminAuthContainer) adminLogoutHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AdminLogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.SessionToken == "" {
		http.Error(w, "Session token is required", http.StatusBadRequest)
		return
	}

	// Get client IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Logout admin
	err := c.adminAuthSvc.LogoutAdmin(req.SessionToken, ipAddress, userAgent)
	if err != nil {
		log.Printf("Admin logout failed: %v", err)
		http.Error(w, fmt.Sprintf("Logout failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := AdminLogoutResponse{
		Success: true,
		Message: "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminSessionHandler validates admin session
func (c *AdminAuthContainer) adminSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Check if it's a Bearer token
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Bearer token required", http.StatusUnauthorized)
		return
	}

	sessionToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate session
	adminUser, err := c.adminAuthSvc.ValidateAdminSession(sessionToken)
	if err != nil {
		log.Printf("Session validation failed: %v", err)
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Return success response
	response := AdminSessionResponse{
		Success: true,
		Message: "Session valid",
		Admin:   adminUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminAuditLogsHandler retrieves admin audit logs
func (c *AdminAuthContainer) adminAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get limit from query parameter
	limit := 100 // Default limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if parsedLimit, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || parsedLimit != 1 {
			limit = 100
		}
	}

	// Get audit logs
	logs, err := c.adminAuthSvc.GetAdminAuditLogs(limit)
	if err != nil {
		log.Printf("Failed to get admin audit logs: %v", err)
		http.Error(w, "Failed to retrieve audit logs", http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminCheckSetupHandler checks if admin setup is required
func (c *AdminAuthContainer) adminCheckSetupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if admin exists
	exists, err := c.adminAuthSvc.CheckAdminExists()
	if err != nil {
		log.Printf("Failed to check admin existence: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	response := map[string]interface{}{
		"setup_required":   !exists,
		"root_admin_email": os.Getenv("ROOT_ADMIN_EMAIL"),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ============================================================================
// INVITATION KEY HANDLERS - ITERATION 4
// ============================================================================

// AdminInvitationRequest represents the request for creating an invitation
type AdminInvitationRequest struct {
	Email   string `json:"email"`
	Role    string `json:"role"`
	MaxUses int    `json:"max_uses"`
}

// AdminInvitationResponse represents the response for invitation creation
type AdminInvitationResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	InvitationID  string `json:"invitation_id,omitempty"`
	InvitationKey string `json:"invitation_key,omitempty"`
	ExpiresAt     string `json:"expires_at,omitempty"`
}

// AdminInvitationValidationRequest represents the request for validating an invitation
type AdminInvitationValidationRequest struct {
	InvitationToken string `json:"invitation_token"`
}

// AdminInvitationValidationResponse represents the response for invitation validation
type AdminInvitationValidationResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	Email       string `json:"email,omitempty"`
	Role        string `json:"role,omitempty"`
	ExpiresAt   string `json:"expires_at,omitempty"`
	MaxUses     int    `json:"max_uses,omitempty"`
	CurrentUses int    `json:"current_uses,omitempty"`
}

// AdminInvitationUseRequest represents the request for using an invitation
type AdminInvitationUseRequest struct {
	InvitationToken string `json:"invitation_token"`
	Password        string `json:"password"`
}

// AdminInvitationUseResponse represents the response for using an invitation
type AdminInvitationUseResponse struct {
	Success      bool             `json:"success"`
	Message      string           `json:"message"`
	Admin        *admin.AdminUser `json:"admin,omitempty"`
	SessionToken string           `json:"session_token,omitempty"`
}

// adminCreateInvitationHandler handles creating invitation keys
func (c *AdminAuthContainer) adminCreateInvitationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get admin ID from context (set by middleware)
	adminID := r.Context().Value("admin_id").(string)
	if adminID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request
	var req AdminInvitationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Email == "" || req.Role == "" {
		http.Error(w, "Email and role are required", http.StatusBadRequest)
		return
	}

	// Set default max uses if not provided
	if req.MaxUses == 0 {
		req.MaxUses = 1
	}

	// Get client IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Create invitation
	invitation, err := c.adminAuthSvc.CreateInvitationKey(adminID, req.Email, req.Role, req.MaxUses, ipAddress, userAgent)
	if err != nil {
		log.Printf("Failed to create invitation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := AdminInvitationResponse{
		Success:      true,
		Message:      "Invitation created successfully",
		InvitationID: invitation.ID,
		ExpiresAt:    invitation.ExpiresAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// adminValidateInvitationHandler handles validating invitation keys
func (c *AdminAuthContainer) adminValidateInvitationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AdminInvitationValidationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.InvitationToken == "" {
		http.Error(w, "Invitation token is required", http.StatusBadRequest)
		return
	}

	// Validate invitation
	invitation, err := c.adminAuthSvc.ValidateInvitationKey(req.InvitationToken)
	if err != nil {
		log.Printf("Invitation validation failed: %v", err)
		http.Error(w, fmt.Sprintf("Invalid invitation: %v", err), http.StatusBadRequest)
		return
	}

	// Return success response
	response := AdminInvitationValidationResponse{
		Success:     true,
		Message:     "Invitation is valid",
		Email:       invitation.Email,
		Role:        invitation.Role,
		ExpiresAt:   invitation.ExpiresAt.Format(time.RFC3339),
		MaxUses:     invitation.MaxUses,
		CurrentUses: invitation.CurrentUses,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminUseInvitationHandler handles using invitation keys to create admin accounts
func (c *AdminAuthContainer) adminUseInvitationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request
	var req AdminInvitationUseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.InvitationToken == "" || req.Password == "" {
		http.Error(w, "Invitation token and password are required", http.StatusBadRequest)
		return
	}

	// Get client IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Use invitation to create admin account
	adminUser, err := c.adminAuthSvc.UseInvitationKey(req.InvitationToken, req.Password, ipAddress, userAgent)
	if err != nil {
		log.Printf("Failed to use invitation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to create admin account: %v", err), http.StatusInternalServerError)
		return
	}

	// Create session for the new admin
	session, err := c.adminAuthSvc.CreateAdminSession(adminUser.ID, ipAddress, userAgent)
	if err != nil {
		log.Printf("Failed to create session for new admin: %v", err)
		// Don't fail the request, just don't return a session token
	}

	// Return success response
	response := AdminInvitationUseResponse{
		Success:      true,
		Message:      "Admin account created successfully",
		Admin:        adminUser,
		SessionToken: session.SessionToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// adminListInvitationsHandler handles listing invitation keys
func (c *AdminAuthContainer) adminListInvitationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get admin ID from context (set by middleware)
	adminID := r.Context().Value("admin_id").(string)
	if adminID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get invitations
	invitations, err := c.adminAuthSvc.ListInvitationKeys(adminID)
	if err != nil {
		log.Printf("Failed to list invitations: %v", err)
		http.Error(w, fmt.Sprintf("Failed to list invitations: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success":     true,
		"invitations": invitations,
		"count":       len(invitations),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRevokeInvitationHandler handles revoking invitation keys
func (c *AdminAuthContainer) adminRevokeInvitationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get admin ID from context (set by middleware)
	adminID := r.Context().Value("admin_id").(string)
	if adminID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get invitation ID from URL path
	invitationID := r.URL.Query().Get("id")
	if invitationID == "" {
		http.Error(w, "Invitation ID is required", http.StatusBadRequest)
		return
	}

	// Get client IP and user agent
	ipAddress := getClientIP(r)
	userAgent := r.Header.Get("User-Agent")

	// Revoke invitation
	err := c.adminAuthSvc.RevokeInvitationKey(adminID, invitationID, ipAddress, userAgent)
	if err != nil {
		log.Printf("Failed to revoke invitation: %v", err)
		http.Error(w, fmt.Sprintf("Failed to revoke invitation: %v", err), http.StatusInternalServerError)
		return
	}

	// Return success response
	response := map[string]interface{}{
		"success": true,
		"message": "Invitation revoked successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
