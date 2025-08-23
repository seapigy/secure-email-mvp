# Test Geolocation Restrictions Feature
# Micro-Iteration 4.11: Country & City-Level Geolocation Restrictions

param(
    [string]$ApiHost = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "testpassword123",
    [string]$TotpCode = "123456"
)

Write-Output "=== Testing Geolocation Restrictions Feature ==="
Write-Output "API Host: $ApiHost"
Write-Output ""

# Function to make API requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Token = $null
    )

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    $params = @{
        Method = $Method
        Uri = "$ApiHost$Endpoint"
        Headers = $headers
    }

    if ($Body) {
        $params.Body = $Body | ConvertTo-Json -Depth 10
    }

    try {
        $response = Invoke-RestMethod @params
        return @{
            Success = $true
            Data = $response
        }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            $reader.Close()

            try {
                $errorData = $errorBody | ConvertFrom-Json
                return @{
                    Success = $false
                    StatusCode = $errorResponse.StatusCode
                    Error = $errorData
                }
            }
            catch {
                return @{
                    Success = $false
                    StatusCode = $errorResponse.StatusCode
                    Error = @{ error = $errorBody }
                }
            }
        }
        else {
            return @{
                Success = $false
                Error = @{ error = $_.Exception.Message }
            }
        }
    }
}

# Test 1: Login to get JWT token
Write-Output "1. Testing Login..."
$loginBody = @{
    email = $TestEmail
    password = $TestPassword
    totp_code = $TotpCode
}

$loginResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody

if (-not $loginResult.Success) {
    Write-Output "❌ Login failed: $($loginResult.Error.error)"
    exit 1
}

$token = $loginResult.Data.token
$userId = $loginResult.Data.user_id
Write-Output "✅ Login successful. User ID: $userId"
Write-Output ""

# Test 2: Send email with country restrictions only
Write-Output "2. Testing Country-Only Restrictions..."
$countryRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Country Restrictions"
    body = "This email can only be accessed from the United States."
    allowedCountries = @("us")
    allowedCities = @()
}

$countryResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $countryRestrictionBody -Token $token

if (-not $countryResult.Success) {
    Write-Output "❌ Country restriction email send failed: $($countryResult.Error.error)"
} else {
    Write-Output "✅ Country restriction email sent successfully. Blob ID: $($countryResult.Data.blob_id)"
    $countryEmailId = $countryResult.Data.blob_id -replace "\.blob$", ""
}
Write-Output ""

# Test 3: Send email with city restrictions only
Write-Output "3. Testing City-Only Restrictions..."
$cityRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - City Restrictions"
    body = "This email can only be accessed from New York."
    allowedCountries = @()
    allowedCities = @("New York")
}

$cityResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $cityRestrictionBody -Token $token

if (-not $cityResult.Success) {
    Write-Output "❌ City restriction email send failed: $($cityResult.Error.error)"
} else {
    Write-Output "✅ City restriction email sent successfully. Blob ID: $($cityResult.Data.blob_id)"
    $cityEmailId = $cityResult.Data.blob_id -replace "\.blob$", ""
}
Write-Output ""

# Test 4: Send email with both country and city restrictions
Write-Output "4. Testing Combined Country and City Restrictions..."
$combinedRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Combined Restrictions"
    body = "This email can only be accessed from New York, US."
    allowedCountries = @("us")
    allowedCities = @("New York")
}

$combinedResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $combinedRestrictionBody -Token $token

if (-not $combinedResult.Success) {
    Write-Output "❌ Combined restriction email send failed: $($combinedResult.Error.error)"
} else {
    Write-Output "✅ Combined restriction email sent successfully. Blob ID: $($combinedResult.Data.blob_id)"
    $combinedEmailId = $combinedResult.Data.blob_id -replace "\.blob$", ""
}
Write-Output ""

# Test 5: Send email with no restrictions (control)
Write-Output "5. Testing No Restrictions (Control)..."
$noRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Restrictions"
    body = "This email has no location restrictions."
    allowedCountries = @()
    allowedCities = @()
}

$noRestrictionResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $noRestrictionBody -Token $token

if (-not $noRestrictionResult.Success) {
    Write-Output "❌ No restriction email send failed: $($noRestrictionResult.Error.error)"
} else {
    Write-Output "✅ No restriction email sent successfully. Blob ID: $($noRestrictionResult.Data.blob_id)"
    $noRestrictionEmailId = $noRestrictionResult.Data.blob_id -replace "\.blob$", ""
}
Write-Output ""

# Test 6: Test viewing emails (this will trigger geolocation checks)
Write-Output "6. Testing Email Viewing (Geolocation Enforcement)..."

if ($noRestrictionEmailId) {
    Write-Output "   Testing no restriction email..."
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$noRestrictionEmailId" -Token $token

    if ($viewResult.Success) {
        Write-Output "   ✅ No restriction email view successful"
    } else {
        Write-Output "   ❌ No restriction email view failed: $($viewResult.Error.error)"
    }
}

if ($countryEmailId) {
    Write-Output "   Testing country restriction email..."
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$countryEmailId" -Token $token

    if ($viewResult.Success) {
        Write-Output "   ✅ Country restriction email view successful (location allowed)"
    } elseif ($viewResult.StatusCode -eq 403 -and $viewResult.Error.code -eq "geo_restricted") {
        Write-Output "   ✅ Country restriction email blocked as expected: $($viewResult.Error.error)"
    } else {
        Write-Output "   ❌ Country restriction email view failed: $($viewResult.Error.error)"
    }
}

if ($cityEmailId) {
    Write-Output "   Testing city restriction email..."
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$cityEmailId" -Token $token

    if ($viewResult.Success) {
        Write-Output "   ✅ City restriction email view successful (location allowed)"
    } elseif ($viewResult.StatusCode -eq 403 -and $viewResult.Error.code -eq "geo_restricted") {
        Write-Output "   ✅ City restriction email blocked as expected: $($viewResult.Error.error)"
    } else {
        Write-Output "   ❌ City restriction email view failed: $($viewResult.Error.error)"
    }
}

if ($combinedEmailId) {
    Write-Output "   Testing combined restriction email..."
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$combinedEmailId" -Token $token

    if ($viewResult.Success) {
        Write-Output "   ✅ Combined restriction email view successful (location allowed)"
    } elseif ($viewResult.StatusCode -eq 403 -and $viewResult.Error.code -eq "geo_restricted") {
        Write-Output "   ✅ Combined restriction email blocked as expected: $($viewResult.Error.error)"
    } else {
        Write-Output "   ❌ Combined restriction email view failed: $($viewResult.Error.error)"
    }
}
Write-Output ""

# Test 7: Test invalid country code validation
Write-Output "7. Testing Invalid Country Code Validation..."
$invalidCountryBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid Country"
    body = "This should fail validation."
    allowedCountries = @("USA", "INVALID")
    allowedCities = @()
}

$invalidCountryResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $invalidCountryBody -Token $token

if (-not $invalidCountryResult.Success -and $invalidCountryResult.StatusCode -eq 400) {
    Write-Output "✅ Invalid country code properly rejected: $($invalidCountryResult.Error.error)"
} else {
    Write-Output "❌ Invalid country code validation failed"
}
Write-Output ""

# Test 8: Test invalid city name validation
Write-Output "8. Testing Invalid City Name Validation..."
$invalidCityBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid City"
    body = "This should fail validation."
    allowedCountries = @()
    allowedCities = @("New York!", "City@123")
}

$invalidCityResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $invalidCityBody -Token $token

if (-not $invalidCityResult.Success -and $invalidCityResult.StatusCode -eq 400) {
    Write-Output "✅ Invalid city name properly rejected: $($invalidCityResult.Error.error)"
} else {
    Write-Output "❌ Invalid city name validation failed"
}
Write-Output ""

Write-Output "=== Geolocation Restrictions Test Complete ==="
Write-Output ""
Write-Output "Summary:"
Write-Output "- Country-only restrictions: Tested"
Write-Output "- City-only restrictions: Tested"
Write-Output "- Combined restrictions: Tested"
Write-Output "- No restrictions (control): Tested"
Write-Output "- Geolocation enforcement: Tested"
Write-Output "- Input validation: Tested"
Write-Output ""
Write-Output "Note: Actual geolocation blocking depends on your current IP location."
Write-Output "If you're using a VPN, the results may vary."
