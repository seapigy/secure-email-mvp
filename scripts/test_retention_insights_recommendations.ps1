# =============================================================================
# SECURE EMAIL MVP - RETENTION INSIGHTS & RECOMMENDATIONS TEST SCRIPT
# =============================================================================
# Test script for Micro-Iteration 4.27: Intelligent Retention Insights & Proactive Policy Recommendations
# =============================================================================

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AdminToken = "",
    [switch]$SkipSetup,
    [switch]$SkipCleanup
)

# Colors for output
$Red = "Red"
$Green = "Green"
$Yellow = "Yellow"
$Blue = "Blue"

function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✅ $Message" $Green
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" $Red
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" $Blue
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" $Yellow
}

# Test configuration
$TestResults = @{
    Total = 0
    Passed = 0
    Failed = 0
    Skipped = 0
}

function Test-Endpoint {
    param(
        [string]$Name,
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [string]$Body = "",
        [scriptblock]$Validation = {}
    )
    
    $TestResults.Total++
    Write-Info "Testing: $Name"
    
    try {
        $uri = "$BaseUrl$Endpoint"
        $headers["Content-Type"] = "application/json"
        
        if ($AdminToken) {
            $headers["Authorization"] = "Bearer $AdminToken"
        }
        
        $params = @{
            Uri = $uri
            Method = $Method
            Headers = $headers
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params -ErrorAction Stop
        
        # Run validation if provided
        if ($Validation) {
            & $Validation $response
        }
        
        Write-Success "$Name - PASSED"
        $TestResults.Passed++
        
    } catch {
        Write-Error "$Name - FAILED: $($_.Exception.Message)"
        $TestResults.Failed++
    }
}

function Test-InsightsEndpoints {
    Write-ColorOutput "`n🔍 Testing Retention Insights Endpoints" $Blue
    
    # Test GET /api/admin/email/retention-insights
    Test-Endpoint -Name "Get Retention Insights" -Method "GET" -Endpoint "/api/admin/email/retention-insights" -Validation {
        param($response)
        if (-not $response.insights) { throw "Missing insights array" }
        if (-not $response.total_count) { throw "Missing total_count" }
    }
    
    # Test GET /api/admin/email/retention-insights with filters
    Test-Endpoint -Name "Get Retention Insights with Filters" -Method "GET" -Endpoint "/api/admin/email/retention-insights?insight_type=daily_rollup&limit=10" -Validation {
        param($response)
        if (-not $response.insights) { throw "Missing insights array" }
        if ($response.limit -ne 10) { throw "Limit not applied correctly" }
    }
    
    # Test GET /api/admin/email/retention-insights/trends
    Test-Endpoint -Name "Get Retention Trends" -Method "GET" -Endpoint "/api/admin/email/retention-insights/trends" -Validation {
        param($response)
        if (-not $response.trend_analysis) { throw "Missing trend_analysis" }
        if (-not $response.date_range) { throw "Missing date_range" }
    }
    
    # Test GET /api/admin/email/retention-insights/trends with custom date range
    Test-Endpoint -Name "Get Retention Trends with Date Range" -Method "GET" -Endpoint "/api/admin/email/retention-insights/trends?start_date=2024-01-01&end_date=2024-01-31" -Validation {
        param($response)
        if (-not $response.trend_analysis) { throw "Missing trend_analysis" }
        if ($response.date_range.start_date -ne "2024-01-01") { throw "Start date not applied" }
        if ($response.date_range.end_date -ne "2024-01-31") { throw "End date not applied" }
    }
    
    # Test CSV export
    Test-Endpoint -Name "Export Trends CSV" -Method "GET" -Endpoint "/api/admin/email/retention-insights/trends?export_csv=true" -Validation {
        param($response)
        if (-not $response) { throw "No CSV content returned" }
    }
}

function Test-RecommendationsEndpoints {
    Write-ColorOutput "`n💡 Testing Retention Recommendations Endpoints" $Blue
    
    # Test GET /api/admin/email/retention-recommendations
    Test-Endpoint -Name "Get Retention Recommendations" -Method "GET" -Endpoint "/api/admin/email/retention-recommendations" -Validation {
        param($response)
        if (-not $response.recommendations) { throw "Missing recommendations array" }
        if (-not $response.total_count) { throw "Missing total_count" }
    }
    
    # Test GET /api/admin/email/retention-recommendations with filters
    Test-Endpoint -Name "Get Recommendations with Filters" -Method "GET" -Endpoint "/api/admin/email/retention-recommendations?priority=high&status=pending" -Validation {
        param($response)
        if (-not $response.recommendations) { throw "Missing recommendations array" }
    }
    
    # Test POST /api/admin/email/retention-recommendations/apply (preview mode)
    $previewBody = @{
        recommendation_id = 1
        preview = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Apply Recommendation (Preview)" -Method "POST" -Endpoint "/api/admin/email/retention-recommendations/apply" -Body $previewBody -Validation {
        param($response)
        if (-not $response.success) { throw "Preview application failed" }
        if (-not $response.preview_mode) { throw "Preview mode not set" }
    }
    
    # Test POST /api/admin/email/retention-recommendations/apply (apply mode)
    $applyBody = @{
        recommendation_id = 1
        preview = $false
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Apply Recommendation (Apply)" -Method "POST" -Endpoint "/api/admin/email/retention-recommendations/apply" -Body $applyBody -Validation {
        param($response)
        if (-not $response.success) { throw "Recommendation application failed" }
        if ($response.preview_mode) { throw "Preview mode should be false" }
    }
}

function Test-DataValidation {
    Write-ColorOutput "`n🔍 Testing Data Validation" $Blue
    
    # Test invalid recommendation ID
    $invalidBody = @{
        recommendation_id = 0
        preview = $true
    } | ConvertTo-Json
    
    Test-Endpoint -Name "Invalid Recommendation ID" -Method "POST" -Endpoint "/api/admin/email/retention-recommendations/apply" -Body $invalidBody -Validation {
        param($response)
        # Should return an error
        if ($response.success) { throw "Should have failed with invalid ID" }
    }
    
    # Test missing authentication
    Test-Endpoint -Name "Unauthenticated Access" -Method "GET" -Endpoint "/api/admin/email/retention-insights" -Headers @{} -Validation {
        param($response)
        # Should return an error
        if (-not $response.error) { throw "Should have returned authentication error" }
    }
}

function Test-Performance {
    Write-ColorOutput "`n⚡ Testing Performance" $Blue
    
    # Test large limit
    Test-Endpoint -Name "Large Limit Handling" -Method "GET" -Endpoint "/api/admin/email/retention-insights?limit=1000" -Validation {
        param($response)
        if ($response.limit -gt 1000) { throw "Limit should be capped at 1000" }
    }
    
    # Test pagination
    Test-Endpoint -Name "Pagination" -Method "GET" -Endpoint "/api/admin/email/retention-insights?limit=10&offset=20" -Validation {
        param($response)
        if ($response.limit -ne 10) { throw "Limit not applied" }
        if ($response.offset -ne 20) { throw "Offset not applied" }
    }
}

function Show-TestResults {
    Write-ColorOutput "`n📊 Test Results Summary" $Blue
    Write-ColorOutput "================================" $Blue
    Write-ColorOutput "Total Tests: $($TestResults.Total)" $Blue
    Write-ColorOutput "Passed: $($TestResults.Passed)" $Green
    Write-ColorOutput "Failed: $($TestResults.Failed)" $Red
    Write-ColorOutput "Skipped: $($TestResults.Skipped)" $Yellow
    
    $successRate = if ($TestResults.Total -gt 0) { [math]::Round(($TestResults.Passed / $TestResults.Total) * 100, 2) } else { 0 }
    Write-ColorOutput "Success Rate: $successRate%" $(if ($successRate -ge 80) { $Green } else { $Red })
}

function Main {
    Write-ColorOutput "🚀 Starting Retention Insights & Recommendations Tests" $Blue
    Write-ColorOutput "=====================================================" $Blue
    Write-ColorOutput "Base URL: $BaseUrl" $Blue
    Write-ColorOutput "Admin Token: $(if ($AdminToken) { "Provided" } else { "Not provided" })" $Blue
    
    if (-not $SkipSetup) {
        Write-Info "Setting up test environment..."
        # Add any setup logic here if needed
    }
    
    try {
        # Run all test suites
        Test-InsightsEndpoints
        Test-RecommendationsEndpoints
        Test-DataValidation
        Test-Performance
        
    } catch {
        Write-Error "Test execution failed: $($_.Exception.Message)"
    } finally {
        if (-not $SkipCleanup) {
            Write-Info "Cleaning up test environment..."
            # Add any cleanup logic here if needed
        }
    }
    
    Show-TestResults
    
    # Exit with appropriate code
    if ($TestResults.Failed -gt 0) {
        Write-Error "Some tests failed. Please review the results above."
        exit 1
    } else {
        Write-Success "All tests passed! 🎉"
        exit 0
    }
}

# Run the main function
Main





