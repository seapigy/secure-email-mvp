package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"
)

// =============================================================================
// SES CONFIGURATION TEST
// =============================================================================
// This test sends a test email to cpigusch@gmail.com to verify Amazon SES
// configuration, domain verification, and email authentication (SPF, DKIM, DMARC)
// =============================================================================

// TestSESConfiguration sends a test email to verify SES setup
func TestSESConfiguration(t *testing.T) {
	// Skip this test if not running integration tests
	if testing.Short() {
		t.Skip("Skipping SES configuration test in short mode")
	}

	// Test configuration
	targetEmail := "cpigusch@gmail.com"
	testEmail := "test@example.com"
	testPassword := "testpass123"

	// Step 1: Test API health
	t.Run("API Health Check", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8080/health")
		if err != nil {
			t.Fatalf("Failed to check API health: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var healthResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&healthResponse); err != nil {
			t.Fatalf("Failed to decode health response: %v", err)
		}

		if status, ok := healthResponse["status"].(string); !ok || status != "ok" {
			t.Fatalf("Expected status 'ok', got %v", status)
		}

		t.Log("✅ API health check passed")
	})

	// Step 2: Login to get authentication token
	var authToken string
	t.Run("Authentication", func(t *testing.T) {
		loginData := map[string]interface{}{
			"email":     testEmail,
			"password":  testPassword,
			"totp_code": "000000", // Default TOTP for testing
		}

		loginJSON, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal login data: %v", err)
		}

		resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(loginJSON))
		if err != nil {
			t.Fatalf("Failed to login: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var loginResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		if token, ok := loginResponse["access_token"].(string); ok && token != "" {
			authToken = token
			t.Log("✅ Authentication successful")
		} else {
			t.Fatalf("No access token received in login response")
		}
	})

	// Step 3: Send SES test email
	t.Run("Send SES Test Email", func(t *testing.T) {
		if authToken == "" {
			t.Skip("Skipping email test - no auth token available")
		}

		emailSubject := "Amazon SES Test Email - Secure Email MVP"
		emailBody := fmt.Sprintf(`This is a test email sent through Amazon SES.

This email is being sent to verify that our Amazon SES configuration is working correctly for the Secure Email MVP project.

Configuration Details:
Domain: securesystem.email
Region: us-east-1
Purpose: Email authentication testing (SPF, DKIM, DMARC)

Expected Results:
If you receive this email, it means:

✅ Amazon SES is properly configured
✅ Domain verification is complete
✅ Email authentication should be working

Next Steps:
Please check the email headers in Gmail to confirm:

SPF = PASS  
DKIM = PASS  
DMARC = PASS  

Test Information:
- Sent at: %s
- From: securesystem.email
- To: %s
- Purpose: SES Configuration Verification

If you see this email, our Amazon SES setup is working correctly!`, time.Now().Format("2006-01-02 15:04:05 UTC"), targetEmail)

		emailData := map[string]interface{}{
			"recipient":                targetEmail,
			"subject":                  emailSubject,
			"body":                     emailBody,
			"selfDestructAfterAttempts": false,
			"burnAfterRead":            false,
			"requireMFA":               false,
			"password":                 "",
			"timeLock":                 false,
			"expiresAt":                time.Now().AddDate(0, 0, 30).Format("2006-01-02T15:04:05Z"),
			"remoteRevoke":             false,
			"stripMetadata":            false,
			"tamperAlerts":             false,
		}

		emailJSON, err := json.Marshal(emailData)
		if err != nil {
			t.Fatalf("Failed to marshal email data: %v", err)
		}

		req, err := http.NewRequest("POST", "http://localhost:8080/api/email/send", bytes.NewBuffer(emailJSON))
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		req.Header.Set("Authorization", "Bearer "+authToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send email: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", resp.StatusCode)
		}

		var emailResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&emailResponse); err != nil {
			t.Fatalf("Failed to decode email response: %v", err)
		}

		if emailID, ok := emailResponse["email_id"].(string); ok && emailID != "" {
			t.Logf("✅ Email sent successfully - Email ID: %s", emailID)
		} else {
			t.Fatalf("No email ID received in response")
		}

		if secureLinkURL, ok := emailResponse["secure_link_url"].(string); ok && secureLinkURL != "" {
			t.Logf("✅ Secure link URL generated: %s", secureLinkURL)
		}

		t.Log("✅ SES test email sent successfully")
	})
}

// TestSESConfigurationManual is a manual test that can be run with go test -run TestSESConfigurationManual
func TestSESConfigurationManual(t *testing.T) {
	t.Log("🔧 Amazon SES Configuration Test")
	t.Log("================================================")
	t.Log("This test sends a test email to cpigusch@gmail.com")
	t.Log("To run this test manually:")
	t.Log("  go test -run TestSESConfigurationManual -v")
	t.Log("")
	t.Log("Make sure the backend server is running on localhost:8080")
	t.Log("")

	// Run the actual test
	TestSESConfiguration(t)

	t.Log("")
	t.Log("📧 Next Steps:")
	t.Log("1. Check the email inbox of cpigusch@gmail.com")
	t.Log("2. Verify the test email was received")
	t.Log("3. Check email headers in Gmail for authentication results:")
	t.Log("   - SPF = PASS")
	t.Log("   - DKIM = PASS")
	t.Log("   - DMARC = PASS")
	t.Log("4. If all headers show PASS, SES configuration is working correctly")
	t.Log("")
	t.Log("🔍 To check email headers in Gmail:")
	t.Log("1. Open the email in Gmail")
	t.Log("2. Click the three dots (⋮) next to Reply")
	t.Log("3. Select 'Show original'")
	t.Log("4. Look for SPF, DKIM, and DMARC results in the headers")
	t.Log("")
	t.Log("🏁 Test completed!")
}
