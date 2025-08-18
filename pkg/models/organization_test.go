package models

import (
	"database/sql"
	"testing"
	"time"

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

func TestCreateOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Test successful organization creation
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("CreateOrganization failed: %v", err)
	}

	if org.ID == "" {
		t.Error("Organization ID should not be empty")
	}
	if org.Name != "Test Organization" {
		t.Errorf("Expected name 'Test Organization', got '%s'", org.Name)
	}
	if org.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if org.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}

	// Test duplicate organization name
	_, err = CreateOrganization(db, "Test Organization")
	if err == nil {
		t.Error("Should fail when creating organization with duplicate name")
	}
}

func TestGetOrganizationByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Test successful retrieval
	retrievedOrg, err := GetOrganizationByID(db, org.ID)
	if err != nil {
		t.Fatalf("GetOrganizationByID failed: %v", err)
	}

	if retrievedOrg.ID != org.ID {
		t.Errorf("Expected ID %s, got %s", org.ID, retrievedOrg.ID)
	}
	if retrievedOrg.Name != org.Name {
		t.Errorf("Expected name %s, got %s", org.Name, retrievedOrg.Name)
	}

	// Test non-existent organization
	_, err = GetOrganizationByID(db, "non-existent-id")
	if err == nil {
		t.Error("Should fail when retrieving non-existent organization")
	}
}

func TestListOrganizations(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create multiple organizations
	org1, err := CreateOrganization(db, "Organization 1")
	if err != nil {
		t.Fatalf("Failed to create organization 1: %v", err)
	}

	org2, err := CreateOrganization(db, "Organization 2")
	if err != nil {
		t.Fatalf("Failed to create organization 2: %v", err)
	}

	// Test listing all organizations
	organizations, err := ListOrganizations(db)
	if err != nil {
		t.Fatalf("ListOrganizations failed: %v", err)
	}

	if len(organizations) != 2 {
		t.Errorf("Expected 2 organizations, got %d", len(organizations))
	}

	// Check that both organizations are in the list
	found1, found2 := false, false
	for _, org := range organizations {
		if org.ID == org1.ID {
			found1 = true
		}
		if org.ID == org2.ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("Organization 1 not found in list")
	}
	if !found2 {
		t.Error("Organization 2 not found in list")
	}
}

func TestUpdateOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Original Name")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	originalUpdatedAt := org.UpdatedAt

	// Wait a bit to ensure updated_at will be different
	time.Sleep(1 * time.Millisecond)

	// Test successful update
	updatedOrg, err := UpdateOrganization(db, org.ID, "Updated Name")
	if err != nil {
		t.Fatalf("UpdateOrganization failed: %v", err)
	}

	if updatedOrg.Name != "Updated Name" {
		t.Errorf("Expected name 'Updated Name', got '%s'", updatedOrg.Name)
	}
	if !updatedOrg.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated")
	}

	// Test update with non-existent organization
	_, err = UpdateOrganization(db, "non-existent-id", "New Name")
	if err == nil {
		t.Error("Should fail when updating non-existent organization")
	}
}

func TestDeleteOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Test successful deletion
	err = DeleteOrganization(db, org.ID)
	if err != nil {
		t.Fatalf("DeleteOrganization failed: %v", err)
	}

	// Verify organization is deleted
	_, err = GetOrganizationByID(db, org.ID)
	if err == nil {
		t.Error("Organization should be deleted")
	}

	// Test deletion of non-existent organization
	err = DeleteOrganization(db, "non-existent-id")
	if err == nil {
		t.Error("Should fail when deleting non-existent organization")
	}
}

func TestAddUserToOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Create test user
	userID := "test-user-id"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
		userID, "test@example.com", "password-hash", "totp-secret")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test successful user assignment
	err = AddUserToOrganization(db, userID, org.ID, RoleEnterpriseAdmin)
	if err != nil {
		t.Fatalf("AddUserToOrganization failed: %v", err)
	}

	// Verify user is assigned to organization
	var assignedOrgID, assignedRole string
	err = db.QueryRow("SELECT organization_id, role FROM users WHERE id = ?", userID).Scan(&assignedOrgID, &assignedRole)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if assignedOrgID != org.ID {
		t.Errorf("Expected organization ID %s, got %s", org.ID, assignedOrgID)
	}
	if assignedRole != string(RoleEnterpriseAdmin) {
		t.Errorf("Expected role %s, got %s", RoleEnterpriseAdmin, assignedRole)
	}
}

func TestRemoveUserFromOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Create test user
	userID := "test-user-id"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, organization_id, role) VALUES (?, ?, ?, ?, ?, ?)",
		userID, "test@example.com", "password-hash", "totp-secret", org.ID, RoleEnterpriseAdmin)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test successful user removal
	err = RemoveUserFromOrganization(db, userID)
	if err != nil {
		t.Fatalf("RemoveUserFromOrganization failed: %v", err)
	}

	// Verify user is removed from organization
	var assignedOrgID sql.NullString
	err = db.QueryRow("SELECT organization_id FROM users WHERE id = ?", userID).Scan(&assignedOrgID)
	if err != nil {
		t.Fatalf("Failed to query user: %v", err)
	}

	if assignedOrgID.Valid {
		t.Error("User should not be assigned to any organization")
	}
}

func TestGetOrganizationMembers(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Create test users
	user1ID := "user1-id"
	user2ID := "user2-id"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, organization_id, role) VALUES (?, ?, ?, ?, ?, ?)",
		user1ID, "user1@example.com", "password-hash", "totp-secret", org.ID, RoleEnterpriseAdmin)
	if err != nil {
		t.Fatalf("Failed to create user 1: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, organization_id, role) VALUES (?, ?, ?, ?, ?, ?)",
		user2ID, "user2@example.com", "password-hash", "totp-secret", org.ID, RoleEnterpriseUser)
	if err != nil {
		t.Fatalf("Failed to create user 2: %v", err)
	}

	// Test getting organization members
	members, err := GetOrganizationMembers(db, org.ID)
	if err != nil {
		t.Fatalf("GetOrganizationMembers failed: %v", err)
	}

	if len(members) != 2 {
		t.Errorf("Expected 2 members, got %d", len(members))
	}

	// Check that both users are in the list
	found1, found2 := false, false
	for _, member := range members {
		if member.UserID == user1ID {
			found1 = true
		}
		if member.UserID == user2ID {
			found2 = true
		}
	}

	if !found1 {
		t.Error("User 1 not found in members list")
	}
	if !found2 {
		t.Error("User 2 not found in members list")
	}
}

func TestGetUserPermissions(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organization
	org, err := CreateOrganization(db, "Test Organization")
	if err != nil {
		t.Fatalf("Failed to create test organization: %v", err)
	}

	// Create test user
	userID := "test-user-id"
	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, organization_id, role) VALUES (?, ?, ?, ?, ?, ?)",
		userID, "test@example.com", "password-hash", "totp-secret", org.ID, RoleEnterpriseAdmin)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test getting user permissions
	perms, err := GetUserPermissions(db, userID)
	if err != nil {
		t.Fatalf("GetUserPermissions failed: %v", err)
	}

	if perms.UserID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, perms.UserID)
	}
	if perms.OrganizationID != org.ID {
		t.Errorf("Expected organization ID %s, got %s", org.ID, perms.OrganizationID)
	}
	if perms.Role != RoleEnterpriseAdmin {
		t.Errorf("Expected role %s, got %s", RoleEnterpriseAdmin, perms.Role)
	}
}

func TestCanUserAccessOrganization(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create test organizations
	org1, err := CreateOrganization(db, "Organization 1")
	if err != nil {
		t.Fatalf("Failed to create organization 1: %v", err)
	}

	org2, err := CreateOrganization(db, "Organization 2")
	if err != nil {
		t.Fatalf("Failed to create organization 2: %v", err)
	}

	// Create test users
	systemAdminID := "system-admin-id"
	enterpriseAdminID := "enterprise-admin-id"
	enterpriseUserID := "enterprise-user-id"

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, role) VALUES (?, ?, ?, ?, ?)",
		systemAdminID, "system@example.com", "password-hash", "totp-secret", RoleSystemAdmin)
	if err != nil {
		t.Fatalf("Failed to create system admin: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, organization_id, role) VALUES (?, ?, ?, ?, ?, ?)",
		enterpriseAdminID, "admin@example.com", "password-hash", "totp-secret", org1.ID, RoleEnterpriseAdmin)
	if err != nil {
		t.Fatalf("Failed to create enterprise admin: %v", err)
	}

	_, err = db.Exec("INSERT INTO users (id, email, password_hash, totp_secret, organization_id, role) VALUES (?, ?, ?, ?, ?, ?)",
		enterpriseUserID, "user@example.com", "password-hash", "totp-secret", org1.ID, RoleEnterpriseUser)
	if err != nil {
		t.Fatalf("Failed to create enterprise user: %v", err)
	}

	// Test system admin access (should have access to all organizations)
	canAccess, err := CanUserAccessOrganization(db, systemAdminID, org1.ID)
	if err != nil {
		t.Fatalf("CanUserAccessOrganization failed for system admin: %v", err)
	}
	if !canAccess {
		t.Error("System admin should have access to all organizations")
	}

	canAccess, err = CanUserAccessOrganization(db, systemAdminID, org2.ID)
	if err != nil {
		t.Fatalf("CanUserAccessOrganization failed for system admin: %v", err)
	}
	if !canAccess {
		t.Error("System admin should have access to all organizations")
	}

	// Test enterprise admin access (should have access to own organization)
	canAccess, err = CanUserAccessOrganization(db, enterpriseAdminID, org1.ID)
	if err != nil {
		t.Fatalf("CanUserAccessOrganization failed for enterprise admin: %v", err)
	}
	if !canAccess {
		t.Error("Enterprise admin should have access to own organization")
	}

	canAccess, err = CanUserAccessOrganization(db, enterpriseAdminID, org2.ID)
	if err != nil {
		t.Fatalf("CanUserAccessOrganization failed for enterprise admin: %v", err)
	}
	if canAccess {
		t.Error("Enterprise admin should not have access to other organizations")
	}

	// Test enterprise user access (should have access to own organization)
	canAccess, err = CanUserAccessOrganization(db, enterpriseUserID, org1.ID)
	if err != nil {
		t.Fatalf("CanUserAccessOrganization failed for enterprise user: %v", err)
	}
	if !canAccess {
		t.Error("Enterprise user should have access to own organization")
	}

	canAccess, err = CanUserAccessOrganization(db, enterpriseUserID, org2.ID)
	if err != nil {
		t.Fatalf("CanUserAccessOrganization failed for enterprise user: %v", err)
	}
	if canAccess {
		t.Error("Enterprise user should not have access to other organizations")
	}
}

func TestUserRoleValidation(t *testing.T) {
	// Test valid roles
	validRoles := []UserRole{RoleSystemAdmin, RoleEnterpriseAdmin, RoleEnterpriseUser}
	for _, role := range validRoles {
		if !IsValidRole(role) {
			t.Errorf("Role %s should be valid", role)
		}
	}

	// Test invalid role
	invalidRole := UserRole("invalid_role")
	if IsValidRole(invalidRole) {
		t.Error("Invalid role should not be valid")
	}
}

func TestRoleHierarchy(t *testing.T) {
	// Test role hierarchy
	if !HasRole(RoleSystemAdmin, RoleSystemAdmin) {
		t.Error("System admin should have system admin role")
	}
	if !HasRole(RoleSystemAdmin, RoleEnterpriseAdmin) {
		t.Error("System admin should have enterprise admin role")
	}
	if !HasRole(RoleSystemAdmin, RoleEnterpriseUser) {
		t.Error("System admin should have enterprise user role")
	}

	if !HasRole(RoleEnterpriseAdmin, RoleEnterpriseAdmin) {
		t.Error("Enterprise admin should have enterprise admin role")
	}
	if !HasRole(RoleEnterpriseAdmin, RoleEnterpriseUser) {
		t.Error("Enterprise admin should have enterprise user role")
	}
	if HasRole(RoleEnterpriseAdmin, RoleSystemAdmin) {
		t.Error("Enterprise admin should not have system admin role")
	}

	if !HasRole(RoleEnterpriseUser, RoleEnterpriseUser) {
		t.Error("Enterprise user should have enterprise user role")
	}
	if HasRole(RoleEnterpriseUser, RoleEnterpriseAdmin) {
		t.Error("Enterprise user should not have enterprise admin role")
	}
	if HasRole(RoleEnterpriseUser, RoleSystemAdmin) {
		t.Error("Enterprise user should not have system admin role")
	}
}
