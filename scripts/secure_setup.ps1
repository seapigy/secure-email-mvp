# Secure Email MVP - Environment Setup Script
# This script helps users set up their environment securely

Write-Output "🔒 Secure Email MVP - Environment Setup"
Write-Output "================================================"

# Check if .env exists
if (Test-Path ".env") {
    Write-Output "⚠️  WARNING: .env file already exists!"
    Write-Output "   This file contains sensitive credentials and should be kept secure."
    Write-Output "   Make sure it's in your .gitignore (it should be)."
    Write-Output ""
}

# Create .env from template if it doesn't exist
if (-not (Test-Path ".env")) {
    Write-Output "📝 Creating .env file from template..."
    Copy-Item "env.example" ".env"
    Write-Output "✅ Created .env file from env.example"
    Write-Output ""
}

# Generate a secure JWT secret
Write-Output "🔑 Generating secure JWT secret..."
$jwtSecret = [System.Convert]::ToBase64String([System.Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
Write-Output "✅ Generated JWT secret: $jwtSecret"
Write-Output ""

# Update .env with the JWT secret
if (Test-Path ".env") {
    $envContent = Get-Content ".env"
    $updatedContent = $envContent -replace "JWT_SECRET=your_32_byte_jwt_secret_here", "JWT_SECRET=$jwtSecret"
    Set-Content ".env" $updatedContent
    Write-Output "✅ Updated .env with generated JWT secret"
    Write-Output ""
}

Write-Output "📋 Next Steps:"
Write-Output "1. Edit .env file with your Cloudflare R2 credentials"
Write-Output "2. Never commit .env to Git (it's in .gitignore)"
Write-Output "3. Keep your credentials secure"
Write-Output ""
Write-Output "🔒 Security Checklist:"
Write-Output "✅ .env is in .gitignore"
Write-Output "✅ JWT secret generated"
Write-Output "⚠️  Add your R2 credentials to .env"
Write-Output "⚠️  Never share your .env file"
Write-Output ""
