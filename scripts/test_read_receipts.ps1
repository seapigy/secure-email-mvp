# =============================================================================
# SECURE EMAIL MVP - READ RECEIPTS TEST SCRIPT
# =============================================================================
# PowerShell script to test read receipts and expiration alerts functionality.
# Micro-Iteration 4.19: Email Read Receipt & Expiration Alerts
# =============================================================================

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpassword123"
)

Write-Host "Starting Read Receipts and Expiration Alerts Test..." -ForegroundColor Green
Write-Host "API URL: $ApiUrl" -ForegroundColor Yellow

# Test 1: User Registration and Login
Write-Host "`n=== Test 1: User Registration and Login ===" -ForegroundColor Cyan

$signupData = @{
    email = $TestEmail
    password = $TestPassword
}

$signupResponse = Invoke-RestMethod -Uri "$ApiUrl/api/auth/signup" -Method POST -Body ($signupData | ConvertTo-Json) -ContentType "application/json"
Write-Host "Signup Response: $($signupResponse | ConvertTo-Json)" -ForegroundColor Gray

# Login to get JWT token
$loginData = @{
    email = $TestEmail
    password = $TestPassword
    totp_code = "123456"  # Default TOTP for testing
}

$loginResponse = Invoke-RestMethod -Uri "$ApiUrl/api/auth/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json"
$jwtToken = $loginResponse.token
$userId = $loginResponse.user_id

Write-Host "Login successful. User ID: $userId" -ForegroundColor Green

# Test 2: Get Read Receipt Preferences
Write-Host "`n=== Test 2: Get Read Receipt Preferences ===" -ForegroundColor Cyan

$headers = @{
    "Authorization" = "Bearer $jwtToken"
    "Content-Type" = "application/json"
}

try {
    $prefsResponse = Invoke-RestMethod -Uri "$ApiUrl/api/read-receipts/preferences" -Method GET -Headers $headers
    Write-Host "Read Receipt Preferences: $($prefsResponse | ConvertTo-Json)" -ForegroundColor Gray
} catch {
    Write-Host "Error getting preferences: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 3: Update Read Receipt Preferences
Write-Host "`n=== Test 3: Update Read Receipt Preferences ===" -ForegroundColor Cyan

$updatePrefsData = @{
    enable_read_receipts = $true
    enable_expiration_alerts = $true
    expiration_alert_hours = 48
    delivery_methods = "email,sms"
}

try {
    $updatePrefsResponse = Invoke-RestMethod -Uri "$ApiUrl/api/read-receipts/preferences" -Method PUT -Headers $headers -Body ($updatePrefsData | ConvertTo-Json)
    Write-Host "Updated Preferences: $($updatePrefsResponse | ConvertTo-Json)" -ForegroundColor Gray
} catch {
    Write-Host "Error updating preferences: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Send Email with Read Receipt Settings
Write-Host "`n=== Test 4: Send Email with Read Receipt Settings ===" -ForegroundColor Cyan

$sendEmailData = @{
    recipient = "recipient@example.com"
    subject = "Test Email with Read Receipts"
    body = "This is a test email to verify read receipt functionality."
    expires_at = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
}

try {
    $sendResponse = Invoke-RestMethod -Uri "$ApiUrl/api/email/send" -Method POST -Headers $headers -Body ($sendEmailData | ConvertTo-Json)
    $emailId = $sendResponse.blob_id
    Write-Host "Email sent successfully. Email ID: $emailId" -ForegroundColor Green
} catch {
    Write-Host "Error sending email: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test 5: Get Email Read Receipt Info
Write-Host "`n=== Test 5: Get Email Read Receipt Info ===" -ForegroundColor Cyan

try {
    $readReceiptInfo = Invoke-RestMethod -Uri "$ApiUrl/api/emails/$emailId/read-receipts" -Method GET -Headers $headers
    Write-Host "Read Receipt Info: $($readReceiptInfo | ConvertTo-Json)" -ForegroundColor Gray
} catch {
    Write-Host "Error getting read receipt info: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Update Email Read Receipt Settings
Write-Host "`n=== Test 6: Update Email Read Receipt Settings ===" -ForegroundColor Cyan

$updateEmailSettingsData = @{
    enable_read_receipts = $true
    enable_expiration_alerts = $true
    expiration_alert_hours = 12
}

try {
    $updateSettingsResponse = Invoke-RestMethod -Uri "$ApiUrl/api/emails/$emailId/read-receipt-settings" -Method PUT -Headers $headers -Body ($updateEmailSettingsData | ConvertTo-Json)
    Write-Host "Email settings updated successfully" -ForegroundColor Green
} catch {
    Write-Host "Error updating email settings: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 7: Simulate Email Read (if recipient access is implemented)
Write-Host "`n=== Test 7: Simulate Email Read ===" -ForegroundColor Cyan

Write-Host "Note: This test requires a recipient user to access the email." -ForegroundColor Yellow
Write-Host "To test read receipts, you would need to:" -ForegroundColor Yellow
Write-Host "1. Create a recipient user account" -ForegroundColor Yellow
Write-Host "2. Access the email via /api/email/{id}/content endpoint" -ForegroundColor Yellow
Write-Host "3. Check that read receipt was sent to sender" -ForegroundColor Yellow

# Test 8: Get Read Events
Write-Host "`n=== Test 8: Get Read Events ===" -ForegroundColor Cyan

try {
    $readEvents = Invoke-RestMethod -Uri "$ApiUrl/api/emails/$emailId/read-events?limit=10" -Method GET -Headers $headers
    Write-Host "Read Events: $($readEvents | ConvertTo-Json)" -ForegroundColor Gray
} catch {
    Write-Host "Error getting read events: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 9: Test Expiration Worker (if running)
Write-Host "`n=== Test 9: Test Expiration Worker ===" -ForegroundColor Cyan

Write-Host "To test expiration alerts, you would need to:" -ForegroundColor Yellow
Write-Host "1. Start the expiration worker: go run cmd/expiration_worker/main.go" -ForegroundColor Yellow
Write-Host "2. Create emails with short expiration times" -ForegroundColor Yellow
Write-Host "3. Wait for expiration alerts to be sent" -ForegroundColor Yellow

# Test 10: Cleanup
Write-Host "`n=== Test 10: Cleanup ===" -ForegroundColor Cyan

try {
    $deleteResponse = Invoke-RestMethod -Uri "$ApiUrl/api/email/$emailId" -Method DELETE -Headers $headers
    Write-Host "Test email deleted successfully" -ForegroundColor Green
} catch {
    Write-Host "Error deleting test email: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "Read Receipts and Expiration Alerts functionality has been tested." -ForegroundColor White
Write-Host "Key features tested:" -ForegroundColor White
Write-Host "- User preference management" -ForegroundColor White
Write-Host "- Email-specific settings" -ForegroundColor White
Write-Host "- Read receipt info retrieval" -ForegroundColor White
Write-Host "- Read events tracking" -ForegroundColor White
Write-Host "`nNote: Actual read receipt sending and expiration alerts require:" -ForegroundColor Yellow
Write-Host "- Recipient user to access emails" -ForegroundColor Yellow
Write-Host "- Expiration worker to be running" -ForegroundColor Yellow
Write-Host "- Email/SMS delivery integration" -ForegroundColor Yellow







