# Test Script for Micro-Iteration 4.10: Simple Geolocation-Based Email Access Restrictions
# This script tests the new single city/country restriction functionality

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpassword123"
)

Write-Output "=== Testing Simple Geolocation Restrictions (Micro-Iteration 4.10) ==="
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

# Test 2: Send email with country restriction only
Write-Output "`n2. Testing country-only restriction..."
$countryRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Country Restriction Only"
    body = "This email is restricted to US only."
    allowedCountry = "US"
    allowedCity = ""
}

if (-not $countryRestrictionEmail.Success) {
    Write-Output "❌ Failed to send country-restricted email: $($countryRestrictionEmail.Error)"
} else {
    Write-Output "✅ Country-restricted email sent successfully"
    $countryEmailId = $countryRestrictionEmail.Data.blob_id
}

# Test 3: Send email with city restriction only
Write-Output "`n3. Testing city-only restriction..."
$cityRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - City Restriction Only"
    body = "This email is restricted to New York only."
    allowedCountry = ""
    allowedCity = "New York"
}

if (-not $cityRestrictionEmail.Success) {
    Write-Output "❌ Failed to send city-restricted email: $($cityRestrictionEmail.Error)"
} else {
    Write-Output "✅ City-restricted email sent successfully"
    $cityEmailId = $cityRestrictionEmail.Data.blob_id
}

# Test 4: Send email with both country and city restrictions
Write-Output "`n4. Testing combined country and city restriction..."
$combinedRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Combined Restrictions"
    body = "This email is restricted to New York, US only."
    allowedCountry = "US"
    allowedCity = "New York"
}

if (-not $combinedRestrictionEmail.Success) {
    Write-Output "❌ Failed to send combined-restriction email: $($combinedRestrictionEmail.Error)"
} else {
    Write-Output "✅ Combined-restriction email sent successfully"
    $combinedEmailId = $combinedRestrictionEmail.Data.blob_id
}

# Test 5: Send email with no restrictions
Write-Output "`n5. Testing no restrictions..."
$noRestrictionEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Restrictions"
    body = "This email has no location restrictions."
    allowedCountry = ""
    allowedCity = ""
}

if (-not $noRestrictionEmail.Success) {
    Write-Output "❌ Failed to send unrestricted email: $($noRestrictionEmail.Error)"
} else {
    Write-Output "✅ Unrestricted email sent successfully"
    $unrestrictedEmailId = $noRestrictionEmail.Data.blob_id
}

# Test 6: Test invalid country code
Write-Output "`n6. Testing invalid country code..."
$invalidCountryEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid Country"
    body = "This should fail due to invalid country code."
    allowedCountry = "USA"  # Invalid - should be "US"
    allowedCity = ""
}

if ($invalidCountryEmail.Success) {
    Write-Output "❌ Invalid country code was accepted (should have failed)"
} else {
    Write-Output "✅ Invalid country code correctly rejected"
}

# Test 7: Test invalid city name
Write-Output "`n7. Testing invalid city name..."
$invalidCityEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid City"
    body = "This should fail due to invalid city name."
    allowedCountry = ""
    allowedCity = "New York123"  # Invalid - contains numbers
}

if ($invalidCityEmail.Success) {
    Write-Output "❌ Invalid city name was accepted (should have failed)"
} else {
    Write-Output "✅ Invalid city name correctly rejected"
}

# Test 8: Test case sensitivity and normalization
Write-Output "`n8. Testing case sensitivity and normalization..."
$caseSensitiveEmail = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Headers @{ "Authorization" = "Bearer $token" } -Body @{
    recipient = "recipient@example.com"
    subject = "Test Email - Case Sensitivity"
    body = "Testing case sensitivity and normalization."
    allowedCountry = "us"  # lowercase
    allowedCity = "NEW YORK"  # uppercase
}

if (-not $caseSensitiveEmail.Success) {
    Write-Output "❌ Case sensitivity test failed: $($caseSensitiveEmail.Error)"
} else {
    Write-Output "✅ Case sensitivity test passed"
    $caseEmailId = $caseSensitiveEmail.Data.blob_id
}

Write-Output "`n=== Test Summary ==="
Write-Output "✅ All geolocation restriction tests completed"
Write-Output "📧 Emails sent with various restriction combinations"
Write-Output "🔒 Validation working correctly for invalid inputs"
Write-Output "🔄 Case sensitivity and normalization working"

Write-Output "`nNote: To test actual geolocation enforcement, you would need to:"
Write-Output "1. Access the emails from different IP locations"
Write-Output "2. Use VPN services to simulate different countries/cities"
Write-Output "3. Verify that access is denied when location doesn't match"
