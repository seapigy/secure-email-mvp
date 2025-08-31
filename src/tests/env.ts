/**
 * Environment setup for tests
 * Sets up environment variables and test configuration
 */

// Set test environment
process.env.NODE_ENV = 'test';

// Database configuration for tests
process.env.DB_CLIENT = 'sqlite3';
process.env.DB_FILE = ':memory:'; // Use in-memory database for tests

// JWT secret for tests
process.env.JWT_SECRET = 'test-jwt-secret-key-for-testing-only';

// Test mode flags
process.env.TEST_MODE = 'true';

// Disable rate limiting for tests (we'll test it explicitly)
process.env.DISABLE_RATE_LIMITING = 'true';

// Logging configuration for tests
process.env.LOG_LEVEL = 'error'; // Only log errors during tests

console.log('🔧 Test environment variables configured');
