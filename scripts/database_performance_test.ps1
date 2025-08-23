# Database Performance Testing Script
# Tests query performance before and after indexing optimization

param(
    [string]$DatabasePath = "/var/db/secure-email.db",
    [int]$TestIterations = 100,
    [int]$ConcurrentUsers = 50,
    [switch]$RunStressTest = $false
)

Write-Output "🔍 Database Performance Testing Script"
Write-Output "================================================"

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
            Write-Output "Query failed: $($_.Exception.Message)"
        }
    }

    if ($successCount -gt 0) {
        $avgTime = $totalTime / $successCount
        Write-Output "✓ $Description`: $avgTime ms average ($successCount/$Iterations successful)"
        return $avgTime
    } else {
        Write-Output "✗ $Description`: All queries failed"
        return 0
    }
}

# Function to run EXPLAIN QUERY PLAN analysis
function Test-QueryPlan {
    param(
        [string]$Query,
        [string]$Description
    )

    Write-Output "`n📊 Query Plan Analysis: $Description"
    Write-Output "Query: $Query"

    $explainQuery = "EXPLAIN QUERY PLAN $Query"
    $result = sqlite3 $DatabasePath $explainQuery 2>$null

    if ($LASTEXITCODE -eq 0) {
        Write-Output "Query Plan:"
        $result | ForEach-Object { Write-Output "  $_" }
    } else {
        Write-Output "Failed to get query plan"
    }
}

# Function to check database statistics
function Get-DatabaseStats {
    Write-Output "`n📈 Database Statistics"

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
            Write-Output "  $($stat.Key): $result"
        }
    }
}

# Function to check index status
function Get-IndexStatus {
    Write-Output "`n🔍 Index Status Check"

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
        Write-Output "Found $($indexes.Count) indexes:"

        foreach ($index in $indexes) {
            $parts = $index -split '\|'
            if ($parts.Length -ge 2) {
                Write-Output "  $($parts[1]) on $($parts[2])"
            }
        }
    } else {
        Write-Output "Failed to get index information"
    }
}

# Function to run performance benchmarks
function Test-PerformanceBenchmarks {
    Write-Output "`n⚡ Performance Benchmarks"

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

    Write-Output "`n🔥 Stress Test - $ConcurrentUsers Concurrent Users"

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

    Write-Output "Stress Test Results:"
    Write-Output "  Total Queries: $totalQueries"
    Write-Output "  Successful: $successCount"
    Write-Output "  Failed: $($totalQueries - $successCount)"
    Write-Output "  Total Time: $totalTime ms"
    Write-Output "  Queries per Second: $qps"
    Write-Output "  Success Rate: $([math]::Round(($successCount / $totalQueries) * 100, 2))%"

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

    Write-Output "`n📊 Performance Comparison Report"
    Write-Output "================================="

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

    Write-Output "`nSummary:"
    Write-Output "  Improved Tests: $improvedTests/$totalTests"
    Write-Output "  Average Improvement: $([math]::Round(($report | Where-Object { $_.Status -eq "✅ Improved" } | ForEach-Object { [double]($_.Improvement -replace '%', '') } | Measure-Object -Average).Average, 2))%" -ForegroundColor Green
}

# Main execution
try {
    # Check if sqlite3 is available
    $sqliteVersion = sqlite3 --version 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Output "❌ sqlite3 command not found. Please install SQLite3."
        exit 1
    }

    Write-Output "SQLite3 Version: $sqliteVersion"

    # Check if database exists
    if (-not (Test-Path $DatabasePath)) {
        Write-Output "❌ Database not found at: $DatabasePath"
        exit 1
    }

    Write-Output "Database: $DatabasePath"

    # Get database statistics
    Get-DatabaseStats

    # Check index status
    Get-IndexStatus

    # Run query plan analysis for critical queries
    Test-QueryPlan -Query "SELECT * FROM users WHERE email = 'test@example.com'" -Description "User Authentication"
    Test-QueryPlan -Query "SELECT * FROM emails WHERE sender_id = 'test-user-id' ORDER BY created_at DESC" -Description "Email Listing"
    Test-QueryPlan -Query "SELECT * FROM audit_log WHERE user_id = 'test-user-id' AND timestamp > datetime('now', '-7 days')" -Description "Audit Log Analysis"

    # Run performance benchmarks
    Write-Output "`n🚀 Running Performance Benchmarks..."
    $benchmarkResults = Test-PerformanceBenchmarks

    # Run stress test if requested
    if ($RunStressTest) {
        Write-Output "`n🔥 Running Stress Test..."
        $stressResults = Test-StressTest -ConcurrentUsers $ConcurrentUsers

        Write-Output "`n📈 Stress Test Summary:"
        Write-Output "  Target QPS: 500"
        Write-Output "  Achieved QPS: $($stressResults.QPS)" -ForegroundColor $(if ($stressResults.QPS -ge 500) { "Green" } else { "Red" })
        Write-Output "  Success Rate: $($stressResults.SuccessRate)%" -ForegroundColor $(if ($stressResults.SuccessRate -ge 95) { "Green" } else { "Yellow" })
    }

    # Performance recommendations
    Write-Output "`n💡 Performance Recommendations:"

    $slowQueries = $benchmarkResults.GetEnumerator() | Where-Object { $_.Value -gt 50 }
    if ($slowQueries) {
        Write-Output "  ⚠️  Slow queries detected:"
        foreach ($query in $slowQueries) {
            Write-Output "    - $($query.Key): $($query.Value) ms"
        }
        Write-Output "  💡 Consider adding indexes for these queries"
    } else {
        Write-Output "  ✅ All queries performing well (< 50ms)"
    }

    Write-Output "`n✅ Performance testing completed!"

} catch {
    Write-Output "❌ Error during performance testing: $($_.Exception.Message)"
    exit 1
}
