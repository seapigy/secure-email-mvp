# Minimal test script for Secure Email API login lockout

$api = "http://localhost:8080/api/auth/login"
$email = "testuser8@example.com"
$pw = "StrongPassword123!"

# 1. Ensure user exists (signup)
$signup = @{
  email = $email
  password = $pw
  fallback_email = "fallback8@example.com"
} | ConvertTo-Json
Invoke-RestMethod -Uri "http://localhost:8080/api/auth/signup" -Method POST -ContentType "application/json" -Body $signup

# 2. Fail login 5 times
for ($i=1; $i -le 5; $i++) {
  $bad = @{
    email = $email
    password = "WrongPassword!"
  } | ConvertTo-Json
  try {
    Invoke-RestMethod -Uri $api -Method POST -ContentType "application/json" -Body $bad
  } catch {
    Write-Host "Failed login $i $($_.Exception.Response.StatusCode) $($_.ErrorDetails.Message)"
  }
}

# 3. Attempt login after lockout
$good = @{
  email = $email
  password = $pw
} | ConvertTo-Json
try {
  Invoke-RestMethod -Uri $api -Method POST -ContentType "application/json" -Body $good
} catch {
  Write-Host "Login after lockout: $($_.Exception.Response.StatusCode) $($_.ErrorDetails.Message)"
} 