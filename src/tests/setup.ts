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

// Global test cleanup
afterAll(async () => {
  console.log('🧹 Cleaning up test environment...');
  
  // Clean up test database
  await cleanupTestDatabase();
  
  // Close database connection
  await closeConnection();
  
  console.log('✅ Test environment cleaned up');
});

// Increase timeout for database operations
jest.setTimeout(30000);
