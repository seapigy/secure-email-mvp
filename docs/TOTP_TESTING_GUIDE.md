# TOTP Testing Guide

## Overview

This guide explains how to properly test TOTP (Time-based One-Time Password) authentication in the Secure Email MVP system.

## The Problem with Hardcoded TOTP Codes

**❌ WRONG**: Using hardcoded TOTP codes like "123456" in tests
```powershell
# This will NEVER work
$loginData = @{
    email = "test@example.com"
    password = "password123"
    totp_code = "123456"  # ❌ This is invalid
}
```

**✅ CORRECT**: Using real TOTP codes generated from actual secrets
```powershell
# This will work correctly
$totpOutput = & "scripts\generate_totp.ps1" -Secret "JBSWY3DPEHPK3PXP"
$currentTOTPCode = $totpOutput.Split("`n") | Where-Object { $_ -match 'CURRENT:' } | ForEach-Object { $_.Split(':')[1].Trim() }

$loginData = @{
    email = "test@example.com"
    password = "password123"
    totp_code = $currentTOTPCode  # ✅ This is valid
}
```

## Test User Configuration

### Default Test User
- **Email**: `test@securesystem.email`
- **Password**: `Test123!@#`
- **TOTP Secret**: `JBSWY3DPEHPK3PXP`
- **Purpose**: Used for automated testing and development

### How TOTP Works
1. **Time-based**: TOTP codes change every 30 seconds
2. **Secret-based**: Each user has a unique secret key
3. **Algorithm**: Uses HMAC-SHA1 with 6-digit output
4. **Tolerance**: Accepts codes from ±2 time steps (±60 seconds)

## Using the TOTP Generator Script

### Basic Usage
```powershell
# Generate current TOTP code for test user
.\scripts\generate_totp.ps1

# Output:
# TOTP Code Generator for Secret: JBSWY3DPEHPK3PXP
# Current Time Step: 58538629
# 
# PREVIOUS (step -3): 814346
# PREVIOUS (step -2): 352369
# PREVIOUS (step -1): 601092
# CURRENT: 885867
# NEXT (step 1): 463919
# NEXT (step 2): 515405
# NEXT (step 3): 822061
```

### Advanced Usage
```powershell
# Generate codes for different secrets
.\scripts\generate_totp.ps1 -Secret "CUSTOM_SECRET_HERE"

# Generate more time steps
.\scripts\generate_totp.ps1 -Steps 5
```

### Integration in Test Scripts
```powershell
# Method 1: Extract current code
$totpOutput = & "scripts\generate_totp.ps1" -Secret "JBSWY3DPEHPK3PXP"
$currentTOTPCode = $totpOutput.Split("`n") | Where-Object { $_ -match 'CURRENT:' } | ForEach-Object { $_.Split(':')[1].Trim() }

# Method 2: Extract specific step
$previousTOTPCode = $totpOutput.Split("`n") | Where-Object { $_ -match 'PREVIOUS \(step -1\):' } | ForEach-Object { $_.Split(':')[1].Trim() }
```

## Updated Test Scripts

The following test scripts have been updated to use real TOTP codes:

1. `tests/comprehensive_user_flow_validation.ps1`
2. `debug_auth.ps1`
3. `tests/test_external_recipient_secure_links.ps1`

## Creating Custom Test Users

### Method 1: Using Signup API
```powershell
# Create a new test user
$signupData = @{
    email = "custom-test@example.com"
    password = "Test123!@#"
    fallback_email = "fallback@example.com"
}

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/signup" -Method POST -Body ($signupData | ConvertTo-Json) -ContentType "application/json"
```

### Method 2: Direct Database Insertion
```sql
-- Insert test user with known TOTP secret
INSERT INTO users (
    email, 
    password, 
    password_hash, 
    totp_secret, 
    fallback_email, 
    fallback_token, 
    fallback_confirmed, 
    fallback_token_expiration, 
    created_at
) VALUES (
    'custom-test@example.com',
    'Test123!@#',
    'Test123!@#',
    'CUSTOM_SECRET_HERE',
    'fallback@example.com',
    'test-token',
    0,
    datetime('now', '+1 day'),
    datetime('now')
);
```

## Troubleshooting

### Common Issues

1. **Rate Limiting**: Too many failed attempts trigger rate limiting
   - **Solution**: Wait 1-5 minutes before retrying
   - **Quick Fix**: Run `.\scripts\reset_rate_limits.ps1`
   - **Test Mode**: Use `.\scripts\setup_test_mode.ps1` to disable rate limiting

2. **Expired TOTP Code**: TOTP codes expire after 30 seconds
   - **Solution**: Generate a fresh code using the generator script

3. **Wrong Secret**: Using wrong TOTP secret for user
   - **Solution**: Check the user's actual TOTP secret in the database

4. **Time Synchronization**: Server and client time out of sync
   - **Solution**: Use the generator script which handles time steps correctly

### Rate Limiting Solutions

#### Option 1: Test Mode (Recommended for Development)
```powershell
# Set up test mode with relaxed rate limits
.\scripts\setup_test_mode.ps1

# Run your tests
.\tests\comprehensive_user_flow_validation.ps1
```

#### Option 2: Reset Rate Limits
```powershell
# Reset rate limits and wait for window to expire
.\scripts\reset_rate_limits.ps1

# Then run your tests
.\tests\comprehensive_user_flow_validation.ps1
```

#### Option 3: Manual Environment Variables
```powershell
# Set environment variables manually
$env:TEST_MODE = "true"
$env:LOGIN_RATE_LIMIT_ENABLED = "0"
$env:RATE_LIMIT_REQUESTS = "100"

# Restart server and run tests
```

### Debug Commands
```powershell
# Check if test user exists
sqlite3 data/secure_email.db "SELECT email, totp_secret FROM users WHERE email='test@securesystem.email';"

# Check current TOTP code
.\scripts\generate_totp.ps1

# Test login with current code
$currentCode = (.\scripts\generate_totp.ps1 | Select-String "CURRENT:").ToString().Split(':')[1].Trim()
curl -s http://localhost:8080/api/auth/login -X POST -H "Content-Type: application/json" -d "{\"email\":\"test@securesystem.email\",\"password\":\"Test123!@#\",\"totp_code\":\"$currentCode\"}"
```

## Best Practices

1. **Always use real TOTP codes** in tests
2. **Generate codes just before use** to avoid expiration
3. **Use the test user** (`test@securesystem.email`) for automated testing
4. **Handle rate limiting** in test scripts
5. **Document custom test users** and their secrets
6. **Use environment variables** for different test environments

## Migration Guide

To update existing test scripts:

1. **Find hardcoded TOTP codes**:
   ```powershell
   Get-ChildItem -Recurse -Include "*.ps1" | Select-String "123456"
   ```

2. **Replace with TOTP generator**:
   ```powershell
   # Replace this:
   totp_code = "123456"
   
   # With this:
   $totpOutput = & "scripts\generate_totp.ps1" -Secret "JBSWY3DPEHPK3PXP"
   $currentTOTPCode = $totpOutput.Split("`n") | Where-Object { $_ -match 'CURRENT:' } | ForEach-Object { $_.Split(':')[1].Trim() }
   totp_code = $currentTOTPCode
   ```

3. **Test the updated script** to ensure it works correctly

## Security Notes

- **Never commit real TOTP secrets** to version control
- **Use test-specific secrets** for automated testing
- **Rotate test secrets** periodically
- **Monitor for suspicious activity** in test environments
