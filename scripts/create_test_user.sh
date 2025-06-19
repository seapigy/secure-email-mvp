#!/bin/bash
set -e

# create_test_user.sh: Create a test user for development
# This script creates a user in the database with a known password and TOTP secret

if [ "$#" -ne 2 ]; then
    echo "Usage: $0 <email> <password>"
    echo "Example: $0 test@securesystem.email testpass123"
    exit 1
fi

EMAIL="$1"
PASSWORD="$2"

# Check if email is valid format
if [[ ! "$EMAIL" =~ ^[^@]+@securesystem\.email$ ]]; then
    echo "Error: Email must be in format user@securesystem.email"
    exit 1
fi

# Check if password is valid length
if [ ${#PASSWORD} -lt 8 ] || [ ${#PASSWORD} -gt 128 ]; then
    echo "Error: Password must be 8-128 characters long"
    exit 1
fi

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Set default database path if not set
DB_PATH="${SQLITE_DB:-/var/db/secure-email.db}"

# Check if database exists
if [ ! -f "$DB_PATH" ]; then
    echo "Error: Database not found at $DB_PATH"
    echo "Please run the setup script first"
    exit 1
fi

echo "Creating test user: $EMAIL"

# Create a temporary Go program to add the user
cat > /tmp/create_user.go << 'EOF'
package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    _ "github.com/mattn/go-sqlite3"
    "github.com/your-username/secure-email-mvp/pkg/auth"
)

func main() {
    if len(os.Args) != 3 {
        log.Fatal("Usage: create_user <email> <password>")
    }
    
    email := os.Args[1]
    password := os.Args[2]
    
    // Connect to database
    dbPath := os.Getenv("SQLITE_DB")
    if dbPath == "" {
        dbPath = "/var/db/secure-email.db"
    }
    
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        log.Fatal("Error opening database:", err)
    }
    defer db.Close()
    
    // Create user
    userID, totpSecret, err := auth.CreateUser(db, email, password)
    if err != nil {
        log.Fatal("Error creating user:", err)
    }
    
    fmt.Printf("User created successfully!\n")
    fmt.Printf("Email: %s\n", email)
    fmt.Printf("User ID: %s\n", userID)
    fmt.Printf("TOTP Secret: %s\n", totpSecret)
    fmt.Printf("\nAdd this secret to your authenticator app:\n")
    fmt.Printf("otpauth://totp/SecureEmail:%s?secret=%s&issuer=SecureEmail\n", email, totpSecret)
}
EOF

# Run the user creation program
cd /opt/secure-email-mvp 2>/dev/null || cd .
go run /tmp/create_user.go "$EMAIL" "$PASSWORD"

# Clean up
rm /tmp/create_user.go

echo ""
echo "Test user created successfully!"
echo "You can now test the login API with these credentials." 