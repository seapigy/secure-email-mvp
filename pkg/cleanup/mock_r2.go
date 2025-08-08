package cleanup

import (
	"context"
	"fmt"
	"log"
)

// MockR2Client is a mock implementation of the R2 client for testing
type MockR2Client struct {
	deletedBlobs map[string]bool
}

// Ensure MockR2Client implements R2ClientInterface
var _ R2ClientInterface = (*MockR2Client)(nil)

// NewMockR2Client creates a new mock R2 client
func NewMockR2Client() *MockR2Client {
	return &MockR2Client{
		deletedBlobs: make(map[string]bool),
	}
}

// DeleteEmail mocks the deletion of an email from R2
func (m *MockR2Client) DeleteEmail(ctx context.Context, blobID string) error {
	if blobID == "" {
		return fmt.Errorf("blobID cannot be empty")
	}

	log.Printf("Mock: Deleting email blob %s from R2", blobID)
	m.deletedBlobs[blobID] = true
	return nil
}

// GetEmail mocks retrieving an email from R2
func (m *MockR2Client) GetEmail(ctx context.Context, blobID string) ([]byte, error) {
	if blobID == "" {
		return nil, fmt.Errorf("blobID cannot be empty")
	}

	// Return mock data
	return []byte("mock-email-content"), nil
}

// UploadEmail mocks uploading an email to R2
func (m *MockR2Client) UploadEmail(ctx context.Context, blobID string, data []byte) error {
	if blobID == "" {
		return fmt.Errorf("blobID cannot be empty")
	}

	log.Printf("Mock: Uploading email blob %s to R2", blobID)
	return nil
}

// EmailExists mocks checking if an email exists in R2
func (m *MockR2Client) EmailExists(ctx context.Context, blobID string) (bool, error) {
	if blobID == "" {
		return false, fmt.Errorf("blobID cannot be empty")
	}

	// Mock that all emails exist
	return true, nil
}

// GetEmailMetadata mocks retrieving email metadata from R2
func (m *MockR2Client) GetEmailMetadata(ctx context.Context, blobID string) (map[string]string, error) {
	if blobID == "" {
		return nil, fmt.Errorf("blobID cannot be empty")
	}

	// Return mock metadata
	return map[string]string{
		"content-type": "application/octet-stream",
		"encryption":   "aes-256-gcm",
		"compression":  "gzip",
	}, nil
}

// GetDeletedBlobs returns the list of blobs that were "deleted" during testing
func (m *MockR2Client) GetDeletedBlobs() []string {
	var blobs []string
	for blobID := range m.deletedBlobs {
		blobs = append(blobs, blobID)
	}
	return blobs
}

// ClearDeletedBlobs clears the list of deleted blobs (useful for test cleanup)
func (m *MockR2Client) ClearDeletedBlobs() {
	m.deletedBlobs = make(map[string]bool)
}
