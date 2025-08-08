# Email Expiration Implementation

## Overview

This document describes the implementation of the **email expiration** functionality for the Secure Email MVP system. This feature allows senders to specify a time after which emails can no longer be accessed, providing time-bound confidentiality for sensitive communications.

## 🎯 Implementation Summary

**Micro-Iteration 4.7** has been **successfully completed** with the following components:

### ✅ Backend Implementation

#### 1. Database Schema
- **Field**: `expires_at DATETIME` in `emails` table (already existed)
- **Purpose**: Tracks when an email should expire and become inaccessible
- **Index**: `idx_emails_expires_at` for performance optimization

#### 2. Send Email Handler (`cmd/api/send_email_handler.go`)
- **Added**: `ExpiresAt` field to `SendEmailRequest` struct
- **Updated**: Database insert to include `expires_at` field
- **Validation**: ISO 8601 UTC format validation and future date checking
- **Conversion**: Proper handling of NULL values for optional expiration

#### 3. View Email Handler (`cmd/api/view_email_handler.go`)
- **Enhanced**: Database query to include `expires_at` field
- **Logic**: Check if current time is past expiration timestamp
- **Response**: Return 410 Gone for expired emails with cleanup
- **Deletion**: Remove from R2 storage and mark as expired in database
- **Logging**: Comprehensive audit logging for expiration events

### ✅ Frontend Integration

The frontend already supports expiration timestamps:
- **Type Definitions**: `expires` property in `SecureEmail` interface
- **UI Components**: Compose modal with expiration date picker
- **Visual Indicators**: Expiration status in email detail view
- **Filtering**: Email inbox filter for expired emails

### ✅ Testing Infrastructure

#### 1. Unit Tests (`cmd/api/view_email_test.go`)
- **Logic Tests**: Validate expiration decision logic
- **Format Tests**: ISO 8601 UTC format validation
- **Interaction Tests**: Expiration and burn-after-read interaction

#### 2. Integration Test Scripts
- **PowerShell**: `scripts/test_email_expiration.ps1` (Windows)
- **Coverage**: Complete end-to-end testing flow

## 🔧 Technical Implementation Details

### Database Schema

```sql
-- When the email auto-expires (e.g., 30 days)
expires_at DATETIME,

-- Index for performance
CREATE INDEX IF NOT EXISTS idx_emails_expires_at ON emails(expires_at);
```

### API Endpoints

#### Send Email (`POST /api/email/send`)
```json
{
  "recipient": "alice@example.com",
  "subject": "Time-Sensitive Information",
  "body": "This email will expire at the specified time.",
  "expiresAt": "2024-12-31T23:59:59Z"
}
```

#### View Email (`GET /api/email/view/{id}`)
**Valid Email (Success)**:
```json
{
  "email_id": "uuid-123",
  "recipient": "alice@example.com",
  "subject": "Time-Sensitive Information",
  "body": "This email will expire at the specified time.",
  "status": "success",
  "expiresAt": "2024-12-31T23:59:59Z",
  "isExpired": false
}
```

**Expired Email (410 Gone)**:
```json
{
  "error": "Email has expired and is no longer accessible",
  "status": "expired",
  "email_id": "uuid-123",
  "expired_at": "2024-12-31T23:59:59Z"
}
```

### Security Flow

1. **Email Creation**: User sends email with `expiresAt` timestamp
2. **Validation**: Server validates ISO 8601 UTC format and future date
3. **Storage**: Email stored normally with expiration metadata
4. **Access Check**: Server checks current time against expiration
5. **Expired Access**: Returns 410 Gone and deletes content
6. **Audit Trail**: All expiration events logged for security monitoring

### Expiration Process

When an expired email is accessed:

1. **Time Check**: Compare current time with `expires_at` timestamp
2. **R2 Storage**: Delete encrypted blob from Cloudflare R2
3. **Database**: Mark encryption fields as NULL (soft delete)
4. **Metadata**: Preserve email metadata for audit purposes
5. **Logging**: Record expiration event with timestamp and user ID

## 🧪 Testing

### Manual Testing

Run the PowerShell test script:
```powershell
.\scripts\test_email_expiration.ps1
```

### Expected Test Results

1. ✅ **API Server**: Running and accessible
2. ✅ **Authentication**: JWT token obtained successfully
3. ✅ **Expired Email**: Correctly returns 410 Gone
4. ✅ **Valid Email**: Accessible until expiration
5. ✅ **Invalid Format**: Rejected with 400 Bad Request
6. ✅ **Past Expiration**: Rejected during creation
7. ✅ **Cleanup**: Email properly deleted when expired

### Unit Tests

Run Go tests:
```bash
go test ./cmd/api -v
```

## 🔐 Security Considerations

### Access Control
- **Authentication**: JWT token required for all operations
- **Authorization**: Users can only access their own emails
- **Time Validation**: Expiration must be in the future during creation

### Data Protection
- **Encryption**: All email content encrypted with AES-256-GCM
- **Secure Deletion**: Complete removal from R2 storage on expiration
- **Audit Logging**: All expiration events recorded

### Privacy Features
- **Time-Bound Access**: Automatic deletion after expiration
- **Metadata Preservation**: Keep audit trail while removing content
- **Format Validation**: Strict ISO 8601 UTC format enforcement

## 📊 Performance Impact

### Database Operations
- **Minimal**: Single additional field in queries
- **Indexed**: `expires_at` field has dedicated index
- **Efficient**: Time comparison before expensive operations

### Storage Operations
- **R2 Deletion**: Asynchronous deletion from Cloudflare R2
- **Graceful Failure**: Continue response even if R2 deletion fails
- **Logging**: Non-blocking audit trail updates

### Network Impact
- **No Change**: Same API response format
- **Status Codes**: Proper HTTP 410 Gone for expired emails
- **Headers**: Standard REST API response headers

## 🚀 Deployment

### Backend Deployment
The implementation is ready for deployment to the Oracle Cloud VM:

1. **Build**: `go build -o api-server ./cmd/api`
2. **Deploy**: Upload to `/home/opc/secure-email-mvp/`
3. **Restart**: `sudo systemctl restart secure-email-api`

### Frontend Integration
The frontend already supports expiration timestamps:
- **No Changes Required**: UI components already implemented
- **Backward Compatible**: Existing emails continue to work
- **Feature Flag**: `expiresAt` field in API requests

## 📈 Monitoring

### Health Checks
- **Endpoint**: `/health` for overall system status
- **Database**: Connection and query performance
- **R2 Storage**: Upload/download success rates

### Security Monitoring
- **Access Logs**: All email access attempts
- **Expiration Events**: Email expiration and deletion events
- **Error Tracking**: Failed operations and security violations

### Metrics
- **Expiration Usage**: Number of emails with expiration set
- **Deletion Success Rate**: Percentage of successful R2 deletions
- **Access Patterns**: Frequency of email access attempts

## 🔮 Future Enhancements

### Planned Features
1. **Bulk Expiration**: Apply expiration to multiple emails
2. **Relative Expiration**: Set expiration relative to creation time
3. **Advanced Analytics**: Detailed expiration statistics and patterns
4. **Admin Interface**: Management tools for expired emails

### Potential Improvements
1. **Caching**: Optimize database queries for expiration checks
2. **Batch Cleanup**: Efficient cleanup of expired emails
3. **Notification System**: Alert users when emails are about to expire
4. **Recovery Options**: Limited-time recovery of expired emails

## ⚡ Integration with Burn-After-Read

### Priority Order
1. **Expiration Check**: First priority - check if email has expired
2. **Burn-After-Read Check**: Second priority - check if already consumed
3. **Content Retrieval**: Final step - serve email content

### Combined Scenarios
- **Expired + Burn-After-Read**: Expiration takes precedence
- **Valid + Burn-After-Read**: Normal burn-after-read behavior
- **Expired + Normal**: Expiration cleanup and 410 Gone
- **Valid + Normal**: Standard email access

## ✅ Completion Status

**Micro-Iteration 4.7: Email Expiration Implementation** is **100% Complete**:

- ✅ **Backend Logic**: Full implementation in view handler
- ✅ **Database Integration**: Schema and queries updated
- ✅ **API Endpoints**: Send and view handlers enhanced
- ✅ **Validation**: ISO 8601 UTC format and future date validation
- ✅ **Error Handling**: Proper HTTP status codes and responses
- ✅ **Security**: Comprehensive access control and logging
- ✅ **Testing**: Unit tests and integration test scripts
- ✅ **Documentation**: Complete implementation guide
- ✅ **Frontend Support**: Already implemented and tested
- ✅ **Integration**: Proper interaction with burn-after-read feature

The email expiration functionality is **production-ready** and provides time-bound confidentiality for sensitive email communications.

