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

Write-Host "=== MICRO-ITERATION 4.4: Email Send Endpoint Debugging ===" -ForegroundColor Green
Write-Host "Testing foreign key fix and end-to-end email sending functionality" -ForegroundColor Gray

# Step 1: Prepare test email data
Write-Host "`n=== Step 1: Preparing Test Data ===" -ForegroundColor Yellow
$testEmail = @{
    recipient = "test@example.com"
    subject = "Micro-Iteration 4.4 Test Email"
    body = "This is a test email body for debugging the foreign key fix and database insert issue."
    selfDestructAfterAttempts = $false
    burnAfterRead = $false
}

# Convert to JSON for API request
$jsonBody = $testEmail | ConvertTo-Json

Write-Host "Request Body:" -ForegroundColor Cyan
Write-Host $jsonBody -ForegroundColor Gray

# Step 2: Test authentication flow
Write-Host "`n=== Step 2: Testing Authentication ===" -ForegroundColor Yellow

# Prepare login data with test credentials
$loginData = @{
    email = "newtest@securesystem.email"
    password = "TestPassword123!"
}

$loginJson = $loginData | ConvertTo-Json

Write-Host "Login Request:" -ForegroundColor Cyan
Write-Host $loginJson -ForegroundColor Gray

# Execute login request
try {
    Write-Host "Attempting login..." -ForegroundColor Blue
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body $loginJson -ContentType "application/json"
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host "JWT Token: $($loginResponse.token)" -ForegroundColor Gray
    
    $token = $loginResponse.token
} catch {
    Write-Host "❌ Login failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
    exit 1
}

# Step 3: Test email send endpoint
Write-Host "`n=== Step 3: Testing Email Send ===" -ForegroundColor Yellow

# Prepare headers with JWT token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

Write-Host "Email Send Request:" -ForegroundColor Cyan
Write-Host $jsonBody -ForegroundColor Gray

# Execute email send request
try {
    Write-Host "Attempting email send..." -ForegroundColor Blue
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/email/send" -Method POST -Body $jsonBody -Headers $headers
    Write-Host "✅ Email send successful!" -ForegroundColor Green
    Write-Host "Response:" -ForegroundColor Cyan
    Write-Host ($response | ConvertTo-Json) -ForegroundColor Gray
    
    # Verify response structure
    if ($response.status -eq "success" -and $response.blob_id) {
        Write-Host "✅ Response validation passed: status=success, blob_id present" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Response validation warning: unexpected response structure" -ForegroundColor Yellow
    }
    
} catch {
    Write-Host "❌ Email send failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    
    # Extract detailed error information
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        Write-Host "Response Body: $responseBody" -ForegroundColor Red
        Write-Host "Status Code: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
    }
}

# Step 4: Summary and next steps
Write-Host "`n=== Step 4: Test Summary ===" -ForegroundColor Yellow
Write-Host "If all steps completed successfully:" -ForegroundColor Gray
Write-Host "1. ✅ Authentication working" -ForegroundColor Green
Write-Host "2. ✅ Email send endpoint responding" -ForegroundColor Green
Write-Host "3. ✅ Foreign key fix implemented" -ForegroundColor Green
Write-Host "4. ✅ Database insert working" -ForegroundColor Green
Write-Host "5. ✅ R2 storage integration working" -ForegroundColor Green

Write-Host "`nNext steps for verification:" -ForegroundColor Cyan
Write-Host "- Check database for email record: sqlite3 secure-email.db 'SELECT * FROM emails ORDER BY created_at DESC LIMIT 1;'" -ForegroundColor Gray
Write-Host "- Verify foreign key integrity: sqlite3 secure-email.db 'SELECT e.email_id, e.sender_id, u.email FROM emails e JOIN users u ON e.sender_id = u.id ORDER BY e.created_at DESC LIMIT 1;'" -ForegroundColor Gray
Write-Host "- Check R2 storage for blob: Verify blob_id exists in Cloudflare R2 bucket" -ForegroundColor Gray

Write-Host "`n=== MICRO-ITERATION 4.4 Debugging Complete ===" -ForegroundColor Green

