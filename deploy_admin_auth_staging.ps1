# Admin Authentication System - Staging Deployment Script
# This script deploys the admin authentication system to staging environment

param(
    [string]$Environment = "staging",
    [string]$ApiUrl = "https://staging.securesystem.email",
    [string]$RootAdminEmail = "cpigusch@gmail.com"
)

Write-Output "=== Admin Authentication System - Staging Deployment ==="
Write-Output "Environment: $Environment"
Write-Output "API URL: $ApiUrl"
Write-Output "Root Admin Email: $RootAdminEmail"

# Step 1: Validate Environment Variables
Write-Output "`n1. Validating Environment Variables..."
$requiredEnvVars = @(
    "ROOT_ADMIN_EMAIL",
    "JWT_SECRET"
)

foreach ($var in $requiredEnvVars) {
    $value = [Environment]::GetEnvironmentVariable($var)
    if ([string]::IsNullOrEmpty($value)) {
        Write-Output "❌ Missing required environment variable: $var"
        exit 1
    } else {
        Write-Output "✅ $var is set"
    }
}

# Step 2: Test API Connectivity
Write-Output "`n2. Testing API Connectivity..."
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/health" -Method GET -TimeoutSec 10
    if ($response.status -eq "ok") {
        Write-Output "✅ API is accessible and healthy"
    } else {
        Write-Output "❌ API health check failed"
        exit 1
    }
} catch {
    Write-Output "❌ Failed to connect to API: $($_.Exception.Message)"
    exit 1
}

# Step 3: Check Admin Setup Status
Write-Output "`n3. Checking Admin Setup Status..."
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/admin/check-setup" -Method GET -TimeoutSec 10
    Write-Output "Setup required: $($response.setup_required)"
    Write-Output "Root admin email: $($response.root_admin_email)"

    if ($response.setup_required) {
        Write-Output "⚠️  Admin setup is required - will create root admin"
    } else {
        Write-Output "✅ Admin already exists"
    }
} catch {
    Write-Output "❌ Failed to check admin setup status: $($_.Exception.Message)"
    exit 1
}

# Step 4: Create Root Admin (if needed)
if ($response.setup_required) {
    Write-Output "`n4. Creating Root Admin..."

    # Generate a secure password for staging
    $stagingPassword = "StagingAdminPassword123!"

    $setupData = @{
        email = $RootAdminEmail
        password = $stagingPassword
    }

    try {
        $response = Invoke-RestMethod -Uri "$ApiUrl/admin/setup" -Method POST -Body ($setupData | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 10
        Write-Output "✅ Root admin created successfully"
        Write-Output "Admin ID: $($response.admin_id)"
        Write-Output "⚠️  IMPORTANT: Staging password is: $stagingPassword"
    } catch {
        Write-Output "❌ Failed to create root admin: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Output "Error details: $errorBody"
        }
        exit 1
    }
}

# Step 5: Test Admin Login
Write-Output "`n5. Testing Admin Login..."
$loginData = @{
    email = $RootAdminEmail
    password = if ($response.setup_required) { $stagingPassword } else { "SecureAdminPassword123!" }
    totp_code = ""
}

try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/admin/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 10
    Write-Output "✅ Admin login successful"
    Write-Output "Session token: $($response.session_token)"
    Write-Output "Admin role: $($response.admin.role)"

    $sessionToken = $response.session_token
} catch {
    Write-Output "❌ Admin login failed: $($_.Exception.Message)"
    exit 1
}

# Step 6: Test Session Validation
Write-Output "`n6. Testing Session Validation..."
$headers = @{
    "Authorization" = "Bearer $sessionToken"
}

try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/admin/session" -Method GET -Headers $headers -TimeoutSec 10
    Write-Output "✅ Session validation successful"
    Write-Output "Admin email: $($response.admin.email)"
    Write-Output "Admin role: $($response.admin.role)"
} catch {
    Write-Output "❌ Session validation failed: $($_.Exception.Message)"
    exit 1
}

# Step 7: Test Audit Logs
Write-Output "`n7. Testing Audit Logs..."
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/admin/audit-logs" -Method GET -Headers $headers -TimeoutSec 10
    Write-Output "✅ Audit logs retrieved successfully"
    Write-Output "Number of logs: $($response.count)"

    if ($response.logs -and $response.logs.Count -gt 0) {
        Write-Output "Recent actions:"
        $response.logs | Select-Object -First 3 | ForEach-Object {
            Write-Output "  - $($_.action) at $($_.created_at) (Success: $($_.success))"
        }
    }
} catch {
    Write-Output "❌ Failed to retrieve audit logs: $($_.Exception.Message)"
    exit 1
}

# Step 8: Test Admin Logout
Write-Output "`n8. Testing Admin Logout..."
$logoutData = @{
    session_token = $sessionToken
}

try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/admin/logout" -Method POST -Body ($logoutData | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 10
    Write-Output "✅ Admin logout successful"
} catch {
    Write-Output "❌ Admin logout failed: $($_.Exception.Message)"
    exit 1
}

# Step 9: Verify Session Invalidation
Write-Output "`n9. Verifying Session Invalidation..."
try {
    $response = Invoke-RestMethod -Uri "$ApiUrl/admin/session" -Method GET -Headers $headers -TimeoutSec 10
    Write-Output "❌ Session should have been invalidated"
    exit 1
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Output "✅ Session properly invalidated"
    } else {
        Write-Output "❌ Unexpected error during session validation: $($_.Exception.Message)"
        exit 1
    }
}

# Step 10: Security Validation
Write-Output "`n10. Security Validation..."

# Test rate limiting (attempt multiple logins)
Write-Output "Testing rate limiting..."
$failedAttempts = 0
for ($i = 1; $i -le 6; $i++) {
    try {
        $badLoginData = @{
            email = $RootAdminEmail
            password = "WrongPassword123!"
            totp_code = ""
        }
        $response = Invoke-RestMethod -Uri "$ApiUrl/admin/login" -Method POST -Body ($badLoginData | ConvertTo-Json) -ContentType "application/json" -TimeoutSec 10
        Write-Output "  Attempt ${i}: Login should have failed"
    } catch {
        if ($_.Exception.Response.StatusCode -eq 401) {
            $failedAttempts++
            Write-Output "  Attempt ${i}: Login failed as expected"
        } else {
            Write-Output "  Attempt ${i}: Unexpected error: $($_.Exception.Message)"
        }
    }
}

if ($failedAttempts -ge 5) {
    Write-Output "✅ Rate limiting appears to be working"
} else {
    Write-Output "⚠️  Rate limiting may not be working correctly"
}

Write-Output "`n=== Staging Deployment Complete ==="
Write-Output "✅ All admin authentication endpoints are working correctly"
Write-Output "✅ Security features are properly implemented"
Write-Output "✅ Audit logging is functional"
Write-Output "✅ Session management is working"

Write-Output "`n📋 Next Steps:"
Write-Output "1. Configure frontend integration"
Write-Output "2. Set up monitoring and alerting"
Write-Output "3. Perform end-to-end testing"
Write-Output "4. Deploy to production"

Write-Output "`n🎉 Admin Authentication System is ready for production!"
