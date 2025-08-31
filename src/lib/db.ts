/**
 * Database connection and configuration
 * Supports both SQLite (for testing) and PostgreSQL (for production)
 */

import knex from 'knex';
import { config } from 'dotenv';

// Load environment variables
config();

// Database configuration
function getDbConfig() {
  return {
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
}

// Create database connection
export const db = knex(getDbConfig());

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
    const usersTableExists = await db.schema.hasTable('users');
    if (!usersTableExists) {
      await db.schema.createTable('users', (table) => {
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
    }

    // Create refresh_tokens table if it doesn't exist
    const refreshTokensTableExists = await db.schema.hasTable('refresh_tokens');
    if (!refreshTokensTableExists) {
      await db.schema.createTable('refresh_tokens', (table) => {
      table.uuid('id').primary();
      table.uuid('user_id').references('id').inTable('users').onDelete('CASCADE');
      table.text('token').notNullable();
      table.uuid('access_token_id'); // Link to access token for revocation
      table.timestamp('expires_at').notNullable();
      table.boolean('is_revoked').defaultTo(false);
      table.timestamp('created_at').defaultTo(db.fn.now());
      });
    }

    // Create emails table if it doesn't exist
    const emailsTableExists = await db.schema.hasTable('emails');
    if (!emailsTableExists) {
      await db.schema.createTable('emails', (table) => {
      table.text('email_id').primary();
      table.text('sender_id').references('id').inTable('users').onDelete('CASCADE');
      table.text('recipient').notNullable();
      table.text('subject');
      table.text('encrypted_blob_url');
      table.text('encrypted_key');
      table.text('encryption_nonce');
      table.text('encryption_auth_tag');
      table.text('compression_algo').defaultTo('gzip');
      table.text('sha256_hash');
      table.boolean('requires_password').defaultTo(false);
      table.text('password_hash');
      table.text('geolocation_json');
      table.timestamp('expires_at');
      table.boolean('burn_after_read').defaultTo(false);
      table.integer('failed_attempts').defaultTo(0);
      table.integer('max_attempts').defaultTo(3);
      table.boolean('self_destruct_after_attempts').defaultTo(false);
      table.boolean('reply_enabled').defaultTo(false);
      table.boolean('reply_requires_password').defaultTo(true);
      table.boolean('allow_forwarding').defaultTo(false);
      table.boolean('show_sender_metadata').defaultTo(false);
      table.boolean('metadata_stripped').defaultTo(true);
      table.boolean('is_honeytoken').defaultTo(false);
      table.text('secure_link_id');
      table.timestamp('link_created_at');
      table.timestamp('last_access_at');
      table.integer('access_count').defaultTo(0);
      table.timestamp('created_at').defaultTo(db.fn.now());
      table.timestamp('updated_at').defaultTo(db.fn.now());
      table.boolean('self_destructed').defaultTo(false);
      table.text('allowed_city');
      table.text('allowed_country');
      table.text('geo_verification_type').defaultTo('none');
      table.text('geo_city');
      table.text('geo_country');
      table.boolean('require_mfa').defaultTo(false);
      table.text('mfa_type');
      table.text('encrypted_totp_secret');
      table.integer('mfa_failed_attempts').defaultTo(0);
      table.timestamp('mfa_locked_until');
      table.integer('brute_force_failed_attempts').defaultTo(0);
      table.timestamp('brute_force_last_failed_attempt');
      table.timestamp('brute_force_lockout_until');
      table.integer('brute_force_max_attempts').defaultTo(3);
      table.integer('brute_force_lockout_duration_minutes').defaultTo(15);
      table.boolean('is_password_protected').defaultTo(false);
      table.text('password_salt');
      table.text('recipient_id').references('id').inTable('users');
      });
    }

    // Create indexes for emails table
    await db.raw('CREATE INDEX IF NOT EXISTS idx_emails_sender_id ON emails(sender_id)');
    await db.raw('CREATE INDEX IF NOT EXISTS idx_emails_recipient ON emails(recipient)');
    await db.raw('CREATE INDEX IF NOT EXISTS idx_emails_recipient_id ON emails(recipient_id)');
    await db.raw('CREATE INDEX IF NOT EXISTS idx_emails_created_at ON emails(created_at)');
    await db.raw('CREATE INDEX IF NOT EXISTS idx_emails_expires_at ON emails(expires_at)');

    console.log('✅ Test database initialized');
  } catch (error) {
    console.error('❌ Failed to initialize test database:', error);
    throw error;
  }
}

// Clean up test database
export async function cleanupTestDatabase(): Promise<void> {
  try {
    // Check if tables exist before trying to delete from them
    const refreshTokensExists = await db.schema.hasTable('refresh_tokens');
    const emailsExists = await db.schema.hasTable('emails');
    const usersExists = await db.schema.hasTable('users');

    if (refreshTokensExists) {
      await db('refresh_tokens').del();
    }
    if (emailsExists) {
      await db('emails').del();
    }
    if (usersExists) {
      await db('users').del();
    }
    
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
