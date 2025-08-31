package dlp

import (
	"context"
	"testing"
	"time"

	"secure-email-mvp/pkg/models"
)

// MockDatabase implements the Database interface for testing
type MockDatabase struct {
	rules []models.DLPRule
}

func (m *MockDatabase) GetActiveDLPRules() ([]models.DLPRule, error) {
	return m.rules, nil
}

func (m *MockDatabase) CreateDLPScanResult(result *models.DLPScanResult) error {
	return nil
}

func (m *MockDatabase) CreateComplianceAuditLog(audit *models.ComplianceAuditLog) error {
	return nil
}

func TestScanBuiltInPatterns(t *testing.T) {
	// Create a mock database
	mockDB := &MockDatabase{
		rules: []models.DLPRule{
			{
				RuleID:      "test_rule",
				RuleName:    "Test Rule",
				RuleType:    "keyword",
				Pattern:     "test",
				Severity:    "medium",
				Action:      "warn",
				IsActive:    true,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Priority:    1,
			},
		},
	}

	service := &Service{
		db: mockDB,
	}

	tests := []struct {
		name     string
		content  string
		expected int // expected number of violations
	}{
		{
			name:     "Email address detection",
			content:  "Contact me at john.doe@example.com for more information",
			expected: 1,
		},
		{
			name:     "Phone number detection",
			content:  "Call me at 555-123-4567 or 555 987 6543",
			expected: 2,
		},
		{
			name:     "Credit card detection",
			content:  "My card number is 1234-5678-9012-3456",
			expected: 1,
		},
		{
			name:     "SSN detection",
			content:  "My SSN is 123-45-6789",
			expected: 1,
		},
		{
			name:     "Multiple patterns",
			content:  "Email: test@example.com, Phone: 555-123-4567, SSN: 123-45-6789",
			expected: 4, // Email + Phone + SSN + Phone (matches both phone numbers)
		},
		{
			name:     "No sensitive data",
			content:  "This is just regular text with no sensitive information",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := models.DLPScanRequest{
				Content:     tt.content,
				ContentType: "email_body",
				LinkID:      "test-link-123",
			}

			resp, err := service.ScanContent(context.Background(), req)
			if err != nil {
				t.Fatalf("ScanContent failed: %v", err)
			}

			if len(resp.Violations) != tt.expected {
				t.Errorf("Expected %d violations, got %d", tt.expected, len(resp.Violations))
			}

			// Verify response structure
			if !resp.Success {
				t.Error("Expected success to be true")
			}

			if resp.ActionTaken == "" {
				t.Error("Expected ActionTaken to be set")
			}
		})
	}
}

func TestScanRegex(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		content  string
		pattern  string
		expected int
	}{
		{
			name:     "Email pattern",
			content:  "Contact: john@example.com and jane@test.org",
			pattern:  `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b`,
			expected: 2,
		},
		{
			name:     "Phone pattern",
			content:  "Call 555-123-4567 or 555 987 6543",
			pattern:  `\b\d{3}[-\s]?\d{3}[-\s]?\d{4}\b`,
			expected: 2,
		},
		{
			name:     "Credit card pattern",
			content:  "Card: 1234-5678-9012-3456",
			pattern:  `\b\d{4}[-\s]?\d{4}[-\s]?\d{4}[-\s]?\d{4}\b`,
			expected: 1,
		},
		{
			name:     "Invalid regex",
			content:  "Test content",
			pattern:  `[invalid`,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := service.scanRegex(tt.content, tt.pattern)
			if len(matches) != tt.expected {
				t.Errorf("Expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestScanKeywords(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		content  string
		keywords string
		expected int
	}{
		{
			name:     "Single keyword",
			content:  "This is a confidential document",
			keywords: "confidential",
			expected: 1,
		},
		{
			name:     "Multiple keywords",
			content:  "This contains confidential and secret information",
			keywords: "confidential|secret|private",
			expected: 2,
		},
		{
			name:     "Case insensitive",
			content:  "This is CONFIDENTIAL information",
			keywords: "confidential",
			expected: 1,
		},
		{
			name:     "No matches",
			content:  "This is regular text",
			keywords: "confidential|secret",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := service.scanKeywords(tt.content, tt.keywords)
			if len(matches) != tt.expected {
				t.Errorf("Expected %d matches, got %d", tt.expected, len(matches))
			}
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	service := &Service{}

	rule := models.DLPRule{
		RuleType: "regex",
		Severity: "high",
	}

	confidence := service.calculateConfidence("test-match", rule)
	if confidence < 0.0 || confidence > 1.0 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}

	// Test with longer match
	confidence = service.calculateConfidence("this-is-a-longer-match-that-should-have-higher-confidence", rule)
	if confidence < 0.0 || confidence > 1.0 {
		t.Errorf("Confidence should be between 0 and 1, got %f", confidence)
	}
}

func TestGetActionPriority(t *testing.T) {
	service := &Service{}

	tests := []struct {
		action   string
		expected int
	}{
		{"block", 3},
		{"warn", 2},
		{"allow", 1},
		{"unknown", 0},
	}

	for _, tt := range tests {
		priority := service.getActionPriority(tt.action)
		if priority != tt.expected {
			t.Errorf("Expected priority %d for action %s, got %d", tt.expected, tt.action, priority)
		}
	}
}

func TestGenerateScanID(t *testing.T) {
	service := &Service{}

	id1 := service.generateScanID()
	id2 := service.generateScanID()

	if id1 == id2 {
		t.Error("Generated scan IDs should be unique")
	}

	if len(id1) == 0 {
		t.Error("Generated scan ID should not be empty")
	}
}

func TestGenerateAuditID(t *testing.T) {
	service := &Service{}

	id1 := service.generateAuditID()
	id2 := service.generateAuditID()

	if id1 == id2 {
		t.Error("Generated audit IDs should be unique")
	}

	if len(id1) == 0 {
		t.Error("Generated audit ID should not be empty")
	}
}
