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
    last_login TIMESTAMP
);

