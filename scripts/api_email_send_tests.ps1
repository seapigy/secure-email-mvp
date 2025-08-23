# PowerShell script to automate Secure Email API tests for /api/email/send

function Invoke-ApiTest {
    param(
        [string]$TestName,
        [string]$Uri,
        [string]$Method = "POST",
        [string]$Body = $null,
        [int]$ExpectedStatus = 200
    )
    Write-Output "\n=== Running Test: $TestName ==="
    Write-Output "Request URI: $Uri"
    Write-Output "Request Method: $Method"
    Write-Output "Request Body: $Body"
    try {
        $response = Invoke-WebRequest -Uri $Uri -Method $Method -ContentType "application/json" -Body $Body -ErrorAction Stop
        $status = $response.StatusCode
        $respBody = $response.Content
    } catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse -is [System.Net.HttpWebResponse]) {
            $status = $errorResponse.StatusCode.value__
            $stream = $errorResponse.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($stream)
            $respBody = $reader.ReadToEnd()
            $reader.Close()
        } else {
            $status = 0
            $respBody = $_.Exception.Message
        }
    }
    Write-Output "Response Status: $status"
    Write-Output "Response Body: $respBody"
    if ($status -eq $ExpectedStatus) {
        Write-Output "Test passed"
    } else {
        Write-Output "Test failed: Expected $ExpectedStatus, got $status"
    }
}

# 1. Valid POST
$validBody = @{
    sender_id = "test-sender"
    recipient = "test@example.com"
    subject = "Test Subject"
    body = "This is a test email body."
} | ConvertTo-Json

Invoke-ApiTest -TestName "Valid POST" -Uri "http://localhost:8080/api/email/send" -Body $validBody -ExpectedStatus 200

# 2. Missing each required field
$fields = @("sender_id", "recipient", "subject", "body")
foreach ($field in $fields) {
    $body = @{
        sender_id = "test-sender"
        recipient = "test@example.com"
        subject = "Test Subject"
        body = "This is a test email body."
    }
    $body.Remove($field)
    $jsonBody = $body | ConvertTo-Json
    Invoke-ApiTest -TestName "Missing field: $field" -Uri "http://localhost:8080/api/email/send" -Body $jsonBody -ExpectedStatus 400
}

# 3. Empty string values for required fields
foreach ($field in $fields) {
    $body = @{
        sender_id = "test-sender"
        recipient = "test@example.com"
        subject = "Test Subject"
        body = "This is a test email body."
    }
    $body[$field] = ""
    $jsonBody = $body | ConvertTo-Json
    Invoke-ApiTest -TestName "Empty value: $field" -Uri "http://localhost:8080/api/email/send" -Body $jsonBody -ExpectedStatus 400
}

# 4. Malformed JSON payloads
$malformedJsons = @(
    '{"sender_id": "test-sender", "recipient": "test@example.com", "subject": "Test Subject", "body": "Missing end brace"',
    '{sender_id: "test-sender", recipient: "test@example.com", subject: "Test Subject", body: "No quotes on keys"}',
    '{"sender_id": "test-sender", "recipient": "test@example.com", "subject": "Test Subject" "body": "Missing comma"}'
)
foreach ($malformed in $malformedJsons) {
    Invoke-ApiTest -TestName "Malformed JSON" -Uri "http://localhost:8080/api/email/send" -Body $malformed -ExpectedStatus 400
}

# 5. Invalid email format in recipient
$invalidEmailBody = @{
    sender_id = "test-sender"
    recipient = "not-an-email"
    subject = "Test Subject"
    body = "This is a test email body."
} | ConvertTo-Json
Invoke-ApiTest -TestName "Invalid recipient email format" -Uri "http://localhost:8080/api/email/send" -Body $invalidEmailBody -ExpectedStatus 400

# =====================
# Signup Endpoint Tests
# =====================

$signupUri = "http://localhost:8080/api/auth/signup"

# 1. Valid signup
$signupValid = @{
    email = "testuser1@example.com"
    password = "StrongPassword123!"
    fallback_email = "fallback1@example.com"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Signup: Valid" -Uri $signupUri -Body $signupValid -ExpectedStatus 201

# 2. Missing fields
$signupFields = @("email", "password", "fallback_email")
foreach ($field in $signupFields) {
    $body = @{
        email = "testuser2@example.com"
        password = "StrongPassword123!"
        fallback_email = "fallback2@example.com"
    }
    $body.Remove($field)
    $jsonBody = $body | ConvertTo-Json
    Invoke-ApiTest -TestName "Signup: Missing field $field" -Uri $signupUri -Body $jsonBody -ExpectedStatus 400
}

# 3. Invalid email
$signupInvalidEmail = @{
    email = "not-an-email"
    password = "StrongPassword123!"
    fallback_email = "fallback3@example.com"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Signup: Invalid email" -Uri $signupUri -Body $signupInvalidEmail -ExpectedStatus 400

# 4. Invalid fallback email
$signupInvalidFallback = @{
    email = "testuser4@example.com"
    password = "StrongPassword123!"
    fallback_email = "not-an-email"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Signup: Invalid fallback email" -Uri $signupUri -Body $signupInvalidFallback -ExpectedStatus 400

# 5. Weak password
$signupWeakPassword = @{
    email = "testuser5@example.com"
    password = "123"
    fallback_email = "fallback5@example.com"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Signup: Weak password" -Uri $signupUri -Body $signupWeakPassword -ExpectedStatus 400

# 6. Duplicate user
$signupDuplicate = @{
    email = "testuser1@example.com"
    password = "AnotherStrongPassword!"
    fallback_email = "fallback1@example.com"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Signup: Duplicate user" -Uri $signupUri -Body $signupDuplicate -ExpectedStatus 400

# =====================
# Login Endpoint Tests
# =====================

$loginUri = "http://localhost:8080/api/auth/login"

# 1. Valid login (assumes testuser1@example.com was created and fallback is confirmed)
$loginValid = @{
    email = "testuser1@example.com"
    password = "StrongPassword123!"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Login: Valid" -Uri $loginUri -Body $loginValid -ExpectedStatus 200

# 2. Missing fields
$loginFields = @("email", "password")
foreach ($field in $loginFields) {
    $body = @{
        email = "testuser1@example.com"
        password = "StrongPassword123!"
    }
    $body.Remove($field)
    $jsonBody = $body | ConvertTo-Json
    Invoke-ApiTest -TestName "Login: Missing field $field" -Uri $loginUri -Body $jsonBody -ExpectedStatus 400
}

# 3. Invalid credentials (wrong password)
$loginWrongPassword = @{
    email = "testuser1@example.com"
    password = "WrongPassword!"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Login: Wrong password" -Uri $loginUri -Body $loginWrongPassword -ExpectedStatus 401

# 4. Invalid credentials (non-existent user)
$loginNoUser = @{
    email = "doesnotexist@example.com"
    password = "AnyPassword123!"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Login: Non-existent user" -Uri $loginUri -Body $loginNoUser -ExpectedStatus 401

# 5. Fallback not confirmed (simulate by registering a new user and not confirming fallback)
$loginUnconfirmed = @{
    email = "testuser6@example.com"
    password = "StrongPassword123!"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Login: Fallback not confirmed" -Uri $loginUri -Body $loginUnconfirmed -ExpectedStatus 403

# 6. Malformed JSON
$loginMalformed = '{"email": "testuser1@example.com", "password": "StrongPassword123!"' # missing closing brace
Invoke-ApiTest -TestName "Login: Malformed JSON" -Uri $loginUri -Body $loginMalformed -ExpectedStatus 400

# =====================
# Login Rate Limiting / Account Lockout Tests
# =====================

# These tests assume LOGIN_RATE_LIMIT_ENABLED=1 and testuser7@example.com exists with fallback confirmed
$lockoutEmail = "testuser7@example.com"
$lockoutPassword = "StrongPassword123!"
$loginUri = "http://localhost:8080/api/auth/login"

# Ensure user exists (signup if needed)
$signupLockout = @{
    email = $lockoutEmail
    password = $lockoutPassword
    fallback_email = "fallback7@example.com"
} | ConvertTo-Json
Invoke-ApiTest -TestName "Lockout: Ensure user exists" -Uri "http://localhost:8080/api/auth/signup" -Body $signupLockout -ExpectedStatus 201

# 1. Exceed failed login attempts to trigger lockout
for ($i = 1; $i -le 5; $i++) {
    $wrongLogin = @{
        email = $lockoutEmail
        password = "WrongPassword!"
    } | ConvertTo-Json
    Invoke-ApiTest -TestName "Lockout: Failed attempt $i" -Uri $loginUri -Body $wrongLogin -ExpectedStatus 401
}

# 2. Attempt login after lockout triggered
$lockedLogin = @{
    email = $lockoutEmail
    password = $lockoutPassword
} | ConvertTo-Json
Invoke-ApiTest -TestName "Lockout: Login during lockout" -Uri $loginUri -Body $lockedLogin -ExpectedStatus 429

# 3. (Optional) Wait for lockout to expire and test login again
Write-Output "If you want to test login after lockout expires, wait 15 minutes, then press Enter to continue."
Read-Host
Invoke-ApiTest -TestName "Lockout: Login after lockout expires" -Uri $loginUri -Body $lockedLogin -ExpectedStatus 200

Write-Output "\n=== Simulated DB Failure Test ==="
Write-Output "To run this test, stop your API server, set the environment variable SIMULATE_DB_FAILURE=1, and restart the server."
Write-Output "Example (PowerShell): $env:SIMULATE_DB_FAILURE = '1'"
Write-Output "Then restart your API server, and press Enter to continue."
Read-Host

Invoke-ApiTest -TestName "Simulated DB insert failure" -Uri "http://localhost:8080/api/email/send" -Body $validBody -ExpectedStatus 500

Write-Output "\nNow unset the environment variable and restart your API server to restore normal behavior."
Write-Output "Example (PowerShell): Remove-Item Env:SIMULATE_DB_FAILURE"
Write-Output "Press Enter after you have reverted and restarted the server."
Read-Host

Invoke-ApiTest -TestName "Post-DB failure revert (should succeed)" -Uri "http://localhost:8080/api/email/send" -Body $validBody -ExpectedStatus 200
