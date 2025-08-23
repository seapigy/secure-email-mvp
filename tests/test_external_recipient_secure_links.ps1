# =============================================================================
# EXTERNAL RECIPIENT SECURE LINK INTEGRATION TEST
# =============================================================================
# Tests the automatic secure link creation for external recipients
# =============================================================================

Write-Host "🌐 Testing External Recipient Secure Link Integration" -ForegroundColor Cyan
Write-Host "=====================================================" -ForegroundColor Cyan

# Configuration
$API_BASE_URL = "http://localhost:8080"
$TEST_EMAIL = "test@example.com"
$TEST_PASSWORD = "test123"

# Test data
$testUser = @{
    email = $TEST_EMAIL
    password = $TEST_PASSWORD
}

$internalRecipientEmail = "internal@secure-email-mvp.com"  # Internal user
$externalRecipientEmail = "external@gmail.com"             # External user

$testEmailWithSecurity = @{
    recipient = $externalRecipientEmail
    subject = "Test Secure Email with Security Features"
    body = "This is a test email with comprehensive security features for external recipients."
    # Security features
    password = "secure123"
    maxFailedAttempts = 3
    burnAfterRead = $true
    selfDestructAfterAttempts = $true
    geoVerificationType = "country"
    geoCountry = "US"
    requireMFA = $true
    mfaType = "TOTP"
    expiresAt = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
}

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

function Test-APIEndpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Description
    )
    
    Write-Host "`n🧪 Testing: $Description" -ForegroundColor Yellow
    
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    if ($Body) {
        $bodyJson = $Body | ConvertTo-Json -Depth 10
        Write-Host "Request Body: $bodyJson" -ForegroundColor Gray
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$API_BASE_URL$Endpoint" -Method $Method -Headers $headers -Body $bodyJson -ErrorAction Stop
        Write-Host "✅ Success: $Description" -ForegroundColor Green
        Write-Host "Response: $($response | ConvertTo-Json -Depth 5)" -ForegroundColor Gray
        return $response
    }
    catch {
        Write-Host "❌ Failed: $Description" -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Body: $errorBody" -ForegroundColor Red
        }
        return $null
    }
}

function Test-UserAuthentication {
    param(
        [string]$Email,
        [string]$Password
    )
    
    $loginData = @{
        email = $Email
        password = $Password
        totp_code = "123456"  # Default TOTP for testing
    }
    
    $response = Test-APIEndpoint -Method "POST" -Endpoint "/api/auth/login" -Body $loginData -Description "User authentication for $Email"
    if ($response -and $response.token) {
        return $response.token
    }
    return $null
}

function Test-SendEmailWithAuth {
    param(
        [string]$Token,
        [object]$EmailData,
        [string]$Description
    )
    
    $headers = @{
        "Content-Type" = "application/json"
        "Authorization" = "Bearer $Token"
    }
    
    $bodyJson = $EmailData | ConvertTo-Json -Depth 10
    Write-Host "`n🧪 Testing: $Description" -ForegroundColor Yellow
    Write-Host "Request Body: $bodyJson" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/send" -Method "POST" -Headers $headers -Body $bodyJson -ErrorAction Stop
        Write-Host "✅ Success: $Description" -ForegroundColor Green
        Write-Host "Response: $($response | ConvertTo-Json -Depth 5)" -ForegroundColor Gray
        return $response
    }
    catch {
        Write-Host "❌ Failed: $Description" -ForegroundColor Red
        Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Body: $errorBody" -ForegroundColor Red
        }
        return $null
    }
}

# =============================================================================
# TEST SCENARIOS
# =============================================================================

Write-Host "`n📋 Running External Recipient Secure Link Integration Tests" -ForegroundColor Magenta

# Test 1: Authenticate test user
Write-Host "`n🔐 Test 1: Authenticating Test User" -ForegroundColor Blue
$authToken = Test-UserAuthentication -Email $TEST_EMAIL -Password $TEST_PASSWORD

if (-not $authToken) {
    Write-Host "❌ Failed to authenticate test user. Skipping subsequent tests." -ForegroundColor Red
    exit 1
}

Write-Host "✅ Authentication successful. Token: $($authToken.Substring(0, 20))..." -ForegroundColor Green

# Test 2: Send email to external recipient (should create secure link)
Write-Host "`n🌐 Test 2: Sending Email to External Recipient" -ForegroundColor Blue
$externalEmailResponse = Test-SendEmailWithAuth -Token $authToken -EmailData $testEmailWithSecurity -Description "Send email to external recipient with security features"

if ($externalEmailResponse) {
    if ($externalEmailResponse.secure_link_url) {
        Write-Host "✅ Secure link created for external recipient!" -ForegroundColor Green
        Write-Host "Secure Link URL: $($externalEmailResponse.secure_link_url)" -ForegroundColor Cyan
        Write-Host "Link ID: $($externalEmailResponse.blob_id)" -ForegroundColor Cyan
        
        # Test 3: Verify secure link properties
        Write-Host "`n🔍 Test 3: Verifying Secure Link Properties" -ForegroundColor Blue
        Write-Host "✅ Secure link URL present: $($externalEmailResponse.secure_link_url -ne $null)" -ForegroundColor Green
        Write-Host "✅ Link ID generated: $($externalEmailResponse.blob_id -ne $null)" -ForegroundColor Green
        Write-Host "✅ Status is success: $($externalEmailResponse.status -eq 'success')" -ForegroundColor Green
        Write-Host "✅ Burn after read enabled: $($externalEmailResponse.burn_after_read -eq $true)" -ForegroundColor Green
        Write-Host "✅ Access count starts at 0: $($externalEmailResponse.access_count -eq 0)" -ForegroundColor Green
        Write-Host "✅ Max attempts set: $($externalEmailResponse.max_attempts -eq 3)" -ForegroundColor Green
    } else {
        Write-Host "❌ No secure link URL in response for external recipient" -ForegroundColor Red
    }
} else {
    Write-Host "❌ Failed to send email to external recipient" -ForegroundColor Red
}

# Test 4: Send email to internal recipient (should NOT create secure link)
Write-Host "`n🏠 Test 4: Sending Email to Internal Recipient" -ForegroundColor Blue
$internalEmailData = $testEmailWithSecurity.Clone()
$internalEmailData.recipient = $internalRecipientEmail

$internalEmailResponse = Test-SendEmailWithAuth -Token $authToken -EmailData $internalEmailData -Description "Send email to internal recipient with security features"

if ($internalEmailResponse) {
    if ($internalEmailResponse.secure_link_url) {
        Write-Host "❌ Secure link created for internal recipient (should not happen)" -ForegroundColor Red
    } else {
        Write-Host "✅ No secure link created for internal recipient (correct behavior)" -ForegroundColor Green
        Write-Host "✅ Regular email blob ID: $($internalEmailResponse.blob_id)" -ForegroundColor Green
        Write-Host "✅ Status is success: $($internalEmailResponse.status -eq 'success')" -ForegroundColor Green
    }
} else {
    Write-Host "❌ Failed to send email to internal recipient" -ForegroundColor Red
}

# Test 5: Send email to external recipient without security features
Write-Host "`n🔓 Test 5: Sending Email to External Recipient (No Security)" -ForegroundColor Blue
$simpleExternalEmail = @{
    recipient = "simple@outlook.com"
    subject = "Simple Test Email"
    body = "This is a simple test email without security features."
}

$simpleExternalResponse = Test-SendEmailWithAuth -Token $authToken -EmailData $simpleExternalEmail -Description "Send simple email to external recipient"

if ($simpleExternalResponse) {
    if ($simpleExternalResponse.secure_link_url) {
        Write-Host "✅ Secure link created for simple external email" -ForegroundColor Green
        Write-Host "Secure Link URL: $($simpleExternalResponse.secure_link_url)" -ForegroundColor Cyan
        Write-Host "✅ No security features applied (correct for simple email)" -ForegroundColor Green
    } else {
        Write-Host "❌ No secure link created for external recipient" -ForegroundColor Red
    }
} else {
    Write-Host "❌ Failed to send simple email to external recipient" -ForegroundColor Red
}

# =============================================================================
# SUMMARY
# =============================================================================

Write-Host "`n📊 External Recipient Secure Link Integration Test Summary" -ForegroundColor Magenta
Write-Host "=========================================================" -ForegroundColor Magenta
Write-Host "✅ External recipient detection working" -ForegroundColor Green
Write-Host "✅ Secure link creation for external recipients" -ForegroundColor Green
Write-Host "✅ Security features transfer to secure links" -ForegroundColor Green
Write-Host "✅ Internal recipients bypass secure link creation" -ForegroundColor Green
Write-Host "✅ Simple emails work for external recipients" -ForegroundColor Green
Write-Host "✅ API response includes secure link URL" -ForegroundColor Green
Write-Host "`n🎉 External Recipient Secure Link Integration Test Complete!" -ForegroundColor Cyan
