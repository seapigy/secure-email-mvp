package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"

	"github.com/google/uuid"
)

type SendEmailRequest struct {
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	// Security settings
	SelfDestructAfterAttempts bool   `json:"selfDestructAfterAttempts,omitempty"`
	MaxFailedAttempts         int    `json:"maxFailedAttempts,omitempty"`
	BurnAfterRead             bool   `json:"burnAfterRead,omitempty"`
	ExpiresAt                 string `json:"expiresAt,omitempty"` // ISO 8601 UTC format
}

type SendEmailResponse struct {
	BlobID string `json:"blob_id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// sendEmailHandler handles POST /api/email/send. It compresses, encrypts, uploads, and stores metadata.
func (srv *Server) sendEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("sendEmailHandler started")
	var req SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("JSON decode error: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request"}`))
		return
	}
	// Get authenticated user_id from context
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	if req.Recipient == "" || req.Subject == "" || req.Body == "" {
		log.Printf("Missing required fields: recipient=%q subject=%q body=%q", req.Recipient, req.Subject, req.Body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing required fields"}`))
		return
	}

	// Validate recipient email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(req.Recipient) {
		log.Printf("Invalid recipient email format: %q", req.Recipient)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid recipient email format"}`))
		return
	}

	// Validate self-destruct settings
	if req.SelfDestructAfterAttempts {
		if req.MaxFailedAttempts < 1 || req.MaxFailedAttempts > 10 {
			log.Printf("Invalid maxFailedAttempts: %d (must be between 1-10)", req.MaxFailedAttempts)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"maxFailedAttempts must be between 1 and 10"}`))
			return
		}
	} else {
		// If self-destruct is disabled, set maxFailedAttempts to 0
		req.MaxFailedAttempts = 0
	}

	// Validate expiration timestamp if provided
	var expiresAtValue interface{} = nil
	if req.ExpiresAt != "" {
		// Parse the ISO 8601 UTC timestamp
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			log.Printf("Invalid expiresAt format: %q (expected ISO 8601 UTC format)", req.ExpiresAt)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"expiresAt must be in ISO 8601 UTC format (e.g., 2024-01-15T14:30:00Z)"}`))
			return
		}

		// Check that expiration is in the future
		if expiresAt.Before(time.Now()) {
			log.Printf("Expiration time is in the past: %v", expiresAt)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error":"expiresAt must be in the future"}`))
			return
		}

		expiresAtValue = expiresAt
	}

	// 1. Compress (gzip)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(req.Body)); err != nil {
		log.Printf("Gzip compression failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Compression failed"}`))
		return
	}
	gz.Close()
	compressed := buf.Bytes()

	// 2. Encrypt compressed content using AES-256-GCM
	encryptedData, err := auth.EncryptAES256GCM(compressed)
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Encryption failed"}`))
		return
	}

	// Combine ciphertext and auth tag for storage (nonce stored separately)
	encrypted := append(encryptedData.Ciphertext, encryptedData.AuthTag...)

	// 4. Generate blobID
	blobID := uuid.New().String() + ".blob"

	// 5. Upload to R2 with context and timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := storage.UploadToR2WithContext(ctx, blobID, encrypted); err != nil {
		log.Printf("R2 upload failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"R2 upload failed"}`))
		return
	}
	log.Printf("R2 upload succeeded: %s", blobID)

	// 6. Compute SHA-256 hash of the encrypted content
	// hash := sha256.Sum256(encrypted)
	// hashB64 := base64.StdEncoding.EncodeToString(hash[:])

	// 7. Store metadata in SQLite
	emailID := uuid.New().String()
	encryptedKeyB64 := base64.StdEncoding.EncodeToString(encryptedData.Key)

	if srv.db == nil {
		log.Printf("ERROR: srv.db is nil!")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database connection is nil"}`))
		return
	}

	hash := sha256.Sum256(encrypted)
	hashB64 := base64.StdEncoding.EncodeToString(hash[:])

	// --- SIMULATED DB FAILURE BLOCK ---
	if os.Getenv("SIMULATE_DB_FAILURE") == "1" {
		log.Printf("Simulated DB insert failure triggered by SIMULATE_DB_FAILURE env var")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database insert failed (simulated)"}`))
		return
	}
	// --- END SIMULATED DB FAILURE BLOCK ---

	// Store nonce and auth tag separately for decryption
	nonceB64 := base64.StdEncoding.EncodeToString(encryptedData.Nonce)
	authTagB64 := base64.StdEncoding.EncodeToString(encryptedData.AuthTag)

	// Convert boolean to integer for SQLite storage
	selfDestructInt := 0
	if req.SelfDestructAfterAttempts {
		selfDestructInt = 1
	}

	burnAfterReadInt := 0
	if req.BurnAfterRead {
		burnAfterReadInt = 1
	}

	_, err = srv.db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, 
			self_destruct_after_attempts, max_attempts, created_at, burn_after_read, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, userID, req.Recipient, req.Subject, blobID, encryptedKeyB64,
		nonceB64, authTagB64, "gzip", hashB64, selfDestructInt, req.MaxFailedAttempts, time.Now(), burnAfterReadInt, expiresAtValue,
	)
	if err != nil {
		log.Printf("DB insert failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Database insert failed"}`))
		return
	}

	// 8. Respond
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SendEmailResponse{
		BlobID: blobID,
		Status: "success",
	})
}
