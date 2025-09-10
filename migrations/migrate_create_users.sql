CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255) NOT NULL UNIQUE,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE,     -- Has user confirmed email?
    verification_token_hash CHAR(128),                 -- PQC-safe email verification token
    verification_exp TIMESTAMP,                        -- Expiration of token
    password_hash CHAR(128) NOT NULL,                 -- Argon2id hashed
    reset_token_hash CHAR(128),                        -- Offline recovery token hash
    reset_token_exp TIMESTAMP,                          -- Expiration
    tier ENUM('free', 'premium', 'enterprise') NOT NULL DEFAULT 'free',
    custom_domain VARCHAR(255),                         -- Premium-specific
    domain_verified BOOLEAN DEFAULT FALSE,             -- Premium only
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    last_login TIMESTAMP
);
