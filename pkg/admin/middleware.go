package admin

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// Context keys for storing admin information
type contextKey string

const (
	AdminIDKey    contextKey = "admin_id"
	AdminEmailKey contextKey = "admin_email"
	AdminRoleKey  contextKey = "admin_role"
)

// AdminMiddleware provides admin authentication and authorization
type AdminMiddleware struct {
	db *sql.DB
}

// NewAdminMiddleware creates a new admin middleware instance
func NewAdminMiddleware(db *sql.DB) *AdminMiddleware {
	return &AdminMiddleware{db: db}
}

// RequireAdminAuth ensures the request has a valid admin session
func (m *AdminMiddleware) RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract session token from Authorization header
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

		sessionToken := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate admin session
		adminUser, err := m.validateAdminSession(sessionToken)
		if err != nil {
			log.Printf("[ADMIN_MIDDLEWARE] Session validation failed: %v", err)
			http.Error(w, `{"error":"Invalid session"}`, http.StatusUnauthorized)
			return
		}

		// Add admin information to request context
		ctx := context.WithValue(r.Context(), AdminIDKey, adminUser.ID)
		ctx = context.WithValue(ctx, AdminEmailKey, adminUser.Email)
		ctx = context.WithValue(ctx, AdminRoleKey, adminUser.Role)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// RequireAdminRole ensures the admin has a specific role
func (m *AdminMiddleware) RequireAdminRole(requiredRole string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			adminRole := r.Context().Value(AdminRoleKey)
			if adminRole == nil {
				http.Error(w, `{"error":"Admin role not found in context"}`, http.StatusInternalServerError)
				return
			}

			role := adminRole.(string)
			if role != requiredRole {
				log.Printf("[ADMIN_MIDDLEWARE] Access denied: admin has role %s, required %s", role, requiredRole)
				http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// RequireRootAdmin ensures the admin is a root admin
func (m *AdminMiddleware) RequireRootAdmin(next http.HandlerFunc) http.HandlerFunc {
	return m.RequireAdminRole("root_admin")(next)
}

// RequireFullAdmin ensures the admin is a full admin or root admin
func (m *AdminMiddleware) RequireFullAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminRole := r.Context().Value(AdminRoleKey)
		if adminRole == nil {
			http.Error(w, `{"error":"Admin role not found in context"}`, http.StatusInternalServerError)
			return
		}

		role := adminRole.(string)
		if role != "root_admin" && role != "full_admin" {
			log.Printf("[ADMIN_MIDDLEWARE] Access denied: admin has role %s, required root_admin or full_admin", role)
			http.Error(w, `{"error":"Insufficient permissions"}`, http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}
}

// validateAdminSession validates an admin session token
func (m *AdminMiddleware) validateAdminSession(sessionToken string) (*AdminUser, error) {
	var admin AdminUser
	var sessionID string

	err := m.db.QueryRow(`
		SELECT s.id, a.id, a.email, a.role, a.totp_enabled, a.is_active, 
		       a.last_login, a.failed_login_attempts, a.locked_until, 
		       a.created_at, a.updated_at
		FROM admin_sessions s
		JOIN admin_users a ON s.admin_id = a.id
		WHERE s.session_token = ? AND s.is_active = 1 AND s.expires_at > ?
	`, sessionToken, time.Now()).Scan(
		&sessionID, &admin.ID, &admin.Email, &admin.Role, &admin.TOTPEnabled,
		&admin.IsActive, &admin.LastLogin, &admin.FailedLoginAttempts,
		&admin.LockedUntil, &admin.CreatedAt, &admin.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid or expired session")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return &admin, nil
}

// GetAdminFromContext extracts admin information from request context
func GetAdminFromContext(ctx context.Context) (adminID, adminEmail, adminRole string, err error) {
	adminIDVal := ctx.Value(AdminIDKey)
	adminEmailVal := ctx.Value(AdminEmailKey)
	adminRoleVal := ctx.Value(AdminRoleKey)

	if adminIDVal == nil || adminEmailVal == nil || adminRoleVal == nil {
		return "", "", "", fmt.Errorf("admin information not found in context")
	}

	adminID, ok := adminIDVal.(string)
	if !ok {
		return "", "", "", fmt.Errorf("invalid admin ID type in context")
	}

	adminEmail, ok = adminEmailVal.(string)
	if !ok {
		return "", "", "", fmt.Errorf("invalid admin email type in context")
	}

	adminRole, ok = adminRoleVal.(string)
	if !ok {
		return "", "", "", fmt.Errorf("invalid admin role type in context")
	}

	return adminID, adminEmail, adminRole, nil
}
