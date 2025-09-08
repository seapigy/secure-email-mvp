# SecureMail Backend Auth Setup Guide

This guide provides step-by-step instructions for setting up the Core Account Creation system.

## Prerequisites

- Go 1.21+ installed
- Database (SQLite for dev, PostgreSQL/Oracle for production)
- Git

## Environment Configuration

### Development Setup

1. **Set DATABASE_URL environment variable:**
   ```bash
   # For SQLite (development)
   export DATABASE_URL="sqlite://./dev.db"
   
   # For PostgreSQL (production)
   export DATABASE_URL="postgres://user:password@localhost:5432/securemail"
   
   # For Oracle (production)
   export DATABASE_URL="oracle://user:password@localhost:1521/XE"
   ```

2. **Install Go dependencies:**
   ```bash
   cd backend
   go mod init securemail-backend
   go get github.com/google/uuid
   go get golang.org/x/crypto/argon2
   go get github.com/mattn/go-sqlite3
   go get github.com/lib/pq
   go get github.com/godror/godror
   ```

## Database Setup

### Run Migrations

1. **Create migration runner (if not exists):**
   ```bash
   # Create a simple migration runner
   cat > backend/cmd/migrate.go << 'EOF'
   package main
   
   import (
       "database/sql"
       "fmt"
       "io/ioutil"
       "log"
       "os"
       "path/filepath"
       "sort"
       "strconv"
       "strings"
       
       _ "github.com/mattn/go-sqlite3"
       _ "github.com/lib/pq"
       _ "github.com/godror/godror"
   )
   
   func main() {
       databaseURL := os.Getenv("DATABASE_URL")
       if databaseURL == "" {
           log.Fatal("DATABASE_URL environment variable is required")
       }
       
       // Determine driver
       var driver string
       if strings.HasPrefix(databaseURL, "sqlite://") {
           driver = "sqlite3"
           databaseURL = strings.TrimPrefix(databaseURL, "sqlite://")
       } else if strings.HasPrefix(databaseURL, "postgres://") || strings.HasPrefix(databaseURL, "postgresql://") {
           driver = "postgres"
       } else if strings.HasPrefix(databaseURL, "oracle://") {
           driver = "godror"
       } else {
           log.Fatal("Unsupported database URL scheme")
       }
       
       db, err := sql.Open(driver, databaseURL)
       if err != nil {
           log.Fatalf("Failed to connect to database: %v", err)
       }
       defer db.Close()
       
       // Run migrations
       if err := runMigrations(db); err != nil {
           log.Fatalf("Failed to run migrations: %v", err)
       }
       
       fmt.Println("Migrations completed successfully")
   }
   
   func runMigrations(db *sql.DB) error {
       // Create migrations table
       createTableSQL := `
           CREATE TABLE IF NOT EXISTS migrations (
               version INTEGER PRIMARY KEY,
               name TEXT NOT NULL,
               applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
           )
       `
       if _, err := db.Exec(createTableSQL); err != nil {
           return err
       }
       
       // Load and apply migrations
       migrationsDir := "db/migrations"
       files, err := ioutil.ReadDir(migrationsDir)
       if err != nil {
           return err
       }
       
       var migrations []string
       for _, file := range files {
           if strings.HasSuffix(file.Name(), ".sql") {
               migrations = append(migrations, file.Name())
           }
       }
       
       sort.Strings(migrations)
       
       for _, migration := range migrations {
           // Check if already applied
           var count int
           err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = ?", migration).Scan(&count)
           if err != nil {
               return err
           }
           
           if count > 0 {
               fmt.Printf("Migration %s already applied, skipping\n", migration)
               continue
           }
           
           // Read and apply migration
           content, err := ioutil.ReadFile(filepath.Join(migrationsDir, migration))
           if err != nil {
               return err
           }
           
           if _, err := db.Exec(string(content)); err != nil {
               return fmt.Errorf("failed to apply migration %s: %v", migration, err)
           }
           
           // Record migration as applied
           if _, err := db.Exec("INSERT INTO migrations (name) VALUES (?)", migration); err != nil {
               return err
           }
           
           fmt.Printf("Applied migration: %s\n", migration)
       }
       
       return nil
   }
   EOF
   ```

2. **Run migrations:**
   ```bash
   cd backend
   go run cmd/migrate.go
   ```

## Testing

### Run Tests

1. **Run all tests:**
   ```bash
   cd backend
   go test ./...
   ```

2. **Run specific auth tests:**
   ```bash
   go test ./tests -v
   ```

3. **Run tests with coverage:**
   ```bash
   go test ./... -cover
   ```

### Test Database

Tests use an in-memory SQLite database for speed and isolation. No persistent test database files are created.

## API Endpoints

### Signup
```bash
POST /api/auth/signup
Content-Type: application/json

{
  "username": "testuser",
  "email": "test@example.com",
  "password": "securepassword123",
  "account_type": "free"
}
```

### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "test@example.com",
  "password": "securepassword123"
}
```

## Security Features

- **Argon2id Password Hashing**: Memory-hard password hashing with configurable parameters
- **Session Management**: Opaque tokens with SHA256 hashing for storage
- **Input Validation**: Username format validation and SQL injection prevention
- **Audit Logging**: All authentication events are logged

## Development Workflow

1. **Make changes to auth code**
2. **Run tests to ensure nothing breaks:**
   ```bash
   go test ./...
   ```
3. **Run linting:**
   ```bash
   go vet ./...
   go fmt ./...
   ```
4. **Commit changes:**
   ```bash
   git add .
   git commit -m "feat(auth): add core users/sessions/recovery migrations + signup/login handlers"
   ```
5. **Create pull request for review**

## Production Deployment

1. **Set production DATABASE_URL**
2. **Run migrations against production database**
3. **Deploy application with proper environment variables**
4. **Monitor logs for authentication events**

## Troubleshooting

### Common Issues

1. **Database connection errors:**
   - Verify DATABASE_URL is set correctly
   - Check database server is running
   - Verify credentials and permissions

2. **Migration failures:**
   - Check database permissions
   - Verify migration files are present
   - Check for conflicting schema changes

3. **Test failures:**
   - Ensure all dependencies are installed
   - Check Go version compatibility
   - Verify test database setup

## Security Considerations

- Never commit database credentials to version control
- Use environment variables for all sensitive configuration
- Regularly rotate session tokens and passwords
- Monitor authentication logs for suspicious activity
- Keep dependencies updated for security patches

---

**IMPORTANT**: Always commit your changes and create a pull request for review before merging to main branch.
