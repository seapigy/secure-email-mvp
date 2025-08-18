package auth

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"secure-email-mvp/pkg/models"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create organizations table
	_, err = db.Exec(`
		CREATE TABLE organizations (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create organizations table: %v", err)
	}

	// Create users table with organization support
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			totp_secret TEXT NOT NULL,
			organization_id TEXT,
			role TEXT CHECK (role IN ('system_admin', 'enterprise_admin', 'enterprise_user')) DEFAULT 'enterprise_user',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (organization_id) REFERENCES organizations(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	return db
}

func createTestUser(db *sql.DB, userID, email, role, organizationID string) error {
	query := `INSERT INTO users (id, email, password_hash, totp_secret, role, organization_id) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(query, userID, email, "password-hash", "totp-secret", role, organizationID)
	return err
}

func createTestOrganization(db *sql.DB, orgID, name string) error {
	query := `INSERT INTO organizations (id, name) VALUES (?, ?)`
	_, err := db.Exec(query, orgID, name)
	return err
}

func TestRequireAuthMissingHeader(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rbac := NewRBACMiddleware(db)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	rbac.RequireAuth(handler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Authorization header required") {
		t.Error("Response should contain authorization header error")
	}
}

func TestRequireAuthInvalidToken(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rbac := NewRBACMiddleware(db)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	rbac.RequireAuth(handler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Handler should not have been called")
	}
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Invalid token") {
		t.Error("Response should contain invalid token error")
	}
}

func TestGetUserFromContext(t *testing.T) {
	// Test with valid context
	ctx := context.WithValue(context.Background(), UserIDKey, "test-user-id")
	ctx = context.WithValue(ctx, UserEmailKey, "test@example.com")
	ctx = context.WithValue(ctx, UserRoleKey, models.RoleEnterpriseAdmin)
	ctx = context.WithValue(ctx, OrganizationIDKey, "org-1")

	userID, userEmail, userRole, orgID, err := GetUserFromContext(ctx)
	if err != nil {
		t.Fatalf("Failed to get user from context: %v", err)
	}

	if userID != "test-user-id" {
		t.Errorf("Expected user ID test-user-id, got %s", userID)
	}
	if userEmail != "test@example.com" {
		t.Errorf("Expected user email test@example.com, got %s", userEmail)
	}
	if userRole != models.RoleEnterpriseAdmin {
		t.Errorf("Expected user role enterprise_admin, got %s", userRole)
	}
	if orgID != "org-1" {
		t.Errorf("Expected organization ID org-1, got %s", orgID)
	}
}

func TestGetUserFromContextEmpty(t *testing.T) {
	// Test with empty context
	ctx := context.Background()

	userID, userEmail, userRole, orgID, err := GetUserFromContext(ctx)
	// We expect an error for empty context
	if err == nil {
		t.Error("Expected error for empty context")
	}

	if userID != "" {
		t.Errorf("Expected empty user ID, got %s", userID)
	}
	if userEmail != "" {
		t.Errorf("Expected empty user email, got %s", userEmail)
	}
	if userRole != "" {
		t.Errorf("Expected empty user role, got %s", userRole)
	}
	if orgID != "" {
		t.Errorf("Expected empty organization ID, got %s", orgID)
	}
}

func TestLogAccess(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rbac := NewRBACMiddleware(db)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	req.RemoteAddr = "127.0.0.1:12345"

	// This should not panic and should log the access
	rbac.LogAccess(req, "TEST_ACTION", "test-resource")
}

func TestRequireEnterpriseMultiTenancy(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	rbac := NewRBACMiddleware(db)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
	})

	// Test with feature disabled (should fail)
	rbac.RequireEnterpriseMultiTenancy(handler).ServeHTTP(w, req)

	if handlerCalled {
		t.Error("Handler should not have been called")
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Enterprise multi-tenancy is disabled") {
		t.Error("Response should contain feature disabled error")
	}
}

func TestRoleHierarchy(t *testing.T) {
	// Test role hierarchy
	if !models.HasRole(models.RoleSystemAdmin, models.RoleSystemAdmin) {
		t.Error("System admin should have system admin role")
	}
	if !models.HasRole(models.RoleSystemAdmin, models.RoleEnterpriseAdmin) {
		t.Error("System admin should have enterprise admin role")
	}
	if !models.HasRole(models.RoleSystemAdmin, models.RoleEnterpriseUser) {
		t.Error("System admin should have enterprise user role")
	}

	if !models.HasRole(models.RoleEnterpriseAdmin, models.RoleEnterpriseAdmin) {
		t.Error("Enterprise admin should have enterprise admin role")
	}
	if !models.HasRole(models.RoleEnterpriseAdmin, models.RoleEnterpriseUser) {
		t.Error("Enterprise admin should have enterprise user role")
	}
	if models.HasRole(models.RoleEnterpriseAdmin, models.RoleSystemAdmin) {
		t.Error("Enterprise admin should not have system admin role")
	}

	if !models.HasRole(models.RoleEnterpriseUser, models.RoleEnterpriseUser) {
		t.Error("Enterprise user should have enterprise user role")
	}
	if models.HasRole(models.RoleEnterpriseUser, models.RoleEnterpriseAdmin) {
		t.Error("Enterprise user should not have enterprise admin role")
	}
	if models.HasRole(models.RoleEnterpriseUser, models.RoleSystemAdmin) {
		t.Error("Enterprise user should not have system admin role")
	}
}

func TestUserRoleValidation(t *testing.T) {
	// Test valid roles
	validRoles := []models.UserRole{models.RoleSystemAdmin, models.RoleEnterpriseAdmin, models.RoleEnterpriseUser}
	for _, role := range validRoles {
		if !models.IsValidRole(role) {
			t.Errorf("Role %s should be valid", role)
		}
	}

	// Test invalid role
	invalidRole := models.UserRole("invalid_role")
	if models.IsValidRole(invalidRole) {
		t.Error("Invalid role should not be valid")
	}
}
