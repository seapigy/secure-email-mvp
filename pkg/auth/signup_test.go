package auth

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSignUpHandler(t *testing.T) {
	db, _ := sql.Open("sqlite3", ":memory:")
	db.Exec("CREATE TABLE users (user_id TEXT, email TEXT UNIQUE, password_hash BLOB, totp_secret TEXT)")
	handler := SignUpHandler(db)

	tests := []struct {
		name     string
		body     string
		status   int
		errorMsg string
	}{
		{
			name:   "Valid sign-up",
			body:   `{"email":"test@securesystem.email","password":"password123","confirm_password":"password123"}`,
			status: http.StatusOK,
		},
		{
			name:     "Invalid email",
			body:     `{"email":"test@example.com","password":"password123","confirm_password":"password123"}`,
			status:   http.StatusBadRequest,
			errorMsg: "Invalid email format",
		},
		{
			name:     "Password mismatch",
			body:     `{"email":"test2@securesystem.email","password":"password123","confirm_password":"different"}`,
			status:   http.StatusBadRequest,
			errorMsg: "Passwords do not match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/auth/signup", bytes.NewBufferString(tt.body))
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != tt.status {
				t.Errorf("Expected status %d, got %d", tt.status, rr.Code)
			}
			if tt.errorMsg != "" && !strings.Contains(rr.Body.String(), tt.errorMsg) {
				t.Errorf("Expected error %s, got %s", tt.errorMsg, rr.Body.String())
			}
		})
	}
}
