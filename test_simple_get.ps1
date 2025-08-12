# Simple test for GET /api/email/{id} endpoint
Write-Host "=== Simple GET Endpoint Test ===" -ForegroundColor Green

# Test configuration
$baseUrl = "http://localhost:8080"
$testEmail = "newtest@securesystem.email"
$testPassword = "TestPassword123!"

# Step 1: Generate TOTP and authenticate
Write-Host "Step 1: Authenticating..." -ForegroundColor Yellow
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
Write-Host "✅ Authentication successful" -ForegroundColor Green

# Step 2: Send test email
Write-Host "Step 2: Sending test email..." -ForegroundColor Yellow
$emailData = @{
    recipient = "test4@example.com"
    subject = "Simple Test Email"
    body = "This is a simple test email."
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "$baseUrl/api/email/send" -Method POST -Body $emailData -Headers $headers
$blobId = $response.blob_id
Write-Host "✅ Email sent successfully" -ForegroundColor Green
Write-Host "Blob ID: $blobId" -ForegroundColor Gray

# Step 3: Get email ID from database
Write-Host "Step 3: Getting email ID..." -ForegroundColor Yellow
$emailId = sqlite3 secure-email.db "SELECT email_id FROM emails WHERE encrypted_blob_url = '$blobId';"
Write-Host "Email ID: $emailId" -ForegroundColor Gray

# Step 4: Test GET endpoint
Write-Host "Step 4: Testing GET endpoint..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/email/$emailId" -Method GET -Headers $headers
    Write-Host "✅ GET endpoint successful!" -ForegroundColor Green
    Write-Host "Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Gray
} catch {
    Write-Host "❌ GET endpoint failed: $($_.Exception.Message)" -ForegroundColor Red
    
    # Try to get more error details
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "Status Code: $statusCode" -ForegroundColor Red
        
        # Try to read response body
        try {
            $stream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Body: $errorBody" -ForegroundColor Red
        } catch {
            Write-Host "Could not read error body" -ForegroundColor Yellow
        }
    }
}

# Step 5: Verify database record
Write-Host "Step 5: Verifying database record..." -ForegroundColor Yellow
$dbResult = sqlite3 secure-email.db "SELECT email_id, sender_id, recipient, subject, encrypted_blob_url FROM emails WHERE email_id = '$emailId';"
Write-Host "Database record: $dbResult" -ForegroundColor Gray

# Step 6: Test JOIN query manually
Write-Host "Step 6: Testing JOIN query manually..." -ForegroundColor Yellow
$joinResult = sqlite3 secure-email.db "SELECT e.encrypted_blob_url, e.encrypted_key, e.encryption_nonce, e.encryption_auth_tag, e.compression_algo, e.sender_id, e.recipient, e.subject, e.created_at, u.email as sender_email FROM emails e JOIN users u ON e.sender_id = u.id WHERE e.email_id = '$emailId';"
Write-Host "JOIN query result: $joinResult" -ForegroundColor Gray
