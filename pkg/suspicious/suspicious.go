package suspicious

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// DetectionType represents the type of suspicious activity detected
type DetectionType string

const (
	DetectionTypeMultipleFailedAttempts DetectionType = "multiple_failed_attempts"
	DetectionTypeUnusualGeolocation     DetectionType = "unusual_geolocation"
	DetectionTypeRapidMultipleIPs       DetectionType = "rapid_multiple_ips"
	DetectionTypeImpossibleTravel       DetectionType = "impossible_travel"
)

// Severity represents the severity level of a detection
type Severity string

const (
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

// DetectionRule represents a configurable detection rule
type DetectionRule struct {
	RuleID            string    `json:"rule_id"`
	RuleName          string    `json:"rule_name"`
	RuleType          string    `json:"rule_type"`
	IsEnabled         bool      `json:"is_enabled"`
	ThresholdValue    int       `json:"threshold_value"`
	TimeWindowMinutes int       `json:"time_window_minutes"`
	Severity          Severity  `json:"severity"`
	Description       string    `json:"description"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// SuspiciousAccessEvent represents a detected suspicious access event
type SuspiciousAccessEvent struct {
	DetectionID       string                 `json:"detection_id"`
	EmailID           string                 `json:"email_id"`
	DetectionType     DetectionType          `json:"detection_type"`
	DetectionRule     string                 `json:"detection_rule"`
	Severity          Severity               `json:"severity"`
	TriggeredAt       time.Time              `json:"triggered_at"`
	ResolvedAt        *time.Time             `json:"resolved_at,omitempty"`
	ResolvedBy        *string                `json:"resolved_by,omitempty"`
	ResolutionNotes   *string                `json:"resolution_notes,omitempty"`
	DetectionMetadata map[string]interface{} `json:"detection_metadata,omitempty"`
}

// SuspiciousAccessPattern represents a detected access pattern
type SuspiciousAccessPattern struct {
	PatternID       string                 `json:"pattern_id"`
	EmailID         string                 `json:"email_id"`
	PatternType     string                 `json:"pattern_type"`
	PatternData     map[string]interface{} `json:"pattern_data"`
	FirstDetectedAt time.Time              `json:"first_detected_at"`
	LastUpdatedAt   time.Time              `json:"last_updated_at"`
	IsActive        bool                   `json:"is_active"`
	ConfidenceScore float64                `json:"confidence_score"`
}

// UserSuspiciousActivityPreferences represents user preferences for suspicious activity detection
type UserSuspiciousActivityPreferences struct {
	UserID                         string    `json:"user_id"`
	EnableSuspiciousDetection      bool      `json:"enable_suspicious_detection"`
	NotifyOnSuspiciousActivity     bool      `json:"notify_on_suspicious_activity"`
	AutoFlagSuspiciousEmails       bool      `json:"auto_flag_suspicious_emails"`
	MinimumSeverityForNotification Severity  `json:"minimum_severity_for_notification"`
	CreatedAt                      time.Time `json:"created_at"`
	UpdatedAt                      time.Time `json:"updated_at"`
}

// DetectionMetadata represents metadata for different detection types
type DetectionMetadata struct {
	FailedAttempts    []time.Time `json:"failed_attempts,omitempty"`
	IPAddresses       []string    `json:"ip_addresses,omitempty"`
	Geolocations      []string    `json:"geolocations,omitempty"`
	TimeWindow        string      `json:"time_window,omitempty"`
	ThresholdExceeded int         `json:"threshold_exceeded,omitempty"`
	PreviousLocations []string    `json:"previous_locations,omitempty"`
	CurrentLocation   string      `json:"current_location,omitempty"`
	TravelTime        string      `json:"travel_time,omitempty"`
	Distance          float64     `json:"distance,omitempty"`
}

// SuspiciousDetectionService provides methods for detecting suspicious access patterns
type SuspiciousDetectionService struct {
	db *sql.DB
}

// NewSuspiciousDetectionService creates a new suspicious detection service
func NewSuspiciousDetectionService(db *sql.DB) *SuspiciousDetectionService {
	return &SuspiciousDetectionService{
		db: db,
	}
}

// ProcessAccessEvent processes a new access event and checks for suspicious patterns
func (s *SuspiciousDetectionService) ProcessAccessEvent(ctx context.Context, emailID, userID, ipAddress, userAgent, country, city, deviceType, failureReason string, eventType string) error {
	// Get user preferences
	prefs, err := s.GetUserPreferences(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user preferences for suspicious detection: %v", err)
		return err
	}

	// Skip detection if disabled
	if !prefs.EnableSuspiciousDetection {
		return nil
	}

	// Get enabled detection rules
	rules, err := s.GetEnabledDetectionRules(ctx)
	if err != nil {
		log.Printf("Failed to get detection rules: %v", err)
		return err
	}

	// Check each rule
	for _, rule := range rules {
		if err := s.checkDetectionRule(ctx, emailID, userID, ipAddress, country, city, eventType, failureReason, rule); err != nil {
			log.Printf("Failed to check detection rule %s: %v", rule.RuleName, err)
			// Continue with other rules
		}
	}

	return nil
}

// checkDetectionRule checks if a specific detection rule is triggered
func (s *SuspiciousDetectionService) checkDetectionRule(ctx context.Context, emailID, _, ipAddress, country, city, _, _ string, rule DetectionRule) error {
	switch rule.RuleType {
	case string(DetectionTypeMultipleFailedAttempts):
		return s.checkMultipleFailedAttempts(ctx, emailID, rule)
	case string(DetectionTypeUnusualGeolocation):
		return s.checkUnusualGeolocation(ctx, emailID, country, city, rule)
	case string(DetectionTypeRapidMultipleIPs):
		return s.checkRapidMultipleIPs(ctx, emailID, ipAddress, rule)
	case string(DetectionTypeImpossibleTravel):
		return s.checkImpossibleTravel(ctx, emailID, country, city, rule)
	default:
		return fmt.Errorf("unknown detection rule type: %s", rule.RuleType)
	}
}

// checkMultipleFailedAttempts checks for multiple failed attempts within a time window
func (s *SuspiciousDetectionService) checkMultipleFailedAttempts(ctx context.Context, emailID string, rule DetectionRule) error {
	// Get failed attempts within the time window
	query := `
		SELECT COUNT(*) FROM access_events 
		WHERE email_id = ? 
		AND event_type = 'failure' 
		AND timestamp >= datetime('now', '-' || ? || ' minutes')
	`

	var failedAttempts int
	err := s.db.QueryRowContext(ctx, query, emailID, rule.TimeWindowMinutes).Scan(&failedAttempts)
	if err != nil {
		return fmt.Errorf("failed to count failed attempts: %w", err)
	}

	// Check if threshold is exceeded
	if failedAttempts >= rule.ThresholdValue {
		// Get the failed attempts for metadata
		failedAttemptsQuery := `
			SELECT timestamp FROM access_events 
			WHERE email_id = ? 
			AND event_type = 'failure' 
			AND timestamp >= datetime('now', '-' || ? || ' minutes')
			ORDER BY timestamp DESC
		`

		rows, err := s.db.QueryContext(ctx, failedAttemptsQuery, emailID, rule.TimeWindowMinutes)
		if err != nil {
			return fmt.Errorf("failed to get failed attempts: %w", err)
		}
		defer rows.Close()

		var timestamps []time.Time
		for rows.Next() {
			var timestamp time.Time
			if err := rows.Scan(&timestamp); err != nil {
				return fmt.Errorf("failed to scan timestamp: %w", err)
			}
			timestamps = append(timestamps, timestamp)
		}

		// Create detection metadata
		metadata := DetectionMetadata{
			FailedAttempts:    timestamps,
			ThresholdExceeded: failedAttempts,
			TimeWindow:        fmt.Sprintf("%d minutes", rule.TimeWindowMinutes),
		}

		// Record the detection
		return s.recordDetection(ctx, emailID, DetectionTypeMultipleFailedAttempts, rule.RuleID, rule.Severity, metadata)
	}

	return nil
}

// checkUnusualGeolocation checks for access from unusual geolocations
func (s *SuspiciousDetectionService) checkUnusualGeolocation(ctx context.Context, emailID, country, city string, rule DetectionRule) error {
	// Get previous successful access locations
	query := `
		SELECT DISTINCT country, city FROM access_events 
		WHERE email_id = ? 
		AND event_type = 'success' 
		AND country IS NOT NULL 
		AND country != ''
		ORDER BY timestamp DESC
		LIMIT 10
	`

	rows, err := s.db.QueryContext(ctx, query, emailID)
	if err != nil {
		return fmt.Errorf("failed to get previous locations: %w", err)
	}
	defer rows.Close()

	var previousLocations []string
	for rows.Next() {
		var prevCountry, prevCity string
		if err := rows.Scan(&prevCountry, &prevCity); err != nil {
			return fmt.Errorf("failed to scan location: %w", err)
		}
		location := fmt.Sprintf("%s, %s", prevCity, prevCountry)
		previousLocations = append(previousLocations, location)
	}

	// Check if current location is unusual
	currentLocation := fmt.Sprintf("%s, %s", city, country)
	isUnusual := true

	for _, prevLocation := range previousLocations {
		if prevLocation == currentLocation {
			isUnusual = false
			break
		}
	}

	if isUnusual && len(previousLocations) > 0 {
		// Create detection metadata
		metadata := DetectionMetadata{
			PreviousLocations: previousLocations,
			CurrentLocation:   currentLocation,
		}

		// Record the detection
		return s.recordDetection(ctx, emailID, DetectionTypeUnusualGeolocation, rule.RuleID, rule.Severity, metadata)
	}

	return nil
}

// checkRapidMultipleIPs checks for rapid access from multiple IP addresses
func (s *SuspiciousDetectionService) checkRapidMultipleIPs(ctx context.Context, emailID, _ string, rule DetectionRule) error {
	// Get unique IP addresses within the time window
	query := `
		SELECT COUNT(DISTINCT ip_address) FROM access_events 
		WHERE email_id = ? 
		AND timestamp >= datetime('now', '-' || ? || ' minutes')
	`

	var uniqueIPs int
	err := s.db.QueryRowContext(ctx, query, emailID, rule.TimeWindowMinutes).Scan(&uniqueIPs)
	if err != nil {
		return fmt.Errorf("failed to count unique IPs: %w", err)
	}

	// Check if threshold is exceeded
	if uniqueIPs >= rule.ThresholdValue {
		// Get the IP addresses for metadata
		ipsQuery := `
			SELECT DISTINCT ip_address FROM access_events 
			WHERE email_id = ? 
			AND timestamp >= datetime('now', '-' || ? || ' minutes')
		`

		rows, err := s.db.QueryContext(ctx, ipsQuery, emailID, rule.TimeWindowMinutes)
		if err != nil {
			return fmt.Errorf("failed to get IP addresses: %w", err)
		}
		defer rows.Close()

		var ipAddresses []string
		for rows.Next() {
			var ip string
			if err := rows.Scan(&ip); err != nil {
				return fmt.Errorf("failed to scan IP: %w", err)
			}
			ipAddresses = append(ipAddresses, ip)
		}

		// Create detection metadata
		metadata := DetectionMetadata{
			IPAddresses:       ipAddresses,
			ThresholdExceeded: uniqueIPs,
			TimeWindow:        fmt.Sprintf("%d minutes", rule.TimeWindowMinutes),
		}

		// Record the detection
		return s.recordDetection(ctx, emailID, DetectionTypeRapidMultipleIPs, rule.RuleID, rule.Severity, metadata)
	}

	return nil
}

// checkImpossibleTravel checks for geographically impossible travel patterns
func (s *SuspiciousDetectionService) checkImpossibleTravel(ctx context.Context, emailID, country, city string, rule DetectionRule) error {
	// Get the most recent successful access
	query := `
		SELECT country, city, timestamp FROM access_events 
		WHERE email_id = ? 
		AND event_type = 'success' 
		AND country IS NOT NULL 
		AND country != ''
		ORDER BY timestamp DESC 
		LIMIT 1
	`

	var prevCountry, prevCity string
	var prevTimestamp time.Time
	err := s.db.QueryRowContext(ctx, query, emailID).Scan(&prevCountry, &prevCity, &prevTimestamp)
	if err != nil {
		if err == sql.ErrNoRows {
			// No previous successful access, skip this check
			return nil
		}
		return fmt.Errorf("failed to get previous access: %w", err)
	}

	// Calculate time difference
	timeDiff := time.Since(prevTimestamp)

	// Simple distance calculation (this could be enhanced with actual geolocation APIs)
	// For now, we'll use a simple heuristic: if locations are different and time is very short
	if prevCountry != country && timeDiff.Minutes() < float64(rule.TimeWindowMinutes) {
		// Create detection metadata
		metadata := DetectionMetadata{
			PreviousLocations: []string{fmt.Sprintf("%s, %s", prevCity, prevCountry)},
			CurrentLocation:   fmt.Sprintf("%s, %s", city, country),
			TravelTime:        timeDiff.String(),
		}

		// Record the detection
		return s.recordDetection(ctx, emailID, DetectionTypeImpossibleTravel, rule.RuleID, rule.Severity, metadata)
	}

	return nil
}

// recordDetection records a suspicious access detection
func (s *SuspiciousDetectionService) recordDetection(ctx context.Context, emailID string, detectionType DetectionType, ruleID string, severity Severity, metadata DetectionMetadata) error {
	detectionID := uuid.New().String()

	// Convert metadata to JSON
	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert detection event
	query := `
		INSERT INTO suspicious_access_events (
			detection_id, email_id, detection_type, detection_rule, severity, detection_metadata
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err = s.db.ExecContext(ctx, query, detectionID, emailID, detectionType, ruleID, severity, string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to insert detection event: %w", err)
	}

	// Set suspicious flag on email if auto-flag is enabled
	if err := s.setSuspiciousFlag(ctx, emailID); err != nil {
		log.Printf("Failed to set suspicious flag: %v", err)
		// Don't return error as the detection was already recorded
	}

	log.Printf("Suspicious access detection recorded: %s for email %s (severity: %s)", detectionType, emailID, severity)
	return nil
}

// setSuspiciousFlag sets the suspicious flag on an email
func (s *SuspiciousDetectionService) setSuspiciousFlag(ctx context.Context, emailID string) error {
	query := `UPDATE emails SET suspicious_flag = TRUE WHERE email_id = ?`
	_, err := s.db.ExecContext(ctx, query, emailID)
	if err != nil {
		return fmt.Errorf("failed to set suspicious flag: %w", err)
	}
	return nil
}

// GetUserPreferences gets user preferences for suspicious activity detection
func (s *SuspiciousDetectionService) GetUserPreferences(ctx context.Context, userID string) (*UserSuspiciousActivityPreferences, error) {
	query := `
		SELECT user_id, enable_suspicious_detection, notify_on_suspicious_activity, 
		       auto_flag_suspicious_emails, minimum_severity_for_notification, created_at, updated_at
		FROM user_suspicious_activity_preferences 
		WHERE user_id = ?
	`

	var prefs UserSuspiciousActivityPreferences
	err := s.db.QueryRowContext(ctx, query, userID).Scan(
		&prefs.UserID, &prefs.EnableSuspiciousDetection, &prefs.NotifyOnSuspiciousActivity,
		&prefs.AutoFlagSuspiciousEmails, &prefs.MinimumSeverityForNotification,
		&prefs.CreatedAt, &prefs.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			// Return default preferences
			return &UserSuspiciousActivityPreferences{
				UserID:                         userID,
				EnableSuspiciousDetection:      true,
				NotifyOnSuspiciousActivity:     true,
				AutoFlagSuspiciousEmails:       true,
				MinimumSeverityForNotification: SeverityMedium,
			}, nil
		}
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}

	return &prefs, nil
}

// GetEnabledDetectionRules gets all enabled detection rules
func (s *SuspiciousDetectionService) GetEnabledDetectionRules(ctx context.Context) ([]DetectionRule, error) {
	query := `
		SELECT rule_id, rule_name, rule_type, is_enabled, threshold_value, 
		       time_window_minutes, severity, description, created_at, updated_at
		FROM detection_rules 
		WHERE is_enabled = TRUE
		ORDER BY severity DESC, rule_name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get detection rules: %w", err)
	}
	defer rows.Close()

	var rules []DetectionRule
	for rows.Next() {
		var rule DetectionRule
		err := rows.Scan(
			&rule.RuleID, &rule.RuleName, &rule.RuleType, &rule.IsEnabled,
			&rule.ThresholdValue, &rule.TimeWindowMinutes, &rule.Severity,
			&rule.Description, &rule.CreatedAt, &rule.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan detection rule: %w", err)
		}
		rules = append(rules, rule)
	}

	return rules, nil
}

// GetSuspiciousAccessEvents gets suspicious access events for an email
func (s *SuspiciousDetectionService) GetSuspiciousAccessEvents(ctx context.Context, emailID string, limit int) ([]SuspiciousAccessEvent, error) {
	query := `
		SELECT detection_id, email_id, detection_type, detection_rule, severity, 
		       triggered_at, resolved_at, resolved_by, resolution_notes, detection_metadata
		FROM suspicious_access_events 
		WHERE email_id = ? 
		ORDER BY triggered_at DESC 
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, emailID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get suspicious access events: %w", err)
	}
	defer rows.Close()

	var events []SuspiciousAccessEvent
	for rows.Next() {
		var event SuspiciousAccessEvent
		var metadataJSON string
		err := rows.Scan(
			&event.DetectionID, &event.EmailID, &event.DetectionType, &event.DetectionRule,
			&event.Severity, &event.TriggeredAt, &event.ResolvedAt, &event.ResolvedBy,
			&event.ResolutionNotes, &metadataJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan suspicious access event: %w", err)
		}

		// Parse metadata JSON
		if metadataJSON != "" {
			if err := json.Unmarshal([]byte(metadataJSON), &event.DetectionMetadata); err != nil {
				log.Printf("Failed to unmarshal detection metadata: %v", err)
			}
		}

		events = append(events, event)
	}

	return events, nil
}

// ClearSuspiciousFlag clears the suspicious flag on an email
func (s *SuspiciousDetectionService) ClearSuspiciousFlag(ctx context.Context, emailID, clearedBy string) error {
	query := `
		UPDATE emails 
		SET suspicious_flag = FALSE, suspicious_flag_cleared_by = ? 
		WHERE email_id = ?
	`
	_, err := s.db.ExecContext(ctx, query, clearedBy, emailID)
	if err != nil {
		return fmt.Errorf("failed to clear suspicious flag: %w", err)
	}
	return nil
}

// ResolveDetectionEvent marks a detection event as resolved
func (s *SuspiciousDetectionService) ResolveDetectionEvent(ctx context.Context, detectionID, resolvedBy, resolutionNotes string) error {
	query := `
		UPDATE suspicious_access_events 
		SET resolved_at = CURRENT_TIMESTAMP, resolved_by = ?, resolution_notes = ? 
		WHERE detection_id = ?
	`
	_, err := s.db.ExecContext(ctx, query, resolvedBy, resolutionNotes, detectionID)
	if err != nil {
		return fmt.Errorf("failed to resolve detection event: %w", err)
	}
	return nil
}




