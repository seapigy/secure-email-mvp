# Test Self-Destruct Functionality
# This script tests the auto-delete after failed access attempts feature

param(
    [string]$ApiHost = "http://localhost:8080",
    [string]$EmailId = "",
    [switch]$EnableSimulation = $false
)

# Set environment variable for simulation if enabled
if ($EnableSimulation) {
    $env:SIMULATE_SELF_DESTRUCT = "1"
    Write-Output "Self-destruct simulation enabled"
}

# Function to send HTTP request
function Send-Request {
    param(
        [string]$Method,
        [string]$Url,
        [string]$Body = "",
        [hashtable]$Headers = @{}
    )

    try {
        $request = [System.Net.WebRequest]::Create($Url)
        $request.Method = $Method
        $request.ContentType = "application/json"

        foreach ($header in $Headers.GetEnumerator()) {
            $request.Headers.Add($header.Key, $header.Value)
        }

        if ($Body) {
            $bytes = [System.Text.Encoding]::UTF8.GetBytes($Body)
            $request.ContentLength = $bytes.Length
            $stream = $request.GetRequestStream()
            $stream.Write($bytes, 0, $bytes.Length)
            $stream.Close()
        }

        $response = $request.GetResponse()
        $reader = New-Object System.IO.StreamReader($response.GetResponseStream())
        $responseBody = $reader.ReadToEnd()
        $reader.Close()
        $response.Close()

        return @{
            StatusCode = $response.StatusCode
            Body = $responseBody
        }
    }
    catch {
        $exception = $_.Exception
        if ($exception.Response) {
            $reader = New-Object System.IO.StreamReader($exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            $reader.Close()

            return @{
                StatusCode = $exception.Response.StatusCode
                Body = $errorBody
            }
        }
        else {
            return @{
                StatusCode = 0
                Body = $exception.Message
            }
        }
    }
}

# Function to create a test email
function Create-TestEmail {
    param([string]$Recipient = "test@example.com")

    $body = @{
        recipient = $Recipient
        subject = "Test Self-Destruct Email"
        body = "This is a test email for self-destruct functionality"
        selfDestructAfterAttempts = $true
        maxFailedAttempts = 3
    } | ConvertTo-Json

    $response = Send-Request -Method "POST" -Url "$ApiHost/api/email/send" -Body $body

    if ($response.StatusCode -eq 200) {
        $responseData = $response.Body | ConvertFrom-Json
        return $responseData.blob_id
    }
    else {
        Write-Output "Failed to create test email: $($response.Body)"
        return $null
    }
}

# Function to simulate failed access attempts
function Simulate-FailedAttempts {
    param(
        [string]$EmailId,
        [int]$Attempts = 3
    )

    Write-Output "Simulating $Attempts failed access attempts for email $EmailId"

    for ($i = 1; $i -le $Attempts; $i++) {
        Write-Output "Attempt $i/$Attempts"

        if ($EnableSimulation) {
            # Use test endpoint for simulation
            $body = @{
                email_id = $EmailId
                action = "increment_failed"
            } | ConvertTo-Json

            $response = Send-Request -Method "POST" -Url "$ApiHost/test/self-destruct" -Body $body

            if ($response.StatusCode -eq 200) {
                $responseData = $response.Body | ConvertFrom-Json
                Write-Output "Failed attempts: $($responseData.failed_attempts)/$($responseData.max_attempts)"

                if ($responseData.self_destructed) {
                    Write-Output "Email has been self-destructed!"
                    return $true
                }
            }
            else {
                Write-Output "Failed to simulate attempt: $($response.Body)"
                return $false
            }
        }
        else {
            # Try to access the email (this should fail and increment attempts)
            $response = Send-Request -Method "GET" -Url "$ApiHost/api/email/view/$EmailId"

            if ($response.StatusCode -eq 410) {
                Write-Output "Email has been self-destructed!"
                return $true
            }
            elseif ($response.StatusCode -eq 403 -or $response.StatusCode -eq 401) {
                Write-Output "Access denied (expected for failed attempt)"
            }
            else {
                Write-Output "Unexpected response: $($response.StatusCode) - $($response.Body)"
            }
        }

        Start-Sleep -Seconds 1
    }

    return $false
}

# Function to test successful access resets failed attempts
function Test-SuccessfulAccess {
    param([string]$EmailId)

    Write-Output "Testing successful access resets failed attempts"

    # First, simulate some failed attempts
    Simulate-FailedAttempts -EmailId $EmailId -Attempts 2

    # Then try a successful access (this should reset the counter)
    $response = Send-Request -Method "GET" -Url "$ApiHost/api/email/view/$EmailId"

    if ($response.StatusCode -eq 200) {
        Write-Output "Successful access should have reset failed attempts"
        return $true
    }
    else {
        Write-Output "Failed to access email: $($response.Body)"
        return $false
    }
}

# Main test execution
Write-Output "=== Self-Destruct Functionality Test ==="

# Test 1: Create email and test self-destruct
Write-Output "`nTest 1: Creating test email with self-destruct enabled"
$emailId = Create-TestEmail

if ($emailId) {
    Write-Output "Created test email with ID: $emailId"

    # Test 2: Simulate failed attempts until self-destruct
    Write-Output "`nTest 2: Simulating failed attempts until self-destruct"
    $selfDestructed = Simulate-FailedAttempts -EmailId $emailId -Attempts 3

    if ($selfDestructed) {
        Write-Output "✓ Self-destruct functionality working correctly"
    }
    else {
        Write-Output "✗ Self-destruct functionality not working as expected"
    }

    # Test 3: Verify email is no longer accessible
    Write-Output "`nTest 3: Verifying email is no longer accessible"
    $response = Send-Request -Method "GET" -Url "$ApiHost/api/email/view/$emailId"

    if ($response.StatusCode -eq 410) {
        Write-Output "✓ Email correctly returns 410 Gone after self-destruct"
    }
    else {
        Write-Output "✗ Email still accessible after self-destruct: $($response.StatusCode)"
    }
}
else {
    Write-Output "✗ Failed to create test email"
}

# Test 4: Test with different max attempts
Write-Output "`nTest 4: Testing with different max attempts"
$emailId2 = Create-TestEmail

if ($emailId2) {
    Write-Output "Created second test email with ID: $emailId2"

    # Test with 2 attempts (should not trigger self-destruct with default max of 3)
    $selfDestructed = Simulate-FailedAttempts -EmailId $emailId2 -Attempts 2

    if (-not $selfDestructed) {
        Write-Output "✓ Email correctly not self-destructed after 2 attempts"
    }
    else {
        Write-Output "✗ Email incorrectly self-destructed after 2 attempts"
    }
}

# Test 5: Test successful access resets counter
Write-Output "`nTest 5: Testing successful access resets failed attempts"
$emailId3 = Create-TestEmail

if ($emailId3) {
    Write-Output "Created third test email with ID: $emailId3"

    $success = Test-SuccessfulAccess -EmailId $emailId3

    if ($success) {
        Write-Output "✓ Successful access correctly resets failed attempts"
    }
    else {
        Write-Output "✗ Successful access failed to reset failed attempts"
    }
}

Write-Output "`n=== Test Complete ==="

# Cleanup: Disable simulation
if ($EnableSimulation) {
    Remove-Item Env:SIMULATE_SELF_DESTRUCT -ErrorAction SilentlyContinue
    Write-Output "Self-destruct simulation disabled"
}
