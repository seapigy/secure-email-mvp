package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"securemail-backend/auth"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Get database URL from environment or use default
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "file:test.db?cache=shared"
	}

	// Open database connection
	db, err := sql.Open("sqlite3", databaseURL)
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Run migrations
	fmt.Println("🔧 Running database migrations...")
	if err := runMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	fmt.Println("✅ Database migrations completed")

	// Set the global DB variable for auth handlers
	// Note: This is a simple approach for demo purposes
	// In production, you'd use dependency injection
	auth.SetDB(db)

	// Create router
	r := mux.NewRouter()

	// Add CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, req)
		})
	})

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes
	api.HandleFunc("/auth/signup", auth.SignupHandler).Methods("POST")
	api.HandleFunc("/auth/login", auth.LoginHandler).Methods("POST")
	api.HandleFunc("/auth/verify-email", auth.VerifyEmailHandler).Methods("POST")
	api.HandleFunc("/account/recover", auth.RecoveryHandler).Methods("POST")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("🚀 Phase 3 Backend Server starting on port %s\n", port)
	fmt.Printf("📊 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("🔐 API base: http://localhost:%s/api\n", port)
	fmt.Printf("💾 Database: %s\n", databaseURL)

	log.Fatal(http.ListenAndServe(":"+port, r))
}

func runMigrations(db *sql.DB) error {
	// Create migrations table to track applied migrations
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %v", err)
	}

	// Load and apply migrations
	migrationsDir := "db/migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %v", err)
	}

	var migrations []string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	// Sort migrations to ensure correct order
	sort.Strings(migrations)

	for _, migration := range migrations {
		// Check if migration already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", migration).Scan(&count)
		if err != nil {
			return fmt.Errorf("failed to check migration status: %v", err)
		}

		if count > 0 {
			fmt.Printf("⏭️  Migration %s already applied, skipping\n", migration)
			continue
		}

		// Read and apply migration
		content, err := ioutil.ReadFile(filepath.Join(migrationsDir, migration))
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %v", migration, err)
		}

		if _, err := db.Exec(string(content)); err != nil {
			// Skip migrations that fail due to existing tables/columns (already applied)
			if strings.Contains(err.Error(), "already exists") || 
			   strings.Contains(err.Error(), "no such column") ||
			   strings.Contains(err.Error(), "duplicate column") {
				fmt.Printf("⚠️  Migration %s skipped (already applied or conflict): %v\n", migration, err)
				// Still record as applied to avoid retrying
				if _, recordErr := db.Exec("INSERT INTO migrations (name) VALUES (?)", migration); recordErr != nil {
					// Ignore record errors for skipped migrations
				}
				continue
			}
			return fmt.Errorf("failed to apply migration %s: %v", migration, err)
		}

		// Record migration as applied
		if _, err := db.Exec("INSERT INTO migrations (name) VALUES (?)", migration); err != nil {
			return fmt.Errorf("failed to record migration %s: %v", migration, err)
		}

		fmt.Printf("✅ Applied migration: %s\n", migration)
	}

	return nil
}
