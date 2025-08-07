# Database Migration Script for Self-Destruct Feature
# This script runs the migration to add self_destruct_after_attempts field

Write-Host "🔧 Running Database Migration for Self-Destruct Feature" -ForegroundColor Green
Write-Host "=====================================================" -ForegroundColor Green

# Check if database file exists
$dbPath = "secure-email.db"
if (-not (Test-Path $dbPath)) {
    Write-Host "❌ Database file not found: $dbPath" -ForegroundColor Red
    Write-Host "Please ensure the database file exists before running migration." -ForegroundColor Yellow
    exit 1
}

Write-Host "📊 Database file found: $dbPath" -ForegroundColor Green

# Read the migration SQL
$migrationPath = "schema/migrate_add_self_destruct.sql"
if (-not (Test-Path $migrationPath)) {
    Write-Host "❌ Migration file not found: $migrationPath" -ForegroundColor Red
    exit 1
}

Write-Host "📝 Reading migration from: $migrationPath" -ForegroundColor Cyan

# Run the migration using sqlite3
try {
    $migrationSQL = Get-Content $migrationPath -Raw
    
    # Create a temporary file with the migration
    $tempFile = [System.IO.Path]::GetTempFileName()
    $migrationSQL | Out-File -FilePath $tempFile -Encoding UTF8
    
    Write-Host "🔄 Running migration..." -ForegroundColor Yellow
    
    # Execute the migration
    $result = sqlite3 $dbPath ".read $tempFile" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
    } else {
        Write-Host "❌ Migration failed with exit code: $LASTEXITCODE" -ForegroundColor Red
        Write-Host "Error output: $result" -ForegroundColor Red
    }
    
    # Clean up temp file
    Remove-Item $tempFile -Force
    
} catch {
    Write-Host "❌ Error running migration: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🔍 Verifying migration..." -ForegroundColor Cyan

# Verify the migration by checking table structure
try {
    $verifyResult = sqlite3 $dbPath "PRAGMA table_info(emails);" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Table structure verified:" -ForegroundColor Green
        Write-Host $verifyResult -ForegroundColor Gray
    } else {
        Write-Host "❌ Failed to verify table structure" -ForegroundColor Red
    }
    
} catch {
    Write-Host "❌ Error verifying migration: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "🎉 Migration process completed!" -ForegroundColor Green
Write-Host "The self_destruct_after_attempts field has been added to the emails table." -ForegroundColor White
