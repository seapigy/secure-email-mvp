# Test MFA Functionality for Micro-Iteration 4.12
# Tests Multi-Factor Authentication features for secure email access

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Email = "test@example.com",
    [string]$Password = "testpassword123"
)

Write-Host "=== Testing MFA Functionality ===" -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Yellow
Write-Host "Test Email: $Email" -ForegroundColor Yellow
Write-Host ""

# Function to make API requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )
    
    $uri = "$BaseUrl$Endpoint"
    $headers["Content-Type"] = "application/json"
    
    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            Write-Host "Request: $Method $uri" -ForegroundColor Cyan
            Write-Host "Body: $jsonBody" -ForegroundColor Gray
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Body $jsonBody -Headers $headers
        } else {
            Write-Host "Request: $Method $uri" -ForegroundColor Cyan
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        
        Write-Host "Response: $($response | ConvertTo-Json -Depth 10)" -ForegroundColor Green
        return $response
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Response: $errorBody" -ForegroundColor Red
        } else {
            Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        }
        return $null
    }
}

# Step 1: Login to get JWT token
Write-Host "Step 1: Logging in..." -ForegroundColor Yellow
$loginBody = @{
    email = $Email
    password = $Password
}

$loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody

if (-not $loginResponse -or -not $loginResponse.token) {
    Write-Host "Login failed. Exiting." -ForegroundColor Red
    exit 1
}

$token = $loginResponse.token
Write-Host "Login successful. Token obtained." -ForegroundColor Green

# Step 2: Send email with MFA enabled (TOTP)
Write-Host "`nStep 2: Sending email with TOTP MFA..." -ForegroundColor Yellow
$emailBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email with TOTP MFA"
    body = "This is a test email with TOTP-based MFA enabled."
    requireMFA = $true
    mfaType = "TOTP"
    burnAfterRead = $true
}

$sendResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $emailBody -Headers @{Authorization = "Bearer $token"}

if (-not $sendResponse -or -not $sendResponse.blob_id) {
    Write-Host "Failed to send email with TOTP MFA." -ForegroundColor Red
    exit 1
}

$emailId = $sendResponse.blob_id
Write-Host "Email sent successfully with TOTP MFA. Email ID: $emailId" -ForegroundColor Green

# Step 3: Get MFA configuration
Write-Host "`nStep 3: Getting MFA configuration..." -ForegroundColor Yellow
$mfaConfigResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/mfa/config/$emailId" -Headers @{Authorization = "Bearer $token"}

if ($mfaConfigResponse) {
    Write-Host "MFA Configuration:" -ForegroundColor Green
    Write-Host "  Require MFA: $($mfaConfigResponse.require_mfa)" -ForegroundColor White
    Write-Host "  MFA Type: $($mfaConfigResponse.mfa_type)" -ForegroundColor White
    Write-Host "  Failed Attempts: $($mfaConfigResponse.failed_attempts)" -ForegroundColor White
}

# Step 4: Try to view email without MFA code (should fail)
Write-Host "`nStep 4: Attempting to view email without MFA code..." -ForegroundColor Yellow
$viewResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$emailId" -Headers @{Authorization = "Bearer $token"}

if ($viewResponse -and $viewResponse.code -eq "mfa_required") {
    Write-Host "Correctly blocked access - MFA code required." -ForegroundColor Green
} else {
    Write-Host "Unexpected response when trying to view without MFA code." -ForegroundColor Red
}

# Step 5: Try to view with invalid MFA code (should fail)
Write-Host "`nStep 5: Attempting to view email with invalid MFA code..." -ForegroundColor Yellow
$viewResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$emailId?mfa_code=123456" -Headers @{Authorization = "Bearer $token"}

if ($viewResponse -and $viewResponse.code -eq "invalid_mfa_code") {
    Write-Host "Correctly rejected invalid MFA code." -ForegroundColor Green
} else {
    Write-Host "Unexpected response when trying to view with invalid MFA code." -ForegroundColor Red
}

# Step 6: Send email with email-based MFA
Write-Host "`nStep 6: Sending email with email-based MFA..." -ForegroundColor Yellow
$emailBody2 = @{
    recipient = "recipient2@example.com"
    subject = "Test Email with Email MFA"
    body = "This is a test email with email-based MFA enabled."
    requireMFA = $true
    mfaType = "EMAIL_CODE"
    burnAfterRead = $true
}

$sendResponse2 = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $emailBody2 -Headers @{Authorization = "Bearer $token"}

if (-not $sendResponse2 -or -not $sendResponse2.blob_id) {
    Write-Host "Failed to send email with email-based MFA." -ForegroundColor Red
    exit 1
}

$emailId2 = $sendResponse2.blob_id
Write-Host "Email sent successfully with email-based MFA. Email ID: $emailId2" -ForegroundColor Green

# Step 7: Generate email code
Write-Host "`nStep 7: Generating email code..." -ForegroundColor Yellow
$emailCodeBody = @{
    email_id = $emailId2
}

$emailCodeResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/mfa/email-code" -Body $emailCodeBody -Headers @{Authorization = "Bearer $token"}

if ($emailCodeResponse -and $emailCodeResponse.success) {
    $emailCode = $emailCodeResponse.code
    Write-Host "Email code generated: $emailCode" -ForegroundColor Green
    
    # Step 8: Try to view email with valid email code
    Write-Host "`nStep 8: Attempting to view email with valid email code..." -ForegroundColor Yellow
    $viewResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$emailId2?mfa_code=$emailCode" -Headers @{Authorization = "Bearer $token"}

    if ($viewResponse -and $viewResponse.body) {
        Write-Host "Successfully viewed email with valid email code!" -ForegroundColor Green
        Write-Host "Email content: $($viewResponse.body)" -ForegroundColor White
    } else {
        Write-Host "Failed to view email with valid email code." -ForegroundColor Red
    }
} else {
    Write-Host "Failed to generate email code." -ForegroundColor Red
}

# Step 9: Test MFA validation endpoint
Write-Host "`nStep 9: Testing MFA validation endpoint..." -ForegroundColor Yellow
$validateBody = @{
    email_id = $emailId2
    mfa_code = "000000"  # Invalid code
}

$validateResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/mfa/validate" -Body $validateBody -Headers @{Authorization = "Bearer $token"}

if ($validateResponse -and -not $validateResponse.success) {
    Write-Host "Correctly rejected invalid MFA code via validation endpoint." -ForegroundColor Green
} else {
    Write-Host "Unexpected response from MFA validation endpoint." -ForegroundColor Red
}

# Step 10: Test brute force protection
Write-Host "`nStep 10: Testing brute force protection..." -ForegroundColor Yellow
for ($i = 1; $i -le 6; $i++) {
    Write-Host "Attempt $i of 6..." -ForegroundColor Yellow
    $validateBody = @{
        email_id = $emailId2
        mfa_code = "000000"
    }
    
    $validateResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/mfa/validate" -Body $validateBody -Headers @{Authorization = "Bearer $token"}
    
    if ($validateResponse -and $validateResponse.code -eq "mfa_locked") {
        Write-Host "MFA correctly locked after $i attempts." -ForegroundColor Green
        break
    }
}

Write-Host "`n=== MFA Functionality Test Complete ===" -ForegroundColor Green
Write-Host "All tests completed. Check the output above for results." -ForegroundColor White
