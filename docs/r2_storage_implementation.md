# Cloudflare R2 Storage Implementation

## Overview

This document describes the implementation of Cloudflare R2 storage for encrypted email uploads in the secure email system. The implementation provides a modular, context-aware storage solution with proper error handling and metadata management.

## Architecture

### Storage Structure
```
R2 Bucket: secure-email-blobs
├── emails/
│   ├── {blobID1}.blob
│   ├── {blobID2}.blob
│   └── ...
└── metadata/
    └── (future use)
```

### File Path Convention
- **Email Blobs**: `emails/{blobID}.blob`
- **Blob ID Format**: `{uuid}.blob` or `email_{timestamp}.blob`
- **Content Type**: `application/octet-stream`

## Implementation Details

### Core Components

#### 1. `R2Config` Struct
```go
type R2Config struct {
    AccessKeyID     string
    SecretAccessKey string
    Bucket          string
    Endpoint        string
    Region          string
}
```

#### 2. `R2Client` Struct
```go
type R2Client struct {
    s3Client *s3.S3
    bucket   string
}
```

#### 3. Key Functions

##### `NewR2ClientFromEnv()`
- Creates client from environment variables
- Validates required configuration
- Returns error if credentials missing

##### `UploadEmail(ctx, blobID, data)`
- Uploads encrypted email data to R2
- Uses proper path structure (`emails/{blobID}`)
- Sets metadata (encryption type, compression, timestamp)
- Supports context for timeout and cancellation

##### `GetEmail(ctx, blobID)`
- Retrieves encrypted email data from R2
- Handles streaming and memory management
- Returns raw bytes for decryption

##### `DeleteEmail(ctx, blobID)`
- Removes email from R2 storage
- Supports cleanup operations

##### `EmailExists(ctx, blobID)`
- Checks if email exists without downloading
- Uses HEAD request for efficiency

##### `GetEmailMetadata(ctx, blobID)`
- Retrieves metadata without downloading content
- Returns upload timestamp, size, encryption info

## Environment Configuration

### Required Environment Variables
```bash
CLOUDFLARE_R2_ACCESS_KEY=your_access_key_id
CLOUDFLARE_R2_SECRET_KEY=your_secret_access_key
CLOUDFLARE_R2_BUCKET=secure-email-blobs
CLOUDFLARE_R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
```

### Optional Configuration
```bash
R2_REGION=auto  # Default for R2
```

## Usage Examples

### Basic Upload
```go
// Upload encrypted email data
err := storage.UploadToR2("email-123.blob", encryptedData)
if err != nil {
    log.Printf("Upload failed: %v", err)
}
```

### Upload with Context
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := storage.UploadToR2WithContext(ctx, "email-123.blob", encryptedData)
if err != nil {
    log.Printf("Upload failed: %v", err)
}
```

### Advanced Client Usage
```go
// Create client
client, err := storage.NewR2ClientFromEnv()
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}

// Upload with metadata
ctx := context.Background()
err = client.UploadEmail(ctx, "email-123.blob", encryptedData)
if err != nil {
    log.Printf("Upload failed: %v", err)
}

// Check if email exists
exists, err := client.EmailExists(ctx, "email-123.blob")
if err != nil {
    log.Printf("Check failed: %v", err)
}

// Get metadata
metadata, err := client.GetEmailMetadata(ctx, "email-123.blob")
if err != nil {
    log.Printf("Metadata retrieval failed: %v", err)
}
```

## Email Upload Flow

### 1. Content Preparation
```go
// Compress email content
var buf bytes.Buffer
gz := gzip.NewWriter(&buf)
gz.Write([]byte(emailContent))
gz.Close()
compressed := buf.Bytes()
```

### 2. Encryption
```go
// Encrypt with AES-256-GCM
encryptedData, err := auth.EncryptAES256GCM(compressed)
if err != nil {
    return fmt.Errorf("encryption failed: %w", err)
}
```

### 3. Blob Preparation
```go
// Combine ciphertext and auth tag
encryptedBlob := append(encryptedData.Ciphertext, encryptedData.AuthTag...)

// Generate blob ID
blobID := uuid.New().String() + ".blob"
```

### 4. R2 Upload
```go
// Upload with context and timeout
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err = storage.UploadToR2WithContext(ctx, blobID, encryptedBlob)
if err != nil {
    return fmt.Errorf("R2 upload failed: %w", err)
}
```

### 5. Database Storage
```sql
INSERT INTO emails (
    email_id, sender_id, recipient, subject,
    encrypted_blob_url, encrypted_key, compression_algo, sha256_hash
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
```

## Metadata Management

### Upload Metadata
```go
input := &s3.PutObjectInput{
    Bucket:        aws.String(bucket),
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
```

### Retrieved Metadata
```go
metadata := map[string]string{
    "upload-timestamp": "2024-01-15T10:30:00Z",
    "content-type":     "application/octet-stream",
    "encryption":       "aes-256-gcm",
    "compression":      "gzip",
    "content-length":   "1024",
    "last-modified":    "2024-01-15T10:30:00Z",
}
```

## Error Handling

### Common Error Scenarios

#### 1. Missing Credentials
```go
if err != nil && strings.Contains(err.Error(), "incomplete R2 configuration") {
    // Handle missing environment variables
    log.Printf("R2 credentials not configured")
}
```

#### 2. Network Timeout
```go
if err != nil && strings.Contains(err.Error(), "context deadline exceeded") {
    // Handle timeout
    log.Printf("Upload timed out")
}
```

#### 3. Invalid Blob ID
```go
if err != nil && strings.Contains(err.Error(), "blobID cannot be empty") {
    // Handle validation error
    log.Printf("Invalid blob ID")
}
```

### Error Response Format
```json
{
    "error": "R2 upload failed",
    "details": "failed to upload email to R2: NoCredentialProviders"
}
```

## Security Features

### 1. **Path Isolation**
- Emails stored in dedicated `emails/` path
- Prevents access to other storage areas
- Clear separation of concerns

### 2. **Content Type Security**
- All uploads use `application/octet-stream`
- Prevents MIME type sniffing attacks
- Consistent handling across all uploads

### 3. **Metadata Security**
- Upload timestamps for audit trails
- Encryption type identification
- Compression algorithm tracking

### 4. **Context Support**
- Timeout protection against hanging requests
- Cancellation support for long-running operations
- Resource cleanup on context cancellation

## Performance Considerations

### 1. **Upload Optimization**
- Streaming uploads for large files
- Memory-efficient handling
- Context-based timeout management

### 2. **Retrieval Optimization**
- HEAD requests for existence checks
- Metadata-only retrieval
- Streaming downloads for large files

### 3. **Error Recovery**
- Automatic retry for transient errors
- Graceful degradation on failures
- Clear error messages for debugging

## Testing

### Unit Tests
- Configuration validation
- Input validation
- Error handling scenarios
- Mock client testing

### Integration Tests
- End-to-end upload flow
- Metadata retrieval
- Error condition testing
- Performance benchmarking

### Test Coverage
- ✅ Client creation and validation
- ✅ Upload functionality
- ✅ Download functionality
- ✅ Metadata operations
- ✅ Error handling
- ✅ Context support

## Monitoring and Logging

### Upload Metrics
```go
log.Printf("Upload started: %s, size: %d bytes", blobID, len(data))
log.Printf("Upload completed: %s, duration: %v", blobID, duration)
log.Printf("Upload failed: %s, error: %v", blobID, err)
```

### Performance Metrics
- Upload duration
- File size distribution
- Error rates
- Success rates

## Future Enhancements

### 1. **Advanced Features**
- Multipart uploads for large files
- Parallel upload/download
- Compression optimization
- Caching layer

### 2. **Security Enhancements**
- Client-side encryption
- Server-side encryption
- Access control lists
- Audit logging

### 3. **Monitoring**
- Real-time metrics
- Alerting on failures
- Performance dashboards
- Cost optimization

## Conclusion

The R2 storage implementation provides:
- **Security**: Proper path isolation and content type handling
- **Reliability**: Context support and error handling
- **Performance**: Streaming operations and timeout management
- **Flexibility**: Multiple upload methods and metadata support
- **Testability**: Comprehensive test coverage and mock support

This implementation meets the requirements for secure email storage while maintaining good performance and reliability. 