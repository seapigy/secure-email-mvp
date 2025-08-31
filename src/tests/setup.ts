/**
 * Test setup file
 * Initializes test environment and database
 */

import { initializeTestDatabase, cleanupTestDatabase, closeConnection } from '../lib/db';

// Global test setup
beforeAll(async () => {
  console.log('🔧 Setting up test environment...');
  
  // Initialize test database
  await initializeTestDatabase();
  
  console.log('✅ Test environment ready');
});

// Ensure database is initialized before each test suite
beforeEach(async () => {
  // Re-initialize database for each test suite to ensure tables exist
  await initializeTestDatabase();
});

// Global test cleanup
afterAll(async () => {
  console.log('🧹 Cleaning up test environment...');
  
  // Clean up test database
  await cleanupTestDatabase();
  
  // Don't close the connection globally - let each test suite manage its own
  // await closeConnection();
  
  console.log('✅ Test environment cleaned up');
});

// Increase timeout for database operations
jest.setTimeout(30000);
