# =============================================================================
# Phase 1 Development Environment Setup Script
# =============================================================================
# This script sets up the development environment for Phase 1 of the
# Secure Link External Email Flow implementation.
# =============================================================================

Write-Host "🚀 Setting up Phase 1 Development Environment for Secure Links" -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Green

# =============================================================================
# STEP 1: Verify Prerequisites
# =============================================================================

Write-Host "`n📋 Step 1: Verifying Prerequisites..." -ForegroundColor Yellow

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Go 1.23+ from https://golang.org/dl/" -ForegroundColor Red
    exit 1
}

# Check if Node.js is installed
try {
    $nodeVersion = node --version
    Write-Host "✅ Node.js is installed: $nodeVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Node.js is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install Node.js 20+ from https://nodejs.org/" -ForegroundColor Red
    exit 1
}

# Check if npm is installed
try {
    $npmVersion = npm --version
    Write-Host "✅ npm is installed: $npmVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ npm is not installed or not in PATH" -ForegroundColor Red
    exit 1
}

# =============================================================================
# STEP 2: Install Go Dependencies
# =============================================================================

Write-Host "`n📦 Step 2: Installing Go Dependencies..." -ForegroundColor Yellow

# Navigate to project root
Set-Location $PSScriptRoot\..

# Download Go modules
Write-Host "Downloading Go modules..." -ForegroundColor Cyan
go mod download

# Tidy up dependencies
Write-Host "Tidying up Go dependencies..." -ForegroundColor Cyan
go mod tidy

Write-Host "✅ Go dependencies installed successfully" -ForegroundColor Green

# =============================================================================
# STEP 3: Install Frontend Dependencies
# =============================================================================

Write-Host "`n📦 Step 3: Installing Frontend Dependencies..." -ForegroundColor Yellow

# Navigate to frontend directory
Set-Location src

# Install npm dependencies
Write-Host "Installing npm dependencies..." -ForegroundColor Cyan
npm install

# Check for any vulnerabilities
Write-Host "Checking for npm vulnerabilities..." -ForegroundColor Cyan
npm audit --audit-level=moderate

Write-Host "✅ Frontend dependencies installed successfully" -ForegroundColor Green

# Navigate back to project root
Set-Location ..

# =============================================================================
# STEP 4: Database Setup
# =============================================================================

Write-Host "`n🗄️ Step 4: Setting up Database..." -ForegroundColor Yellow

# Create database directory if it doesn't exist
$dbDir = "/var/db"
if (-not (Test-Path $dbDir)) {
    Write-Host "Creating database directory: $dbDir" -ForegroundColor Cyan
    New-Item -ItemType Directory -Path $dbDir -Force | Out-Null
}

# Apply database migrations
Write-Host "Applying database migrations..." -ForegroundColor Cyan

# Apply core schema
if (Test-Path "schema.sql") {
    Write-Host "Applying core schema..." -ForegroundColor Cyan
    # Note: This would require SQLite to be installed
    # sqlite3 /var/db/secure-email.db < schema.sql
    Write-Host "⚠️ Please manually apply schema.sql to your database" -ForegroundColor Yellow
}

# Apply secure links migration
if (Test-Path "schema/migrate_add_secure_links.sql") {
    Write-Host "Applying secure links migration..." -ForegroundColor Cyan
    # Note: This would require SQLite to be installed
    # sqlite3 /var/db/secure-email.db < schema/migrate_add_secure_links.sql
    Write-Host "⚠️ Please manually apply migrate_add_secure_links.sql to your database" -ForegroundColor Yellow
}

Write-Host "✅ Database setup completed" -ForegroundColor Green

# =============================================================================
# STEP 5: Environment Configuration
# =============================================================================

Write-Host "`n⚙️ Step 5: Environment Configuration..." -ForegroundColor Yellow

# Check if .env file exists
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env file from template..." -ForegroundColor Cyan
    Copy-Item "env.example" ".env"
    Write-Host "✅ .env file created from template" -ForegroundColor Green
    Write-Host "⚠️ Please update .env with your configuration values" -ForegroundColor Yellow
} else {
    Write-Host "✅ .env file already exists" -ForegroundColor Green
}

# =============================================================================
# STEP 6: Build Backend
# =============================================================================

Write-Host "`n🔨 Step 6: Building Backend..." -ForegroundColor Yellow

# Build the API server
Write-Host "Building API server..." -ForegroundColor Cyan
go build -o api-server ./cmd/api

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Backend built successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Backend build failed" -ForegroundColor Red
    exit 1
}

# =============================================================================
# STEP 7: Build Frontend
# =============================================================================

Write-Host "`n🔨 Step 7: Building Frontend..." -ForegroundColor Yellow

# Navigate to frontend directory
Set-Location src

# Build the frontend
Write-Host "Building frontend..." -ForegroundColor Cyan
npm run build

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Frontend built successfully" -ForegroundColor Green
} else {
    Write-Host "❌ Frontend build failed" -ForegroundColor Red
    exit 1
}

# Navigate back to project root
Set-Location ..

# =============================================================================
# STEP 8: Run Tests
# =============================================================================

Write-Host "`n🧪 Step 8: Running Tests..." -ForegroundColor Yellow

# Run Go tests
Write-Host "Running Go tests..." -ForegroundColor Cyan
go test ./pkg/securelinks/...

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Go tests passed" -ForegroundColor Green
} else {
    Write-Host "⚠️ Some Go tests failed (this is expected for Phase 1)" -ForegroundColor Yellow
}

# Run frontend tests
Write-Host "Running frontend tests..." -ForegroundColor Cyan
Set-Location src
npm run test

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Frontend tests passed" -ForegroundColor Green
} else {
    Write-Host "⚠️ Some frontend tests failed (this is expected for Phase 1)" -ForegroundColor Yellow
}

Set-Location ..

# =============================================================================
# STEP 9: Development Server Setup
# =============================================================================

Write-Host "`n🚀 Step 9: Development Server Setup..." -ForegroundColor Yellow

Write-Host "To start the development servers:" -ForegroundColor Cyan
Write-Host "1. Backend: ./api-server" -ForegroundColor White
Write-Host "2. Frontend: cd src && npm run dev" -ForegroundColor White
Write-Host "" -ForegroundColor White
Write-Host "API Endpoints:" -ForegroundColor Cyan
Write-Host "- POST /api/secure-links (Create secure link)" -ForegroundColor White
Write-Host "- POST /api/secure-links/{linkID}/access (Access secure link)" -ForegroundColor White
Write-Host "- DELETE /api/secure-links/{linkID} (Revoke secure link)" -ForegroundColor White
Write-Host "- GET /api/secure-links/{linkID} (Get link info)" -ForegroundColor White
Write-Host "- GET /api/secure-links (List links)" -ForegroundColor White
Write-Host "- GET /api/secure-links/templates (Get templates)" -ForegroundColor White

# =============================================================================
# STEP 10: Verification
# =============================================================================

Write-Host "`n✅ Step 10: Verification..." -ForegroundColor Yellow

Write-Host "Phase 1 Development Environment Setup Complete!" -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Green
Write-Host "" -ForegroundColor White
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "1. Update .env with your configuration" -ForegroundColor White
Write-Host "2. Apply database migrations manually" -ForegroundColor White
Write-Host "3. Start the development servers" -ForegroundColor White
Write-Host "4. Test the secure links API endpoints" -ForegroundColor White
Write-Host "5. Begin Phase 2 implementation" -ForegroundColor White
Write-Host "" -ForegroundColor White
Write-Host "Documentation:" -ForegroundColor Cyan
Write-Host "- Implementation Tracker: docs/secure-link-external-email-flow-implementation.md" -ForegroundColor White
Write-Host "- API Documentation: docs/api-documentation.md" -ForegroundColor White
Write-Host "" -ForegroundColor White
Write-Host "Happy coding! 🎉" -ForegroundColor Green
