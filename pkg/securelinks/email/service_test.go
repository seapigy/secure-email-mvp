package email

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/email"
)

// MockSESHandler is a mock implementation of the SES handler for testing
type MockSESHandler struct {
	shouldFail  bool
	transaction *email.SESTransaction
}

func (m *MockSESHandler) SendEmailViaSES(ctx context.Context, emailID, senderID, recipient, subject, body string) (*email.SESTransaction, error) {
	if m.shouldFail {
		return nil, fmt.Errorf("mock send failure")
	}

	return &email.SESTransaction{
		TransactionID: "mock_transaction_123",
		MessageID:     emailID,
		SenderID:      senderID,
		Recipient:     recipient,
		BlobID:        emailID,
		Timestamp:     time.Now(),
		Status:        "sent",
		RetryCount:    0,
	}, nil
}

func TestSendSecureLinkEmail(t *testing.T) {
	// Create in-memory database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Create test tables
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS link_audit_log (
			id TEXT PRIMARY KEY,
			link_id TEXT NOT NULL,
			event_type TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			ip_address TEXT,
			user_agent TEXT,
			details TEXT,
			ses_transaction_id TEXT,
			recipient_email TEXT
		);
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	tests := []struct {
		name        string
		request     SecureLinkEmailRequest
		shouldFail  bool
		expectError bool
	}{
		{
			name: "Valid secure link email request",
			request: SecureLinkEmailRequest{
				LinkID:         "test_link_123",
				RecipientEmail: "test@example.com",
				SenderName:     "Test Sender",
				SenderEmail:    "sender@example.com",
				SecurityContext: SecurityContext{
					RequirePassword: true,
					RequireMFA:      false,
					ReadOnce:        true,
					AutoDestruct:    false,
				},
			},
			shouldFail:  false,
			expectError: false,
		},
		{
			name: "Invalid recipient email",
			request: SecureLinkEmailRequest{
				LinkID:         "test_link_123",
				RecipientEmail: "invalid-email",
				SenderEmail:    "sender@example.com",
			},
			shouldFail:  false,
			expectError: true,
		},
		{
			name: "Missing link ID",
			request: SecureLinkEmailRequest{
				RecipientEmail: "test@example.com",
				SenderEmail:    "sender@example.com",
			},
			shouldFail:  false,
			expectError: true,
		},
		{
			name: "SES send failure",
			request: SecureLinkEmailRequest{
				LinkID:         "test_link_123",
				RecipientEmail: "test@example.com",
				SenderEmail:    "sender@example.com",
			},
			shouldFail:  true,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock SES handler
			mockSES := &MockSESHandler{
				shouldFail: tt.shouldFail,
			}

			// Create email service
			service := NewService(db, mockSES, "https://test.example.com")

			// Send email
			response, err := service.SendSecureLinkEmail(context.Background(), tt.request)

			// Check results
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if response != nil && response.Success {
					t.Errorf("Expected failure response but got success")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if response == nil {
					t.Errorf("Expected response but got nil")
				} else if !response.Success {
					t.Errorf("Expected success response but got failure: %s", response.Error)
				}
				if response.TransactionID == "" {
					t.Errorf("Expected transaction ID but got empty")
				}
			}
		})
	}
}

func TestGenerateEmailContent(t *testing.T) {
	service := &Service{
		baseURL: "https://test.example.com",
	}

	req := SecureLinkEmailRequest{
		LinkID:         "test_link_123",
		RecipientEmail: "test@example.com",
		SenderName:     "Test Sender",
		SenderEmail:    "sender@example.com",
		SecurityContext: SecurityContext{
			RequirePassword: true,
			RequireMFA:      true,
			MFAType:         "TOTP",
			ReadOnce:        true,
			AutoDestruct:    false,
		},
	}

	subject, body, err := service.generateEmailContent(req)
	if err != nil {
		t.Fatalf("Failed to generate email content: %v", err)
	}

	// Check subject
	if subject == "" {
		t.Errorf("Expected non-empty subject")
	}
	if subject != "Secure Message from Test Sender" {
		t.Errorf("Expected subject 'Secure Message from Test Sender', got '%s'", subject)
	}

	// Check body
	if body == "" {
		t.Errorf("Expected non-empty body")
	}
	if !strings.Contains(body, "test_link_123") {
		t.Errorf("Expected body to contain link ID")
	}
	if !strings.Contains(body, "https://test.example.com/v/test_link_123") {
		t.Errorf("Expected body to contain secure URL")
	}
	if !strings.Contains(body, "password protection") {
		t.Errorf("Expected body to contain security features description")
	}
}

func TestBuildSecurityFeaturesDescription(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name     string
		context  SecurityContext
		expected string
	}{
		{
			name: "Password only",
			context: SecurityContext{
				RequirePassword: true,
			},
			expected: "password protection",
		},
		{
			name: "Password and MFA",
			context: SecurityContext{
				RequirePassword: true,
				RequireMFA:      true,
				MFAType:         "TOTP",
			},
			expected: "password protection and totp verification",
		},
		{
			name: "Multiple features",
			context: SecurityContext{
				RequirePassword: true,
				RequireMFA:      true,
				MFAType:         "TOTP",
				ReadOnce:        true,
				AutoDestruct:    true,
			},
			expected: "password protection, totp verification, one-time viewing, and auto-destruct protection",
		},
		{
			name:     "No features",
			context:  SecurityContext{},
			expected: "standard security measures",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.buildSecurityFeaturesDescription(tt.context)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestValidateRequest(t *testing.T) {
	service := &Service{}

	tests := []struct {
		name        string
		request     SecureLinkEmailRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: SecureLinkEmailRequest{
				LinkID:         "test_link_123",
				RecipientEmail: "test@example.com",
				SenderEmail:    "sender@example.com",
			},
			expectError: false,
		},
		{
			name: "Missing link ID",
			request: SecureLinkEmailRequest{
				RecipientEmail: "test@example.com",
				SenderEmail:    "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "Missing recipient email",
			request: SecureLinkEmailRequest{
				LinkID:      "test_link_123",
				SenderEmail: "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "Invalid recipient email",
			request: SecureLinkEmailRequest{
				LinkID:         "test_link_123",
				RecipientEmail: "invalid-email",
				SenderEmail:    "sender@example.com",
			},
			expectError: true,
		},
		{
			name: "Missing sender email",
			request: SecureLinkEmailRequest{
				LinkID:         "test_link_123",
				RecipientEmail: "test@example.com",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.validateRequest(tt.request)
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}
