package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
)

// EnhancedAuditEvent represents an enhanced audit log entry with structured data
type EnhancedAuditEvent struct {
	EventID        string                 `json:"event_id"`
	Timestamp      time.Time              `json:"timestamp"`
	EventType      string                 `json:"event_type"`
	Severity       Severity               `json:"severity"`
	Category       string                 `json:"category"`
	UserID         string                 `json:"user_id,omitempty"`
	IPAddress      string                 `json:"ip_address"`
	UserAgent      string                 `json:"user_agent"`
	LinkID         string                 `json:"link_id,omitempty"`
	EmailID        string                 `json:"email_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	RequestID      string                 `json:"request_id,omitempty"`
	Geolocation    *GeolocationData       `json:"geolocation,omitempty"`
	DeviceInfo     *DeviceInfo            `json:"device_info,omitempty"`
	SecurityFlags  *SecurityFlags         `json:"security_flags,omitempty"`
	Outcome        string                 `json:"outcome"`
	Details        map[string]interface{} `json:"details"`
	CorrelationID  string                 `json:"correlation_id,omitempty"`
	ParentEventID  string                 `json:"parent_event_id,omitempty"`
	Tags           []string               `json:"tags,omitempty"`
	RiskScore      float64                `json:"risk_score"`
	IsSuspicious   bool                   `json:"is_suspicious"`
}

// GeolocationData contains geolocation information
type GeolocationData struct {
	Country   string  `json:"country,omitempty"`
	City      string  `json:"city,omitempty"`
	Region    string  `json:"region,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Timezone  string  `json:"timezone,omitempty"`
	ISP       string  `json:"isp,omitempty"`
	ASN       string  `json:"asn,omitempty"`
}

// DeviceInfo contains device fingerprinting information
type DeviceInfo struct {
	DeviceType    string `json:"device_type,omitempty"`
	Browser       string `json:"browser,omitempty"`
	OS            string `json:"os,omitempty"`
	ScreenSize    string `json:"screen_size,omitempty"`
	Language      string `json:"language,omitempty"`
	Timezone      string `json:"timezone,omitempty"`
	Fingerprint   string `json:"fingerprint,omitempty"`
}

// SecurityFlags contains security-related flags
type SecurityFlags struct {
	IsVPN           bool `json:"is_vpn"`
	IsProxy         bool `json:"is_proxy"`
	IsTor           bool `json:"is_tor"`
	IsDataCenter    bool `json:"is_data_center"`
	IsMobile        bool `json:"is_mobile"`
	IsAutomated     bool `json:"is_automated"`
	HasValidSession bool `json:"has_valid_session"`
}

// EnhancedAuditLogger provides enhanced audit logging with structured JSON
type EnhancedAuditLogger struct {
	db *sql.DB
}

// NewEnhancedAuditLogger creates a new enhanced audit logger
func NewEnhancedAuditLogger(db *sql.DB) *EnhancedAuditLogger {
	return &EnhancedAuditLogger{
		db: db,
	}
}

// LogEvent logs an enhanced audit event with structured data
func (eal *EnhancedAuditLogger) LogEvent(ctx context.Context, event *EnhancedAuditEvent) error {
	// Generate event ID if not provided
	if event.EventID == "" {
		event.EventID = uuid.New().String()
	}

	// Set timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Calculate risk score
	event.RiskScore = eal.calculateRiskScore(event)

	// Check for suspicious activity
	event.IsSuspicious = eal.detectSuspiciousActivity(event)

	// Convert to JSON for structured logging
	jsonData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Log to structured log
	log.Printf("[AUDIT] %s", string(jsonData))

	// Store in database
	return eal.storeEvent(ctx, event)
}

// LogSecureLinkAccess logs secure link access events with enhanced context
func (eal *EnhancedAuditLogger) LogSecureLinkAccess(ctx context.Context, linkID, ipAddress, userAgent string, outcome string, details map[string]interface{}) error {
	event := &EnhancedAuditEvent{
		EventType:     "secure_link_access",
		Category:      "access_control",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		LinkID:        linkID,
		Outcome:       outcome,
		Details:       details,
		Severity:      eal.determineSeverity(outcome),
		SecurityFlags: eal.analyzeSecurityFlags(ipAddress, userAgent),
		Tags:          []string{"secure_link", "external_access"},
	}

	return eal.LogEvent(ctx, event)
}

// LogSecurityValidation logs security validation events
func (eal *EnhancedAuditLogger) LogSecurityValidation(ctx context.Context, linkID, ipAddress, userAgent, validationType, outcome string, details map[string]interface{}) error {
	event := &EnhancedAuditEvent{
		EventType:     "security_validation",
		Category:      "authentication",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		LinkID:        linkID,
		Outcome:       outcome,
		Details:       details,
		Severity:      eal.determineSeverity(outcome),
		SecurityFlags: eal.analyzeSecurityFlags(ipAddress, userAgent),
		Tags:          []string{"security_validation", validationType},
	}

	return eal.LogEvent(ctx, event)
}

// LogReplyEvent logs secure link reply events
func (eal *EnhancedAuditLogger) LogReplyEvent(ctx context.Context, linkID, replyID, ipAddress, userAgent string, outcome string, details map[string]interface{}) error {
	event := &EnhancedAuditEvent{
		EventType:     "secure_link_reply",
		Category:      "communication",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		LinkID:        linkID,
		Outcome:       outcome,
		Details:       details,
		Severity:      eal.determineSeverity(outcome),
		SecurityFlags: eal.analyzeSecurityFlags(ipAddress, userAgent),
		Tags:          []string{"secure_reply", "external_communication"},
	}

	return eal.LogEvent(ctx, event)
}

// LogSuspiciousActivity logs suspicious activity events
func (eal *EnhancedAuditLogger) LogSuspiciousActivity(ctx context.Context, activityType, ipAddress, userAgent string, details map[string]interface{}) error {
	event := &EnhancedAuditEvent{
		EventType:     "suspicious_activity",
		Category:      "security_monitoring",
		IPAddress:     ipAddress,
		UserAgent:     userAgent,
		Outcome:       "detected",
		Details:       details,
		Severity:      SeverityWarning,
		SecurityFlags: eal.analyzeSecurityFlags(ipAddress, userAgent),
		Tags:          []string{"suspicious_activity", activityType},
		IsSuspicious:  true,
	}

	return eal.LogEvent(ctx, event)
}

// calculateRiskScore calculates a risk score for the event
func (eal *EnhancedAuditLogger) calculateRiskScore(event *EnhancedAuditEvent) float64 {
	score := 0.0

	// Base score by severity
	switch event.Severity {
	case SeverityCritical:
		score += 100
	case SeverityError:
		score += 75
	case SeverityWarning:
		score += 50
	case SeverityInfo:
		score += 25
	}

	// Risk factors
	if event.SecurityFlags != nil {
		if event.SecurityFlags.IsVPN {
			score += 10
		}
		if event.SecurityFlags.IsProxy {
			score += 15
		}
		if event.SecurityFlags.IsTor {
			score += 25
		}
		if event.SecurityFlags.IsDataCenter {
			score += 20
		}
		if event.SecurityFlags.IsAutomated {
			score += 30
		}
	}

	// Outcome-based scoring
	if event.Outcome == "failure" || event.Outcome == "blocked" {
		score += 25
	}

	// Normalize score to 0-100 range
	if score > 100 {
		score = 100
	}

	return score
}

// detectSuspiciousActivity detects suspicious patterns
func (eal *EnhancedAuditLogger) detectSuspiciousActivity(event *EnhancedAuditEvent) bool {
	// Check for automated access
	if event.SecurityFlags != nil && event.SecurityFlags.IsAutomated {
		return true
	}

	// Check for high-risk IP types
	if event.SecurityFlags != nil {
		if event.SecurityFlags.IsTor || event.SecurityFlags.IsDataCenter {
			return true
		}
	}

	// Check for suspicious user agents
	suspiciousPatterns := []string{
		"bot", "crawler", "spider", "scraper",
		"curl", "wget", "python", "java",
		"automation", "headless",
	}

	userAgentLower := event.UserAgent
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(strings.ToLower(userAgentLower), pattern) {
			return true
		}
	}

	return false
}

// determineSeverity determines the severity level based on outcome
func (eal *EnhancedAuditLogger) determineSeverity(outcome string) Severity {
	switch outcome {
	case "success":
		return SeverityInfo
	case "failure":
		return SeverityWarning
	case "blocked":
		return SeverityError
	case "critical_failure":
		return SeverityCritical
	default:
		return SeverityInfo
	}
}

// analyzeSecurityFlags analyzes security-related flags
func (eal *EnhancedAuditLogger) analyzeSecurityFlags(ipAddress, userAgent string) *SecurityFlags {
	flags := &SecurityFlags{}

	// Analyze IP address (simplified - in production, use IP intelligence service)
	if strings.Contains(ipAddress, "10.") || strings.Contains(ipAddress, "192.168.") {
		flags.IsVPN = true
	}

	// Analyze user agent
	userAgentLower := strings.ToLower(userAgent)
	if strings.Contains(userAgentLower, "mobile") || strings.Contains(userAgentLower, "android") || strings.Contains(userAgentLower, "iphone") {
		flags.IsMobile = true
	}

	if strings.Contains(userAgentLower, "bot") || strings.Contains(userAgentLower, "crawler") {
		flags.IsAutomated = true
	}

	return flags
}

// storeEvent stores the event in the database
func (eal *EnhancedAuditLogger) storeEvent(ctx context.Context, event *EnhancedAuditEvent) error {
	query := `
		INSERT INTO enhanced_audit_log (
			event_id, timestamp, event_type, severity, category, user_id, ip_address, 
			user_agent, link_id, email_id, session_id, request_id, geolocation_data, 
			device_info, security_flags, outcome, details, correlation_id, parent_event_id, 
			tags, risk_score, is_suspicious
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	geolocationJSON, _ := json.Marshal(event.Geolocation)
	deviceInfoJSON, _ := json.Marshal(event.DeviceInfo)
	securityFlagsJSON, _ := json.Marshal(event.SecurityFlags)
	detailsJSON, _ := json.Marshal(event.Details)
	tagsJSON, _ := json.Marshal(event.Tags)

	_, err := eal.db.ExecContext(ctx, query,
		event.EventID, event.Timestamp, event.EventType, event.Severity, event.Category,
		event.UserID, event.IPAddress, event.UserAgent, event.LinkID, event.EmailID,
		event.SessionID, event.RequestID, geolocationJSON, deviceInfoJSON, securityFlagsJSON,
		event.Outcome, detailsJSON, event.CorrelationID, event.ParentEventID, tagsJSON,
		event.RiskScore, event.IsSuspicious,
	)

	return err
}

// GetEvents retrieves audit events with filtering
func (eal *EnhancedAuditLogger) GetEvents(ctx context.Context, filters map[string]interface{}) ([]*EnhancedAuditEvent, error) {
	query := `SELECT * FROM enhanced_audit_log WHERE 1=1`
	args := []interface{}{}

	// Add filters
	if linkID, ok := filters["link_id"].(string); ok {
		query += " AND link_id = ?"
		args = append(args, linkID)
	}

	if severity, ok := filters["severity"].(string); ok {
		query += " AND severity = ?"
		args = append(args, severity)
	}

	if isSuspicious, ok := filters["is_suspicious"].(bool); ok {
		query += " AND is_suspicious = ?"
		args = append(args, isSuspicious)
	}

	query += " ORDER BY timestamp DESC LIMIT 100"

	rows, err := eal.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*EnhancedAuditEvent
	for rows.Next() {
		event := &EnhancedAuditEvent{}
		// Scan the row into the event struct
		// Implementation depends on the exact column structure
		events = append(events, event)
	}

	return events, nil
}
