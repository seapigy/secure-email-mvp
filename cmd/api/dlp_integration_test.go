package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"secure-email-mvp/pkg/models"
	"secure-email-mvp/pkg/securelinks/dlp"
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

// TestDLPScanEndpoint tests the DLP scan endpoint with various payloads
func TestDLPScanEndpoint(t *testing.T) {
	// Create a test server with mock DLP service
	srv := &Server{}
	
	// Create a mock DLP service for testing
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
	
	dlpService := dlp.NewService(mockDB, nil)
	srv.dlpService = dlpService

	tests := []struct {
		name           string
		linkID         string
		payload        models.DLPScanRequest
		expectedStatus int
		expectedError  bool
	}{
		{
			name:   "Email address detection",
			linkID: "test-link-123",
			payload: models.DLPScanRequest{
				Content:     "Contact me at john.doe@example.com for more information",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Credit card detection",
			linkID: "test-link-456",
			payload: models.DLPScanRequest{
				Content:     "My credit card number is 1234-5678-9012-3456",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "SSN detection",
			linkID: "test-link-789",
			payload: models.DLPScanRequest{
				Content:     "My SSN is 123-45-6789",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Phone number detection",
			linkID: "test-link-101",
			payload: models.DLPScanRequest{
				Content:     "Call me at 555-123-4567",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Multiple sensitive data",
			linkID: "test-link-202",
			payload: models.DLPScanRequest{
				Content:     "Email: test@example.com, Phone: 555-123-4567, SSN: 123-45-6789",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Clean content",
			linkID: "test-link-303",
			payload: models.DLPScanRequest{
				Content:     "This is just regular text with no sensitive information",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
		{
			name:   "Empty content",
			linkID: "test-link-404",
			payload: models.DLPScanRequest{
				Content:     "",
				ContentType: "email_body",
			},
			expectedStatus: http.StatusOK,
			expectedError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set the LinkID from the test case
			tt.payload.LinkID = tt.linkID

			// Create request body
			body, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/v/"+tt.linkID+"/dlp/scan", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			srv.dlpScanHandler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// If we expect an error, don't parse the response
			if tt.expectedError {
				return
			}

			// Parse response
			var response models.DLPScanResponse
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Verify response structure
			if !response.Success {
				t.Error("Expected success to be true")
			}

			if response.ActionTaken == "" {
				t.Error("Expected ActionTaken to be set")
			}

			// Log findings for debugging
			t.Logf("Test: %s, Violations: %d, Action: %s", tt.name, len(response.Violations), response.ActionTaken)
		})
	}
}

// TestDLPScanEndpointInvalidRequest tests the DLP scan endpoint with invalid requests
func TestDLPScanEndpointInvalidRequest(t *testing.T) {
	srv := &Server{}
	
	// Create a mock DLP service for testing
	mockDB := &MockDatabase{
		rules: []models.DLPRule{},
	}
	
	dlpService := dlp.NewService(mockDB, nil)
	srv.dlpService = dlpService

	tests := []struct {
		name           string
		linkID         string
		body           string
		expectedStatus int
	}{
		{
			name:           "Invalid JSON",
			linkID:         "test-link-123",
			body:           `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing content",
			linkID:         "test-link-456",
			body:           `{"content_type": "email_body"}`,
			expectedStatus: http.StatusOK, // This should still work as content can be empty
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/v/"+tt.linkID+"/dlp/scan", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call the handler
			srv.dlpScanHandler(w, req)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

// TestDLPScanResponseFormat tests that the DLP scan response format is correct
func TestDLPScanResponseFormat(t *testing.T) {
	srv := &Server{}
	
	// Create a mock DLP service for testing
	mockDB := &MockDatabase{
		rules: []models.DLPRule{},
	}
	
	dlpService := dlp.NewService(mockDB, nil)
	srv.dlpService = dlpService

	// Test payload with known sensitive data
	payload := models.DLPScanRequest{
		Content:     "My credit card is 1234-5678-9012-3456 and SSN is 123-45-6789",
		ContentType: "email_body",
		LinkID:      "test-link-format",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest("POST", "/api/v/test-link-format/dlp/scan", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.dlpScanHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response models.DLPScanResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify response format
	if !response.Success {
		t.Error("Expected success to be true")
	}

	if response.ActionTaken == "" {
		t.Error("Expected ActionTaken to be set")
	}

	if response.Message == "" {
		t.Error("Expected Message to be set")
	}

	// Should have violations for credit card and SSN
	if len(response.Violations) < 2 {
		t.Errorf("Expected at least 2 violations, got %d", len(response.Violations))
	}

	// Check that violations have required fields
	for _, violation := range response.Violations {
		if violation.ScanID == "" {
			t.Error("Expected ScanID to be set")
		}
		if violation.RuleID == "" {
			t.Error("Expected RuleID to be set")
		}
		if violation.ContentType == "" {
			t.Error("Expected ContentType to be set")
		}
		if violation.ActionTaken == "" {
			t.Error("Expected ActionTaken to be set")
		}
		if violation.MatchedContent == nil {
			t.Error("Expected MatchedContent to be set")
		}
	}
}
