# Security & Logging Standards

## Zero Visibility Compliance

This document outlines the security and logging standards for the Secure Email MVP, with a focus on **Zero Visibility** compliance to ensure no Personally Identifiable Information (PII) is ever exposed in logs, error messages, or system outputs.

---

## 🔒 Zero Visibility Principles

### Core Rules

1. **No PII in Logs**: Never log email addresses, user IDs, tokens, or any identifying information
2. **Generic Error Messages**: All error responses must be user-friendly and generic
3. **Structured Logging**: Use structured JSON logging with sanitized fields
4. **User Isolation**: Strict enforcement that users cannot access each other's data
5. **Secure Data Handling**: All sensitive data must be properly sanitized before logging

### What Constitutes PII

- **Email addresses** (e.g., `user@example.com`)
- **User IDs** (unless they are UUIDs in specific contexts)
- **Authentication tokens** (JWT, refresh tokens, etc.)
- **Passwords** (even hashed ones)
- **Personal names** or identifiers
- **IP addresses** (in some contexts)
- **Session IDs** or cookies

### What is Safe to Log

- **UUIDs** (when they are not tied to PII)
- **Operation status** (success/failure)
- **HTTP status codes**
- **Request methods** and endpoints
- **Timestamps**
- **Generic error types** (not specific error messages)
- **Performance metrics** (without PII context)

---

## 📝 Structured Logging Standards

### Log Entry Format

All logs must follow this structured JSON format:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "event": "InboxList",
  "component": "inbox_api",
  "status": 200,
  "duration_ms": 150,
  "user_id": "anonymous",
  "ip_address": "192.168.1.1",
  "user_agent": "Mozilla/5.0...",
  "endpoint": "/api/inbox/list",
  "method": "GET",
  "error": "",
  "details": {
    "email_count": 5,
    "operation": "inbox_request"
  },
  "message": "Inbox list operation completed"
}
```

### Log Levels

- **DEBUG**: Development-only information
- **INFO**: General operational information
- **WARNING**: Potential issues that don't affect functionality
- **ERROR**: Errors that affect functionality but are recoverable
- **CRITICAL**: Critical errors that may affect system stability

### Sanitization Rules

#### User ID Sanitization
```go
// ✅ Safe - UUID format
"user_id": "550e8400-e29b-41d4-a716-446655440000"

// ✅ Safe - Anonymous for non-UUID
"user_id": "anonymous"

// ❌ Forbidden - Email address
"user_id": "user@example.com"

// ❌ Forbidden - Plain text ID
"user_id": "user123"
```

#### IP Address Sanitization
```go
// ✅ Safe - Masked IP
"ip_address": "192.168.1.0"

// ✅ Safe - Unknown for empty
"ip_address": "unknown"

// ❌ Forbidden - Full IP with sensitive parts
"ip_address": "192.168.1.100"
```

#### User Agent Sanitization
```go
// ✅ Safe - Truncated
"user_agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36..."

// ❌ Forbidden - Contains PII
"user_agent": "CustomApp/1.0 (user@example.com)"
```

---

## 🛡️ Security Implementation

### Backend Security

#### Authentication Enforcement
```go
// ✅ Correct - Check authentication first
func (srv *Server) listInboxHandler(w http.ResponseWriter, r *http.Request) {
    userID, ok := GetUserIDFromContext(r)
    if !ok {
        LogSecurityEvent(r, "unauthorized_access", "Authentication required", LogLevelWarning, "", nil)
        w.WriteHeader(http.StatusUnauthorized)
        w.Write([]byte(`{"error":"Authentication required"}`))
        return
    }
    // ... rest of handler
}
```

#### User Isolation
```go
// ✅ Correct - Always filter by user_id
rows, err := srv.db.Query(`
    SELECT e.email_id, e.sender_id, e.subject, im.created_at, ...
    FROM inbox_messages im
    INNER JOIN emails e ON im.email_id = e.email_id
    WHERE im.user_id = ? AND im.is_deleted = 0
    ORDER BY im.created_at DESC`,
    userID, // Always use authenticated user's ID
)
```

#### Generic Error Messages
```go
// ✅ Correct - Generic error messages
w.Write([]byte(`{"error":"Service temporarily unavailable"}`))
w.Write([]byte(`{"error":"Email not found"}`))
w.Write([]byte(`{"error":"Authentication required"}`))

// ❌ Forbidden - Specific error messages
w.Write([]byte(`{"error":"Database connection failed for user@example.com"}`))
w.Write([]byte(`{"error":"SQL error: no such table users"}`))
```

### Frontend Security

#### Error Handling
```typescript
// ✅ Correct - Generic error handling
export const handleInboxError = (error: unknown): string => {
  if (typeof error === 'string') {
    return error;
  }
  if (error instanceof Error) {
    return error.message;
  }
  return 'Failed to load inbox. Please try again.';
};

// ❌ Forbidden - Exposing internal errors
export const handleInboxError = (error: unknown): string => {
  if (error instanceof Error) {
    return `Database error: ${error.message}`; // Exposes internal details
  }
  return 'Error occurred';
};
```

#### Data Transformation
```typescript
// ✅ Correct - Safe data transformation
export const transformInboxEmailToSecureEmail = (inboxItem: InboxEmailItem): SecureEmail => {
  return {
    id: inboxItem.email_id,
    from: `user-${inboxItem.sender_id}@securesystem.email`, // Safe format
    to: 'current-user@securesystem.email', // Generic recipient
    // ... other fields
  };
};

// ❌ Forbidden - Exposing raw PII
export const transformInboxEmailToSecureEmail = (inboxItem: InboxEmailItem): SecureEmail => {
  return {
    id: inboxItem.email_id,
    from: inboxItem.sender_email, // Raw email address
    to: inboxItem.recipient_email, // Raw email address
    // ... other fields
  };
};
```

---

## 🧪 Testing Standards

### Backend Security Tests

#### User Isolation Tests
```go
func testUserIsolation(t *testing.T, srv *Server, user1ID, user2ID, email1ID, email2ID string) {
    // Test that user1 cannot access user2's email
    req := httptest.NewRequest("GET", "/api/inbox/"+email2ID, nil)
    req = setUserContext(req, user1ID)
    
    w := httptest.NewRecorder()
    srv.getInboxEmailHandler(w, req)
    
    if w.Code != http.StatusNotFound {
        t.Errorf("Expected status 404, got %d", w.Code)
    }
}
```

#### Zero Visibility Log Tests
```go
func testZeroVisibilityLogging(t *testing.T, srv *Server, userID, emailID string) {
    // Capture log output
    var logBuffer bytes.Buffer
    log.SetOutput(&logBuffer)
    defer log.SetOutput(os.Stderr)
    
    // Make request
    req := httptest.NewRequest("GET", "/api/inbox/"+emailID, nil)
    req = setUserContext(req, userID)
    srv.getInboxEmailHandler(w, req)
    
    // Check for PII patterns
    logOutput := logBuffer.String()
    piiPatterns := []string{"@", "password", "token", "secret"}
    
    for _, pattern := range piiPatterns {
        if strings.Contains(logOutput, pattern) {
            t.Errorf("PII pattern '%s' found in logs", pattern)
        }
    }
}
```

### Frontend Security Tests

#### Console Output Tests
```typescript
it('should not log PII in console output', async () => {
  // Mock console methods
  const consoleOutput: string[] = [];
  console.log = vi.fn((...args) => {
    consoleOutput.push(args.join(' '));
  });
  
  // Render component
  render(<EmailInbox />);
  
  // Check for PII patterns
  const consoleText = consoleOutput.join(' ');
  const piiPatterns = ['@', 'password', 'token', 'secret'];
  
  piiPatterns.forEach(pattern => {
    expect(consoleText).not.toContain(pattern);
  });
});
```

#### Error Message Tests
```typescript
it('should display generic error messages', async () => {
  // Mock API error
  const mockApi = {
    get: vi.fn().mockRejectedValue(new Error('Database connection failed'))
  };
  
  render(<EmailInbox />);
  
  await waitFor(() => {
    expect(screen.getByText(/Failed to load inbox/)).toBeInTheDocument();
  });
  
  // Check that error is generic
  const errorMessage = screen.getByText(/Failed to load inbox/);
  expect(errorMessage.textContent).not.toContain('Database');
  expect(errorMessage.textContent).not.toContain('connection');
});
```

---

## 🔍 Log Scanning Utility

### Automated PII Detection

The test suite includes automated scanning for PII patterns:

```go
// PII patterns to detect
var piiPatterns = []string{
    "@",                    // Email addresses
    "password",             // Password references
    "token",                // Token references
    "secret",               // Secret references
    "user@",                // User email patterns
    "admin@",               // Admin email patterns
    "test@",                // Test email patterns
}

// Scan logs for PII
func scanLogsForPII(logOutput string) []string {
    var foundPII []string
    for _, pattern := range piiPatterns {
        if strings.Contains(logOutput, pattern) {
            foundPII = append(foundPII, pattern)
        }
    }
    return foundPII
}
```

### Test Integration

```go
func TestZeroVisibilityCompliance(t *testing.T) {
    // Capture logs during test
    var logBuffer bytes.Buffer
    log.SetOutput(&logBuffer)
    defer log.SetOutput(os.Stderr)
    
    // Run inbox operations
    runInboxOperations()
    
    // Scan for PII
    logOutput := logBuffer.String()
    foundPII := scanLogsForPII(logOutput)
    
    if len(foundPII) > 0 {
        t.Errorf("PII patterns found in logs: %v", foundPII)
    }
}
```

---

## 📋 Compliance Checklist

### Backend Compliance

- [ ] All endpoints require authentication
- [ ] User isolation enforced in all queries
- [ ] Generic error messages only
- [ ] Structured logging implemented
- [ ] No PII in log output
- [ ] User IDs sanitized (UUID or "anonymous")
- [ ] IP addresses masked
- [ ] User agents truncated
- [ ] Security events logged
- [ ] Database operations logged safely

### Frontend Compliance

- [ ] No PII in console output
- [ ] Generic error messages displayed
- [ ] No raw email addresses in UI
- [ ] Safe data transformation
- [ ] User isolation maintained
- [ ] Authentication errors handled gracefully
- [ ] Network errors handled safely
- [ ] No stack traces exposed

### Testing Compliance

- [ ] User isolation tests pass
- [ ] Zero visibility log tests pass
- [ ] Generic error message tests pass
- [ ] Authentication enforcement tests pass
- [ ] PII scanning tests pass
- [ ] Console output tests pass
- [ ] Error handling tests pass

---

## 🚨 Security Violations

### Critical Violations

1. **PII in Logs**: Any email address, user ID, or token appearing in logs
2. **Cross-User Access**: Users able to access other users' data
3. **Specific Error Messages**: Technical details exposed to users
4. **Missing Authentication**: Endpoints accessible without auth
5. **Raw PII in UI**: Email addresses or tokens displayed to users

### Response to Violations

1. **Immediate Fix**: Address the violation immediately
2. **Log Review**: Scan all logs for similar violations
3. **Test Enhancement**: Add tests to prevent recurrence
4. **Documentation Update**: Update this document if needed
5. **Team Notification**: Inform team of the violation and fix

---

## 📚 References

- [OWASP Logging Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html)
- [GDPR Article 32](https://gdpr-info.eu/art-32-gdpr/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [Zero Trust Security Model](https://www.nist.gov/publications/zero-trust-architecture)

---

## 📞 Contact

For questions about security and logging standards:

- **Security Team**: security@securesystem.email
- **DevOps Team**: devops@securesystem.email
- **Emergency**: security-emergency@securesystem.email

**Last Updated**: August 30, 2025  
**Version**: 1.0  
**Status**: Active
