/**
 * Authentication routes
 * Handles signup, login, logout, and token management
 */

import { Router, Request, Response } from 'express';
import { db } from '../lib/db';
import { verifyTOTP } from '../lib/totp';
import * as argon2 from 'argon2';
import * as jwt from 'jsonwebtoken';
import { v4 as uuidv4 } from 'uuid';

const router = Router();

// JWT secret (in production, this should be in environment variables)
const JWT_SECRET = process.env.JWT_SECRET || 'test-secret-key';

interface SignupRequest {
  email: string;
  password: string;
  name: string;
}

interface SignupResponse {
  user_id: string;
  email: string;
  name: string;
  message: string;
}

interface LoginRequest {
  email: string;
  password: string;
  totp_code?: string; // Optional for non-TOTP login
}

interface SimpleLoginRequest {
  email: string;
  password: string;
}

interface LoginResponse {
  access_token: string;
  refresh_token: string;
  token_type: string;
  expires_in: number;
  user_id: string;
  email: string;
}

interface LogoutRequest {
  refresh_token: string;
}

/**
 * POST /api/auth/signup
 * Creates a new user account
 */
router.post('/signup', async (req: Request, res: Response) => {
  try {
    const { email, password, name }: SignupRequest = req.body;

    // Validate required fields
    if (!email || !password || name === undefined || name === null) {
      return res.status(400).json({ 
        error: 'Email, password, and name are required' 
      });
    }

    // Validate email format (trim first)
    const trimmedEmail = email.trim();
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(trimmedEmail)) {
      return res.status(400).json({ 
        error: 'Invalid email format' 
      });
    }

    // Validate password strength (minimum 8 characters)
    if (password.length < 8) {
      return res.status(400).json({ 
        error: 'Password must be at least 8 characters long' 
      });
    }

    // Validate name (not empty and reasonable length)
    if (name.trim().length === 0 || name.length > 100) {
      return res.status(400).json({ 
        error: 'Name must be between 1 and 100 characters' 
      });
    }

    // Check if user already exists (case-insensitive)
    const existingUser = await db('users')
      .whereRaw('LOWER(email) = ?', [trimmedEmail.toLowerCase()])
      .first();

    if (existingUser) {
      return res.status(409).json({ 
        error: 'User with this email already exists' 
      });
    }

    // Generate user ID
    const userId = uuidv4();

    // Hash password using Argon2id
    const passwordHash = await argon2.hash(password, {
      type: argon2.argon2id,
      memoryCost: 2 ** 16, // 64MB
      timeCost: 3,         // 3 iterations
      parallelism: 2,      // 2 threads
    });

    // Generate TOTP secret for MFA
    const totpSecret = require('crypto').randomBytes(20).toString('base64').replace(/[^A-Z2-7]/g, '').substring(0, 32);

    // Insert new user
    await db('users').insert({
      id: userId,
      email: trimmedEmail.toLowerCase(),
      name: name.trim(),
      password: '', // Keep empty for legacy compatibility
      password_hash: passwordHash,
      totp_secret: totpSecret,
      created_at: new Date(),
      failed_login_attempts: 0,
      fallback_confirmed: false,
      is_active: true,
      email_verified: false, // New users need to verify their email
      updated_at: new Date()
    });

    // Return success response
    const response: SignupResponse = {
      user_id: userId,
      email: trimmedEmail.toLowerCase(),
      name: name.trim(),
      message: 'User account created successfully'
    };

    res.status(201).json(response);
  } catch (error) {
    console.error('Signup error:', error);
    res.status(500).json({ 
      error: 'Internal server error' 
    });
  }
});

/**
 * POST /api/auth/login
 * Authenticates user with email and password (TOTP optional)
 */
router.post('/login', async (req, res) => {
  try {
    const { email, password, totp_code }: LoginRequest = req.body;

    // Validate required fields
    if (!email || !password) {
      return res.status(400).json({ error: 'Email and password are required' });
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return res.status(400).json({ error: 'Invalid email format' });
    }

    // Find user by email
    const user = await db('users')
      .where('email', email)
      .where('is_active', true)
      .first();

    if (!user) {
      return res.status(401).json({ error: 'Invalid credentials' });
    }

    // Verify password
    const passwordValid = await argon2.verify(user.password_hash, password);
    if (!passwordValid) {
      return res.status(401).json({ error: 'Invalid credentials' });
    }

    // Verify TOTP code if provided
    if (totp_code) {
      const totpValid = verifyTOTP(user.totp_secret, totp_code);
      if (!totpValid) {
        return res.status(401).json({ error: 'Invalid credentials' });
      }
    }

    // Generate access token with jti (JWT ID) for revocation tracking
    const accessTokenId = uuidv4();
    const accessToken = jwt.sign(
      {
        user_id: user.id,
        email: user.email,
        type: 'access',
        jti: accessTokenId
      },
      JWT_SECRET,
      { expiresIn: '1h' }
    );

    // Generate refresh token
    const refreshToken = uuidv4();
    const refreshTokenExpiry = new Date();
    refreshTokenExpiry.setDate(refreshTokenExpiry.getDate() + 7); // 7 days

    // Store refresh token in database
    await db('refresh_tokens').insert({
      id: uuidv4(),
      user_id: user.id,
      token: refreshToken,
      access_token_id: accessTokenId, // Link to access token
      expires_at: refreshTokenExpiry,
      is_revoked: false
    });

    // Return response
    const response: LoginResponse = {
      access_token: accessToken,
      refresh_token: refreshToken,
      token_type: 'Bearer',
      expires_in: 3600, // 1 hour
      user_id: user.id,
      email: user.email
    };

    res.status(200).json(response);
  } catch (error) {
    console.error('Login error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /api/auth/logout
 * Logs out user and revokes refresh token
 */
router.post('/logout', async (req, res) => {
  try {
    const { refresh_token }: LogoutRequest = req.body;

    if (!refresh_token) {
      return res.status(400).json({ error: 'Refresh token is required' });
    }

    // Check if refresh token exists
    const tokenRecord = await db('refresh_tokens')
      .where('token', refresh_token)
      .first();

    if (!tokenRecord) {
      return res.status(400).json({ error: 'Invalid refresh token' });
    }

    // Revoke refresh token and associated access token
    await db('refresh_tokens')
      .where('token', refresh_token)
      .update({ is_revoked: true });

    res.status(200).json({ message: 'Logout successful' });
  } catch (error) {
    console.error('Logout error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /api/auth/simple-login
 * Simple login with email and password only (no TOTP)
 */
router.post('/simple-login', async (req, res) => {
  try {
    const { email, password }: SimpleLoginRequest = req.body;

    // Validate required fields
    if (!email || !password) {
      return res.status(400).json({ error: 'Email and password are required' });
    }

    // Validate email format
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
    if (!emailRegex.test(email)) {
      return res.status(400).json({ error: 'Invalid email format' });
    }

    // Find user by email
    const user = await db('users')
      .where('email', email)
      .where('is_active', true)
      .first();

    if (!user) {
      return res.status(401).json({ error: 'Invalid credentials' });
    }

    // Verify password
    const passwordValid = await argon2.verify(user.password_hash, password);
    if (!passwordValid) {
      return res.status(401).json({ error: 'Invalid credentials' });
    }

    // Generate access token
    const accessTokenId = uuidv4();
    const accessToken = jwt.sign(
      {
        user_id: user.id,
        email: user.email,
        type: 'access',
        jti: accessTokenId
      },
      JWT_SECRET,
      { expiresIn: '1h' }
    );

    // Generate refresh token
    const refreshToken = uuidv4();
    const refreshTokenExpiry = new Date();
    refreshTokenExpiry.setDate(refreshTokenExpiry.getDate() + 7); // 7 days

    // Store refresh token in database
    await db('refresh_tokens').insert({
      id: uuidv4(),
      user_id: user.id,
      token: refreshToken,
      access_token_id: accessTokenId,
      expires_at: refreshTokenExpiry,
      is_revoked: false
    });

    // Return response
    const response: LoginResponse = {
      access_token: accessToken,
      refresh_token: refreshToken,
      token_type: 'Bearer',
      expires_in: 3600, // 1 hour
      user_id: user.id,
      email: user.email
    };

    res.status(200).json(response);
  } catch (error) {
    console.error('Simple login error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /api/auth/refresh
 * Refreshes access token using refresh token
 */
router.post('/refresh', async (req, res) => {
  try {
    const { refresh_token } = req.body;

    if (!refresh_token) {
      return res.status(400).json({ error: 'Refresh token is required' });
    }

    // Find valid refresh token
    const tokenRecord = await db('refresh_tokens')
      .where('token', refresh_token)
      .where('is_revoked', false)
      .where('expires_at', '>', new Date())
      .first();

    if (!tokenRecord) {
      return res.status(401).json({ error: 'Invalid or expired refresh token' });
    }

    // Get user information
    const user = await db('users')
      .where('id', tokenRecord.user_id)
      .where('is_active', true)
      .first();

    if (!user) {
      return res.status(401).json({ error: 'User not found' });
    }

    // Generate new access token
    const accessToken = jwt.sign(
      {
        user_id: user.id,
        email: user.email,
        type: 'access'
      },
      JWT_SECRET,
      { expiresIn: '1h' }
    );

    res.status(200).json({
      access_token: accessToken,
      token_type: 'Bearer',
      expires_in: 3600
    });
  } catch (error) {
    console.error('Token refresh error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export { router as authRoutes };
