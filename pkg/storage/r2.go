package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// R2Config holds Cloudflare R2 configuration
type R2Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	Endpoint        string
	Region          string
}

// R2Client wraps the S3 client for R2 operations
type R2Client struct {
	s3Client *s3.S3
	bucket   string
}

// NewR2Client creates a new R2 client with the given configuration
func NewR2Client(config *R2Config) (*R2Client, error) {
	if config.AccessKeyID == "" || config.SecretAccessKey == "" || config.Bucket == "" || config.Endpoint == "" {
		return nil, fmt.Errorf("incomplete R2 configuration: missing required fields")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String(config.Region),
		Credentials:      credentials.NewStaticCredentials(config.AccessKeyID, config.SecretAccessKey, ""),
		Endpoint:         aws.String(config.Endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 session: %w", err)
	}

	return &R2Client{
		s3Client: s3.New(sess),
		bucket:   config.Bucket,
	}, nil
}

// NewR2ClientFromEnv creates a new R2 client using environment variables
func NewR2ClientFromEnv() (*R2Client, error) {
	config := &R2Config{
		AccessKeyID:     os.Getenv("CLOUDFLARE_R2_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("CLOUDFLARE_R2_SECRET_KEY"),
		Bucket:          os.Getenv("CLOUDFLARE_R2_BUCKET"),
		Endpoint:        os.Getenv("CLOUDFLARE_R2_ENDPOINT"),
		Region:          "auto", // R2 uses "auto" region
	}

	return NewR2Client(config)
}

// UploadEmail uploads encrypted email data to R2 with proper path structure
func (r *R2Client) UploadEmail(ctx context.Context, blobID string, data []byte) error {
	// Validate inputs
	if blobID == "" {
		return fmt.Errorf("blobID cannot be empty")
	}
	if len(data) == 0 {
		return fmt.Errorf("data cannot be empty")
	}

	// Construct the full path for email storage
	objectKey := path.Join("emails", blobID)

	// Create upload input with proper metadata
	input := &s3.PutObjectInput{
		Bucket:        aws.String(r.bucket),
		Key:           aws.String(objectKey),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String("application/octet-stream"),
		ContentLength: aws.Int64(int64(len(data))),
		Metadata: map[string]*string{
			"upload-timestamp": aws.String(time.Now().UTC().Format(time.RFC3339)),
			"content-type":     aws.String("application/octet-stream"),
			"encryption":       aws.String("aes-256-gcm"),
			"compression":      aws.String("gzip"),
		},
	}

	// Upload with context
	_, err := r.s3Client.PutObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload email to R2: %w", err)
	}

	return nil
}

// UploadToR2 is a convenience function that creates a client and uploads data
// This maintains backward compatibility with existing code
func UploadToR2(blobID string, data []byte) error {
	client, err := NewR2ClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create R2 client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return client.UploadEmail(ctx, blobID, data)
}

// UploadToR2WithContext uploads data with a custom context
func UploadToR2WithContext(ctx context.Context, blobID string, data []byte) error {
	// Check if we're in test mode
	if os.Getenv("TEST_MODE") == "1" {
		// In test mode, just log the upload and return success
		log.Printf("TEST_MODE: Mocking R2 upload for blob %s", blobID)
		return nil
	}

	client, err := NewR2ClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create R2 client: %w", err)
	}

	return client.UploadEmail(ctx, blobID, data)
}

// GetEmailFromR2 is a convenience function that creates a client and retrieves data
func GetEmailFromR2(ctx context.Context, blobID string) ([]byte, error) {
	client, err := NewR2ClientFromEnv()
	if err != nil {
		return nil, fmt.Errorf("failed to create R2 client: %w", err)
	}

	return client.GetEmail(ctx, blobID)
}

// DeleteBlob is a convenience function that creates a client and deletes data
func DeleteBlob(ctx context.Context, blobID string) error {
	client, err := NewR2ClientFromEnv()
	if err != nil {
		return fmt.Errorf("failed to create R2 client: %w", err)
	}

	return client.DeleteEmail(ctx, blobID)
}

// GetEmail retrieves encrypted email data from R2
func (r *R2Client) GetEmail(ctx context.Context, blobID string) ([]byte, error) {
	if blobID == "" {
		return nil, fmt.Errorf("blobID cannot be empty")
	}

	objectKey := path.Join("emails", blobID)

	input := &s3.GetObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	}

	result, err := r.s3Client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve email from R2: %w", err)
	}
	defer result.Body.Close()

	// Read the entire body
	var buf bytes.Buffer
	_, err = buf.ReadFrom(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read email data: %w", err)
	}

	return buf.Bytes(), nil
}

// DeleteEmail removes an email from R2
func (r *R2Client) DeleteEmail(ctx context.Context, blobID string) error {
	if blobID == "" {
		return fmt.Errorf("blobID cannot be empty")
	}

	objectKey := path.Join("emails", blobID)

	input := &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	}

	_, err := r.s3Client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete email from R2: %w", err)
	}

	return nil
}

// EmailExists checks if an email exists in R2
func (r *R2Client) EmailExists(ctx context.Context, blobID string) (bool, error) {
	if blobID == "" {
		return false, fmt.Errorf("blobID cannot be empty")
	}

	objectKey := path.Join("emails", blobID)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	}

	_, err := r.s3Client.HeadObjectWithContext(ctx, input)
	if err != nil {
		// Check if it's a "not found" error
		if request.IsErrorRetryable(err) || request.IsErrorThrottle(err) {
			return false, fmt.Errorf("failed to check email existence: %w", err)
		}
		// Assume it's a "not found" error
		return false, nil
	}

	return true, nil
}

// GetEmailMetadata retrieves metadata for an email without downloading the content
func (r *R2Client) GetEmailMetadata(ctx context.Context, blobID string) (map[string]string, error) {
	if blobID == "" {
		return nil, fmt.Errorf("blobID cannot be empty")
	}

	objectKey := path.Join("emails", blobID)

	input := &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(objectKey),
	}

	result, err := r.s3Client.HeadObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get email metadata: %w", err)
	}

	metadata := make(map[string]string)
	for key, value := range result.Metadata {
		if value != nil {
			metadata[key] = *value
		}
	}

	// Add standard metadata
	if result.ContentLength != nil {
		metadata["content-length"] = fmt.Sprintf("%d", *result.ContentLength)
	}
	if result.LastModified != nil {
		metadata["last-modified"] = result.LastModified.Format(time.RFC3339)
	}

	return metadata, nil
}
