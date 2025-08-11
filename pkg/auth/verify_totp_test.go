package auth

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
	_ "modernc.org/sqlite"
)

func TestVerifyTotpHandler(t *testing.T) {
	db, _ := sql.Open("sqlite", ":memory:")
	db.Exec("CREATE TABLE users (user_id TEXT, email TEXT UNIQUE, password_hash BLOB, totp_secret TEXT)")
	os.Setenv("JWT_SECRET", "test-secret")
	handler := VerifyTotpHandler(db)

	// Setup temp state
	tempID := "test-uuid"
	secret := "JBSWY3DPEHPK3PXP"
	totpCode, _ := totp.GenerateCode(secret, time.Now())

	tests := []struct {
		name           string
		body           string
		status         int
		errorMsg       string
		setupTempState bool
	}{
		{
			name:           "Valid TOTP",
			body:           `{"temp_id":"test-uuid","totp_code":"` + totpCode + `"}`,
			status:         http.StatusOK,
			setupTempState: true,
		},
		{
			name:           "Invalid TOTP",
			body:           `{"temp_id":"test-uuid","totp_code":"123456"}`,
			status:         http.StatusBadRequest,
			errorMsg:       "Invalid TOTP code",
			setupTempState: true,
		},
		{
			name:           "Invalid temp ID",
			body:           `{"temp_id":"wrong-uuid","totp_code":"123456"}`,
			status:         http.StatusBadRequest,
			errorMsg:       "Invalid or expired temp ID",
			setupTempState: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup temp state for each test that needs it
			if tt.setupTempState {
				tempStore.Store(tempID, TempState{
					Email:        "test@securesystem.email",
					PasswordHash: []byte("hashed"),
					TotpSecret:   secret,
					ExpiresAt:    time.Now().Add(5 * time.Minute),
				})
			}

			req, _ := http.NewRequest("POST", "/api/auth/verify-totp", bytes.NewBufferString(tt.body))
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
