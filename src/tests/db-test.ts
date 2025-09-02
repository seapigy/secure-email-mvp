/**
 * Simple database test to verify table creation
 */

import { db, initializeTestDatabase, cleanupTestDatabase } from '../lib/db';

describe('Database Setup', () => {
  beforeAll(async () => {
    await initializeTestDatabase();
  });

  afterAll(async () => {
    await cleanupTestDatabase();
  });

  it('should have users table', async () => {
    const usersExists = await db.schema.hasTable('users');
    expect(usersExists).toBe(true);
  });

  it('should have refresh_tokens table', async () => {
    const refreshTokensExists = await db.schema.hasTable('refresh_tokens');
    expect(refreshTokensExists).toBe(true);
  });

  it('should have emails table', async () => {
    const emailsExists = await db.schema.hasTable('emails');
    expect(emailsExists).toBe(true);
  });

  it('should be able to insert and query from users table', async () => {
    const testUser = {
      id: 'test-user-123',
      email: 'test@example.com',
      password: 'hashedpassword',
      name: 'Test User'
    };

    await db('users').insert(testUser);
    
    const user = await db('users').where('id', testUser.id).first();
    expect(user).toBeDefined();
    expect(user.email).toBe(testUser.email);
  });
});
