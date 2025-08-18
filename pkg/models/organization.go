package models

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleSystemAdmin     UserRole = "system_admin"
	RoleEnterpriseAdmin UserRole = "enterprise_admin"
	RoleEnterpriseUser  UserRole = "enterprise_user"
)

// Organization represents an enterprise organization
type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// OrganizationMember represents a user's membership in an organization
type OrganizationMember struct {
	UserID           string    `json:"user_id"`
	Email            string    `json:"email"`
	Role             UserRole  `json:"role"`
	OrganizationID   string    `json:"organization_id"`
	OrganizationName string    `json:"organization_name"`
	UserCreatedAt    time.Time `json:"user_created_at"`
	OrgCreatedAt     time.Time `json:"organization_created_at"`
}

// AdminPermissions represents admin permissions for a user
type AdminPermissions struct {
	UserID                    string   `json:"user_id"`
	Email                     string   `json:"email"`
	Role                      UserRole `json:"role"`
	OrganizationID            string   `json:"organization_id"`
	OrganizationName          string   `json:"organization_name"`
	CanManageOrganizations    bool     `json:"can_manage_organizations"`
	CanManageAllOrganizations bool     `json:"can_manage_all_organizations"`
	CanManageUsers            bool     `json:"can_manage_users"`
}

// CreateOrganization creates a new organization
func CreateOrganization(db *sql.DB, name string) (*Organization, error) {
	id := uuid.New().String()

	query := `INSERT INTO organizations (id, name) VALUES (?, ?)`
	_, err := db.Exec(query, id, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %v", err)
	}

	// Fetch the created organization
	return GetOrganizationByID(db, id)
}

// GetOrganizationByID retrieves an organization by ID
func GetOrganizationByID(db *sql.DB, id string) (*Organization, error) {
	query := `SELECT id, name, created_at, updated_at FROM organizations WHERE id = ?`

	var org Organization
	err := db.QueryRow(query, id).Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %v", err)
	}

	return &org, nil
}

// GetOrganizationByName retrieves an organization by name
func GetOrganizationByName(db *sql.DB, name string) (*Organization, error) {
	query := `SELECT id, name, created_at, updated_at FROM organizations WHERE name = ?`

	var org Organization
	err := db.QueryRow(query, name).Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("organization not found")
		}
		return nil, fmt.Errorf("failed to get organization: %v", err)
	}

	return &org, nil
}

// ListOrganizations retrieves all organizations
func ListOrganizations(db *sql.DB) ([]*Organization, error) {
	query := `SELECT id, name, created_at, updated_at FROM organizations ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %v", err)
	}
	defer rows.Close()

	var organizations []*Organization
	for rows.Next() {
		var org Organization
		err := rows.Scan(&org.ID, &org.Name, &org.CreatedAt, &org.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %v", err)
		}
		organizations = append(organizations, &org)
	}

	return organizations, nil
}

// UpdateOrganization updates an organization
func UpdateOrganization(db *sql.DB, id, name string) (*Organization, error) {
	query := `UPDATE organizations SET name = ? WHERE id = ?`
	_, err := db.Exec(query, name, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update organization: %v", err)
	}

	// Fetch the updated organization
	return GetOrganizationByID(db, id)
}

// DeleteOrganization deletes an organization
func DeleteOrganization(db *sql.DB, id string) error {
	// First, check if there are any users in this organization
	var count int
	query := `SELECT COUNT(*) FROM users WHERE organization_id = ?`
	err := db.QueryRow(query, id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check organization users: %v", err)
	}

	if count > 0 {
		return fmt.Errorf("cannot delete organization with %d users", count)
	}

	// Delete the organization
	query = `DELETE FROM organizations WHERE id = ?`
	_, err = db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %v", err)
	}

	return nil
}

// GetOrganizationMembers retrieves all members of an organization
func GetOrganizationMembers(db *sql.DB, organizationID string) ([]*OrganizationMember, error) {
	query := `
		SELECT 
			u.id as user_id,
			u.email,
			u.role,
			o.id as organization_id,
			o.name as organization_name,
			u.created_at as user_created_at,
			o.created_at as organization_created_at
		FROM users u
		LEFT JOIN organizations o ON u.organization_id = o.id
		WHERE u.organization_id = ?
		ORDER BY u.email
	`

	rows, err := db.Query(query, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization members: %v", err)
	}
	defer rows.Close()

	var members []*OrganizationMember
	for rows.Next() {
		var member OrganizationMember
		err := rows.Scan(
			&member.UserID,
			&member.Email,
			&member.Role,
			&member.OrganizationID,
			&member.OrganizationName,
			&member.UserCreatedAt,
			&member.OrgCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization member: %v", err)
		}
		members = append(members, &member)
	}

	return members, nil
}

// AddUserToOrganization adds a user to an organization with a specific role
func AddUserToOrganization(db *sql.DB, organizationID, userEmail string, role UserRole) error {
	// First, check if the user exists
	var userID string
	query := `SELECT id FROM users WHERE email = ?`
	err := db.QueryRow(query, userEmail).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("user not found: %s", userEmail)
		}
		return fmt.Errorf("failed to check user: %v", err)
	}

	// Check if the organization exists
	query = `SELECT id FROM organizations WHERE id = ?`
	err = db.QueryRow(query, organizationID).Scan(&organizationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("organization not found: %s", organizationID)
		}
		return fmt.Errorf("failed to check organization: %v", err)
	}

	// Update the user's organization and role
	query = `UPDATE users SET organization_id = ?, role = ? WHERE id = ?`
	_, err = db.Exec(query, organizationID, role, userID)
	if err != nil {
		return fmt.Errorf("failed to add user to organization: %v", err)
	}

	log.Printf("Added user %s to organization %s with role %s", userEmail, organizationID, role)
	return nil
}

// RemoveUserFromOrganization removes a user from an organization
func RemoveUserFromOrganization(db *sql.DB, userEmail string) error {
	// Update the user to have no organization and default role
	query := `UPDATE users SET organization_id = 'system-default', role = 'enterprise_user' WHERE email = ?`
	result, err := db.Exec(query, userEmail)
	if err != nil {
		return fmt.Errorf("failed to remove user from organization: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found: %s", userEmail)
	}

	log.Printf("Removed user %s from organization", userEmail)
	return nil
}

// GetUserPermissions retrieves admin permissions for a user
func GetUserPermissions(db *sql.DB, userID string) (*AdminPermissions, error) {
	query := `
		SELECT 
			u.id as user_id,
			u.email,
			u.role,
			u.organization_id,
			o.name as organization_name,
			CASE 
				WHEN u.role = 'system_admin' THEN 1
				WHEN u.role = 'enterprise_admin' THEN 1
				ELSE 0
			END as can_manage_organizations,
			CASE 
				WHEN u.role = 'system_admin' THEN 1
				ELSE 0
			END as can_manage_all_organizations,
			CASE 
				WHEN u.role = 'system_admin' THEN 1
				WHEN u.role = 'enterprise_admin' THEN 1
				ELSE 0
			END as can_manage_users
		FROM users u
		LEFT JOIN organizations o ON u.organization_id = o.id
		WHERE u.id = ?
	`

	var perms AdminPermissions
	err := db.QueryRow(query, userID).Scan(
		&perms.UserID,
		&perms.Email,
		&perms.Role,
		&perms.OrganizationID,
		&perms.OrganizationName,
		&perms.CanManageOrganizations,
		&perms.CanManageAllOrganizations,
		&perms.CanManageUsers,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user permissions: %v", err)
	}

	return &perms, nil
}

// CanUserAccessOrganization checks if a user can access a specific organization
func CanUserAccessOrganization(db *sql.DB, userID, organizationID string) (bool, error) {
	perms, err := GetUserPermissions(db, userID)
	if err != nil {
		return false, err
	}

	// System admins can access all organizations
	if perms.Role == RoleSystemAdmin {
		return true, nil
	}

	// Enterprise admins can only access their own organization
	if perms.Role == RoleEnterpriseAdmin {
		return perms.OrganizationID == organizationID, nil
	}

	// Enterprise users can only access their own organization
	return perms.OrganizationID == organizationID, nil
}

// ValidateRole validates if a role is valid
func ValidateRole(role string) bool {
	switch UserRole(role) {
	case RoleSystemAdmin, RoleEnterpriseAdmin, RoleEnterpriseUser:
		return true
	default:
		return false
	}
}

// GetDefaultOrganizationID returns the default organization ID
func GetDefaultOrganizationID() string {
	return "system-default"
}

// GetDefaultRole returns the default user role
func GetDefaultRole() UserRole {
	return RoleEnterpriseUser
}

// IsValidRole checks if a role is valid
func IsValidRole(role UserRole) bool {
	switch role {
	case RoleSystemAdmin, RoleEnterpriseAdmin, RoleEnterpriseUser:
		return true
	default:
		return false
	}
}

// HasRole checks if a user role has the specified permission level
func HasRole(userRole UserRole, requiredRole UserRole) bool {
	// Define role hierarchy (higher index = higher privilege)
	roleHierarchy := map[UserRole]int{
		RoleEnterpriseUser:  0,
		RoleEnterpriseAdmin: 1,
		RoleSystemAdmin:     2,
	}

	userLevel, userExists := roleHierarchy[userRole]
	requiredLevel, requiredExists := roleHierarchy[requiredRole]

	if !userExists || !requiredExists {
		return false
	}

	return userLevel >= requiredLevel
}
