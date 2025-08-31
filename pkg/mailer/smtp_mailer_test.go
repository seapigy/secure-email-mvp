package mailer

import (
	"os"
	"strings"
	"testing"
)

func TestNewSMTPMailer(t *testing.T) {
	// Test with missing environment variables
	t.Run("Missing SMTP_HOST", func(t *testing.T) {
		// Clear environment
		os.Unsetenv("SES_SMTP_HOST")
		os.Unsetenv("SES_SMTP_PORT")
		os.Unsetenv("SES_SMTP_USERNAME")
		os.Unsetenv("SES_SMTP_PASSWORD")

		_, err := NewSMTPMailer()
		if err == nil {
			t.Error("Expected error when SES_SMTP_HOST is missing")
		}
		if err.Error() != "SES_SMTP_HOST environment variable is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Missing SES_SMTP_PORT", func(t *testing.T) {
		// Set required env vars
		os.Setenv("SES_SMTP_HOST", "test.example.com")
		os.Unsetenv("SES_SMTP_PORT")

		_, err := NewSMTPMailer()
		if err == nil {
			t.Error("Expected error when SES_SMTP_PORT is missing")
		}
		if err.Error() != "SES_SMTP_PORT environment variable is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Missing SES_SMTP_USERNAME", func(t *testing.T) {
		// Set required env vars
		os.Setenv("SES_SMTP_HOST", "test.example.com")
		os.Setenv("SES_SMTP_PORT", "587")
		os.Unsetenv("SES_SMTP_USERNAME")

		_, err := NewSMTPMailer()
		if err == nil {
			t.Error("Expected error when SES_SMTP_USERNAME is missing")
		}
		if err.Error() != "SES_SMTP_USERNAME environment variable is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Missing SES_SMTP_PASSWORD", func(t *testing.T) {
		// Set required env vars
		os.Setenv("SES_SMTP_HOST", "test.example.com")
		os.Setenv("SES_SMTP_PORT", "587")
		os.Setenv("SES_SMTP_USERNAME", "testuser")
		os.Unsetenv("SES_SMTP_PASSWORD")

		_, err := NewSMTPMailer()
		if err == nil {
			t.Error("Expected error when SES_SMTP_PASSWORD is missing")
		}
		if err.Error() != "SES_SMTP_PASSWORD environment variable is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Valid Configuration", func(t *testing.T) {
		// Set all required env vars
		os.Setenv("SES_SMTP_HOST", "test.example.com")
		os.Setenv("SES_SMTP_PORT", "587")
		os.Setenv("SES_SMTP_USERNAME", "testuser")
		os.Setenv("SES_SMTP_PASSWORD", "testpass")

		mailer, err := NewSMTPMailer()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if mailer == nil {
			t.Error("Expected mailer to be created")
		}

		// Test GetConfig
		config := mailer.GetConfig()
		if config["host"] != "test.example.com" {
			t.Errorf("Expected host to be 'test.example.com', got: %s", config["host"])
		}
		if config["port"] != "587" {
			t.Errorf("Expected port to be '587', got: %s", config["port"])
		}
		// Ensure sensitive data is not included
		if _, exists := config["username"]; exists {
			t.Error("Username should not be included in config")
		}
		if _, exists := config["password"]; exists {
			t.Error("Password should not be included in config")
		}
	})
}

func TestValidateFromAddress(t *testing.T) {
	// Set up a valid mailer for testing
	os.Setenv("SES_SMTP_HOST", "test.example.com")
	os.Setenv("SES_SMTP_PORT", "587")
	os.Setenv("SES_SMTP_USERNAME", "testuser")
	os.Setenv("SES_SMTP_PASSWORD", "testpass")

	mailer, err := NewSMTPMailer()
	if err != nil {
		t.Fatalf("Failed to create mailer: %v", err)
	}

	t.Run("Valid From Address", func(t *testing.T) {
		validAddresses := []string{
			"alice@securesystem.email",
			"bob@securesystem.email",
			"test.user@securesystem.email",
			"user123@securesystem.email",
		}

		for _, addr := range validAddresses {
			err := mailer.ValidateFromAddress(addr)
			if err != nil {
				t.Errorf("Expected valid address %s to pass validation, got error: %v", addr, err)
			}
		}
	})

	t.Run("Invalid From Address", func(t *testing.T) {
		testCases := []struct {
			address string
			expect  string
		}{
			{"", "from address is required"},
			{"alice@example.com", "from address must end with @securesystem.email"},
			{"invalid-email", "from address must end with @securesystem.email"},
			{"@securesystem.email", "local part of email address cannot be empty"},
			{"alice@securesystem.com", "from address must end with @securesystem.email"},
		}

		for _, tc := range testCases {
			err := mailer.ValidateFromAddress(tc.address)
			if err == nil {
				t.Errorf("Expected invalid address %s to fail validation", tc.address)
			} else if !strings.Contains(err.Error(), tc.expect) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.expect, err)
			}
		}
	})
}

func TestEmailMessageValidation(t *testing.T) {
	// Set up a valid mailer for testing
	os.Setenv("SES_SMTP_HOST", "test.example.com")
	os.Setenv("SES_SMTP_PORT", "587")
	os.Setenv("SES_SMTP_USERNAME", "testuser")
	os.Setenv("SES_SMTP_PASSWORD", "testpass")

	mailer, err := NewSMTPMailer()
	if err != nil {
		t.Fatalf("Failed to create mailer: %v", err)
	}

	t.Run("Missing To Address", func(t *testing.T) {
		msg := EmailMessage{
			Subject: "Test Subject",
			Body:    "Test Body",
			From:    "test@securesystem.email",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when To address is missing")
		}
		if err.Error() != "recipient email address is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Missing Subject", func(t *testing.T) {
		msg := EmailMessage{
			To:   "test@example.com",
			Body: "Test Body",
			From: "test@securesystem.email",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when Subject is missing")
		}
		if err.Error() != "email subject is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Missing Body", func(t *testing.T) {
		msg := EmailMessage{
			To:      "test@example.com",
			Subject: "Test Subject",
			From:    "test@securesystem.email",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when Body is missing")
		}
		if err.Error() != "email body is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Missing From Address", func(t *testing.T) {
		msg := EmailMessage{
			To:      "test@example.com",
			Subject: "Test Subject",
			Body:    "Test Body",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when From address is missing")
		}
		if err.Error() != "invalid from address: from address is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Invalid From Address Domain", func(t *testing.T) {
		msg := EmailMessage{
			To:      "test@example.com",
			Subject: "Test Subject",
			Body:    "Test Body",
			From:    "test@example.com",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when From address has invalid domain")
		}
		if err.Error() != "invalid from address: from address must end with @securesystem.email, got: test@example.com" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Invalid From Address Format", func(t *testing.T) {
		msg := EmailMessage{
			To:      "test@example.com",
			Subject: "Test Subject",
			Body:    "Test Body",
			From:    "invalid-email",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when From address has invalid format")
		}
		if err.Error() != "invalid from address: from address must end with @securesystem.email, got: invalid-email" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Empty Local Part", func(t *testing.T) {
		msg := EmailMessage{
			To:      "test@example.com",
			Subject: "Test Subject",
			Body:    "Test Body",
			From:    "@securesystem.email",
		}

		err := mailer.SendEmail(msg)
		if err == nil {
			t.Error("Expected error when From address has empty local part")
		}
		if err.Error() != "invalid from address: local part of email address cannot be empty" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Valid Message", func(t *testing.T) {
		msg := EmailMessage{
			To:      "test@example.com",
			Subject: "Test Subject",
			Body:    "Test Body",
			From:    "test@securesystem.email",
		}

		// This will fail due to SMTP connection, but should pass validation
		err := mailer.SendEmail(msg)
		if err != nil {
			// Should fail with SMTP connection error, not validation error
			if err.Error() == "recipient email address is required" ||
				err.Error() == "email subject is required" ||
				err.Error() == "email body is required" ||
				err.Error() == "invalid from address:" {
				t.Errorf("Unexpected validation error: %v", err)
			}
		}
	})
}

func TestTestConnection(t *testing.T) {
	// Set up a valid mailer for testing
	os.Setenv("SES_SMTP_HOST", "test.example.com")
	os.Setenv("SES_SMTP_PORT", "587")
	os.Setenv("SES_SMTP_USERNAME", "testuser")
	os.Setenv("SES_SMTP_PASSWORD", "testpass")

	mailer, err := NewSMTPMailer()
	if err != nil {
		t.Fatalf("Failed to create mailer: %v", err)
	}

	t.Run("Empty Test Email", func(t *testing.T) {
		err := mailer.TestConnection("")
		if err == nil {
			t.Error("Expected error when test email is empty")
		}
		if err.Error() != "test email address is required" {
			t.Errorf("Expected specific error message, got: %v", err)
		}
	})

	t.Run("Valid Test Email", func(t *testing.T) {
		// This will fail due to SMTP connection, but should pass validation
		err := mailer.TestConnection("test@example.com")
		if err != nil {
			// Should fail with SMTP connection error, not validation error
			if err.Error() == "test email address is required" {
				t.Errorf("Unexpected validation error: %v", err)
			}
		}
	})
}
