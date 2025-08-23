# Test MFA Functionality for Micro-Iteration 4.12
# Tests Multi-Factor Authentication features for secure email access

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Email = "test@example.com",
    [string]$Password = "testpassword123"
)

Write-Output "=== Testing MFA Functionality ==="
Write-Output "Base URL: $BaseUrl"
Write-Output "Test Email: $Email"
Write-Output ""

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
            Write-Output "Request: $Method $uri"
            Write-Output "Body: $jsonBody"
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Body $jsonBody -Headers $headers
        } else {
            Write-Output "Request: $Method $uri"
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }

        Write-Output "Response: $($response | ConvertTo-Json -Depth 10)"
        return $response
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Output "Error Response: $errorBody"
        } else {
            Write-Output "Error: $($_.Exception.Message)"
        }
        return $null
    }
}

# Step 1: Login to get JWT token
Write-Output "Step 1: Logging in..."
$loginBody = @{
    email = $Email
    password = $Password
}

$loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody

if (-not $loginResponse -or -not $loginResponse.token) {
    Write-Output "Login failed. Exiting."
    exit 1
}

$token = $loginResponse.token
Write-Output "Login successful. Token obtained."

# Step 2: Send email with MFA enabled (TOTP)
Write-Output "`nStep 2: Sending email with TOTP MFA..."
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
    Write-Output "Failed to send email with TOTP MFA."
    exit 1
}

$emailId = $sendResponse.blob_id
Write-Output "Email sent successfully with TOTP MFA. Email ID: $emailId"

# Step 3: Get MFA configuration
Write-Output "`nStep 3: Getting MFA configuration..."
$mfaConfigResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/mfa/config/$emailId" -Headers @{Authorization = "Bearer $token"}

if ($mfaConfigResponse) {
    Write-Output "MFA Configuration:"
    Write-Output "  Require MFA: $($mfaConfigResponse.require_mfa)"
    Write-Output "  MFA Type: $($mfaConfigResponse.mfa_type)"
    Write-Output "  Failed Attempts: $($mfaConfigResponse.failed_attempts)"
}

# Step 4: Try to view email without MFA code (should fail)
Write-Output "`nStep 4: Attempting to view email without MFA code..."
$viewResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$emailId" -Headers @{Authorization = "Bearer $token"}

if ($viewResponse -and $viewResponse.code -eq "mfa_required") {
    Write-Output "Correctly blocked access - MFA code required."
} else {
    Write-Output "Unexpected response when trying to view without MFA code."
}

# Step 5: Try to view with invalid MFA code (should fail)
Write-Output "`nStep 5: Attempting to view email with invalid MFA code..."
$viewResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$emailId?mfa_code=123456" -Headers @{Authorization = "Bearer $token"}

if ($viewResponse -and $viewResponse.code -eq "invalid_mfa_code") {
    Write-Output "Correctly rejected invalid MFA code."
} else {
    Write-Output "Unexpected response when trying to view with invalid MFA code."
}

# Step 6: Send email with email-based MFA
Write-Output "`nStep 6: Sending email with email-based MFA..."
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
    Write-Output "Failed to send email with email-based MFA."
    exit 1
}

$emailId2 = $sendResponse2.blob_id
Write-Output "Email sent successfully with email-based MFA. Email ID: $emailId2"

# Step 7: Generate email code
Write-Output "`nStep 7: Generating email code..."
$emailCodeBody = @{
    email_id = $emailId2
}

$emailCodeResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/mfa/email-code" -Body $emailCodeBody -Headers @{Authorization = "Bearer $token"}

if ($emailCodeResponse -and $emailCodeResponse.success) {
    $emailCode = $emailCodeResponse.code
    Write-Output "Email code generated: $emailCode"

    # Step 8: Try to view email with valid email code
    Write-Output "`nStep 8: Attempting to view email with valid email code..."
    $viewResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$emailId2?mfa_code=$emailCode" -Headers @{Authorization = "Bearer $token"}

    if ($viewResponse -and $viewResponse.body) {
        Write-Output "Successfully viewed email with valid email code!"
        Write-Output "Email content: $($viewResponse.body)"
    } else {
        Write-Output "Failed to view email with valid email code."
    }
} else {
    Write-Output "Failed to generate email code."
}

# Step 9: Test MFA validation endpoint
Write-Output "`nStep 9: Testing MFA validation endpoint..."
$validateBody = @{
    email_id = $emailId2
    mfa_code = "000000"  # Invalid code
}

$validateResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/mfa/validate" -Body $validateBody -Headers @{Authorization = "Bearer $token"}

if ($validateResponse -and -not $validateResponse.success) {
    Write-Output "Correctly rejected invalid MFA code via validation endpoint."
} else {
    Write-Output "Unexpected response from MFA validation endpoint."
}

# Step 10: Test brute force protection
Write-Output "`nStep 10: Testing brute force protection..."
for ($i = 1; $i -le 6; $i++) {
    Write-Output "Attempt $i of 6..."
    $validateBody = @{
        email_id = $emailId2
        mfa_code = "000000"
    }

    $validateResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/mfa/validate" -Body $validateBody -Headers @{Authorization = "Bearer $token"}

    if ($validateResponse -and $validateResponse.code -eq "mfa_locked") {
        Write-Output "MFA correctly locked after $i attempts."
        break
    }
}

Write-Output "`n=== MFA Functionality Test Complete ==="
Write-Output "All tests completed. Check the output above for results."
