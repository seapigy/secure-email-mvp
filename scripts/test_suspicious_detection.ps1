# =============================================================================
# SECURE EMAIL MVP - SUSPICIOUS ACCESS DETECTION INTEGRATION TESTS
# =============================================================================
# Integration tests for Micro-Iteration 4.18: Suspicious Access Pattern Detection
# Tests all API endpoints and suspicious detection functionality
# =============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "TestPassword123!"
)

# Helper function to make API requests
function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [string]$Body = "",
        [string]$Token = ""
    )

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($Token) {
        $headers["Authorization"] = "Bearer $Token"
    }

    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl$Endpoint" -Method $Method -Headers $headers -Body $Body -ErrorAction Stop
        return @{
            Success = $true
            Data = $response
        }
    }
    catch {
        $errorResponse = $_.Exception.Response
        if ($errorResponse) {
            $reader = New-Object System.IO.StreamReader($errorResponse.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            $reader.Close()

            return @{
                Success = $false
                StatusCode = $errorResponse.StatusCode
                Error = $errorBody
            }
        }
        return @{
            Success = $false
            Error = $_.Exception.Message
        }
    }
}

# Test user registration and login
function Test-UserAuth {
    Write-Output "Testing user authentication..."

    # Register user
    $registerBody = @{
        email = $TestEmail
        password = $TestPassword
    } | ConvertTo-Json

    $registerResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $registerBody

    if (-not $registerResult.Success) {
        Write-Output "User registration failed: $($registerResult.Error)"
        return $null
    }

    Write-Output "User registered successfully"

    # Login user
    $loginBody = @{
        email = $TestEmail
        password = $TestPassword
    } | ConvertTo-Json

    $loginResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody

    if (-not $loginResult.Success) {
        Write-Output "User login failed: $($loginResult.Error)"
        return $null
    }

    Write-Output "User logged in successfully"
    return $loginResult.Data.token
}

# Test suspicious access detection preferences
function Test-SuspiciousPreferences {
    param([string]$Token)

    Write-Output "Testing suspicious access detection preferences..."

    # Get user preferences
    $getPrefsResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/preferences" -Token $Token

    if (-not $getPrefsResult.Success) {
        Write-Output "Failed to get user preferences: $($getPrefsResult.Error)"
        return $false
    }

    Write-Output "Current preferences: $($getPrefsResult.Data | ConvertTo-Json)"

    # Update user preferences
    $updatePrefsBody = @{
        enable_suspicious_detection = $true
        notify_on_suspicious_activity = $true
        auto_flag_suspicious_emails = $true
        minimum_severity_for_notification = "medium"
    } | ConvertTo-Json

    $updatePrefsResult = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/suspicious/preferences" -Body $updatePrefsBody -Token $Token

    if (-not $updatePrefsResult.Success) {
        Write-Output "Failed to update user preferences: $($updatePrefsResult.Error)"
        return $false
    }

    Write-Output "User preferences updated successfully"
    return $true
}

# Test detection rules
function Test-DetectionRules {
    param([string]$Token)

    Write-Output "Testing detection rules..."

    $rulesResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/rules" -Token $Token

    if (-not $rulesResult.Success) {
        Write-Output "Failed to get detection rules: $($rulesResult.Error)"
        return $false
    }

    Write-Output "Detection rules: $($rulesResult.Data | ConvertTo-Json)"

    if ($rulesResult.Data.Count -eq 0) {
        Write-Output "No detection rules found"
        return $false
    }

    Write-Output "Detection rules retrieved successfully"
    return $true
}

# Test suspicious emails list
function Test-SuspiciousEmails {
    param([string]$Token)

    Write-Output "Testing suspicious emails list..."

    $emailsResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/emails" -Token $Token

    if (-not $emailsResult.Success) {
        Write-Output "Failed to get suspicious emails: $($emailsResult.Error)"
        return $false
    }

    Write-Output "Suspicious emails: $($emailsResult.Data | ConvertTo-Json)"
    Write-Output "Suspicious emails retrieved successfully"
    return $true
}

# Test suspicious activity for a specific email
function Test-SuspiciousActivity {
    param([string]$Token, [string]$EmailID)

    Write-Output "Testing suspicious activity for email $EmailID..."

    $activityResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/activity/$EmailID" -Token $Token

    if (-not $activityResult.Success) {
        Write-Output "Failed to get suspicious activity: $($activityResult.Error)"
        return $false
    }

    Write-Output "Suspicious activity: $($activityResult.Data | ConvertTo-Json)"
    Write-Output "Suspicious activity retrieved successfully"
    return $true
}

# Test clearing suspicious flag
function Test-ClearSuspiciousFlag {
    param([string]$Token, [string]$EmailID)

    Write-Output "Testing clear suspicious flag for email $EmailID..."

    $clearBody = @{
        resolution_notes = "Test resolution - false positive"
    } | ConvertTo-Json

    $clearResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/suspicious/clear-flag/$EmailID" -Body $clearBody -Token $Token

    if (-not $clearResult.Success) {
        Write-Output "Failed to clear suspicious flag: $($clearResult.Error)"
        return $false
    }

    Write-Output "Suspicious flag cleared successfully"
    return $true
}

# Test resolving detection event
function Test-ResolveDetection {
    param([string]$Token, [string]$DetectionID)

    Write-Output "Testing resolve detection event $DetectionID..."

    $resolveBody = @{
        resolution_notes = "Test resolution - false positive"
    } | ConvertTo-Json

    $resolveResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/suspicious/resolve/$DetectionID" -Body $resolveBody -Token $Token

    if (-not $resolveResult.Success) {
        Write-Output "Failed to resolve detection event: $($resolveResult.Error)"
        return $false
    }

    Write-Output "Detection event resolved successfully"
    return $true
}

# Test creating and accessing emails to trigger suspicious detection
function Test-SuspiciousDetection {
    param([string]$Token)

    Write-Output "Testing suspicious access detection..."

    # Create a test email
    $createEmailBody = @{
        recipient = "recipient@example.com"
        subject = "Test Email for Suspicious Detection"
        content = "This is a test email for suspicious access detection testing."
    } | ConvertTo-Json

    $createResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/emails/send" -Body $createEmailBody -Token $Token

    if (-not $createResult.Success) {
        Write-Output "Failed to create test email: $($createResult.Error)"
        return $null
    }

    $emailID = $createResult.Data.email_id
    Write-Output "Test email created with ID: $emailID"

    # Simulate multiple failed access attempts to trigger detection
    Write-Output "Simulating multiple failed access attempts..."

    for ($i = 1; $i -le 4; $i++) {
        Write-Output "Attempt $i..."

        # Try to access the email with wrong credentials or from different IPs
        $accessResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/get" -Body "{}" -Token $Token

        # Wait a bit between attempts
        Start-Sleep -Seconds 2
    }

    # Check if suspicious activity was detected
    Start-Sleep -Seconds 5  # Give time for detection to process

    $activityResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/activity/$emailID" -Token $Token

    if ($activityResult.Success) {
        Write-Output "Suspicious activity detected: $($activityResult.Data | ConvertTo-Json)"

        # Test clearing the suspicious flag
        Test-ClearSuspiciousFlag -Token $Token -EmailID $emailID

        return $emailID
    } else {
        Write-Output "No suspicious activity detected yet"
        return $emailID
    }
}

# Main test execution
function Main {
    Write-Output "Starting Suspicious Access Detection Integration Tests"
    Write-Output "Base URL: $BaseUrl"
    Write-Output "Test Email: $TestEmail"
    Write-Output "=================================================="

    # Test user authentication
    $token = Test-UserAuth
    if (-not $token) {
        Write-Output "Authentication failed. Exiting tests."
        return
    }

    Write-Output "Authentication successful. Token: $($token.Substring(0, 20))..."

    # Test suspicious detection preferences
    if (-not (Test-SuspiciousPreferences -Token $token)) {
        Write-Output "Suspicious preferences test failed"
        return
    }

    # Test detection rules
    if (-not (Test-DetectionRules -Token $token)) {
        Write-Output "Detection rules test failed"
        return
    }

    # Test suspicious emails list
    if (-not (Test-SuspiciousEmails -Token $token)) {
        Write-Output "Suspicious emails test failed"
        return
    }

    # Test suspicious detection
    $emailID = Test-SuspiciousDetection -Token $token
    if ($emailID) {
        # Test suspicious activity for the email
        Test-SuspiciousActivity -Token $token -EmailID $emailID
    }

    Write-Output "=================================================="
    Write-Output "Suspicious Access Detection Integration Tests Completed"
}

# Run the tests
Main












