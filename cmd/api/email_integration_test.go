package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// TestCompleteEmailFlow tests the entire email lifecycle:
// 1. Send email (compress -> encrypt -> upload to R2 -> store metadata)
// 2. Retrieve email (fetch from R2 -> decrypt -> decompress)
// 3. Verify original content matches retrieved content
func TestCompleteEmailFlow(t *testing.T) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		t.Skipf("Skipping test: no .env file found: %v", err)
	}

	// Check if R2 credentials are available
	if os.Getenv("CLOUDFLARE_R2_ACCESS_KEY") == "" {
		t.Skip("Skipping test: R2 credentials not available")
	}

	// Create test server
	srv := createTestServer(t)
	defer srv.db.Close()

	// Test data
	originalEmail := SendEmailRequest{
		Recipient: "test@example.com",
		Subject:   "Test Email Subject",
		Body:      "This is a test email body with some content to compress and encrypt. It should be longer to make compression worthwhile.",
	}

	// Step 1: Send email
	t.Log("Step 1: Sending email...")
	sendResp := sendEmail(t, srv, originalEmail)
	if sendResp.Error != "" {
		t.Fatalf("Send email failed: %s", sendResp.Error)
	}
	t.Logf("Email sent successfully, blob ID: %s", sendResp.BlobID)

	// Step 2: Retrieve email
	t.Log("Step 2: Retrieving email...")
	getResp := getEmail(t, srv, sendResp.BlobID)
	if getResp.Error != "" {
		t.Fatalf("Get email failed: %s", getResp.Error)
	}

	// Step 3: Verify content matches
	t.Log("Step 3: Verifying content...")
	if getResp.Body != originalEmail.Body {
		t.Errorf("Content mismatch!\nOriginal: %q\nRetrieved: %q", originalEmail.Body, getResp.Body)
	} else {
		t.Log("✅ Content verification successful")
	}

	// Step 4: Verify metadata
	// Note: SenderID is now obtained from authenticated context, so we can't verify it in this test
	// since the test doesn't have authentication. The actual sender_id will be set by the authenticated user.
	if getResp.Recipient != originalEmail.Recipient {
		t.Errorf("Recipient mismatch: expected %q, got %q", originalEmail.Recipient, getResp.Recipient)
	}
	if getResp.Subject != originalEmail.Subject {
		t.Errorf("Subject mismatch: expected %q, got %q", originalEmail.Subject, getResp.Subject)
	}

	t.Log("✅ Complete email flow test passed")
}

// TestEncryptionDecryptionFlow tests the encryption/decryption cycle in isolation
func TestEncryptionDecryptionFlow(t *testing.T) {
	// Test data
	originalContent := "This is test content that will be compressed and encrypted, then decrypted and decompressed to verify the process works correctly."

	// Step 1: Compress
	t.Log("Step 1: Compressing content...")
	compressed := compressContent(t, originalContent)
	t.Logf("Compressed size: %d bytes (original: %d bytes)", len(compressed), len(originalContent))

	// Step 2: Encrypt
	t.Log("Step 2: Encrypting content...")
	encryptedData, err := auth.EncryptAES256GCM(compressed)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}
	t.Logf("Encrypted data size: %d bytes", len(encryptedData.Ciphertext))

	// Step 3: Decrypt
	t.Log("Step 3: Decrypting content...")
	decryptedCompressed, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	// Step 4: Decompress
	t.Log("Step 4: Decompressing content...")
	decryptedContent := decompressContent(t, decryptedCompressed)

	// Step 5: Verify
	if decryptedContent != originalContent {
		t.Errorf("Content mismatch after encryption/decryption cycle!\nOriginal: %q\nDecrypted: %q", originalContent, decryptedContent)
	} else {
		t.Log("✅ Encryption/decryption cycle successful")
	}
}

// TestR2StorageFlow tests the R2 storage and retrieval in isolation
func TestR2StorageFlow(t *testing.T) {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		t.Skipf("Skipping test: no .env file found: %v", err)
	}

	// Check if R2 credentials are available
	if os.Getenv("CLOUDFLARE_R2_ACCESS_KEY") == "" {
		t.Skip("Skipping test: R2 credentials not available")
	}

	// Test data
	testData := []byte("This is test data for R2 storage verification")
	blobID := fmt.Sprintf("test-r2-flow-%d.blob", time.Now().Unix())

	// Step 1: Upload to R2
	t.Log("Step 1: Uploading to R2...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := storage.UploadToR2WithContext(ctx, blobID, testData)
	if err != nil {
		t.Fatalf("R2 upload failed: %v", err)
	}
	t.Logf("Upload successful: %s", blobID)

	// Step 2: Verify file exists
	t.Log("Step 2: Verifying file exists...")
	client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		t.Fatalf("Failed to create R2 client: %v", err)
	}

	exists, err := client.EmailExists(ctx, blobID)
	if err != nil {
		t.Fatalf("Failed to check file existence: %v", err)
	}
	if !exists {
		t.Fatal("Uploaded file not found in R2")
	}
	t.Log("✅ File exists in R2")

	// Step 3: Retrieve from R2
	t.Log("Step 3: Retrieving from R2...")
	retrievedData, err := storage.GetEmailFromR2(ctx, blobID)
	if err != nil {
		t.Fatalf("R2 retrieval failed: %v", err)
	}

	// Step 4: Verify content matches
	if !bytes.Equal(testData, retrievedData) {
		t.Errorf("Retrieved data doesn't match original!\nOriginal: %q\nRetrieved: %q", testData, retrievedData)
	} else {
		t.Log("✅ R2 storage/retrieval successful")
	}

	// Step 5: Clean up
	t.Log("Step 4: Cleaning up...")
	err = client.DeleteEmail(ctx, blobID)
	if err != nil {
		t.Logf("Warning: Failed to delete test file: %v", err)
	} else {
		t.Log("✅ Cleanup successful")
	}
}

// Helper functions

func createTestServer(t *testing.T) *Server {
	// Create temporary database
	dbPath := fmt.Sprintf("/tmp/test-email-flow-%d.db", time.Now().Unix())
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Apply schema
	schema, err := os.ReadFile("../../schema/emails.sql")
	if err != nil {
		t.Fatalf("Failed to read schema: %v", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		t.Fatalf("Failed to apply schema: %v", err)
	}

	return &Server{db: db}
}

func sendEmail(t *testing.T, srv *Server, req SendEmailRequest) SendEmailResponse {
	// Create request
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq := httptest.NewRequest("POST", "/api/email/send", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	srv.sendEmailHandler(w, httpReq)

	// Parse response
	var resp SendEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return resp
}

func getEmail(t *testing.T, srv *Server, emailID string) GetEmailResponse {
	// Create request
	req := GetEmailRequest{EmailID: emailID}
	reqBody, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	// Create HTTP request
	httpReq := httptest.NewRequest("POST", "/api/email/get", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	srv.getEmailHandler(w, httpReq)

	// Parse response
	var resp GetEmailResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	return resp
}

func compressContent(t *testing.T, content string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(content)); err != nil {
		t.Fatalf("Compression failed: %v", err)
	}
	gz.Close()
	return buf.Bytes()
}

func decompressContent(t *testing.T, compressed []byte) string {
	reader := bytes.NewReader(compressed)
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		t.Fatalf("Failed to read decompressed content: %v", err)
	}

	return string(decompressed)
}

// TestErrorHandling tests various error scenarios
func TestErrorHandling(t *testing.T) {
	// Test 1: Invalid JSON in send request
	t.Run("InvalidJSON", func(t *testing.T) {
		httpReq := httptest.NewRequest("POST", "/api/email/send", strings.NewReader("invalid json"))
		httpReq.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create a minimal server for testing
		srv := &Server{db: nil} // No database needed for this test
		srv.sendEmailHandler(w, httpReq)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})
}
