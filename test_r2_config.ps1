# Test R2 Configuration
# Checks if R2 environment variables are loaded correctly

Write-Host "=== Testing R2 Configuration ===" -ForegroundColor Blue

# Check if .env file exists and load it
if (Test-Path ".env") {
    Write-Host "✅ .env file found" -ForegroundColor Green
    
    # Read and display R2 environment variables
    $envContent = Get-Content ".env" | Where-Object { $_ -like "*R2*" }
    Write-Host "R2 Environment Variables in .env:" -ForegroundColor Yellow
    $envContent | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
    
    # Check if variables are set in current session
    Write-Host "`nChecking environment variables in current session:" -ForegroundColor Yellow
    $r2Vars = @("R2_ACCESS_KEY_ID", "R2_SECRET_ACCESS_KEY", "R2_BUCKET", "R2_ENDPOINT", "R2_REGION")
    
    foreach ($var in $r2Vars) {
        $value = [Environment]::GetEnvironmentVariable($var)
        if ($value) {
            if ($var -like "*SECRET*") {
                Write-Host "  ✅ $var = $($value.Substring(0, [Math]::Min(8, $value.Length)))..." -ForegroundColor Green
            } else {
                Write-Host "  ✅ $var = $value" -ForegroundColor Green
            }
        } else {
            Write-Host "  ❌ $var = NOT SET" -ForegroundColor Red
        }
    }
} else {
    Write-Host "❌ .env file not found" -ForegroundColor Red
}

# Test R2 connection via API
Write-Host "`n=== Testing R2 Connection via API ===" -ForegroundColor Blue

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Host "✅ API server is running" -ForegroundColor Green
    
    # Try to send a test email to see if R2 works
    Write-Host "`nTesting email sending with R2..." -ForegroundColor Yellow
    
    # Get TOTP code
    $totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email "test@securesystem.email"
    Write-Host "TOTP Code: $totpCode" -ForegroundColor Gray
    
    # Login
    $loginData = @{
        email = "test@securesystem.email"
        password = "TestPassword123!"
        totp_code = $totpCode
    }
    
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body ($loginData | ConvertTo-Json)
    $accessToken = $loginResponse.access_token
    Write-Host "✅ Login successful" -ForegroundColor Green
    
    # Send email
    $emailData = @{
        recipient = "test@example.com"
        subject = "R2 Test Email"
        body = "This is a test email to verify R2 storage is working."
    }
    
    $emailResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/email/send" -Method POST -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer $accessToken"} -Body ($emailData | ConvertTo-Json)
    Write-Host "✅ Email sent successfully!" -ForegroundColor Green
    Write-Host "Response: $($emailResponse | ConvertTo-Json)" -ForegroundColor Gray
    
} catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Host "Response: $errorBody" -ForegroundColor Red
    }
}

