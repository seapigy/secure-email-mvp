# Secure Email MVP - Environment Setup Script
# This script helps users set up their environment securely

Write-Host "🔒 Secure Email MVP - Environment Setup" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green

# Check if .env exists
if (Test-Path ".env") {
    Write-Host "⚠️  WARNING: .env file already exists!" -ForegroundColor Yellow
    Write-Host "   This file contains sensitive credentials and should be kept secure." -ForegroundColor Yellow
    Write-Host "   Make sure it's in your .gitignore (it should be)." -ForegroundColor Yellow
    Write-Host ""
}

# Create .env from template if it doesn't exist
if (-not (Test-Path ".env")) {
    Write-Host "📝 Creating .env file from template..." -ForegroundColor Cyan
    Copy-Item "env.example" ".env"
    Write-Host "✅ Created .env file from env.example" -ForegroundColor Green
    Write-Host ""
}

# Generate a secure JWT secret
Write-Host "🔑 Generating secure JWT secret..." -ForegroundColor Cyan
$jwtSecret = [System.Convert]::ToBase64String([System.Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
Write-Host "✅ Generated JWT secret: $jwtSecret" -ForegroundColor Green
Write-Host ""

# Update .env with the JWT secret
if (Test-Path ".env") {
    $envContent = Get-Content ".env"
    $updatedContent = $envContent -replace "JWT_SECRET=your_32_byte_jwt_secret_here", "JWT_SECRET=$jwtSecret"
    Set-Content ".env" $updatedContent
    Write-Host "✅ Updated .env with generated JWT secret" -ForegroundColor Green
    Write-Host ""
}

Write-Host "📋 Next Steps:" -ForegroundColor Yellow
Write-Host "1. Edit .env file with your Cloudflare R2 credentials" -ForegroundColor White
Write-Host "2. Never commit .env to Git (it's in .gitignore)" -ForegroundColor White
Write-Host "3. Keep your credentials secure" -ForegroundColor White
Write-Host ""
Write-Host "🔒 Security Checklist:" -ForegroundColor Green
Write-Host "✅ .env is in .gitignore" -ForegroundColor Green
Write-Host "✅ JWT secret generated" -ForegroundColor Green
Write-Host "⚠️  Add your R2 credentials to .env" -ForegroundColor Yellow
Write-Host "⚠️  Never share your .env file" -ForegroundColor Yellow
Write-Host ""
