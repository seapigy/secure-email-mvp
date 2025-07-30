package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/storage"
)

// EmailUploadExample demonstrates the complete email upload flow
func EmailUploadExample() {
	fmt.Println("=== Email Upload Example ===")

	// Simulate email content
	emailContent := `Subject: Secure Email Test
From: sender@securesystem.email
To: recipient@example.com
Date: ` + time.Now().UTC().Format(time.RFC3339) + `

This is a test email that will be compressed, encrypted, and uploaded to Cloudflare R2.

The content includes:
- Subject and headers
- Email body with multiple lines
- Special characters and formatting
- This demonstrates the complete encryption and upload flow

Best regards,
Secure Email System`

	fmt.Printf("Original email content length: %d bytes\n", len(emailContent))

	// Step 1: Compress the email content
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(emailContent)); err != nil {
		log.Fatalf("Compression failed: %v", err)
	}
	gz.Close()
	compressed := buf.Bytes()

	fmt.Printf("Compressed content length: %d bytes (%.1f%% compression)\n",
		len(compressed), float64(len(compressed))/float64(len(emailContent))*100)

	// Step 2: Encrypt the compressed content
	encryptedData, err := auth.EncryptAES256GCM(compressed)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	fmt.Println("\n=== Encryption Components ===")
	fmt.Printf("Key (32 bytes): %s\n", base64.StdEncoding.EncodeToString(encryptedData.Key))
	fmt.Printf("Nonce (12 bytes): %s\n", base64.StdEncoding.EncodeToString(encryptedData.Nonce))
	fmt.Printf("Auth Tag (16 bytes): %s\n", base64.StdEncoding.EncodeToString(encryptedData.AuthTag))
	fmt.Printf("Ciphertext length: %d bytes\n", len(encryptedData.Ciphertext))

	// Step 3: Combine ciphertext and auth tag for R2 storage
	encryptedBlob := append(encryptedData.Ciphertext, encryptedData.AuthTag...)
	fmt.Printf("Encrypted blob length: %d bytes\n", len(encryptedBlob))

	// Step 4: Generate blob ID
	blobID := fmt.Sprintf("email_%d.blob", time.Now().Unix())
	fmt.Printf("Blob ID: %s\n", blobID)

	// Step 5: Upload to R2 (if credentials are available)
	fmt.Println("\n=== R2 Upload ===")

	// Check if R2 credentials are available
	if os.Getenv("R2_ACCESS_KEY_ID") == "" {
		fmt.Println("⚠️  R2 credentials not available - skipping upload")
		fmt.Println("   Set R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET, R2_ENDPOINT")
		fmt.Println("   to test actual upload functionality")
	} else {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Upload to R2
		err = storage.UploadToR2WithContext(ctx, blobID, encryptedBlob)
		if err != nil {
			fmt.Printf("❌ R2 upload failed: %v\n", err)
		} else {
			fmt.Printf("✅ Successfully uploaded to R2: %s\n", blobID)

			// Test retrieval (optional)
			fmt.Println("\n=== Testing Retrieval ===")
			client, err := storage.NewR2ClientFromEnv()
			if err != nil {
				fmt.Printf("❌ Failed to create R2 client: %v\n", err)
			} else {
				// Check if email exists
				exists, err := client.EmailExists(ctx, blobID)
				if err != nil {
					fmt.Printf("❌ Failed to check email existence: %v\n", err)
				} else {
					fmt.Printf("📧 Email exists in R2: %t\n", exists)
				}

				// Get metadata
				metadata, err := client.GetEmailMetadata(ctx, blobID)
				if err != nil {
					fmt.Printf("❌ Failed to get metadata: %v\n", err)
				} else {
					fmt.Println("📋 Email metadata:")
					for key, value := range metadata {
						fmt.Printf("   %s: %s\n", key, value)
					}
				}
			}
		}
	}

	// Step 6: Prepare database metadata
	fmt.Println("\n=== Database Metadata ===")
	encryptedKeyB64 := base64.StdEncoding.EncodeToString(encryptedData.Key)
	encryptedNonceB64 := base64.StdEncoding.EncodeToString(encryptedData.Nonce)
	encryptedAuthTagB64 := base64.StdEncoding.EncodeToString(encryptedData.AuthTag)

	fmt.Printf("Encrypted Key (base64): %s\n", encryptedKeyB64)
	fmt.Printf("Encrypted Nonce (base64): %s\n", encryptedNonceB64)
	fmt.Printf("Encrypted Auth Tag (base64): %s\n", encryptedAuthTagB64)

	// Step 7: Simulate database insertion
	emailRecord := map[string]interface{}{
		"email_id":           "test-email-123",
		"sender_id":          "sender@securesystem.email",
		"recipient":          "recipient@example.com",
		"subject":            "Secure Email Test",
		"encrypted_blob_url": blobID,
		"encrypted_key":      encryptedKeyB64,
		"compression_algo":   "gzip",
		"created_at":         time.Now().UTC().Format(time.RFC3339),
	}

	fmt.Println("\n=== Database Record ===")
	dbJSON, _ := json.MarshalIndent(emailRecord, "", "  ")
	fmt.Println(string(dbJSON))

	// Step 8: Demonstrate decryption (for verification)
	fmt.Println("\n=== Decryption Verification ===")
	decrypted, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err)
	}

	// Decompress
	reader := bytes.NewReader(decrypted)
	gzReader, err := gzip.NewReader(reader)
	if err != nil {
		log.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer gzReader.Close()

	decompressed, err := io.ReadAll(gzReader)
	if err != nil {
		log.Fatalf("Decompression failed: %v", err)
	}

	// Verify the result matches the original
	if string(decompressed) == emailContent {
		fmt.Println("✅ Decryption and decompression successful - content matches original")
	} else {
		fmt.Println("❌ Content mismatch after decryption")
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("1. Email content compression with gzip")
	fmt.Println("2. AES-256-GCM encryption with component separation")
	fmt.Println("3. R2 upload with proper path structure (emails/{blobID})")
	fmt.Println("4. Database metadata preparation")
	fmt.Println("5. Verification through decryption")
	fmt.Println("6. Error handling and validation")

	fmt.Println("\nStorage locations:")
	fmt.Println("- Encrypted blob: Cloudflare R2 (emails/{blobID})")
	fmt.Println("- Encryption key: SQLite (encrypted with user key)")
	fmt.Println("- Nonce: SQLite (encrypted with user key)")
	fmt.Println("- Auth tag: SQLite (encrypted with user key)")
	fmt.Println("- Metadata: SQLite (email record)")
}
