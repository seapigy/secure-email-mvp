package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestSignupHandlerV2(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply the migration to create the required table structure
	migrationSQL := `
		CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			plan TEXT DEFAULT 'free',
			company_code TEXT,
			status TEXT DEFAULT 'pending_verification',
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`
	_, err = db.Exec(migrationSQL)
	if err != nil {
		t.Fatalf("Failed to apply migration: %v", err)
	}

	// Create the handler
	handler := signupHandlerV2Factory(db)

	tests := []struct {
		name           string
		requestBody    SignupRequestV2
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Valid Free Plan Signup",
			requestBody: SignupRequestV2{
				Plan:     "free",
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Valid Paid Plan Signup",
			requestBody: SignupRequestV2{
				Plan:     "paid",
				Email:    "paid@example.com",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Valid Company Plan Signup",
			requestBody: SignupRequestV2{
				Plan:        "company",
				Email:       "company@example.com",
				Password:    "SecurePass123!",
				CompanyCode: "COMP123",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Company Plan Without Company Code",
			requestBody: SignupRequestV2{
				Plan:     "company",
				Email:    "company@example.com",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Company code required for company plans",
		},
		{
			name: "Invalid Plan",
			requestBody: SignupRequestV2{
				Plan:     "invalid",
				Email:    "test@example.com",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid plan type",
		},
		{
			name: "Invalid Email Format",
			requestBody: SignupRequestV2{
				Plan:     "free",
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid email format",
		},
		{
			name: "Weak Password",
			requestBody: SignupRequestV2{
				Plan:     "free",
				Email:    "test@example.com",
				Password: "weak",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Password does not meet security requirements",
		},
		{
			name: "Missing Required Fields",
			requestBody: SignupRequestV2{
				Plan: "free",
				// Missing email and password
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Missing required fields",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request body
			body, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/api/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.ServeHTTP(rr, req)

			// Check status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response body for errors
			if tt.expectedError != "" {
				var errorResponse map[string]string
				if err := json.Unmarshal(rr.Body.Bytes(), &errorResponse); err != nil {
					t.Fatalf("Failed to unmarshal error response: %v", err)
				}
				if errorResponse["error"] != tt.expectedError {
					t.Errorf("Expected error '%s', got '%s'", tt.expectedError, errorResponse["error"])
				}
			} else {
				// Check success response
				var response SignupResponseV2
				if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
					t.Fatalf("Failed to unmarshal success response: %v", err)
				}
				if response.Status != "success" {
					t.Errorf("Expected status 'success', got '%s'", response.Status)
				}
				if response.NextStep != "verify_email" {
					t.Errorf("Expected next_step 'verify_email', got '%s'", response.NextStep)
				}
				if response.UserID == "" {
					t.Error("Expected non-empty user_id")
				}
			}
		})
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "SecurePass123!"

	// Test password hashing
	hashedPassword, err := hashPasswordWithArgon2id(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Verify hash format
	if !strings.HasPrefix(hashedPassword, "$argon2id$") {
		t.Error("Expected Argon2id hash format")
	}

	// Verify hash is different from original password
	if hashedPassword == password {
		t.Error("Hash should not equal original password")
	}
}

func TestPasswordValidation(t *testing.T) {
	tests := []struct {
		password string
		expected bool
	}{
		{"SecurePass123!", true},   // Valid password
		{"weak", false},            // Too short
		{"nouppercase123!", false}, // No uppercase
		{"NOLOWERCASE123!", false}, // No lowercase
		{"NoNumbers!", false},      // No numbers
		{"NoSpecial123", false},    // No special characters
		{"", false},                // Empty
	}

	for _, tt := range tests {
		result := isValidPasswordStrength(tt.password)
		if result != tt.expected {
			t.Errorf("Password '%s': expected %v, got %v", tt.password, tt.expected, result)
		}
	}
}
