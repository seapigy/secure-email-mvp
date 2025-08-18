# =============================================================================
# SECURE EMAIL MVP - USER COMPLIANCE TRANSPARENCY TEST SCRIPT
# =============================================================================
# PowerShell script to test Micro-Iteration 4.31: User Transparency Layer
# Tests user-facing compliance endpoints and admin transparency controls

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$AdminToken = "",
    [string]$UserToken = "",
    [switch]$EnableUserPortal = $false
)

# Test configuration
$TestConfig = @{
    BaseUrl = $BaseUrl
    AdminToken = $AdminToken
    UserToken = $UserToken
    EnableUserPortal = $EnableUserPortal
}

# Test results tracking
$TestResults = @{
    TotalTests = 0
    PassedTests = 0
    FailedTests = 0
    Errors = @()
}

# Utility functions
function Write-TestHeader {
    param([string]$Title)
    Write-Host "`n" -NoNewline
    Write-Host "=" * 80 -ForegroundColor Cyan
    Write-Host " $Title" -ForegroundColor Cyan
    Write-Host "=" * 80 -ForegroundColor Cyan
}

function Write-TestResult {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Message = ""
    )
    
    $TestResults.TotalTests++
    
    if ($Success) {
        $TestResults.PassedTests++
        Write-Host "✅ PASS: $TestName" -ForegroundColor Green
        if ($Message) { Write-Host "   $Message" -ForegroundColor Gray }
    } else {
        $TestResults.FailedTests++
        Write-Host "❌ FAIL: $TestName" -ForegroundColor Red
        if ($Message) { Write-Host "   $Message" -ForegroundColor Red }
    }
}

function Invoke-ApiRequest {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [string]$Body = ""
    )
    
    $uri = "$($TestConfig.BaseUrl)$Endpoint"
    
    $requestHeaders = @{
        "Content-Type" = "application/json"
    }
    
    foreach ($key in $Headers.Keys) {
        $requestHeaders[$key] = $Headers[$key]
    }
    
    try {
        if ($Method -eq "GET" -or $Body -eq "") {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $requestHeaders -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $requestHeaders -Body $Body -ErrorAction Stop
        }
        return @{
            Success = $true
            Data = $response
            StatusCode = 200
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        $errorMessage = $_.Exception.Message
        
        try {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
        } catch {
            $errorBody = "Unable to read error response"
        }
        
        return @{
            Success = $false
            StatusCode = $statusCode
            Error = $errorMessage
            ErrorBody = $errorBody
        }
    }
}

# Test data generation
function New-TestUserData {
    return @{
        user_id = "test-user-$(Get-Random)"
        domain = "testcompany.com"
        email = "testuser@testcompany.com"
    }
}

function New-TestEnterpriseData {
    return @{
        org_name = "Test Enterprise Organization"
        org_domain = "testcompany.com"
        org_type = "healthcare"
        primary_framework_id = 1  # GDPR
        compliance_contact_email = "compliance@testcompany.com"
        compliance_contact_name = "Test Compliance Officer"
    }
}

# Test functions
function Test-UserCompliancePortalDisabled {
    Write-TestHeader "Testing User Compliance Portal (Disabled)"
    
    # Test user compliance status endpoint when portal is disabled
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/status" -Headers @{
        "Authorization" = "Bearer $($TestConfig.UserToken)"
    }
    
    $expectedStatus = 503  # Service Unavailable
    $success = -not $response.Success -and $response.StatusCode -eq $expectedStatus
    
    Write-TestResult -TestName "User Compliance Portal Disabled" -Success $success -Message "Expected 503 Service Unavailable when portal is disabled"
}

function Test-UserComplianceStatus {
    Write-TestHeader "Testing User Compliance Status Endpoint"
    
    # Test user compliance status endpoint
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/status" -Headers @{
        "Authorization" = "Bearer $($TestConfig.UserToken)"
    }
    
    if ($response.Success) {
        $data = $response.Data
        
        # Validate response structure
        $validStructure = $data.PSObject.Properties.Name -contains "success" -and
                         $data.PSObject.Properties.Name -contains "data" -and
                         $data.success -eq $true
        
        Write-TestResult -TestName "User Compliance Status Response Structure" -Success $validStructure
        
        if ($validStructure -and $data.data) {
            $userData = $data.data
            
            # Validate user compliance status fields
            $validFields = $userData.PSObject.Properties.Name -contains "user_id" -and
                          $userData.PSObject.Properties.Name -contains "domain" -and
                          $userData.PSObject.Properties.Name -contains "is_enterprise_user" -and
                          $userData.PSObject.Properties.Name -contains "applicable_policies" -and
                          $userData.PSObject.Properties.Name -contains "compliance_score" -and
                          $userData.PSObject.Properties.Name -contains "transparency_settings"
            
            Write-TestResult -TestName "User Compliance Status Data Fields" -Success $validFields
            
            # Validate transparency settings
            if ($userData.transparency_settings) {
                $validSettings = $userData.transparency_settings.PSObject.Properties.Name -contains "show_retention_rules" -and
                                $userData.transparency_settings.PSObject.Properties.Name -contains "show_compliance_frameworks" -and
                                $userData.transparency_settings.PSObject.Properties.Name -contains "show_violations" -and
                                $userData.transparency_settings.PSObject.Properties.Name -contains "cache_ttl_minutes"
                
                Write-TestResult -TestName "Transparency Settings Structure" -Success $validSettings
            }
        }
    } else {
        Write-TestResult -TestName "User Compliance Status Request" -Success $false -Message "Request failed: $($response.StatusCode) - $($response.Error)"
    }
}

function Test-UserCompliancePolicies {
    Write-TestHeader "Testing User Compliance Policies Endpoint"
    
    # Test user compliance policies endpoint
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/policies" -Headers @{
        "Authorization" = "Bearer $($TestConfig.UserToken)"
    }
    
    if ($response.Success) {
        $data = $response.Data
        
        # Validate response structure
        $validStructure = $data.PSObject.Properties.Name -contains "success" -and
                         $data.PSObject.Properties.Name -contains "data" -and
                         $data.success -eq $true
        
        Write-TestResult -TestName "User Compliance Policies Response Structure" -Success $validStructure
        
        if ($validStructure -and $data.data) {
            $policies = $data.data
            
            # Validate policies array
            $validArray = $policies -is [array]
            Write-TestResult -TestName "Policies Array Structure" -Success $validArray
            
            if ($validArray -and $policies.Count -gt 0) {
                $firstPolicy = $policies[0]
                
                # Validate policy structure
                $validPolicy = $firstPolicy.PSObject.Properties.Name -contains "policy_id" -and
                              $firstPolicy.PSObject.Properties.Name -contains "policy_name" -and
                              $firstPolicy.PSObject.Properties.Name -contains "policy_type" -and
                              $firstPolicy.PSObject.Properties.Name -contains "retention_period_days" -and
                              $firstPolicy.PSObject.Properties.Name -contains "human_readable_summary"
                
                Write-TestResult -TestName "Policy Data Structure" -Success $validPolicy
                
                # Validate human-readable summary
                if ($firstPolicy.human_readable_summary) {
                    $validSummary = $firstPolicy.human_readable_summary.Length -gt 0 -and
                                   $firstPolicy.human_readable_summary -match "days"
                    
                    Write-TestResult -TestName "Human-Readable Policy Summary" -Success $validSummary
                }
            }
        }
    } else {
        Write-TestResult -TestName "User Compliance Policies Request" -Success $false -Message "Request failed: $($response.StatusCode) - $($response.Error)"
    }
}

function Test-AdminTransparencySettings {
    Write-TestHeader "Testing Admin Transparency Settings Endpoints"
    
    # Test GET admin transparency settings
    $getResponse = Invoke-ApiRequest -Method "GET" -Endpoint "/api/admin/compliance/settings/user-transparency" -Headers @{
        "Authorization" = "Bearer $($TestConfig.AdminToken)"
    }
    
    if ($getResponse.Success) {
        $data = $getResponse.Data
        
        # Validate response structure
        $validStructure = $data.PSObject.Properties.Name -contains "success" -and
                         $data.PSObject.Properties.Name -contains "data" -and
                         $data.success -eq $true
        
        Write-TestResult -TestName "Admin GET Transparency Settings Response" -Success $validStructure
        
        if ($validStructure -and $data.data) {
            $settings = $data.data
            
            # Validate settings structure
            $validSettings = $settings.PSObject.Properties.Name -contains "show_retention_rules" -and
                            $settings.PSObject.Properties.Name -contains "show_compliance_frameworks" -and
                            $settings.PSObject.Properties.Name -contains "show_violations" -and
                            $settings.PSObject.Properties.Name -contains "show_compliance_rules" -and
                            $settings.PSObject.Properties.Name -contains "cache_ttl_minutes"
            
            Write-TestResult -TestName "Admin Transparency Settings Structure" -Success $validSettings
            
            # Test PUT admin transparency settings
            $updateSettings = @{
                show_retention_rules = $true
                show_compliance_frameworks = $true
                show_violations = $false
                show_compliance_rules = $true
                cache_ttl_minutes = 30
            }
            
            $putResponse = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/admin/compliance/settings/user-transparency" -Headers @{
                "Authorization" = "Bearer $($TestConfig.AdminToken)"
            } -Body ($updateSettings | ConvertTo-Json)
            
            if ($putResponse.Success) {
                $putData = $putResponse.Data
                $validPutResponse = $putData.PSObject.Properties.Name -contains "success" -and
                                   $putData.success -eq $true
                
                Write-TestResult -TestName "Admin PUT Transparency Settings" -Success $validPutResponse
            } else {
                Write-TestResult -TestName "Admin PUT Transparency Settings" -Success $false -Message "Request failed: $($putResponse.StatusCode)"
            }
        }
    } else {
        Write-TestResult -TestName "Admin GET Transparency Settings" -Success $false -Message "Request failed: $($getResponse.StatusCode) - $($getResponse.Error)"
    }
}

function Test-UserComplianceUnauthorized {
    Write-TestHeader "Testing User Compliance Unauthorized Access"
    
    # Test without authentication
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/status"
    
    $expectedStatus = 401  # Unauthorized
    $success = -not $response.Success -and $response.StatusCode -eq $expectedStatus
    
    Write-TestResult -TestName "User Compliance Status Unauthorized" -Success $success -Message "Expected 401 Unauthorized without authentication"
    
    # Test with invalid token
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/user/compliance/status" -Headers @{
        "Authorization" = "Bearer invalid-token"
    }
    
    $expectedStatus = 401  # Unauthorized
    $success = -not $response.Success -and $response.StatusCode -eq $expectedStatus
    
    Write-TestResult -TestName "User Compliance Status Invalid Token" -Success $success -Message "Expected 401 Unauthorized with invalid token"
}

function Test-AdminTransparencyUnauthorized {
    Write-TestHeader "Testing Admin Transparency Settings Unauthorized Access"
    
    # Test without authentication
    $response = Invoke-ApiRequest -Method "GET" -Endpoint "/api/admin/compliance/settings/user-transparency"
    
    $expectedStatus = 401  # Unauthorized
    $success = -not $response.Success -and $response.StatusCode -eq $expectedStatus
    
    Write-TestResult -TestName "Admin Transparency Settings Unauthorized" -Success $success -Message "Expected 401 Unauthorized without authentication"
}

function Test-TransparencySettingsValidation {
    Write-TestHeader "Testing Transparency Settings Validation"
    
    # Test invalid cache TTL (too low)
    $invalidSettings = @{
        show_retention_rules = $true
        show_compliance_frameworks = $true
        show_violations = $false
        show_compliance_rules = $true
        cache_ttl_minutes = 0  # Invalid: must be >= 1
    }
    
    $response = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/admin/compliance/settings/user-transparency" -Headers @{
        "Authorization" = "Bearer $($TestConfig.AdminToken)"
    } -Body ($invalidSettings | ConvertTo-Json)
    
    $expectedStatus = 400  # Bad Request
    $success = -not $response.Success -and $response.StatusCode -eq $expectedStatus
    
    Write-TestResult -TestName "Invalid Cache TTL Validation" -Success $success -Message "Expected 400 Bad Request for invalid cache TTL"
    
    # Test invalid cache TTL (too high)
    $invalidSettings.cache_ttl_minutes = 1441  # Invalid: must be <= 1440
    
    $response = Invoke-ApiRequest -Method "PUT" -Endpoint "/api/admin/compliance/settings/user-transparency" -Headers @{
        "Authorization" = "Bearer $($TestConfig.AdminToken)"
    } -Body ($invalidSettings | ConvertTo-Json)
    
    $expectedStatus = 400  # Bad Request
    $success = -not $response.Success -and $response.StatusCode -eq $expectedStatus
    
    Write-TestResult -TestName "Invalid Cache TTL (High) Validation" -Success $success -Message "Expected 400 Bad Request for invalid cache TTL"
}

# Main test execution
function Start-UserComplianceTests {
    Write-Host "`n" -NoNewline
    Write-Host "🚀 Starting Micro-Iteration 4.31: User Transparency Layer Tests" -ForegroundColor Yellow
    Write-Host "Base URL: $($TestConfig.BaseUrl)" -ForegroundColor Gray
    
    # Check if user portal is enabled
    if (-not $TestConfig.EnableUserPortal) {
        Write-Host "`n⚠️  User Compliance Portal is disabled. Some tests will be skipped." -ForegroundColor Yellow
        Test-UserCompliancePortalDisabled
    } else {
        Write-Host "`n✅ User Compliance Portal is enabled. Running full test suite." -ForegroundColor Green
        
        # Run user compliance tests
        Test-UserComplianceStatus
        Test-UserCompliancePolicies
        Test-UserComplianceUnauthorized
        
        # Run admin transparency tests
        Test-AdminTransparencySettings
        Test-AdminTransparencyUnauthorized
        Test-TransparencySettingsValidation
    }
    
    # Display test results
    Write-TestHeader "Test Results Summary"
    Write-Host "Total Tests: $($TestResults.TotalTests)" -ForegroundColor White
    Write-Host "Passed: $($TestResults.PassedTests)" -ForegroundColor Green
    Write-Host "Failed: $($TestResults.FailedTests)" -ForegroundColor Red
    
    if ($TestResults.FailedTests -gt 0) {
        Write-Host "`n❌ Some tests failed. Check the output above for details." -ForegroundColor Red
        exit 1
    } else {
        Write-Host "`n✅ All tests passed!" -ForegroundColor Green
    }
}

# Script execution
if (-not $TestConfig.UserToken -or -not $TestConfig.AdminToken) {
    Write-Host "❌ Error: Both UserToken and AdminToken are required for testing." -ForegroundColor Red
    Write-Host "Usage: .\test_user_compliance_transparency.ps1 -UserToken 'your-user-token' -AdminToken 'your-admin-token' [-EnableUserPortal]" -ForegroundColor Yellow
    exit 1
}

Start-UserComplianceTests
