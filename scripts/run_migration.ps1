# Database Migration Script for Self-Destruct Feature
# This script runs the migration to add self_destruct_after_attempts field

Write-Output "🔧 Running Database Migration for Self-Destruct Feature"
Write-Output "====================================================="

# Check if database file exists
$dbPath = "secure-email.db"
if (-not (Test-Path $dbPath)) {
    Write-Output "❌ Database file not found: $dbPath"
    Write-Output "Please ensure the database file exists before running migration."
    exit 1
}

Write-Output "📊 Database file found: $dbPath"

# Read the migration SQL
$migrationPath = "schema/migrate_add_self_destruct.sql"
if (-not (Test-Path $migrationPath)) {
    Write-Output "❌ Migration file not found: $migrationPath"
    exit 1
}

Write-Output "📝 Reading migration from: $migrationPath"

# Run the migration using sqlite3
try {
    $migrationSQL = Get-Content $migrationPath -Raw

    # Create a temporary file with the migration
    $tempFile = [System.IO.Path]::GetTempFileName()
    $migrationSQL | Out-File -FilePath $tempFile -Encoding UTF8

    Write-Output "🔄 Running migration..."

    # Execute the migration
    $result = sqlite3 $dbPath ".read $tempFile" 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Output "✅ Migration completed successfully!"
    } else {
        Write-Output "❌ Migration failed with exit code: $LASTEXITCODE"
        Write-Output "Error output: $result"
    }

    # Clean up temp file
    Remove-Item $tempFile -Force

} catch {
    Write-Output "❌ Error running migration: $($_.Exception.Message)"
    exit 1
}

Write-Output ""
Write-Output "🔍 Verifying migration..."

# Verify the migration by checking table structure
try {
    $verifyResult = sqlite3 $dbPath "PRAGMA table_info(emails);" 2>&1

    if ($LASTEXITCODE -eq 0) {
        Write-Output "✅ Table structure verified:"
        Write-Host $verifyResult -ForegroundColor Gray
    } else {
        Write-Output "❌ Failed to verify table structure"
    }

} catch {
    Write-Output "❌ Error verifying migration: $($_.Exception.Message)"
}

Write-Output ""
Write-Output "🎉 Migration process completed!"
Write-Output "The self_destruct_after_attempts field has been added to the emails table."
