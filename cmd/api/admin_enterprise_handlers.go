package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/models"
)

// OrganizationRequest represents the request body for creating/updating organizations
type OrganizationRequest struct {
	Name string `json:"name"`
}

// OrganizationResponse represents the response for organization operations
type OrganizationResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// OrganizationListResponse represents the response for listing organizations
type OrganizationListResponse struct {
	Organizations []*OrganizationResponse `json:"organizations"`
	Total         int                     `json:"total"`
}

// OrganizationDetailResponse represents the detailed response for a single organization
type OrganizationDetailResponse struct {
	Organization *OrganizationResponse        `json:"organization"`
	Members      []*models.OrganizationMember `json:"members"`
	MemberCount  int                          `json:"member_count"`
}

// UserAssignmentRequest represents the request body for assigning users to organizations
type UserAssignmentRequest struct {
	Email string          `json:"email"`
	Role  models.UserRole `json:"role"`
}

// UserAssignmentResponse represents the response for user assignment operations
type UserAssignmentResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id,omitempty"`
}

// createOrganizationHandler creates a new organization (system_admin only)
// TODO: Wire up to routes in future micro-iteration
func createOrganizationHandler(db *sql.DB) http.HandlerFunc { //nolint:unused
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req OrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", err.Error())
			return
		}

		// Validate organization name
		if req.Name == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization name is required", "MISSING_NAME", "")
			return
		}

		if len(req.Name) > 100 {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization name too long", "NAME_TOO_LONG", "Maximum 100 characters")
			return
		}

		// Create organization
		org, err := models.CreateOrganization(db, req.Name)
		if err != nil {
			log.Printf("Failed to create organization: %v", err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to create organization", "CREATE_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "CREATE_ORGANIZATION", "organizations")

		// Return response
		response := &OrganizationResponse{
			ID:        org.ID,
			Name:      org.Name,
			CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		auth.WriteSuccessResponse(w, response)
	}
}

// listOrganizationsHandler lists all organizations (system_admin only)
// TODO: Wire up to routes in future micro-iteration
func listOrganizationsHandler(db *sql.DB) http.HandlerFunc { //nolint:unused
	return func(w http.ResponseWriter, r *http.Request) {
		// Get query parameters for pagination (future enhancement)
		_ = r.URL.Query().Get("page")
		_ = r.URL.Query().Get("limit")

		// List organizations
		organizations, err := models.ListOrganizations(db)
		if err != nil {
			log.Printf("Failed to list organizations: %v", err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to list organizations", "LIST_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "LIST_ORGANIZATIONS", "organizations")

		// Convert to response format
		var orgResponses []*OrganizationResponse
		for _, org := range organizations {
			orgResponses = append(orgResponses, &OrganizationResponse{
				ID:        org.ID,
				Name:      org.Name,
				CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt: org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			})
		}

		response := &OrganizationListResponse{
			Organizations: orgResponses,
			Total:         len(orgResponses),
		}

		auth.WriteSuccessResponse(w, response)
	}
}

// getOrganizationHandler gets a single organization with its members
func getOrganizationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from URL
		orgID := r.URL.Query().Get("id")
		if orgID == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization ID is required", "MISSING_ID", "")
			return
		}

		// Get organization details
		org, err := models.GetOrganizationByID(db, orgID)
		if err != nil {
			log.Printf("Failed to get organization %s: %v", orgID, err)
			auth.WriteErrorResponse(w, http.StatusNotFound, "Organization not found", "NOT_FOUND", err.Error())
			return
		}

		// Get organization members
		members, err := models.GetOrganizationMembers(db, orgID)
		if err != nil {
			log.Printf("Failed to get organization members for %s: %v", orgID, err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get organization members", "MEMBERS_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "GET_ORGANIZATION", "organizations/"+orgID)

		// Return response
		orgResponse := &OrganizationResponse{
			ID:        org.ID,
			Name:      org.Name,
			CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		response := &OrganizationDetailResponse{
			Organization: orgResponse,
			Members:      members,
			MemberCount:  len(members),
		}

		auth.WriteSuccessResponse(w, response)
	}
}

// addUserToOrganizationHandler adds a user to an organization
func addUserToOrganizationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from URL
		orgID := r.URL.Query().Get("id")
		if orgID == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization ID is required", "MISSING_ID", "")
			return
		}

		// Parse request body
		var req UserAssignmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", err.Error())
			return
		}

		// Validate email
		if req.Email == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "User email is required", "MISSING_EMAIL", "")
			return
		}

		// Validate role
		if !models.ValidateRole(string(req.Role)) {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Invalid role", "INVALID_ROLE", "Must be system_admin, enterprise_admin, or enterprise_user")
			return
		}

		// Get current user context
		userID, _, userRole, _, err := auth.GetUserFromContext(r.Context())
		if err != nil {
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user context", "CONTEXT_FAILED", err.Error())
			return
		}

		// Check permissions
		// System admins can assign any role to any organization
		// Enterprise admins can only assign enterprise_user and enterprise_admin roles to their own organization
		if userRole == models.RoleEnterpriseAdmin {
			// Check if user is trying to assign to their own organization
			canAccess, err := models.CanUserAccessOrganization(db, userID, orgID)
			if err != nil {
				auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to verify permissions", "PERMISSION_CHECK_FAILED", err.Error())
				return
			}
			if !canAccess {
				auth.WriteErrorResponse(w, http.StatusForbidden, "Cannot assign users to other organizations", "ACCESS_DENIED", "")
				return
			}

			// Enterprise admins cannot assign system_admin role
			if req.Role == models.RoleSystemAdmin {
				auth.WriteErrorResponse(w, http.StatusForbidden, "Cannot assign system_admin role", "ROLE_DENIED", "")
				return
			}
		}

		// Add user to organization
		err = models.AddUserToOrganization(db, orgID, req.Email, req.Role)
		if err != nil {
			log.Printf("Failed to add user %s to organization %s: %v", req.Email, orgID, err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to add user to organization", "ASSIGNMENT_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "ADD_USER_TO_ORG", "organizations/"+orgID+"/users")

		// Return response
		response := &UserAssignmentResponse{
			Message: "User added to organization successfully",
		}

		auth.WriteSuccessResponse(w, response)
	}
}

// removeUserFromOrganizationHandler removes a user from an organization
func removeUserFromOrganizationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract user email from URL
		userEmail := r.URL.Query().Get("email")
		if userEmail == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "User email is required", "MISSING_EMAIL", "")
			return
		}

		// Get current user context
		userID, _, userRole, _, err := auth.GetUserFromContext(r.Context())
		if err != nil {
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user context", "CONTEXT_FAILED", err.Error())
			return
		}

		// Check permissions
		// System admins can remove any user
		// Enterprise admins can only remove users from their own organization
		if userRole == models.RoleEnterpriseAdmin {
			// Get the user's organization to check if it matches
			perms, err := models.GetUserPermissions(db, userID)
			if err != nil {
				auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get user permissions", "PERMISSION_CHECK_FAILED", err.Error())
				return
			}

			// Check if the user being removed is in the same organization
			userToRemovePerms, err := models.GetUserPermissions(db, userEmail) // Note: this should be user ID, not email
			if err != nil {
				auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get target user permissions", "TARGET_USER_CHECK_FAILED", err.Error())
				return
			}

			if perms.OrganizationID != userToRemovePerms.OrganizationID {
				auth.WriteErrorResponse(w, http.StatusForbidden, "Cannot remove users from other organizations", "ACCESS_DENIED", "")
				return
			}
		}

		// Remove user from organization
		err = models.RemoveUserFromOrganization(db, userEmail)
		if err != nil {
			log.Printf("Failed to remove user %s from organization: %v", userEmail, err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to remove user from organization", "REMOVAL_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "REMOVE_USER_FROM_ORG", "organizations/users/"+userEmail)

		// Return response
		response := &UserAssignmentResponse{
			Message: "User removed from organization successfully",
		}

		auth.WriteSuccessResponse(w, response)
	}
}

// updateOrganizationHandler updates an organization
func updateOrganizationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from URL
		orgID := r.URL.Query().Get("id")
		if orgID == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization ID is required", "MISSING_ID", "")
			return
		}

		// Parse request body
		var req OrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST", err.Error())
			return
		}

		// Validate organization name
		if req.Name == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization name is required", "MISSING_NAME", "")
			return
		}

		if len(req.Name) > 100 {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization name too long", "NAME_TOO_LONG", "Maximum 100 characters")
			return
		}

		// Update organization
		org, err := models.UpdateOrganization(db, orgID, req.Name)
		if err != nil {
			log.Printf("Failed to update organization %s: %v", orgID, err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update organization", "UPDATE_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "UPDATE_ORGANIZATION", "organizations/"+orgID)

		// Return response
		response := &OrganizationResponse{
			ID:        org.ID,
			Name:      org.Name,
			CreatedAt: org.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt: org.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		auth.WriteSuccessResponse(w, response)
	}
}

// deleteOrganizationHandler deletes an organization
func deleteOrganizationHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract organization ID from URL
		orgID := r.URL.Query().Get("id")
		if orgID == "" {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Organization ID is required", "MISSING_ID", "")
			return
		}

		// Check if it's the default organization
		if orgID == models.GetDefaultOrganizationID() {
			auth.WriteErrorResponse(w, http.StatusBadRequest, "Cannot delete default organization", "DELETE_DENIED", "")
			return
		}

		// Delete organization
		err := models.DeleteOrganization(db, orgID)
		if err != nil {
			log.Printf("Failed to delete organization %s: %v", orgID, err)
			auth.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to delete organization", "DELETE_FAILED", err.Error())
			return
		}

		// Log the action
		rbac := auth.NewRBACMiddleware(db)
		rbac.LogAccess(r, "DELETE_ORGANIZATION", "organizations/"+orgID)

		// Return response
		response := map[string]string{
			"message": "Organization deleted successfully",
		}

		auth.WriteSuccessResponse(w, response)
	}
}
