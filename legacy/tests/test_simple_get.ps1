# Simple test for GET /api/email/{id} endpoint
Write-Output "=== Simple GET Endpoint Test ==="

# Test configuration
$baseUrl = "http://localhost:8080"
$testEmail = "newtest@securesystem.email"
$testPassword = "TestPassword123!"

# Step 1: Generate TOTP and authenticate
Write-Output "Step 1: Authenticating..."
$totpCode = & .\totp_generator.exe "67TD4B73KBSUZ7TYIKAQSY7RFZEPJQXN"
$loginData = @{
    email = $testEmail
    password = $testPassword
    totp_code = $totpCode
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
$token = $response.access_token
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
Write-Output "✅ Authentication successful"

# Step 2: Send test email
Write-Output "Step 2: Sending test email..."
$emailData = @{
    recipient = "test4@example.com"
    subject = "Simple Test Email"
    body = "This is a simple test email."
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$baseUrl/api/email/send" -Method POST -Body $emailData -Headers $headers
$blobId = $response.blob_id
Write-Output "✅ Email sent successfully"
Write-Output "Blob ID: $blobId"

# Step 3: Get email ID from database
Write-Output "Step 3: Getting email ID..."
$emailId = sqlite3 secure-email.db "SELECT email_id FROM emails WHERE encrypted_blob_url = '$blobId';"
Write-Output "Email ID: $emailId"

# Step 4: Test GET endpoint
Write-Output "Step 4: Testing GET endpoint..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/email/$emailId" -Method GET -Headers $headers
    Write-Output "✅ GET endpoint successful!"
    Write-Output "Response: $($response | ConvertTo-Json -Depth 3)"
} catch {
    Write-Output "❌ GET endpoint failed: $($_.Exception.Message)"

    # Try to get more error details
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Output "Status Code: $statusCode"

        # Try to read response body
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $errorBody = $reader.ReadToEnd()
            Write-Output "Error Body: $errorBody"
        } catch {
            Write-Output "Could not read error body"
        }
    }
}

# Step 5: Verify database record
Write-Output "Step 5: Verifying database record..."
$dbResult = sqlite3 secure-email.db "SELECT email_id, sender_id, recipient, subject, encrypted_blob_url FROM emails WHERE email_id = '$emailId';"
Write-Output "Database record: $dbResult"

# Step 6: Test JOIN query manually
Write-Output "Step 6: Testing JOIN query manually..."
$joinResult = sqlite3 secure-email.db "SELECT e.encrypted_blob_url, e.encrypted_key, e.encryption_nonce, e.encryption_auth_tag, e.compression_algo, e.sender_id, e.recipient, e.subject, e.created_at, u.email as sender_email FROM emails e JOIN users u ON e.sender_id = u.id WHERE e.email_id = '$emailId';"
Write-Output "JOIN query result: $joinResult"
