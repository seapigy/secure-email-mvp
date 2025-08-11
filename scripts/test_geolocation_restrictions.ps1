# Test Geolocation Restrictions Feature
# Micro-Iteration 4.11: Country & City-Level Geolocation Restrictions

param(
    [string]$ApiHost = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "testpassword123",
    [string]$TotpCode = "123456"
)

Write-Host "=== Testing Geolocation Restrictions Feature ===" -ForegroundColor Green
Write-Host "API Host: $ApiHost" -ForegroundColor Yellow
Write-Host ""

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
Write-Host "1. Testing Login..." -ForegroundColor Cyan
$loginBody = @{
    email = $TestEmail
    password = $TestPassword
    totp_code = $TotpCode
}

$loginResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody

if (-not $loginResult.Success) {
    Write-Host "❌ Login failed: $($loginResult.Error.error)" -ForegroundColor Red
    exit 1
}

$token = $loginResult.Data.token
$userId = $loginResult.Data.user_id
Write-Host "✅ Login successful. User ID: $userId" -ForegroundColor Green
Write-Host ""

# Test 2: Send email with country restrictions only
Write-Host "2. Testing Country-Only Restrictions..." -ForegroundColor Cyan
$countryRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Country Restrictions"
    body = "This email can only be accessed from the United States."
    allowedCountries = @("us")
    allowedCities = @()
}

$countryResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $countryRestrictionBody -Token $token

if (-not $countryResult.Success) {
    Write-Host "❌ Country restriction email send failed: $($countryResult.Error.error)" -ForegroundColor Red
} else {
    Write-Host "✅ Country restriction email sent successfully. Blob ID: $($countryResult.Data.blob_id)" -ForegroundColor Green
    $countryEmailId = $countryResult.Data.blob_id -replace "\.blob$", ""
}
Write-Host ""

# Test 3: Send email with city restrictions only
Write-Host "3. Testing City-Only Restrictions..." -ForegroundColor Cyan
$cityRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - City Restrictions"
    body = "This email can only be accessed from New York."
    allowedCountries = @()
    allowedCities = @("New York")
}

$cityResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $cityRestrictionBody -Token $token

if (-not $cityResult.Success) {
    Write-Host "❌ City restriction email send failed: $($cityResult.Error.error)" -ForegroundColor Red
} else {
    Write-Host "✅ City restriction email sent successfully. Blob ID: $($cityResult.Data.blob_id)" -ForegroundColor Green
    $cityEmailId = $cityResult.Data.blob_id -replace "\.blob$", ""
}
Write-Host ""

# Test 4: Send email with both country and city restrictions
Write-Host "4. Testing Combined Country and City Restrictions..." -ForegroundColor Cyan
$combinedRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Combined Restrictions"
    body = "This email can only be accessed from New York, US."
    allowedCountries = @("us")
    allowedCities = @("New York")
}

$combinedResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $combinedRestrictionBody -Token $token

if (-not $combinedResult.Success) {
    Write-Host "❌ Combined restriction email send failed: $($combinedResult.Error.error)" -ForegroundColor Red
} else {
    Write-Host "✅ Combined restriction email sent successfully. Blob ID: $($combinedResult.Data.blob_id)" -ForegroundColor Green
    $combinedEmailId = $combinedResult.Data.blob_id -replace "\.blob$", ""
}
Write-Host ""

# Test 5: Send email with no restrictions (control)
Write-Host "5. Testing No Restrictions (Control)..." -ForegroundColor Cyan
$noRestrictionBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - No Restrictions"
    body = "This email has no location restrictions."
    allowedCountries = @()
    allowedCities = @()
}

$noRestrictionResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $noRestrictionBody -Token $token

if (-not $noRestrictionResult.Success) {
    Write-Host "❌ No restriction email send failed: $($noRestrictionResult.Error.error)" -ForegroundColor Red
} else {
    Write-Host "✅ No restriction email sent successfully. Blob ID: $($noRestrictionResult.Data.blob_id)" -ForegroundColor Green
    $noRestrictionEmailId = $noRestrictionResult.Data.blob_id -replace "\.blob$", ""
}
Write-Host ""

# Test 6: Test viewing emails (this will trigger geolocation checks)
Write-Host "6. Testing Email Viewing (Geolocation Enforcement)..." -ForegroundColor Cyan

if ($noRestrictionEmailId) {
    Write-Host "   Testing no restriction email..." -ForegroundColor Yellow
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$noRestrictionEmailId" -Token $token
    
    if ($viewResult.Success) {
        Write-Host "   ✅ No restriction email view successful" -ForegroundColor Green
    } else {
        Write-Host "   ❌ No restriction email view failed: $($viewResult.Error.error)" -ForegroundColor Red
    }
}

if ($countryEmailId) {
    Write-Host "   Testing country restriction email..." -ForegroundColor Yellow
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$countryEmailId" -Token $token
    
    if ($viewResult.Success) {
        Write-Host "   ✅ Country restriction email view successful (location allowed)" -ForegroundColor Green
    } elseif ($viewResult.StatusCode -eq 403 -and $viewResult.Error.code -eq "geo_restricted") {
        Write-Host "   ✅ Country restriction email blocked as expected: $($viewResult.Error.error)" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Country restriction email view failed: $($viewResult.Error.error)" -ForegroundColor Red
    }
}

if ($cityEmailId) {
    Write-Host "   Testing city restriction email..." -ForegroundColor Yellow
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$cityEmailId" -Token $token
    
    if ($viewResult.Success) {
        Write-Host "   ✅ City restriction email view successful (location allowed)" -ForegroundColor Green
    } elseif ($viewResult.StatusCode -eq 403 -and $viewResult.Error.code -eq "geo_restricted") {
        Write-Host "   ✅ City restriction email blocked as expected: $($viewResult.Error.error)" -ForegroundColor Green
    } else {
        Write-Host "   ❌ City restriction email view failed: $($viewResult.Error.error)" -ForegroundColor Red
    }
}

if ($combinedEmailId) {
    Write-Host "   Testing combined restriction email..." -ForegroundColor Yellow
    $viewResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/email/view/$combinedEmailId" -Token $token
    
    if ($viewResult.Success) {
        Write-Host "   ✅ Combined restriction email view successful (location allowed)" -ForegroundColor Green
    } elseif ($viewResult.StatusCode -eq 403 -and $viewResult.Error.code -eq "geo_restricted") {
        Write-Host "   ✅ Combined restriction email blocked as expected: $($viewResult.Error.error)" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Combined restriction email view failed: $($viewResult.Error.error)" -ForegroundColor Red
    }
}
Write-Host ""

# Test 7: Test invalid country code validation
Write-Host "7. Testing Invalid Country Code Validation..." -ForegroundColor Cyan
$invalidCountryBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid Country"
    body = "This should fail validation."
    allowedCountries = @("USA", "INVALID")
    allowedCities = @()
}

$invalidCountryResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $invalidCountryBody -Token $token

if (-not $invalidCountryResult.Success -and $invalidCountryResult.StatusCode -eq 400) {
    Write-Host "✅ Invalid country code properly rejected: $($invalidCountryResult.Error.error)" -ForegroundColor Green
} else {
    Write-Host "❌ Invalid country code validation failed" -ForegroundColor Red
}
Write-Host ""

# Test 8: Test invalid city name validation
Write-Host "8. Testing Invalid City Name Validation..." -ForegroundColor Cyan
$invalidCityBody = @{
    recipient = "recipient@example.com"
    subject = "Test Email - Invalid City"
    body = "This should fail validation."
    allowedCountries = @()
    allowedCities = @("New York!", "City@123")
}

$invalidCityResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/send" -Body $invalidCityBody -Token $token

if (-not $invalidCityResult.Success -and $invalidCityResult.StatusCode -eq 400) {
    Write-Host "✅ Invalid city name properly rejected: $($invalidCityResult.Error.error)" -ForegroundColor Green
} else {
    Write-Host "❌ Invalid city name validation failed" -ForegroundColor Red
}
Write-Host ""

Write-Host "=== Geolocation Restrictions Test Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "- Country-only restrictions: Tested" -ForegroundColor White
Write-Host "- City-only restrictions: Tested" -ForegroundColor White
Write-Host "- Combined restrictions: Tested" -ForegroundColor White
Write-Host "- No restrictions (control): Tested" -ForegroundColor White
Write-Host "- Geolocation enforcement: Tested" -ForegroundColor White
Write-Host "- Input validation: Tested" -ForegroundColor White
Write-Host ""
Write-Host "Note: Actual geolocation blocking depends on your current IP location." -ForegroundColor Cyan
Write-Host "If you're using a VPN, the results may vary." -ForegroundColor Cyan
