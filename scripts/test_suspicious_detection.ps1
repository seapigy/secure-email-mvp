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
    Write-Host "Testing user authentication..." -ForegroundColor Green
    
    # Register user
    $registerBody = @{
        email = $TestEmail
        password = $TestPassword
    } | ConvertTo-Json
    
    $registerResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/signup" -Body $registerBody
    
    if (-not $registerResult.Success) {
        Write-Host "User registration failed: $($registerResult.Error)" -ForegroundColor Red
        return $null
    }
    
    Write-Host "User registered successfully" -ForegroundColor Green
    
    # Login user
    $loginBody = @{
        email = $TestEmail
        password = $TestPassword
    } | ConvertTo-Json
    
    $loginResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginBody
    
    if (-not $loginResult.Success) {
        Write-Host "User login failed: $($loginResult.Error)" -ForegroundColor Red
        return $null
    }
    
    Write-Host "User logged in successfully" -ForegroundColor Green
    return $loginResult.Data.token
}

# Test suspicious access detection preferences
function Test-SuspiciousPreferences {
    param([string]$Token)
    
    Write-Host "Testing suspicious access detection preferences..." -ForegroundColor Green
    
    # Get user preferences
    $getPrefsResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/preferences" -Token $Token
    
    if (-not $getPrefsResult.Success) {
        Write-Host "Failed to get user preferences: $($getPrefsResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Current preferences: $($getPrefsResult.Data | ConvertTo-Json)" -ForegroundColor Yellow
    
    # Update user preferences
    $updatePrefsBody = @{
        enable_suspicious_detection = $true
        notify_on_suspicious_activity = $true
        auto_flag_suspicious_emails = $true
        minimum_severity_for_notification = "medium"
    } | ConvertTo-Json
    
    $updatePrefsResult = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/suspicious/preferences" -Body $updatePrefsBody -Token $Token
    
    if (-not $updatePrefsResult.Success) {
        Write-Host "Failed to update user preferences: $($updatePrefsResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "User preferences updated successfully" -ForegroundColor Green
    return $true
}

# Test detection rules
function Test-DetectionRules {
    param([string]$Token)
    
    Write-Host "Testing detection rules..." -ForegroundColor Green
    
    $rulesResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/rules" -Token $Token
    
    if (-not $rulesResult.Success) {
        Write-Host "Failed to get detection rules: $($rulesResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Detection rules: $($rulesResult.Data | ConvertTo-Json)" -ForegroundColor Yellow
    
    if ($rulesResult.Data.Count -eq 0) {
        Write-Host "No detection rules found" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Detection rules retrieved successfully" -ForegroundColor Green
    return $true
}

# Test suspicious emails list
function Test-SuspiciousEmails {
    param([string]$Token)
    
    Write-Host "Testing suspicious emails list..." -ForegroundColor Green
    
    $emailsResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/emails" -Token $Token
    
    if (-not $emailsResult.Success) {
        Write-Host "Failed to get suspicious emails: $($emailsResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Suspicious emails: $($emailsResult.Data | ConvertTo-Json)" -ForegroundColor Yellow
    Write-Host "Suspicious emails retrieved successfully" -ForegroundColor Green
    return $true
}

# Test suspicious activity for a specific email
function Test-SuspiciousActivity {
    param([string]$Token, [string]$EmailID)
    
    Write-Host "Testing suspicious activity for email $EmailID..." -ForegroundColor Green
    
    $activityResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/activity/$EmailID" -Token $Token
    
    if (-not $activityResult.Success) {
        Write-Host "Failed to get suspicious activity: $($activityResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Suspicious activity: $($activityResult.Data | ConvertTo-Json)" -ForegroundColor Yellow
    Write-Host "Suspicious activity retrieved successfully" -ForegroundColor Green
    return $true
}

# Test clearing suspicious flag
function Test-ClearSuspiciousFlag {
    param([string]$Token, [string]$EmailID)
    
    Write-Host "Testing clear suspicious flag for email $EmailID..." -ForegroundColor Green
    
    $clearBody = @{
        resolution_notes = "Test resolution - false positive"
    } | ConvertTo-Json
    
    $clearResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/suspicious/clear-flag/$EmailID" -Body $clearBody -Token $Token
    
    if (-not $clearResult.Success) {
        Write-Host "Failed to clear suspicious flag: $($clearResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Suspicious flag cleared successfully" -ForegroundColor Green
    return $true
}

# Test resolving detection event
function Test-ResolveDetection {
    param([string]$Token, [string]$DetectionID)
    
    Write-Host "Testing resolve detection event $DetectionID..." -ForegroundColor Green
    
    $resolveBody = @{
        resolution_notes = "Test resolution - false positive"
    } | ConvertTo-Json
    
    $resolveResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/suspicious/resolve/$DetectionID" -Body $resolveBody -Token $Token
    
    if (-not $resolveResult.Success) {
        Write-Host "Failed to resolve detection event: $($resolveResult.Error)" -ForegroundColor Red
        return $false
    }
    
    Write-Host "Detection event resolved successfully" -ForegroundColor Green
    return $true
}

# Test creating and accessing emails to trigger suspicious detection
function Test-SuspiciousDetection {
    param([string]$Token)
    
    Write-Host "Testing suspicious access detection..." -ForegroundColor Green
    
    # Create a test email
    $createEmailBody = @{
        recipient = "recipient@example.com"
        subject = "Test Email for Suspicious Detection"
        content = "This is a test email for suspicious access detection testing."
    } | ConvertTo-Json
    
    $createResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/emails/send" -Body $createEmailBody -Token $Token
    
    if (-not $createResult.Success) {
        Write-Host "Failed to create test email: $($createResult.Error)" -ForegroundColor Red
        return $null
    }
    
    $emailID = $createResult.Data.email_id
    Write-Host "Test email created with ID: $emailID" -ForegroundColor Green
    
    # Simulate multiple failed access attempts to trigger detection
    Write-Host "Simulating multiple failed access attempts..." -ForegroundColor Yellow
    
    for ($i = 1; $i -le 4; $i++) {
        Write-Host "Attempt $i..." -ForegroundColor Yellow
        
        # Try to access the email with wrong credentials or from different IPs
        $accessResult = Invoke-ApiRequest -Method "POST" -Endpoint "/api/email/get" -Body "{}" -Token $Token
        
        # Wait a bit between attempts
        Start-Sleep -Seconds 2
    }
    
    # Check if suspicious activity was detected
    Start-Sleep -Seconds 5  # Give time for detection to process
    
    $activityResult = Invoke-ApiRequest -Method "GET" -Endpoint "/api/suspicious/activity/$emailID" -Token $Token
    
    if ($activityResult.Success) {
        Write-Host "Suspicious activity detected: $($activityResult.Data | ConvertTo-Json)" -ForegroundColor Green
        
        # Test clearing the suspicious flag
        Test-ClearSuspiciousFlag -Token $Token -EmailID $emailID
        
        return $emailID
    } else {
        Write-Host "No suspicious activity detected yet" -ForegroundColor Yellow
        return $emailID
    }
}

# Main test execution
function Main {
    Write-Host "Starting Suspicious Access Detection Integration Tests" -ForegroundColor Cyan
    Write-Host "Base URL: $BaseUrl" -ForegroundColor Cyan
    Write-Host "Test Email: $TestEmail" -ForegroundColor Cyan
    Write-Host "==================================================" -ForegroundColor Cyan
    
    # Test user authentication
    $token = Test-UserAuth
    if (-not $token) {
        Write-Host "Authentication failed. Exiting tests." -ForegroundColor Red
        return
    }
    
    Write-Host "Authentication successful. Token: $($token.Substring(0, 20))..." -ForegroundColor Green
    
    # Test suspicious detection preferences
    if (-not (Test-SuspiciousPreferences -Token $token)) {
        Write-Host "Suspicious preferences test failed" -ForegroundColor Red
        return
    }
    
    # Test detection rules
    if (-not (Test-DetectionRules -Token $token)) {
        Write-Host "Detection rules test failed" -ForegroundColor Red
        return
    }
    
    # Test suspicious emails list
    if (-not (Test-SuspiciousEmails -Token $token)) {
        Write-Host "Suspicious emails test failed" -ForegroundColor Red
        return
    }
    
    # Test suspicious detection
    $emailID = Test-SuspiciousDetection -Token $token
    if ($emailID) {
        # Test suspicious activity for the email
        Test-SuspiciousActivity -Token $token -EmailID $emailID
    }
    
    Write-Host "==================================================" -ForegroundColor Cyan
    Write-Host "Suspicious Access Detection Integration Tests Completed" -ForegroundColor Cyan
}

# Run the tests
Main







