$signupData = @{
    email = "test2@securesystem.email"
    password = "testpassword123"
    fallback_email = "test2@example.com"
}

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/signup" -Method POST -Body ($signupData | ConvertTo-Json) -ContentType "application/json"
    Write-Host "Signup successful: $($response | ConvertTo-Json)"
} catch {
    Write-Host "Signup failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error details: $errorBody"
    }
}



















