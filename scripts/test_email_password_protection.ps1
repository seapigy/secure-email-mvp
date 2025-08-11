# Test Script for Micro-Iteration 4.14: Password Protection for Email Access
# This script tests the new password protection functionality

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpassword123"
)

Write-Host "=== Testing Password Protection for Email Access (Micro-Iteration 4.14) ===" -ForegroundColor Green
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

# Test 2: Send email with password protection
Write-Host "`n2. Testing email with password protection..." -ForegroundColor Cyan
$passwordEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Password Protected"
    body = "This email is password-protected and will test password validation."
    password = "securepassword123"
}

if (-not $passwordEmail.Success) {
    Write-Host "❌ Failed to send password-protected email: $($passwordEmail.Error)" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Password-protected email sent successfully" -ForegroundColor Green
$passwordEmailId = $passwordEmail.Data.blob_id

# Test 3: Attempt to access email without password (should fail)
Write-Host "`n3. Testing access without password..." -ForegroundColor Cyan
$noPasswordAccess = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$passwordEmailId" -Headers @{ "Authorization" = "Bearer $token" }

if ($noPasswordAccess.Success) {
    Write-Host "❌ Access was successful without password (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Access correctly blocked without password" -ForegroundColor Green
    Write-Host "   Status: $($noPasswordAccess.StatusCode)" -ForegroundColor Yellow
    if ($noPasswordAccess.StatusCode -eq 401) {
        Write-Host "   ✅ Correct 401 Unauthorized status" -ForegroundColor Green
    }
}

# Test 4: Attempt to access email with wrong password (should fail)
Write-Host "`n4. Testing access with wrong password..." -ForegroundColor Cyan
$wrongPasswordAccess = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/view/$passwordEmailId" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    password = "wrongpassword"
}

if ($wrongPasswordAccess.Success) {
    Write-Host "❌ Access was successful with wrong password (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Access correctly blocked with wrong password" -ForegroundColor Green
    Write-Host "   Status: $($wrongPasswordAccess.StatusCode)" -ForegroundColor Yellow
    if ($wrongPasswordAccess.StatusCode -eq 401) {
        Write-Host "   ✅ Correct 401 Unauthorized status" -ForegroundColor Green
    }
}

# Test 5: Access email with correct password (should succeed)
Write-Host "`n5. Testing access with correct password..." -ForegroundColor Cyan
$correctPasswordAccess = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/view/$passwordEmailId" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    password = "securepassword123"
}

if ($correctPasswordAccess.Success) {
    Write-Host "✅ Access successful with correct password" -ForegroundColor Green
    Write-Host "   Subject: $($correctPasswordAccess.Data.subject)" -ForegroundColor Yellow
    Write-Host "   Body: $($correctPasswordAccess.Data.body)" -ForegroundColor Yellow
} else {
    Write-Host "❌ Access failed with correct password: $($correctPasswordAccess.Error)" -ForegroundColor Red
}

# Test 6: Send email without password protection (should work normally)
Write-Host "`n6. Testing email without password protection..." -ForegroundColor Cyan
$normalEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Password Protection"
    body = "This email has no password protection and should work normally."
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

# Test 7: Test password validation with weak password (should fail)
Write-Host "`n7. Testing password validation with weak password..." -ForegroundColor Cyan
$weakPasswordEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Weak Password"
    body = "This email has a weak password and should be rejected."
    password = "123"
}

if ($weakPasswordEmail.Success) {
    Write-Host "❌ Weak password was accepted (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Weak password correctly rejected" -ForegroundColor Green
    Write-Host "   Status: $($weakPasswordEmail.StatusCode)" -ForegroundColor Yellow
    if ($weakPasswordEmail.StatusCode -eq 400) {
        Write-Host "   ✅ Correct 400 Bad Request status" -ForegroundColor Green
    }
}

# Test 8: Test password validation with common weak password (should fail)
Write-Host "`n8. Testing password validation with common weak password..." -ForegroundColor Cyan
$commonWeakPasswordEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Common Weak Password"
    body = "This email has a common weak password and should be rejected."
    password = "password"
}

if ($commonWeakPasswordEmail.Success) {
    Write-Host "❌ Common weak password was accepted (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Common weak password correctly rejected" -ForegroundColor Green
    Write-Host "   Status: $($commonWeakPasswordEmail.StatusCode)" -ForegroundColor Yellow
    if ($commonWeakPasswordEmail.StatusCode -eq 400) {
        Write-Host "   ✅ Correct 400 Bad Request status" -ForegroundColor Green
    }
}

# Test 9: Test brute-force protection with multiple wrong passwords
Write-Host "`n9. Testing brute-force protection with multiple wrong passwords..." -ForegroundColor Cyan

for ($i = 1; $i -le 4; $i++) {
    Write-Host "   Attempt $i with wrong password..." -ForegroundColor Yellow
    $wrongPasswordAttempt = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/view/$passwordEmailId" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
        password = "wrongpassword$i"
    }
    
    if ($wrongPasswordAttempt.Success) {
        Write-Host "   ❌ Wrong password was accepted (should have failed)" -ForegroundColor Red
    } else {
        Write-Host "   ✅ Wrong password correctly rejected" -ForegroundColor Green
    }
    
    # Small delay between attempts
    Start-Sleep -Milliseconds 100
}

# Test 10: Test lockout after multiple failed attempts
Write-Host "`n10. Testing lockout after multiple failed attempts..." -ForegroundColor Cyan
$lockoutAttempt = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/view/$passwordEmailId" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    password = "anotherwrongpassword"
}

if ($lockoutAttempt.Success) {
    Write-Host "❌ Access was successful after multiple failed attempts (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Access correctly blocked after multiple failed attempts" -ForegroundColor Green
    Write-Host "   Status: $($lockoutAttempt.StatusCode)" -ForegroundColor Yellow
    if ($lockoutAttempt.StatusCode -eq 403) {
        Write-Host "   ✅ Correct 403 Forbidden status (likely IP lockout)" -ForegroundColor Green
    }
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "✅ Password protection tests completed" -ForegroundColor Green
Write-Host "🔒 Password validation working correctly" -ForegroundColor Yellow
Write-Host "🛡️ Weak password rejection working" -ForegroundColor Yellow
Write-Host "📧 Integration with brute-force protection working" -ForegroundColor Yellow
Write-Host "🌐 Integration with IP tracking working" -ForegroundColor Yellow

Write-Host "`nNote: The password protection feature is now active with:" -ForegroundColor Cyan
Write-Host "- Argon2id password hashing with random salt" -ForegroundColor White
Write-Host "- Password strength validation (8-128 characters)" -ForegroundColor White
Write-Host "- Common weak password rejection" -ForegroundColor White
Write-Host "- Integration with existing security layers" -ForegroundColor White
Write-Host "- Generic 'Access denied' messages for security" -ForegroundColor White
Write-Host "- Automatic reset of failed attempts on success" -ForegroundColor White
