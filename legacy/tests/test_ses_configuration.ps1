# =============================================================================
# AMAZON SES CONFIGURATION TEST
# =============================================================================
# This script sends a test email to cpigusch@gmail.com to verify Amazon SES
# configuration, domain verification, and email authentication (SPF, DKIM, DMARC)
# =============================================================================

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@securesystem.email",
    [string]$TestPassword = "testpass123",
    [string]$TargetEmail = "cpigusch@gmail.com"
)

# =============================================================================
# CONFIGURATION
# =============================================================================

$ErrorActionPreference = "Stop"
$ProgressPreference = "SilentlyContinue"

Write-Host "🔧 Amazon SES Configuration Test" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host "API URL: $ApiUrl" -ForegroundColor Yellow
Write-Host "Test Email: $TestEmail" -ForegroundColor Yellow
Write-Host "Target Email: $TargetEmail" -ForegroundColor Yellow
Write-Host ""

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

function Test-APIHealth {
    Write-Host "🔍 Testing API health..." -ForegroundColor Blue
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/health" -Method GET -TimeoutSec 10
        if ($response.status -eq "ok") {
            Write-Host "✅ API is healthy" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ API health check failed" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ API health check failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-Login {
    param([string]$Email, [string]$Password)
    
    Write-Host "🔐 Testing login for $Email..." -ForegroundColor Blue
    
    $loginData = @{
        email = $Email
        password = $Password
        totp_code = "000000"  # Default TOTP for testing
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/api/auth/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 10
        if ($response.access_token) {
            Write-Host "✅ Login successful" -ForegroundColor Green
            return $response.access_token
        } else {
            Write-Host "❌ Login failed: No token received" -ForegroundColor Red
            return $null
        }
    } catch {
        Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

function Send-SESTestEmail {
    param(
        [string]$Token,
        [string]$Recipient
    )
    
    Write-Host "📧 Sending Amazon SES test email..." -ForegroundColor Blue
    
    $headers = @{
        "Authorization" = "Bearer $Token"
        "Content-Type" = "application/json"
    }
    
    $emailSubject = "Amazon SES Test Email - Secure Email MVP"
    
    $emailBody = @"
This is a test email sent through Amazon SES.

This email is being sent to verify that our Amazon SES configuration is working correctly for the Secure Email MVP project.

Configuration Details:
Domain: securesystem.email
Region: us-east-1
Purpose: Email authentication testing (SPF, DKIM, DMARC)

Expected Results:
If you receive this email, it means:

✅ Amazon SES is properly configured
✅ Domain verification is complete
✅ Email authentication should be working

Next Steps:
Please check the email headers in Gmail to confirm:

SPF = PASS  
DKIM = PASS  
DMARC = PASS  

Test Information:
- Sent at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC')
- From: securesystem.email
- To: $Recipient
- Purpose: SES Configuration Verification

If you see this email, our Amazon SES setup is working correctly!
"@
    
    $emailData = @{
        recipient = $Recipient
        subject = $emailSubject
        body = $emailBody
        # Basic security features for testing
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
    
    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/api/email/send" -Method POST -Headers $headers -Body ($emailData | ConvertTo-Json) -TimeoutSec 30
        
        Write-Host "✅ Email sent successfully" -ForegroundColor Green
        Write-Host "   Email ID: $($response.email_id)" -ForegroundColor Gray
        if ($response.secure_link_url) {
            Write-Host "   Secure Link URL: $($response.secure_link_url)" -ForegroundColor Gray
        }
        
        return $response
    } catch {
        Write-Host "❌ Email send failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "   Error details: $errorBody" -ForegroundColor Red
        }
        return $null
    }
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

Write-Host "🚀 Starting Amazon SES Configuration Test" -ForegroundColor Green
Write-Host ""

# Step 1: Test API Health
if (-not (Test-APIHealth)) {
    Write-Host "❌ API health check failed. Exiting test." -ForegroundColor Red
    exit 1
}

# Step 2: Login to get authentication token
$authToken = Test-Login -Email $TestEmail -Password $TestPassword
if (-not $authToken) {
    Write-Host "❌ Login failed. Exiting test." -ForegroundColor Red
    exit 1
}

# Step 3: Send SES test email
$emailResponse = Send-SESTestEmail -Token $authToken -Recipient $TargetEmail

if (-not $emailResponse) {
    Write-Host "❌ SES test email send failed. Exiting test." -ForegroundColor Red
    exit 1
}

# =============================================================================
# TEST SUMMARY
# =============================================================================

Write-Host ""
Write-Host "🎉 Amazon SES Configuration Test Completed Successfully!" -ForegroundColor Green
Write-Host "================================================================" -ForegroundColor Green
Write-Host "✅ API Health Check: PASSED" -ForegroundColor Green
Write-Host "✅ Authentication: PASSED" -ForegroundColor Green
Write-Host "✅ SES Test Email Send: PASSED" -ForegroundColor Green
Write-Host ""
Write-Host "📋 Test Results Summary:" -ForegroundColor Cyan
Write-Host "   Email ID: $($emailResponse.email_id)" -ForegroundColor Gray
Write-Host "   Recipient: $TargetEmail" -ForegroundColor Gray
Write-Host "   Subject: Amazon SES Test Email - Secure Email MVP" -ForegroundColor Gray
Write-Host "   Sent at: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss UTC')" -ForegroundColor Gray
Write-Host ""
Write-Host "📧 Next Steps:" -ForegroundColor Yellow
Write-Host "   1. Check the email inbox of $TargetEmail" -ForegroundColor White
Write-Host "   2. Verify the test email was received" -ForegroundColor White
Write-Host "   3. Check email headers in Gmail for authentication results:" -ForegroundColor White
Write-Host "      - SPF = PASS" -ForegroundColor White
Write-Host "      - DKIM = PASS" -ForegroundColor White
Write-Host "      - DMARC = PASS" -ForegroundColor White
Write-Host "   4. If all headers show PASS, SES configuration is working correctly" -ForegroundColor White
Write-Host ""
Write-Host "🔍 To check email headers in Gmail:" -ForegroundColor Cyan
Write-Host "   1. Open the email in Gmail" -ForegroundColor White
Write-Host "   2. Click the three dots (⋮) next to Reply" -ForegroundColor White
Write-Host "   3. Select 'Show original'" -ForegroundColor White
Write-Host "   4. Look for SPF, DKIM, and DMARC results in the headers" -ForegroundColor White
Write-Host ""
Write-Host "🏁 Test completed successfully!" -ForegroundColor Green
