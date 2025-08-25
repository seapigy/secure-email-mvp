# Password Validation & Breach Check Service

## Overview

The Secure Email MVP includes a comprehensive password validation service that enforces strong password requirements and checks against known data breaches using the HaveIBeenPwned (HIBP) API. This service is integrated into both signup and password reset flows to ensure all user passwords meet security standards.

## Features

### Password Strength Validation
- **Minimum Length**: 12 characters
- **Character Requirements**:
  - At least one uppercase letter (A-Z)
  - At least one lowercase letter (a-z)
  - At least one number (0-9)
  - At least one special character (!@#$%^&*()_+-=[]{}|;':",./<>?)
- **Common Password Blacklist**: Blocks passwords found in common password lists
- **Strength Scoring**: Provides a 0-100 score based on complexity and entropy

### Breach Detection
- **HaveIBeenPwned Integration**: Uses the Pwned Passwords API
- **K-Anonymity**: Protects user privacy by only sending password hash prefixes
- **Real-time Checking**: Validates passwords against known data breaches
- **Graceful Fallback**: Continues operation if API is unavailable

### User Experience
- **Generic Error Messages**: Prevents information disclosure about specific requirements
- **Improvement Suggestions**: Provides helpful guidance for password creation
- **Audit Logging**: Records validation attempts and failures for security monitoring

## Configuration

### Environment Variables

```bash
# HaveIBeenPwned API key (optional, increases rate limits)
HIBP_API_KEY=your_hibp_api_key_here
```

### Default Settings

```go
type PasswordConfig struct {
    MinLength:           12,
    RequireUppercase:    true,
    RequireLowercase:    true,
    RequireNumbers:      true,
    RequireSpecialChars: true,
    HIBPTimeout:         10 * time.Second,
    UserAgent:           "SecureEmail-MVP/1.0",
}
```

## API Integration

### Signup Flow

The password validation service is integrated into the signup handler (`cmd/api/signup_handler.go`):

```go
// Validate password using comprehensive password service
passwordService := password.NewPasswordService()
passwordResult, err := passwordService.ValidatePassword(ctx, req.Password)
if err != nil {
    log.Printf("Password validation failed: %v", err)
    http.Error(w, `{"error":"Password validation error"}`, http.StatusInternalServerError)
    return
}

if !passwordResult.IsValid {
    // Log validation failures for audit
    log.Printf("Password validation failed for email %s: %v", req.Email, passwordResult.Errors)
    
    // Return generic error message without revealing specific validation details
    http.Error(w, `{"error":"Password does not meet security requirements"}`, http.StatusBadRequest)
    return
}
```

### Password Reset Flow

The service is also integrated into password reset functionality (`cmd/api/password_reset_handler.go`):

```go
// Validate new password using comprehensive password service
passwordService := password.NewPasswordService()
ctx := context.Background()

passwordResult, err := passwordService.ValidatePassword(ctx, req.NewPassword)
if err != nil {
    log.Printf("Password validation failed during reset for email %s: %v", req.Email, err)
    http.Error(w, `{"error":"Password validation error"}`, http.StatusInternalServerError)
    return
}

if !passwordResult.IsValid {
    // Log validation failures for audit
    log.Printf("Password reset validation failed for email %s: %v", req.Email, passwordResult.Errors)
    
    // Return generic error message without revealing specific validation details
    http.Error(w, `{"error":"New password does not meet security requirements"}`, http.StatusBadRequest)
    return
}
```

## Security Implementation

### K-Anonymity for Breach Checking

The service uses k-anonymity to protect user privacy when checking passwords against the HaveIBeenPwned database:

1. **Password Hashing**: Creates SHA1 hash of the password
2. **Prefix Extraction**: Sends only the first 5 characters of the hash to HIBP
3. **Local Matching**: Receives a list of hash suffixes and matches locally
4. **Privacy Protection**: HIBP never receives the full password or hash

```go
// Create SHA1 hash of password
hash := sha1.Sum([]byte(password))
hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

// Use k-anonymity: send only first 5 characters of hash
prefix := hashHex[:5]
suffix := hashHex[5:]

// Make request to HIBP API
url := fmt.Sprintf("https://api.pwnedpasswords.com/range/%s", prefix)
```

### Common Password Detection

The service maintains a comprehensive blacklist of common passwords including:
- Simple patterns (123456, qwerty, password)
- Common names and words
- Keyboard patterns
- Default passwords

### Error Handling

The service implements robust error handling to maintain service availability:

- **API Failures**: Gracefully continues without breach checking
- **Network Timeouts**: Configurable timeout with fallback
- **Invalid Responses**: Logs errors but doesn't block signup
- **Rate Limiting**: Respects API rate limits

## Testing

### Unit Tests

Comprehensive unit tests cover all validation scenarios:

```bash
go test ./pkg/password -v
```

Test coverage includes:
- Password strength validation
- Character requirement checking
- Common password detection
- Breach checking (with and without API key)
- Error handling and fallback behavior

### Integration Tests

PowerShell integration tests verify end-to-end functionality:

```powershell
.\scripts\test_password_validation.ps1
```

Test scenarios include:
- Weak password rejection
- Common password rejection
- Missing requirement validation
- Strong password acceptance
- Password reset validation

## Monitoring and Logging

### Audit Logs

The service logs all validation attempts for security monitoring:

```
2025/08/12 12:48:41 Password validation failed for email user@example.com: [Password must be at least 12 characters long, Password must contain at least one uppercase letter]
2025/08/12 12:48:42 Password validation passed for email user@example.com (score: 75, breach count: 0)
```

### Metrics

Key metrics to monitor:
- Password validation success/failure rates
- Breach detection frequency
- API response times and failure rates
- Common password usage patterns

## Getting Started

### 1. Configure Environment Variables

Add to your `.env` file:

```bash
# HaveIBeenPwned API key (optional)
HIBP_API_KEY=your_hibp_api_key_here
```

### 2. Get HIBP API Key (Optional)

1. Visit https://haveibeenpwned.com/API/Key
2. Sign up for a free account
3. Generate an API key
4. Add the key to your environment variables

### 3. Test the Integration

```powershell
# Test password validation
.\scripts\test_password_validation.ps1
```

### 4. Monitor Logs

Check application logs for password validation events:

```bash
tail -f /var/log/api.log | grep "Password validation"
```

## Security Considerations

### Privacy Protection
- **K-Anonymity**: Only password hash prefixes are sent to HIBP
- **No PII**: No personal information is transmitted
- **Local Processing**: Password matching happens locally

### Rate Limiting
- **API Respect**: Service respects HIBP rate limits
- **Graceful Degradation**: Continues operation during API outages
- **Configurable Timeouts**: Prevents hanging requests

### Error Handling
- **Generic Messages**: Prevents information disclosure
- **Audit Logging**: Records validation attempts for security monitoring
- **Fallback Behavior**: Maintains service availability

## Troubleshooting

### Common Issues

1. **API Key Not Configured**
   - Service continues without breach checking
   - Logs: "HIBP API key not configured, skipping breach check"

2. **Network Timeouts**
   - Service falls back gracefully
   - Logs: "Password breach check failed: timeout"

3. **Rate Limiting**
   - Service respects API limits
   - Consider upgrading to paid HIBP plan for higher limits

### Debug Mode

Enable debug logging to troubleshoot issues:

```go
// Add debug logging to password service
log.Printf("Password validation result: %+v", result)
```

## Future Enhancements

### Planned Features
- **Password History**: Prevent reuse of recent passwords
- **Custom Requirements**: Configurable password policies per organization
- **Real-time Feedback**: Frontend integration for live password strength
- **Advanced Scoring**: Machine learning-based password strength assessment

### Integration Opportunities
- **Frontend Components**: React components for password input
- **Admin Dashboard**: Password policy management interface
- **Analytics**: Password strength distribution reporting
- **Compliance**: GDPR and regulatory compliance features

## Support

For issues or questions about the password validation service:

1. Check the logs for error messages
2. Verify environment variable configuration
3. Test with the integration test script
4. Review the unit tests for expected behavior

## References

- [HaveIBeenPwned API Documentation](https://haveibeenpwned.com/API/v3)
- [Pwned Passwords API](https://haveibeenpwned.com/API/v3#PwnedPasswords)
- [K-Anonymity Explanation](https://www.troyhunt.com/ive-just-launched-pwned-passwords-version-2/)
- [Password Security Best Practices](https://owasp.org/www-project-cheat-sheets/cheatsheets/Authentication_Cheat_Sheet.html)












