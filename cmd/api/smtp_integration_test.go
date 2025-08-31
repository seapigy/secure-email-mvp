package main

import (
	"database/sql"
	"os"
	"testing"

	"secure-email-mvp/pkg/email"
	"secure-email-mvp/pkg/mailer"
)

func TestSMTPIntegration(t *testing.T) {
	// Skip if integration tests are not enabled
	if os.Getenv("INTEGRATION_TESTS") != "1" {
		t.Skip("Skipping SMTP integration test. Set INTEGRATION_TESTS=1 to run.")
	}

	// Test SMTP mailer initialization
	t.Run("SMTP Mailer Initialization", func(t *testing.T) {
		// Set test environment variables
		os.Setenv("SES_SMTP_HOST", "email-smtp.us-east-1.amazonaws.com")
		os.Setenv("SES_SMTP_PORT", "587")
		os.Setenv("SES_SMTP_USERNAME", "test-user")
		os.Setenv("SES_SMTP_PASSWORD", "test-password")

		// Initialize SMTP mailer
		smtpMailer, err := mailer.NewSMTPMailer()
		if err != nil {
			t.Fatalf("Failed to initialize SMTP mailer: %v", err)
		}

		if smtpMailer == nil {
			t.Fatal("SMTP mailer should not be nil")
		}

		// Test configuration retrieval
		config := smtpMailer.GetConfig()
		if config["host"] != "email-smtp.us-east-1.amazonaws.com" {
			t.Errorf("Expected host to be 'email-smtp.us-east-1.amazonaws.com', got: %s", config["host"])
		}
		if config["port"] != "587" {
			t.Errorf("Expected port to be '587', got: %s", config["port"])
		}

		// Verify sensitive data is not exposed
		if _, exists := config["username"]; exists {
			t.Error("Username should not be exposed in config")
		}
		if _, exists := config["password"]; exists {
			t.Error("Password should not be exposed in config")
		}
	})

	// Test EmailSecurityService with SMTP integration
	t.Run("EmailSecurityService SMTP Integration", func(t *testing.T) {
		// Create a test database
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			t.Fatalf("Failed to create test database: %v", err)
		}
		defer db.Close()

		// Set test environment variables
		os.Setenv("SES_SMTP_HOST", "email-smtp.us-east-1.amazonaws.com")
		os.Setenv("SES_SMTP_PORT", "587")
		os.Setenv("SES_SMTP_USERNAME", "test-user")
		os.Setenv("SES_SMTP_PASSWORD", "test-password")

		// Initialize EmailSecurityService
		service := email.NewEmailSecurityService(db)
		if service == nil {
			t.Fatal("EmailSecurityService should not be nil")
		}

		// Test notification email sending (this will fail due to SMTP connection, but should not crash)
		err = service.SendNotificationEmail("test@example.com", "test-email-id", "test@securesystem.email")
		if err != nil {
			// Expected to fail due to SMTP connection, but should not crash
			t.Logf("SMTP notification failed as expected: %v", err)
		}
	})
}

func TestSMTPConfigurationValidation(t *testing.T) {
	t.Run("Missing Environment Variables", func(t *testing.T) {
		// Clear environment variables
		os.Unsetenv("SES_SMTP_HOST")
		os.Unsetenv("SES_SMTP_PORT")
		os.Unsetenv("SES_SMTP_USERNAME")
		os.Unsetenv("SES_SMTP_PASSWORD")

		// Try to initialize SMTP mailer
		_, err := mailer.NewSMTPMailer()
		if err == nil {
			t.Error("Expected error when SMTP environment variables are missing")
		}
	})

	t.Run("Partial Environment Variables", func(t *testing.T) {
		// Set only some environment variables
		os.Setenv("SES_SMTP_HOST", "test.example.com")
		os.Setenv("SES_SMTP_PORT", "587")
		// Missing username, password

		// Try to initialize SMTP mailer
		_, err := mailer.NewSMTPMailer()
		if err == nil {
			t.Error("Expected error when SMTP environment variables are incomplete")
		}
	})
}
