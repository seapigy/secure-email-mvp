package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"
)

type GetEmailRequest struct {
	EmailID string `json:"email_id"`
}

type GetEmailResponse struct {
	EmailID   string    `json:"email_id"`
	SenderID  string    `json:"sender_id"`
	Recipient string    `json:"recipient"`
	Subject   string    `json:"subject"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
}

// getEmailHandler handles GET /api/email/get. It retrieves, decrypts, and decompresses email content.
func (srv *Server) getEmailHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("getEmailHandler started")

	// Parse request
	var req GetEmailRequest
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

	if req.EmailID == "" {
		log.Printf("Missing email_id")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Missing email_id"}`))
		return
	}

	// 1. Retrieve email metadata from database
	var (
		blobID, encryptedKeyB64, nonceB64, authTagB64, compressionAlgo string
		senderID, recipient, subject                                   string
		createdAt                                                      time.Time
	)

	err := srv.db.QueryRow(`
		SELECT encrypted_blob_url, encrypted_key, encryption_nonce, encryption_auth_tag, 
		       compression_algo, sender_id, recipient, subject, created_at
		FROM emails WHERE email_id = ?`,
		req.EmailID,
	).Scan(&blobID, &encryptedKeyB64, &nonceB64, &authTagB64, &compressionAlgo,
		&senderID, &recipient, &subject, &createdAt)

	if err != nil {
		log.Printf("Database query failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error":"Email not found"}`))
		return
	}

	// Check if the authenticated user is the sender of this email
	if senderID != userID {
		log.Printf("Unauthorized access attempt: user %s trying to access email from sender %s", userID, senderID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error":"Access denied"}`))
		return
	}

	// 2. Decode base64 fields
	encryptedKey, err := base64.StdEncoding.DecodeString(encryptedKeyB64)
	if err != nil {
		log.Printf("Failed to decode encrypted key: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encryption key"}`))
		return
	}

	nonce, err := base64.StdEncoding.DecodeString(nonceB64)
	if err != nil {
		log.Printf("Failed to decode nonce: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encryption nonce"}`))
		return
	}

	authTag, err := base64.StdEncoding.DecodeString(authTagB64)
	if err != nil {
		log.Printf("Failed to decode auth tag: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encryption auth tag"}`))
		return
	}

	// 3. Retrieve encrypted blob from R2
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	encryptedBlob, err := storage.GetEmailFromR2(ctx, blobID)
	if err != nil {
		log.Printf("Failed to retrieve blob from R2: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve encrypted content"}`))
		return
	}

	// 4. Separate ciphertext and auth tag
	if len(encryptedBlob) < 16 {
		log.Printf("Encrypted blob too short: %d bytes", len(encryptedBlob))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Invalid encrypted content"}`))
		return
	}

	ciphertext := encryptedBlob[:len(encryptedBlob)-16]
	blobAuthTag := encryptedBlob[len(encryptedBlob)-16:]

	// 5. Verify auth tag matches
	if !bytes.Equal(authTag, blobAuthTag) {
		log.Printf("Auth tag mismatch")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Content integrity check failed"}`))
		return
	}

	// 6. Create EncryptedData struct for decryption
	encryptedData := &auth.EncryptedData{
		Ciphertext: ciphertext,
		Key:        encryptedKey,
		Nonce:      nonce,
		AuthTag:    authTag,
	}

	// 7. Decrypt the content
	compressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Printf("Decryption failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Decryption failed"}`))
		return
	}

	// 8. Decompress the content
	var plaintext []byte
	if compressionAlgo == "gzip" {
		reader := bytes.NewReader(compressed)
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			log.Printf("Failed to create gzip reader: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Decompression failed"}`))
			return
		}
		defer gzReader.Close()

		plaintext, err = io.ReadAll(gzReader)
		if err != nil {
			log.Printf("Failed to read decompressed content: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"Decompression failed"}`))
			return
		}
	} else {
		// No compression or unknown algorithm
		plaintext = compressed
	}

	// 9. Update access tracking
	_, err = srv.db.Exec(`
		UPDATE emails SET 
			last_access_at = CURRENT_TIMESTAMP,
			access_count = access_count + 1
		WHERE email_id = ?`,
		req.EmailID,
	)
	if err != nil {
		log.Printf("Failed to update access tracking: %v", err)
		// Don't fail the request for tracking errors
	}

	// 10. Return decrypted email
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(GetEmailResponse{
		EmailID:   req.EmailID,
		SenderID:  senderID,
		Recipient: recipient,
		Subject:   subject,
		Body:      string(plaintext),
		CreatedAt: createdAt,
		Status:    "success",
	})
}
