# PowerShell script to send a test email to the Secure Email API

# Define the API endpoint
$uri = "http://localhost:8080/api/email/send"

# Prepare the JSON payload
$body = @{
  sender_id = "test-sender"
  recipient = "test@example.com"
  subject = "Test Subject"
  body = "This is a test email body."
} | ConvertTo-Json

# Send the POST request and print the response
$response = Invoke-RestMethod -Uri $uri -Method POST -ContentType "application/json" -Body $body

Write-Host "API Response:"
$response | ConvertTo-Json -Depth 5 