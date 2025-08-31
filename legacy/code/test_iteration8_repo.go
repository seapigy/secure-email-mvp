package main

import (
	"database/sql"
	"fmt"
	"log"
	"secure-email-mvp/pkg/securelinks/watermarking"

	_ "modernc.org/sqlite"
)

func main() {
	fmt.Println("=== Watermarking Repository Isolation Test ===")

	// Connect to database
	db, err := sql.Open("sqlite", "data/secure_email.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	fmt.Println("✅ Database connection successful")

	// Create repository
	repo := watermarking.NewSQLiteWatermarkRepository(db)
	fmt.Println("✅ Repository created successfully")

	// Test ListTemplates with no filters
	fmt.Println("\n--- Testing ListTemplates (no filters) ---")
	templates, err := repo.ListTemplates("", "")
	if err != nil {
		log.Printf("❌ ListTemplates failed: %v", err)
	} else {
		fmt.Printf("✅ ListTemplates returned %d templates\n", len(templates))
		for i, template := range templates {
			fmt.Printf("  Template %d: %s (%s)\n", i+1, template.TemplateName, template.TemplateID)
		}
	}

	// Test ListTemplates with watermark type filter
	fmt.Println("\n--- Testing ListTemplates (watermark_type=text) ---")
	templates, err = repo.ListTemplates("text", "")
	if err != nil {
		log.Printf("❌ ListTemplates with watermark_type filter failed: %v", err)
	} else {
		fmt.Printf("✅ ListTemplates with watermark_type filter returned %d templates\n", len(templates))
	}

	// Test ListTemplates with content type filter
	fmt.Println("\n--- Testing ListTemplates (content_type=pdf) ---")
	templates, err = repo.ListTemplates("", "pdf")
	if err != nil {
		log.Printf("❌ ListTemplates with content_type filter failed: %v", err)
	} else {
		fmt.Printf("✅ ListTemplates with content_type filter returned %d templates\n", len(templates))
	}

	// Test ListTemplates with both filters
	fmt.Println("\n--- Testing ListTemplates (watermark_type=text, content_type=pdf) ---")
	templates, err = repo.ListTemplates("text", "pdf")
	if err != nil {
		log.Printf("❌ ListTemplates with both filters failed: %v", err)
	} else {
		fmt.Printf("✅ ListTemplates with both filters returned %d templates\n", len(templates))
	}

	fmt.Println("\n=== Repository Isolation Test Complete ===")
}
