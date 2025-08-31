package security

import (
	"context"
	"testing"
	"time"

	"secure-email-mvp/pkg/models"
	"secure-email-mvp/pkg/securelinks"
)

// MockSystemSecurityRepository implements SystemSecurityRepository for testing
type MockSystemSecurityRepository struct {
	policies []models.SystemSecurityPolicy
}

func (m *MockSystemSecurityRepository) GetAllSystemPolicies() ([]models.SystemSecurityPolicy, error) {
	return m.policies, nil
}

func (m *MockSystemSecurityRepository) GetSystemPolicyByID(policyID string) (*models.SystemSecurityPolicy, error) {
	for _, policy := range m.policies {
		if policy.PolicyID == policyID {
			return &policy, nil
		}
	}
	return nil, nil
}

func (m *MockSystemSecurityRepository) GetSystemPoliciesByType(policyType string) ([]models.SystemSecurityPolicy, error) {
	var filtered []models.SystemSecurityPolicy
	for _, policy := range m.policies {
		if policy.PolicyType == policyType {
			filtered = append(filtered, policy)
		}
	}
	return filtered, nil
}

func (m *MockSystemSecurityRepository) GetSystemPoliciesByCategory(category string) ([]models.SystemSecurityPolicy, error) {
	var filtered []models.SystemSecurityPolicy
	for _, policy := range m.policies {
		if policy.PolicyCategory == category {
			filtered = append(filtered, policy)
		}
	}
	return filtered, nil
}

func (m *MockSystemSecurityRepository) CreateSystemPolicy(policy *models.SystemSecurityPolicy) error {
	m.policies = append(m.policies, *policy)
	return nil
}

func (m *MockSystemSecurityRepository) UpdateSystemPolicy(policy *models.SystemSecurityPolicy) error {
	for i, existingPolicy := range m.policies {
		if existingPolicy.PolicyID == policy.PolicyID {
			m.policies[i] = *policy
			return nil
		}
	}
	return nil
}

func (m *MockSystemSecurityRepository) DeleteSystemPolicy(policyID string) error {
	for i, policy := range m.policies {
		if policy.PolicyID == policyID {
			m.policies = append(m.policies[:i], m.policies[i+1:]...)
			return nil
		}
	}
	return nil
}

// MockDatabase implements Database interface for testing
type MockDatabase struct{}

func (m *MockDatabase) CreateSecurityPolicy(policy *models.SecurityPolicy) error { return nil }
func (m *MockDatabase) GetSecurityPolicy(linkID string) (*models.SecurityPolicy, error) {
	return nil, nil
}
func (m *MockDatabase) UpdateSecurityPolicy(policy *models.SecurityPolicy) error { return nil }
func (m *MockDatabase) GetSecurityPolicyTemplate(templateID string) (*models.SecurityPolicyTemplate, error) {
	return nil, nil
}
func (m *MockDatabase) GetSecurityPolicyTemplates() ([]models.SecurityPolicyTemplate, error) {
	return nil, nil
}
func (m *MockDatabase) CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error { return nil }
func (m *MockDatabase) GetSecureLink(linkID string) (*securelinks.SecureLink, error)    { return nil, nil }
func (m *MockDatabase) UpdateSecureLink(link *securelinks.SecureLink) error             { return nil }

func TestGetSystemSecurityPolicies(t *testing.T) {
	// Create test data
	testPolicies := []models.SystemSecurityPolicy{
		{
			PolicyID:         "sys_pwd_001",
			PolicyName:       "Password Complexity",
			PolicyType:       "password",
			IsActive:         true,
			PolicyValue:      `{"min_length": 12}`,
			PolicyCategory:   "authentication",
			Severity:         "high",
			EnforcementLevel: "required",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			PolicyID:         "sys_mfa_001",
			PolicyName:       "MFA Requirement",
			PolicyType:       "mfa",
			IsActive:         true,
			PolicyValue:      `{"enabled": true}`,
			PolicyCategory:   "authentication",
			Severity:         "critical",
			EnforcementLevel: "mandatory",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	mockRepo := &MockSystemSecurityRepository{policies: testPolicies}
	mockDB := &MockDatabase{}
	service := NewService(mockDB, mockRepo)

	// Test getting all policies
	resp, err := service.GetSystemSecurityPolicies(context.Background())
	if err != nil {
		t.Fatalf("GetSystemSecurityPolicies failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if len(resp.Policies) != 2 {
		t.Errorf("Expected 2 policies, got %d", len(resp.Policies))
	}

	if resp.Message == "" {
		t.Error("Expected message to be set")
	}
}

func TestGetSystemSecurityPoliciesByType(t *testing.T) {
	// Create test data
	testPolicies := []models.SystemSecurityPolicy{
		{
			PolicyID:         "sys_pwd_001",
			PolicyName:       "Password Complexity",
			PolicyType:       "password",
			IsActive:         true,
			PolicyValue:      `{"min_length": 12}`,
			PolicyCategory:   "authentication",
			Severity:         "high",
			EnforcementLevel: "required",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			PolicyID:         "sys_mfa_001",
			PolicyName:       "MFA Requirement",
			PolicyType:       "mfa",
			IsActive:         true,
			PolicyValue:      `{"enabled": true}`,
			PolicyCategory:   "authentication",
			Severity:         "critical",
			EnforcementLevel: "mandatory",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	mockRepo := &MockSystemSecurityRepository{policies: testPolicies}
	mockDB := &MockDatabase{}
	service := NewService(mockDB, mockRepo)

	// Test getting policies by type
	resp, err := service.GetSystemSecurityPoliciesByType(context.Background(), "password")
	if err != nil {
		t.Fatalf("GetSystemSecurityPoliciesByType failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if len(resp.Policies) != 1 {
		t.Errorf("Expected 1 password policy, got %d", len(resp.Policies))
	}

	if resp.Policies[0].PolicyType != "password" {
		t.Errorf("Expected policy type 'password', got %s", resp.Policies[0].PolicyType)
	}
}

func TestUpdateSystemSecurityPolicy(t *testing.T) {
	// Create test data
	testPolicies := []models.SystemSecurityPolicy{
		{
			PolicyID:         "sys_pwd_001",
			PolicyName:       "Password Complexity",
			PolicyType:       "password",
			IsActive:         true,
			PolicyValue:      `{"min_length": 12}`,
			PolicyCategory:   "authentication",
			Severity:         "high",
			EnforcementLevel: "required",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	mockRepo := &MockSystemSecurityRepository{policies: testPolicies}
	mockDB := &MockDatabase{}
	service := NewService(mockDB, mockRepo)

	// Test updating a policy
	req := models.SystemSecurityPolicyUpdateRequest{
		PolicyID:    "sys_pwd_001",
		PolicyValue: `{"min_length": 16}`,
		IsActive:    boolPtr(false),
		UpdatedBy:   stringPtr("test_user"),
	}

	resp, err := service.UpdateSystemSecurityPolicy(context.Background(), req)
	if err != nil {
		t.Fatalf("UpdateSystemSecurityPolicy failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if len(resp.Policies) != 1 {
		t.Errorf("Expected 1 updated policy, got %d", len(resp.Policies))
	}

	updatedPolicy := resp.Policies[0]
	if updatedPolicy.PolicyValue != `{"min_length": 16}` {
		t.Errorf("Expected updated policy value, got %s", updatedPolicy.PolicyValue)
	}

	if updatedPolicy.IsActive {
		t.Error("Expected policy to be inactive")
	}

	if *updatedPolicy.LastModifiedBy != "test_user" {
		t.Errorf("Expected last modified by 'test_user', got %s", *updatedPolicy.LastModifiedBy)
	}
}

func TestUpdateSystemSecurityPolicyNotFound(t *testing.T) {
	mockRepo := &MockSystemSecurityRepository{policies: []models.SystemSecurityPolicy{}}
	mockDB := &MockDatabase{}
	service := NewService(mockDB, mockRepo)

	// Test updating a non-existent policy
	req := models.SystemSecurityPolicyUpdateRequest{
		PolicyID:    "non_existent",
		PolicyValue: `{"min_length": 16}`,
	}

	_, err := service.UpdateSystemSecurityPolicy(context.Background(), req)
	if err == nil {
		t.Error("Expected error for non-existent policy")
	}

	if err.Error() != "policy not found: non_existent" {
		t.Errorf("Expected specific error message, got %v", err)
	}
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}
