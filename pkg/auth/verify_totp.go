package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
)

type VerifyTotpRequest struct {
	TempID   string `json:"temp_id"`
	TotpCode string `json:"totp_code"`
}

type VerifyTotpResponse struct {
	Token string `json:"token"`
}

func VerifyTotpHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req VerifyTotpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
			log.Printf("TOTP verification failed: %v", err)
			return
		}

		// Retrieve temp state
		data, ok := tempStore.Load(req.TempID)
		if !ok {
			http.Error(w, `{"error":"Invalid or expired temp ID"}`, http.StatusBadRequest)
			return
		}
		state := data.(TempState)
		if time.Now().After(state.ExpiresAt) {
			tempStore.Delete(req.TempID)
			http.Error(w, `{"error":"Temp ID expired"}`, http.StatusBadRequest)
			return
		}

		// Validate TOTP
		if !totp.Validate(req.TotpCode, state.TotpSecret) {
			http.Error(w, `{"error":"Invalid TOTP code"}`, http.StatusBadRequest)
			return
		}

		// Create user
		userID := uuid.New().String()
		_, err := db.Exec(
			"INSERT INTO users (user_id, email, password_hash, totp_secret) VALUES (?, ?, ?, ?)",
			userID, state.Email, state.PasswordHash, state.TotpSecret,
		)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("User creation failed: %v", err)
			return
		}

		// Generate JWT
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": userID,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("JWT generation failed: %v", err)
			return
		}

		// Clean up temp state
		tempStore.Delete(req.TempID)

		// Respond
		resp := VerifyTotpResponse{Token: tokenString}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("TOTP verification response failed: %v", err)
		}
		log.Printf("User created for %s", state.Email)
	}
}
