package auth

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/argon2"
)

type SignUpRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type SignUpResponse struct {
	TempID string `json:"temp_id"`
	TotpQr string `json:"totp_qr"`
}

type TempState struct {
	Email        string
	PasswordHash []byte
	TotpSecret   string
	ExpiresAt    time.Time
}

var tempStore = sync.Map{} // In-memory store for temp TOTP state

func SignUpHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SignUpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
			log.Printf("Sign-up failed: %v", err)
			return
		}

		// Validate email
		if !regexp.MustCompile(`^[^@]+@securesystem.email$`).MatchString(req.Email) {
			http.Error(w, `{"error":"Invalid email format"}`, http.StatusBadRequest)
			return
		}

		// Validate passwords
		if req.Password != req.ConfirmPassword {
			http.Error(w, `{"error":"Passwords do not match"}`, http.StatusBadRequest)
			return
		}
		if len(req.Password) < 8 || len(req.Password) > 128 {
			http.Error(w, `{"error":"Password must be 8–128 characters"}`, http.StatusBadRequest)
			return
		}

		// Check user count
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count); err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("Sign-up failed: %v", err)
			return
		}
		if count >= 100 {
			http.Error(w, `{"error":"Max 100 users reached"}`, http.StatusForbidden)
			return
		}

		// Check email uniqueness
		var exists int
		if err := db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", req.Email).Scan(&exists); err != nil || exists > 0 {
			http.Error(w, `{"error":"Email already exists"}`, http.StatusBadRequest)
			return
		}

		// Hash password
		salt := make([]byte, 16)
		if _, err := rand.Read(salt); err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("Sign-up failed: %v", err)
			return
		}
		hash := argon2.IDKey([]byte(req.Password), salt, 1, 64*1024, 4, 32)
		passwordHash := append(salt, hash...)

		// Generate TOTP secret
		key, err := totp.Generate(totp.GenerateOpts{
			Issuer:      "SecureEmailMVP",
			AccountName: req.Email,
		})
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("Sign-up failed: %v", err)
			return
		}

		// Generate QR code
		qr, err := key.Image(160, 160)
		if err != nil {
			http.Error(w, `{"error":"Internal server error"}`, http.StatusInternalServerError)
			log.Printf("Sign-up failed: %v", err)
			return
		}
		var buf bytes.Buffer
		png.Encode(&buf, qr)
		totpQr := base64.StdEncoding.EncodeToString(buf.Bytes())

		// Store temp state
		tempID := uuid.New().String()
		tempStore.Store(tempID, TempState{
			Email:        req.Email,
			PasswordHash: passwordHash,
			TotpSecret:   key.Secret(),
			ExpiresAt:    time.Now().Add(5 * time.Minute),
		})

		// Respond
		resp := SignUpResponse{TempID: tempID, TotpQr: totpQr}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			log.Printf("Sign-up response failed: %v", err)
		}
		log.Printf("Sign-up initiated for %s", req.Email)
	}
}
