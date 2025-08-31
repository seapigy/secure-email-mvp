# Test Admin Authentication System
# This script tests the admin authentication endpoints

$baseUrl = "http://localhost:8080"

Write-Output "=== Testing Admin Authentication System ==="

# Test 1: Check if admin setup is required
Write-Output "`n1. Checking if admin setup is required..."
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/admin/check-setup" -Method GET -ContentType "application/json"
    Write-Output "Setup required: $($response.setup_required)"
    Write-Output "Root admin email: $($response.root_admin_email)"
} catch {
    Write-Output "Error checking setup: $($_.Exception.Message)"
}

# Test 2: Create root admin (if setup is required)
Write-Output "`n2. Attempting to create root admin..."
$setupData = @{
    email = "cpigusch@gmail.com"
    password = "SecureAdminPassword123!"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/admin/setup" -Method POST -Body ($setupData | ConvertTo-Json) -ContentType "application/json"
    Write-Output "Admin created successfully: $($response.success)"
    Write-Output "Admin ID: $($response.admin_id)"
} catch {
    Write-Output "Error creating admin: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Output "Error details: $errorBody"
    }
}

# Test 3: Admin login
Write-Output "`n3. Testing admin login..."
$loginData = @{
    email = "cpigusch@gmail.com"
    password = "SecureAdminPassword123!"
    totp_code = ""  # TOTP not enabled initially
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/admin/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json"
    Write-Output "Login successful: $($response.success)"
    Write-Output "Session token: $($response.session_token)"
    Write-Output "Admin role: $($response.admin.role)"

    $sessionToken = $response.session_token
} catch {
    Write-Output "Error logging in: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Output "Error details: $errorBody"
    }
    $sessionToken = $null
}

# Test 4: Validate session (if login was successful)
if ($sessionToken) {
    Write-Output "`n4. Testing session validation..."
    $headers = @{
        "Authorization" = "Bearer $sessionToken"
    }

    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/session" -Method GET -Headers $headers -ContentType "application/json"
        Write-Output "Session valid: $($response.success)"
        Write-Output "Admin email: $($response.admin.email)"
        Write-Output "Admin role: $($response.admin.role)"
    } catch {
        Write-Output "Error validating session: $($_.Exception.Message)"
    }
}

# Test 5: Get audit logs (if session is valid)
if ($sessionToken) {
    Write-Output "`n5. Testing audit logs retrieval..."
    $headers = @{
        "Authorization" = "Bearer $sessionToken"
    }

    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/audit-logs" -Method GET -Headers $headers -ContentType "application/json"
        Write-Output "Audit logs retrieved: $($response.success)"
        Write-Output "Number of logs: $($response.count)"

        if ($response.logs -and $response.logs.Count -gt 0) {
            Write-Output "Recent actions:"
            $response.logs | Select-Object -First 3 | ForEach-Object {
                Write-Output "  - $($_.action) at $($_.created_at) (Success: $($_.success))"
            }
        }
    } catch {
        Write-Output "Error retrieving audit logs: $($_.Exception.Message)"
    }
}

# Test 6: Admin logout (if session is valid)
if ($sessionToken) {
    Write-Output "`n6. Testing admin logout..."
    $logoutData = @{
        session_token = $sessionToken
    }

    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/logout" -Method POST -Body ($logoutData | ConvertTo-Json) -ContentType "application/json"
        Write-Output "Logout successful: $($response.success)"
    } catch {
        Write-Output "Error logging out: $($_.Exception.Message)"
    }
}

Write-Output "`n=== Admin Authentication Test Complete ==="
