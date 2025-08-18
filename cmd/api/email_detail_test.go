package main

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"

	_ "modernc.org/sqlite"
)

// TestGetEmailDetailAuthorized tests successful email detail retrieval for authorized sender
func TestGetEmailDetailAuthorized(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-authorized-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test data
	testUserID := "test-user-authorized"
	emailID := "email-authorized-test"
	recipient := "recipient@test.com"
	subject := "Test Subject"
	body := "This is a test email body (mock decryption)"

	// Encrypt and compress the body
	_, key, nonce, authTag, err := encryptAndCompressBody(body)
	if err != nil {
		t.Fatalf("Failed to encrypt body: %v", err)
	}

	// Insert test email
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, recipient, subject, "blob-url",
		base64.StdEncoding.EncodeToString(key),
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(authTag),
		"gzip", "hash-1", time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL parameters
	router := mux.NewRouter()
	router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response EmailDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response fields
	if response.EmailID != emailID {
		t.Errorf("Expected email_id '%s', got '%s'", emailID, response.EmailID)
	}
	if response.Recipient != recipient {
		t.Errorf("Expected recipient '%s', got '%s'", recipient, response.Recipient)
	}
	if response.Subject != subject {
		t.Errorf("Expected subject '%s', got '%s'", subject, response.Subject)
	}
	if response.Body == nil {
		t.Error("Expected body to be present, got nil")
	} else if *response.Body != body {
		t.Errorf("Expected body '%s', got '%s'", body, *response.Body)
	}
	if response.Status != "delivered" {
		t.Errorf("Expected status 'delivered', got '%s'", response.Status)
	}
}

// TestGetEmailDetailUnauthorized tests that non-sender gets 403 Forbidden
func TestGetEmailDetailUnauthorized(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-unauthorized-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test data
	senderID := "sender-user"
	unauthorizedUserID := "unauthorized-user"
	emailID := "email-unauthorized-test"

	// Insert test email
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, senderID, "recipient@test.com", "Test Subject", "blob-url",
		"key", "nonce", "tag", "gzip", "hash-1", time.Now(),
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for unauthorized user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": unauthorizedUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL parameters
	router := mux.NewRouter()
	router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Access forbidden" {
			t.Errorf("Expected error 'Access forbidden', got '%s'", response["error"])
		}
	}
}

// TestGetEmailDetailNotFound tests 404 for nonexistent email
func TestGetEmailDetailNotFound(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-notfound-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "test-user",
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request for nonexistent email
	req := httptest.NewRequest("GET", "/api/email/detail/nonexistent-email", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL parameters
	router := mux.NewRouter()
	router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
	}

	// Check error message
	var response map[string]string
	if err := json.NewDecoder(w.Body).Decode(&response); err == nil {
		if response["error"] != "Email not found" {
			t.Errorf("Expected error 'Email not found', got '%s'", response["error"])
		}
	}
}

// TestGetEmailDetailExpired tests that expired emails return status expired and body null
func TestGetEmailDetailExpired(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-expired-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test data
	testUserID := "test-user-expired"
	emailID := "email-expired-test"
	expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

	// Insert expired email
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at, expires_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@test.com", "Expired Subject", "blob-url",
		"key", "nonce", "tag", "gzip", "hash-1", time.Now(), expiresAt,
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL parameters
	router := mux.NewRouter()
	router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response EmailDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if response.Status != "expired" {
		t.Errorf("Expected status 'expired', got '%s'", response.Status)
	}
	if response.Body != nil {
		t.Error("Expected body to be null for expired email, got non-nil")
	}
	if response.ExpiresAt == nil {
		t.Error("Expected expires_at to be present")
	}
}

// TestGetEmailDetailDeleted tests that deleted emails return status deleted and body null
func TestGetEmailDetailDeleted(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-deleted-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test data
	testUserID := "test-user-deleted"
	emailID := "email-deleted-test"

	// Insert deleted email
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at, self_destructed
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@test.com", "Deleted Subject", "blob-url",
		"key", "nonce", "tag", "gzip", "hash-1", time.Now(), 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// Create request
	req := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	// Create response recorder
	w := httptest.NewRecorder()

	// Set up router with URL parameters
	router := mux.NewRouter()
	router.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
		t.Logf("Response body: %s", w.Body.String())
		return
	}

	// Parse response
	var response EmailDetailResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify response
	if response.Status != "deleted" {
		t.Errorf("Expected status 'deleted', got '%s'", response.Status)
	}
	if response.Body != nil {
		t.Error("Expected body to be null for deleted email, got non-nil")
	}
}

// TestGetEmailDetailBurnAfterRead tests burn-after-read functionality
func TestGetEmailDetailBurnAfterRead(t *testing.T) {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-detail-burn-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer db.Close()

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	// Create test data
	testUserID := "test-user-burn"
	emailID := "email-burn-test"
	body := "This is a test email body (mock decryption)"

	// Encrypt and compress the body
	_, key, nonce, authTag, err := encryptAndCompressBody(body)
	if err != nil {
		t.Fatalf("Failed to encrypt body: %v", err)
	}

	// Insert burn-after-read email
	_, err = db.Exec(`
		INSERT INTO emails (
			email_id, sender_id, recipient, subject, encrypted_blob_url, encrypted_key, 
			encryption_nonce, encryption_auth_tag, compression_algo, sha256_hash, created_at, burn_after_read
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		emailID, testUserID, "recipient@test.com", "Burn After Read", "blob-url",
		base64.StdEncoding.EncodeToString(key),
		base64.StdEncoding.EncodeToString(nonce),
		base64.StdEncoding.EncodeToString(authTag),
		"gzip", "hash-1", time.Now(), 1,
	)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Create server with test database
	srv := &Server{db: db}

	// Set JWT secret for testing
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-signing")
	defer os.Unsetenv("JWT_SECRET")

	// Generate JWT token for test user
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": testUserID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte("test-secret-key-for-jwt-signing"))
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}

	// First access - should return body and status "delivered"
	req1 := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req1.Header.Set("Authorization", "Bearer "+tokenString)
	w1 := httptest.NewRecorder()

	router1 := mux.NewRouter()
	router1.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router1.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Errorf("First access: Expected status 200, got %d", w1.Code)
		t.Logf("Response body: %s", w1.Body.String())
		return
	}

	var response1 EmailDetailResponse
	if err := json.NewDecoder(w1.Body).Decode(&response1); err != nil {
		t.Fatalf("Failed to decode first response: %v", err)
	}

	if response1.Status != "delivered" {
		t.Errorf("First access: Expected status 'delivered', got '%s'", response1.Status)
	}
	if response1.Body == nil {
		t.Error("First access: Expected body to be present, got nil")
	} else if *response1.Body != body {
		t.Errorf("First access: Expected body '%s', got '%s'", body, *response1.Body)
	}

	// Second access - should return status "deleted" and body null
	req2 := httptest.NewRequest("GET", "/api/email/detail/"+emailID, nil)
	req2.Header.Set("Authorization", "Bearer "+tokenString)
	w2 := httptest.NewRecorder()

	router2 := mux.NewRouter()
	router2.Handle("/api/email/detail/{email_id}", jwtMiddleware(http.HandlerFunc(srv.getEmailDetailHandler))).Methods("GET")
	router2.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("Second access: Expected status 200, got %d", w2.Code)
		t.Logf("Response body: %s", w2.Body.String())
		return
	}

	var response2 EmailDetailResponse
	if err := json.NewDecoder(w2.Body).Decode(&response2); err != nil {
		t.Fatalf("Failed to decode second response: %v", err)
	}

	if response2.Status != "deleted" {
		t.Errorf("Second access: Expected status 'deleted', got '%s'", response2.Status)
	}
	if response2.Body != nil {
		t.Error("Second access: Expected body to be null, got non-nil")
	}
}

// Helper function to encrypt and compress email body for testing
func encryptAndCompressBody(body string) ([]byte, []byte, []byte, []byte, error) {
	// Compress the body
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write([]byte(body)); err != nil {
		return nil, nil, nil, nil, err
	}
	gw.Close()
	compressed := buf.Bytes()

	// Generate encryption key
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, nil, nil, nil, err
	}

	// Generate nonce
	nonce := make([]byte, 12) // GCM nonce
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, nil, nil, err
	}

	// Encrypt
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	ciphertext := aesGCM.Seal(nil, nonce, compressed, nil)

	// Extract auth tag (last 16 bytes)
	authTag := ciphertext[len(ciphertext)-16:]
	encryptedData := ciphertext[:len(ciphertext)-16]

	return encryptedData, key, nonce, authTag, nil
}
