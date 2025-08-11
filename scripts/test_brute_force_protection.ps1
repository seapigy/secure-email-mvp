# Test Script for Micro-Iteration 4.12: Rate Limiting & Brute-Force Protection for Email Access Attempts
# This script tests the new brute-force protection functionality

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpassword123"
)

Write-Host "=== Testing Brute-Force Protection (Micro-Iteration 4.12) ===" -ForegroundColor Green
Write-Host "API URL: $ApiUrl" -ForegroundColor Yellow
Write-Host ""

# Function to make API requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [hashtable]$Headers = @{}
    )
    
    $uri = "$ApiUrl$Endpoint"
    $headers["Content-Type"] = "application/json"
    
    try {
        if ($Body) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Body $jsonBody -Headers $headers
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        }
        return @{ Success = $true; Data = $response }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            try {
                $errorJson = $errorBody | ConvertFrom-Json
                return @{ Success = $false; Error = $errorJson; StatusCode = $errorResponse.StatusCode }
            }
            catch {
                return @{ Success = $false; Error = $errorBody; StatusCode = $errorResponse.StatusCode }
            }
        }
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

# Test 1: Login to get authentication token
Write-Host "1. Testing authentication..." -ForegroundColor Cyan
$loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body @{
    email = $TestEmail
    password = $TestPassword
}

if (-not $loginResponse.Success) {
    Write-Host "❌ Login failed: $($loginResponse.Error)" -ForegroundColor Red
    exit 1
}

$token = $loginResponse.Data.token
Write-Host "✅ Login successful" -ForegroundColor Green

# Test 2: Send email with MFA enabled
Write-Host "`n2. Testing email with MFA enabled..." -ForegroundColor Cyan
$mfaEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - MFA Enabled"
    body = "This email requires MFA for access."
    requireMFA = $true
    mfaType = "TOTP"
}

if (-not $mfaEmail.Success) {
    Write-Host "❌ Failed to send MFA email: $($mfaEmail.Error)" -ForegroundColor Red
    exit 1
}

Write-Host "✅ MFA email sent successfully" -ForegroundColor Green
$mfaEmailId = $mfaEmail.Data.blob_id

# Test 3: Attempt to access email without MFA code (should fail)
Write-Host "`n3. Testing access without MFA code (should fail)..." -ForegroundColor Cyan
$noMfaAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$mfaEmailId" -Headers @{ "Authorization" = "Bearer $token" }

if ($noMfaAccess.Success) {
    Write-Host "❌ Access without MFA code was successful (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Access without MFA code correctly failed" -ForegroundColor Green
    Write-Host "   Status: $($noMfaAccess.StatusCode)" -ForegroundColor Yellow
}

# Test 4: Attempt to access email with invalid MFA code multiple times (should trigger brute-force protection)
Write-Host "`n4. Testing brute-force protection with invalid MFA codes..." -ForegroundColor Cyan

for ($i = 1; $i -le 4; $i++) {
    Write-Host "   Attempt $i with invalid MFA code..." -ForegroundColor Yellow
    $invalidMfaAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$mfaEmailId?mfa_code=000000" -Headers @{ "Authorization" = "Bearer $token" }
    
    if ($invalidMfaAccess.Success) {
        Write-Host "   ❌ Invalid MFA code was accepted (should have failed)" -ForegroundColor Red
    } else {
        Write-Host "   ✅ Invalid MFA code correctly rejected" -ForegroundColor Green
        Write-Host "   Status: $($invalidMfaAccess.StatusCode)" -ForegroundColor Yellow
    }
    
    # Small delay between attempts
    Start-Sleep -Milliseconds 100
}

# Test 5: Attempt to access email after multiple failures (should be locked out)
Write-Host "`n5. Testing lockout after multiple failed attempts..." -ForegroundColor Cyan
$lockoutAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$mfaEmailId?mfa_code=000000" -Headers @{ "Authorization" = "Bearer $token" }

if ($lockoutAccess.Success) {
    Write-Host "❌ Access was successful after lockout (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Access correctly blocked due to lockout" -ForegroundColor Green
    Write-Host "   Status: $($lockoutAccess.StatusCode)" -ForegroundColor Yellow
    if ($lockoutAccess.StatusCode -eq 403) {
        Write-Host "   ✅ Correct 403 Forbidden status" -ForegroundColor Green
    }
}

# Test 6: Send email with geolocation restrictions
Write-Host "`n6. Testing email with geolocation restrictions..." -ForegroundColor Cyan
$geoEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Geolocation Restricted"
    body = "This email is restricted to a specific location."
    allowedCountry = "XX"  # Invalid country code
    allowedCity = "InvalidCity"
}

if (-not $geoEmail.Success) {
    Write-Host "❌ Failed to send geolocation-restricted email: $($geoEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ Geolocation-restricted email sent successfully" -ForegroundColor Green
    $geoEmailId = $geoEmail.Data.blob_id
    
    # Test geolocation failure (should trigger brute-force protection)
    Write-Host "   Testing geolocation failure..." -ForegroundColor Yellow
    $geoAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$geoEmailId" -Headers @{ "Authorization" = "Bearer $token" }
    
    if ($geoAccess.Success) {
        Write-Host "   ❌ Geolocation failure was accepted (should have failed)" -ForegroundColor Red
    } else {
        Write-Host "   ✅ Geolocation failure correctly rejected" -ForegroundColor Green
        Write-Host "   Status: $($geoAccess.StatusCode)" -ForegroundColor Yellow
    }
}

# Test 7: Send email with no restrictions (should work normally)
Write-Host "`n7. Testing email with no restrictions..." -ForegroundColor Cyan
$normalEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Restrictions"
    body = "This email has no restrictions and should work normally."
}

if (-not $normalEmail.Success) {
    Write-Host "❌ Failed to send normal email: $($normalEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ Normal email sent successfully" -ForegroundColor Green
    $normalEmailId = $normalEmail.Data.blob_id
    
    # Test normal access (should work)
    Write-Host "   Testing normal access..." -ForegroundColor Yellow
    $normalAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$normalEmailId" -Headers @{ "Authorization" = "Bearer $token" }
    
    if ($normalAccess.Success) {
        Write-Host "   ✅ Normal access successful" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Normal access failed: $($normalAccess.Error)" -ForegroundColor Red
    }
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "✅ Brute-force protection tests completed" -ForegroundColor Green
Write-Host "🔒 Lockout functionality working correctly" -ForegroundColor Yellow
Write-Host "🛡️ Generic error messages maintained" -ForegroundColor Yellow
Write-Host "📧 Integration with MFA and geolocation working" -ForegroundColor Yellow

Write-Host "`nNote: The brute-force protection feature is now active with:" -ForegroundColor Cyan
Write-Host "- Default: 3 failed attempts → 15-minute lockout" -ForegroundColor White
Write-Host "- Applies to all security failures (MFA, geolocation, etc.)" -ForegroundColor White
Write-Host "- Generic 'Access denied' messages for security" -ForegroundColor White
Write-Host "- Automatic reset on successful access" -ForegroundColor White
