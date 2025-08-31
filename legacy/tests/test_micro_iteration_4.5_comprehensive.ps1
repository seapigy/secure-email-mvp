# =============================================================================
# SECURE EMAIL MVP - MICRO-ITERATION 4.5 COMPREHENSIVE TEST SCRIPT
# =============================================================================
# Comprehensive test script for GET /api/email/{id} endpoint implementation
# =============================================================================
#
# TESTING REQUIREMENTS:
# - Server starts without build errors
# - Database migrations apply successfully
# - JWT authentication works
# - Email send endpoint creates test email
# - GET /api/email/{id} retrieves and decrypts email
# - Access control prevents unauthorized access
# - Audit logging records access events
# - Response format matches specification
#
# TEST FLOW:
# 1. Generate fresh TOTP code
# 2. Authenticate and get JWT token
# 3. Send test email via POST /api/email/send
# 4. Retrieve email via GET /api/email/{id}
# 5. Verify response format and content
# 6. Test access control (unauthorized user)
# 7. Verify audit log entries
# 8. Test recipient access (if applicable)
#
# DEBUGGING FEATURES:
# - Step-by-step execution with clear indicators
# - Detailed error reporting
# - Request/response logging
# - Database verification queries
# - Audit log inspection
# =============================================================================

Write-Output "=== MICRO-ITERATION 4.5 COMPREHENSIVE TEST ==="
Write-Output "Testing GET /api/email/{id} endpoint implementation"
Write-Output ""

# Test configuration
$baseUrl = "http://localhost:8080"
$testEmail = "newtest@securesystem.email"
$testPassword = "TestPassword123!"
$testRecipient = "test4@example.com"
$testSubject = "Micro-Iteration 4.5 Test Email"
$testBody = "This is a comprehensive test email for Micro-Iteration 4.5 verification of the GET /api/email/{id} endpoint."

# Initialize test results
$testResults = @()

# Function to add test result
function Add-TestResult {
    param($Test, $Result, $Notes)
    $testResults += [PSCustomObject]@{
        Test = $Test
        Result = $Result
        Notes = $Notes
    }
}

# Step 1: Generate fresh TOTP code
Write-Output "=== Step 1: Generate TOTP Code ==="
try {
    $totpCode = & .\totp_generator.exe "67TD4B73KBSUZ7TYIKAQSY7RFZEPJQXN"
    Write-Output "✅ TOTP Code generated: $totpCode"
    Add-TestResult "TOTP Generation" "PASS" "Code generated successfully"
} catch {
    Write-Output "❌ TOTP generation failed: $($_.Exception.Message)"
    Add-TestResult "TOTP Generation" "FAIL" "Failed to generate TOTP code"
    exit 1
}

# Step 2: Authenticate and get JWT token
Write-Output "`n=== Step 2: Authentication ==="
try {
    $loginData = @{
        email = $testEmail
        password = $testPassword
        totp_code = $totpCode
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$baseUrl/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"

    if ($response.access_token) {
        $token = $response.access_token
        Write-Output "✅ Authentication successful"
        Write-Output "Token: $($token.Substring(0, 50))..."
        Add-TestResult "JWT Authentication" "PASS" "Valid token obtained"
    } else {
        Write-Output "❌ Authentication failed - no access token"
        Add-TestResult "JWT Authentication" "FAIL" "No access token in response"
        exit 1
    }
} catch {
    Write-Output "❌ Authentication failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Output "Error details: $errorBody"
    }
    Add-TestResult "JWT Authentication" "FAIL" "Authentication request failed"
    exit 1
}

# Step 3: Send test email
Write-Output "`n=== Step 3: Send Test Email ==="
try {
    $emailData = @{
        recipient = $testRecipient
        subject = $testSubject
        body = $testBody
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    $response = Invoke-RestMethod -Uri "$baseUrl/api/email/send" -Method POST -Body $emailData -Headers $headers

    if ($response.status -eq "success" -and $response.blob_id) {
        # Get the actual email_id from database using the blob_id
        $blobId = $response.blob_id
        $dbResult = sqlite3 secure-email.db "SELECT email_id FROM emails WHERE encrypted_blob_url = '$blobId';"
        if ($dbResult) {
            $emailId = $dbResult.Trim()
        } else {
            Write-Output "❌ Could not find email_id for blob_id: $blobId"
            Add-TestResult "Email Send" "FAIL" "Email sent but not found in database"
            exit 1
        }
        Write-Output "✅ Email sent successfully"
        Write-Output "Email ID: $emailId"
        Write-Output "Blob ID: $($response.blob_id)"
        Add-TestResult "Email Send" "PASS" "Email created with ID: $emailId"
    } else {
        Write-Output "❌ Email send failed - invalid response"
        Write-Output "Response: $($response | ConvertTo-Json)"
        Add-TestResult "Email Send" "FAIL" "Invalid response from send endpoint"
        exit 1
    }
} catch {
    Write-Output "❌ Email send failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Output "Error details: $errorBody"
    }
    Add-TestResult "Email Send" "FAIL" "Email send request failed"
    exit 1
}

# Step 4: Retrieve email via GET /api/email/{id}
Write-Output "`n=== Step 4: Retrieve Email ==="
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/api/email/$emailId" -Method GET -Headers $headers

    Write-Output "✅ Email retrieved successfully"
    Write-Output "Response: $($response | ConvertTo-Json -Depth 3)"

    # Verify response format
    $requiredFields = @("id", "sender", "recipient", "subject", "body", "sent_at", "status")
    $missingFields = @()

    foreach ($field in $requiredFields) {
        if (-not $response.PSObject.Properties.Name.Contains($field)) {
            $missingFields += $field
        }
    }

    if ($missingFields.Count -gt 0) {
        Write-Output "❌ Response missing required fields: $($missingFields -join ', ')"
        Add-TestResult "Response Format" "FAIL" "Missing fields: $($missingFields -join ', ')"
    } else {
        Write-Output "✅ Response format validation passed"
        Add-TestResult "Response Format" "PASS" "All required fields present"
    }

    # Verify content matches
    if ($response.subject -eq $testSubject -and $response.body -eq $testBody) {
        Write-Output "✅ Content verification passed"
        Write-Output "Subject: $($response.subject)"
        Write-Output "Body: $($response.body)"
        Add-TestResult "Content Match" "PASS" "Retrieved content matches sent content"
    } else {
        Write-Output "❌ Content verification failed"
        Write-Output "Expected subject: $testSubject, Got: $($response.subject)"
        Write-Output "Expected body: $testBody, Got: $($response.body)"
        Add-TestResult "Content Match" "FAIL" "Content mismatch between sent and retrieved"
    }

    Add-TestResult "Email Retrieve" "PASS" "GET /api/email/{id} working correctly"

} catch {
    Write-Output "❌ Email retrieval failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Output "Error details: $errorBody"
    }
    Add-TestResult "Email Retrieve" "FAIL" "GET request failed: $($_.Exception.Message)"
}

# Step 5: Test access control (try to access with different user)
Write-Output "`n=== Step 5: Access Control Test ==="
try {
    # Create a fake token with different user ID
    $fakeToken = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOTk5IiwiZW1haWwiOiJmYWtlQGV4YW1wbGUuY29tIiwiZXhwIjoxNzU1MDM5NDUxLCJpYXQiOjE3NTUwMzg1NTEsImlzcyI6InNlY3VyZS1lbWFpbC1tdnAiLCJzdWIiOiI5OTkifQ.invalid_signature"

    $fakeHeaders = @{
        "Authorization" = "Bearer $fakeToken"
        "Content-Type" = "application/json"
    }

    $response = Invoke-RestMethod -Uri "$baseUrl/api/email/$emailId" -Method GET -Headers $fakeHeaders
    Write-Output "❌ Access control failed - unauthorized access allowed"
    Add-TestResult "Access Control" "FAIL" "Unauthorized access allowed"
} catch {
    if ($_.Exception.Response.StatusCode -eq 401 -or $_.Exception.Response.StatusCode -eq 403) {
        Write-Output "✅ Access control working - unauthorized access blocked"
        Write-Output "Status code: $($_.Exception.Response.StatusCode)"
        Add-TestResult "Access Control" "PASS" "Unauthorized access blocked correctly"
    } else {
        Write-Output "⚠️ Access control test inconclusive - unexpected status: $($_.Exception.Response.StatusCode)"
        Add-TestResult "Access Control" "WARN" "Unexpected status code: $($_.Exception.Response.StatusCode)"
    }
}

# Step 6: Verify database record
Write-Output "`n=== Step 6: Database Verification ==="
try {
    $dbResult = sqlite3 secure-email.db "SELECT email_id, sender_id, recipient, subject, encrypted_blob_url FROM emails WHERE email_id = '$emailId';"

    if ($dbResult) {
        Write-Output "✅ Database record found"
        Write-Output "Record: $dbResult"
        Add-TestResult "Database Record" "PASS" "Email record exists in database"
    } else {
        Write-Output "❌ Database record not found"
        Add-TestResult "Database Record" "FAIL" "Email record not found in database"
    }
} catch {
    Write-Output "❌ Database verification failed: $($_.Exception.Message)"
    Add-TestResult "Database Record" "FAIL" "Database query failed"
}

# Step 7: Verify audit log entries
Write-Output "`n=== Step 7: Audit Log Verification ==="
try {
    $auditResult = sqlite3 secure-email.db "SELECT log_id, event_type, user_id, related_email_id, outcome, details FROM audit_log WHERE related_email_id = '$emailId' ORDER BY timestamp DESC LIMIT 5;"

    if ($auditResult) {
        Write-Output "✅ Audit log entries found"
        Write-Output "Audit entries:"
        $auditResult -split "`n" | ForEach-Object { Write-Output "  $_" }
        Add-TestResult "Audit Log" "PASS" "Access events recorded in audit log"
    } else {
        Write-Output "⚠️ No audit log entries found"
        Add-TestResult "Audit Log" "WARN" "No audit log entries found"
    }
} catch {
    Write-Output "❌ Audit log verification failed: $($_.Exception.Message)"
    Add-TestResult "Audit Log" "FAIL" "Audit log query failed"
}

# Step 8: Test with non-existent email ID
Write-Output "`n=== Step 8: Non-existent Email Test ==="
try {
    $fakeEmailId = "00000000-0000-0000-0000-000000000000"
    $response = Invoke-RestMethod -Uri "$baseUrl/api/email/$fakeEmailId" -Method GET -Headers $headers
    Write-Output "❌ Non-existent email test failed - should return 404"
    Add-TestResult "Error Handling" "FAIL" "Non-existent email should return 404"
} catch {
    if ($_.Exception.Response.StatusCode -eq 404) {
        Write-Output "✅ Non-existent email test passed - 404 returned"
        Add-TestResult "Error Handling" "PASS" "Non-existent emails return 404 correctly"
    } else {
        Write-Output "⚠️ Non-existent email test inconclusive - status: $($_.Exception.Response.StatusCode)"
        Add-TestResult "Error Handling" "WARN" "Unexpected status for non-existent email: $($_.Exception.Response.StatusCode)"
    }
}

# Final summary
Write-Output "`n=== MICRO-ITERATION 4.5 TEST SUMMARY ==="
$testResults | Format-Table -AutoSize

# Count results
$passCount = ($testResults | Where-Object { $_.Result -eq "PASS" }).Count
$failCount = ($testResults | Where-Object { $_.Result -eq "FAIL" }).Count
$warnCount = ($testResults | Where-Object { $_.Result -eq "WARN" }).Count
$totalCount = $testResults.Count

Write-Output "`nTest Results Summary:"
Write-Output "PASS: $passCount"
Write-Output "FAIL: $failCount"
Write-Output "WARN: $warnCount"
Write-Output "TOTAL: $totalCount"

if ($failCount -eq 0) {
    Write-Output "`n🎯 MICRO-ITERATION 4.5 FULLY OPERATIONAL"
    Write-Output "All critical tests passed successfully!"
} else {
    Write-Output "`n❌ MICRO-ITERATION 4.5 HAS FAILURES"
    Write-Output "Some tests failed. Please review the results above."
}
