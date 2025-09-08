package auth

// DO NOT EDIT EXISTING CODE - new file added
// Role-based access control middleware for Enterprise features

import (
	"database/sql"
	"net/http"
)

// RequireAccountType middleware ensures user has required account type
func RequireAccountType(requiredTypes ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from session context
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Get user's account type
			var accountType string
			err := DB.QueryRow("SELECT account_type_new FROM users WHERE id = ?", userID).Scan(&accountType)
			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "user not found", http.StatusNotFound)
					return
				}
				http.Error(w, "database error", http.StatusServiceUnavailable)
				return
			}

			// Check if user has required account type
			hasRequiredType := false
			for _, requiredType := range requiredTypes {
				if accountType == requiredType {
					hasRequiredType = true
					break
				}
			}

			if !hasRequiredType {
				http.Error(w, "account type not authorized", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireOrganizationRole middleware ensures user has required organization role
func RequireOrganizationRole(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from session context
			userID, ok := GetUserIDFromContext(r.Context())
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Get user's organization role
			var role sql.NullString
			err := DB.QueryRow(`
				SELECT om.role 
				FROM organization_members om
				WHERE om.user_id = ? AND om.status = 'active'
			`, userID).Scan(&role)

			if err != nil {
				if err == sql.ErrNoRows {
					http.Error(w, "not a member of any organization", http.StatusForbidden)
					return
				}
				http.Error(w, "database error", http.StatusServiceUnavailable)
				return
			}

			if !role.Valid {
				http.Error(w, "no organization role assigned", http.StatusForbidden)
				return
			}

			// Check if user has required role
			hasRequiredRole := false
			for _, requiredRole := range requiredRoles {
				if role.String == requiredRole {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				http.Error(w, "insufficient organization role", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePremiumOrEnterprise middleware for domain management features
func RequirePremiumOrEnterprise(next http.Handler) http.Handler {
	return RequireAccountType("premium", "enterprise")(next)
}

// RequireEnterprise middleware for organization management features
func RequireEnterprise(next http.Handler) http.Handler {
	return RequireAccountType("enterprise")(next)
}

// RequireOrganizationAdmin middleware for admin-only organization features
func RequireOrganizationAdmin(next http.Handler) http.Handler {
	return RequireOrganizationRole("admin")(next)
}
