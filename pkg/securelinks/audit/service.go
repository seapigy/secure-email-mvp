package audit

import (
	"time"
	"secure-email-mvp/pkg/models"
)

// Service provides audit logging functionality
type Service struct {
	repository Repository
}

// NewService creates a new audit service
func NewService(repo Repository) *Service {
	return &Service{
		repository: repo,
	}
}

// LogAction creates an audit log entry for a user action
func (s *Service) LogAction(userID, action, entity, details, severity string) error {
	log := &models.AuditLog{
		Timestamp: time.Now(),
		UserID:    userID,
		Action:    action,
		Entity:    entity,
		Details:   details,
		Severity:  severity,
	}
	
	return s.repository.CreateAuditLog(log)
}

// GetAuditLogs retrieves audit logs with optional filtering and pagination
func (s *Service) GetAuditLogs(filters models.AuditLogFilters) ([]models.AuditLog, int, error) {
	// Set default values if not provided
	if filters.Limit <= 0 {
		filters.Limit = 50 // Default page size
	}
	if filters.Offset < 0 {
		filters.Offset = 0
	}
	
	return s.repository.GetAuditLogs(filters)
}

// LogAdminLogin logs an admin login event
func (s *Service) LogAdminLogin(adminID string) error {
	return s.LogAction(adminID, "login", "admin", "Admin login successful", "low")
}

// LogPolicyUpdate logs a security policy update
func (s *Service) LogPolicyUpdate(adminID, policyName, details string) error {
	return s.LogAction(adminID, "update_policy", "system_security_policy", details, "high")
}

// LogDLPScan logs a DLP scan event
func (s *Service) LogDLPScan(userID, details string, severity string) error {
	return s.LogAction(userID, "dlp_scan", "secure_link", details, severity)
}

// LogSecureLinkCreation logs a secure link creation
func (s *Service) LogSecureLinkCreation(userID, details string) error {
	return s.LogAction(userID, "create_secure_link", "secure_link", details, "medium")
}

// LogAuditLogView logs when audit logs are viewed
func (s *Service) LogAuditLogView(adminID string) error {
	return s.LogAction(adminID, "view_audit_logs", "audit_logs", "Viewed audit log entries", "low")
}






