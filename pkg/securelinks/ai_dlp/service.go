package ai_dlp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"secure-email-mvp/pkg/models"
)

// Service handles AI-powered data loss prevention scanning
type Service struct {
	db         Database
	classifier Classifier
	config     *Config
}

// Database interface for AI DLP operations
type Database interface {
	GetAIDLPPolicy(policyID string) (*models.AIDLPPolicy, error)
	GetDefaultAIDLPPolicy() (*models.AIDLPPolicy, error)
	GetAIDLPScanResult(scanID string) (*models.AIDLPScanResult, error)
	CreateAIDLPScanResult(result *models.AIDLPScanResult) error
	CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error
	UpdateAIDLPScanResult(result *models.AIDLPScanResult) error
}

// Classifier interface for AI content classification
type Classifier interface {
	ClassifyContent(content string) (*models.AIContentClassification, error)
	ExtractEntities(content string) ([]models.Entity, error)
	CalculateRiskScore(classification *models.AIContentClassification) float64
}

// Config holds AI DLP service configuration
type Config struct {
	ModelVersion               string
	DefaultConfidenceThreshold float64
	DefaultRiskThreshold       float64
	ProcessingTimeout          time.Duration
	EnableEntityExtraction     bool
	EnableContextAnalysis      bool
}

// NewService creates a new AI DLP service
func NewService(db Database, classifier Classifier, config *Config) *Service {
	return &Service{
		db:         db,
		classifier: classifier,
		config:     config,
	}
}

// ScanContent performs AI-powered DLP scanning
func (s *Service) ScanContent(ctx context.Context, req models.AIDLPScanRequest) (*models.AIDLPScanResponse, error) {
	startTime := time.Now()

	// Get policy (specific or default)
	var policy *models.AIDLPPolicy
	var err error
	if req.PolicyID != nil {
		policy, err = s.db.GetAIDLPPolicy(*req.PolicyID)
	} else {
		policy, err = s.db.GetDefaultAIDLPPolicy()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get AI DLP policy: %w", err)
	}

	// Generate content hash
	contentHash := s.generateContentHash(req.Content)

	// Perform AI classification
	classification, err := s.classifier.ClassifyContent(req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to classify content: %w", err)
	}

	// Extract entities if enabled
	if s.config.EnableEntityExtraction {
		entities, err := s.classifier.ExtractEntities(req.Content)
		if err == nil {
			classification.Entities = entities
		}
	}

	// Calculate risk score
	riskScore := s.classifier.CalculateRiskScore(classification)

	// Determine severity level
	severityLevel := models.GetSeverityFromScore(riskScore)

	// Get recommended action based on policy
	actionRecommended := s.getRecommendedAction(severityLevel, policy)

	// Determine final action
	actionTaken := s.determineFinalAction(&actionRecommended, req.UserRole, policy)

	// Calculate processing time
	processingTime := float64(time.Since(startTime).Milliseconds())

	// Create scan result
	scanResult := models.AIDLPScanResult{
		ScanID:            s.generateScanID(),
		LinkID:            &req.LinkID,
		ReplyID:           req.ReplyID,
		AttachmentID:      req.AttachmentID,
		ContentType:       req.ContentType,
		ContentHash:       contentHash,
		Classification:    classification,
		SeverityScore:     riskScore,
		RiskLevel:         severityLevel,
		ActionRecommended: actionRecommended,
		ActionTaken:       actionTaken,
		ModelVersion:      s.config.ModelVersion,
		ProcessingTime:    processingTime,
		ScanTimestamp:     time.Now(),
		CreatedBy:         req.UserID,
	}

	// Store scan result
	if err := s.db.CreateAIDLPScanResult(&scanResult); err != nil {
		return nil, fmt.Errorf("failed to store AI DLP scan result: %w", err)
	}

	// Log compliance audit event
	auditDetails := map[string]interface{}{
		"content_type":       req.ContentType,
		"classification":     classification.Category,
		"severity_score":     riskScore,
		"risk_level":         severityLevel,
		"action_recommended": actionRecommended,
		"action_taken":       actionTaken,
		"processing_time":    processingTime,
		"model_version":      s.config.ModelVersion,
		"policy_id":          policy.PolicyID,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:            s.generateAuditID(),
		EventType:          "ai_dlp_scan",
		LinkID:             &req.LinkID,
		ReplyID:            req.ReplyID,
		AttachmentID:       req.AttachmentID,
		IPAddress:          nil, // Will be set by caller
		UserAgent:          nil, // Will be set by caller
		Severity:           s.getSeverityFromRiskLevel(severityLevel),
		ComplianceCategory: stringPtr("ai_dlp"),
		RetentionRequired:  true,
		CreatedAt:          time.Now(),
	}

	if err := auditLog.SetEventDetails(auditDetails); err != nil {
		return nil, fmt.Errorf("failed to set audit details: %w", err)
	}

	if err := s.db.CreateComplianceAuditLog(&auditLog); err != nil {
		return nil, fmt.Errorf("failed to create compliance audit log: %w", err)
	}

	// Determine if override is allowed
	canOverride := s.canOverride(actionTaken, req.UserRole, policy)

	return &models.AIDLPScanResponse{
		Success:           true,
		ScanID:            scanResult.ScanID,
		Classification:    classification,
		SeverityScore:     riskScore,
		RiskLevel:         severityLevel,
		ActionRecommended: actionRecommended,
		ActionTaken:       actionTaken,
		CanOverride:       canOverride,
		ProcessingTime:    processingTime,
		ModelVersion:      s.config.ModelVersion,
		Message:           s.generateMessage(severityLevel, actionTaken),
	}, nil
}

// OverrideDecision allows authorized users to override AI DLP decisions
func (s *Service) OverrideDecision(ctx context.Context, req models.AIDLPOverrideRequest) (*models.AIDLPOverrideResponse, error) {
	// Get the original scan result
	scanResult, err := s.db.GetAIDLPScanResult(req.ScanID)
	if err != nil {
		return nil, fmt.Errorf("failed to get scan result: %w", err)
	}

	// Check if override is allowed
	if !s.canOverride(scanResult.ActionTaken, &req.UserRole, nil) {
		return &models.AIDLPOverrideResponse{
			Success:   false,
			Error:     "Override not allowed for this action or user role",
			ErrorCode: "OVERRIDE_NOT_ALLOWED",
		}, nil
	}

	// Update scan result with override information
	scanResult.ActionTaken = "overridden"
	scanResult.OverrideReason = &req.OverrideReason
	scanResult.OverrideBy = &req.UserID

	if err := s.db.UpdateAIDLPScanResult(scanResult); err != nil {
		return nil, fmt.Errorf("failed to update scan result: %w", err)
	}

	// Log override event
	auditDetails := map[string]interface{}{
		"original_action": scanResult.ActionRecommended,
		"override_reason": req.OverrideReason,
		"justification":   req.Justification,
		"user_role":       req.UserRole,
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:            s.generateAuditID(),
		EventType:          "ai_dlp_override",
		LinkID:             scanResult.LinkID,
		ReplyID:            scanResult.ReplyID,
		AttachmentID:       scanResult.AttachmentID,
		UserID:             &req.UserID,
		IPAddress:          nil, // Will be set by caller
		UserAgent:          nil, // Will be set by caller
		Severity:           "warning",
		ComplianceCategory: stringPtr("ai_dlp"),
		RetentionRequired:  true,
		CreatedAt:          time.Now(),
	}

	if err := auditLog.SetEventDetails(auditDetails); err != nil {
		return nil, fmt.Errorf("failed to set audit details: %w", err)
	}

	if err := s.db.CreateComplianceAuditLog(&auditLog); err != nil {
		return nil, fmt.Errorf("failed to create compliance audit log: %w", err)
	}

	return &models.AIDLPOverrideResponse{
		Success:     true,
		OverrideID:  s.generateOverrideID(),
		ActionTaken: "overridden",
		Message:     "AI DLP decision overridden successfully",
	}, nil
}

// getRecommendedAction determines the recommended action based on severity and policy
func (s *Service) getRecommendedAction(severityLevel string, policy *models.AIDLPPolicy) string {
	// Check if policy has specific action mapping
	if policy.Actions != nil {
		if action, exists := policy.Actions[severityLevel]; exists {
			return action
		}
	}

	// Use default action mapping
	return models.GetActionForSeverity(severityLevel)
}

// determineFinalAction determines the final action based on user role and policy
func (s *Service) determineFinalAction(recommendedAction, userRole *string, policy *models.AIDLPPolicy) string {
	if recommendedAction != nil && *recommendedAction == "block" && userRole != nil {
		// Check if user role can override blocks
		if policy.AllowOverride && policy.OverrideRoles != nil {
			for _, role := range policy.OverrideRoles {
				if role == *userRole {
					return "warn" // Allow override for authorized roles
				}
			}
		}
	}
	if recommendedAction != nil {
		return *recommendedAction
	}
	return "allow"
}

// canOverride checks if the current action can be overridden by the user role
func (s *Service) canOverride(actionTaken string, userRole *string, policy *models.AIDLPPolicy) bool {
	if actionTaken != "blocked" {
		return false
	}

	if userRole == nil || policy == nil {
		return false
	}

	if !policy.AllowOverride {
		return false
	}

	for _, role := range policy.OverrideRoles {
		if role == *userRole {
			return true
		}
	}

	return false
}

// getSeverityFromRiskLevel converts risk level to audit severity
func (s *Service) getSeverityFromRiskLevel(riskLevel string) string {
	switch riskLevel {
	case "critical":
		return "critical"
	case "high":
		return "error"
	case "medium":
		return "warning"
	case "low":
		return "info"
	default:
		return "info"
	}
}

// generateMessage creates a user-friendly message based on severity and action
func (s *Service) generateMessage(severityLevel, actionTaken string) string {
	switch actionTaken {
	case "blocked":
		return fmt.Sprintf("Content blocked due to %s risk level. Please review and remove sensitive information.", severityLevel)
	case "warned":
		return fmt.Sprintf("Content flagged with %s risk level. Please review before sending.", severityLevel)
	case "allowed":
		return "Content allowed with monitoring."
	case "overridden":
		return "Content allowed after override by authorized user."
	default:
		return "Content processed successfully."
	}
}

// generateContentHash creates a SHA256 hash of the content
func (s *Service) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// generateScanID generates a unique scan ID
func (s *Service) generateScanID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "ai_dlp_scan_" + hex.EncodeToString(bytes)
}

// generateOverrideID generates a unique override ID
func (s *Service) generateOverrideID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "override_" + hex.EncodeToString(bytes)
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
