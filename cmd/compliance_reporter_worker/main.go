package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"secure-email-mvp/pkg/email"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// ComplianceReporterWorker handles automated compliance reporting and certification generation
type ComplianceReporterWorker struct {
	db                       *sql.DB
	complianceService        *email.ComplianceService
	reportingInterval        time.Duration
	enableAutoCertification  bool
	enableViolationDetection bool
	stopChan                 chan struct{}
}

// NewComplianceReporterWorker creates a new compliance reporter worker
func NewComplianceReporterWorker(dbPath string, reportingInterval time.Duration, enableAutoCertification, enableViolationDetection bool) (*ComplianceReporterWorker, error) {
	// Connect to database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize compliance service
	complianceService := email.NewComplianceService(db)

	return &ComplianceReporterWorker{
		db:                       db,
		complianceService:        complianceService,
		reportingInterval:        reportingInterval,
		enableAutoCertification:  enableAutoCertification,
		enableViolationDetection: enableViolationDetection,
		stopChan:                 make(chan struct{}),
	}, nil
}

// Start begins the compliance reporting worker
func (w *ComplianceReporterWorker) Start() {
	log.Printf("Starting Compliance Reporter Worker...")
	log.Printf("Reporting interval: %v", w.reportingInterval)
	log.Printf("Auto certification enabled: %v", w.enableAutoCertification)
	log.Printf("Violation detection enabled: %v", w.enableViolationDetection)

	// Start the main reporting loop
	go w.runReportingLoop()

	// Start violation detection if enabled
	if w.enableViolationDetection {
		go w.runViolationDetection()
	}

	log.Printf("Compliance Reporter Worker started successfully")
}

// Stop gracefully stops the worker
func (w *ComplianceReporterWorker) Stop() {
	log.Printf("Stopping Compliance Reporter Worker...")
	close(w.stopChan)
	w.db.Close()
	log.Printf("Compliance Reporter Worker stopped")
}

// runReportingLoop runs the main reporting loop
func (w *ComplianceReporterWorker) runReportingLoop() {
	ticker := time.NewTicker(w.reportingInterval)
	defer ticker.Stop()

	// Run initial report generation
	w.generateComplianceReports(context.Background())

	for {
		select {
		case <-ticker.C:
			w.generateComplianceReports(context.Background())
		case <-w.stopChan:
			return
		}
	}
}

// runViolationDetection runs continuous violation detection
func (w *ComplianceReporterWorker) runViolationDetection() {
	ticker := time.NewTicker(1 * time.Hour) // Check every hour
	defer ticker.Stop()

	// Run initial violation detection
	w.detectComplianceViolations(context.Background())

	for {
		select {
		case <-ticker.C:
			w.detectComplianceViolations(context.Background())
		case <-w.stopChan:
			return
		}
	}
}

// generateComplianceReports generates compliance reports for all active frameworks
func (w *ComplianceReporterWorker) generateComplianceReports(ctx context.Context) {
	log.Printf("Generating compliance reports...")

	// Get all active compliance frameworks
	frameworks, err := w.complianceService.GetComplianceFrameworks(ctx)
	if err != nil {
		log.Printf("Failed to get compliance frameworks: %v", err)
		return
	}

	for _, framework := range frameworks {
		if !framework.IsActive {
			continue
		}

		// Generate monthly report if it's the first of the month
		if w.shouldGenerateMonthlyReport() {
			err := w.generateMonthlyCertification(ctx, framework)
			if err != nil {
				log.Printf("Failed to generate monthly certification for %s: %v", framework.FrameworkName, err)
			}
		}

		// Generate quarterly report if it's the start of a quarter
		if w.shouldGenerateQuarterlyReport() {
			err := w.generateQuarterlyCertification(ctx, framework)
			if err != nil {
				log.Printf("Failed to generate quarterly certification for %s: %v", framework.FrameworkName, err)
			}
		}

		// Generate annual report if it's the start of the year
		if w.shouldGenerateAnnualReport() {
			err := w.generateAnnualCertification(ctx, framework)
			if err != nil {
				log.Printf("Failed to generate annual certification for %s: %v", framework.FrameworkName, err)
			}
		}
	}

	log.Printf("Compliance report generation completed")
}

// generateMonthlyCertification generates a monthly compliance certification
func (w *ComplianceReporterWorker) generateMonthlyCertification(ctx context.Context, framework email.ComplianceFramework) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)

	log.Printf("Generating monthly certification for %s (period: %s to %s)",
		framework.FrameworkName, periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))

	certification, err := w.complianceService.GenerateComplianceCertification(
		ctx, framework.ID, "monthly", periodStart, periodEnd, "compliance_reporter_worker",
	)
	if err != nil {
		return fmt.Errorf("failed to generate monthly certification: %w", err)
	}

	log.Printf("Generated monthly certification %s with compliance score: %.2f%%",
		certification.CertificationID, certification.ComplianceScore*100)

	return nil
}

// generateQuarterlyCertification generates a quarterly compliance certification
func (w *ComplianceReporterWorker) generateQuarterlyCertification(ctx context.Context, framework email.ComplianceFramework) error {
	now := time.Now()
	quarter := (now.Month()-1)/3 + 1
	periodStart := time.Date(now.Year(), time.Month((quarter-1)*3+1), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 3, 0).Add(-time.Second)

	log.Printf("Generating quarterly certification for %s (Q%d %d, period: %s to %s)",
		framework.FrameworkName, quarter, now.Year(), periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))

	certification, err := w.complianceService.GenerateComplianceCertification(
		ctx, framework.ID, "quarterly", periodStart, periodEnd, "compliance_reporter_worker",
	)
	if err != nil {
		return fmt.Errorf("failed to generate quarterly certification: %w", err)
	}

	log.Printf("Generated quarterly certification %s with compliance score: %.2f%%",
		certification.CertificationID, certification.ComplianceScore*100)

	return nil
}

// generateAnnualCertification generates an annual compliance certification
func (w *ComplianceReporterWorker) generateAnnualCertification(ctx context.Context, framework email.ComplianceFramework) error {
	now := time.Now()
	periodStart := time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(1, 0, 0).Add(-time.Second)

	log.Printf("Generating annual certification for %s (%d, period: %s to %s)",
		framework.FrameworkName, now.Year(), periodStart.Format("2006-01-02"), periodEnd.Format("2006-01-02"))

	certification, err := w.complianceService.GenerateComplianceCertification(
		ctx, framework.ID, "annual", periodStart, periodEnd, "compliance_reporter_worker",
	)
	if err != nil {
		return fmt.Errorf("failed to generate annual certification: %w", err)
	}

	log.Printf("Generated annual certification %s with compliance score: %.2f%%",
		certification.CertificationID, certification.ComplianceScore*100)

	return nil
}

// detectComplianceViolations detects compliance violations across all frameworks
func (w *ComplianceReporterWorker) detectComplianceViolations(ctx context.Context) {
	log.Printf("Detecting compliance violations...")

	// Get all active compliance frameworks
	frameworks, err := w.complianceService.GetComplianceFrameworks(ctx)
	if err != nil {
		log.Printf("Failed to get compliance frameworks: %v", err)
		return
	}

	for _, framework := range frameworks {
		if !framework.IsActive {
			continue
		}

		// Get compliance rules for this framework
		rules, err := w.complianceService.GetComplianceRules(ctx, framework.ID)
		if err != nil {
			log.Printf("Failed to get compliance rules for %s: %v", framework.FrameworkName, err)
			continue
		}

		// Check each rule for violations
		for _, rule := range rules {
			if !rule.AutoEnforcementEnabled {
				continue
			}

			err := w.checkRuleViolations(ctx, framework, rule)
			if err != nil {
				log.Printf("Failed to check violations for rule %s: %v", rule.RuleCode, err)
			}
		}
	}

	log.Printf("Compliance violation detection completed")
}

// checkRuleViolations checks for violations of a specific compliance rule
func (w *ComplianceReporterWorker) checkRuleViolations(ctx context.Context, framework email.ComplianceFramework, rule email.ComplianceRule) error {
	// This is a simplified implementation - in a real system, this would be much more complex
	// and would check actual email data against compliance requirements

	// Check for retention period violations
	if rule.RetentionPeriodDays != nil {
		err := w.checkRetentionViolations(ctx, framework, rule)
		if err != nil {
			return fmt.Errorf("failed to check retention violations: %w", err)
		}
	}

	// Check for encryption violations
	if rule.EncryptionRequired {
		err := w.checkEncryptionViolations(ctx, framework, rule)
		if err != nil {
			return fmt.Errorf("failed to check encryption violations: %w", err)
		}
	}

	// Check for archival violations
	if rule.ArchivalRequired {
		err := w.checkArchivalViolations(ctx, framework, rule)
		if err != nil {
			return fmt.Errorf("failed to check archival violations: %w", err)
		}
	}

	return nil
}

// checkRetentionViolations checks for retention period violations
func (w *ComplianceReporterWorker) checkRetentionViolations(ctx context.Context, framework email.ComplianceFramework, rule email.ComplianceRule) error {
	// Query for emails that exceed the retention period
	query := `
		SELECT COUNT(*) FROM emails 
		WHERE created_at < datetime('now', '-' || ? || ' days') 
		AND encrypted_blob_url IS NOT NULL
	`

	var count int
	err := w.db.QueryRowContext(ctx, query, *rule.RetentionPeriodDays).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count retention violations: %w", err)
	}

	if count > 0 {
		// Create violation record
		violation := &email.ComplianceViolation{
			FrameworkID:          framework.ID,
			ComplianceRuleID:     rule.ID,
			ViolationType:        "retention_exceeded",
			ViolationSeverity:    rule.SeverityLevel,
			ViolationDescription: fmt.Sprintf("%d emails exceed retention period of %d days", count, *rule.RetentionPeriodDays),
			Status:               "open",
			AffectedEmailsCount:  count,
			DaysOverLimit:        int(time.Since(time.Now().AddDate(0, 0, -*rule.RetentionPeriodDays)).Hours() / 24),
		}

		err = w.complianceService.CreateComplianceViolation(ctx, violation)
		if err != nil {
			return fmt.Errorf("failed to create retention violation: %w", err)
		}

		log.Printf("Created retention violation for %s rule %s: %s",
			framework.FrameworkName, rule.RuleCode, violation.ViolationDescription)
	}

	return nil
}

// checkEncryptionViolations checks for encryption violations
func (w *ComplianceReporterWorker) checkEncryptionViolations(ctx context.Context, framework email.ComplianceFramework, rule email.ComplianceRule) error {
	// Query for emails without encryption
	query := `
		SELECT COUNT(*) FROM emails 
		WHERE encrypted_blob_url IS NULL OR encrypted_blob_url = ''
	`

	var count int
	err := w.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count encryption violations: %w", err)
	}

	if count > 0 {
		// Create violation record
		violation := &email.ComplianceViolation{
			FrameworkID:          framework.ID,
			ComplianceRuleID:     rule.ID,
			ViolationType:        "encryption_missing",
			ViolationSeverity:    rule.SeverityLevel,
			ViolationDescription: fmt.Sprintf("%d emails lack required encryption", count),
			Status:               "open",
			AffectedEmailsCount:  count,
		}

		err = w.complianceService.CreateComplianceViolation(ctx, violation)
		if err != nil {
			return fmt.Errorf("failed to create encryption violation: %w", err)
		}

		log.Printf("Created encryption violation for %s rule %s: %s",
			framework.FrameworkName, rule.RuleCode, violation.ViolationDescription)
	}

	return nil
}

// checkArchivalViolations checks for archival violations
func (w *ComplianceReporterWorker) checkArchivalViolations(ctx context.Context, framework email.ComplianceFramework, rule email.ComplianceRule) error {
	// Query for emails that should be archived but aren't
	query := `
		SELECT COUNT(*) FROM emails e
		LEFT JOIN archived_emails ae ON e.email_id = ae.original_email_id
		WHERE ae.original_email_id IS NULL 
		AND e.created_at < datetime('now', '-90 days')
		AND e.encrypted_blob_url IS NOT NULL
	`

	var count int
	err := w.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count archival violations: %w", err)
	}

	if count > 0 {
		// Create violation record
		violation := &email.ComplianceViolation{
			FrameworkID:          framework.ID,
			ComplianceRuleID:     rule.ID,
			ViolationType:        "archival_missing",
			ViolationSeverity:    rule.SeverityLevel,
			ViolationDescription: fmt.Sprintf("%d emails should be archived but are not", count),
			Status:               "open",
			AffectedEmailsCount:  count,
		}

		err = w.complianceService.CreateComplianceViolation(ctx, violation)
		if err != nil {
			return fmt.Errorf("failed to create archival violation: %w", err)
		}

		log.Printf("Created archival violation for %s rule %s: %s",
			framework.FrameworkName, rule.RuleCode, violation.ViolationDescription)
	}

	return nil
}

// shouldGenerateMonthlyReport checks if we should generate a monthly report
func (w *ComplianceReporterWorker) shouldGenerateMonthlyReport() bool {
	now := time.Now()
	return now.Day() == 1 && now.Hour() == 0 && now.Minute() < 10
}

// shouldGenerateQuarterlyReport checks if we should generate a quarterly report
func (w *ComplianceReporterWorker) shouldGenerateQuarterlyReport() bool {
	now := time.Now()
	quarter := (now.Month()-1)/3 + 1
	return now.Day() == 1 && now.Month() == time.Month((quarter-1)*3+1) && now.Hour() == 0 && now.Minute() < 10
}

// shouldGenerateAnnualReport checks if we should generate an annual report
func (w *ComplianceReporterWorker) shouldGenerateAnnualReport() bool {
	now := time.Now()
	return now.Day() == 1 && now.Month() == 1 && now.Hour() == 0 && now.Minute() < 10
}

func main() {
	// Parse command line flags
	var (
		dbPath                   = flag.String("db", "/var/db/secure-email.db", "Database path")
		reportingInterval        = flag.Duration("interval", 24*time.Hour, "Reporting interval")
		enableAutoCertification  = flag.Bool("auto-cert", true, "Enable automatic certification generation")
		enableViolationDetection = flag.Bool("violation-detection", true, "Enable violation detection")
		envFile                  = flag.String("env", ".env", "Environment file path")
	)
	flag.Parse()

	// Load environment variables
	if err := godotenv.Load(*envFile); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Create and start the worker
	worker, err := NewComplianceReporterWorker(*dbPath, *reportingInterval, *enableAutoCertification, *enableViolationDetection)
	if err != nil {
		log.Fatalf("Failed to create compliance reporter worker: %v", err)
	}

	// Start the worker
	worker.Start()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Wait for shutdown signal
	<-sigChan
	log.Printf("Received shutdown signal, stopping worker...")
	worker.Stop()
}





