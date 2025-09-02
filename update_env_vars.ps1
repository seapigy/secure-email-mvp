# Update Environment Variable References
# This script updates all R2 environment variable references to use the new naming convention

Write-Output "Updating environment variable references..."

# Get all Go files
$goFiles = Get-ChildItem -Recurse -Include "*.go" | Where-Object { $_.FullName -notlike "*vendor*" -and $_.FullName -notlike "*node_modules*" }

# Get all PowerShell files
$psFiles = Get-ChildItem -Recurse -Include "*.ps1" | Where-Object { $_.FullName -notlike "*vendor*" -and $_.FullName -notlike "*node_modules*" }

# Get all markdown files
$mdFiles = Get-ChildItem -Recurse -Include "*.md" | Where-Object { $_.FullName -notlike "*vendor*" -and $_.FullName -notlike "*node_modules*" }

# Combine all files
$allFiles = @($goFiles) + @($psFiles) + @($mdFiles)

$updatedCount = 0

foreach ($file in $allFiles) {
    $content = Get-Content $file.FullName -Raw
    $originalContent = $content
    
    # Update R2 environment variables
    $content = $content -replace 'CLOUDFLARE_R2_ACCESS_KEY', 'CLOUDFLARE_R2_ACCESS_KEY'
    $content = $content -replace 'CLOUDFLARE_R2_SECRET_KEY', 'CLOUDFLARE_R2_SECRET_KEY'
    $content = $content -replace 'CLOUDFLARE_R2_BUCKET', 'CLOUDFLARE_CLOUDFLARE_R2_BUCKET'
    $content = $content -replace 'CLOUDFLARE_R2_ENDPOINT', 'CLOUDFLARE_CLOUDFLARE_R2_ENDPOINT'
    
    # Only write if content changed
    if ($content -ne $originalContent) {
        Set-Content $file.FullName $content -NoNewline
        Write-Output "Updated: $($file.FullName)"
        $updatedCount++
    }
}

Write-Output "Updated $updatedCount files"
Write-Output "Environment variable references have been updated to use the new naming convention:"
Write-Output "  - CLOUDFLARE_R2_ACCESS_KEY -> CLOUDFLARE_R2_ACCESS_KEY"
Write-Output "  - CLOUDFLARE_R2_SECRET_KEY -> CLOUDFLARE_R2_SECRET_KEY"
Write-Output "  - CLOUDFLARE_R2_BUCKET -> CLOUDFLARE_CLOUDFLARE_R2_BUCKET"
Write-Output "  - CLOUDFLARE_R2_ENDPOINT -> CLOUDFLARE_CLOUDFLARE_R2_ENDPOINT"









