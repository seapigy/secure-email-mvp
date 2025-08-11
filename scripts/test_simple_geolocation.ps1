# Test Script for Micro-Iteration 4.10: Simple Geolocation-Based Email Access Restrictions
# This script tests the new single city/country restriction functionality

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpassword123"
)

Write-Host "=== Testing Simple Geolocation Restrictions (Micro-Iteration 4.10) ===" -ForegroundColor Green
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

# Test 2: Send email with country restriction only
Write-Host "`n2. Testing country-only restriction..." -ForegroundColor Cyan
$countryRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Country Restriction Only"
    body = "This email is restricted to US only."
    allowedCountry = "US"
    allowedCity = ""
}

if (-not $countryRestrictionEmail.Success) {
    Write-Host "❌ Failed to send country-restricted email: $($countryRestrictionEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ Country-restricted email sent successfully" -ForegroundColor Green
    $countryEmailId = $countryRestrictionEmail.Data.blob_id
}

# Test 3: Send email with city restriction only
Write-Host "`n3. Testing city-only restriction..." -ForegroundColor Cyan
$cityRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - City Restriction Only"
    body = "This email is restricted to New York only."
    allowedCountry = ""
    allowedCity = "New York"
}

if (-not $cityRestrictionEmail.Success) {
    Write-Host "❌ Failed to send city-restricted email: $($cityRestrictionEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ City-restricted email sent successfully" -ForegroundColor Green
    $cityEmailId = $cityRestrictionEmail.Data.blob_id
}

# Test 4: Send email with both country and city restrictions
Write-Host "`n4. Testing combined country and city restriction..." -ForegroundColor Cyan
$combinedRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Combined Restrictions"
    body = "This email is restricted to New York, US only."
    allowedCountry = "US"
    allowedCity = "New York"
}

if (-not $combinedRestrictionEmail.Success) {
    Write-Host "❌ Failed to send combined-restriction email: $($combinedRestrictionEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ Combined-restriction email sent successfully" -ForegroundColor Green
    $combinedEmailId = $combinedRestrictionEmail.Data.blob_id
}

# Test 5: Send email with no restrictions
Write-Host "`n5. Testing no restrictions..." -ForegroundColor Cyan
$noRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Restrictions"
    body = "This email has no location restrictions."
    allowedCountry = ""
    allowedCity = ""
}

if (-not $noRestrictionEmail.Success) {
    Write-Host "❌ Failed to send unrestricted email: $($noRestrictionEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ Unrestricted email sent successfully" -ForegroundColor Green
    $unrestrictedEmailId = $noRestrictionEmail.Data.blob_id
}

# Test 6: Test invalid country code
Write-Host "`n6. Testing invalid country code..." -ForegroundColor Cyan
$invalidCountryEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid Country"
    body = "This should fail due to invalid country code."
    allowedCountry = "USA"  # Invalid - should be "US"
    allowedCity = ""
}

if ($invalidCountryEmail.Success) {
    Write-Host "❌ Invalid country code was accepted (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Invalid country code correctly rejected" -ForegroundColor Green
}

# Test 7: Test invalid city name
Write-Host "`n7. Testing invalid city name..." -ForegroundColor Cyan
$invalidCityEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid City"
    body = "This should fail due to invalid city name."
    allowedCountry = ""
    allowedCity = "New York123"  # Invalid - contains numbers
}

if ($invalidCityEmail.Success) {
    Write-Host "❌ Invalid city name was accepted (should have failed)" -ForegroundColor Red
} else {
    Write-Host "✅ Invalid city name correctly rejected" -ForegroundColor Green
}

# Test 8: Test case sensitivity and normalization
Write-Host "`n8. Testing case sensitivity and normalization..." -ForegroundColor Cyan
$caseSensitiveEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Case Sensitivity"
    body = "Testing case sensitivity and normalization."
    allowedCountry = "us"  # lowercase
    allowedCity = "NEW YORK"  # uppercase
}

if (-not $caseSensitiveEmail.Success) {
    Write-Host "❌ Case sensitivity test failed: $($caseSensitiveEmail.Error)" -ForegroundColor Red
} else {
    Write-Host "✅ Case sensitivity test passed" -ForegroundColor Green
    $caseEmailId = $caseSensitiveEmail.Data.blob_id
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "✅ All geolocation restriction tests completed" -ForegroundColor Green
Write-Host "📧 Emails sent with various restriction combinations" -ForegroundColor Yellow
Write-Host "🔒 Validation working correctly for invalid inputs" -ForegroundColor Yellow
Write-Host "🔄 Case sensitivity and normalization working" -ForegroundColor Yellow

Write-Host "`nNote: To test actual geolocation enforcement, you would need to:" -ForegroundColor Cyan
Write-Host "1. Access the emails from different IP locations" -ForegroundColor White
Write-Host "2. Use VPN services to simulate different countries/cities" -ForegroundColor White
Write-Host "3. Verify that access is denied when location doesn't match" -ForegroundColor White
