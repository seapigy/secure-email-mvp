# DELETE /api/email/{id} Endpoint

## Overview

The `DELETE /api/email/{id}` endpoint allows authenticated users to securely delete emails they sent. This endpoint performs a complete cleanup by removing both the encrypted content from Cloudflare R2 and the metadata from the SQLite database.

## Security Features

- **JWT Authentication Required**: All requests must include a valid JWT token
- **User Authorization**: Users can only delete emails they sent (sender_id == user_id)
- **Complete Cleanup**: Removes encrypted content from R2 and metadata from database
- **Access Logging**: Logs all deletion attempts for audit purposes
- **Atomic Operations**: Ensures both R2 and database operations succeed or fail together

## API Specification

### Endpoint
```
DELETE /api/email/{id}
```

### Headers
```
Authorization: Bearer <jwt_token>
Content-Type: application/json
```

### Path Parameters
- `{id}`: The unique email ID to delete

### Response Format

#### Success Response (200 OK)
```json
{
  "status": "deleted",
  "email_id": "ccaa1234-5678-9abc-def0-123456789abc"
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

### Deletion Flow
1. Query SQLite for email metadata (blob URL, sender_id, etc.)
2. Verify user ownership of the email
3. Delete encrypted blob from Cloudflare R2 using `DeleteBlob()`
4. Remove email record from SQLite database
5. Return success response with confirmation

### Error Handling
- **Missing/Invalid JWT**: Returns 401 Unauthorized
- **Missing email_id**: Returns 400 Bad Request
- **Email not found**: Returns 404 Not Found
- **Unauthorized access**: Returns 403 Forbidden
- **Database errors**: Returns 500 Internal Server Error
- **R2 deletion errors**: Returns 500 Internal Server Error

## Usage Examples

### cURL Example
```bash
curl -X DELETE "http://localhost:8080/api/email/ccaa1234-5678-9abc-def0-123456789abc" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json"
```

### JavaScript Example
```javascript
const response = await fetch('/api/email/ccaa1234-5678-9abc-def0-123456789abc', {
  method: 'DELETE',
  headers: {
    'Authorization': `Bearer ${jwtToken}`,
    'Content-Type': 'application/json'
  }
});

if (response.ok) {
  const result = await response.json();
  console.log('Email deleted:', result.email_id);
} else {
  const error = await response.json();
  console.error('Error:', error.error);
}
```

## Database Schema

The endpoint queries the `emails` table with the following fields:
- `email_id`: Primary key
- `sender_id`: User who sent the email
- `encrypted_blob_url`: R2 blob identifier
- `created_at`: Email creation timestamp

## Security Considerations

1. **JWT Validation**: Tokens are validated using the `JWT_SECRET` environment variable
2. **User Isolation**: Users can only delete their own emails via sender_id filtering
3. **Complete Cleanup**: Both R2 content and database metadata are removed
4. **Access Logging**: All deletion attempts are logged for audit purposes
5. **Timeout Protection**: R2 operations have 30-second timeout
6. **Error Sanitization**: Internal errors are sanitized before returning to client

## Testing

The endpoint includes comprehensive test coverage:
- Authentication tests (missing/invalid JWT tokens)
- Authorization tests (access control)
- Error handling tests (database, R2, deletion errors)
- Response structure validation
- URL parameter validation

## Dependencies

- `github.com/gorilla/mux`: URL parameter extraction
- `secure-email-mvp/pkg/storage`: R2 blob deletion
- `context`: Timeout and cancellation support

## Cleanup Process

1. **R2 Deletion**: Removes encrypted blob from Cloudflare R2 storage
2. **Database Cleanup**: Removes email metadata from SQLite database
3. **Logging**: Records successful deletion with user and email ID
4. **Error Handling**: If R2 deletion fails, database deletion is not attempted

## Performance Considerations

- **Timeout**: 30-second timeout for R2 operations
- **Atomicity**: Database deletion only occurs after successful R2 deletion
- **Error Recovery**: Failed R2 deletions prevent database cleanup to maintain consistency 