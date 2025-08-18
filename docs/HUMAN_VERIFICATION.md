# Human Verification System

## Overview

The Human Verification System is an anti-automation layer designed to prevent automated attacks while maintaining a good user experience for legitimate users. It implements both CAPTCHA and proof-of-work verification methods to ensure that sensitive operations are performed by humans rather than bots.

## Features

### Verification Methods

1. **Proof-of-Work Verification**
   - Server generates cryptographic challenges
   - Client solves challenges by finding hash preimages
   - Configurable difficulty levels
   - No external dependencies

2. **CAPTCHA Verification**
   - Integration with Google reCAPTCHA v3
   - Score-based verification (0.0 = bot, 1.0 = human)
   - Configurable score thresholds
   - Fallback to proof-of-work if CAPTCHA is unavailable

### Protected Endpoints

The following endpoints require human verification:

- `POST /api/email/{id}/trust-device` - Trust current device
- `POST /api/email/{id}/session` - Generate session token

### Security Features

- **Audit Logging**: All verification attempts are logged with IP, user agent, and result
- **Failure Tracking**: High failure rates trigger monitoring alerts
- **Configurable Thresholds**: Adjustable failure thresholds and ban durations
- **Generic Error Messages**: Prevents information leakage on failures

## Architecture

### Database Schema

```sql
CREATE TABLE human_verification_logs (
    id TEXT PRIMARY KEY,
    email_id TEXT NOT NULL,
    ip_address TEXT NOT NULL,
    user_agent TEXT,
    verification_type TEXT NOT NULL, -- 'captcha' or 'proof_of_work'
    challenge_id TEXT,
    result TEXT NOT NULL, -- 'success', 'failure', 'timeout'
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    details TEXT, -- JSON field for additional verification details
    FOREIGN KEY (email_id) REFERENCES emails(email_id) ON DELETE CASCADE
);

-- Indexes for abuse detection and audit queries
CREATE INDEX idx_human_verification_logs_email_ip ON human_verification_logs(email_id, ip_address);
CREATE INDEX idx_human_verification_logs_timestamp ON human_verification_logs(timestamp);
CREATE INDEX idx_human_verification_logs_ip_timestamp ON human_verification_logs(ip_address, timestamp);
CREATE INDEX idx_human_verification_logs_result ON human_verification_logs(result, timestamp);
```

### Service Architecture

```
HumanVerificationService (Interface)
├── VerifyResponse() - Verify verification tokens
├── GenerateChallenge() - Generate proof-of-work challenges
├── LogVerification() - Log verification attempts
└── GetVerificationStats() - Get abuse detection statistics

Implementations:
├── HumanVerificationServiceImpl - Production implementation
└── MockHumanVerificationService - Testing implementation
```

### Middleware Integration

The `HumanVerificationMiddleware` wraps protected endpoints and:

1. Checks if verification is enabled
2. Extracts verification tokens from headers or query parameters
3. Verifies tokens using the appropriate method
4. Logs verification attempts
5. Returns challenges when no token is provided
6. Handles verification failures with generic error messages

## Configuration

### Environment Variables

```bash
# Enable/disable human verification
HUMAN_VERIFICATION_ENABLED=true

# Verification method: 'captcha' or 'proof_of_work'
HUMAN_VERIFICATION_TYPE=proof_of_work

# CAPTCHA Configuration (for reCAPTCHA v3)
CAPTCHA_SECRET_KEY=your_recaptcha_secret_key
CAPTCHA_SITE_KEY=your_recaptcha_site_key
CAPTCHA_ENDPOINT=https://www.google.com/recaptcha/api/siteverify

# Proof-of-Work Configuration
PROOF_OF_WORK_DIFFICULTY=4  # Number of leading zeros required
PROOF_OF_WORK_MAX_NONCE=1000000  # Maximum nonce value

# Abuse Detection
HUMAN_VERIFICATION_FAILURE_THRESHOLD=5  # Max failures before alert
HUMAN_VERIFICATION_BAN_DURATION=15m  # Ban duration for repeated failures
```

### Default Configuration

```go
Config{
    Enabled:              true,
    VerificationType:     "proof_of_work",
    CAPTCHAEndpoint:      "https://www.google.com/recaptcha/api/siteverify",
    ProofOfWorkDifficulty: 4,
    MaxNonce:             1000000,
    FailureThreshold:     5,
    BanDuration:          15 * time.Minute,
}
```

## API Endpoints

### Generate Challenge

**GET** `/api/human-verification/challenge`

Returns a verification challenge based on the configured verification type.

#### Response (Proof-of-Work)

```json
{
    "challenge": {
        "id": "challenge-uuid",
        "prefix": "random-prefix",
        "target": "0000",
        "max_nonce": 1000000
    },
    "type": "proof_of_work"
}
```

#### Response (CAPTCHA)

```json
{
    "captcha_site_key": "your_site_key",
    "type": "captcha"
}
```

### Protected Endpoints

When accessing protected endpoints without a verification token, the response will be:

```json
{
    "verification_required": true,
    "verification_type": "proof_of_work",
    "challenge": {
        "id": "challenge-uuid",
        "prefix": "random-prefix",
        "target": "0000",
        "max_nonce": 1000000
    },
    "message": "Human verification required. Solve the proof-of-work challenge and include the solution in X-Human-Verification-Token header."
}
```

## Usage Examples

### Proof-of-Work Verification

1. **Request Challenge**
   ```bash
   curl -X GET http://localhost:8080/api/human-verification/challenge
   ```

2. **Solve Challenge** (Client-side)
   ```javascript
   // Find nonce such that SHA256(prefix + nonce) starts with target
   const challenge = response.challenge;
   let nonce = 0;
   while (nonce < challenge.max_nonce) {
       const data = challenge.prefix + nonce;
       const hash = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(data));
       const hashHex = Array.from(new Uint8Array(hash))
           .map(b => b.toString(16).padStart(2, '0'))
           .join('');
       
       if (hashHex.startsWith(challenge.target)) {
           break;
       }
       nonce++;
   }
   
   const solution = challenge.id + ':' + nonce;
   ```

3. **Use Solution**
   ```bash
   curl -X POST http://localhost:8080/api/email/email-id/trust-device \
     -H "Authorization: Bearer jwt-token" \
     -H "X-Human-Verification-Token: challenge-id:nonce"
   ```

### CAPTCHA Verification

1. **Request Challenge**
   ```bash
   curl -X GET http://localhost:8080/api/human-verification/challenge
   ```

2. **Complete CAPTCHA** (Client-side)
   ```javascript
   // Use reCAPTCHA v3
   grecaptcha.ready(function() {
       grecaptcha.execute('site_key', {action: 'verify'})
           .then(function(token) {
               // Use token in request
           });
   });
   ```

3. **Use Token**
   ```bash
   curl -X POST http://localhost:8080/api/email/email-id/trust-device \
     -H "Authorization: Bearer jwt-token" \
     -H "X-Human-Verification-Token: captcha-token"
   ```

## Security Considerations

### Information Leakage Prevention

- Generic error messages prevent attackers from determining if an email exists
- No distinction between invalid tokens and missing tokens
- Failed verification attempts increment self-destruct counters

### Rate Limiting

- Verification attempts are logged and tracked
- High failure rates trigger monitoring alerts
- Configurable thresholds for abuse detection

### Privacy

- IP addresses are logged for abuse detection
- User agents are logged for pattern analysis
- Verification details are stored in JSON format for flexibility

## Monitoring and Analytics

### Verification Statistics

The system tracks:

- Total verification attempts per email/IP
- Success/failure rates
- Failure patterns over time
- Geographic distribution of attempts

### Alerting

High failure rates or suspicious patterns can trigger:

- Log alerts for manual review
- Automatic blocking of IP addresses
- Integration with existing security monitoring

## Testing

### Unit Tests

```bash
go test ./pkg/humanverification/... -v
```

### Integration Tests

```bash
go test ./cmd/api/... -run TestHumanVerification -v
```

### Mock Service

For testing, use `MockHumanVerificationService`:

```go
mockSvc := humanverification.NewMockHumanVerificationService()
mockSvc.SetVerificationResult(false) // Force failure
mockSvc.SetSpecificVerification("token", true) // Specific token success
```

## Troubleshooting

### Common Issues

1. **Verification Always Fails**
   - Check CAPTCHA secret key configuration
   - Verify proof-of-work difficulty settings
   - Review server logs for verification errors

2. **Challenges Not Generated**
   - Ensure human verification is enabled
   - Check database connectivity
   - Verify migration has been applied

3. **High Failure Rates**
   - Review verification configuration
   - Check for client-side implementation issues
   - Monitor for automated attack patterns

### Debug Mode

Enable debug logging by setting:

```bash
export HUMAN_VERIFICATION_DEBUG=true
```

This will log detailed verification attempts and challenge generation.

## Future Enhancements

### Planned Features

1. **Additional Verification Methods**
   - hCaptcha integration
   - Custom challenge types
   - Biometric verification

2. **Advanced Analytics**
   - Machine learning-based threat detection
   - Real-time abuse prevention
   - Geographic threat intelligence

3. **Performance Optimizations**
   - Challenge caching
   - Distributed verification
   - Async verification processing

### Integration Opportunities

- SIEM system integration
- Threat intelligence feeds
- Automated response systems
- User behavior analytics
