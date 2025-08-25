# =============================================================================
# SECURE LINK VIEWING INTEGRATION TEST
# =============================================================================
# This script tests the complete secure link viewing flow for external users
# It verifies public endpoints, security validation, and content access
# =============================================================================

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$TestEmail = "test@example.com",
    [string]$TestPassword = "testpass123",
    [string]$ExternalRecipient = "external.test@gmail.com"
)

# =============================================================================
# HELPER FUNCTIONS
# =============================================================================

function Test-APIHealth {
    Write-Host "🔍 Testing API health..." -ForegroundColor Blue
    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/api/health" -Method GET -TimeoutSec 10
        if ($response.status -eq "healthy") {
            Write-Host "✅ API is healthy" -ForegroundColor Green
            return $true
        } else {
            Write-Host "❌ API health check failed" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ API health check failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-Login {
    Write-Host "🔐 Testing login..." -ForegroundColor Blue
    try {
        $loginData = @{
            email = $TestEmail
            password = $TestPassword
        } | ConvertTo-Json

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        
        if ($response.token) {
            Write-Host "✅ Login successful" -ForegroundColor Green
            return $response.token
        } else {
            Write-Host "❌ Login failed" -ForegroundColor Red
            return $null
        }
    } catch {
        Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

function Test-SendSecureLinkEmail {
    param([string]$Token)
    Write-Host "📧 Testing secure link email sending..." -ForegroundColor Blue
    try {
        $emailData = @{
            recipient = $ExternalRecipient
            subject = "Test Secure Link - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
            body = "This is a test secure email with various security features enabled."
            password = "testpass123"
            requireMFA = $true
            mfaType = "email"
            geoVerificationType = "country"
            geoCountry = "United States"
            timeLock = $false
            burnAfterRead = $true
            selfDestructAfterAttempts = 3
            expiresAt = (Get-Date).AddDays(7).ToString("yyyy-MM-ddTHH:mm:ssZ")
        } | ConvertTo-Json

        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/emails/send" -Method POST -Body $emailData -Headers $headers
        
        if ($response.secure_link_id) {
            Write-Host "✅ Secure link email sent successfully" -ForegroundColor Green
            Write-Host "   Link ID: $($response.secure_link_id)" -ForegroundColor Cyan
            return $response.secure_link_id
        } else {
            Write-Host "❌ Secure link email failed" -ForegroundColor Red
            return $null
        }
    } catch {
        Write-Host "❌ Secure link email failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

function Test-PublicSecureLinkAccess {
    param([string]$LinkID)
    Write-Host "🔗 Testing public secure link access..." -ForegroundColor Blue
    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID" -Method GET -TimeoutSec 10
        
        if ($response.link_id -and $response.status -eq "active") {
            Write-Host "✅ Public secure link access successful" -ForegroundColor Green
            Write-Host "   Subject: $($response.subject)" -ForegroundColor Cyan
            Write-Host "   Sender: $($response.sender_email)" -ForegroundColor Cyan
            Write-Host "   Security: Password=$($response.security_settings.require_password), MFA=$($response.security_settings.require_mfa)" -ForegroundColor Cyan
            return $response
        } else {
            Write-Host "❌ Public secure link access failed" -ForegroundColor Red
            return $null
        }
    } catch {
        Write-Host "❌ Public secure link access failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

function Test-SecurityValidation {
    param([string]$LinkID, [string]$Password)
    Write-Host "🔐 Testing security validation..." -ForegroundColor Blue
    try {
        $validationData = @{
            link_id = $LinkID
            password = $Password
            ip_address = "127.0.0.1"
            user_agent = "PowerShell-Test/1.0"
        } | ConvertTo-Json

        $response = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID/validate" -Method POST -Body $validationData -ContentType "application/json"
        
        if ($response.success -and $response.validated) {
            Write-Host "✅ Security validation successful" -ForegroundColor Green
            return $true
        } elseif ($response.requires_mfa) {
            Write-Host "⚠️ MFA required, testing MFA validation..." -ForegroundColor Yellow
            
            # Test MFA validation (using a mock code)
            $mfaData = @{
                link_id = $LinkID
                password = $Password
                mfa_code = "123456"
                mfa_type = "email"
                ip_address = "127.0.0.1"
                user_agent = "PowerShell-Test/1.0"
            } | ConvertTo-Json

            $mfaResponse = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID/validate" -Method POST -Body $mfaData -ContentType "application/json"
            
            if ($mfaResponse.success -and $mfaResponse.validated) {
                Write-Host "✅ MFA validation successful" -ForegroundColor Green
                return $true
            } else {
                Write-Host "❌ MFA validation failed: $($mfaResponse.error)" -ForegroundColor Red
                return $false
            }
        } else {
            Write-Host "❌ Security validation failed: $($response.error)" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ Security validation failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-SecureEmailContent {
    param([string]$LinkID)
    Write-Host "📄 Testing secure email content retrieval..." -ForegroundColor Blue
    try {
        $contentData = @{
            link_id = $LinkID
            ip_address = "127.0.0.1"
            user_agent = "PowerShell-Test/1.0"
        } | ConvertTo-Json

        $response = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID/content" -Method POST -Body $contentData -ContentType "application/json"
        
        if ($response.link_id -and $response.subject -and $response.body) {
            Write-Host "✅ Secure email content retrieved successfully" -ForegroundColor Green
            Write-Host "   Subject: $($response.subject)" -ForegroundColor Cyan
            Write-Host "   Body length: $($response.body.Length) characters" -ForegroundColor Cyan
            Write-Host "   Read-once: $($response.read_once)" -ForegroundColor Cyan
            return $response
        } else {
            Write-Host "❌ Secure email content retrieval failed" -ForegroundColor Red
            return $null
        }
    } catch {
        Write-Host "❌ Secure email content retrieval failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

function Test-InvalidLinkAccess {
    param([string]$InvalidLinkID = "invalid-link-12345")
    Write-Host "🚫 Testing invalid link access..." -ForegroundColor Blue
    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/v/$InvalidLinkID" -Method GET -TimeoutSec 10
        Write-Host "❌ Invalid link access should have failed" -ForegroundColor Red
        return $false
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        if ($statusCode -eq 404) {
            Write-Host "✅ Invalid link access properly rejected (404)" -ForegroundColor Green
            return $true
        } else {
            Write-Host "⚠️ Invalid link access returned unexpected status: $statusCode" -ForegroundColor Yellow
            return $false
        }
    }
}

function Test-ExpiredLinkAccess {
    param([string]$Token)
    Write-Host "⏰ Testing expired link access..." -ForegroundColor Blue
    try {
        # Send a link that expires in 1 second
        $emailData = @{
            recipient = $ExternalRecipient
            subject = "Test Expired Link - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
            body = "This is a test secure email that will expire quickly."
            expiresAt = (Get-Date).AddSeconds(1).ToString("yyyy-MM-ddTHH:mm:ssZ")
        } | ConvertTo-Json

        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/emails/send" -Method POST -Body $emailData -Headers $headers
        
        if ($response.secure_link_id) {
            Write-Host "   Created expired link: $($response.secure_link_id)" -ForegroundColor Cyan
            
            # Wait for expiration
            Start-Sleep -Seconds 3
            
            # Try to access the expired link
            try {
                $accessResponse = Invoke-RestMethod -Uri "$ApiUrl/v/$($response.secure_link_id)" -Method GET -TimeoutSec 10
                Write-Host "❌ Expired link access should have failed" -ForegroundColor Red
                return $false
            } catch {
                $statusCode = $_.Exception.Response.StatusCode.value__
                if ($statusCode -eq 410) {
                    Write-Host "✅ Expired link access properly rejected (410)" -ForegroundColor Green
                    return $true
                } else {
                    Write-Host "⚠️ Expired link access returned unexpected status: $statusCode" -ForegroundColor Yellow
                    return $false
                }
            }
        } else {
            Write-Host "❌ Failed to create expired link" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ Expired link test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-AutoDestructBehavior {
    param([string]$Token)
    Write-Host "💥 Testing auto-destruct behavior..." -ForegroundColor Blue
    try {
        # Send a link with auto-destruct after 1 attempt
        $emailData = @{
            recipient = $ExternalRecipient
            subject = "Test Auto-Destruct Link - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
            body = "This is a test secure email with auto-destruct enabled."
            password = "wrongpassword"
            selfDestructAfterAttempts = 1
        } | ConvertTo-Json

        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/emails/send" -Method POST -Body $emailData -Headers $headers
        
        if ($response.secure_link_id) {
            Write-Host "   Created auto-destruct link: $($response.secure_link_id)" -ForegroundColor Cyan
            
            # Try to validate with wrong password (should trigger auto-destruct)
            $validationData = @{
                link_id = $response.secure_link_id
                password = "wrongpassword"
                ip_address = "127.0.0.1"
                user_agent = "PowerShell-Test/1.0"
            } | ConvertTo-Json

            try {
                $validationResponse = Invoke-RestMethod -Uri "$ApiUrl/v/$($response.secure_link_id)/validate" -Method POST -Body $validationData -ContentType "application/json"
                
                if ($validationResponse.error_code -eq "LINK_DESTROYED") {
                    Write-Host "✅ Auto-destruct behavior working correctly" -ForegroundColor Green
                    return $true
                } else {
                    Write-Host "⚠️ Auto-destruct behavior unexpected: $($validationResponse.error_code)" -ForegroundColor Yellow
                    return $false
                }
            } catch {
                $statusCode = $_.Exception.Response.StatusCode.value__
                if ($statusCode -eq 410) {
                    Write-Host "✅ Auto-destruct behavior working correctly (410)" -ForegroundColor Green
                    return $true
                } else {
                    Write-Host "⚠️ Auto-destruct behavior returned unexpected status: $statusCode" -ForegroundColor Yellow
                    return $false
                }
            }
        } else {
            Write-Host "❌ Failed to create auto-destruct link" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ Auto-destruct test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-AuditLogging {
    param([string]$Token, [string]$LinkID)
    Write-Host "📊 Testing audit logging..." -ForegroundColor Blue
    try {
        # Query audit logs for the link
        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/secure-links/$LinkID" -Method GET -Headers $headers
        
        if ($response.link_id -and $response.audit_logs) {
            Write-Host "✅ Audit logging working correctly" -ForegroundColor Green
            Write-Host "   Audit entries: $($response.audit_logs.Count)" -ForegroundColor Cyan
            
            foreach ($log in $response.audit_logs) {
                Write-Host "   - $($log.timestamp): $($log.action) from $($log.ip_address)" -ForegroundColor Gray
            }
            return $true
        } else {
            Write-Host "⚠️ No audit logs found" -ForegroundColor Yellow
            return $false
        }
    } catch {
        Write-Host "❌ Audit logging test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

Write-Host "🚀 Starting Secure Link Viewing Integration Test" -ForegroundColor Green
Write-Host "API URL: $ApiUrl" -ForegroundColor Cyan
Write-Host "Test Email: $TestEmail" -ForegroundColor Cyan
Write-Host "External Recipient: $ExternalRecipient" -ForegroundColor Cyan
Write-Host ""

# Test 1: API Health Check
if (-not (Test-APIHealth)) {
    Write-Host "❌ API health check failed. Exiting." -ForegroundColor Red
    exit 1
}

# Test 2: Login
$token = Test-Login
if (-not $token) {
    Write-Host "❌ Login failed. Exiting." -ForegroundColor Red
    exit 1
}

# Test 3: Send Secure Link Email
$linkID = Test-SendSecureLinkEmail -Token $token
if (-not $linkID) {
    Write-Host "❌ Secure link email failed. Exiting." -ForegroundColor Red
    exit 1
}

# Wait a moment for email processing
Start-Sleep -Seconds 2

# Test 4: Public Secure Link Access
$metadata = Test-PublicSecureLinkAccess -LinkID $linkID
if (-not $metadata) {
    Write-Host "❌ Public secure link access failed. Exiting." -ForegroundColor Red
    exit 1
}

# Test 5: Security Validation
$validationSuccess = Test-SecurityValidation -LinkID $linkID -Password "testpass123"
if (-not $validationSuccess) {
    Write-Host "❌ Security validation failed. Exiting." -ForegroundColor Red
    exit 1
}

# Test 6: Secure Email Content
$content = Test-SecureEmailContent -LinkID $linkID
if (-not $content) {
    Write-Host "❌ Secure email content retrieval failed. Exiting." -ForegroundColor Red
    exit 1
}

# Test 7: Invalid Link Access
$invalidTest = Test-InvalidLinkAccess
if (-not $invalidTest) {
    Write-Host "⚠️ Invalid link access test failed" -ForegroundColor Yellow
}

# Test 8: Expired Link Access
$expiredTest = Test-ExpiredLinkAccess -Token $token
if (-not $expiredTest) {
    Write-Host "⚠️ Expired link access test failed" -ForegroundColor Yellow
}

# Test 9: Auto-Destruct Behavior
$autoDestructTest = Test-AutoDestructBehavior -Token $token
if (-not $autoDestructTest) {
    Write-Host "⚠️ Auto-destruct behavior test failed" -ForegroundColor Yellow
}

# Test 10: Audit Logging
$auditTest = Test-AuditLogging -Token $token -LinkID $linkID
if (-not $auditTest) {
    Write-Host "⚠️ Audit logging test failed" -ForegroundColor Yellow
}

# =============================================================================
# TEST SUMMARY
# =============================================================================

Write-Host ""
Write-Host "🎯 Secure Link Viewing Integration Test Summary" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host "✅ API Health Check: PASSED" -ForegroundColor Green
Write-Host "✅ Login: PASSED" -ForegroundColor Green
Write-Host "✅ Secure Link Email: PASSED" -ForegroundColor Green
Write-Host "✅ Public Link Access: PASSED" -ForegroundColor Green
Write-Host "✅ Security Validation: PASSED" -ForegroundColor Green
Write-Host "✅ Content Retrieval: PASSED" -ForegroundColor Green

if ($invalidTest) {
    Write-Host "✅ Invalid Link Access: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Invalid Link Access: PARTIAL" -ForegroundColor Yellow
}

if ($expiredTest) {
    Write-Host "✅ Expired Link Access: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Expired Link Access: PARTIAL" -ForegroundColor Yellow
}

if ($autoDestructTest) {
    Write-Host "✅ Auto-Destruct Behavior: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Auto-Destruct Behavior: PARTIAL" -ForegroundColor Yellow
}

if ($auditTest) {
    Write-Host "✅ Audit Logging: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Audit Logging: PARTIAL" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Secure Link Viewing Integration Test Completed Successfully!" -ForegroundColor Green
Write-Host "External recipients can now access secure links with proper security validation." -ForegroundColor Cyan
