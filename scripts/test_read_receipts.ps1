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

Write-Output "Starting Read Receipts and Expiration Alerts Test..."
Write-Output "API URL: $ApiUrl"

# Test 1: User Registration and Login
Write-Output "`n=== Test 1: User Registration and Login ==="

$signupData = @{
    email = $TestEmail
    password = $TestPassword
}

$signupResponse = Invoke-RestMethod -Uri "$ApiUrl/api/auth/signup" -Method POST -Body ($signupData | ConvertTo-Json) -ContentType "application/json"
Write-Output "Signup Response: $($signupResponse | ConvertTo-Json)"

# Login to get JWT token
$loginData = @{
    email = $TestEmail
    password = $TestPassword
    totp_code = "123456"  # Default TOTP for testing
}

$loginResponse = Invoke-RestMethod -Uri "$ApiUrl/api/auth/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json"
$jwtToken = $loginResponse.token
$userId = $loginResponse.user_id

Write-Output "Login successful. User ID: $userId"

# Test 2: Get Read Receipt Preferences
Write-Output "`n=== Test 2: Get Read Receipt Preferences ==="

$headers = @{
    "Authorization" = "Bearer $jwtToken"
    "Content-Type" = "application/json"
}

try {
    $prefsResponse = Invoke-RestMethod -Uri "$ApiUrl/api/read-receipts/preferences" -Method GET -Headers $headers
    Write-Output "Read Receipt Preferences: $($prefsResponse | ConvertTo-Json)"
} catch {
    Write-Output "Error getting preferences: $($_.Exception.Message)"
}

# Test 3: Update Read Receipt Preferences
Write-Output "`n=== Test 3: Update Read Receipt Preferences ==="

$updatePrefsData = @{
    enable_read_receipts = $true
    enable_expiration_alerts = $true
    expiration_alert_hours = 48
    delivery_methods = "email,sms"
}

try {
    $updatePrefsResponse = Invoke-RestMethod -Uri "$ApiUrl/api/read-receipts/preferences" -Method PUT -Headers $headers -Body ($updatePrefsData | ConvertTo-Json)
    Write-Output "Updated Preferences: $($updatePrefsResponse | ConvertTo-Json)"
} catch {
    Write-Output "Error updating preferences: $($_.Exception.Message)"
}

# Test 4: Send Email with Read Receipt Settings
Write-Output "`n=== Test 4: Send Email with Read Receipt Settings ==="

$sendEmailData = @{
    recipient = "recipient@example.com"
    subject = "Test Email with Read Receipts"
    body = "This is a test email to verify read receipt functionality."
    expires_at = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
}

try {
    $sendResponse = Invoke-RestMethod -Uri "$ApiUrl/api/email/send" -Method POST -Headers $headers -Body ($sendEmailData | ConvertTo-Json)
    $emailId = $sendResponse.blob_id
    Write-Output "Email sent successfully. Email ID: $emailId"
} catch {
    Write-Output "Error sending email: $($_.Exception.Message)"
    exit 1
}

# Test 5: Get Email Read Receipt Info
Write-Output "`n=== Test 5: Get Email Read Receipt Info ==="

try {
    $readReceiptInfo = Invoke-RestMethod -Uri "$ApiUrl/api/emails/$emailId/read-receipts" -Method GET -Headers $headers
    Write-Output "Read Receipt Info: $($readReceiptInfo | ConvertTo-Json)"
} catch {
    Write-Output "Error getting read receipt info: $($_.Exception.Message)"
}

# Test 6: Update Email Read Receipt Settings
Write-Output "`n=== Test 6: Update Email Read Receipt Settings ==="

$updateEmailSettingsData = @{
    enable_read_receipts = $true
    enable_expiration_alerts = $true
    expiration_alert_hours = 12
}

try {
    $updateSettingsResponse = Invoke-RestMethod -Uri "$ApiUrl/api/emails/$emailId/read-receipt-settings" -Method PUT -Headers $headers -Body ($updateEmailSettingsData | ConvertTo-Json)
    Write-Output "Email settings updated successfully"
} catch {
    Write-Output "Error updating email settings: $($_.Exception.Message)"
}

# Test 7: Simulate Email Read (if recipient access is implemented)
Write-Output "`n=== Test 7: Simulate Email Read ==="

Write-Output "Note: This test requires a recipient user to access the email."
Write-Output "To test read receipts, you would need to:"
Write-Output "1. Create a recipient user account"
Write-Output "2. Access the email via /api/email/{id}/content endpoint"
Write-Output "3. Check that read receipt was sent to sender"

# Test 8: Get Read Events
Write-Output "`n=== Test 8: Get Read Events ==="

try {
    $readEvents = Invoke-RestMethod -Uri "$ApiUrl/api/emails/$emailId/read-events?limit=10" -Method GET -Headers $headers
    Write-Output "Read Events: $($readEvents | ConvertTo-Json)"
} catch {
    Write-Output "Error getting read events: $($_.Exception.Message)"
}

# Test 9: Test Expiration Worker (if running)
Write-Output "`n=== Test 9: Test Expiration Worker ==="

Write-Output "To test expiration alerts, you would need to:"
Write-Output "1. Start the expiration worker: go run cmd/expiration_worker/main.go"
Write-Output "2. Create emails with short expiration times"
Write-Output "3. Wait for expiration alerts to be sent"

# Test 10: Cleanup
Write-Output "`n=== Test 10: Cleanup ==="

try {
    $deleteResponse = Invoke-RestMethod -Uri "$ApiUrl/api/email/$emailId" -Method DELETE -Headers $headers
    Write-Output "Test email deleted successfully"
} catch {
    Write-Output "Error deleting test email: $($_.Exception.Message)"
}

Write-Output "`n=== Test Summary ==="
Write-Output "Read Receipts and Expiration Alerts functionality has been tested."
Write-Output "Key features tested:"
Write-Output "- User preference management"
Write-Output "- Email-specific settings"
Write-Output "- Read receipt info retrieval"
Write-Output "- Read events tracking"
Write-Output "`nNote: Actual read receipt sending and expiration alerts require:"
Write-Output "- Recipient user to access emails"
Write-Output "- Expiration worker to be running"
Write-Output "- Email/SMS delivery integration"





















