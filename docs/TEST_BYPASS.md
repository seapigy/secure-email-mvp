# Test Bypass Mode Documentation

## Overview

The Secure Email API includes a safe test bypass mode that allows full testing of all endpoints without being blocked by production security features. This mode is designed for development and testing purposes only and should never be enabled in production.

## ⚠️ Security Warning

**This feature is for testing only and should NEVER be enabled in production environments.**
- Test bypass mode disables critical security features
- Only use in development/testing environments
- Always verify TEST_BYPASS=false in production deployments

## Configuration

### Environment Variable

Set the following environment variable to enable test bypass mode:

```bash
TEST_BYPASS=true
```

### Default Test Account

When test bypass is enabled, a default test account is automatically created:

- **Email**: `test@example.com`
- **Password**: `Test1234!`
- **User ID**: `test-user-12345`

## Features

### Bypassed Security Layers

When test bypass is enabled, the following security features are bypassed for the test user:

1. **TOTP/MFA Authentication** - No TOTP code required
2. **Password Complexity Checks** - Simple password accepted
3. **Brute Force Protection** - No lockouts or rate limiting
4. **Geolocation Restrictions** - No location-based blocking
5. **Human Verification** - No CAPTCHA or verification challenges
6. **IP Reputation Checks** - No IP-based blocking

### Authentication Flow

1. **Login**: Use test credentials without TOTP
2. **Token Generation**: Full JWT + refresh token pair
3. **API Access**: All endpoints accessible with test user token
4. **Session Management**: Normal session handling applies

## Usage Examples

### Login with Test Account

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test1234!",
    "totp_code": "123456"
  }'
```

### Access Protected Endpoints

```bash
# Get user info
curl -X GET http://localhost:8080/api/auth/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# List inbox
curl -X GET http://localhost:8080/api/inbox/list \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"

# Get specific email
curl -X GET http://localhost:8080/api/inbox/email-id \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Implementation Details

### Files Modified

1. **`pkg/testbypass/config.go`** - Configuration management
2. **`pkg/testbypass/seeder.go`** - Test user creation
3. **`cmd/api/login_handler.go`** - Login bypass logic
4. **`cmd/api/main.go`** - Startup integration

### Key Functions

- `testbypass.LoadConfig()` - Load configuration from environment
- `testbypass.SeedTestUser()` - Create test user account
- `testbypass.Config.IsTestUser()` - Check if email is test user

### Safety Measures

1. **Environment Check**: Only active when TEST_BYPASS=true
2. **Specific User**: Only affects test@example.com
3. **Production Safe**: No impact on real user accounts
4. **Clear Logging**: All bypass actions are logged

## Testing Workflow

### 1. Enable Test Bypass

```bash
export TEST_BYPASS=true
```

### 2. Start Server

```bash
go run ./cmd/api
```

### 3. Verify Test User Creation

Look for these log messages:
```
[TEST_BYPASS] 🔧 Test bypass mode enabled - seeding test user account
[TEST_BYPASS] ✅ Test user created successfully
```

### 4. Test Authentication

```bash
# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test1234!","totp_code":"123456"}'
```

### 5. Test API Endpoints

Use the returned access token to test all protected endpoints.

## Production Deployment

### Disable Test Bypass

Ensure TEST_BYPASS is not set or explicitly set to false:

```bash
export TEST_BYPASS=false
# or
unset TEST_BYPASS
```

### Verification

1. Check environment variables
2. Verify no test bypass logs in production
3. Confirm all security features are active
4. Test with real user accounts only

## Troubleshooting

### Test User Not Created

- Check TEST_BYPASS environment variable
- Verify database connection
- Check server startup logs

### Login Fails

- Ensure using correct test credentials
- Check server is running
- Verify test bypass is enabled

### Endpoints Still Protected

- Confirm access token is valid
- Check token expiration
- Verify endpoint requires authentication

## Security Considerations

1. **Never enable in production**
2. **Use only for development/testing**
3. **Monitor logs for bypass usage**
4. **Regular security audits**
5. **Clear separation from production data**

## Support

For issues with test bypass mode:
1. Check server logs for error messages
2. Verify environment configuration
3. Ensure database is accessible
4. Test with provided examples
