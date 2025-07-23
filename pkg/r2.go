package pkg

import (
	"bytes"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// UploadToR2 uploads the given data to Cloudflare R2 under the specified blobID (object key).
// It uses environment variables for credentials and endpoint configuration.
func UploadToR2(blobID string, data []byte) error {
	accessKey := os.Getenv("CLOUDFLARE_R2_ACCESS_KEY")
	secretKey := os.Getenv("CLOUDFLARE_R2_SECRET_KEY")
	bucket := os.Getenv("CLOUDFLARE_R2_BUCKET")
	endpoint := os.Getenv("CLOUDFLARE_R2_ENDPOINT")

	if accessKey == "" || secretKey == "" || bucket == "" || endpoint == "" {
		return fmt.Errorf("Cloudflare R2 environment variables not set")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create R2 session: %w", err)
	}

	s3Client := s3.New(sess)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(blobID),
		Body:   bytes.NewReader(data),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to R2: %w", err)
	}
	return nil
}
