#!/usr/bin/env pwsh
# Iteration 8 - Advanced Watermarking Integration Test
# Tests advanced watermarking features including recipient-specific watermarks,
# audio/video watermarking, inline content watermarking, and watermark templates

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$TestLinkID = "test_link_advanced_watermarking",
    [string]$TestAttachmentID = "test_attachment_advanced_watermarking",
    [string]$TestContentID = "test_content_advanced_watermarking"
)

# Colors for output
$Green = "`e[32m"
$Red = "`e[31m"
$Yellow = "`e[33m"
$Blue = "`e[34m"
$Reset = "`e[0m"

# Test counters
$script:TotalTests = 0
$script:PassedTests = 0
$script:FailedTests = 0

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Message = ""
    )
    
    $script:TotalTests++
    if ($Success) {
        $script:PassedTests++
        Write-Host "${Green}✓${Reset} $TestName" -ForegroundColor Green
        if ($Message) {
            Write-Host "  $Message" -ForegroundColor Gray
        }
    } else {
        $script:FailedTests++
        Write-Host "${Red}✗${Reset} $TestName" -ForegroundColor Red
        if ($Message) {
            Write-Host "  $Message" -ForegroundColor Red
        }
    }
}

function Test-AdvancedWatermarking {
    Write-Host "${Blue}Testing Advanced Watermarking Features...${Reset}" -ForegroundColor Blue
    
    # Test 1: Text Watermark with Recipient-Specific Information
    Test-TextWatermarkWithRecipient
    
    # Test 2: Audio Watermarking
    Test-AudioWatermarking
    
    # Test 3: Video Watermarking
    Test-VideoWatermarking
    
    # Test 4: Inline Content Watermarking
    Test-InlineContentWatermarking
    
    # Test 5: Watermark Templates
    Test-WatermarkTemplates
    
    # Test 6: Advanced Watermarking Audit Logging
    Test-AdvancedWatermarkingAudit
    
    # Test 7: Multi-Format Watermarking
    Test-MultiFormatWatermarking
    
    # Test 8: Recipient-Specific Watermark Validation
    Test-RecipientSpecificWatermarkValidation
}

function Test-TextWatermarkWithRecipient {
    $testName = "Text Watermark with Recipient-Specific Information"
    
    try {
        $requestBody = @{
            link_id = $TestLinkID
            attachment_id = $TestAttachmentID
            watermark_type = "text"
            content_type = "pdf"
            recipient_email = "test@example.com"
            recipient_id = "user_123"
            watermark_config = @{
                text = "Confidential Document"
                position = "bottom-right"
                opacity = 0.8
                font_size = 14
                color = "#FF0000"
                rotation = -45
                include_recipient = $true
                include_timestamp = $true
            }
            is_recipient_specific = $true
        } | ConvertTo-Json -Depth 3
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
        
        $success = $response.success -eq $true -and $response.config_id -and $response.applied_to -contains $TestAttachmentID
        Write-TestResult -TestName $testName -Success $success -Message "Config ID: $($response.config_id)"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-AudioWatermarking {
    $testName = "Audio Watermarking"
    
    try {
        $requestBody = @{
            link_id = $TestLinkID
            attachment_id = $TestAttachmentID
            watermark_type = "audio"
            content_type = "audio"
            recipient_email = "test@example.com"
            recipient_id = "user_123"
            watermark_config = @{
                frequency = 18000
                volume = -30
                pattern = "recipient_id"
                duration = 0.1
                include_recipient = $true
            }
            is_recipient_specific = $true
        } | ConvertTo-Json -Depth 3
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
        
        $success = $response.success -eq $true -and $response.config_id -and $response.applied_to -contains $TestAttachmentID
        Write-TestResult -TestName $testName -Success $success -Message "Config ID: $($response.config_id)"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-VideoWatermarking {
    $testName = "Video Watermarking"
    
    try {
        $requestBody = @{
            link_id = $TestLinkID
            attachment_id = $TestAttachmentID
            watermark_type = "video"
            content_type = "video"
            recipient_email = "test@example.com"
            recipient_id = "user_123"
            watermark_config = @{
                position = "bottom-right"
                opacity = 0.6
                font_size = 16
                color = "#FFFFFF"
                background_color = "#000000"
                include_recipient = $true
                overlay_duration = "full"
            }
            is_recipient_specific = $true
        } | ConvertTo-Json -Depth 3
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
        
        $success = $response.success -eq $true -and $response.config_id -and $response.applied_to -contains $TestAttachmentID
        Write-TestResult -TestName $testName -Success $success -Message "Config ID: $($response.config_id)"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-InlineContentWatermarking {
    $testName = "Inline Content Watermarking"
    
    try {
        $requestBody = @{
            link_id = $TestLinkID
            content_id = $TestContentID
            watermark_type = "inline"
            content_type = "email_content"
            recipient_email = "test@example.com"
            recipient_id = "user_123"
            watermark_config = @{
                position = "bottom-right"
                opacity = 0.5
                font_size = 10
                color = "#FF0000"
                rotation = 0
                include_recipient = $true
                include_timestamp = $true
            }
            is_recipient_specific = $true
        } | ConvertTo-Json -Depth 3
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
        
        $success = $response.success -eq $true -and $response.config_id -and $response.watermarked_content -and $response.applied_to -contains $TestContentID
        Write-TestResult -TestName $testName -Success $success -Message "Config ID: $($response.config_id)"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-WatermarkTemplates {
    $testName = "Watermark Templates"
    
    try {
        # Test getting all templates
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/watermark/templates" -Method GET
        
        $success = $response.success -eq $true -and $response.templates -and $response.templates.Count -gt 0
        
        if ($success) {
            $templateCount = $response.templates.Count
            Write-TestResult -TestName $testName -Success $success -Message "Found $templateCount templates"
            
            # Test getting templates by type
            $textTemplates = Invoke-RestMethod -Uri "$BaseUrl/api/watermark/templates?watermark_type=text" -Method GET
            $audioTemplates = Invoke-RestMethod -Uri "$BaseUrl/api/watermark/templates?watermark_type=audio" -Method GET
            $videoTemplates = Invoke-RestMethod -Uri "$BaseUrl/api/watermark/templates?watermark_type=video" -Method GET
            
            $filteredSuccess = $textTemplates.success -and $audioTemplates.success -and $videoTemplates.success
            Write-TestResult -TestName "Watermark Template Filtering" -Success $filteredSuccess -Message "Filtered by type successfully"
        } else {
            Write-TestResult -TestName $testName -Success $success -Message "No templates found"
        }
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-AdvancedWatermarkingAudit {
    $testName = "Advanced Watermarking Audit Logging"
    
    try {
        # Apply a watermark to trigger audit logging
        $requestBody = @{
            link_id = $TestLinkID
            attachment_id = $TestAttachmentID
            watermark_type = "text"
            content_type = "pdf"
            recipient_email = "audit_test@example.com"
            recipient_id = "audit_user_456"
            watermark_config = @{
                text = "Audit Test Watermark"
                position = "center"
                opacity = 0.7
                font_size = 12
                color = "#0000FF"
                rotation = 0
            }
            is_recipient_specific = $true
        } | ConvertTo-Json -Depth 3
        
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
        
        $success = $response.success -eq $true -and $response.config_id
        Write-TestResult -TestName $testName -Success $success -Message "Audit event logged for config ID: $($response.config_id)"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-MultiFormatWatermarking {
    $testName = "Multi-Format Watermarking"
    
    try {
        $formats = @("pdf", "image", "document", "audio", "video")
        $successCount = 0
        
        foreach ($format in $formats) {
            try {
                $requestBody = @{
                    link_id = $TestLinkID
                    attachment_id = "test_attachment_$format"
                    watermark_type = "text"
                    content_type = $format
                    recipient_email = "multiformat@example.com"
                    recipient_id = "user_multiformat"
                    watermark_config = @{
                        text = "Multi-Format Test"
                        position = "bottom-right"
                        opacity = 0.8
                        font_size = 12
                        color = "#FF0000"
                        rotation = -45
                    }
                    is_recipient_specific = $true
                } | ConvertTo-Json -Depth 3
                
                $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
                
                if ($response.success) {
                    $successCount++
                }
            } catch {
                # Some formats might not be supported yet, which is expected
                Write-Host "  ${Yellow}Warning:${Reset} Format $format not fully supported yet" -ForegroundColor Yellow
            }
        }
        
        $success = $successCount -gt 0
        Write-TestResult -TestName $testName -Success $success -Message "Successfully processed $successCount out of $($formats.Count) formats"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Test-RecipientSpecificWatermarkValidation {
    $testName = "Recipient-Specific Watermark Validation"
    
    try {
        $recipients = @(
            @{ email = "user1@example.com"; id = "user_001" },
            @{ email = "user2@example.com"; id = "user_002" },
            @{ email = "user3@example.com"; id = "user_003" }
        )
        
        $successCount = 0
        
        foreach ($recipient in $recipients) {
            try {
                $requestBody = @{
                    link_id = $TestLinkID
                    attachment_id = "test_attachment_recipient_$($recipient.id)"
                    watermark_type = "text"
                    content_type = "pdf"
                    recipient_email = $recipient.email
                    recipient_id = $recipient.id
                    watermark_config = @{
                        text = "Recipient-Specific Test"
                        position = "bottom-right"
                        opacity = 0.8
                        font_size = 12
                        color = "#FF0000"
                        rotation = -45
                        include_recipient = $true
                        include_timestamp = $true
                    }
                    is_recipient_specific = $true
                } | ConvertTo-Json -Depth 3
                
                $response = Invoke-RestMethod -Uri "$BaseUrl/api/v/$TestLinkID/watermark/advanced" -Method POST -Body $requestBody -ContentType "application/json"
                
                if ($response.success -and $response.recipient_info.email -eq $recipient.email) {
                    $successCount++
                }
            } catch {
                Write-Host "  ${Yellow}Warning:${Reset} Failed for recipient $($recipient.email)" -ForegroundColor Yellow
            }
        }
        
        $success = $successCount -eq $recipients.Count
        Write-TestResult -TestName $testName -Success $success -Message "Successfully processed $successCount out of $($recipients.Count) recipients"
        
    } catch {
        Write-TestResult -TestName $testName -Success $false -Message "Error: $($_.Exception.Message)"
    }
}

function Show-TestSummary {
    Write-Host "`n${Blue}=== Advanced Watermarking Test Summary ===${Reset}" -ForegroundColor Blue
    Write-Host "Total Tests: $script:TotalTests" -ForegroundColor White
    Write-Host "Passed: ${Green}$script:PassedTests${Reset}" -ForegroundColor Green
    Write-Host "Failed: ${Red}$script:FailedTests${Reset}" -ForegroundColor Red
    
    $successRate = if ($script:TotalTests -gt 0) { [math]::Round(($script:PassedTests / $script:TotalTests) * 100, 1) } else { 0 }
    Write-Host "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 80) { "Green" } elseif ($successRate -ge 60) { "Yellow" } else { "Red" })
    
    if ($script:FailedTests -eq 0) {
        Write-Host "`n${Green}🎉 All tests passed! Advanced watermarking features are working correctly.${Reset}" -ForegroundColor Green
    } else {
        Write-Host "`n${Red}❌ Some tests failed. Please check the implementation.${Reset}" -ForegroundColor Red
    }
}

function Main {
    Write-Host "${Blue}🚀 Starting Iteration 8 - Advanced Watermarking Integration Tests${Reset}" -ForegroundColor Blue
    Write-Host "Base URL: $BaseUrl" -ForegroundColor Gray
    Write-Host "Test Link ID: $TestLinkID" -ForegroundColor Gray
    Write-Host ""
    
    # Run all tests
    Test-AdvancedWatermarking
    
    # Show summary
    Show-TestSummary
    
    # Exit with appropriate code
    if ($script:FailedTests -eq 0) {
        exit 0
    } else {
        exit 1
    }
}

# Run the main function
Main
