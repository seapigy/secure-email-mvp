package dlp

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"secure-email-mvp/pkg/models"
	monitoring "secure-email-mvp/pkg/securelinks/monitoring"
)

// Service handles data loss prevention scanning
type Service struct {
	db Database
	monitoringService *monitoring.Service
}

// Database interface for DLP operations
type Database interface {
	GetActiveDLPRules() ([]models.DLPRule, error)
	CreateDLPScanResult(result *models.DLPScanResult) error
	CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error
}

// NewService creates a new DLP service
func NewService(db Database, monitoringService *monitoring.Service) *Service {
	return &Service{
		db: db,
		monitoringService: monitoringService,
	}
}

// ScanContent scans content for DLP violations
func (s *Service) ScanContent(ctx context.Context, req models.DLPScanRequest) (*models.DLPScanResponse, error) {
	var violations []models.DLPScanResult
	highestAction := "allowed"

	// First, scan with built-in patterns for common sensitive data
	builtInResults := s.scanBuiltInPatterns(req.Content)
	for _, result := range builtInResults {
		// Set request-specific fields
		result.LinkID = &req.LinkID
		result.ReplyID = req.ReplyID
		result.AttachmentID = req.AttachmentID
		result.ContentType = req.ContentType
		
		// Store scan result
		if err := s.db.CreateDLPScanResult(&result); err != nil {
			return nil, fmt.Errorf("failed to store DLP scan result: %w", err)
		}

		violations = append(violations, result)

		// Update highest action based on severity
		if s.getActionPriority(result.ActionTaken) > s.getActionPriority(highestAction) {
			highestAction = result.ActionTaken
		}
	}

	// Get active DLP rules from database
	rules, err := s.db.GetActiveDLPRules()
	if err != nil {
		// Log error but continue with built-in patterns
		fmt.Printf("Warning: failed to get DLP rules from database: %v\n", err)
	} else {
		// Scan content against each rule
		for _, rule := range rules {
			if !rule.IsActive {
				continue
			}

			matches := s.scanRule(req.Content, rule)
			if len(matches) > 0 {
				// Create scan result for each match
				for _, match := range matches {
					scanResult := models.DLPScanResult{
						ScanID:         s.generateScanID(),
						LinkID:         &req.LinkID,
						ReplyID:        req.ReplyID,
						AttachmentID:   req.AttachmentID,
						RuleID:         rule.RuleID,
						ContentType:    req.ContentType,
						MatchedContent: &match,
						ConfidenceScore: s.calculateConfidence(match, rule),
						ActionTaken:    rule.Action,
						ScanTimestamp:  time.Now(),
					}

					// Store scan result
					if err := s.db.CreateDLPScanResult(&scanResult); err != nil {
						return nil, fmt.Errorf("failed to store DLP scan result: %w", err)
					}

					violations = append(violations, scanResult)

					// Update highest action based on severity
					if s.getActionPriority(rule.Action) > s.getActionPriority(highestAction) {
						highestAction = rule.Action
					}
				}
			}
		}
	}

	// Log compliance audit event
	auditDetails := map[string]interface{}{
		"content_type": req.ContentType,
		"violations_count": len(violations),
		"action_taken": highestAction,
		"rules_checked": len(rules),
	}

	auditLog := models.ComplianceAuditLog{
		AuditID:           s.generateAuditID(),
		EventType:         "dlp_scan",
		LinkID:            &req.LinkID,
		ReplyID:           req.ReplyID,
		AttachmentID:      req.AttachmentID,
		IPAddress:         nil, // Will be set by caller
		UserAgent:         nil, // Will be set by caller
		Severity:          s.getSeverityFromAction(highestAction),
		ComplianceCategory: stringPtr("dlp"),
		RetentionRequired: true,
		CreatedAt:         time.Now(),
	}

	if err := auditLog.SetEventDetails(auditDetails); err != nil {
		return nil, fmt.Errorf("failed to set audit details: %w", err)
	}

	if err := s.db.CreateComplianceAuditLog(&auditLog); err != nil {
		return nil, fmt.Errorf("failed to create compliance audit log: %w", err)
	}

	// Log monitoring event
	if s.monitoringService != nil {
		scanResult := "clean"
		if len(violations) > 0 {
			scanResult = fmt.Sprintf("%d_violations_%s", len(violations), highestAction)
		}
		
		event := models.CreateDLPScanEvent(
			req.ContentType,
			scanResult,
			0.0, // TODO: Add actual processing time measurement
		)
		if err := s.monitoringService.LogEvent(event); err != nil {
			// Log error but don't fail the scan
			fmt.Printf("Failed to log DLP monitoring event: %v\n", err)
		}
	}

	// Convert violations to findings format for enhanced response
	var findings []map[string]interface{}
	for _, violation := range violations {
		finding := map[string]interface{}{
			"type":    violation.RuleID,
			"match":   violation.MatchedContent,
			"action":  violation.ActionTaken,
			"confidence": violation.ConfidenceScore,
		}
		findings = append(findings, finding)
	}

	return &models.DLPScanResponse{
		Success:     true,
		Violations:  violations,
		ActionTaken: highestAction,
		Message:     fmt.Sprintf("DLP scan completed. Found %d violations.", len(violations)),
	}, nil
}

// scanRule scans content against a specific DLP rule
func (s *Service) scanRule(content string, rule models.DLPRule) []string {
	var matches []string

	switch rule.RuleType {
	case "regex":
		matches = s.scanRegex(content, rule.Pattern)
	case "keyword":
		matches = s.scanKeywords(content, rule.Pattern)
	case "ai_pattern":
		// Placeholder for future AI-based pattern matching
		matches = s.scanAIPattern(content, rule.Pattern)
	default:
		// Unknown rule type, skip
		return matches
	}

	return matches
}

// scanRegex scans content using regex pattern
func (s *Service) scanRegex(content, pattern string) []string {
	var matches []string

	// Compile regex pattern
	re, err := regexp.Compile(pattern)
	if err != nil {
		// Invalid regex, return empty matches
		return matches
	}

	// Find all matches
	foundMatches := re.FindAllString(content, -1)
	for _, match := range foundMatches {
		// Add unique matches only
		if !s.containsString(matches, match) {
			matches = append(matches, match)
		}
	}

	return matches
}

// scanKeywords scans content for keyword matches
func (s *Service) scanKeywords(content, keywords string) []string {
	var matches []string

	// Split keywords by pipe (|) for multiple keywords
	keywordList := strings.Split(keywords, "|")
	contentLower := strings.ToLower(content)

	for _, keyword := range keywordList {
		keyword = strings.TrimSpace(keyword)
		if keyword == "" {
			continue
		}

		// Check if keyword exists in content (case-insensitive)
		if strings.Contains(contentLower, strings.ToLower(keyword)) {
			// Find the actual occurrence in original content
			keywordRegex := regexp.QuoteMeta(keyword)
			re, err := regexp.Compile("(?i)" + keywordRegex)
			if err == nil {
				foundMatches := re.FindAllString(content, -1)
				for _, match := range foundMatches {
					if !s.containsString(matches, match) {
						matches = append(matches, match)
					}
				}
			}
		}
	}

	return matches
}

// scanAIPattern is a placeholder for future AI-based pattern matching
func (s *Service) scanAIPattern(content, pattern string) []string {
	// TODO: Implement AI-based pattern matching
	// For now, return empty matches
	return []string{}
}

// scanBuiltInPatterns scans content for built-in sensitive data patterns
func (s *Service) scanBuiltInPatterns(content string) []models.DLPScanResult {
	var results []models.DLPScanResult
	
	// Email addresses pattern
	emailPattern := `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`
	emailMatches := s.scanRegex(content, emailPattern)
	for _, match := range emailMatches {
		results = append(results, models.DLPScanResult{
			ScanID:         s.generateScanID(),
			RuleID:         "builtin_email",
			ContentType:    "email_body",
			MatchedContent: &match,
			ConfidenceScore: 0.9,
			ActionTaken:    "allow",
			ScanTimestamp:  time.Now(),
		})
	}
	
	// Phone numbers pattern (US format)
	phonePattern := `\b\d{3}[-\s]?\d{3}[-\s]?\d{4}\b`
	phoneMatches := s.scanRegex(content, phonePattern)
	for _, match := range phoneMatches {
		results = append(results, models.DLPScanResult{
			ScanID:         s.generateScanID(),
			RuleID:         "builtin_phone",
			ContentType:    "email_body",
			MatchedContent: &match,
			ConfidenceScore: 0.8,
			ActionTaken:    "allow",
			ScanTimestamp:  time.Now(),
		})
	}
	
	// Credit card numbers pattern
	ccPattern := `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`
	ccMatches := s.scanRegex(content, ccPattern)
	for _, match := range ccMatches {
		results = append(results, models.DLPScanResult{
			ScanID:         s.generateScanID(),
			RuleID:         "builtin_credit_card",
			ContentType:    "email_body",
			MatchedContent: &match,
			ConfidenceScore: 0.95,
			ActionTaken:    "warn",
			ScanTimestamp:  time.Now(),
		})
	}
	
	// Social Security numbers pattern (US format)
	ssnPattern := `\b\d{3}[-\s]?\d{2}[-\s]?\d{4}\b`
	ssnMatches := s.scanRegex(content, ssnPattern)
	for _, match := range ssnMatches {
		results = append(results, models.DLPScanResult{
			ScanID:         s.generateScanID(),
			RuleID:         "builtin_ssn",
			ContentType:    "email_body",
			MatchedContent: &match,
			ConfidenceScore: 0.95,
			ActionTaken:    "warn",
			ScanTimestamp:  time.Now(),
		})
	}
	
	return results
}

// calculateConfidence calculates confidence score for a match
func (s *Service) calculateConfidence(match string, rule models.DLPRule) float64 {
	baseConfidence := 0.5

	// Adjust confidence based on rule type
	switch rule.RuleType {
	case "regex":
		// Regex patterns are more precise
		baseConfidence = 0.8
	case "keyword":
		// Keyword matches are less precise
		baseConfidence = 0.6
	case "ai_pattern":
		// AI patterns would have variable confidence
		baseConfidence = 0.7
	}

	// Adjust based on match length (longer matches might be more reliable)
	if len(match) > 10 {
		baseConfidence += 0.1
	}

	// Adjust based on rule severity
	switch rule.Severity {
	case "critical":
		baseConfidence += 0.2
	case "high":
		baseConfidence += 0.1
	case "medium":
		// No adjustment
	case "low":
		baseConfidence -= 0.1
	}

	// Ensure confidence is between 0 and 1
	if baseConfidence > 1.0 {
		baseConfidence = 1.0
	} else if baseConfidence < 0.0 {
		baseConfidence = 0.0
	}

	return baseConfidence
}

// getActionPriority returns priority for action types
func (s *Service) getActionPriority(action string) int {
	switch action {
	case "block":
		return 3
	case "warn":
		return 2
	case "allow":
		return 1
	default:
		return 0
	}
}

// getSeverityFromAction converts action to severity level
func (s *Service) getSeverityFromAction(action string) string {
	switch action {
	case "block":
		return "critical"
	case "warn":
		return "warning"
	case "allow":
		return "info"
	default:
		return "info"
	}
}

// containsString checks if a slice contains a string
func (s *Service) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateScanID generates a unique scan ID
func (s *Service) generateScanID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "dlp_scan_" + hex.EncodeToString(bytes)
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
