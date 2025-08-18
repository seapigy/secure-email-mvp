package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"secure-email-mvp/pkg/storage"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using system environment variables: %v", err)
	}

	// Check required environment variables
	requiredEnvVars := []string{
		"CLOUDFLARE_R2_ACCESS_KEY",
		"CLOUDFLARE_R2_SECRET_KEY",
		"CLOUDFLARE_R2_BUCKET",
		"CLOUDFLARE_R2_ENDPOINT",
	}

	missingVars := []string{}
	for _, envVar := range requiredEnvVars {
		if os.Getenv(envVar) == "" {
			missingVars = append(missingVars, envVar)
		}
	}

	if len(missingVars) > 0 {
		log.Fatalf("Missing required environment variables: %v", missingVars)
	}

	// Print configuration (without sensitive data)
	log.Printf("R2 Configuration:")
	log.Printf("  Bucket: %s", os.Getenv("CLOUDFLARE_R2_BUCKET"))
	log.Printf("  Endpoint: %s", os.Getenv("CLOUDFLARE_R2_ENDPOINT"))
	log.Printf("  Access Key ID: %s...", os.Getenv("CLOUDFLARE_R2_ACCESS_KEY")[:8])
	log.Printf("  Secret Key: [HIDDEN]")

	// Prepare test data
	testData := []byte("this is a test blob")
	blobID := fmt.Sprintf("test-upload-%d.blob", time.Now().Unix())

	log.Printf("Starting upload test:")
	log.Printf("  Blob ID: %s", blobID)
	log.Printf("  Data size: %d bytes", len(testData))

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt upload
	log.Printf("Uploading to R2...")
	err := storage.UploadToR2WithContext(ctx, blobID, testData)

	if err != nil {
		log.Printf("❌ Upload failed: %v", err)
		log.Printf("Error details:")
		log.Printf("  - Check your R2 credentials")
		log.Printf("  - Verify bucket exists and is accessible")
		log.Printf("  - Ensure endpoint URL is correct")
		log.Printf("  - Check network connectivity")
		os.Exit(1)
	}

	log.Printf("✅ Upload successful!")
	log.Printf("  Blob uploaded to: emails/%s", blobID)
	log.Printf("  Bucket: %s", os.Getenv("CLOUDFLARE_R2_BUCKET"))

	// Optional: Verify the upload by checking if the file exists
	log.Printf("Verifying upload...")
	client, err := storage.NewR2ClientFromEnv()
	if err != nil {
		log.Printf("⚠️  Warning: Could not create client for verification: %v", err)
		return
	}

	exists, err := client.EmailExists(ctx, blobID)
	if err != nil {
		log.Printf("⚠️  Warning: Could not verify upload: %v", err)
		return
	}

	if exists {
		log.Printf("✅ Verification successful: File exists in R2")
	} else {
		log.Printf("⚠️  Warning: File not found during verification")
	}

	// Optional: Clean up test file
	log.Printf("Cleaning up test file...")
	err = client.DeleteEmail(ctx, blobID)
	if err != nil {
		log.Printf("⚠️  Warning: Could not delete test file: %v", err)
	} else {
		log.Printf("✅ Test file cleaned up successfully")
	}
}
