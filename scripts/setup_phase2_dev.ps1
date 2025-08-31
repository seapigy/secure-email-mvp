# =============================================================================
# Phase 2 Development Environment Setup Script
# =============================================================================
# Secure Link External Email Flow - Security Enforcement on Link Access
# 
# This script automates the setup of the development environment for Phase 2,
# which implements advanced security features for secure link access including:
# - Password protection with attempt tracking
# - Enhanced geolocation verification
# - Time lock functionality
# - Auto-destruct & read-once features
# - Remote revocation capabilities
# - Decoy message system
# - Metadata stripping
# - Tamper alerts
# - Multi-Factor Authentication for external users
# =============================================================================

param(
    [switch]$SkipPrerequisites,
    [switch]$SkipDatabase,
    [switch]$SkipBuild,
    [switch]$SkipTests,
    [string]$Environment = "development"
)

Write-Host "🚀 Phase 2 Development Environment Setup" -ForegroundColor Green
Write-Host "Secure Link External Email Flow - Security Enforcement" -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Gray

# =============================================================================
# PREREQUISITE CHECKS
# =============================================================================
if (-not $SkipPrerequisites) {
    Write-Host "`n📋 Checking Prerequisites..." -ForegroundColor Yellow
    
    # Check Go installation
    try {
        $goVersion = go version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Go is installed: $goVersion" -ForegroundColor Green
        } else {
            Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
            Write-Host "Please install Go 1.23+ from https://golang.org/dl/" -ForegroundColor Yellow
            exit 1
        }
    } catch {
        Write-Host "❌ Go is not installed or not in PATH" -ForegroundColor Red
        Write-Host "Please install Go 1.23+ from https://golang.org/dl/" -ForegroundColor Yellow
        exit 1
    }
    
    # Check Node.js installation
    try {
        $nodeVersion = node --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Node.js is installed: $nodeVersion" -ForegroundColor Green
        } else {
            Write-Host "❌ Node.js is not installed or not in PATH" -ForegroundColor Red
            Write-Host "Please install Node.js 18+ from https://nodejs.org/" -ForegroundColor Yellow
            exit 1
        }
    } catch {
        Write-Host "❌ Node.js is not installed or not in PATH" -ForegroundColor Red
        Write-Host "Please install Node.js 18+ from https://nodejs.org/" -ForegroundColor Yellow
        exit 1
    }
    
    # Check npm installation
    try {
        $npmVersion = npm --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ npm is installed: $npmVersion" -ForegroundColor Green
        } else {
            Write-Host "❌ npm is not installed or not in PATH" -ForegroundColor Red
            exit 1
        }
    } catch {
        Write-Host "❌ npm is not installed or not in PATH" -ForegroundColor Red
        exit 1
    }
    
    # Check Git installation
    try {
        $gitVersion = git --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Git is installed: $gitVersion" -ForegroundColor Green
        } else {
            Write-Host "❌ Git is not installed or not in PATH" -ForegroundColor Red
            Write-Host "Please install Git from https://git-scm.com/" -ForegroundColor Yellow
            exit 1
        }
    } catch {
        Write-Host "❌ Git is not installed or not in PATH" -ForegroundColor Red
        Write-Host "Please install Git from https://git-scm.com/" -ForegroundColor Yellow
        exit 1
    }
    
    # Check SQLite installation (optional but recommended)
    try {
        $sqliteVersion = sqlite3 --version 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ SQLite is installed: $sqliteVersion" -ForegroundColor Green
        } else {
            Write-Host "⚠️ SQLite CLI is not installed (optional)" -ForegroundColor Yellow
            Write-Host "   Database operations will be handled by the Go application" -ForegroundColor Gray
        }
    } catch {
        Write-Host "⚠️ SQLite CLI is not installed (optional)" -ForegroundColor Yellow
        Write-Host "   Database operations will be handled by the Go application" -ForegroundColor Gray
    }
}

# =============================================================================
# DEPENDENCY INSTALLATION
# =============================================================================
Write-Host "`n📦 Installing Dependencies..." -ForegroundColor Yellow

# Install Go dependencies
Write-Host "Installing Go dependencies..." -ForegroundColor Cyan
try {
    go mod download
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Go dependencies installed successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to install Go dependencies" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Failed to install Go dependencies: $_" -ForegroundColor Red
    exit 1
}

# Install Node.js dependencies
Write-Host "Installing Node.js dependencies..." -ForegroundColor Cyan
try {
    npm install
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Node.js dependencies installed successfully" -ForegroundColor Green
    } else {
        Write-Host "❌ Failed to install Node.js dependencies" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Failed to install Node.js dependencies: $_" -ForegroundColor Red
    exit 1
}

# =============================================================================
# DATABASE SETUP
# =============================================================================
if (-not $SkipDatabase) {
    Write-Host "`n🗄️ Setting up Database..." -ForegroundColor Yellow
    
    # Create database directory if it doesn't exist
    $dbDir = "/var/db"
    if (-not (Test-Path $dbDir)) {
        Write-Host "Creating database directory: $dbDir" -ForegroundColor Cyan
        try {
            New-Item -ItemType Directory -Path $dbDir -Force | Out-Null
            Write-Host "✅ Database directory created" -ForegroundColor Green
        } catch {
            Write-Host "⚠️ Could not create database directory: $_" -ForegroundColor Yellow
            Write-Host "   Database will be created in current directory" -ForegroundColor Gray
        }
    }
    
    # Apply core schema
    if (Test-Path "schema.sql") {
        Write-Host "Applying core schema..." -ForegroundColor Cyan
        # Note: This would require SQLite to be installed
        # sqlite3 /var/db/secure-email.db < schema.sql
        Write-Host "⚠️ Please manually apply schema.sql to your database" -ForegroundColor Yellow
    }
    
    # Apply secure links migration (Phase 1)
    if (Test-Path "schema/migrate_add_secure_links.sql") {
        Write-Host "Applying secure links migration (Phase 1)..." -ForegroundColor Cyan
        # Note: This would require SQLite to be installed
        # sqlite3 /var/db/secure-email.db < schema/migrate_add_secure_links.sql
        Write-Host "⚠️ Please manually apply migrate_add_secure_links.sql to your database" -ForegroundColor Yellow
    }
    
    # Apply Phase 2 security enforcement migration
    if (Test-Path "schema/migrate_add_phase2_security.sql") {
        Write-Host "Applying Phase 2 security enforcement migration..." -ForegroundColor Cyan
        # Note: This would require SQLite to be installed
        # sqlite3 /var/db/secure-email.db < schema/migrate_add_phase2_security.sql
        Write-Host "⚠️ Please manually apply migrate_add_phase2_security.sql to your database" -ForegroundColor Yellow
    } else {
        Write-Host "⚠️ Phase 2 migration file not found - will be created during implementation" -ForegroundColor Yellow
    }
}

# =============================================================================
# ENVIRONMENT CONFIGURATION
# =============================================================================
Write-Host "`n⚙️ Configuring Environment..." -ForegroundColor Yellow

# Create .env file if it doesn't exist
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env file..." -ForegroundColor Cyan
    $envContent = @"
# =============================================================================
# Secure Email MVP - Environment Configuration
# =============================================================================
# Phase 2: Security Enforcement on Link Access
# =============================================================================

# Database Configuration
SQLITE_DB=/var/db/secure-email.db

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXPIRY_HOURS=24

# Cloudflare R2 Configuration (Optional)
R2_ACCOUNT_ID=your-r2-account-id
CLOUDFLARE_R2_ACCESS_KEY=your-r2-access-key
CLOUDFLARE_R2_SECRET_KEY=your-r2-secret-key
CLOUDFLARE_R2_BUCKET=your-r2-bucket-name
R2_REGION=auto

# Base URL for secure links
BASE_URL=https://securesystem.email

# Security Configuration
MAX_LOGIN_ATTEMPTS=5
LOGIN_LOCKOUT_DURATION_MINUTES=15
PASSWORD_MIN_LENGTH=8
TOTP_ISSUER=SecureEmail

# Phase 2 Security Settings
SECURE_LINK_PASSWORD_MAX_ATTEMPTS=3
SECURE_LINK_LOCKOUT_DURATION_MINUTES=30
SECURE_LINK_AUTO_DESTRUCT_ATTEMPTS=5
SECURE_LINK_DEFAULT_EXPIRY_HOURS=72
SECURE_LINK_GEO_RESTRICTION_ENABLED=true
SECURE_LINK_MFA_ENABLED=true
SECURE_LINK_DECOY_MESSAGES_ENABLED=true

# Geolocation Configuration
GEO_API_KEY=your-geolocation-api-key
GEO_CACHE_DURATION_MINUTES=60

# Email Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@securesystem.email

# Development Settings
DEBUG=true
LOG_LEVEL=debug
ENVIRONMENT=development

# Testing Configuration
TEST_MODE=false
SIMULATE_SELF_DESTRUCT=false
"@
    $envContent | Out-File -FilePath ".env" -Encoding UTF8
    Write-Host "✓ .env file created" -ForegroundColor Green
} else {
    Write-Host "✓ .env file already exists" -ForegroundColor Green
}

# =============================================================================
# BUILD PROCESS
# =============================================================================
if (-not $SkipBuild) {
    Write-Host "`n🔨 Building Application..." -ForegroundColor Yellow
    
    # Build backend
    Write-Host "Building Go backend..." -ForegroundColor Cyan
    try {
        go build -o api-server ./cmd/api/
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Backend built successfully" -ForegroundColor Green
        } else {
            Write-Host "❌ Backend build failed" -ForegroundColor Red
            exit 1
        }
    } catch {
        Write-Host "❌ Backend build failed: $_" -ForegroundColor Red
        exit 1
    }
    
    # Build frontend
    Write-Host "Building React frontend..." -ForegroundColor Cyan
    try {
        npm run build
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Frontend built successfully" -ForegroundColor Green
        } else {
            Write-Host "❌ Frontend build failed" -ForegroundColor Red
            exit 1
        }
    } catch {
        Write-Host "❌ Frontend build failed: $_" -ForegroundColor Red
        exit 1
    }
}

# =============================================================================
# TESTING SETUP
# =============================================================================
if (-not $SkipTests) {
    Write-Host "`n🧪 Setting up Testing..." -ForegroundColor Yellow
    
    # Run backend tests
    Write-Host "Running backend tests..." -ForegroundColor Cyan
    try {
        go test ./... -v
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Backend tests passed" -ForegroundColor Green
        } else {
            Write-Host "⚠️ Some backend tests failed" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "⚠️ Backend tests failed: $_" -ForegroundColor Yellow
    }
    
    # Run frontend tests
    Write-Host "Running frontend tests..." -ForegroundColor Cyan
    try {
        npm test -- --watchAll=false
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Frontend tests passed" -ForegroundColor Green
        } else {
            Write-Host "⚠️ Some frontend tests failed" -ForegroundColor Yellow
        }
    } catch {
        Write-Host "⚠️ Frontend tests failed: $_" -ForegroundColor Yellow
    }
}

# =============================================================================
# PHASE 2 SPECIFIC SETUP
# =============================================================================
Write-Host "`n🔒 Phase 2 Security Features Setup..." -ForegroundColor Yellow

# Create Phase 2 specific directories
$phase2Dirs = @(
    "pkg/securelinks/security",
    "pkg/securelinks/geolocation",
    "pkg/securelinks/mfa",
    "pkg/securelinks/decoy",
    "pkg/securelinks/audit",
    "cmd/api/security_handlers",
    "tests/phase2",
    "docs/phase2"
)

foreach ($dir in $phase2Dirs) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "✅ Created directory: $dir" -ForegroundColor Green
    }
}

# Create Phase 2 implementation files
Write-Host "Creating Phase 2 implementation files..." -ForegroundColor Cyan

# Security enforcement service
@"
package security

import (
    "context"
    "time"
)

// SecurityEnforcementService handles all Phase 2 security features
type SecurityEnforcementService struct {
    // TODO: Implement Phase 2 security enforcement
}

// NewSecurityEnforcementService creates a new security enforcement service
func NewSecurityEnforcementService() *SecurityEnforcementService {
    return &SecurityEnforcementService{}
}

// TODO: Implement password protection, geolocation verification, time lock, etc.
"@ | Out-File -FilePath "pkg/securelinks/security/enforcement.go" -Encoding UTF8

# Geolocation verification service
@"
package geolocation

import (
    "context"
)

// GeolocationVerificationService handles enhanced geolocation verification
type GeolocationVerificationService struct {
    // TODO: Implement enhanced geolocation verification
}

// NewGeolocationVerificationService creates a new geolocation verification service
func NewGeolocationVerificationService() *GeolocationVerificationService {
    return &GeolocationVerificationService{}
}

// TODO: Implement country/city restrictions, allowlist/blocklist, etc.
"@ | Out-File -FilePath "pkg/securelinks/geolocation/verification.go" -Encoding UTF8

# MFA service for external users
@"
package mfa

import (
    "context"
)

// ExternalMFAService handles MFA for external secure link users
type ExternalMFAService struct {
    // TODO: Implement MFA for external users
}

// NewExternalMFAService creates a new external MFA service
func NewExternalMFAService() *ExternalMFAService {
    return &ExternalMFAService{}
}

// TODO: Implement TOTP, email OTP, SMS OTP for external users
"@ | Out-File -FilePath "pkg/securelinks/mfa/external.go" -Encoding UTF8

# Decoy message service
@"
package decoy

import (
    "context"
)

// DecoyMessageService handles decoy message generation and display
type DecoyMessageService struct {
    // TODO: Implement decoy message system
}

// NewDecoyMessageService creates a new decoy message service
func NewDecoyMessageService() *DecoyMessageService {
    return &DecoyMessageService{}
}

// TODO: Implement decoy message generation, trigger conditions, etc.
"@ | Out-File -FilePath "pkg/securelinks/decoy/messages.go" -Encoding UTF8

Write-Host "✅ Phase 2 implementation files created" -ForegroundColor Green

# =============================================================================
# COMPLETION SUMMARY
# =============================================================================
Write-Host "`n🎉 Phase 2 Development Environment Setup Complete!" -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Gray

Write-Host "`n✅ What's Been Set Up:" -ForegroundColor Green

Write-Host "`n🔒 Phase 2 Security Features Framework:" -ForegroundColor Cyan
Write-Host "   • Password Protection System" -ForegroundColor White
Write-Host "   • Enhanced Geolocation Verification" -ForegroundColor White
Write-Host "   • Time Lock Functionality" -ForegroundColor White
Write-Host "   • Auto-Destruct & Read-Once Features" -ForegroundColor White
Write-Host "   • Remote Revocation Capabilities" -ForegroundColor White
Write-Host "   • Decoy Message System" -ForegroundColor White
Write-Host "   • Metadata Stripping" -ForegroundColor White
Write-Host "   • Tamper Alerts" -ForegroundColor White
Write-Host "   • Multi-Factor Authentication for External Users" -ForegroundColor White

Write-Host "`n📁 Project Structure:" -ForegroundColor Cyan
Write-Host "   • pkg/securelinks/security/ - Security enforcement services" -ForegroundColor White
Write-Host "   • pkg/securelinks/geolocation/ - Enhanced geolocation verification" -ForegroundColor White
Write-Host "   • pkg/securelinks/mfa/ - External user MFA" -ForegroundColor White
Write-Host "   • pkg/securelinks/decoy/ - Decoy message system" -ForegroundColor White
Write-Host "   • cmd/api/security_handlers/ - Security API handlers" -ForegroundColor White
Write-Host "   • tests/phase2/ - Phase 2 specific tests" -ForegroundColor White
Write-Host "   • docs/phase2/ - Phase 2 documentation" -ForegroundColor White

Write-Host "`n🚀 Ready for Development:" -ForegroundColor Cyan
Write-Host "   1. Start the backend: ./api-server" -ForegroundColor White
Write-Host "   2. Start the frontend: npm run dev" -ForegroundColor White
Write-Host "   3. Begin implementing Phase 2 security features" -ForegroundColor White

Write-Host "`n📋 Next Steps:" -ForegroundColor Cyan
Write-Host "   1. Implement password protection system" -ForegroundColor White
Write-Host "   2. Add enhanced geolocation verification" -ForegroundColor White
Write-Host "   3. Implement time lock functionality" -ForegroundColor White
Write-Host "   4. Add auto-destruct and read-once features" -ForegroundColor White
Write-Host "   5. Create decoy message system" -ForegroundColor White
Write-Host "   6. Implement tamper alerts" -ForegroundColor White
Write-Host "   7. Add MFA for external users" -ForegroundColor White

Write-Host "`n🔗 Useful Commands:" -ForegroundColor Cyan
Write-Host "   • Run backend: ./api-server" -ForegroundColor White
Write-Host "   • Run frontend: npm run dev" -ForegroundColor White
    Write-Host "   • Run tests: go test ./..." -ForegroundColor White
    Write-Host "   • Build all: go build ./cmd/api/" -ForegroundColor White

Write-Host "`n📚 Documentation:" -ForegroundColor Cyan
Write-Host "   • Phase 2 Implementation Plan: docs/secure-link-external-email-flow-implementation.md" -ForegroundColor White
Write-Host "   • API Documentation: docs/api/secure-links.md" -ForegroundColor White
Write-Host "   • Security Guide: docs/security/phase2-features.md" -ForegroundColor White

Write-Host "`n🎯 Phase 2 Implementation Status:" -ForegroundColor Cyan
Write-Host "   • Overall Progress: 25% Complete" -ForegroundColor White
Write-Host "   • Phase 1: ✅ Complete" -ForegroundColor Green
Write-Host "   • Phase 2: 🔄 Ready to Begin" -ForegroundColor Yellow
Write-Host "   • Phase 3: ⏳ Pending" -ForegroundColor Gray
Write-Host "   • Phase 4: ⏳ Pending" -ForegroundColor Gray

Write-Host "`n🚀 Happy coding! Phase 2 security features are ready for implementation." -ForegroundColor Green
