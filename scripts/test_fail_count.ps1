# Test Fail Count Functionality
# This script tests the auto-delete after failed access attempts feature

param(
    [string]$ApiHost = "http://localhost:8080"
)

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
        subject = "Test Fail Count Email"
        body = "This is a test email for fail count functionality"
    } | ConvertTo-Json

    $response = Send-Request -Method "POST" -Url "$ApiHost/api/email/send" -Body $body

    if ($response.StatusCode -eq 200) {
        $responseData = $response.Body | ConvertFrom-Json
        return $responseData.email_id
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

        # Try to access the email with wrong credentials
        $body = @{
            email_id = $EmailId
        } | ConvertTo-Json

        $response = Send-Request -Method "POST" -Url "$ApiHost/api/email/get" -Body $body

        if ($response.StatusCode -eq 410) {
            Write-Output "Email has been deleted due to too many failed attempts!"
            return $true
        }
        elseif ($response.StatusCode -eq 403 -or $response.StatusCode -eq 401) {
            Write-Output "Access denied (expected for failed attempt)"
        }
        else {
            Write-Output "Unexpected response: $($response.StatusCode) - $($response.Body)"
        }

        Start-Sleep -Seconds 1
    }

    return $false
}

# Main test execution
Write-Output "=== Fail Count Functionality Test ==="

# Test 1: Create email and test fail count
Write-Output "`nTest 1: Creating test email"
$emailId = Create-TestEmail

if ($emailId) {
    Write-Output "Created test email with ID: $emailId"

    # Test 2: Simulate failed attempts until deletion
    Write-Output "`nTest 2: Simulating failed attempts until deletion"
    $deleted = Simulate-FailedAttempts -EmailId $emailId -Attempts 3

    if ($deleted) {
        Write-Output "✓ Fail count functionality working correctly"
    }
    else {
        Write-Output "✗ Fail count functionality not working as expected"
    }

    # Test 3: Verify email is no longer accessible
    Write-Output "`nTest 3: Verifying email is no longer accessible"
    $body = @{
        email_id = $emailId
    } | ConvertTo-Json

    $response = Send-Request -Method "POST" -Url "$ApiHost/api/email/get" -Body $body

    if ($response.StatusCode -eq 404) {
        Write-Output "✓ Email correctly returns 404 Not Found after deletion"
    }
    else {
        Write-Output "✗ Email still accessible after deletion: $($response.StatusCode)"
    }
}
else {
    Write-Output "✗ Failed to create test email"
}

# Test 4: Test with different number of attempts
Write-Output "`nTest 4: Testing with 2 attempts (should not trigger deletion)"
$emailId2 = Create-TestEmail

if ($emailId2) {
    Write-Output "Created second test email with ID: $emailId2"

    # Test with 2 attempts (should not trigger deletion with default limit of 3)
    $deleted = Simulate-FailedAttempts -EmailId $emailId2 -Attempts 2

    if (-not $deleted) {
        Write-Output "✓ Email correctly not deleted after 2 attempts"
    }
    else {
        Write-Output "✗ Email incorrectly deleted after 2 attempts"
    }
}

Write-Output "`n=== Test Complete ==="
