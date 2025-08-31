package migrations

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// Migrator handles database migrations
type Migrator struct {
	db *sql.DB
}

// NewMigrator creates a new migrator instance
func NewMigrator(db *sql.DB) *Migrator {
	return &Migrator{db: db}
}

// Initialize creates the schema_migrations table if it doesn't exist
func (m *Migrator) Initialize() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		description TEXT NOT NULL,
		applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		checksum TEXT NOT NULL
	);`

	_, err := m.db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	log.Printf("[MIGRATION] Schema migrations table initialized")
	return nil
}

// GetCurrentVersion returns the highest applied migration version
func (m *Migrator) GetCurrentVersion() (int, error) {
	var version int
	err := m.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get current migration version: %w", err)
	}
	return version, nil
}

// LoadMigrations loads all migration files from the migrations directory
func (m *Migrator) LoadMigrations(migrationsDir string) ([]Migration, error) {
	var migrations []Migration

	// Read all files in the migrations directory
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		// Parse filename: version__description.sql
		parts := strings.Split(strings.TrimSuffix(file.Name(), ".sql"), "__")
		if len(parts) != 2 {
			log.Printf("[MIGRATION] Skipping invalid migration file: %s", file.Name())
			continue
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			log.Printf("[MIGRATION] Skipping migration with invalid version: %s", file.Name())
			continue
		}

		description := parts[1]

		// Read migration SQL
		content, err := os.ReadFile(filepath.Join(migrationsDir, file.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration file %s: %w", file.Name(), err)
		}

		migrations = append(migrations, Migration{
			Version:     version,
			Description: description,
			SQL:         string(content),
		})
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// ApplyMigration applies a single migration
func (m *Migrator) ApplyMigration(migration Migration) error {
	// Start transaction
	tx, err := m.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Apply the migration
	_, err = tx.Exec(migration.SQL)
	if err != nil {
		return fmt.Errorf("failed to apply migration %d (%s): %w", migration.Version, migration.Description, err)
	}

	// Record the migration
	checksum := fmt.Sprintf("%d-%s", migration.Version, migration.Description)
	_, err = tx.Exec(
		"INSERT INTO schema_migrations (version, description, checksum) VALUES (?, ?, ?)",
		migration.Version,
		migration.Description,
		checksum,
	)
	if err != nil {
		return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
	}

	log.Printf("[MIGRATION] Applied migration %d: %s", migration.Version, migration.Description)
	return nil
}

// RunMigrations runs all pending migrations
func (m *Migrator) RunMigrations(migrationsDir string) error {
	// Initialize migrations table
	if err := m.Initialize(); err != nil {
		return err
	}

	// Get current version
	currentVersion, err := m.GetCurrentVersion()
	if err != nil {
		return err
	}

	// Load all migrations
	migrations, err := m.LoadMigrations(migrationsDir)
	if err != nil {
		return err
	}

	// Find pending migrations
	var pendingMigrations []Migration
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			pendingMigrations = append(pendingMigrations, migration)
		}
	}

	if len(pendingMigrations) == 0 {
		log.Printf("[MIGRATION] Database is up-to-date (v%d)", currentVersion)
		return nil
	}

	log.Printf("[MIGRATION] Found %d pending migrations (current: v%d)", len(pendingMigrations), currentVersion)

	// Apply pending migrations
	for _, migration := range pendingMigrations {
		if err := m.ApplyMigration(migration); err != nil {
			return fmt.Errorf("migration failed at version %d: %w", migration.Version, err)
		}
	}

	log.Printf("[MIGRATION] All migrations completed successfully (v%d)", pendingMigrations[len(pendingMigrations)-1].Version)
	return nil
}
