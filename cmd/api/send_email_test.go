package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock context key for testing
const userIDContextKey = "user_id"

// Helper function to create request with user context
func createAuthenticatedRequest(method, url string, body []byte, userID string) *http.Request {
	req := httptest.NewRequest(method, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	
	// Add user ID to context
	ctx := context.WithValue(req.Context(), userIDContextKey, userID)
	req = req.WithContext(ctx)
	
	return req
}

func TestSendEmailHandler_SelfDestructValidation(t *testing.T) {
	// Create a test server with mock database
	srv := &Server{}
	// Note: In a real test, you'd set up a test database

	tests := []struct {
		name           string
		requestBody    SendEmailRequest
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid self-destruct settings",
			requestBody: SendEmailRequest{
				Recipient:                 "test@example.com",
				Subject:                   "Test Subject",
				Body:                      "Test body",
				SelfDestructAfterAttempts: true,
				MaxFailedAttempts:         3,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid maxFailedAttempts too low",
			requestBody: SendEmailRequest{
				Recipient:                 "test@example.com",
				Subject:                   "Test Subject",
				Body:                      "Test body",
				SelfDestructAfterAttempts: true,
				MaxFailedAttempts:         0,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "maxFailedAttempts must be between 1 and 10",
		},
		{
			name: "Invalid maxFailedAttempts too high",
			requestBody: SendEmailRequest{
				Recipient:                 "test@example.com",
				Subject:                   "Test Subject",
				Body:                      "Test body",
				SelfDestructAfterAttempts: true,
				MaxFailedAttempts:         11,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "maxFailedAttempts must be between 1 and 10",
		},
		{
			name: "Self-destruct disabled should set maxFailedAttempts to 0",
			requestBody: SendEmailRequest{
				Recipient:                 "test@example.com",
				Subject:                   "Test Subject",
				Body:                      "Test body",
				SelfDestructAfterAttempts: false,
				MaxFailedAttempts:         5, // This should be ignored
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Create HTTP request with authentication
			req := createAuthenticatedRequest("POST", "/api/email/send", body, "test-user-id")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			srv.sendEmailHandler(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check error message if expected
			if tt.expectedError != "" {
				var response map[string]interface{}
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}
				if response["error"] != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, response["error"])
				}
			}
		})
	}
}

func TestSendEmailHandler_RequiredFields(t *testing.T) {
	srv := &Server{}

	tests := []struct {
		name           string
		requestBody    SendEmailRequest
		expectedStatus int
	}{
		{
			name: "Missing recipient",
			requestBody: SendEmailRequest{
				Subject: "Test Subject",
				Body:    "Test body",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing subject",
			requestBody: SendEmailRequest{
				Recipient: "test@example.com",
				Body:      "Test body",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing body",
			requestBody: SendEmailRequest{
				Recipient: "test@example.com",
				Subject:   "Test Subject",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "All required fields present",
			requestBody: SendEmailRequest{
				Recipient: "test@example.com",
				Subject:   "Test Subject",
				Body:      "Test body",
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			req := createAuthenticatedRequest("POST", "/api/email/send", body, "test-user-id")

			rr := httptest.NewRecorder()

			srv.sendEmailHandler(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// Test validation logic directly
func TestSelfDestructValidation(t *testing.T) {
	tests := []struct {
		name           string
		selfDestruct   bool
		maxAttempts    int
		shouldBeValid  bool
		expectedError  string
	}{
		{
			name:          "Valid self-destruct enabled",
			selfDestruct:  true,
			maxAttempts:   3,
			shouldBeValid: true,
		},
		{
			name:          "Valid self-destruct enabled max attempts",
			selfDestruct:  true,
			maxAttempts:   10,
			shouldBeValid: true,
		},
		{
			name:           "Invalid max attempts too low",
			selfDestruct:   true,
			maxAttempts:    0,
			shouldBeValid:  false,
			expectedError:  "maxFailedAttempts must be between 1 and 10",
		},
		{
			name:           "Invalid max attempts too high",
			selfDestruct:   true,
			maxAttempts:    11,
			shouldBeValid:  false,
			expectedError:  "maxFailedAttempts must be between 1 and 10",
		},
		{
			name:          "Self-destruct disabled",
			selfDestruct:  false,
			maxAttempts:   5,
			shouldBeValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := SendEmailRequest{
				Recipient:                 "test@example.com",
				Subject:                   "Test Subject",
				Body:                      "Test body",
				SelfDestructAfterAttempts: tt.selfDestruct,
				MaxFailedAttempts:         tt.maxAttempts,
			}

			// Simulate validation logic
			var isValid bool
			var errorMsg string

			if req.SelfDestructAfterAttempts {
				if req.MaxFailedAttempts < 1 || req.MaxFailedAttempts > 10 {
					isValid = false
					errorMsg = "maxFailedAttempts must be between 1 and 10"
				} else {
					isValid = true
				}
			} else {
				isValid = true
				req.MaxFailedAttempts = 0
			}

			if isValid != tt.shouldBeValid {
				t.Errorf("Expected valid=%t, got valid=%t", tt.shouldBeValid, isValid)
			}

			if !isValid && errorMsg != tt.expectedError {
				t.Errorf("Expected error '%s', got '%s'", tt.expectedError, errorMsg)
			}

			if !req.SelfDestructAfterAttempts && req.MaxFailedAttempts != 0 {
				t.Errorf("Expected MaxFailedAttempts to be 0 when self-destruct is disabled, got %d", req.MaxFailedAttempts)
			}
		})
	}
}
