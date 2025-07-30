package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"log"

	"secure-email-mvp/pkg/auth"
)

// Example demonstrating AES-256-GCM encryption for email content
func main() {
	// Simulate email content (subject + body)
	emailContent := `Subject: Test Email
Body: This is a test email with some content that will be compressed and encrypted using AES-256-GCM.

The content includes multiple lines and various characters to test the encryption process.
This simulates what would happen in the /api/email/send handler.`

	fmt.Println("=== Email Encryption Example ===")
	fmt.Printf("Original content length: %d bytes\n", len(emailContent))

	// Step 1: Compress the content (like in the email handler)
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write([]byte(emailContent)); err != nil {
		log.Fatalf("Compression failed: %v", err)
	}
	gz.Close()
	compressed := buf.Bytes()

	fmt.Printf("Compressed content length: %d bytes (%.1f%% compression)\n",
		len(compressed), float64(len(compressed))/float64(len(emailContent))*100)

	// Step 2: Encrypt the compressed content using AES-256-GCM
	encryptedData, err := auth.EncryptAES256GCM(compressed)
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	// Step 3: Display the encryption components
	fmt.Println("\n=== Encryption Components ===")
	fmt.Printf("Key (32 bytes): %s\n", base64.StdEncoding.EncodeToString(encryptedData.Key))
	fmt.Printf("Nonce (12 bytes): %s\n", base64.StdEncoding.EncodeToString(encryptedData.Nonce))
	fmt.Printf("Auth Tag (16 bytes): %s\n", base64.StdEncoding.EncodeToString(encryptedData.AuthTag))
	fmt.Printf("Ciphertext length: %d bytes\n", len(encryptedData.Ciphertext))

	// Step 4: Combine ciphertext and auth tag for storage (like in the handler)
	encryptedBlob := append(encryptedData.Ciphertext, encryptedData.AuthTag...)
	fmt.Printf("Encrypted blob length: %d bytes\n", len(encryptedBlob))

	// Step 5: Validate the encrypted data structure
	if err := auth.ValidateEncryptedData(encryptedData); err != nil {
		log.Fatalf("Validation failed: %v", err)
	}
	fmt.Println("✓ Encrypted data validation passed")

	// Step 6: Decrypt and verify (for demonstration)
	decrypted, err := auth.DecryptAES256GCM(encryptedData)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err)
	}

	// Step 7: Decompress and verify the original content
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

	// Step 8: Verify the result matches the original
	if string(decompressed) == emailContent {
		fmt.Println("✓ Decryption and decompression successful - content matches original")
	} else {
		fmt.Println("✗ Content mismatch after decryption")
	}

	fmt.Println("\n=== Summary ===")
	fmt.Println("This example demonstrates:")
	fmt.Println("1. Content compression with gzip")
	fmt.Println("2. AES-256-GCM encryption with random key and nonce")
	fmt.Println("3. Separation of encryption components (key, nonce, auth tag)")
	fmt.Println("4. Validation of encrypted data structure")
	fmt.Println("5. Successful decryption and decompression")
	fmt.Println("\nThe encryption components would be stored separately:")
	fmt.Println("- Key: In SQLite (encrypted with user's key)")
	fmt.Println("- Nonce: In SQLite (encrypted with user's key)")
	fmt.Println("- Auth Tag: In SQLite (encrypted with user's key)")
	fmt.Println("- Ciphertext + Auth Tag: In Cloudflare R2")
}
