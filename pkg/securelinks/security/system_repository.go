package security

import (
	"database/sql"
	"secure-email-mvp/pkg/models"
)

// SystemSecurityRepository interface for system security policy operations
type SystemSecurityRepository interface {
	GetAllSystemPolicies() ([]models.SystemSecurityPolicy, error)
	GetSystemPolicyByID(policyID string) (*models.SystemSecurityPolicy, error)
	GetSystemPoliciesByType(policyType string) ([]models.SystemSecurityPolicy, error)
	GetSystemPoliciesByCategory(category string) ([]models.SystemSecurityPolicy, error)
	CreateSystemPolicy(policy *models.SystemSecurityPolicy) error
	UpdateSystemPolicy(policy *models.SystemSecurityPolicy) error
	DeleteSystemPolicy(policyID string) error
}

// SQLiteSystemSecurityRepository implements SystemSecurityRepository for SQLite
type SQLiteSystemSecurityRepository struct {
	db *sql.DB
}

// NewSQLiteSystemSecurityRepository creates a new SQLite repository for system security policies
func NewSQLiteSystemSecurityRepository(db *sql.DB) *SQLiteSystemSecurityRepository {
	return &SQLiteSystemSecurityRepository{db: db}
}

// GetAllSystemPolicies retrieves all system security policies
func (r *SQLiteSystemSecurityRepository) GetAllSystemPolicies() ([]models.SystemSecurityPolicy, error) {
	query := `
		SELECT policy_id, policy_name, policy_description, policy_type, is_active,
		       policy_value, policy_category, severity, enforcement_level,
		       created_at, updated_at, created_by, last_modified_by
		FROM system_security_policies 
		ORDER BY policy_category, policy_type, policy_name
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.SystemSecurityPolicy
	for rows.Next() {
		var policy models.SystemSecurityPolicy
		err := rows.Scan(
			&policy.PolicyID,
			&policy.PolicyName,
			&policy.PolicyDescription,
			&policy.PolicyType,
			&policy.IsActive,
			&policy.PolicyValue,
			&policy.PolicyCategory,
			&policy.Severity,
			&policy.EnforcementLevel,
			&policy.CreatedAt,
			&policy.UpdatedAt,
			&policy.CreatedBy,
			&policy.LastModifiedBy,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

// GetSystemPolicyByID retrieves a system security policy by ID
func (r *SQLiteSystemSecurityRepository) GetSystemPolicyByID(policyID string) (*models.SystemSecurityPolicy, error) {
	query := `
		SELECT policy_id, policy_name, policy_description, policy_type, is_active,
		       policy_value, policy_category, severity, enforcement_level,
		       created_at, updated_at, created_by, last_modified_by
		FROM system_security_policies 
		WHERE policy_id = ?
	`

	var policy models.SystemSecurityPolicy
	err := r.db.QueryRow(query, policyID).Scan(
		&policy.PolicyID,
		&policy.PolicyName,
		&policy.PolicyDescription,
		&policy.PolicyType,
		&policy.IsActive,
		&policy.PolicyValue,
		&policy.PolicyCategory,
		&policy.Severity,
		&policy.EnforcementLevel,
		&policy.CreatedAt,
		&policy.UpdatedAt,
		&policy.CreatedBy,
		&policy.LastModifiedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &policy, nil
}

// GetSystemPoliciesByType retrieves system security policies by type
func (r *SQLiteSystemSecurityRepository) GetSystemPoliciesByType(policyType string) ([]models.SystemSecurityPolicy, error) {
	query := `
		SELECT policy_id, policy_name, policy_description, policy_type, is_active,
		       policy_value, policy_category, severity, enforcement_level,
		       created_at, updated_at, created_by, last_modified_by
		FROM system_security_policies 
		WHERE policy_type = ?
		ORDER BY policy_name
	`

	rows, err := r.db.Query(query, policyType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.SystemSecurityPolicy
	for rows.Next() {
		var policy models.SystemSecurityPolicy
		err := rows.Scan(
			&policy.PolicyID,
			&policy.PolicyName,
			&policy.PolicyDescription,
			&policy.PolicyType,
			&policy.IsActive,
			&policy.PolicyValue,
			&policy.PolicyCategory,
			&policy.Severity,
			&policy.EnforcementLevel,
			&policy.CreatedAt,
			&policy.UpdatedAt,
			&policy.CreatedBy,
			&policy.LastModifiedBy,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

// GetSystemPoliciesByCategory retrieves system security policies by category
func (r *SQLiteSystemSecurityRepository) GetSystemPoliciesByCategory(category string) ([]models.SystemSecurityPolicy, error) {
	query := `
		SELECT policy_id, policy_name, policy_description, policy_type, is_active,
		       policy_value, policy_category, severity, enforcement_level,
		       created_at, updated_at, created_by, last_modified_by
		FROM system_security_policies 
		WHERE policy_category = ?
		ORDER BY policy_type, policy_name
	`

	rows, err := r.db.Query(query, category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var policies []models.SystemSecurityPolicy
	for rows.Next() {
		var policy models.SystemSecurityPolicy
		err := rows.Scan(
			&policy.PolicyID,
			&policy.PolicyName,
			&policy.PolicyDescription,
			&policy.PolicyType,
			&policy.IsActive,
			&policy.PolicyValue,
			&policy.PolicyCategory,
			&policy.Severity,
			&policy.EnforcementLevel,
			&policy.CreatedAt,
			&policy.UpdatedAt,
			&policy.CreatedBy,
			&policy.LastModifiedBy,
		)
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return policies, nil
}

// CreateSystemPolicy creates a new system security policy
func (r *SQLiteSystemSecurityRepository) CreateSystemPolicy(policy *models.SystemSecurityPolicy) error {
	query := `
		INSERT INTO system_security_policies (
			policy_id, policy_name, policy_description, policy_type, is_active,
			policy_value, policy_category, severity, enforcement_level,
			created_at, updated_at, created_by, last_modified_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		policy.PolicyID,
		policy.PolicyName,
		policy.PolicyDescription,
		policy.PolicyType,
		policy.IsActive,
		policy.PolicyValue,
		policy.PolicyCategory,
		policy.Severity,
		policy.EnforcementLevel,
		policy.CreatedAt,
		policy.UpdatedAt,
		policy.CreatedBy,
		policy.LastModifiedBy,
	)

	return err
}

// UpdateSystemPolicy updates an existing system security policy
func (r *SQLiteSystemSecurityRepository) UpdateSystemPolicy(policy *models.SystemSecurityPolicy) error {
	query := `
		UPDATE system_security_policies SET
			policy_name = ?, policy_description = ?, policy_type = ?, is_active = ?,
			policy_value = ?, policy_category = ?, severity = ?, enforcement_level = ?,
			updated_at = ?, last_modified_by = ?
		WHERE policy_id = ?
	`

	_, err := r.db.Exec(query,
		policy.PolicyName,
		policy.PolicyDescription,
		policy.PolicyType,
		policy.IsActive,
		policy.PolicyValue,
		policy.PolicyCategory,
		policy.Severity,
		policy.EnforcementLevel,
		policy.UpdatedAt,
		policy.LastModifiedBy,
		policy.PolicyID,
	)

	return err
}

// DeleteSystemPolicy deletes a system security policy
func (r *SQLiteSystemSecurityRepository) DeleteSystemPolicy(policyID string) error {
	query := `DELETE FROM system_security_policies WHERE policy_id = ?`
	_, err := r.db.Exec(query, policyID)
	return err
}







