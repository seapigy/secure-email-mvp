#!/usr/bin/env pwsh

<#
.SYNOPSIS
    Test script for Micro-Iteration 4.30: Automated Compliance & Retention Certification

.DESCRIPTION
    This script tests the compliance system features including:
    - Enterprise organization management
    - Compliance frameworks and rules
    - Violation detection and management
    - Certification generation and approval
    - API endpoints and responses

.PARAMETER ApiUrl
    The base URL of the API server (default: http://localhost:8080)

.PARAMETER AdminToken
    Admin authentication token for testing

.PARAMETER TestEnterprise
    Whether to test enterprise organization features (default: true)

.PARAMETER TestViolations
    Whether to test violation detection and management (default: true)

.PARAMETER TestCertifications
    Whether to test certification generation and approval (default: true)

.EXAMPLE
    .\test_compliance_system.ps1 -ApiUrl "http://localhost:8080" -AdminToken "your-admin-token"
#>

param(
    [string]$ApiUrl = "http://localhost:8080",
    [string]$AdminToken = "",
    [bool]$TestEnterprise = $true,
    [bool]$TestViolations = $true,
    [bool]$TestCertifications = $true
)

# Color output functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✅ $Message" "Green"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" "Cyan"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
}

# Test API endpoint function
function Test-ApiEndpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Description
    )

    Write-Info "Testing: $Description"
    Write-Info "  $Method $Endpoint"

    $headers = @{
        "Content-Type" = "application/json"
    }

    if ($AdminToken) {
        $headers["Authorization"] = "Bearer $AdminToken"
    }

    try {
        $uri = "$ApiUrl$Endpoint"

        if ($Method -eq "GET") {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers
        } else {
            $jsonBody = if ($Body) { $Body | ConvertTo-Json -Depth 10 } else { "{}" }
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $headers -Body $jsonBody
        }

        Write-Success "  Response received successfully"
        Write-Info "  Status: $($response.status)"
        Write-Info "  Message: $($response.message)"

        if ($response.data) {
            Write-Info "  Data present: $($response.data.GetType().Name)"
        }

        return $response
    }
    catch {
        Write-Error "  Failed: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode
            Write-Error "  Status Code: $statusCode"
        }
        return $null
    }
}

# Generate test data function
function Generate-TestData {
    Write-Info "Generating test data for compliance system..."

    # Test enterprise organization
    $testOrg = @{
        org_name = "Test Enterprise Corp"
        org_domain = "testenterprise.com"
        org_type = "general"
        primary_framework_id = 1
        compliance_contact_email = "compliance@testenterprise.com"
        compliance_contact_name = "John Compliance Officer"
    }

    # Test violation acknowledgment
    $testAcknowledgment = @{
        notes = "Test acknowledgment - investigating the violation"
    }

    # Test violation resolution
    $testResolution = @{
        notes = "Test resolution - violation has been addressed"
    }

    # Test certification approval
    $testApproval = @{
        notes = "Test approval - certification reviewed and approved"
    }

    # Test certification generation
    $testCertification = @{
        framework_id = 1
        certification_type = "monthly"
        period_start = (Get-Date).AddDays(-30).ToString("yyyy-MM-ddTHH:mm:ssZ")
        period_end = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    }

    return @{
        Organization = $testOrg
        Acknowledgment = $testAcknowledgment
        Resolution = $testResolution
        Approval = $testApproval
        Certification = $testCertification
    }
}

# Test compliance status endpoints
function Test-ComplianceStatusEndpoints {
    Write-Info "Testing compliance status endpoints..."

    # Test compliance status
    $statusResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/status" -Description "Get compliance status"
    if ($statusResponse) {
        Write-Success "Compliance status endpoint working"

        if ($statusResponse.data) {
            Write-Info "  Enterprise enabled: $($statusResponse.data.enterprise_enabled)"
            Write-Info "  Compliance enabled: $($statusResponse.data.compliance_enabled)"
            Write-Info "  Compliance score: $($statusResponse.data.compliance_score)"
            Write-Info "  Total violations: $($statusResponse.data.total_violations)"
            Write-Info "  Open violations: $($statusResponse.data.open_violations)"
            Write-Info "  Total certifications: $($statusResponse.data.total_certifications)"
        }
    }

    # Test compliance reports
    $reportsResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/reports" -Description "Get compliance reports"
    if ($reportsResponse) {
        Write-Success "Compliance reports endpoint working"

        if ($reportsResponse.data) {
            Write-Info "  Reports count: $($reportsResponse.data.Count)"
        }
    }

    # Test compliance violations
    $violationsResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/violations" -Description "Get compliance violations"
    if ($violationsResponse) {
        Write-Success "Compliance violations endpoint working"

        if ($violationsResponse.data) {
            Write-Info "  Violations count: $($violationsResponse.data.Count)"
        }
    }

    # Test compliance certifications
    $certificationsResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/certifications" -Description "Get compliance certifications"
    if ($certificationsResponse) {
        Write-Success "Compliance certifications endpoint working"

        if ($certificationsResponse.data) {
            Write-Info "  Certifications count: $($certificationsResponse.data.Count)"
        }
    }
}

# Test enterprise organization endpoints
function Test-EnterpriseEndpoints {
    if (-not $TestEnterprise) {
        Write-Warning "Skipping enterprise organization tests"
        return
    }

    Write-Info "Testing enterprise organization endpoints..."

    $testData = Generate-TestData

    # Test create enterprise organization
    $createResponse = Test-ApiEndpoint -Method "POST" -Endpoint "/api/admin/compliance/enterprise" -Body $testData.Organization -Description "Create enterprise organization"
    if ($createResponse) {
        Write-Success "Enterprise organization creation working"

        if ($createResponse.data) {
            $orgId = $createResponse.data.org_id
            Write-Info "  Created organization ID: $orgId"
        }
    }

    # Test get enterprise organization
    $getResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/enterprise" -Description "Get enterprise organization"
    if ($getResponse) {
        Write-Success "Enterprise organization retrieval working"

        if ($getResponse.data) {
            Write-Info "  Organization name: $($getResponse.data.org_name)"
            Write-Info "  Organization domain: $($getResponse.data.org_domain)"
            Write-Info "  Enterprise enabled: $($getResponse.data.enterprise_enabled)"
            Write-Info "  Compliance enabled: $($getResponse.data.compliance_enabled)"
        }
    }
}

# Test violation management endpoints
function Test-ViolationEndpoints {
    if (-not $TestViolations) {
        Write-Warning "Skipping violation management tests"
        return
    }

    Write-Info "Testing violation management endpoints..."

    # First get violations to find one to test with
    $violationsResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/violations" -Description "Get violations for testing"

    if ($violationsResponse -and $violationsResponse.data -and $violationsResponse.data.Count -gt 0) {
        $testViolation = $violationsResponse.data[0]
        $violationId = $testViolation.violation_id

        Write-Info "  Using violation ID: $violationId for testing"

        $testData = Generate-TestData

        # Test acknowledge violation
        $ackResponse = Test-ApiEndpoint -Method "POST" -Endpoint "/api/admin/compliance/violations/$violationId/acknowledge" -Body $testData.Acknowledgment -Description "Acknowledge violation"
        if ($ackResponse) {
            Write-Success "Violation acknowledgment working"
        }

        # Test resolve violation
        $resolveResponse = Test-ApiEndpoint -Method "POST" -Endpoint "/api/admin/compliance/violations/$violationId/resolve" -Body $testData.Resolution -Description "Resolve violation"
        if ($resolveResponse) {
            Write-Success "Violation resolution working"
        }
    } else {
        Write-Warning "  No violations found for testing - skipping violation management tests"
    }
}

# Test certification management endpoints
function Test-CertificationEndpoints {
    if (-not $TestCertifications) {
        Write-Warning "Skipping certification management tests"
        return
    }

    Write-Info "Testing certification management endpoints..."

    $testData = Generate-TestData

    # Test generate certification
    $generateResponse = Test-ApiEndpoint -Method "POST" -Endpoint "/api/admin/compliance/certifications/generate" -Body $testData.Certification -Description "Generate certification"
    if ($generateResponse) {
        Write-Success "Certification generation working"

        if ($generateResponse.data) {
            $certificationId = $generateResponse.data.certification_id
            Write-Info "  Generated certification ID: $certificationId"

            # Test approve certification
            $approveResponse = Test-ApiEndpoint -Method "POST" -Endpoint "/api/admin/compliance/certifications/$certificationId/approve" -Body $testData.Approval -Description "Approve certification"
            if ($approveResponse) {
                Write-Success "Certification approval working"
            }
        }
    }
}

# Test compliance framework endpoints
function Test-FrameworkEndpoints {
    Write-Info "Testing compliance framework endpoints..."

    # Test get compliance frameworks (this would be a custom endpoint if implemented)
    # For now, we'll test the status endpoint which includes framework information
    $statusResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/status" -Description "Get compliance status with frameworks"
    if ($statusResponse -and $statusResponse.data -and $statusResponse.data.active_frameworks) {
        Write-Success "Compliance frameworks accessible"

        foreach ($framework in $statusResponse.data.active_frameworks) {
            Write-Info "  Framework: $($framework.framework_name) v$($framework.framework_version)"
        }
    }
}

# Test error handling
function Test-ErrorHandling {
    Write-Info "Testing error handling..."

    # Test invalid endpoint
    $invalidResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/invalid" -Description "Test invalid endpoint"
    if (-not $invalidResponse) {
        Write-Success "Error handling working for invalid endpoints"
    }

    # Test without authentication (if no token provided)
    if (-not $AdminToken) {
        $noAuthResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/status" -Description "Test without authentication"
        if (-not $noAuthResponse) {
            Write-Success "Authentication required properly enforced"
        }
    }
}

# Test configuration
function Test-Configuration {
    Write-Info "Testing configuration..."

    # Check if enterprise compliance is enabled
    $statusResponse = Test-ApiEndpoint -Method "GET" -Endpoint "/api/admin/compliance/status" -Description "Check compliance configuration"
    if ($statusResponse -and $statusResponse.data) {
        Write-Info "  Enterprise compliance enabled: $($statusResponse.data.enterprise_enabled)"
        Write-Info "  Compliance enabled: $($statusResponse.data.compliance_enabled)"

        if ($statusResponse.data.enterprise_enabled) {
            Write-Success "Enterprise compliance features are enabled"
        } else {
            Write-Warning "Enterprise compliance features are disabled"
        }
    }
}

# Show summary
function Show-Summary {
    Write-Info "=== Compliance System Test Summary ==="
    Write-Info "API URL: $ApiUrl"
    Write-Info "Admin Token: $(if ($AdminToken) { "Provided" } else { "Not provided" })"
    Write-Info "Enterprise Tests: $(if ($TestEnterprise) { "Enabled" } else { "Disabled" })"
    Write-Info "Violation Tests: $(if ($TestViolations) { "Enabled" } else { "Disabled" })"
    Write-Info "Certification Tests: $(if ($TestCertifications) { "Enabled" } else { "Disabled" })"
    Write-Info "====================================="
}

# Main test execution
function Main {
    Write-ColorOutput "🚀 Starting Compliance System Tests (Micro-Iteration 4.30)" "Magenta"
    Write-Info "Testing automated compliance and retention certification features"

    Show-Summary

    # Test basic endpoints
    Test-ComplianceStatusEndpoints

    # Test enterprise features
    Test-EnterpriseEndpoints

    # Test violation management
    Test-ViolationEndpoints

    # Test certification management
    Test-CertificationEndpoints

    # Test framework endpoints
    Test-FrameworkEndpoints

    # Test error handling
    Test-ErrorHandling

    # Test configuration
    Test-Configuration

    Write-ColorOutput "🎉 Compliance System Tests Completed!" "Magenta"
    Write-Info "Check the output above for test results and any issues that need attention."
}

# Run the main function
Main






