# Database Performance Testing Script
# Tests query performance before and after indexing optimization

param(
    [string]$DatabasePath = "/var/db/secure-email.db",
    [int]$TestIterations = 100,
    [int]$ConcurrentUsers = 50,
    [switch]$RunStressTest = $false
)

Write-Host "🔍 Database Performance Testing Script" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan

# Function to measure query execution time
function Measure-QueryPerformance {
    param(
        [string]$Query,
        [string]$Description,
        [int]$Iterations = 10
    )
    
    $totalTime = 0
    $successCount = 0
    
    for ($i = 0; $i -lt $Iterations; $i++) {
        $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
        
        try {
            $result = sqlite3 $DatabasePath $Query 2>$null
            $stopwatch.Stop()
            
            if ($LASTEXITCODE -eq 0) {
                $totalTime += $stopwatch.ElapsedMilliseconds
                $successCount++
            }
        }
        catch {
            Write-Host "Query failed: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    
    if ($successCount -gt 0) {
        $avgTime = $totalTime / $successCount
        Write-Host "✓ $Description`: $avgTime ms average ($successCount/$Iterations successful)" -ForegroundColor Green
        return $avgTime
    } else {
        Write-Host "✗ $Description`: All queries failed" -ForegroundColor Red
        return 0
    }
}

# Function to run EXPLAIN QUERY PLAN analysis
function Test-QueryPlan {
    param(
        [string]$Query,
        [string]$Description
    )
    
    Write-Host "`n📊 Query Plan Analysis: $Description" -ForegroundColor Yellow
    Write-Host "Query: $Query" -ForegroundColor Gray
    
    $explainQuery = "EXPLAIN QUERY PLAN $Query"
    $result = sqlite3 $DatabasePath $explainQuery 2>$null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Query Plan:" -ForegroundColor Cyan
        $result | ForEach-Object { Write-Host "  $_" -ForegroundColor White }
    } else {
        Write-Host "Failed to get query plan" -ForegroundColor Red
    }
}

# Function to check database statistics
function Get-DatabaseStats {
    Write-Host "`n📈 Database Statistics" -ForegroundColor Yellow
    
    $queries = @{
        "Total Users" = "SELECT COUNT(*) FROM users;"
        "Total Emails" = "SELECT COUNT(*) FROM emails;"
        "Total Audit Logs" = "SELECT COUNT(*) FROM audit_log;"
        "Total Sessions" = "SELECT COUNT(*) FROM sessions;"
        "Active Sessions" = "SELECT COUNT(*) FROM sessions WHERE expires_at > datetime('now');"
        "Recent Emails (24h)" = "SELECT COUNT(*) FROM emails WHERE created_at > datetime('now', '-1 day');"
        "Recent Audit Logs (24h)" = "SELECT COUNT(*) FROM audit_log WHERE timestamp > datetime('now', '-1 day');"
    }
    
    foreach ($stat in $queries.GetEnumerator()) {
        $result = sqlite3 $DatabasePath $stat.Value 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  $($stat.Key): $result" -ForegroundColor White
        }
    }
}

# Function to check index status
function Get-IndexStatus {
    Write-Host "`n🔍 Index Status Check" -ForegroundColor Yellow
    
    $indexQuery = @"
SELECT 
    name as index_name,
    tbl_name as table_name,
    sql as definition
FROM sqlite_master 
WHERE type = 'index' 
ORDER BY tbl_name, name;
"@
    
    $result = sqlite3 $DatabasePath $indexQuery 2>$null
    
    if ($LASTEXITCODE -eq 0) {
        $indexes = $result -split "`n" | Where-Object { $_ -match '\|' }
        Write-Host "Found $($indexes.Count) indexes:" -ForegroundColor Green
        
        foreach ($index in $indexes) {
            $parts = $index -split '\|'
            if ($parts.Length -ge 2) {
                Write-Host "  $($parts[1]) on $($parts[2])" -ForegroundColor White
            }
        }
    } else {
        Write-Host "Failed to get index information" -ForegroundColor Red
    }
}

# Function to run performance benchmarks
function Test-PerformanceBenchmarks {
    Write-Host "`n⚡ Performance Benchmarks" -ForegroundColor Yellow
    
    $benchmarks = @{
        "User Authentication Lookup" = "SELECT * FROM users WHERE email = 'test@example.com' LIMIT 1;"
        "Email Listing by Sender" = "SELECT * FROM emails WHERE sender_id = 'test-user-id' ORDER BY created_at DESC LIMIT 50;"
        "Recent Audit Logs" = "SELECT * FROM audit_log WHERE user_id = 'test-user-id' AND timestamp > datetime('now', '-7 days') ORDER BY timestamp DESC LIMIT 100;"
        "Session Validation" = "SELECT * FROM sessions WHERE token_hash = 'test-hash' AND expires_at > datetime('now') LIMIT 1;"
        "Email Count by Status" = "SELECT status, COUNT(*) FROM emails GROUP BY status;"
        "User Email Count" = "SELECT sender_id, COUNT(*) FROM emails GROUP BY sender_id ORDER BY COUNT(*) DESC LIMIT 10;"
        "Recent Security Events" = "SELECT * FROM security_events WHERE severity IN ('high', 'critical') AND timestamp > datetime('now', '-24 hours') ORDER BY timestamp DESC LIMIT 50;"
    }
    
    $results = @{}
    
    foreach ($benchmark in $benchmarks.GetEnumerator()) {
        $avgTime = Measure-QueryPerformance -Query $benchmark.Value -Description $benchmark.Key -Iterations $TestIterations
        $results[$benchmark.Key] = $avgTime
    }
    
    return $results
}

# Function to run stress test
function Test-StressTest {
    param([int]$ConcurrentUsers = 50)
    
    Write-Host "`n🔥 Stress Test - $ConcurrentUsers Concurrent Users" -ForegroundColor Yellow
    
    $queries = @(
        "SELECT * FROM users WHERE email = 'test@example.com' LIMIT 1;",
        "SELECT * FROM emails WHERE sender_id = 'test-user-id' ORDER BY created_at DESC LIMIT 10;",
        "SELECT COUNT(*) FROM audit_log WHERE timestamp > datetime('now', '-1 hour');",
        "SELECT * FROM sessions WHERE expires_at > datetime('now') LIMIT 1;"
    )
    
    $stopwatch = [System.Diagnostics.Stopwatch]::StartNew()
    $successCount = 0
    $totalQueries = $ConcurrentUsers * $queries.Length
    
    # Simulate concurrent users
    $jobs = @()
    
    for ($user = 0; $user -lt $ConcurrentUsers; $user++) {
        foreach ($query in $queries) {
            $job = Start-Job -ScriptBlock {
                param($dbPath, $sql)
                $result = sqlite3 $dbPath $sql 2>$null
                return @{
                    Success = ($LASTEXITCODE -eq 0)
                    Time = Get-Date
                }
            } -ArgumentList $DatabasePath, $query
            
            $jobs += $job
        }
    }
    
    # Wait for all jobs to complete
    $results = $jobs | Wait-Job | Receive-Job
    $stopwatch.Stop()
    
    $successCount = ($results | Where-Object { $_.Success }).Count
    $totalTime = $stopwatch.ElapsedMilliseconds
    $qps = [math]::Round($totalQueries / ($totalTime / 1000), 2)
    
    Write-Host "Stress Test Results:" -ForegroundColor Green
    Write-Host "  Total Queries: $totalQueries" -ForegroundColor White
    Write-Host "  Successful: $successCount" -ForegroundColor White
    Write-Host "  Failed: $($totalQueries - $successCount)" -ForegroundColor White
    Write-Host "  Total Time: $totalTime ms" -ForegroundColor White
    Write-Host "  Queries per Second: $qps" -ForegroundColor White
    Write-Host "  Success Rate: $([math]::Round(($successCount / $totalQueries) * 100, 2))%" -ForegroundColor White
    
    # Clean up jobs
    $jobs | Remove-Job -Force
    
    return @{
        TotalQueries = $totalQueries
        Successful = $successCount
        Failed = $totalQueries - $successCount
        TotalTime = $totalTime
        QPS = $qps
        SuccessRate = [math]::Round(($successCount / $totalQueries) * 100, 2)
    }
}

# Function to generate performance report
function Write-PerformanceReport {
    param(
        [hashtable]$BeforeResults,
        [hashtable]$AfterResults
    )
    
    Write-Host "`n📊 Performance Comparison Report" -ForegroundColor Cyan
    Write-Host "=================================" -ForegroundColor Cyan
    
    $report = @()
    
    foreach ($test in $BeforeResults.Keys) {
        $before = $BeforeResults[$test]
        $after = $AfterResults[$test]
        
        if ($before -gt 0 -and $after -gt 0) {
            $improvement = [math]::Round((($before - $after) / $before) * 100, 2)
            $report += [PSCustomObject]@{
                Test = $test
                Before = "$before ms"
                After = "$after ms"
                Improvement = "$improvement%"
                Status = if ($improvement -gt 0) { "✅ Improved" } else { "❌ Degraded" }
            }
        }
    }
    
    $report | Format-Table -AutoSize
    
    # Summary
    $improvedTests = ($report | Where-Object { $_.Status -eq "✅ Improved" }).Count
    $totalTests = $report.Count
    
    Write-Host "`nSummary:" -ForegroundColor Yellow
    Write-Host "  Improved Tests: $improvedTests/$totalTests" -ForegroundColor Green
    Write-Host "  Average Improvement: $([math]::Round(($report | Where-Object { $_.Status -eq "✅ Improved" } | ForEach-Object { [double]($_.Improvement -replace '%', '') } | Measure-Object -Average).Average, 2))%" -ForegroundColor Green
}

# Main execution
try {
    # Check if sqlite3 is available
    $sqliteVersion = sqlite3 --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ sqlite3 command not found. Please install SQLite3." -ForegroundColor Red
        exit 1
    }
    
    Write-Host "SQLite3 Version: $sqliteVersion" -ForegroundColor Green
    
    # Check if database exists
    if (-not (Test-Path $DatabasePath)) {
        Write-Host "❌ Database not found at: $DatabasePath" -ForegroundColor Red
        exit 1
    }
    
    Write-Host "Database: $DatabasePath" -ForegroundColor Green
    
    # Get database statistics
    Get-DatabaseStats
    
    # Check index status
    Get-IndexStatus
    
    # Run query plan analysis for critical queries
    Test-QueryPlan -Query "SELECT * FROM users WHERE email = 'test@example.com'" -Description "User Authentication"
    Test-QueryPlan -Query "SELECT * FROM emails WHERE sender_id = 'test-user-id' ORDER BY created_at DESC" -Description "Email Listing"
    Test-QueryPlan -Query "SELECT * FROM audit_log WHERE user_id = 'test-user-id' AND timestamp > datetime('now', '-7 days')" -Description "Audit Log Analysis"
    
    # Run performance benchmarks
    Write-Host "`n🚀 Running Performance Benchmarks..." -ForegroundColor Yellow
    $benchmarkResults = Test-PerformanceBenchmarks
    
    # Run stress test if requested
    if ($RunStressTest) {
        Write-Host "`n🔥 Running Stress Test..." -ForegroundColor Yellow
        $stressResults = Test-StressTest -ConcurrentUsers $ConcurrentUsers
        
        Write-Host "`n📈 Stress Test Summary:" -ForegroundColor Cyan
        Write-Host "  Target QPS: 500" -ForegroundColor White
        Write-Host "  Achieved QPS: $($stressResults.QPS)" -ForegroundColor $(if ($stressResults.QPS -ge 500) { "Green" } else { "Red" })
        Write-Host "  Success Rate: $($stressResults.SuccessRate)%" -ForegroundColor $(if ($stressResults.SuccessRate -ge 95) { "Green" } else { "Yellow" })
    }
    
    # Performance recommendations
    Write-Host "`n💡 Performance Recommendations:" -ForegroundColor Cyan
    
    $slowQueries = $benchmarkResults.GetEnumerator() | Where-Object { $_.Value -gt 50 }
    if ($slowQueries) {
        Write-Host "  ⚠️  Slow queries detected:" -ForegroundColor Yellow
        foreach ($query in $slowQueries) {
            Write-Host "    - $($query.Key): $($query.Value) ms" -ForegroundColor Red
        }
        Write-Host "  💡 Consider adding indexes for these queries" -ForegroundColor Yellow
    } else {
        Write-Host "  ✅ All queries performing well (< 50ms)" -ForegroundColor Green
    }
    
    Write-Host "`n✅ Performance testing completed!" -ForegroundColor Green
    
} catch {
    Write-Host "❌ Error during performance testing: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
