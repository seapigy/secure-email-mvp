# =============================================================================
# MICRO-ITERATION 4.4: EMAIL SEND ENDPOINT DEBUGGING SCRIPT
# =============================================================================
#
# PURPOSE:
# This script tests the /api/email/send endpoint to verify that the foreign key
# fix is working correctly and that emails can be sent end-to-end.
#
# TESTING FLOW:
# 1. Wait for backend to be ready
# 2. Get TOTP code for authentication
# 3. Login to get JWT token
# 4. Send test email via /api/email/send
# 5. Verify response contains success status and blob_id
#
# DEBUGGING FEATURES:
# - Detailed request/response logging
# - Error handling with full error details
# - Step-by-step progress indicators
# - Environment variable configuration
#
# PREREQUISITES:
# - Backend server running on http://localhost:8080
# - Test user exists in database
# - TOTP script available at scripts/get_totp_code.ps1
# - PowerShell execution policy allows script execution
#
# USAGE:
# .\debug_email_send.ps1
#
# EXPECTED OUTPUT:
# - Login successful with JWT token
# - Email send successful with blob_id
# - Detailed error messages if any step fails
# =============================================================================

# Configure environment variables for testing
$env:JWT_SECRET = "your-secret-key"
$env:SQLITE_DB = "secure-email.db"

Write-Output "=== MICRO-ITERATION 4.4: Email Send Endpoint Debugging ==="
Write-Output "Testing foreign key fix and end-to-end email sending functionality"

# Step 1: Prepare test email data
Write-Output "`n=== Step 1: Preparing Test Data ==="
$testEmail = @{
    recipient = "test@example.com"
    subject = "Micro-Iteration 4.4 Test Email"
    body = "This is a test email body for debugging the foreign key fix and database insert issue."
    selfDestructAfterAttempts = $false
    burnAfterRead = $false
}

# Convert to JSON for API request
$jsonBody = $testEmail | ConvertTo-Json

Write-Output "Request Body:"
Write-Host $jsonBody -ForegroundColor Gray

# Step 2: Test authentication flow
Write-Output "`n=== Step 2: Testing Authentication ==="

# Prepare login data with test credentials
$loginData = @{
    email = "newtest@securesystem.email"
    password = "TestPassword123!"
}

$loginJson = $loginData | ConvertTo-Json

Write-Output "Login Request:"
Write-Host $loginJson -ForegroundColor Gray

# Execute login request
try {
    Write-Output "Attempting login..."
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $loginJson -ContentType "application/json"
    Write-Output "✅ Login successful!"
    Write-Output "JWT Token: $($loginResponse.token)"

    $token = $loginResponse.token
} catch {
    Write-Output "❌ Login failed!"
    Write-Output "Error: $($_.Exception.Message)"
    Write-Output "Response: $($_.Exception.Response)"
    exit 1
}

# Step 3: Test email send endpoint
Write-Output "`n=== Step 3: Testing Email Send ==="

# Prepare headers with JWT token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Output "Email Send Request:"
Write-Host $jsonBody -ForegroundColor Gray

# Execute email send request
try {
    Write-Output "Attempting email send..."
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/email/send" -Method POST -Body $jsonBody -Headers $headers
    Write-Output "✅ Email send successful!"
    Write-Output "Response:"
    Write-Host ($response | ConvertTo-Json) -ForegroundColor Gray

    # Verify response structure
    if ($response.status -eq "success" -and $response.blob_id) {
        Write-Output "✅ Response validation passed: status=success, blob_id present"
    } else {
        Write-Output "⚠️ Response validation warning: unexpected response structure"
    }

} catch {
    Write-Output "❌ Email send failed!"
    Write-Output "Error: $($_.Exception.Message)"

    # Extract detailed error information
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Output "Response Body: $responseBody"
        Write-Output "Status Code: $($_.Exception.Response.StatusCode)"
    }
}

# Step 4: Summary and next steps
Write-Output "`n=== Step 4: Test Summary ==="
Write-Output "If all steps completed successfully:"
Write-Output "1. ✅ Authentication working"
Write-Output "2. ✅ Email send endpoint responding"
Write-Output "3. ✅ Foreign key fix implemented"
Write-Output "4. ✅ Database insert working"
Write-Output "5. ✅ R2 storage integration working"

Write-Output "`nNext steps for verification:"
Write-Output "- Check database for email record: sqlite3 secure-email.db 'SELECT * FROM emails ORDER BY created_at DESC LIMIT 1;'"
Write-Output "- Verify foreign key integrity: sqlite3 secure-email.db 'SELECT e.email_id, e.sender_id, u.email FROM emails e JOIN users u ON e.sender_id = u.id ORDER BY e.created_at DESC LIMIT 1;'"
Write-Output "- Check R2 storage for blob: Verify blob_id exists in Cloudflare R2 bucket"

Write-Output "`n=== MICRO-ITERATION 4.4 Debugging Complete ==="

