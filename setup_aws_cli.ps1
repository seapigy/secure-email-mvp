#!/usr/bin/env pwsh
<#
.SYNOPSIS
    AWS CLI Setup Script for Secure Email MVP
.DESCRIPTION
    This script helps configure AWS CLI for the Secure Email MVP project.
    It sets up credentials, configures the region, and verifies the installation.
.PARAMETER AccessKeyId
    AWS Access Key ID
.PARAMETER SecretAccessKey
    AWS Secret Access Key
.PARAMETER Region
    AWS Region (default: us-east-1)
.EXAMPLE
    .\setup_aws_cli.ps1 -AccessKeyId "AKIA..." -SecretAccessKey "secret..."
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$AccessKeyId,
    
    [Parameter(Mandatory=$true)]
    [string]$SecretAccessKey,
    
    [Parameter(Mandatory=$false)]
    [string]$Region = "us-east-1"
)

Write-Host "🚀 AWS CLI Setup for Secure Email MVP" -ForegroundColor Green
Write-Host "================================================" -ForegroundColor Green

# Step 1: Verify AWS CLI installation
Write-Host "`n1. Verifying AWS CLI installation..." -ForegroundColor Yellow
try {
    $awsVersion = aws --version 2>$null
    if ($awsVersion) {
        Write-Host "   ✅ AWS CLI is installed: $awsVersion" -ForegroundColor Green
    } else {
        Write-Host "   ❌ AWS CLI not found in PATH" -ForegroundColor Red
        Write-Host "   Please install AWS CLI v2 first:" -ForegroundColor Yellow
        Write-Host "   winget install Amazon.AWSCLI" -ForegroundColor Cyan
        exit 1
    }
} catch {
    Write-Host "   ❌ AWS CLI not found in PATH" -ForegroundColor Red
    Write-Host "   Please install AWS CLI v2 first:" -ForegroundColor Yellow
    Write-Host "   winget install Amazon.AWSCLI" -ForegroundColor Cyan
    exit 1
}

# Step 2: Configure AWS CLI
Write-Host "`n2. Configuring AWS CLI..." -ForegroundColor Yellow

# Create AWS directory if it doesn't exist
$awsDir = "$env:USERPROFILE\.aws"
if (!(Test-Path $awsDir)) {
    New-Item -ItemType Directory -Path $awsDir -Force | Out-Null
    Write-Host "   📁 Created AWS directory: $awsDir" -ForegroundColor Cyan
}

# Configure AWS CLI
Write-Host "   🔧 Running aws configure..." -ForegroundColor Cyan
$configInput = @"
$AccessKeyId
$SecretAccessKey
$Region
json
"@

$configInput | aws configure

if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ AWS CLI configured successfully" -ForegroundColor Green
} else {
    Write-Host "   ❌ Failed to configure AWS CLI" -ForegroundColor Red
    exit 1
}

# Step 3: Verify configuration files
Write-Host "`n3. Verifying configuration files..." -ForegroundColor Yellow

$credentialsFile = "$awsDir\credentials"
$configFile = "$awsDir\config"

if (Test-Path $credentialsFile) {
    Write-Host "   ✅ Credentials file created: $credentialsFile" -ForegroundColor Green
} else {
    Write-Host "   ❌ Credentials file not found" -ForegroundColor Red
}

if (Test-Path $configFile) {
    Write-Host "   ✅ Config file created: $configFile" -ForegroundColor Green
} else {
    Write-Host "   ❌ Config file not found" -ForegroundColor Red
}

# Step 4: Test authentication
Write-Host "`n4. Testing AWS authentication..." -ForegroundColor Yellow
try {
    $callerIdentity = aws sts get-caller-identity 2>$null | ConvertFrom-Json
    if ($callerIdentity) {
        Write-Host "   ✅ Authentication successful!" -ForegroundColor Green
        Write-Host "   📋 Account ID: $($callerIdentity.Account)" -ForegroundColor Cyan
        Write-Host "   👤 User ID: $($callerIdentity.UserId)" -ForegroundColor Cyan
        Write-Host "   🏷️  ARN: $($callerIdentity.Arn)" -ForegroundColor Cyan
    } else {
        Write-Host "   ❌ Authentication failed" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "   ❌ Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 5: Test SES access
Write-Host "`n5. Testing SES access..." -ForegroundColor Yellow
try {
    $sesQuota = aws ses get-send-quota --region $Region 2>$null | ConvertFrom-Json
    if ($sesQuota) {
        Write-Host "   ✅ SES access successful!" -ForegroundColor Green
        Write-Host "   📧 Daily quota: $($sesQuota.Max24HourSend) emails" -ForegroundColor Cyan
        Write-Host "   📊 Sent today: $($sesQuota.SentLast24Hours) emails" -ForegroundColor Cyan
    } else {
        Write-Host "   ❌ SES access failed" -ForegroundColor Red
    }
} catch {
    Write-Host "   ⚠️  SES access failed: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host "   This might be normal if SES is not set up yet" -ForegroundColor Yellow
}

# Step 6: Security reminder
Write-Host "`n6. Security check..." -ForegroundColor Yellow

# Check if .gitignore includes AWS credentials
$gitignorePath = ".gitignore"
if (Test-Path $gitignorePath) {
    $gitignoreContent = Get-Content $gitignorePath -Raw
    if ($gitignoreContent -match "\.aws/") {
        Write-Host "   ✅ .gitignore includes AWS credentials protection" -ForegroundColor Green
    } else {
        Write-Host "   ⚠️  .gitignore should include .aws/ directory" -ForegroundColor Yellow
    }
} else {
    Write-Host "   ⚠️  .gitignore file not found" -ForegroundColor Yellow
}

Write-Host "`n🎉 AWS CLI setup completed successfully!" -ForegroundColor Green
Write-Host "`n📝 Next steps:" -ForegroundColor Yellow
Write-Host "   1. Test SES email sending: python ses_test.py" -ForegroundColor Cyan
Write-Host "   2. Verify domain in SES console: https://console.aws.amazon.com/ses/" -ForegroundColor Cyan
Write-Host "   3. Check email authentication headers in Gmail" -ForegroundColor Cyan

Write-Host "`n🔒 Security Notes:" -ForegroundColor Yellow
Write-Host "   - AWS credentials are stored in: $awsDir" -ForegroundColor Cyan
Write-Host "   - Never commit credentials to git" -ForegroundColor Red
Write-Host "   - Use IAM roles in production environments" -ForegroundColor Cyan
