# Test Admin Authentication System
# This script tests the admin authentication endpoints

$baseUrl = "http://localhost:8080"

Write-Host "=== Testing Admin Authentication System ===" -ForegroundColor Green

# Test 1: Check if admin setup is required
Write-Host "`n1. Checking if admin setup is required..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$baseUrl/admin/check-setup" -Method GET -ContentType "application/json"
    Write-Host "Setup required: $($response.setup_required)" -ForegroundColor Cyan
    Write-Host "Root admin email: $($response.root_admin_email)" -ForegroundColor Cyan
} catch {
    Write-Host "Error checking setup: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Create root admin (if setup is required)
Write-Host "`n2. Attempting to create root admin..." -ForegroundColor Yellow
$setupData = @{
    email = "cpigusch@gmail.com"
    password = "SecureAdminPassword123!"
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/admin/setup" -Method POST -Body ($setupData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Admin created successfully: $($response.success)" -ForegroundColor Green
    Write-Host "Admin ID: $($response.admin_id)" -ForegroundColor Cyan
} catch {
    Write-Host "Error creating admin: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody" -ForegroundColor Red
    }
}

# Test 3: Admin login
Write-Host "`n3. Testing admin login..." -ForegroundColor Yellow
$loginData = @{
    email = "cpigusch@gmail.com"
    password = "SecureAdminPassword123!"
    totp_code = ""  # TOTP not enabled initially
}

try {
    $response = Invoke-RestMethod -Uri "$baseUrl/admin/login" -Method POST -Body ($loginData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Login successful: $($response.success)" -ForegroundColor Green
    Write-Host "Session token: $($response.session_token)" -ForegroundColor Cyan
    Write-Host "Admin role: $($response.admin.role)" -ForegroundColor Cyan
    
    $sessionToken = $response.session_token
} catch {
    Write-Host "Error logging in: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody" -ForegroundColor Red
    }
    $sessionToken = $null
}

# Test 4: Validate session (if login was successful)
if ($sessionToken) {
    Write-Host "`n4. Testing session validation..." -ForegroundColor Yellow
    $headers = @{
        "Authorization" = "Bearer $sessionToken"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/session" -Method GET -Headers $headers -ContentType "application/json"
        Write-Host "Session valid: $($response.success)" -ForegroundColor Green
        Write-Host "Admin email: $($response.admin.email)" -ForegroundColor Cyan
        Write-Host "Admin role: $($response.admin.role)" -ForegroundColor Cyan
    } catch {
        Write-Host "Error validating session: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 5: Get audit logs (if session is valid)
if ($sessionToken) {
    Write-Host "`n5. Testing audit logs retrieval..." -ForegroundColor Yellow
    $headers = @{
        "Authorization" = "Bearer $sessionToken"
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/audit-logs" -Method GET -Headers $headers -ContentType "application/json"
        Write-Host "Audit logs retrieved: $($response.success)" -ForegroundColor Green
        Write-Host "Number of logs: $($response.count)" -ForegroundColor Cyan
        
        if ($response.logs -and $response.logs.Count -gt 0) {
            Write-Host "Recent actions:" -ForegroundColor Cyan
            $response.logs | Select-Object -First 3 | ForEach-Object {
                Write-Host "  - $($_.action) at $($_.created_at) (Success: $($_.success))" -ForegroundColor White
            }
        }
    } catch {
        Write-Host "Error retrieving audit logs: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 6: Admin logout (if session is valid)
if ($sessionToken) {
    Write-Host "`n6. Testing admin logout..." -ForegroundColor Yellow
    $logoutData = @{
        session_token = $sessionToken
    }
    
    try {
        $response = Invoke-RestMethod -Uri "$baseUrl/admin/logout" -Method POST -Body ($logoutData | ConvertTo-Json) -ContentType "application/json"
        Write-Host "Logout successful: $($response.success)" -ForegroundColor Green
    } catch {
        Write-Host "Error logging out: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n=== Admin Authentication Test Complete ===" -ForegroundColor Green
