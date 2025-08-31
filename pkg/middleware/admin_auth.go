package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dgrijalva/jwt-go"
)

// Context keys for storing admin information
type contextKey string

const (
	AdminIDKey    contextKey = "admin_id"
	AdminEmailKey contextKey = "admin_email"
	AdminRoleKey  contextKey = "admin_role"
)

// AdminAuthMiddleware provides admin authentication and authorization
type AdminAuthMiddleware struct{}

// NewAdminAuthMiddleware creates a new admin auth middleware
func NewAdminAuthMiddleware() *AdminAuthMiddleware {
	return &AdminAuthMiddleware{}
}

// RequireAdminAuth ensures the request has a valid admin JWT token
func (m *AdminAuthMiddleware) RequireAdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract JWT from Authorization header
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
		claims, err := m.validateJWT(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error":"Invalid token: %v"}`, err), http.StatusUnauthorized)
			return
		}

		// Verify role is admin
		role, ok := claims["role"].(string)
		if !ok || role != "admin" {
			http.Error(w, `{"error":"Admin role required"}`, http.StatusUnauthorized)
			return
		}

		// Extract admin information
		email, _ := claims["email"].(string)
		sub, _ := claims["sub"].(string)
		
		// Extract admin ID from subject (format: "admin:123")
		var adminID string
		if strings.HasPrefix(sub, "admin:") {
			adminID = strings.TrimPrefix(sub, "admin:")
		}

		// Add admin information to request context
		ctx := context.WithValue(r.Context(), AdminIDKey, adminID)
		ctx = context.WithValue(ctx, AdminEmailKey, email)
		ctx = context.WithValue(ctx, AdminRoleKey, role)

		// Call next handler with enriched context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// validateJWT validates and parses a JWT token
func (m *AdminAuthMiddleware) validateJWT(tokenString string) (jwt.MapClaims, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET not configured")
	}

	// Parse and validate token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	return claims, nil
}

// GetAdminIDFromContext extracts admin ID from request context
func GetAdminIDFromContext(ctx context.Context) string {
	if adminID, ok := ctx.Value(AdminIDKey).(string); ok {
		return adminID
	}
	return ""
}

// GetAdminEmailFromContext extracts admin email from request context
func GetAdminEmailFromContext(ctx context.Context) string {
	if adminEmail, ok := ctx.Value(AdminEmailKey).(string); ok {
		return adminEmail
	}
	return ""
}

// GetAdminRoleFromContext extracts admin role from request context
func GetAdminRoleFromContext(ctx context.Context) string {
	if adminRole, ok := ctx.Value(AdminRoleKey).(string); ok {
		return adminRole
	}
	return ""
}






