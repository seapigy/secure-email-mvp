$body = @{
    email = "test@securesystem.email"
    password = "TestPassword123!"
    fallback_email = "fallback@example.com"
} | ConvertTo-Json

Write-Output "Request body: $body"

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/auth/signup" -Method POST -Headers @{"Content-Type"="application/json"} -Body $body
    Write-Output "Success: $($response | ConvertTo-Json)"
} catch {
    Write-Output "Error: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Output "Response: $errorBody"
    }
}












