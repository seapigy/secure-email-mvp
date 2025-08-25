# TOTP Testing Solution for Integration Tests

## Overview

This document describes the solution implemented to fix TOTP login validation for integration testing. The authentication system now supports proper TOTP code generation for testing purposes while maintaining security in production.

## Problem

The original authentication system had the following issues for integration testing:

1. **Invalid TOTP Codes**: Test scripts were using hardcoded TOTP codes like "123456" that didn't match the actual TOTP secrets stored in the database.
2. **Login Failures**: The auth package correctly rejected invalid TOTP codes, causing integration tests to fail.
3. **No TOTP Generation**: There was no way to generate valid TOTP codes for testing purposes.

## Solution

### 1. TOTP Generator Tool

Created a Go-based TOTP generator tool (`cmd/totp_generator/main.go`) that can generate valid TOTP codes for any given base32 secret.

**Usage:**
```bash
# Build the tool
go build -o totp_generator.exe ./cmd/totp_generator

# Generate TOTP code for a secret
.\totp_generator.exe JBSWY3DPEHPK3PXP
# Output: 843987 (example)
```

### 2. PowerShell TOTP Utility

Created a PowerShell utility script (`scripts/get_totp_code.ps1`) that:
- Retrieves the TOTP secret from the database for a given user email
- Generates a valid TOTP code using the TOTP generator tool
- Returns the TOTP code for use in integration tests

**Usage:**
```powershell
# Get TOTP code for a user
$totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email "test@securesystem.email"
# Returns: 364259 (example)
```

### 3. Updated Integration Tests

Updated integration test scripts to use the new TOTP generation functionality:

- `test_login_with_totp.ps1`: Tests login with dynamically generated TOTP codes
- `test_integration_complete.ps1`: Complete integration test with signup and login
- `scripts/test_geo_restrictions.ps1`: Updated geo-restrictions test with proper authentication

## Files Created/Modified

### New Files
- `cmd/totp_generator/main.go`: TOTP code generator tool
- `pkg/auth/totp_test_utils.go`: TOTP utility functions for testing
- `scripts/get_totp_code.ps1`: PowerShell utility for getting TOTP codes
- `test_login_with_totp.ps1`: Login test with valid TOTP codes
- `test_integration_complete.ps1`: Complete integration test
- `docs/totp_testing_solution.md`: This documentation

### Modified Files
- `scripts/test_geo_restrictions.ps1`: Updated to use proper TOTP generation

## How to Use

### For Integration Testing

1. **Build the TOTP generator:**
   ```bash
   go build -o totp_generator.exe ./cmd/totp_generator
   ```

2. **Run integration tests:**
   ```powershell
   # Test complete authentication flow
   powershell -ExecutionPolicy Bypass -File test_integration_complete.ps1
   
   # Test geo-restrictions with authentication
   powershell -ExecutionPolicy Bypass -File scripts/test_geo_restrictions.ps1
   ```

3. **Get TOTP codes for custom testing:**
   ```powershell
   $totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email "user@example.com"
   ```

### For Development

1. **Create a test user:**
   ```powershell
   $signupData = @{
       email = "test@securesystem.email"
       password = "TestPassword123!"
       fallback_email = "fallback@example.com"
   }
   $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body ($signupData | ConvertTo-Json)
   ```

2. **Login with generated TOTP:**
   ```powershell
   $totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email "test@securesystem.email"
   $loginData = @{
       email = "test@securesystem.email"
       password = "TestPassword123!"
       totp_code = $totpCode
   }
   $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body ($loginData | ConvertTo-Json)
   ```

## Security Considerations

### Production Security
- **TOTP Validation**: Production TOTP validation remains secure and enforced
- **No Bypass**: No test-mode flags or bypasses are implemented for production
- **Proper Secrets**: TOTP secrets are properly generated and stored

### Testing Security
- **Isolated Tools**: TOTP generation tools are separate from production code
- **Database Access**: TOTP secrets are retrieved from the database for testing only
- **No Hardcoded Codes**: No hardcoded TOTP codes are used in production

## Testing Results

### Before Fix
- ❌ Login failed with "Invalid credentials"
- ❌ Integration tests blocked by TOTP validation
- ❌ No way to generate valid TOTP codes

### After Fix
- ✅ Login successful with dynamically generated TOTP codes
- ✅ Complete integration tests pass
- ✅ Geo-restrictions integration tests pass
- ✅ Authentication system fully functional for testing

## Example Test Output

```
[INFO] === Complete Integration Test ===
[INFO] Test Email: integration20250812140909@securesystem.email
[INFO] Step 1: Creating user...
[SUCCESS] User created successfully: User created
[INFO] Step 2: Getting TOTP secret from database...
[INFO] TOTP Secret: V6J5ATOZIWKXKGR3JVCN5Y7Y74NPDFZ7
[INFO] Step 3: Generating valid TOTP code...
[INFO] Generated TOTP Code: 629368
[INFO] Step 4: Testing login...
[SUCCESS] Login successful!
[INFO] Access Token: eyJhbGciOiJIUzI1NiIs...
[SUCCESS] === Integration Test PASSED ===
```

## Troubleshooting

### Common Issues

1. **TOTP Generator Not Found**
   - Ensure `totp_generator.exe` is built and in the current directory
   - Run: `go build -o totp_generator.exe ./cmd/totp_generator`

2. **Database Path Issues**
   - Ensure the correct database path is used (`C:\var\db\secure-email.db`)
   - Check that the user exists in the database

3. **PowerShell Execution Policy**
   - Run: `Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope CurrentUser`
   - Or use: `powershell -ExecutionPolicy Bypass -File script.ps1`

4. **Invalid TOTP Secret**
   - Ensure the TOTP secret is valid base32 format
   - Check that the user was created successfully

### Debugging

1. **Check TOTP Secret:**
   ```sql
   sqlite3 C:\var\db\secure-email.db "SELECT email, totp_secret FROM users WHERE email = 'test@securesystem.email';"
   ```

2. **Test TOTP Generation:**
   ```bash
   .\totp_generator.exe JBSWY3DPEHPK3PXP
   ```

3. **Test PowerShell Utility:**
   ```powershell
   powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email "test@securesystem.email"
   ```

## Conclusion

The TOTP testing solution successfully resolves the authentication issues for integration testing while maintaining security in production. Integration tests can now run end-to-end with proper authentication, enabling comprehensive testing of Micro-Iteration 4.7 (GeoIP Country Restriction) and other features.












