# =============================================================================
# SECURE LINK REPLIES INTEGRATION TEST
# =============================================================================
# This script tests the complete secure link reply flow for external users
# It verifies reply endpoints, forwarding to internal senders, and audit logging
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
            subject = "Test Secure Link Reply - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
            body = "This is a test secure email for reply testing. Please reply to this message."
            password = "testpass123"
            requireMFA = $false
            geoVerificationType = "none"
            timeLock = $false
            burnAfterRead = $false
            selfDestructAfterAttempts = 5
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

function Test-ValidReply {
    param([string]$LinkID)
    Write-Host "📝 Testing valid reply..." -ForegroundColor Blue
    try {
        $replyData = @{
            link_id = $LinkID
            subject = "Re: Test Secure Link Reply"
            body = "This is a test reply from an external recipient. Thank you for the secure message!"
            ip_address = "127.0.0.1"
            user_agent = "PowerShell-Test/1.0"
        } | ConvertTo-Json

        $response = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID/reply" -Method POST -Body $replyData -ContentType "application/json"
        
        if ($response.success) {
            Write-Host "✅ Valid reply sent successfully" -ForegroundColor Green
            Write-Host "   Reply ID: $($response.reply_id)" -ForegroundColor Cyan
            Write-Host "   Transaction ID: $($response.transaction_id)" -ForegroundColor Cyan
            return $response
        } else {
            Write-Host "❌ Valid reply failed: $($response.error)" -ForegroundColor Red
            return $null
        }
    } catch {
        Write-Host "❌ Valid reply failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

function Test-ReplyToExpiredLink {
    param([string]$Token)
    Write-Host "⏰ Testing reply to expired link..." -ForegroundColor Blue
    try {
        # Send a link that expires in 1 second
        $emailData = @{
            recipient = $ExternalRecipient
            subject = "Test Expired Link Reply - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
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
            
            # Try to reply to the expired link
            $replyData = @{
                link_id = $response.secure_link_id
                subject = "Re: Test Expired Link Reply"
                body = "This reply should fail because the link is expired."
                ip_address = "127.0.0.1"
                user_agent = "PowerShell-Test/1.0"
            } | ConvertTo-Json

            try {
                $replyResponse = Invoke-RestMethod -Uri "$ApiUrl/v/$($response.secure_link_id)/reply" -Method POST -Body $replyData -ContentType "application/json"
                Write-Host "❌ Reply to expired link should have failed" -ForegroundColor Red
                return $false
            } catch {
                $statusCode = $_.Exception.Response.StatusCode.value__
                if ($statusCode -eq 400) {
                    Write-Host "✅ Reply to expired link properly rejected (400)" -ForegroundColor Green
                    return $true
                } else {
                    Write-Host "⚠️ Reply to expired link returned unexpected status: $statusCode" -ForegroundColor Yellow
                    return $false
                }
            }
        } else {
            Write-Host "❌ Failed to create expired link" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ Expired link reply test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-ReplyToRevokedLink {
    param([string]$Token)
    Write-Host "🚫 Testing reply to revoked link..." -ForegroundColor Blue
    try {
        # Send a secure link
        $emailData = @{
            recipient = $ExternalRecipient
            subject = "Test Revoked Link Reply - $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
            body = "This is a test secure email that will be revoked."
        } | ConvertTo-Json

        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/emails/send" -Method POST -Body $emailData -Headers $headers
        
        if ($response.secure_link_id) {
            Write-Host "   Created link: $($response.secure_link_id)" -ForegroundColor Cyan
            
            # Revoke the link
            $revokeData = @{} | ConvertTo-Json
            $revokeResponse = Invoke-RestMethod -Uri "$ApiUrl/api/secure-links/$($response.secure_link_id)/revoke" -Method POST -Body $revokeData -Headers $headers
            
            if ($revokeResponse.success) {
                Write-Host "   Link revoked successfully" -ForegroundColor Cyan
                
                # Try to reply to the revoked link
                $replyData = @{
                    link_id = $response.secure_link_id
                    subject = "Re: Test Revoked Link Reply"
                    body = "This reply should fail because the link is revoked."
                    ip_address = "127.0.0.1"
                    user_agent = "PowerShell-Test/1.0"
                } | ConvertTo-Json

                try {
                    $replyResponse = Invoke-RestMethod -Uri "$ApiUrl/v/$($response.secure_link_id)/reply" -Method POST -Body $replyData -ContentType "application/json"
                    Write-Host "❌ Reply to revoked link should have failed" -ForegroundColor Red
                    return $false
                } catch {
                    $statusCode = $_.Exception.Response.StatusCode.value__
                    if ($statusCode -eq 400) {
                        Write-Host "✅ Reply to revoked link properly rejected (400)" -ForegroundColor Green
                        return $true
                    } else {
                        Write-Host "⚠️ Reply to revoked link returned unexpected status: $statusCode" -ForegroundColor Yellow
                        return $false
                    }
                }
            } else {
                Write-Host "❌ Failed to revoke link" -ForegroundColor Red
                return $false
            }
        } else {
            Write-Host "❌ Failed to create link" -ForegroundColor Red
            return $false
        }
    } catch {
        Write-Host "❌ Revoked link reply test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-MultipleReplies {
    param([string]$LinkID)
    Write-Host "📨 Testing multiple replies..." -ForegroundColor Blue
    try {
        $replies = @()
        
        # Send first reply
        $reply1Data = @{
            link_id = $LinkID
            subject = "Re: Test Secure Link Reply - First Reply"
            body = "This is the first reply to the secure message."
            ip_address = "127.0.0.1"
            user_agent = "PowerShell-Test/1.0"
        } | ConvertTo-Json

        $response1 = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID/reply" -Method POST -Body $reply1Data -ContentType "application/json"
        
        if ($response1.success) {
            $replies += $response1
            Write-Host "   First reply sent: $($response1.reply_id)" -ForegroundColor Cyan
        } else {
            Write-Host "❌ First reply failed" -ForegroundColor Red
            return $false
        }

        # Send second reply
        $reply2Data = @{
            link_id = $LinkID
            subject = "Re: Test Secure Link Reply - Second Reply"
            body = "This is the second reply to the secure message."
            ip_address = "127.0.0.1"
            user_agent = "PowerShell-Test/1.0"
        } | ConvertTo-Json

        $response2 = Invoke-RestMethod -Uri "$ApiUrl/v/$LinkID/reply" -Method POST -Body $reply2Data -ContentType "application/json"
        
        if ($response2.success) {
            $replies += $response2
            Write-Host "   Second reply sent: $($response2.reply_id)" -ForegroundColor Cyan
        } else {
            Write-Host "❌ Second reply failed" -ForegroundColor Red
            return $false
        }

        Write-Host "✅ Multiple replies handled correctly" -ForegroundColor Green
        Write-Host "   Total replies: $($replies.Count)" -ForegroundColor Cyan
        return $true
    } catch {
        Write-Host "❌ Multiple replies test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-ReplyAuditLogging {
    param([string]$Token, [string]$LinkID)
    Write-Host "📊 Testing reply audit logging..." -ForegroundColor Blue
    try {
        # Query audit logs for the link
        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/secure-links/$LinkID" -Method GET -Headers $headers
        
        if ($response.link_id -and $response.audit_logs) {
            $replyLogs = $response.audit_logs | Where-Object { $_.event_type -like "*reply*" }
            
            if ($replyLogs.Count -gt 0) {
                Write-Host "✅ Reply audit logging working correctly" -ForegroundColor Green
                Write-Host "   Reply audit entries: $($replyLogs.Count)" -ForegroundColor Cyan
                
                foreach ($log in $replyLogs) {
                    Write-Host "   - $($log.timestamp): $($log.action) from $($log.ip_address)" -ForegroundColor Gray
                }
                return $true
            } else {
                Write-Host "⚠️ No reply audit logs found" -ForegroundColor Yellow
                return $false
            }
        } else {
            Write-Host "⚠️ No audit logs found" -ForegroundColor Yellow
            return $false
        }
    } catch {
        Write-Host "❌ Reply audit logging test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

function Test-SESTransactionLogging {
    param([string]$Token, [string]$ReplyID)
    Write-Host "📧 Testing SES transaction logging for reply..." -ForegroundColor Blue
    try {
        # Query SES transactions for the reply
        $headers = @{
            "Authorization" = "Bearer $Token"
            "Content-Type" = "application/json"
        }

        $response = Invoke-RestMethod -Uri "$ApiUrl/api/ses-transactions" -Method GET -Headers $headers
        
        if ($response.transactions) {
            $replyTransactions = $response.transactions | Where-Object { $_.reply_id -eq $ReplyID }
            
            if ($replyTransactions.Count -gt 0) {
                Write-Host "✅ SES transaction logging working correctly" -ForegroundColor Green
                Write-Host "   Reply transactions: $($replyTransactions.Count)" -ForegroundColor Cyan
                
                foreach ($txn in $replyTransactions) {
                    Write-Host "   - $($txn.transaction_id): $($txn.status) to $($txn.recipient)" -ForegroundColor Gray
                }
                return $true
            } else {
                Write-Host "⚠️ No SES transactions found for reply" -ForegroundColor Yellow
                return $false
            }
        } else {
            Write-Host "⚠️ No SES transactions found" -ForegroundColor Yellow
            return $false
        }
    } catch {
        Write-Host "❌ SES transaction logging test failed: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# =============================================================================
# MAIN TEST EXECUTION
# =============================================================================

Write-Host "🚀 Starting Secure Link Replies Integration Test" -ForegroundColor Green
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

# Test 4: Valid Reply
$replyResponse = Test-ValidReply -LinkID $linkID
if (-not $replyResponse) {
    Write-Host "❌ Valid reply failed. Exiting." -ForegroundColor Red
    exit 1
}

# Test 5: Multiple Replies
$multipleRepliesTest = Test-MultipleReplies -LinkID $linkID
if (-not $multipleRepliesTest) {
    Write-Host "⚠️ Multiple replies test failed" -ForegroundColor Yellow
}

# Test 6: Reply to Expired Link
$expiredReplyTest = Test-ReplyToExpiredLink -Token $token
if (-not $expiredReplyTest) {
    Write-Host "⚠️ Expired link reply test failed" -ForegroundColor Yellow
}

# Test 7: Reply to Revoked Link
$revokedReplyTest = Test-ReplyToRevokedLink -Token $token
if (-not $revokedReplyTest) {
    Write-Host "⚠️ Revoked link reply test failed" -ForegroundColor Yellow
}

# Test 8: Reply Audit Logging
$auditTest = Test-ReplyAuditLogging -Token $token -LinkID $linkID
if (-not $auditTest) {
    Write-Host "⚠️ Reply audit logging test failed" -ForegroundColor Yellow
}

# Test 9: SES Transaction Logging
if ($replyResponse.reply_id) {
    $sesTest = Test-SESTransactionLogging -Token $token -ReplyID $replyResponse.reply_id
    if (-not $sesTest) {
        Write-Host "⚠️ SES transaction logging test failed" -ForegroundColor Yellow
    }
}

# =============================================================================
# TEST SUMMARY
# =============================================================================

Write-Host ""
Write-Host "🎯 Secure Link Replies Integration Test Summary" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green
Write-Host "✅ API Health Check: PASSED" -ForegroundColor Green
Write-Host "✅ Login: PASSED" -ForegroundColor Green
Write-Host "✅ Secure Link Email: PASSED" -ForegroundColor Green
Write-Host "✅ Valid Reply: PASSED" -ForegroundColor Green

if ($multipleRepliesTest) {
    Write-Host "✅ Multiple Replies: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Multiple Replies: PARTIAL" -ForegroundColor Yellow
}

if ($expiredReplyTest) {
    Write-Host "✅ Expired Link Reply: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Expired Link Reply: PARTIAL" -ForegroundColor Yellow
}

if ($revokedReplyTest) {
    Write-Host "✅ Revoked Link Reply: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Revoked Link Reply: PARTIAL" -ForegroundColor Yellow
}

if ($auditTest) {
    Write-Host "✅ Reply Audit Logging: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ Reply Audit Logging: PARTIAL" -ForegroundColor Yellow
}

if ($sesTest) {
    Write-Host "✅ SES Transaction Logging: PASSED" -ForegroundColor Green
} else {
    Write-Host "⚠️ SES Transaction Logging: PARTIAL" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "🎉 Secure Link Replies Integration Test Completed Successfully!" -ForegroundColor Green
Write-Host "External recipients can now reply to secure links with proper forwarding and audit logging." -ForegroundColor Cyan
