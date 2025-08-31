# Simple SES Configuration Test
# This script tests the login and email sending functionality step by step

Write-Host "🔧 Simple SES Configuration Test" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan

# Step 1: Test API Health
Write-Host "🔍 Testing API health..." -ForegroundColor Blue
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 10
    if ($response.status -eq "ok") {
        Write-Host "✅ API is healthy" -ForegroundColor Green
    } else {
        Write-Host "❌ API health check failed" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ API health check failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2: Test Login
Write-Host "🔐 Testing login..." -ForegroundColor Blue
$loginData = @{
    email = "test@securesystem.email"
    password = "TestPassword123!"
    totp_code = "000000"
}

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 10
    if ($loginResponse.access_token) {
        Write-Host "✅ Login successful" -ForegroundColor Green
        $authToken = $loginResponse.access_token
    } else {
        Write-Host "❌ Login failed: No token received" -ForegroundColor Red
        Write-Host "Response: $($loginResponse | ConvertTo-Json)" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody" -ForegroundColor Red
    }
    exit 1
}

# Step 3: Send Test Email
Write-Host "📧 Sending test email..." -ForegroundColor Blue
$emailData = @{
    recipient = "cpigusch@gmail.com"
    subject = "Amazon SES Test Email - Secure Email MVP"
    body = "This is a test email sent through Amazon SES to verify configuration."
    selfDestructAfterAttempts = $false
    burnAfterRead = $false
    requireMFA = $false
    password = ""
    timeLock = $false
    expiresAt = (Get-Date).AddDays(30).ToString("yyyy-MM-ddTHH:mm:ssZ")
    remoteRevoke = $false
    stripMetadata = $false
    tamperAlerts = $false
}

$headers = @{
    "Authorization" = "Bearer $authToken"
    "Content-Type" = "application/json"
}

try {
    $emailResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/email/send" -Method POST -Headers $headers -Body ($emailData | ConvertTo-Json) -TimeoutSec 30
    Write-Host "✅ Email sent successfully" -ForegroundColor Green
    Write-Host "Email ID: $($emailResponse.email_id)" -ForegroundColor Gray
    if ($emailResponse.secure_link_url) {
        Write-Host "Secure Link: $($emailResponse.secure_link_url)" -ForegroundColor Gray
    }
} catch {
    Write-Host "❌ Email send failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody" -ForegroundColor Red
    }
    exit 1
}

Write-Host ""
Write-Host "🎉 Test completed successfully!" -ForegroundColor Green
Write-Host "Check cpigusch@gmail.com for the test email." -ForegroundColor Yellow
