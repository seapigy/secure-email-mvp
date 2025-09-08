package auth

// DO NOT EDIT EXISTING CODE - new file added
// Organization management handlers: create, manage users, and enterprise features

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type createOrganizationRequest struct {
	Name string `json:"name"`
}

type createOrganizationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	OrganizationID string `json:"organization_id"`
}

type addUserToOrgRequest struct {
	OrganizationID string `json:"organization_id"`
	UserEmail      string `json:"user_email"`
	Role           string `json:"role,omitempty"` // admin, member, viewer
}

type addUserToOrgResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type removeUserFromOrgRequest struct {
	OrganizationID string `json:"organization_id"`
	UserID         string `json:"user_id"`
}

type removeUserFromOrgResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type listOrgUsersResponse struct {
	Success bool   `json:"success"`
	Users   []OrgUser `json:"users"`
}

type OrgUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	JoinedAt string `json:"joined_at"`
}

// POST /api/org/create
func CreateOrganizationHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "organization name required", http.StatusBadRequest)
		return
	}

	// Check if user has enterprise account
	var accountType string
	err := DB.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
	if err != nil {
		log.Printf("ERROR getting user account type: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if accountType != "enterprise" {
		http.Error(w, "enterprise account required to create organizations", http.StatusForbidden)
		return
	}

	// Check if user already has an organization
	var existingOrgID string
	err = DB.QueryRow("SELECT id FROM organizations WHERE admin_user_id = ?", userID).Scan(&existingOrgID)
	if err == nil {
		http.Error(w, "user already has an organization", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("ERROR checking existing organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Create organization
	orgID := uuid.New().String()
	now := time.Now().UTC()

	_, err = DB.Exec(`
		INSERT INTO organizations (id, name, admin_user_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`, orgID, req.Name, userID, now, now)

	if err != nil {
		log.Printf("ERROR creating organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Add admin user to organization
	_, err = DB.Exec(`
		INSERT INTO organization_members (id, organization_id, user_id, role, status, joined_at)
		VALUES (?, ?, ?, 'admin', 'active', ?)
	`, uuid.New().String(), orgID, userID, now)

	if err != nil {
		log.Printf("ERROR adding admin to organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Update user's organization_id
	_, err = DB.Exec(`
		UPDATE users 
		SET organization_id = ?, updated_at = ?
		WHERE id = ?
	`, orgID, now, userID)

	if err != nil {
		log.Printf("ERROR updating user organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log organization creation (non-sensitive)
	log.Printf("INFO organization_created user_id=%s org_id=%s", userID, orgID)

	resp := createOrganizationResponse{
		Success: true,
		Message: "Organization created successfully",
		OrganizationID: orgID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/org/add-user
func AddUserToOrgHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req addUserToOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OrganizationID == "" || req.UserEmail == "" {
		http.Error(w, "organization ID and user email required", http.StatusBadRequest)
		return
	}

	// Set default role
	if req.Role == "" {
		req.Role = "member"
	}

	// Validate role
	if req.Role != "admin" && req.Role != "member" && req.Role != "viewer" {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	// Check if user is admin of the organization
	var isAdmin bool
	err := DB.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM organization_members 
		WHERE organization_id = ? AND user_id = ? AND role = 'admin' AND status = 'active'
	`, req.OrganizationID, userID).Scan(&isAdmin)

	if err != nil {
		log.Printf("ERROR checking admin status: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if !isAdmin {
		http.Error(w, "admin access required", http.StatusForbidden)
		return
	}

	// Find user by email
	var targetUserID string
	err = DB.QueryRow("SELECT id FROM users WHERE email = ?", req.UserEmail).Scan(&targetUserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		log.Printf("ERROR finding user: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Check if user is already in organization
	var existingMemberID string
	err = DB.QueryRow(`
		SELECT id FROM organization_members 
		WHERE organization_id = ? AND user_id = ?
	`, req.OrganizationID, targetUserID).Scan(&existingMemberID)

	if err == nil {
		http.Error(w, "user already in organization", http.StatusConflict)
		return
	} else if err != sql.ErrNoRows {
		log.Printf("ERROR checking existing membership: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Add user to organization
	memberID := uuid.New().String()
	now := time.Now().UTC()

	_, err = DB.Exec(`
		INSERT INTO organization_members (id, organization_id, user_id, role, invited_by, status, joined_at)
		VALUES (?, ?, ?, ?, ?, 'active', ?)
	`, memberID, req.OrganizationID, targetUserID, req.Role, userID, now)

	if err != nil {
		log.Printf("ERROR adding user to organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Update user's organization_id
	_, err = DB.Exec(`
		UPDATE users 
		SET organization_id = ?, updated_at = ?
		WHERE id = ?
	`, req.OrganizationID, now, targetUserID)

	if err != nil {
		log.Printf("ERROR updating user organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log user addition (non-sensitive)
	log.Printf("INFO user_added_to_org org_id=%s user_id=%s role=%s", req.OrganizationID, targetUserID, req.Role)

	resp := addUserToOrgResponse{
		Success: true,
		Message: "User added to organization successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/org/remove-user
func RemoveUserFromOrgHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req removeUserFromOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.OrganizationID == "" || req.UserID == "" {
		http.Error(w, "organization ID and user ID required", http.StatusBadRequest)
		return
	}

	// Check if user is admin of the organization
	var isAdmin bool
	err := DB.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM organization_members 
		WHERE organization_id = ? AND user_id = ? AND role = 'admin' AND status = 'active'
	`, req.OrganizationID, userID).Scan(&isAdmin)

	if err != nil {
		log.Printf("ERROR checking admin status: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if !isAdmin {
		http.Error(w, "admin access required", http.StatusForbidden)
		return
	}

	// Check if target user is in organization
	var memberID string
	err = DB.QueryRow(`
		SELECT id FROM organization_members 
		WHERE organization_id = ? AND user_id = ?
	`, req.OrganizationID, req.UserID).Scan(&memberID)

	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "user not in organization", http.StatusNotFound)
			return
		}
		log.Printf("ERROR checking user membership: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Remove user from organization
	_, err = DB.Exec("DELETE FROM organization_members WHERE id = ?", memberID)
	if err != nil {
		log.Printf("ERROR removing user from organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Update user's organization_id to NULL
	_, err = DB.Exec(`
		UPDATE users 
		SET organization_id = NULL, updated_at = ?
		WHERE id = ?
	`, time.Now().UTC(), req.UserID)

	if err != nil {
		log.Printf("ERROR updating user organization: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	// Log user removal (non-sensitive)
	log.Printf("INFO user_removed_from_org org_id=%s user_id=%s", req.OrganizationID, req.UserID)

	resp := removeUserFromOrgResponse{
		Success: true,
		Message: "User removed from organization successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// POST /api/org/list-users
func ListOrgUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session context
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	// Get organization ID from query parameter
	orgID := r.URL.Query().Get("organization_id")
	if orgID == "" {
		http.Error(w, "organization ID required", http.StatusBadRequest)
		return
	}

	// Check if user is member of the organization
	var isMember bool
	err := DB.QueryRow(`
		SELECT COUNT(*) > 0 
		FROM organization_members 
		WHERE organization_id = ? AND user_id = ? AND status = 'active'
	`, orgID, userID).Scan(&isMember)

	if err != nil {
		log.Printf("ERROR checking membership: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	if !isMember {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	// Get organization users
	rows, err := DB.Query(`
		SELECT u.id, u.email, om.role, om.status, om.joined_at
		FROM users u
		JOIN organization_members om ON u.id = om.user_id
		WHERE om.organization_id = ?
		ORDER BY om.joined_at ASC
	`, orgID)

	if err != nil {
		log.Printf("ERROR getting organization users: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}
	defer rows.Close()

	var users []OrgUser
	for rows.Next() {
		var user OrgUser
		var joinedAt time.Time
		err := rows.Scan(&user.ID, &user.Email, &user.Role, &user.Status, &joinedAt)
		if err != nil {
			log.Printf("ERROR scanning user: %v", err)
			continue
		}
		user.JoinedAt = joinedAt.Format(time.RFC3339)
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		log.Printf("ERROR iterating users: %v", err)
		http.Error(w, "database error", http.StatusServiceUnavailable)
		return
	}

	resp := listOrgUsersResponse{
		Success: true,
		Users:   users,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
