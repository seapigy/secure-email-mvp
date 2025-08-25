package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"secure-email-mvp/pkg/models"
	"secure-email-mvp/pkg/securelinks"
)

// Service handles security policies and controls
type Service struct {
	db Database
}

// Database interface for security operations
type Database interface {
	CreateSecurityPolicy(policy *models.SecurityPolicy) error
	GetSecurityPolicy(linkID string) (*models.SecurityPolicy, error)
	UpdateSecurityPolicy(policy *models.SecurityPolicy) error
	GetSecurityPolicyTemplate(templateID string) (*models.SecurityPolicyTemplate, error)
	GetSecurityPolicyTemplates() ([]models.SecurityPolicyTemplate, error)
	CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error
	GetSecureLink(linkID string) (*securelinks.SecureLink, error)
	UpdateSecureLink(link *securelinks.SecureLink) error
}

// NewService creates a new security service
func NewService(db Database) *Service {
	return &Service{
		db: db,
	}
}

// CreateSecurityPolicy creates a new security policy
func (s *Service) CreateSecurityPolicy(ctx context.Context, req models.CreateSecurityPolicyRequest) (*models.SecurityPolicyResponse, error) {
	// Generate policy ID
	policyID := s.generatePolicyID()

	// Create security policy
	policy := models.SecurityPolicy{
		PolicyID:             policyID,
		LinkID:               req.LinkID,
		ReplyID:              req.ReplyID,
		EmailID:              req.EmailID,
		DLPEnabled:           req.DLPEnabled,
		WatermarkEnabled:     req.WatermarkEnabled,
		DownloadDisabled:     req.DownloadDisabled,
		ForwardingDisabled:   req.ForwardingDisabled,
		AutoRevokeAfterReply: req.AutoRevokeAfterReply,
		MaxViews:             req.MaxViews,
		ExpiresAt:            req.ExpiresAt,
		ExpiresAfterViews:    req.ExpiresAfterViews,
		NotifyOnExpiry:       req.NotifyOnExpiry,
		NotifyOnRevoke:       req.NotifyOnRevoke,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// If template ID is provided, apply template settings
	if req.TemplateID != nil {
		template, err := s.db.GetSecurityPolicyTemplate(*req.TemplateID)
		if err != nil {
			return nil, fmt.Errorf("failed to get security policy template: %w", err)
		}

		// Apply template settings (only if not explicitly set in request)
		if !req.DLPEnabled {
			policy.DLPEnabled = template.DLPEnabled
		}
		if !req.WatermarkEnabled {
			policy.WatermarkEnabled = template.WatermarkEnabled
		}
		if !req.DownloadDisabled {
			policy.DownloadDisabled = template.DownloadDisabled
		}
		if !req.ForwardingDisabled {
			policy.ForwardingDisabled = template.ForwardingDisabled
		}
		if !req.AutoRevokeAfterReply {
			policy.AutoRevokeAfterReply = template.AutoRevokeAfterReply
		}
		if req.MaxViews == nil {
			policy.MaxViews = template.MaxViews
		}
		if req.ExpiresAt == nil && template.DefaultExpiryHours != nil {
			expiryTime := time.Now().Add(time.Duration(*template.DefaultExpiryHours) * time.Hour)
			policy.ExpiresAt = &expiryTime
		}
		if !req.NotifyOnExpiry {
			policy.NotifyOnExpiry = template.NotifyOnExpiry
		}
		if !req.NotifyOnRevoke {
			policy.NotifyOnRevoke = template.NotifyOnRevoke
		}
	}

	// Store security policy
	if err := s.db.CreateSecurityPolicy(&policy); err != nil {
		return nil, fmt.Errorf("failed to create security policy: %w", err)
	}

	// Log compliance audit event
	auditDetails := map[string]interface{}{
		"policy_id": policyID,
		"link_id": req.LinkID,
		"dlp_enabled": policy.DLPEnabled,
		"watermark_enabled": policy.WatermarkEnabled,
		"download_disabled": policy.DownloadDisabled,
		"forwarding_disabled": policy.ForwardingDisabled,
		"auto_revoke_after_reply": policy.AutoRevokeAfterReply,
		"max_views": policy.MaxViews,
		"expires_at": policy.ExpiresAt,
		"expires_after_views": policy.ExpiresAfterViews,
		"template_id": req.TemplateID,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:           s.generateAuditID(),
		EventType:         "policy_created",
		LinkID:            &req.LinkID,
		ReplyID:           req.ReplyID,
		PolicyID:          &policyID,
		IPAddress:         nil, // Will be set by caller
		UserAgent:         nil, // Will be set by caller
		Severity:          "info",
		ComplianceCategory: stringPtr("access_control"),
		RetentionRequired: true,
		CreatedAt:         time.Now(),
	}

	if err := auditLog.SetEventDetails(auditDetails); err != nil {
		return nil, fmt.Errorf("failed to set audit details: %w", err)
	}

	if err := s.db.CreateComplianceAuditLog(&auditLog); err != nil {
		return nil, fmt.Errorf("failed to create compliance audit log: %w", err)
	}

	return &models.SecurityPolicyResponse{
		Success:  true,
		PolicyID: policyID,
		Policy:   &policy,
		Message:  "Security policy created successfully",
	}, nil
}

// GetSecurityPolicy retrieves a security policy
func (s *Service) GetSecurityPolicy(ctx context.Context, linkID string) (*models.SecurityPolicyResponse, error) {
	policy, err := s.db.GetSecurityPolicy(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get security policy: %w", err)
	}

	return &models.SecurityPolicyResponse{
		Success: true,
		Policy:  policy,
	}, nil
}

// UpdateSecurityPolicy updates an existing security policy
func (s *Service) UpdateSecurityPolicy(ctx context.Context, policy *models.SecurityPolicy) (*models.SecurityPolicyResponse, error) {
	policy.UpdatedAt = time.Now()

	if err := s.db.UpdateSecurityPolicy(policy); err != nil {
		return nil, fmt.Errorf("failed to update security policy: %w", err)
	}

	// Log compliance audit event
	auditDetails := map[string]interface{}{
		"policy_id": policy.PolicyID,
		"link_id": policy.LinkID,
		"updated_at": policy.UpdatedAt,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:           s.generateAuditID(),
		EventType:         "policy_updated",
		LinkID:            &policy.LinkID,
		ReplyID:           policy.ReplyID,
		PolicyID:          &policy.PolicyID,
		IPAddress:         nil, // Will be set by caller
		UserAgent:         nil, // Will be set by caller
		Severity:          "info",
		ComplianceCategory: stringPtr("access_control"),
		RetentionRequired: true,
		CreatedAt:         time.Now(),
	}

	if err := auditLog.SetEventDetails(auditDetails); err != nil {
		return nil, fmt.Errorf("failed to set audit details: %w", err)
	}

	if err := s.db.CreateComplianceAuditLog(&auditLog); err != nil {
		return nil, fmt.Errorf("failed to create compliance audit log: %w", err)
	}

	return &models.SecurityPolicyResponse{
		Success: true,
		Policy:  policy,
		Message: "Security policy updated successfully",
	}, nil
}

// GetSecurityPolicyTemplates retrieves all security policy templates
func (s *Service) GetSecurityPolicyTemplates(ctx context.Context) (*models.SecurityPolicyTemplateResponse, error) {
	templates, err := s.db.GetSecurityPolicyTemplates()
	if err != nil {
		return nil, fmt.Errorf("failed to get security policy templates: %w", err)
	}

	return &models.SecurityPolicyTemplateResponse{
		Success:   true,
		Templates: templates,
	}, nil
}

// CheckAccessControl checks if access is allowed based on security policy
func (s *Service) CheckAccessControl(ctx context.Context, linkID string, userID string, ipAddress string, userAgent string) (*models.SecurityPolicyResponse, error) {
	// Get security policy
	policy, err := s.db.GetSecurityPolicy(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get security policy: %w", err)
	}

	// Get secure link to check current state
	secureLink, err := s.db.GetSecureLink(linkID)
	if err != nil {
		return nil, fmt.Errorf("failed to get secure link: %w", err)
	}

	// Check if link is expired
	if policy.IsExpired() {
		// Log expiration event
		s.logExpirationEvent(linkID, userID, ipAddress, userAgent, policy)
		return &models.SecurityPolicyResponse{
			Success: false,
			Error:   "Link has expired",
			ErrorCode: "LINK_EXPIRED",
		}, nil
	}

	// Check if link should expire after views
	if policy.ShouldExpireAfterViews(secureLink.AccessCount) {
		// Log expiration event
		s.logExpirationEvent(linkID, userID, ipAddress, userAgent, policy)
		return &models.SecurityPolicyResponse{
			Success: false,
			Error:   "Link has expired after maximum views",
			ErrorCode: "LINK_EXPIRED_VIEWS",
		}, nil
	}

	// Check if link is revoked
	if secureLink.Status == "revoked" {
		// Log revocation event
		s.logRevocationEvent(linkID, userID, ipAddress, userAgent, policy)
		return &models.SecurityPolicyResponse{
			Success: false,
			Error:   "Link has been revoked",
			ErrorCode: "LINK_REVOKED",
		}, nil
	}

	// Increment access count
	secureLink.AccessCount++
	now := time.Now()
	secureLink.LastAccessed = &now
	if err := s.db.UpdateSecureLink(secureLink); err != nil {
		return nil, fmt.Errorf("failed to update secure link: %w", err)
	}

	// Log access event
	s.logAccessEvent(linkID, userID, ipAddress, userAgent, policy)

	return &models.SecurityPolicyResponse{
		Success: true,
		Policy:  policy,
		Message: "Access granted",
	}, nil
}

// logExpirationEvent logs when a link expires
func (s *Service) logExpirationEvent(linkID, userID, ipAddress, userAgent string, policy *models.SecurityPolicy) {
	auditDetails := map[string]interface{}{
		"link_id": linkID,
		"user_id": userID,
		"expiration_reason": "time_expired",
		"policy_id": policy.PolicyID,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:           s.generateAuditID(),
		EventType:         "expiration_triggered",
		LinkID:            &linkID,
		PolicyID:          &policy.PolicyID,
		UserID:            &userID,
		IPAddress:         &ipAddress,
		UserAgent:         &userAgent,
		Severity:          "warning",
		ComplianceCategory: stringPtr("expiration"),
		RetentionRequired: true,
		CreatedAt:         time.Now(),
	}

	auditLog.SetEventDetails(auditDetails)
	s.db.CreateComplianceAuditLog(&auditLog)
}

// logRevocationEvent logs when a link is revoked
func (s *Service) logRevocationEvent(linkID, userID, ipAddress, userAgent string, policy *models.SecurityPolicy) {
	auditDetails := map[string]interface{}{
		"link_id": linkID,
		"user_id": userID,
		"revocation_reason": "manual_revoke",
		"policy_id": policy.PolicyID,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:           s.generateAuditID(),
		EventType:         "revocation_triggered",
		LinkID:            &linkID,
		PolicyID:          &policy.PolicyID,
		UserID:            &userID,
		IPAddress:         &ipAddress,
		UserAgent:         &userAgent,
		Severity:          "warning",
		ComplianceCategory: stringPtr("revocation"),
		RetentionRequired: true,
		CreatedAt:         time.Now(),
	}

	auditLog.SetEventDetails(auditDetails)
	s.db.CreateComplianceAuditLog(&auditLog)
}

// logAccessEvent logs when access is granted
func (s *Service) logAccessEvent(linkID, userID, ipAddress, userAgent string, policy *models.SecurityPolicy) {
	auditDetails := map[string]interface{}{
		"link_id": linkID,
		"user_id": userID,
		"access_granted": true,
		"policy_id": policy.PolicyID,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:           s.generateAuditID(),
		EventType:         "policy_enforced",
		LinkID:            &linkID,
		PolicyID:          &policy.PolicyID,
		UserID:            &userID,
		IPAddress:         &ipAddress,
		UserAgent:         &userAgent,
		Severity:          "info",
		ComplianceCategory: stringPtr("access_control"),
		RetentionRequired: true,
		CreatedAt:         time.Now(),
	}

	auditLog.SetEventDetails(auditDetails)
	s.db.CreateComplianceAuditLog(&auditLog)
}

// generatePolicyID generates a unique policy ID
func (s *Service) generatePolicyID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "policy_" + hex.EncodeToString(bytes)
}

// generateAuditID generates a unique audit ID
func (s *Service) generateAuditID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "audit_" + hex.EncodeToString(bytes)
}

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
