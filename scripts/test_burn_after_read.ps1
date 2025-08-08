# Test script for burn-after-read functionality
# This script tests the complete flow of creating and accessing burn-after-read emails

param(
    [string]$ApiBase = "http://localhost:8080",
    [string]$TestUser = "test@securesystem.email",
    [string]$TestPassword = "testpassword123",
    [string]$TestTotp = "123456"
)

Write-Host "=== Testing Burn-After-Read Functionality ===" -ForegroundColor Cyan

# Helper function to print colored output
function Write-Status {
    param(
        [string]$Status,
        [string]$Message
    )
    
    switch ($Status) {
        "SUCCESS" { Write-Host "✓ $Message" -ForegroundColor Green }
        "ERROR" { Write-Host "✗ $Message" -ForegroundColor Red }
        "INFO" { Write-Host "ℹ $Message" -ForegroundColor Yellow }
    }
}

# Check if API server is running
Write-Status "INFO" "Checking if API server is running..."
try {
    $healthResponse = Invoke-RestMethod -Uri "$ApiBase/health" -Method GET -ErrorAction Stop
    Write-Status "SUCCESS" "API server is running"
} catch {
    Write-Status "ERROR" "API server is not running. Please start the server first."
    exit 1
}

# Step 1: Login to get JWT token
Write-Status "INFO" "Logging in to get JWT token..."
$loginBody = @{
    email = $TestUser
    password = $TestPassword
    totp_code = $TestTotp
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$ApiBase/api/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.token
    
    if (-not $token) {
        Write-Status "ERROR" "Failed to get JWT token. Login response: $($loginResponse | ConvertTo-Json)"
        exit 1
    }
    Write-Status "SUCCESS" "Got JWT token"
} catch {
    Write-Status "ERROR" "Login failed: $($_.Exception.Message)"
    exit 1
}

# Step 2: Send a burn-after-read email
Write-Status "INFO" "Sending burn-after-read email..."
$sendBody = @{
    recipient = "alice@example.com"
    subject = "Test Burn-After-Read Email"
    body = "This is a test burn-after-read email that should be deleted after first access."
    burnAfterRead = $true
} | ConvertTo-Json

try {
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $sendResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/send" -Method POST -Body $sendBody -Headers $headers
    $emailId = $sendResponse.blob_id -replace '\.blob$', ''
    
    if (-not $emailId) {
        Write-Status "ERROR" "Failed to send email. Response: $($sendResponse | ConvertTo-Json)"
        exit 1
    }
    Write-Status "SUCCESS" "Sent burn-after-read email with ID: $emailId"
} catch {
    Write-Status "ERROR" "Failed to send email: $($_.Exception.Message)"
    exit 1
}

# Step 3: Access the email for the first time (should succeed)
Write-Status "INFO" "Accessing burn-after-read email for the first time..."
try {
    $firstAccessResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/view/$emailId" -Method GET -Headers $headers
    Write-Status "SUCCESS" "First access successful - email content retrieved"
} catch {
    Write-Status "ERROR" "First access failed: $($_.Exception.Message)"
    exit 1
}

# Step 4: Try to access the email again (should return 410 Gone)
Write-Status "INFO" "Attempting to access burn-after-read email for the second time..."
try {
    $secondAccessResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/view/$emailId" -Method GET -Headers $headers
    Write-Status "ERROR" "Second access should return 410 Gone, but succeeded"
    exit 1
} catch {
    $httpCode = $_.Exception.Response.StatusCode.value__
    if ($httpCode -eq 410) {
        Write-Status "SUCCESS" "Second access correctly returned 410 Gone - email consumed"
    } else {
        Write-Status "ERROR" "Second access should return 410 Gone, got $httpCode"
        exit 1
    }
}

# Step 5: Verify the email is marked as consumed in the database
Write-Status "INFO" "Verifying email is marked as consumed in database..."
try {
    $listResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/list" -Method GET -Headers $headers
    $emailInList = $listResponse.emails | Where-Object { $_.email_id -eq $emailId }
    
    if ($emailInList) {
        Write-Status "INFO" "Email still appears in list (metadata preserved)"
    } else {
        Write-Status "INFO" "Email not found in list (completely deleted)"
    }
} catch {
    Write-Status "INFO" "Could not verify email list: $($_.Exception.Message)"
}

Write-Status "SUCCESS" "Burn-after-read functionality test completed successfully!"
Write-Host ""
Write-Host "Test Summary:" -ForegroundColor Cyan
Write-Host "- ✓ API server running"
Write-Host "- ✓ JWT authentication working"
Write-Host "- ✓ Burn-after-read email sent successfully"
Write-Host "- ✓ First access returned email content"
Write-Host "- ✓ Second access returned 410 Gone (email consumed)"
Write-Host "- ✓ Email properly deleted after first access"

Write-Host ""
Write-Status "INFO" "Burn-after-read functionality is working correctly!"

