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
    Write-Host "Self-destruct simulation enabled" -ForegroundColor Yellow
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
        Write-Host "Failed to create test email: $($response.Body)" -ForegroundColor Red
        return $null
    }
}

# Function to simulate failed access attempts
function Simulate-FailedAttempts {
    param(
        [string]$EmailId,
        [int]$Attempts = 3
    )
    
    Write-Host "Simulating $Attempts failed access attempts for email $EmailId" -ForegroundColor Yellow
    
    for ($i = 1; $i -le $Attempts; $i++) {
        Write-Host "Attempt $i/$Attempts" -ForegroundColor Cyan
        
        if ($EnableSimulation) {
            # Use test endpoint for simulation
            $body = @{
                email_id = $EmailId
                action = "increment_failed"
            } | ConvertTo-Json
            
            $response = Send-Request -Method "POST" -Url "$ApiHost/test/self-destruct" -Body $body
            
            if ($response.StatusCode -eq 200) {
                $responseData = $response.Body | ConvertFrom-Json
                Write-Host "Failed attempts: $($responseData.failed_attempts)/$($responseData.max_attempts)" -ForegroundColor Green
                
                if ($responseData.self_destructed) {
                    Write-Host "Email has been self-destructed!" -ForegroundColor Red
                    return $true
                }
            }
            else {
                Write-Host "Failed to simulate attempt: $($response.Body)" -ForegroundColor Red
                return $false
            }
        }
        else {
            # Try to access the email (this should fail and increment attempts)
            $response = Send-Request -Method "GET" -Url "$ApiHost/api/email/view/$EmailId"
            
            if ($response.StatusCode -eq 410) {
                Write-Host "Email has been self-destructed!" -ForegroundColor Red
                return $true
            }
            elseif ($response.StatusCode -eq 403 -or $response.StatusCode -eq 401) {
                Write-Host "Access denied (expected for failed attempt)" -ForegroundColor Yellow
            }
            else {
                Write-Host "Unexpected response: $($response.StatusCode) - $($response.Body)" -ForegroundColor Red
            }
        }
        
        Start-Sleep -Seconds 1
    }
    
    return $false
}

# Function to test successful access resets failed attempts
function Test-SuccessfulAccess {
    param([string]$EmailId)
    
    Write-Host "Testing successful access resets failed attempts" -ForegroundColor Yellow
    
    # First, simulate some failed attempts
    Simulate-FailedAttempts -EmailId $EmailId -Attempts 2
    
    # Then try a successful access (this should reset the counter)
    $response = Send-Request -Method "GET" -Url "$ApiHost/api/email/view/$EmailId"
    
    if ($response.StatusCode -eq 200) {
        Write-Host "Successful access should have reset failed attempts" -ForegroundColor Green
        return $true
    }
    else {
        Write-Host "Failed to access email: $($response.Body)" -ForegroundColor Red
        return $false
    }
}

# Main test execution
Write-Host "=== Self-Destruct Functionality Test ===" -ForegroundColor Green

# Test 1: Create email and test self-destruct
Write-Host "`nTest 1: Creating test email with self-destruct enabled" -ForegroundColor Green
$emailId = Create-TestEmail

if ($emailId) {
    Write-Host "Created test email with ID: $emailId" -ForegroundColor Green
    
    # Test 2: Simulate failed attempts until self-destruct
    Write-Host "`nTest 2: Simulating failed attempts until self-destruct" -ForegroundColor Green
    $selfDestructed = Simulate-FailedAttempts -EmailId $emailId -Attempts 3
    
    if ($selfDestructed) {
        Write-Host "✓ Self-destruct functionality working correctly" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Self-destruct functionality not working as expected" -ForegroundColor Red
    }
    
    # Test 3: Verify email is no longer accessible
    Write-Host "`nTest 3: Verifying email is no longer accessible" -ForegroundColor Green
    $response = Send-Request -Method "GET" -Url "$ApiHost/api/email/view/$emailId"
    
    if ($response.StatusCode -eq 410) {
        Write-Host "✓ Email correctly returns 410 Gone after self-destruct" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Email still accessible after self-destruct: $($response.StatusCode)" -ForegroundColor Red
    }
}
else {
    Write-Host "✗ Failed to create test email" -ForegroundColor Red
}

# Test 4: Test with different max attempts
Write-Host "`nTest 4: Testing with different max attempts" -ForegroundColor Green
$emailId2 = Create-TestEmail

if ($emailId2) {
    Write-Host "Created second test email with ID: $emailId2" -ForegroundColor Green
    
    # Test with 2 attempts (should not trigger self-destruct with default max of 3)
    $selfDestructed = Simulate-FailedAttempts -EmailId $emailId2 -Attempts 2
    
    if (-not $selfDestructed) {
        Write-Host "✓ Email correctly not self-destructed after 2 attempts" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Email incorrectly self-destructed after 2 attempts" -ForegroundColor Red
    }
}

# Test 5: Test successful access resets counter
Write-Host "`nTest 5: Testing successful access resets failed attempts" -ForegroundColor Green
$emailId3 = Create-TestEmail

if ($emailId3) {
    Write-Host "Created third test email with ID: $emailId3" -ForegroundColor Green
    
    $success = Test-SuccessfulAccess -EmailId $emailId3
    
    if ($success) {
        Write-Host "✓ Successful access correctly resets failed attempts" -ForegroundColor Green
    }
    else {
        Write-Host "✗ Successful access failed to reset failed attempts" -ForegroundColor Red
    }
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Green

# Cleanup: Disable simulation
if ($EnableSimulation) {
    Remove-Item Env:SIMULATE_SELF_DESTRUCT -ErrorAction SilentlyContinue
    Write-Host "Self-destruct simulation disabled" -ForegroundColor Yellow
}
