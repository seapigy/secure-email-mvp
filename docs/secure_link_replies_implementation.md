# Secure Link Replies Implementation

## Overview

This document describes the implementation of Iteration 3 - External Reply Flow for the Secure Email MVP. This feature enables external recipients to reply to secure emails through a secure web interface, with replies being forwarded to the original sender while maintaining security and audit logging.

## Architecture

### Backend Components

#### 1. Public Reply Handlers (`cmd/api/public_reply_handlers.go`)

**Purpose**: Handle external reply requests and process them securely.

**Key Functions**:
- `replyHandler`: Main endpoint handler for POST `/v/{linkID}/reply`
- `processExternalReply`: Core logic for processing external replies
- `getOriginalEmailForReply`: Retrieve original email details for reply context
- `validateLinkForReply`: Validate that a link is still valid for replies
- `storeSecureReply`: Store reply data in the database
- `forwardReplyToSender`: Forward reply to the original sender via SES
- `updateReplyStatus`: Update reply status after processing
- `logSESTransaction`: Log SES transactions for audit purposes

**Data Structures**:
```go
type ReplyRequest struct {
    LinkID      string `json:"link_id"`
    Subject     string `json:"subject"`
    Body        string `json:"body"`
    IPAddress   string `json:"ip_address,omitempty"`
    UserAgent   string `json:"user_agent,omitempty"`
    AccessToken string `json:"access_token,omitempty"`
}

type ReplyResponse struct {
    Success       bool   `json:"success"`
    ReplyID       string `json:"reply_id,omitempty"`
    Message       string `json:"message,omitempty"`
    Error         string `json:"error,omitempty"`
    ErrorCode     string `json:"error_code,omitempty"`
    TransactionID string `json:"transaction_id,omitempty"`
}

type SecureReplyData struct {
    ReplyID       string    `json:"reply_id"`
    LinkID        string    `json:"link_id"`
    EmailChainID  string    `json:"email_chain_id"`
    ParentEmailID string    `json:"parent_email_id"`
    Subject       string    `json:"subject"`
    Body          string    `json:"body"`
    SenderEmail   string    `json:"sender_email"`
    SenderName    string    `json:"sender_name,omitempty"`
    IPAddress     string    `json:"ip_address"`
    UserAgent     string    `json:"user_agent"`
    CreatedAt     time.Time `json:"created_at"`
    Status        string    `json:"status"`
}
```

#### 2. API Endpoints

**POST `/v/{linkID}/reply`**
- **Purpose**: Accept external replies to secure links
- **Authentication**: None (public endpoint)
- **Request Body**: `ReplyRequest`
- **Response**: `ReplyResponse`
- **Security**: Validates link status, expiration, revocation, and access attempts

### Frontend Components

#### 1. ReplyComposer (`src/components/external/ReplyComposer.tsx`)

**Purpose**: Provide a user-friendly interface for external recipients to compose and send replies.

**Features**:
- Simple, intuitive reply form
- Subject line with automatic "Re:" prefix
- Rich text area for message body
- Real-time validation
- Loading states and error handling
- Success confirmation
- Keyboard shortcuts (Ctrl+Enter to send)

**Props**:
```typescript
interface ReplyComposerProps {
  linkID: string;
  originalSubject: string;
  originalSenderEmail: string;
  originalSenderName?: string;
  onReplySent?: (replyID: string) => void;
  onCancel?: () => void;
  isOpen: boolean;
}
```

**State Management**:
- Form data (subject, body)
- Loading states
- Error handling
- Success confirmation
- Reply ID tracking

#### 2. SecureEmailViewer Integration

**Purpose**: Integrate reply functionality into the existing secure email viewer.

**Features**:
- Reply button in the email content view
- Modal integration for reply composer
- Success handling and UI updates
- Proper state management

### Database Schema

#### 1. Secure Replies Table

```sql
CREATE TABLE secure_replies (
    reply_id VARCHAR(255) PRIMARY KEY,
    link_id VARCHAR(255) NOT NULL,
    email_chain_id VARCHAR(255) NOT NULL,
    parent_email_id VARCHAR(255) NOT NULL,
    subject TEXT NOT NULL,
    body TEXT NOT NULL,
    sender_email VARCHAR(255) NOT NULL,
    sender_name VARCHAR(255),
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    status VARCHAR(50) DEFAULT 'pending',
    FOREIGN KEY (link_id) REFERENCES secure_links(link_id)
);
```

#### 2. SES Transactions Table (Extended)

The existing `ses_transactions` table has been extended to support reply tracking:

```sql
ALTER TABLE ses_transactions ADD COLUMN reply_id VARCHAR(255);
```

## Security Features

### 1. Link Validation

**Expiration Check**: Replies are blocked if the secure link has expired.
```go
if expiresAt != nil && *expiresAt < time.Now().Unix() {
    return fmt.Errorf("secure link has expired")
}
```

**Revocation Check**: Replies are blocked if the secure link has been revoked.
```go
if revokedAt != "" {
    return fmt.Errorf("secure link has been revoked")
}
```

**Access Attempts**: Replies are blocked if the link has exceeded maximum access attempts.
```go
if currentAttempts >= maxAccessAttempts {
    return fmt.Errorf("secure link has been destroyed")
}
```

**Read-Once Protection**: Replies are blocked if a read-once link has been read.
```go
if readOnce {
    var readCount int
    err := srv.db.QueryRowContext(ctx, "SELECT read_count FROM secure_links WHERE link_id = ?", linkID).Scan(&readCount)
    if err == nil && readCount > 0 {
        return fmt.Errorf("secure link has been read and cannot be replied to")
    }
}
```

### 2. Audit Logging

**Reply Attempts**: All reply attempts are logged with:
- Link ID
- IP address
- User agent
- Success/failure status
- SES transaction ID (if successful)

**SES Transaction Tracking**: All reply emails are tracked in the SES transactions table with:
- Transaction ID
- Message ID
- Reply ID
- Recipient (original sender)
- Status and timestamp

### 3. Input Validation

**Content Validation**: Reply content is validated for:
- Required fields (subject, body)
- Content length limits
- XSS prevention
- Malicious content detection

**Rate Limiting**: Reply attempts are rate-limited to prevent abuse.

## Email Forwarding

### 1. Reply Email Format

Replies are formatted as professional HTML emails with:

**Header Section**:
- Secure reply notification
- Sender information
- Timestamp

**Content Section**:
- Original subject (with "Re:" prefix)
- Reply message body
- Proper formatting and styling

**Footer Section**:
- Reply details (sender, date, IP, link ID)
- Security notice
- System branding

### 2. SES Integration

**Email Sending**: Replies are sent via Amazon SES using the existing SES handler.

**Transaction Tracking**: Each reply email generates a unique SES transaction ID for tracking.

**Error Handling**: Failed email sends are logged but don't prevent reply storage.

## Testing

### 1. Integration Tests (`tests/test_secure_link_replies.ps1`)

**Test Scenarios**:
- Valid reply acceptance and forwarding
- Expired link reply rejection
- Revoked link reply rejection
- Multiple replies handling
- Audit logging verification
- SES transaction logging verification

**Test Functions**:
- `Test-APIHealth`: Verify API availability
- `Test-Login`: Authenticate test user
- `Test-SendSecureLinkEmail`: Create test secure link
- `Test-ValidReply`: Test successful reply
- `Test-ReplyToExpiredLink`: Test expired link rejection
- `Test-ReplyToRevokedLink`: Test revoked link rejection
- `Test-MultipleReplies`: Test multiple replies to same link
- `Test-ReplyAuditLogging`: Verify audit log entries
- `Test-SESTransactionLogging`: Verify SES transaction logging

### 2. Frontend Testing

**Component Testing**:
- ReplyComposer form validation
- Error handling and display
- Success state management
- Modal integration
- Keyboard shortcuts

**Integration Testing**:
- End-to-end reply flow
- Security validation integration
- UI state management
- Error recovery

## Configuration

### 1. Environment Variables

```bash
# SES Configuration (existing)
AWS_SES_SMTP_HOST=smtp.email.us-east-1.amazonaws.com
AWS_SES_SMTP_PORT=587
AWS_SES_SMTP_USERNAME=your-ses-username
AWS_SES_SMTP_PASSWORD=your-ses-password
AWS_SES_REGION=us-east-1

# Reply Configuration
REPLY_RATE_LIMIT=10
REPLY_MAX_LENGTH=10000
REPLY_SYSTEM_SENDER=noreply@securemail.com
```

### 2. Database Configuration

**Required Tables**:
- `secure_replies` (new)
- `ses_transactions` (extended)
- `secure_links` (existing)
- `emails` (existing)
- `users` (existing)

## Performance Considerations

### 1. Database Optimization

**Indexes**:
```sql
CREATE INDEX idx_secure_replies_link_id ON secure_replies(link_id);
CREATE INDEX idx_secure_replies_created_at ON secure_replies(created_at);
CREATE INDEX idx_ses_transactions_reply_id ON ses_transactions(reply_id);
```

**Query Optimization**:
- Efficient joins for reply processing
- Batch operations for audit logging
- Connection pooling for high concurrency

### 2. Caching Strategy

**Reply Validation Caching**:
- Cache link validation results
- TTL-based cache invalidation
- Memory-efficient caching

### 3. Rate Limiting

**Reply Rate Limits**:
- Per-link rate limiting
- Per-IP rate limiting
- Global rate limiting
- Exponential backoff for violations

## Security Considerations

### 1. Input Sanitization

**Content Sanitization**:
- HTML entity encoding
- Script tag removal
- XSS prevention
- Content length validation

### 2. Access Control

**Link Validation**:
- Comprehensive link status checking
- Expiration enforcement
- Revocation enforcement
- Access attempt tracking

### 3. Audit Trail

**Comprehensive Logging**:
- All reply attempts logged
- IP address tracking
- User agent tracking
- Success/failure status
- SES transaction correlation

## Error Handling

### 1. Backend Error Handling

**Validation Errors**:
- Link not found: 404
- Link expired: 400 with specific error code
- Link revoked: 400 with specific error code
- Link destroyed: 400 with specific error code
- Invalid input: 400 with validation details

**Processing Errors**:
- Database errors: 500 with logging
- SES errors: 500 with fallback handling
- System errors: 500 with generic message

### 2. Frontend Error Handling

**User-Friendly Messages**:
- Clear error descriptions
- Actionable error messages
- Retry mechanisms
- Graceful degradation

## Monitoring and Alerting

### 1. Metrics

**Reply Metrics**:
- Reply success rate
- Reply volume by time period
- Error rates by type
- Response times

**Security Metrics**:
- Failed reply attempts
- Suspicious activity patterns
- Rate limit violations
- Audit log volume

### 2. Alerts

**Critical Alerts**:
- High error rates
- System failures
- Security violations
- Performance degradation

**Informational Alerts**:
- High reply volume
- New patterns detected
- System health status

## Future Enhancements

### 1. Advanced Features

**Rich Text Support**:
- Markdown support
- Image embedding
- File attachments
- Formatting options

**Reply Threading**:
- Visual thread display
- Reply history
- Context preservation
- Thread management

### 2. Security Enhancements

**Advanced Validation**:
- Content analysis
- Spam detection
- Malware scanning
- Reputation scoring

**Enhanced Audit**:
- Behavioral analysis
- Anomaly detection
- Risk scoring
- Automated response

### 3. Performance Improvements

**Scalability**:
- Horizontal scaling
- Load balancing
- Database sharding
- CDN integration

**Optimization**:
- Query optimization
- Caching strategies
- Async processing
- Resource management

## Conclusion

The Secure Link Replies implementation provides a complete, secure, and user-friendly solution for external recipients to reply to secure emails. The system maintains security standards while providing a seamless user experience, comprehensive audit logging, and robust error handling.

Key achievements:
- ✅ Secure reply endpoints with comprehensive validation
- ✅ User-friendly frontend interface
- ✅ Complete audit logging and SES transaction tracking
- ✅ Comprehensive error handling and security measures
- ✅ Full integration testing and validation
- ✅ Performance optimization and monitoring capabilities

The implementation is production-ready and provides a solid foundation for future enhancements and scaling.
