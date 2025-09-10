-- Production Database Initialization
-- This script sets up the production MySQL database with security best practices

-- Create the main database
CREATE DATABASE IF NOT EXISTS securesystem CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- Create application user with limited privileges
CREATE USER IF NOT EXISTS 'secureuser'@'%' IDENTIFIED BY '${DB_PASSWORD}';
CREATE USER IF NOT EXISTS 'secureuser'@'localhost' IDENTIFIED BY '${DB_PASSWORD}';

-- Grant necessary privileges to application user
GRANT SELECT, INSERT, UPDATE, DELETE ON securesystem.* TO 'secureuser'@'%';
GRANT SELECT, INSERT, UPDATE, DELETE ON securesystem.* TO 'secureuser'@'localhost';

-- Revoke unnecessary privileges
REVOKE CREATE, DROP, ALTER, INDEX, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, EVENT, TRIGGER ON *.* FROM 'secureuser'@'%';
REVOKE CREATE, DROP, ALTER, INDEX, CREATE TEMPORARY TABLES, LOCK TABLES, EXECUTE, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, EVENT, TRIGGER ON *.* FROM 'secureuser'@'localhost';

-- Use the database
USE securesystem;

-- Create users table with production optimizations
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,
    verification_token_hash CHAR(128),
    verification_exp TIMESTAMP,
    password_hash CHAR(128) NOT NULL,
    reset_token_hash CHAR(128),
    reset_token_exp TIMESTAMP,
    tier ENUM('free', 'premium', 'enterprise') NOT NULL DEFAULT 'free',
    custom_domain VARCHAR(255),
    domain_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    
    -- Add indexes for performance
    INDEX idx_email (email),
    INDEX idx_verification_token (verification_token_hash),
    INDEX idx_reset_token (reset_token_hash),
    INDEX idx_tier (tier),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create audit log table for security monitoring
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    details JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_user_id (user_id),
    INDEX idx_action (action),
    INDEX idx_created_at (created_at),
    
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create email verification attempts table
CREATE TABLE IF NOT EXISTS email_verification_attempts (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL,
    token_hash CHAR(128) NOT NULL,
    attempts INT DEFAULT 1,
    last_attempt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_email (email),
    INDEX idx_token_hash (token_hash),
    INDEX idx_last_attempt (last_attempt)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Flush privileges
FLUSH PRIVILEGES;

-- Security hardening
-- Disable root remote access
DELETE FROM mysql.user WHERE User='root' AND Host NOT IN ('localhost', '127.0.0.1', '::1');

-- Remove anonymous users
DELETE FROM mysql.user WHERE User='';

-- Remove test database
DROP DATABASE IF EXISTS test;

-- Flush privileges again
FLUSH PRIVILEGES;
