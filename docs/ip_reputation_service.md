# IP Reputation Service

## Overview

The IP Reputation Service is a security feature that integrates with AbuseIPDB to check the reputation of IP addresses attempting to sign up or log in to the Secure Email MVP. This helps prevent automated attacks, bot registrations, and access from known malicious IP addresses.

## Features

- **Real-time IP reputation checking** on all signup and login requests
- **Automatic blocking** of IPs with abuse confidence scores above the configured threshold
- **Graceful fallback** - allows access if the reputation service is unavailable
- **Proxy support** - correctly identifies client IPs behind proxies, load balancers, and CDNs
- **IPv4 and IPv6 support** - handles both address formats
- **Private IP filtering** - skips reputation checks for private/local IPs
- **Configurable threshold** - adjustable sensitivity for different environments

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `IP_REPUTATION_API_KEY` | Your AbuseIPDB API key | None | Yes (for full protection) |
| `IP_REPUTATION_THRESHOLD` | Malicious threshold (0-100) | 25 | No |

### Getting an AbuseIPDB API Key

1. Visit [AbuseIPDB API](https://www.abuseipdb.com/api)
2. Sign up for a free account
3. Generate an API key
4. Add the key to your `.env` file

### Threshold Configuration

The threshold determines how strict the IP reputation checking is:

- **0-10**: Very strict (blocks most suspicious IPs)
- **10-25**: Moderate (recommended default)
- **25-50**: Permissive (allows more IPs through)
- **50-100**: Very permissive (only blocks highly malicious IPs)

## Implementation Details

### Service Architecture

The IP reputation service is implemented as a Go package (`pkg/reputation/`) with the following components:

- **ReputationService**: Main service for checking IP reputation
- **ReputationConfig**: Configuration management
- **AbuseIPDBResponse**: Response structure for API calls
- **Client IP extraction**: Support for various proxy headers

### Integration Points

#### Signup Handler (`cmd/api/signup_handler.go`)

```go
// Get client IP address
clientIP := reputation.GetClientIP(r)

// Check IP reputation before processing signup
reputationService := reputation.NewReputationService()
ctx := context.Background()

isMalicious, err := reputationService.CheckIPReputation(ctx, clientIP)
if err != nil {
    log.Printf("IP reputation check failed for IP %s: %v", clientIP, err)
    // Continue processing on reputation check failure
} else if isMalicious {
    log.Printf("Signup blocked due to IP reputation for IP %s", clientIP)
    w.WriteHeader(http.StatusForbidden)
    json.NewEncoder(w).Encode(map[string]string{
        "error": "Access denied due to IP reputation",
    })
    return
}
```

#### Login Handler (`cmd/api/login_handler.go`)

Similar integration is implemented in the login handler to check IP reputation before authentication.

### Client IP Detection

The service supports multiple methods for detecting the real client IP:

1. **X-Forwarded-For**: Standard proxy header
2. **X-Real-IP**: Nginx and other reverse proxies
3. **CF-Connecting-IP**: Cloudflare specific header
4. **RemoteAddr**: Direct connection fallback

### Error Handling

The service implements graceful error handling:

- **API failures**: Logs the error but allows access to prevent accidental lockouts
- **Timeouts**: Configurable timeout with fallback to allow access
- **Invalid responses**: Parses responses safely and falls back on errors
- **Missing API key**: Allows all access when not configured

## API Response Format

### Successful Response (Clean IP)

```json
{
  "data": {
    "ipAddress": "8.8.8.8",
    "countryCode": "US",
    "usageType": "Residential",
    "isp": "Google LLC",
    "domain": "google.com",
    "hostnames": "",
    "totalReports": 0,
    "numDistinctUsers": 0,
    "lastReportedAt": null,
    "abuseConfidenceScore": 0
  }
}
```

### Successful Response (Malicious IP)

```json
{
  "data": {
    "ipAddress": "203.0.113.1",
    "countryCode": "XX",
    "usageType": "Hosting",
    "isp": "Malicious ISP",
    "domain": "malicious.com",
    "hostnames": "",
    "totalReports": 150,
    "numDistinctUsers": 45,
    "lastReportedAt": "2024-01-15T10:30:00Z",
    "abuseConfidenceScore": 85
  }
}
```

## Testing

### Unit Tests

Run the unit tests to verify the service functionality:

```bash
go test ./pkg/reputation -v
```

### Integration Tests

Test the integration with the API endpoints:

```powershell
# Test IP reputation integration
.\scripts\test_ip_reputation.ps1
```

### Manual Testing

To test with a known malicious IP:

1. Configure the service with a valid API key
2. Use a VPN or proxy with a flagged IP address
3. Attempt to sign up or log in
4. Verify the request is blocked with a 403 status

## Security Considerations

### Privacy

- Only the IP address is sent to the reputation service
- No user data, credentials, or email content is transmitted
- The service does not store IP addresses locally

### Rate Limiting

- AbuseIPDB has rate limits on their free tier
- The service includes a 10-second timeout to prevent hanging requests
- Failed requests are logged but don't block legitimate users

### Fallback Behavior

- If the reputation service is unavailable, access is allowed
- This prevents accidental lockouts during service outages
- All failures are logged for monitoring

## Monitoring and Logging

### Log Messages

The service logs various events:

```
IP 8.8.8.8 reputation check passed (score: 0%, threshold: 25%)
IP 203.0.113.1 flagged as malicious (score: 85%, threshold: 25%)
IP reputation API request failed for IP 8.8.8.8: timeout
IP reputation API key not configured, allowing access for IP: 8.8.8.8
```

### Metrics to Monitor

- API request success/failure rates
- Number of blocked IPs
- Average response times
- Error rates and types

## Troubleshooting

### Common Issues

1. **API Key Not Working**
   - Verify the key is correct and active
   - Check AbuseIPDB account status
   - Ensure the key has the necessary permissions

2. **All IPs Being Blocked**
   - Check the threshold setting (may be too low)
   - Verify the API is returning correct data
   - Review logs for API response issues

3. **No IPs Being Blocked**
   - Verify the API key is configured
   - Check the threshold setting (may be too high)
   - Review logs for API failures

4. **Proxy Issues**
   - Ensure proxy headers are properly configured
   - Verify the service can detect the real client IP
   - Check network configuration

### Debug Mode

Enable debug logging by setting the log level to debug in your application configuration.

## Future Enhancements

Potential improvements to the IP reputation service:

- **Caching**: Cache reputation results to reduce API calls
- **Multiple providers**: Support for additional reputation services
- **Whitelist/blacklist**: Local IP lists for custom rules
- **Geolocation integration**: Combine with geographic restrictions
- **Machine learning**: Local reputation scoring based on behavior patterns












