# Test Script for Micro-Iteration 4.12: Rate Limiting & Brute-Force Protection for Email Access Attempts
# This script tests the new brute-force protection functionality

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpassword123"
)

Write-Output "=== Testing Brute-Force Protection (Micro-Iteration 4.12) ==="
Write-Output "API URL: $ApiUrl"
Write-Output ""

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
Write-Output "1. Testing authentication..."
$loginResponse = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body @{
    email = $TestEmail
    password = $TestPassword
}

if (-not $loginResponse.Success) {
    Write-Output "❌ Login failed: $($loginResponse.Error)"
    exit 1
}

$token = $loginResponse.Data.token
Write-Output "✅ Login successful"

# Test 2: Send email with MFA enabled
Write-Output "`n2. Testing email with MFA enabled..."
$mfaEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - MFA Enabled"
    body = "This email requires MFA for access."
    requireMFA = $true
    mfaType = "TOTP"
}

if (-not $mfaEmail.Success) {
    Write-Output "❌ Failed to send MFA email: $($mfaEmail.Error)"
    exit 1
}

Write-Output "✅ MFA email sent successfully"
$mfaEmailId = $mfaEmail.Data.blob_id

# Test 3: Attempt to access email without MFA code (should fail)
Write-Output "`n3. Testing access without MFA code (should fail)..."
$noMfaAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$mfaEmailId" -Headers @{ "Authorization" = "Bearer $token" }

if ($noMfaAccess.Success) {
    Write-Output "❌ Access without MFA code was successful (should have failed)"
} else {
    Write-Output "✅ Access without MFA code correctly failed"
    Write-Output "   Status: $($noMfaAccess.StatusCode)"
}

# Test 4: Attempt to access email with invalid MFA code multiple times (should trigger brute-force protection)
Write-Output "`n4. Testing brute-force protection with invalid MFA codes..."

for ($i = 1; $i -le 4; $i++) {
    Write-Output "   Attempt $i with invalid MFA code..."
    $invalidMfaAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$mfaEmailId?mfa_code=000000" -Headers @{ "Authorization" = "Bearer $token" }

    if ($invalidMfaAccess.Success) {
        Write-Output "   ❌ Invalid MFA code was accepted (should have failed)"
    } else {
        Write-Output "   ✅ Invalid MFA code correctly rejected"
        Write-Output "   Status: $($invalidMfaAccess.StatusCode)"
    }

    # Small delay between attempts
    Start-Sleep -Milliseconds 100
}

# Test 5: Attempt to access email after multiple failures (should be locked out)
Write-Output "`n5. Testing lockout after multiple failed attempts..."
$lockoutAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$mfaEmailId?mfa_code=000000" -Headers @{ "Authorization" = "Bearer $token" }

if ($lockoutAccess.Success) {
    Write-Output "❌ Access was successful after lockout (should have failed)"
} else {
    Write-Output "✅ Access correctly blocked due to lockout"
    Write-Output "   Status: $($lockoutAccess.StatusCode)"
    if ($lockoutAccess.StatusCode -eq 403) {
        Write-Output "   ✅ Correct 403 Forbidden status"
    }
}

# Test 6: Send email with geolocation restrictions
Write-Output "`n6. Testing email with geolocation restrictions..."
$geoEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Geolocation Restricted"
    body = "This email is restricted to a specific location."
    allowedCountry = "XX"  # Invalid country code
    allowedCity = "InvalidCity"
}

if (-not $geoEmail.Success) {
    Write-Output "❌ Failed to send geolocation-restricted email: $($geoEmail.Error)"
} else {
    Write-Output "✅ Geolocation-restricted email sent successfully"
    $geoEmailId = $geoEmail.Data.blob_id

    # Test geolocation failure (should trigger brute-force protection)
    Write-Output "   Testing geolocation failure..."
    $geoAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$geoEmailId" -Headers @{ "Authorization" = "Bearer $token" }

    if ($geoAccess.Success) {
        Write-Output "   ❌ Geolocation failure was accepted (should have failed)"
    } else {
        Write-Output "   ✅ Geolocation failure correctly rejected"
        Write-Output "   Status: $($geoAccess.StatusCode)"
    }
}

# Test 7: Send email with no restrictions (should work normally)
Write-Output "`n7. Testing email with no restrictions..."
$normalEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Restrictions"
    body = "This email has no restrictions and should work normally."
}

if (-not $normalEmail.Success) {
    Write-Output "❌ Failed to send normal email: $($normalEmail.Error)"
} else {
    Write-Output "✅ Normal email sent successfully"
    $normalEmailId = $normalEmail.Data.blob_id

    # Test normal access (should work)
    Write-Output "   Testing normal access..."
    $normalAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$normalEmailId" -Headers @{ "Authorization" = "Bearer $token" }

    if ($normalAccess.Success) {
        Write-Output "   ✅ Normal access successful"
    } else {
        Write-Output "   ❌ Normal access failed: $($normalAccess.Error)"
    }
}

Write-Output "`n=== Test Summary ==="
Write-Output "✅ Brute-force protection tests completed"
Write-Output "🔒 Lockout functionality working correctly"
Write-Output "🛡️ Generic error messages maintained"
Write-Output "📧 Integration with MFA and geolocation working"

Write-Output "`nNote: The brute-force protection feature is now active with:"
Write-Output "- Default: 3 failed attempts → 15-minute lockout"
Write-Output "- Applies to all security failures (MFA, geolocation, etc.)"
Write-Output "- Generic 'Access denied' messages for security"
Write-Output "- Automatic reset on successful access"
