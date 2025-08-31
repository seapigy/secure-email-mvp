/**
 * Database connection and configuration
 * Supports both SQLite (for testing) and PostgreSQL (for production)
 */

import knex from 'knex';
import { config } from 'dotenv';

// Load environment variables
config();

// Database configuration
const dbConfig = {
  client: process.env.DB_CLIENT || 'sqlite3',
  connection: process.env.DATABASE_URL || {
    filename: process.env.DB_FILE || './data/test.db'
  },
  useNullAsDefault: true,
  migrations: {
    directory: '../migrations'
  },
  seeds: {
    directory: '../seeds'
  },
  pool: {
    min: 2,
    max: 10
  }
};

// Create database connection
export const db = knex(dbConfig);

// Test database connection
export async function testConnection(): Promise<boolean> {
  try {
    await db.raw('SELECT 1');
    console.log('✅ Database connection successful');
    return true;
  } catch (error) {
    console.error('❌ Database connection failed:', error);
    return false;
  }
}

// Initialize database tables for testing
export async function initializeTestDatabase(): Promise<void> {
  try {
    // Create users table if it doesn't exist
    await db.schema.createTableIfNotExists('users', (table) => {
      table.text('id').primary();
      table.text('email').unique().notNullable();
      table.text('password').notNullable();
      table.text('fallback_email');
      table.text('fallback_token');
      table.boolean('fallback_confirmed').defaultTo(false);
      table.timestamp('fallback_token_expiration');
      table.timestamp('created_at').defaultTo(db.fn.now());
      table.integer('failed_login_attempts').defaultTo(0);
      table.timestamp('last_failed_login');
      table.timestamp('account_locked_until');
      table.text('phone_number');
      table.text('totp_secret');
      table.text('password_hash');
      table.text('name');
      table.boolean('is_active').defaultTo(true);
      table.boolean('email_verified').defaultTo(false);
      table.timestamp('updated_at').defaultTo(db.fn.now());
    });

    // Create refresh_tokens table if it doesn't exist
    await db.schema.createTableIfNotExists('refresh_tokens', (table) => {
      table.uuid('id').primary();
      table.uuid('user_id').references('id').inTable('users').onDelete('CASCADE');
      table.text('token').notNullable();
      table.uuid('access_token_id'); // Link to access token for revocation
      table.timestamp('expires_at').notNullable();
      table.boolean('is_revoked').defaultTo(false);
      table.timestamp('created_at').defaultTo(db.fn.now());
    });

    console.log('✅ Test database initialized');
  } catch (error) {
    console.error('❌ Failed to initialize test database:', error);
    throw error;
  }
}

// Clean up test database
export async function cleanupTestDatabase(): Promise<void> {
  try {
    await db('refresh_tokens').del();
    await db('users').del();
    console.log('✅ Test database cleaned up');
  } catch (error) {
    console.error('❌ Failed to cleanup test database:', error);
    throw error;
  }
}

// Close database connection
export async function closeConnection(): Promise<void> {
  try {
    await db.destroy();
    console.log('✅ Database connection closed');
  } catch (error) {
    console.error('❌ Failed to close database connection:', error);
    throw error;
  }
}
