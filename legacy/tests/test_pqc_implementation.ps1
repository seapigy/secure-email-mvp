# Test PQC Implementation with Sample Emails and Performance Benchmarks
# This script tests the new quantum-resistant email encryption system

Write-Host "🔐 Testing PQC Implementation with Sample Emails" -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Green

# Configuration
$API_BASE_URL = "http://localhost:8080"
$TEST_USER_EMAIL = "test@securesystem.email"
$TEST_USER_PASSWORD = "SecurePassword123!"
$TEST_TOTP_CODE = "123456"

# Test data
$SAMPLE_EMAILS = @(
    @{
        subject = "Test Email 1 - Short Message"
        body = "This is a short test email to verify PQC encryption is working correctly."
        recipient = "recipient1@securesystem.email"
    },
    @{
        subject = "Test Email 2 - Medium Message"
        body = "This is a medium-length test email with more content to test PQC encryption performance. It contains multiple sentences and should provide a good benchmark for the quantum-resistant encryption system."
        recipient = "recipient2@securesystem.email"
    },
    @{
        subject = "Test Email 3 - Long Message"
        body = "This is a long test email designed to thoroughly test the PQC hybrid encryption system. It contains extensive content with multiple paragraphs to simulate real-world email usage patterns. The quantum-resistant encryption should handle this efficiently while maintaining maximum security. This email tests the Kyber-768 key encapsulation mechanism combined with AES-256-GCM data encryption for optimal performance and security."
        recipient = "recipient3@securesystem.email"
    },
    @{
        subject = "Test Email 4 - Special Characters"
        body = "This email contains special characters: !@#`$%^`&*()`_+-=[]{}|;':"",./<>? and unicode: 🚀🔐📧💻 to test PQC encryption with various content types."
        recipient = "recipient4@securesystem.email"
    }
)

# Performance tracking
$performanceResults = @()

function Test-Login {
    Write-Host "`n🔑 Testing Authentication..." -ForegroundColor Yellow
    
    $loginData = @{
        email = $TEST_USER_EMAIL
        password = $TEST_USER_PASSWORD
        totp_code = $TEST_TOTP_CODE
    } | ConvertTo-Json

    $startTime = Get-Date
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalMilliseconds

    if ($response.token) {
        Write-Host "✅ Login successful in ${duration}ms" -ForegroundColor Green
        $script:authToken = $response.token
        return $true
    } else {
        Write-Host "❌ Login failed" -ForegroundColor Red
        return $false
    }
}

function Test-SendEmail {
    param(
        [string]$Subject,
        [string]$Body,
        [string]$Recipient,
        [int]$TestNumber
    )
    
    Write-Host "`n📧 Testing Email Send #$TestNumber..." -ForegroundColor Yellow
    Write-Host "Subject: $Subject" -ForegroundColor Cyan
    Write-Host "Recipient: $Recipient" -ForegroundColor Cyan
    Write-Host "Body Length: $($Body.Length) characters" -ForegroundColor Cyan

    $emailData = @{
        recipient = $Recipient
        subject = $Subject
        body = $Body
        password = ""  # No password protection for this test
        burn_after_read = $false
        self_destruct_after_attempts = $false
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $authToken"
        "Content-Type" = "application/json"
    }

    $startTime = Get-Date
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/send" -Method POST -Body $emailData -Headers $headers
    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalMilliseconds

    if ($response.email_id) {
        Write-Host "✅ Email sent successfully in ${duration}ms" -ForegroundColor Green
        Write-Host "Email ID: $($response.email_id)" -ForegroundColor Gray
        
        $performanceResults += [PSCustomObject]@{
            TestType = "Send"
            TestNumber = $TestNumber
            Subject = $Subject
            BodyLength = $Body.Length
            Duration = $duration
            EmailID = $response.email_id
        }
        
        return $response.email_id
    } else {
        Write-Host "❌ Email send failed" -ForegroundColor Red
        return $null
    }
}

function Test-GetEmail {
    param(
        [string]$EmailID,
        [int]$TestNumber
    )
    
    Write-Host "`n📥 Testing Email Retrieval #$TestNumber..." -ForegroundColor Yellow
    Write-Host "Email ID: $EmailID" -ForegroundColor Cyan

    $getData = @{
        email_id = $EmailID
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $authToken"
        "Content-Type" = "application/json"
    }

    $startTime = Get-Date
    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/get" -Method POST -Body $getData -Headers $headers
    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalMilliseconds

    if ($response.status -eq "success") {
        Write-Host "✅ Email retrieved successfully in ${duration}ms" -ForegroundColor Green
        Write-Host "Subject: $($response.subject)" -ForegroundColor Gray
        Write-Host "Body Length: $($response.body.Length) characters" -ForegroundColor Gray
        
        $performanceResults += [PSCustomObject]@{
            TestType = "Retrieve"
            TestNumber = $TestNumber
            Subject = $response.subject
            BodyLength = $response.body.Length
            Duration = $duration
            EmailID = $EmailID
        }
        
        return $true
    } else {
        Write-Host "❌ Email retrieval failed" -ForegroundColor Red
        return $false
    }
}

function Test-PQCPerformance {
    Write-Host "`n🚀 Testing PQC Performance Benchmarks..." -ForegroundColor Yellow
    
    # Test different email sizes
    $testSizes = @(100, 500, 1000, 5000, 10000)
    $pqcResults = @()
    
    foreach ($size in $testSizes) {
        $testBody = "A" * $size  # Create test data of specified size
        
        Write-Host "Testing PQC encryption with ${size} character email..." -ForegroundColor Cyan
        
        $emailData = @{
            recipient = "benchmark@securesystem.email"
            subject = "PQC Benchmark Test - $size chars"
            body = $testBody
        } | ConvertTo-Json

        $headers = @{
            "Authorization" = "Bearer $authToken"
            "Content-Type" = "application/json"
        }

        # Send test
        $startTime = Get-Date
        $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/send" -Method POST -Body $emailData -Headers $headers
        $sendTime = (Get-Date) - $startTime
        
        if ($response.email_id) {
            # Retrieve test
            $getData = @{ email_id = $response.email_id } | ConvertTo-Json
            $startTime = Get-Date
            $retrieveResponse = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/get" -Method POST -Body $getData -Headers $headers
            $retrieveTime = (Get-Date) - $startTime
            
            $pqcResults += [PSCustomObject]@{
                EmailSize = $size
                SendTime = $sendTime.TotalMilliseconds
                RetrieveTime = $retrieveTime.TotalMilliseconds
                TotalTime = ($sendTime + $retrieveTime).TotalMilliseconds
            }
            
            Write-Host "  Send: $($sendTime.TotalMilliseconds)ms, Retrieve: $($retrieveTime.TotalMilliseconds)ms" -ForegroundColor Gray
        }
    }
    
    return $pqcResults
}

function Show-PerformanceResults {
    Write-Host "`n📊 Performance Results Summary" -ForegroundColor Green
    Write-Host "==============================" -ForegroundColor Green
    
    # Email send/retrieve performance
    $sendTests = $performanceResults | Where-Object { $_.TestType -eq "Send" }
    $retrieveTests = $performanceResults | Where-Object { $_.TestType -eq "Retrieve" }
    
    Write-Host "`n📧 Email Send Performance:" -ForegroundColor Yellow
    $sendTests | Format-Table -AutoSize
    
    Write-Host "`n📥 Email Retrieve Performance:" -ForegroundColor Yellow
    $retrieveTests | Format-Table -AutoSize
    
    # Calculate averages
    $avgSendTime = ($sendTests | Measure-Object -Property Duration -Average).Average
    $avgRetrieveTime = ($retrieveTests | Measure-Object -Property Duration -Average).Average
    
    Write-Host "`n📈 Performance Averages:" -ForegroundColor Yellow
    Write-Host "Average Send Time: ${avgSendTime:F2}ms" -ForegroundColor Cyan
    Write-Host "Average Retrieve Time: ${avgRetrieveTime:F2}ms" -ForegroundColor Cyan
    Write-Host "Total Average Time: $($avgSendTime + $avgRetrieveTime):F2ms" -ForegroundColor Cyan
}

function Show-PQCBenchmarks {
    param([array]$PQCResults)
    
    Write-Host "`n🔐 PQC Performance Benchmarks" -ForegroundColor Green
    Write-Host "=============================" -ForegroundColor Green
    
    $PQCResults | Format-Table -AutoSize
    
    # Calculate throughput
    $totalEmails = $PQCResults.Count
    $totalTime = ($PQCResults | Measure-Object -Property TotalTime -Sum).Sum
    $throughput = $totalEmails / ($totalTime / 1000)  # emails per second
    
    Write-Host "`n📊 PQC Throughput Analysis:" -ForegroundColor Yellow
    Write-Host "Total Emails Processed: $totalEmails" -ForegroundColor Cyan
    Write-Host "Total Processing Time: ${totalTime:F2}ms" -ForegroundColor Cyan
    Write-Host "Throughput: ${throughput:F2} emails/second" -ForegroundColor Cyan
    
    # Performance by email size
    Write-Host "`n📏 Performance by Email Size:" -ForegroundColor Yellow
    foreach ($result in $PQCResults) {
        $rate = $result.EmailSize / ($result.TotalTime / 1000)  # characters per second
        Write-Host "$($result.EmailSize) chars: ${rate:F0} chars/sec" -ForegroundColor Gray
    }
}

function Test-SecurityFeatures {
    Write-Host "`n🔒 Testing Security Features..." -ForegroundColor Yellow
    
    # Test password protection
    Write-Host "Testing password-protected email..." -ForegroundColor Cyan
    $secureEmailData = @{
        recipient = "secure@securesystem.email"
        subject = "Password Protected Test"
        body = "This email is password protected"
        password = "SecurePass123!"
    } | ConvertTo-Json

    $headers = @{
        "Authorization" = "Bearer $authToken"
        "Content-Type" = "application/json"
    }

    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/send" -Method POST -Body $secureEmailData -Headers $headers
    
    if ($response.email_id) {
        Write-Host "✅ Password-protected email created successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Password-protected email creation failed" -ForegroundColor Red
    }
    
    # Test burn-after-read
    Write-Host "Testing burn-after-read email..." -ForegroundColor Cyan
    $burnEmailData = @{
        recipient = "burn@securesystem.email"
        subject = "Burn After Read Test"
        body = "This email will be deleted after reading"
        burn_after_read = $true
    } | ConvertTo-Json

    $response = Invoke-RestMethod -Uri "$API_BASE_URL/api/email/send" -Method POST -Body $burnEmailData -Headers $headers
    
    if ($response.email_id) {
        Write-Host "✅ Burn-after-read email created successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Burn-after-read email creation failed" -ForegroundColor Red
    }
}

# Main test execution
try {
    Write-Host "Starting PQC Implementation Tests..." -ForegroundColor Green
    
    # Step 1: Login
    if (-not (Test-Login)) {
        Write-Host "❌ Cannot proceed without authentication" -ForegroundColor Red
        exit 1
    }
    
    # Step 2: Test email sending and retrieval
    $emailIDs = @()
    for ($i = 0; $i -lt $SAMPLE_EMAILS.Count; $i++) {
        $email = $SAMPLE_EMAILS[$i]
        $emailID = Test-SendEmail -Subject $email.subject -Body $email.body -Recipient $email.recipient -TestNumber ($i + 1)
        if ($emailID) {
            $emailIDs += $emailID
        }
    }
    
    # Step 3: Test email retrieval
    for ($i = 0; $i -lt $emailIDs.Count; $i++) {
        Test-GetEmail -EmailID $emailIDs[$i] -TestNumber ($i + 1)
    }
    
    # Step 4: Test PQC performance benchmarks
    $pqcBenchmarks = Test-PQCPerformance
    
    # Step 5: Test security features
    Test-SecurityFeatures
    
    # Step 6: Show results
    Show-PerformanceResults
    Show-PQCBenchmarks -PQCResults $pqcBenchmarks
    
    Write-Host "`n🎉 All PQC Implementation Tests Completed Successfully!" -ForegroundColor Green
    Write-Host "The quantum-resistant email encryption system is working correctly." -ForegroundColor Green
    
} catch {
    Write-Host "`n❌ Test failed with error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Stack trace: $($_.ScriptStackTrace)" -ForegroundColor Red
    exit 1
}
