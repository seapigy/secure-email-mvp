# Test R2 Configuration
# Checks if R2 environment variables are loaded correctly

Write-Output "=== Testing R2 Configuration ==="

# Check if .env file exists and load it
if (Test-Path ".env") {
    Write-Output "✅ .env file found"

    # Read and display R2 environment variables
    $envContent = Get-Content ".env" | Where-Object { $_ -like "*R2*" }
    Write-Output "R2 Environment Variables in .env:"
    $envContent | ForEach-Object { Write-Output "  $_" }

    # Check if variables are set in current session
    Write-Output "`nChecking environment variables in current session:"
    $r2Vars = @("CLOUDFLARE_R2_ACCESS_KEY", "CLOUDFLARE_R2_SECRET_KEY", "CLOUDFLARE_R2_BUCKET", "CLOUDFLARE_R2_ENDPOINT", "R2_REGION")

    foreach ($var in $r2Vars) {
        $value = [Environment]::GetEnvironmentVariable($var)
        if ($value) {
            if ($var -like "*SECRET*") {
                Write-Output "  ✅ $var = $($value.Substring(0, [Math]::Min(8, $value.Length)))..."
            } else {
                Write-Output "  ✅ $var = $value"
            }
        } else {
            Write-Output "  ❌ $var = NOT SET"
        }
    }
} else {
    Write-Output "❌ .env file not found"
}

# Test R2 connection via API
Write-Output "`n=== Testing R2 Connection via API ==="

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET
    Write-Output "✅ API server is running"

    # Try to send a test email to see if R2 works
    Write-Output "`nTesting email sending with R2..."

    # Get TOTP code
    $totpCode = & powershell -ExecutionPolicy Bypass -File "scripts/get_totp_code.ps1" -Email "test@securesystem.email"
    Write-Output "TOTP Code: $totpCode"

    # Login
    $loginData = @{
        email = "test@securesystem.email"
        password = "TestPassword123!"
        totp_code = $totpCode
    }

    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/login" -Method POST -Headers @{"Content-Type"="application/json"} -Body ($loginData | ConvertTo-Json)
    $accessToken = $loginResponse.access_token
    Write-Output "✅ Login successful"

    # Send email
    $emailData = @{
        recipient = "test@example.com"
        subject = "R2 Test Email"
        body = "This is a test email to verify R2 storage is working."
    }

    $emailResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/email/send" -Method POST -Headers @{"Content-Type"="application/json"; "Authorization"="Bearer $accessToken"} -Body ($emailData | ConvertTo-Json)
    Write-Output "✅ Email sent successfully!"
    Write-Output "Response: $($emailResponse | ConvertTo-Json)"

} catch {
    Write-Output "❌ Error: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Output "Response: $errorBody"
    }
}













