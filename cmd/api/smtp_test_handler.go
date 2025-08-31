package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"secure-email-mvp/pkg/mailer"
)

// SMTPTestRequest represents the request body for SMTP testing
type SMTPTestRequest struct {
	TestEmail string `json:"test_email"` // Email address to send test email to
}

// SMTPTestResponse represents the response for SMTP testing
type SMTPTestResponse struct {
	Status  string            `json:"status"`          // "success" or "error"
	Message string            `json:"message"`         // Human-readable message
	Config  map[string]string `json:"config"`          // SMTP configuration (without sensitive data)
	Error   string            `json:"error,omitempty"` // Error details if status is "error"
}

// smtpTestHandler handles SMTP connectivity testing
func (srv *Server) smtpTestHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req SMTPTestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("❌ Failed to parse SMTP test request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate test email
	if req.TestEmail == "" {
		log.Printf("❌ Test email address is required")
		http.Error(w, "Test email address is required", http.StatusBadRequest)
		return
	}

	// Basic email validation
	if !isValidEmailAddress(req.TestEmail) {
		log.Printf("❌ Invalid email address: %s", req.TestEmail)
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	log.Printf("🧪 Testing SMTP connectivity to: %s", req.TestEmail)

	// Initialize SMTP mailer
	smtpMailer, err := mailer.NewSMTPMailer()
	if err != nil {
		log.Printf("❌ Failed to initialize SMTP mailer: %v", err)

		// Return detailed error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SMTPTestResponse{
			Status:  "error",
			Message: "SMTP configuration error",
			Error:   err.Error(),
			Config:  map[string]string{},
		})
		return
	}

	// Test SMTP connection by sending a test email
	if err := smtpMailer.TestConnection(req.TestEmail); err != nil {
		log.Printf("❌ SMTP test failed: %v", err)

		// Return detailed error response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(SMTPTestResponse{
			Status:  "error",
			Message: "SMTP test failed",
			Error:   err.Error(),
			Config:  smtpMailer.GetConfig(),
		})
		return
	}

	// Success response
	log.Printf("✅ SMTP test successful - test email sent to: %s", req.TestEmail)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SMTPTestResponse{
		Status:  "success",
		Message: fmt.Sprintf("SMTP test successful - test email sent to %s", req.TestEmail),
		Config:  smtpMailer.GetConfig(),
	})
}

// isValidEmailAddress performs basic email validation
func isValidEmailAddress(email string) bool {
	// Basic email validation
	if !strings.Contains(email, "@") {
		return false
	}

	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}

	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return false
	}

	if !strings.Contains(parts[1], ".") {
		return false
	}

	return true
}

// smtpConfigHandler returns SMTP configuration status (without sensitive data)
func (srv *Server) smtpConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	log.Printf("📋 Checking SMTP configuration status")

	// Check if SMTP environment variables are set
	config := map[string]string{}
	missingVars := []string{}

	requiredVars := []string{"SES_SMTP_HOST", "SES_SMTP_PORT", "SES_SMTP_USERNAME", "SES_SMTP_PASSWORD"}

	for _, varName := range requiredVars {
		value := os.Getenv(varName)
		if value == "" {
			missingVars = append(missingVars, varName)
		} else {
			// Don't expose sensitive data
			if varName == "SES_SMTP_PASSWORD" {
				config[varName] = "***HIDDEN***"
			} else if varName == "SES_SMTP_USERNAME" {
				config[varName] = "***HIDDEN***"
			} else {
				config[varName] = value
			}
		}
	}

	// Determine status
	var status, message string
	if len(missingVars) > 0 {
		status = "error"
		message = fmt.Sprintf("Missing SMTP configuration variables: %s", strings.Join(missingVars, ", "))
	} else {
		status = "configured"
		message = "SMTP configuration appears to be complete"
	}

	log.Printf("📋 SMTP configuration status: %s", status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  status,
		"message": message,
		"config":  config,
		"missing": missingVars,
	})
}
