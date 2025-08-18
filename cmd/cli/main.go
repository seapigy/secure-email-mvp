package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"secure-email-mvp/pkg/audit"

	_ "modernc.org/sqlite"
)

// CLI tool for admin access log queries (Micro-Iteration 4.23)
func main() {
	// Parse command line flags
	var (
		dbPath    = flag.String("db", "/var/db/secure-email.db", "Path to SQLite database")
		emailID   = flag.String("email", "", "Filter by email ID")
		userID    = flag.String("user", "", "Filter by user ID")
		result    = flag.String("result", "", "Filter by result type (success, failed_password, expired, etc.)")
		hours     = flag.Int("hours", 24, "Time window in hours for failed attempts summary")
		limit     = flag.Int("limit", 50, "Limit number of results")
		offset    = flag.Int("offset", 0, "Offset for pagination")
		showStats = flag.Bool("stats", false, "Show failed attempts statistics")
		help      = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help {
		fmt.Println("Secure Email MVP - Admin Access Log CLI")
		fmt.Println("Usage: cli [options]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  cli -email=abc123                    # Show access logs for specific email")
		fmt.Println("  cli -user=user123                    # Show access logs for specific user")
		fmt.Println("  cli -result=failed_password          # Show failed password attempts")
		fmt.Println("  cli -stats -hours=48                 # Show failed attempts stats for last 48 hours")
		fmt.Println("  cli -limit=100 -offset=50            # Paginated results")
		return
	}

	// Connect to database
	db, err := sql.Open("sqlite", *dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize email access auditor
	auditor := audit.NewEmailAccessAuditor(db, audit.DefaultRateLimitConfig)

	ctx := context.Background()

	// Show failed attempts statistics if requested
	if *showStats {
		fmt.Printf("=== Failed Attempts Statistics (Last %d hours) ===\n", *hours)
		stats, err := auditor.GetFailedAttemptsSummary(ctx, *hours)
		if err != nil {
			log.Fatalf("Failed to get failed attempts summary: %v", err)
		}

		fmt.Printf("Total Failed Attempts: %d\n", stats["total_failed_attempts"])
		fmt.Printf("Unique Emails: %d\n", stats["unique_emails"])
		fmt.Printf("Unique IPs: %d\n", stats["unique_ips"])
		fmt.Printf("Unique Users: %d\n", stats["unique_users"])
		fmt.Printf("Time Window: %d hours\n", stats["time_window_hours"])

		if topIPs, ok := stats["top_ips"].([]map[string]interface{}); ok && len(topIPs) > 0 {
			fmt.Println("\nTop IPs with Failed Attempts:")
			for i, ipData := range topIPs {
				fmt.Printf("  %d. %s: %v attempts\n", i+1, ipData["ip_address"], ipData["attempt_count"])
			}
		}
		return
	}

	// Build filters for access logs query
	filters := make(map[string]string)
	if *emailID != "" {
		filters["email_id"] = *emailID
	}
	if *userID != "" {
		filters["user_id"] = *userID
	}
	if *result != "" {
		filters["result"] = *result
	}

	// Get access logs
	fmt.Printf("=== Access Logs ===\n")
	fmt.Printf("Filters: %+v\n", filters)
	fmt.Printf("Limit: %d, Offset: %d\n\n", *limit, *offset)

	logs, err := auditor.GetAccessLogsForAdmin(ctx, filters, *limit, *offset)
	if err != nil {
		log.Fatalf("Failed to get access logs: %v", err)
	}

	if len(logs) == 0 {
		fmt.Println("No access logs found matching the criteria.")
		return
	}

	// Display access logs
	fmt.Printf("Found %d access logs:\n\n", len(logs))
	for i, log := range logs {
		fmt.Printf("%d. Email: %s\n", i+1, log.EmailID)
		if log.UserID != nil {
			fmt.Printf("   User: %s\n", *log.UserID)
		}
		fmt.Printf("   IP: %s\n", log.IPAddress)
		if log.UserAgent != "" {
			fmt.Printf("   User Agent: %s\n", log.UserAgent)
		}
		fmt.Printf("   Status: %s\n", log.Status)
		fmt.Printf("   Result: %s\n", log.Result)
		fmt.Printf("   Attempt Count: %d\n", log.AttemptCount)
		fmt.Printf("   Timestamp: %s\n", log.CreatedAt.Format(time.RFC3339))
		fmt.Println()
	}

	// Get total count for pagination info
	totalCount, err := auditor.GetAccessLogsCountForAdmin(ctx, filters)
	if err != nil {
		log.Printf("Warning: Failed to get total count: %v", err)
	} else {
		fmt.Printf("Total matching records: %d\n", totalCount)
		if (*offset + *limit) < totalCount {
			fmt.Printf("More results available. Use -offset=%d to see next page.\n", *offset+*limit)
		}
	}
}
