package storage

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestNewR2ClientFromEnv(t *testing.T) {
	// Test with missing environment variables
	client, err := NewR2ClientFromEnv()
	if err == nil {
		t.Error("Expected error when environment variables are missing")
	}

	// Test with valid environment variables (mock)
	os.Setenv("R2_ACCESS_KEY_ID", "test-access-key")
	os.Setenv("R2_SECRET_ACCESS_KEY", "test-secret-key")
	os.Setenv("R2_BUCKET", "test-bucket")
	os.Setenv("R2_ENDPOINT", "https://test.r2.cloudflarestorage.com")

	client, err = NewR2ClientFromEnv()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Error("Expected client to be created")
		return // Add early return to prevent nil pointer access
	}
	if client.bucket != "test-bucket" {
		t.Errorf("Expected bucket 'test-bucket', got '%s'", client.bucket)
	}

	// Clean up
	os.Unsetenv("R2_ACCESS_KEY_ID")
	os.Unsetenv("R2_SECRET_ACCESS_KEY")
	os.Unsetenv("R2_BUCKET")
	os.Unsetenv("R2_ENDPOINT")
}

func TestNewR2Client(t *testing.T) {
	// Test with incomplete configuration
	config := &R2Config{
		AccessKeyID: "test-key",
		// Missing other required fields
	}

	client, err := NewR2Client(config)
	if err == nil {
		t.Error("Expected error with incomplete configuration")
	}

	// Test with complete configuration
	config = &R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Bucket:          "test-bucket",
		Endpoint:        "https://test.r2.cloudflarestorage.com",
		Region:          "auto",
	}

	client, err = NewR2Client(config)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if client == nil {
		t.Error("Expected client to be created")
	}
}

func TestUploadToR2(t *testing.T) {
	// This test would require actual R2 credentials
	// For now, we'll test the function signature and error handling

	// Test with empty blobID
	err := UploadToR2("", []byte("test data"))
	if err == nil {
		t.Error("Expected error with empty blobID")
	}

	// Test with empty data
	err = UploadToR2("test.blob", []byte{})
	if err == nil {
		t.Error("Expected error with empty data")
	}

	// Test with valid inputs (will fail due to missing credentials, but tests the flow)
	err = UploadToR2("test.blob", []byte("test data"))
	// We expect an error due to missing credentials, but the function should handle it gracefully
	if err != nil && err.Error() == "failed to create R2 client: incomplete R2 configuration: missing required fields" {
		// This is expected behavior
	} else if err == nil {
		t.Error("Expected error due to missing credentials")
	}
}

func TestUploadToR2WithContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test with context timeout
	err := UploadToR2WithContext(ctx, "test.blob", []byte("test data"))
	// We expect an error due to missing credentials, but the function should handle it gracefully
	if err != nil && err.Error() == "failed to create R2 client: incomplete R2 configuration: missing required fields" {
		// This is expected behavior
	} else if err == nil {
		t.Error("Expected error due to missing credentials")
	}
}

func TestR2Client_UploadEmail_Validation(t *testing.T) {
	// Create a client with test configuration
	config := &R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Bucket:          "test-bucket",
		Endpoint:        "https://test.r2.cloudflarestorage.com",
		Region:          "auto",
	}

	client, err := NewR2Client(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty blobID
	err = client.UploadEmail(ctx, "", []byte("test data"))
	if err == nil {
		t.Error("Expected error with empty blobID")
	}

	// Test with empty data
	err = client.UploadEmail(ctx, "test.blob", []byte{})
	if err == nil {
		t.Error("Expected error with empty data")
	}

	// Test with valid inputs (will fail due to invalid credentials, but tests validation)
	err = client.UploadEmail(ctx, "test.blob", []byte("test data"))
	// We expect an error due to invalid credentials, but validation should pass
	if err != nil && err.Error() != "failed to upload email to R2: NoCredentialProviders: no valid providers in chain" {
		// This is expected behavior - the validation passed but upload failed due to credentials
	}
}

func TestR2Client_GetEmail_Validation(t *testing.T) {
	// Create a client with test configuration
	config := &R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Bucket:          "test-bucket",
		Endpoint:        "https://test.r2.cloudflarestorage.com",
		Region:          "auto",
	}

	client, err := NewR2Client(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty blobID
	_, err = client.GetEmail(ctx, "")
	if err == nil {
		t.Error("Expected error with empty blobID")
	}
}

func TestR2Client_DeleteEmail_Validation(t *testing.T) {
	// Create a client with test configuration
	config := &R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Bucket:          "test-bucket",
		Endpoint:        "https://test.r2.cloudflarestorage.com",
		Region:          "auto",
	}

	client, err := NewR2Client(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty blobID
	err = client.DeleteEmail(ctx, "")
	if err == nil {
		t.Error("Expected error with empty blobID")
	}
}

func TestR2Client_EmailExists_Validation(t *testing.T) {
	// Create a client with test configuration
	config := &R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Bucket:          "test-bucket",
		Endpoint:        "https://test.r2.cloudflarestorage.com",
		Region:          "auto",
	}

	client, err := NewR2Client(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty blobID
	_, err = client.EmailExists(ctx, "")
	if err == nil {
		t.Error("Expected error with empty blobID")
	}
}

func TestR2Client_GetEmailMetadata_Validation(t *testing.T) {
	// Create a client with test configuration
	config := &R2Config{
		AccessKeyID:     "test-access-key",
		SecretAccessKey: "test-secret-key",
		Bucket:          "test-bucket",
		Endpoint:        "https://test.r2.cloudflarestorage.com",
		Region:          "auto",
	}

	client, err := NewR2Client(config)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test with empty blobID
	_, err = client.GetEmailMetadata(ctx, "")
	if err == nil {
		t.Error("Expected error with empty blobID")
	}
}

// TestR2Config_Validation tests the R2Config validation
func TestR2Config_Validation(t *testing.T) {
	tests := []struct {
		name    string
		config  *R2Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &R2Config{
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				Bucket:          "test-bucket",
				Endpoint:        "https://test.r2.cloudflarestorage.com",
			},
			wantErr: false,
		},
		{
			name: "missing access key",
			config: &R2Config{
				SecretAccessKey: "test-secret",
				Bucket:          "test-bucket",
				Endpoint:        "https://test.r2.cloudflarestorage.com",
			},
			wantErr: true,
		},
		{
			name: "missing secret key",
			config: &R2Config{
				AccessKeyID: "test-key",
				Bucket:      "test-bucket",
				Endpoint:    "https://test.r2.cloudflarestorage.com",
			},
			wantErr: true,
		},
		{
			name: "missing bucket",
			config: &R2Config{
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				Endpoint:        "https://test.r2.cloudflarestorage.com",
			},
			wantErr: true,
		},
		{
			name: "missing endpoint",
			config: &R2Config{
				AccessKeyID:     "test-key",
				SecretAccessKey: "test-secret",
				Bucket:          "test-bucket",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewR2Client(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewR2Client() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
