package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"secure-email-mvp/pkg/models"
)

// Context keys for storing user information
type contextKey string

const (
	UserIDKey         contextKey = "user_id"
	UserEmailKey      contextKey = "user_email"
	UserRoleKey       contextKey = "user_role"
	OrganizationIDKey contextKey = "organization_id"
	FilterOrgIDKey    contextKey = "filter_organization_id"
)

// RBACMiddleware provides role-based access control
type RBACMiddleware struct {
	db *sql.DB
}

// NewRBACMiddleware creates a new RBAC middleware instance
func NewRBACMiddleware(db *sql.DB) *RBACMiddleware {
	return &RBACMiddleware{db: db}
}

// RequireAuth ensures the request has a valid JWT token
func (m *RBACMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Bearer token required"}`, http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate JWT token
		userID, userEmail, err := ValidateJWT(tokenString)
		if err != nil {
			log.Printf("[RBAC] JWT validation failed: %v", err)
			http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Get user permissions from database
		perms, err := models.GetUserPermissions(m.db, userID)
		if err != nil {
			log.Printf("[RBAC] Failed to get user permissions for %s: %v", userID, err)
			http.Error(w, `{"error":"User not found"}`, http.StatusUnauthorized)
			return
		}

		// Add user information to request context
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserEmailKey, userEmail)
		ctx = context.WithValue(ctx, UserRoleKey, perms.Role)
		ctx = context.WithValue(ctx, OrganizationIDKey, perms.OrganizationID)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireRole ensures the user has a specific role
func (m *RBACMiddleware) RequireRole(requiredRole models.UserRole) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value(UserRoleKey)
			if userRole == nil {
				http.Error(w, `{"error":"User role not found in context"}`, http.StatusInternalServerError)
				return
			}

			role := userRole.(models.UserRole)
			if role != requiredRole {
				log.Printf("[RBAC] Access denied: user has role %s, required %s", role, requiredRole)
				http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireSystemAdmin ensures the user is a system admin
func (m *RBACMiddleware) RequireSystemAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole(models.RoleSystemAdmin)(next)
}

// RequireEnterpriseAdmin ensures the user is an enterprise admin
func (m *RBACMiddleware) RequireEnterpriseAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireRole(models.RoleEnterpriseAdmin)(next)
}

// RequireAdmin ensures the user is either a system admin or enterprise admin
func (m *RBACMiddleware) RequireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(UserRoleKey)
		if userRole == nil {
			http.Error(w, `{"error":"User role not found in context"}`, http.StatusInternalServerError)
			return
		}

		role := userRole.(models.UserRole)
		if role != models.RoleSystemAdmin && role != models.RoleEnterpriseAdmin {
			log.Printf("[RBAC] Access denied: user has role %s, requires admin role", role)
			http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// RequireOrganizationAccess ensures the user can access the specified organization
func (m *RBACMiddleware) RequireOrganizationAccess(organizationID string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userID := r.Context().Value(UserIDKey)
			if userID == nil {
				http.Error(w, `{"error":"User ID not found in context"}`, http.StatusInternalServerError)
				return
			}

			// Check if user can access the organization
			canAccess, err := models.CanUserAccessOrganization(m.db, userID.(string), organizationID)
			if err != nil {
				log.Printf("[RBAC] Failed to check organization access: %v", err)
				http.Error(w, `{"error":"Failed to verify permissions"}`, http.StatusInternalServerError)
				return
			}

			if !canAccess {
				log.Printf("[RBAC] Access denied: user %s cannot access organization %s", userID, organizationID)
				http.Error(w, `{"error":"Access denied to organization"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireEnterpriseMultiTenancy ensures enterprise multi-tenancy is enabled
func (m *RBACMiddleware) RequireEnterpriseMultiTenancy(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enabled := os.Getenv("ENABLE_ENTERPRISE_MULTI_TENANCY") == "true"
		if !enabled {
			http.Error(w, `{"error":"Enterprise multi-tenancy is disabled"}`, http.StatusServiceUnavailable)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// ScopeToOrganization ensures responses are scoped to the user's organization
func (m *RBACMiddleware) ScopeToOrganization(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userRole := r.Context().Value(UserRoleKey)
		userOrgID := r.Context().Value(OrganizationIDKey)

		if userRole == nil || userOrgID == nil {
			http.Error(w, `{"error":"User context not found"}`, http.StatusInternalServerError)
			return
		}

		role := userRole.(models.UserRole)
		orgID := userOrgID.(string)

		// System admins can see all data, so no scoping needed
		if role == models.RoleSystemAdmin {
			next.ServeHTTP(w, r)
			return
		}

		// For enterprise admins and users, scope to their organization
		// Add organization filter to request context
		ctx := context.WithValue(r.Context(), FilterOrgIDKey, orgID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext extracts user information from request context
func GetUserFromContext(ctx context.Context) (userID, userEmail string, userRole models.UserRole, organizationID string, err error) {
	userIDVal := ctx.Value(UserIDKey)
	userEmailVal := ctx.Value(UserEmailKey)
	userRoleVal := ctx.Value(UserRoleKey)
	orgIDVal := ctx.Value(OrganizationIDKey)

	if userIDVal == nil || userEmailVal == nil || userRoleVal == nil || orgIDVal == nil {
		return "", "", "", "", fmt.Errorf("user context not found")
	}

	userID = userIDVal.(string)
	userEmail = userEmailVal.(string)
	userRole = userRoleVal.(models.UserRole)
	organizationID = orgIDVal.(string)

	return userID, userEmail, userRole, organizationID, nil
}

// GetFilterOrganizationID gets the organization ID for filtering from context
func GetFilterOrganizationID(ctx context.Context) (string, bool) {
	orgID := ctx.Value(FilterOrgIDKey)
	if orgID == nil {
		return "", false
	}
	return orgID.(string), true
}

// LogAccess logs access attempts for audit purposes
func (m *RBACMiddleware) LogAccess(r *http.Request, action string, resource string) {
	userID := r.Context().Value(UserIDKey)
	userEmail := r.Context().Value(UserEmailKey)
	userRole := r.Context().Value(UserRoleKey)
	orgID := r.Context().Value(OrganizationIDKey)

	if userID != nil {
		log.Printf("[RBAC_AUDIT] %s | User: %s (%s) | Role: %s | Org: %s | Resource: %s | IP: %s",
			action,
			userEmail,
			userID,
			userRole,
			orgID,
			resource,
			r.RemoteAddr,
		)
	}
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// WriteErrorResponse writes a standardized error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, error, code, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Code:    code,
		Details: details,
	}

	json.NewEncoder(w).Encode(response)
}

// WriteSuccessResponse writes a standardized success response
func WriteSuccessResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
