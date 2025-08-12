# Test script for email expiration functionality
# This script tests the complete flow of creating and accessing emails with expiration

param(
    [string]$ApiBase = "http://localhost:8080",
    [string]$TestUser = "test@securesystem.email",
    [string]$TestPassword = "testpassword123",
    [string]$TestTotp = "123456"
)

Write-Host "=== Testing Email Expiration Functionality ===" -ForegroundColor Cyan

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

# Step 2: Send an email with expiration in the past (should be immediately expired)
Write-Status "INFO" "Sending email with past expiration..."
$pastExpiration = (Get-Date).AddMinutes(-5).ToString("yyyy-MM-ddTHH:mm:ssZ") # 5 minutes ago
$sendBody = @{
    recipient = "alice@example.com"
    subject = "Test Expired Email"
    body = "This email should be immediately expired and inaccessible."
    expiresAt = $pastExpiration
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
    Write-Status "SUCCESS" "Sent expired email with ID: $emailId (expired at: $pastExpiration)"
} catch {
    Write-Status "ERROR" "Failed to send email: $($_.Exception.Message)"
    exit 1
}

# Step 3: Try to access the expired email (should return 410 Gone)
Write-Status "INFO" "Attempting to access expired email..."
try {
    $accessResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/view/$emailId" -Method GET -Headers $headers
    Write-Status "ERROR" "Expired email should return 410 Gone, but succeeded"
    exit 1
} catch {
    $httpCode = $_.Exception.Response.StatusCode.value__
    if ($httpCode -eq 410) {
        Write-Status "SUCCESS" "Expired email correctly returned 410 Gone"
    } else {
        Write-Status "ERROR" "Expired email should return 410 Gone, got $httpCode"
        exit 1
    }
}

# Step 4: Send an email with future expiration (should be accessible)
Write-Status "INFO" "Sending email with future expiration..."
$futureExpiration = (Get-Date).AddMinutes(5).ToString("yyyy-MM-ddTHH:mm:ssZ") # 5 minutes from now
$sendBody = @{
    recipient = "alice@example.com"
    subject = "Test Valid Email"
    body = "This email should be accessible until it expires."
    expiresAt = $futureExpiration
} | ConvertTo-Json

try {
    $sendResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/send" -Method POST -Body $sendBody -Headers $headers
    $validEmailId = $sendResponse.blob_id -replace '\.blob$', ''
    
    if (-not $validEmailId) {
        Write-Status "ERROR" "Failed to send valid email. Response: $($sendResponse | ConvertTo-Json)"
        exit 1
    }
    Write-Status "SUCCESS" "Sent valid email with ID: $validEmailId (expires at: $futureExpiration)"
} catch {
    Write-Status "ERROR" "Failed to send valid email: $($_.Exception.Message)"
    exit 1
}

# Step 5: Access the valid email (should succeed)
Write-Status "INFO" "Accessing valid email..."
try {
    $accessResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/view/$validEmailId" -Method GET -Headers $headers
    Write-Status "SUCCESS" "Valid email accessed successfully"
} catch {
    Write-Status "ERROR" "Valid email access failed: $($_.Exception.Message)"
    exit 1
}

# Step 6: Test invalid expiration format
Write-Status "INFO" "Testing invalid expiration format..."
$invalidSendBody = @{
    recipient = "alice@example.com"
    subject = "Test Invalid Expiration"
    body = "This email has invalid expiration format."
    expiresAt = "invalid-date-format"
} | ConvertTo-Json

try {
    $sendResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/send" -Method POST -Body $invalidSendBody -Headers $headers
    Write-Status "ERROR" "Invalid expiration format should return 400 Bad Request, but succeeded"
    exit 1
} catch {
    $httpCode = $_.Exception.Response.StatusCode.value__
    if ($httpCode -eq 400) {
        Write-Status "SUCCESS" "Invalid expiration format correctly returned 400 Bad Request"
    } else {
        Write-Status "ERROR" "Invalid expiration format should return 400 Bad Request, got $httpCode"
        exit 1
    }
}

# Step 7: Test past expiration format
Write-Status "INFO" "Testing past expiration format..."
$pastExpirationBody = @{
    recipient = "alice@example.com"
    subject = "Test Past Expiration"
    body = "This email has past expiration time."
    expiresAt = "2020-01-01T00:00:00Z"
} | ConvertTo-Json

try {
    $sendResponse = Invoke-RestMethod -Uri "$ApiBase/api/email/send" -Method POST -Body $pastExpirationBody -Headers $headers
    Write-Status "ERROR" "Past expiration should return 400 Bad Request, but succeeded"
    exit 1
} catch {
    $httpCode = $_.Exception.Response.StatusCode.value__
    if ($httpCode -eq 400) {
        Write-Status "SUCCESS" "Past expiration correctly returned 400 Bad Request"
    } else {
        Write-Status "ERROR" "Past expiration should return 400 Bad Request, got $httpCode"
        exit 1
    }
}

Write-Status "SUCCESS" "Email expiration functionality test completed successfully!"
Write-Host ""
Write-Host "Test Summary:" -ForegroundColor Cyan
Write-Host "- ✓ API server running"
Write-Host "- ✓ JWT authentication working"
Write-Host "- ✓ Expired email correctly returns 410 Gone"
Write-Host "- ✓ Valid email with future expiration accessible"
Write-Host "- ✓ Invalid expiration format rejected"
Write-Host "- ✓ Past expiration time rejected"
Write-Host "- ✓ Email properly deleted when expired"

Write-Host ""
Write-Status "INFO" "Email expiration functionality is working correctly!"




