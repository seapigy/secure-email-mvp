# GET /api/email/view/{id} Endpoint

## Overview

The `GET /api/email/view/{id}` endpoint allows authenticated users to retrieve and view the full contents of individual emails they sent. This endpoint provides secure access to decrypted email content with proper authorization controls.

## Security Features

- **JWT Authentication Required**: All requests must include a valid JWT token
- **User Authorization**: Users can only access emails they sent (sender_id == user_id)
- **Encrypted Storage**: Email content is encrypted at rest in Cloudflare R2
- **Integrity Verification**: Auth tag verification ensures content integrity
- **Access Tracking**: Logs access attempts and updates access counters

## API Specification

### Endpoint
```
GET /api/email/view/{id}
```

### Headers
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Path Parameters
- `{id}`: The unique email ID to retrieve

### Response Format

#### Success Response (200 OK)
```json
{
  "email_id": "ccaa1234-5678-9abc-def0-123456789abc",
  "recipient": "user@example.com",
  "subject": "Confidential Details",
  "body": "This is the decrypted body of the email.",
  "created_at": "2025-07-30T12:00:00Z",
  "status": "success"
}
```

#### Error Responses

**401 Unauthorized**
```json
{
  "error": "Authentication required"
}
```

**400 Bad Request**
```json
{
  "error": "Missing email_id"
}
```

**403 Forbidden**
```json
{
  "error": "Access denied"
}
```

**404 Not Found**
```json
{
  "error": "Email not found"
}
```

**500 Internal Server Error**
```json
{
  "error": "Database connection unavailable"
}
```

## Implementation Details

### Authentication Flow
1. Extract JWT token from `Authorization` header
2. Validate token using JWT secret
3. Extract `user_id` from token claims
4. Inject `user_id` into request context

### Authorization Flow
1. Retrieve email metadata from database using `email_id`
2. Compare `sender_id` from database with authenticated `user_id`
3. Return 403 Forbidden if user doesn't own the email

### Data Retrieval Flow
1. Query SQLite for email metadata (blob URL, encryption keys, etc.)
2. Download encrypted blob from Cloudflare R2
3. Decode base64-encoded encryption parameters
4. Verify auth tag integrity
5. Decrypt content using AES-256-GCM
6. Decompress content using gzip (if applicable)
7. Update access tracking in database

### Error Handling
- **Missing/Invalid JWT**: Returns 401 Unauthorized
- **Missing email_id**: Returns 400 Bad Request
- **Email not found**: Returns 404 Not Found
- **Unauthorized access**: Returns 403 Forbidden
- **Database errors**: Returns 500 Internal Server Error
- **R2 retrieval errors**: Returns 500 Internal Server Error
- **Decryption errors**: Returns 500 Internal Server Error
- **Decompression errors**: Returns 500 Internal Server Error

## Usage Examples

### cURL Example
```bash
curl -X GET "http://localhost:8080/api/email/view/ccaa1234-5678-9abc-def0-123456789abc" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

### JavaScript Example
```javascript
const response = await fetch('/api/email/view/ccaa1234-5678-9abc-def0-123456789abc', {
  method: 'GET',
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
    'Content-Type': 'application/json'
  }
});

if (response.ok) {
  const email = await response.json();
  console.log('Email subject:', email.subject);
  console.log('Email body:', email.body);
} else {
  const error = await response.json();
  console.error('Error:', error.error);
}
```

## Database Schema

The endpoint queries the `emails` table with the following fields:
- `email_id`: Primary key
- `sender_id`: User who sent the email
- `recipient`: Email recipient
- `subject`: Email subject
- `encrypted_blob_url`: R2 blob identifier
- `encrypted_key`: Base64-encoded encryption key
- `encryption_nonce`: Base64-encoded nonce
- `encryption_auth_tag`: Base64-encoded auth tag
- `compression_algo`: Compression algorithm used
- `created_at`: Email creation timestamp

## Security Considerations

1. **JWT Validation**: Tokens are validated using the `JWT_SECRET` environment variable
2. **User Isolation**: Users can only access their own emails via sender_id filtering
3. **Content Integrity**: Auth tag verification prevents tampering
4. **Access Logging**: All access attempts are logged for audit purposes
5. **Timeout Protection**: R2 operations have 30-second timeout
6. **Error Sanitization**: Internal errors are sanitized before returning to client

## Testing

The endpoint includes comprehensive test coverage:
- Authentication tests (missing/invalid JWT tokens)
- Authorization tests (access control)
- Error handling tests (database, R2, decryption errors)
- Response structure validation
- URL parameter validation

## Dependencies

- `github.com/gorilla/mux`: URL parameter extraction
- `crypto/aes` and `crypto/cipher`: AES-256-GCM decryption
- `compress/gzip`: Content decompression
- `encoding/base64`: Parameter decoding
- `secure-email-mvp/pkg/auth`: JWT validation and decryption
- `secure-email-mvp/pkg/storage`: R2 blob retrieval 