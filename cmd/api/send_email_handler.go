package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	"secure-email-mvp/pkg"

	"github.com/google/uuid"
)

type SendEmailRequest struct {
	SenderID  string `json:"sender_id"`
	Recipient string `json:"recipient"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
}

type SendEmailResponse struct {
	BlobID string `json:"blob_id,omitempty"`
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Helper: AES-256-GCM encrypt
func encryptAESGCM(key, plaintext []byte) (ciphertext, nonce []byte, err error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, err
	}
	nonce = make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, err
	}
	ciphertext = aesgcm.Seal(nil, nonce, plaintext, nil)
	return ciphertext, nonce, nil
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
	if req.SenderID == "" || req.Recipient == "" || req.Subject == "" || req.Body == "" {
		log.Printf("Missing required fields: sender_id=%q recipient=%q subject=%q body=%q", req.SenderID, req.Recipient, req.Subject, req.Body)
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

	// 2. Generate per-email AES-256 key
	dataKey := make([]byte, 32)
	if _, err := rand.Read(dataKey); err != nil {
		log.Printf("Key generation failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Encryption key generation failed"}`))
		return
	}

	// 3. Encrypt compressed content
	ciphertext, nonce, err := encryptAESGCM(dataKey, compressed)
	if err != nil {
		log.Printf("Encryption failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Encryption failed"}`))
		return
	}
	encrypted := append(nonce, ciphertext...) // Store nonce+ct together

	// 4. Generate blobID
	blobID := uuid.New().String() + ".blob"

	// 5. Upload to R2
	if err := pkg.UploadToR2(blobID, encrypted); err != nil {
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
	encryptedKeyB64 := base64.StdEncoding.EncodeToString(dataKey) // For demo, store raw key; in production, encrypt with user key

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

	_, err = srv.db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, sha256_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, req.SenderID, req.Recipient, req.Subject, blobID, encryptedKeyB64, hashB64, time.Now(),
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
