package models

import (
	"time"
)

// AIContentClassification represents AI-based content classification
type AIContentClassification struct {
	Category     string   `json:"category"`      // 'pii', 'financial', 'healthcare', 'legal', 'confidential', 'none'
	Confidence   float64  `json:"confidence"`    // 0.0 to 1.0
	Severity     string   `json:"severity"`      // 'low', 'medium', 'high', 'critical'
	RiskScore    float64  `json:"risk_score"`    // 0.0 to 1.0
	Keywords     []string `json:"keywords"`      // Detected sensitive keywords
	Entities     []Entity `json:"entities"`      // Named entities detected
	Context      string   `json:"context"`       // Contextual information
	ModelVersion string   `json:"model_version"` // AI model version used
}

// Entity represents a named entity detected by AI
type Entity struct {
	Type       string  `json:"type"`       // 'person', 'organization', 'location', 'date', 'amount', 'account'
	Value      string  `json:"value"`      // The actual entity value
	Confidence float64 `json:"confidence"` // Entity detection confidence
	StartPos   int     `json:"start_pos"`  // Start position in text
	EndPos     int     `json:"end_pos"`    // End position in text
}

// AIDLPScanResult represents AI-powered DLP scan results
type AIDLPScanResult struct {
	ScanID            string                   `json:"scan_id" db:"scan_id"`
	LinkID            *string                  `json:"link_id,omitempty" db:"link_id"`
	ReplyID           *string                  `json:"reply_id,omitempty" db:"reply_id"`
	AttachmentID      *string                  `json:"attachment_id,omitempty" db:"attachment_id"`
	ContentType       string                   `json:"content_type" db:"content_type"` // 'email_body', 'reply_body', 'attachment'
	ContentHash       string                   `json:"content_hash" db:"content_hash"`
	Classification    *AIContentClassification `json:"classification,omitempty" db:"classification"`
	SeverityScore     float64                  `json:"severity_score" db:"severity_score"`
	RiskLevel         string                   `json:"risk_level" db:"risk_level"`                 // 'low', 'medium', 'high', 'critical'
	ActionRecommended string                   `json:"action_recommended" db:"action_recommended"` // 'allow', 'warn', 'block', 'review'
	ActionTaken       string                   `json:"action_taken" db:"action_taken"`             // 'allowed', 'warned', 'blocked', 'overridden'
	OverrideReason    *string                  `json:"override_reason,omitempty" db:"override_reason"`
	OverrideBy        *string                  `json:"override_by,omitempty" db:"override_by"`
	ModelVersion      string                   `json:"model_version" db:"model_version"`
	ProcessingTime    float64                  `json:"processing_time" db:"processing_time"` // milliseconds
	ScanTimestamp     time.Time                `json:"scan_timestamp" db:"scan_timestamp"`
	CreatedBy         *string                  `json:"created_by,omitempty" db:"created_by"`
}

// AIDLPPolicy represents AI DLP policy configuration
type AIDLPPolicy struct {
	PolicyID            string             `json:"policy_id" db:"policy_id"`
	PolicyName          string             `json:"policy_name" db:"policy_name"`
	Description         *string            `json:"description,omitempty" db:"description"`
	IsActive            bool               `json:"is_active" db:"is_active"`
	Categories          []string           `json:"categories" db:"categories"`                   // JSON array of categories to scan
	SeverityThresholds  map[string]float64 `json:"severity_thresholds" db:"severity_thresholds"` // JSON map of severity thresholds
	Actions             map[string]string  `json:"actions" db:"actions"`                         // JSON map of severity -> action
	ConfidenceThreshold float64            `json:"confidence_threshold" db:"confidence_threshold"`
	RiskThreshold       float64            `json:"risk_threshold" db:"risk_threshold"`
	AllowOverride       bool               `json:"allow_override" db:"allow_override"`
	OverrideRoles       []string           `json:"override_roles" db:"override_roles"` // JSON array of roles that can override
	CreatedAt           time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time          `json:"updated_at" db:"updated_at"`
	CreatedBy           *string            `json:"created_by,omitempty" db:"created_by"`
}

// AIDLPScanRequest represents a request for AI DLP scanning
type AIDLPScanRequest struct {
	Content      string  `json:"content"`
	ContentType  string  `json:"content_type"` // 'email_body', 'reply_body', 'attachment'
	LinkID       string  `json:"link_id"`
	ReplyID      *string `json:"reply_id,omitempty"`
	AttachmentID *string `json:"attachment_id,omitempty"`
	PolicyID     *string `json:"policy_id,omitempty"` // Optional specific policy
	UserID       *string `json:"user_id,omitempty"`
	UserRole     *string `json:"user_role,omitempty"`
}

// AIDLPScanResponse represents the response from AI DLP scanning
type AIDLPScanResponse struct {
	Success           bool                     `json:"success"`
	ScanID            string                   `json:"scan_id"`
	Classification    *AIContentClassification `json:"classification,omitempty"`
	SeverityScore     float64                  `json:"severity_score"`
	RiskLevel         string                   `json:"risk_level"`
	ActionRecommended string                   `json:"action_recommended"`
	ActionTaken       string                   `json:"action_taken"`
	CanOverride       bool                     `json:"can_override"`
	OverrideReason    *string                  `json:"override_reason,omitempty"`
	ProcessingTime    float64                  `json:"processing_time"`
	ModelVersion      string                   `json:"model_version"`
	Message           string                   `json:"message,omitempty"`
	Error             string                   `json:"error,omitempty"`
	ErrorCode         string                   `json:"error_code,omitempty"`
}

// AIDLPOverrideRequest represents a request to override AI DLP decision
type AIDLPOverrideRequest struct {
	ScanID         string `json:"scan_id"`
	OverrideReason string `json:"override_reason"`
	UserID         string `json:"user_id"`
	UserRole       string `json:"user_role"`
	Justification  string `json:"justification"`
}

// AIDLPOverrideResponse represents the response from override request
type AIDLPOverrideResponse struct {
	Success     bool   `json:"success"`
	OverrideID  string `json:"override_id"`
	ActionTaken string `json:"action_taken"`
	Message     string `json:"message,omitempty"`
	Error       string `json:"error,omitempty"`
	ErrorCode   string `json:"error_code,omitempty"`
}

// AIDLPMetrics represents AI DLP performance metrics
type AIDLPMetrics struct {
	TotalScans     int64     `json:"total_scans"`
	AverageTime    float64   `json:"average_time"`
	Accuracy       float64   `json:"accuracy"`
	FalsePositives int64     `json:"false_positives"`
	FalseNegatives int64     `json:"false_negatives"`
	Overrides      int64     `json:"overrides"`
	BlockedContent int64     `json:"blocked_content"`
	WarnedContent  int64     `json:"warned_content"`
	AllowedContent int64     `json:"allowed_content"`
	ModelVersion   string    `json:"model_version"`
	LastUpdated    time.Time `json:"last_updated"`
}

// ContentCategory represents predefined content categories
type ContentCategory struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Patterns    []string `json:"patterns"`
	Severity    string   `json:"severity"`
	RiskWeight  float64  `json:"risk_weight"`
}

// Predefined content categories
var ContentCategories = map[string]ContentCategory{
	"pii": {
		ID:          "pii",
		Name:        "Personally Identifiable Information",
		Description: "Personal data that can identify individuals",
		Keywords:    []string{"ssn", "social security", "passport", "driver license", "date of birth", "address"},
		Patterns:    []string{`\b\d{3}-\d{2}-\d{4}\b`, `\b[A-Z]{2}\d{7}\b`},
		Severity:    "high",
		RiskWeight:  0.8,
	},
	"financial": {
		ID:          "financial",
		Name:        "Financial Information",
		Description: "Banking, credit card, and financial data",
		Keywords:    []string{"credit card", "bank account", "routing number", "swift code", "iban", "account number"},
		Patterns:    []string{`\b\d{4}[- ]?\d{4}[- ]?\d{4}[- ]?\d{4}\b`, `\b\d{9,17}\b`},
		Severity:    "critical",
		RiskWeight:  0.9,
	},
	"healthcare": {
		ID:          "healthcare",
		Name:        "Healthcare Information",
		Description: "Medical records and health-related data",
		Keywords:    []string{"diagnosis", "treatment", "medication", "patient", "medical record", "phi"},
		Patterns:    []string{`\b[A-Z]{2}\d{10}\b`, `\b\d{10}\b`},
		Severity:    "critical",
		RiskWeight:  0.95,
	},
	"legal": {
		ID:          "legal",
		Name:        "Legal Information",
		Description: "Legal documents and privileged information",
		Keywords:    []string{"attorney", "legal", "privileged", "confidential", "case number", "court"},
		Patterns:    []string{},
		Severity:    "high",
		RiskWeight:  0.7,
	},
	"confidential": {
		ID:          "confidential",
		Name:        "Confidential Information",
		Description: "Trade secrets and confidential business data",
		Keywords:    []string{"confidential", "secret", "proprietary", "internal", "restricted", "classified"},
		Patterns:    []string{},
		Severity:    "medium",
		RiskWeight:  0.6,
	},
}

// GetSeverityThreshold returns the severity threshold for a given severity level
func GetSeverityThreshold(severity string) float64 {
	switch severity {
	case "critical":
		return 0.9
	case "high":
		return 0.7
	case "medium":
		return 0.5
	case "low":
		return 0.3
	default:
		return 0.5
	}
}

// GetActionForSeverity returns the recommended action for a given severity level
func GetActionForSeverity(severity string) string {
	switch severity {
	case "critical":
		return "block"
	case "high":
		return "warn"
	case "medium":
		return "warn"
	case "low":
		return "allow"
	default:
		return "allow"
	}
}

// GetSeverityFromScore returns severity level based on risk score
func GetSeverityFromScore(score float64) string {
	switch {
	case score >= 0.9:
		return "critical"
	case score >= 0.7:
		return "high"
	case score >= 0.5:
		return "medium"
	case score >= 0.3:
		return "low"
	default:
		return "none"
	}
}
