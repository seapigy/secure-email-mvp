# R2 Upload Test

This test verifies that the Cloudflare R2 upload functionality works correctly with real credentials.

## Setup

### 1. Environment Variables

Create a `.env` file in the project root with your R2 credentials:

```bash
# Required R2 Configuration
R2_ACCESS_KEY_ID=your_access_key_id
R2_SECRET_ACCESS_KEY=your_secret_access_key
R2_BUCKET=secure-email-blobs
R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
```

### 2. Manual Environment Variables (Alternative)

If you prefer to set environment variables manually:

**Windows PowerShell:**
```powershell
$env:R2_ACCESS_KEY_ID="your_access_key_id"
$env:R2_SECRET_ACCESS_KEY="your_secret_access_key"
$env:R2_BUCKET="secure-email-blobs"
$env:R2_ENDPOINT="https://your-account-id.r2.cloudflarestorage.com"
```

**Windows Command Prompt:**
```cmd
set R2_ACCESS_KEY_ID=your_access_key_id
set R2_SECRET_ACCESS_KEY=your_secret_access_key
set R2_BUCKET=secure-email-blobs
set R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
```

**Linux/macOS:**
```bash
export R2_ACCESS_KEY_ID=your_access_key_id
export R2_SECRET_ACCESS_KEY=your_secret_access_key
export R2_BUCKET=secure-email-blobs
export R2_ENDPOINT=https://your-account-id.r2.cloudflarestorage.com
```

## Running the Test

### Option 1: Run with .env file
```bash
go run test_r2_upload.go
```

### Option 2: Run with manual environment variables
```bash
R2_ACCESS_KEY_ID=your_key R2_SECRET_ACCESS_KEY=your_secret R2_BUCKET=your_bucket R2_ENDPOINT=your_endpoint go run test_r2_upload.go
```

## Expected Output

### Success Case:
```
2024/01/XX XX:XX:XX R2 Configuration:
2024/01/XX XX:XX:XX   Bucket: secure-email-blobs
2024/01/XX XX:XX:XX   Endpoint: https://your-account-id.r2.cloudflarestorage.com
2024/01/XX XX:XX:XX   Access Key ID: your_key...
2024/01/XX XX:XX:XX   Secret Key: [HIDDEN]
2024/01/XX XX:XX:XX Starting upload test:
2024/01/XX XX:XX:XX   Blob ID: test-upload-1703123456.blob
2024/01/XX XX:XX:XX   Data size: 20 bytes
2024/01/XX XX:XX:XX Uploading to R2...
2024/01/XX XX:XX:XX ✅ Upload successful!
2024/01/XX XX:XX:XX   Blob uploaded to: emails/test-upload-1703123456.blob
2024/01/XX XX:XX:XX   Bucket: secure-email-blobs
2024/01/XX XX:XX:XX Verifying upload...
2024/01/XX XX:XX:XX ✅ Verification successful: File exists in R2
2024/01/XX XX:XX:XX Cleaning up test file...
2024/01/XX XX:XX:XX ✅ Test file cleaned up successfully
```

### Error Case:
```
2024/01/XX XX:XX:XX Missing required environment variables: [R2_ACCESS_KEY_ID R2_SECRET_ACCESS_KEY R2_BUCKET R2_ENDPOINT]
```

## What the Test Does

1. **Environment Check**: Validates all required environment variables are set
2. **Configuration Display**: Shows R2 configuration (without sensitive data)
3. **Test Data Preparation**: Creates a test blob with timestamp-based ID
4. **Upload Test**: Attempts to upload the test blob to R2
5. **Verification**: Checks if the uploaded file exists in R2
6. **Cleanup**: Deletes the test file from R2

## Troubleshooting

### Common Issues:

1. **Missing Environment Variables**
   - Ensure all required variables are set
   - Check `.env` file exists and is readable

2. **Invalid Credentials**
   - Verify your R2 access key and secret are correct
   - Check that the credentials have proper permissions

3. **Bucket Issues**
   - Ensure the bucket exists in your R2 account
   - Verify bucket permissions allow upload/delete operations

4. **Network Issues**
   - Check internet connectivity
   - Verify firewall settings allow outbound HTTPS connections

5. **Endpoint Issues**
   - Ensure the endpoint URL is correct for your account
   - Format should be: `https://your-account-id.r2.cloudflarestorage.com`

## Security Notes

- The test file is automatically cleaned up after verification
- Sensitive credentials are not logged in the output
- The test uses a unique timestamp-based filename to avoid conflicts
- All operations use proper context timeouts

## Next Steps

After successful upload test:
1. Verify the R2 storage integration works in your main application
2. Test the complete email upload flow with encryption
3. Monitor R2 usage and costs in your Cloudflare dashboard 