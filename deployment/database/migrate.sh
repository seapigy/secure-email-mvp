#!/bin/bash

# Database Migration Script for Production
set -e

# Configuration
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-3306}"
DB_USER="${DB_USER:-secureuser}"
DB_PASSWORD="${DB_PASSWORD}"
DB_NAME="${DB_NAME:-securesystem}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting database migration...${NC}"

# Wait for database to be ready
echo -e "${YELLOW}Waiting for database to be ready...${NC}"
until mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1" >/dev/null 2>&1; do
    echo "Database is unavailable - sleeping"
    sleep 2
done

echo -e "${GREEN}Database is ready!${NC}"

# Run initialization script
echo -e "${YELLOW}Running database initialization...${NC}"
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" < /opt/secure-email/deployment/database/init.sql

# Run migration files
echo -e "${YELLOW}Running migration files...${NC}"
for migration_file in /opt/secure-email/migrations/*.sql; do
    if [ -f "$migration_file" ]; then
        echo "Running migration: $(basename "$migration_file")"
        mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" < "$migration_file"
    fi
done

# Verify tables exist
echo -e "${YELLOW}Verifying database schema...${NC}"
TABLES=$(mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SHOW TABLES;" -s --skip-column-names)

if echo "$TABLES" | grep -q "users"; then
    echo -e "${GREEN}✓ Users table created successfully${NC}"
else
    echo -e "${RED}✗ Users table not found${NC}"
    exit 1
fi

if echo "$TABLES" | grep -q "audit_logs"; then
    echo -e "${GREEN}✓ Audit logs table created successfully${NC}"
else
    echo -e "${RED}✗ Audit logs table not found${NC}"
    exit 1
fi

if echo "$TABLES" | grep -q "email_verification_attempts"; then
    echo -e "${GREEN}✓ Email verification attempts table created successfully${NC}"
else
    echo -e "${RED}✗ Email verification attempts table not found${NC}"
    exit 1
fi

# Test database connection
echo -e "${YELLOW}Testing database connection...${NC}"
mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SELECT COUNT(*) as user_count FROM users;" >/dev/null 2>&1

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Database connection test successful${NC}"
else
    echo -e "${RED}✗ Database connection test failed${NC}"
    exit 1
fi

echo -e "${GREEN}Database migration completed successfully!${NC}"
