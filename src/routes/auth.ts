/**
 * Authentication routes
 * Handles login, logout, and token management
 */

import { Router } from 'express';
import { db } from '../lib/db';
import { verifyTOTP } from '../lib/totp';
import * as argon2 from 'argon2';
import * as jwt from 'jsonwebtoken';
import { v4 as uuidv4 } from 'uuid';

const router = Router();

// JWT secret (in production, this should be in environment variables)
const JWT_SECRET = process.env.JWT_SECRET || 'test-secret-key';

interface LoginRequest {
  email: string;
  password: string;
  totp_code: string;
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
 * POST /api/auth/login
 * Authenticates user with email, password, and TOTP code
 */
router.post('/login', async (req, res) => {
  try {
    const { email, password, totp_code }: LoginRequest = req.body;

    // Validate required fields
    if (!email || !password || !totp_code) {
      return res.status(400).json({ error: 'Email, password, and TOTP code are required' });
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

    // Verify TOTP code
    const totpValid = verifyTOTP(user.totp_secret, totp_code);
    if (!totpValid) {
      return res.status(401).json({ error: 'Invalid credentials' });
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
