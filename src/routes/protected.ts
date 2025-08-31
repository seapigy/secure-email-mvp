/**
 * Protected routes
 * Routes that require authentication
 */

import { Router, Request, Response, NextFunction } from 'express';
import * as jwt from 'jsonwebtoken';
import { db } from '../lib/db';

const router = Router();

// JWT secret (in production, this should be in environment variables)
const JWT_SECRET = process.env.JWT_SECRET || 'test-secret-key';

interface JWTPayload {
  user_id: string;
  email: string;
  type: string;
  jti: string;
}

// Extend Express Request type to include user property
interface AuthenticatedRequest extends Request {
  user?: {
    id: string;
    email: string;
  };
}

/**
 * Authentication middleware
 * Verifies JWT token and adds user info to request
 */
const authenticateToken = async (req: AuthenticatedRequest, res: Response, next: NextFunction) => {
  const authHeader = req.headers['authorization'];
  const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

  if (!token) {
    return res.status(401).json({ error: 'Access token required' });
  }

  try {
    const decoded = jwt.verify(token, JWT_SECRET) as JWTPayload;
    
    // Verify token type
    if (decoded.type !== 'access') {
      return res.status(401).json({ error: 'Invalid token type' });
    }

    // Check if access token has been revoked (only if jti exists)
    if (decoded.jti) {
      const revokedToken = await db('refresh_tokens')
        .where('access_token_id', decoded.jti)
        .where('is_revoked', true)
        .first();

      if (revokedToken) {
        return res.status(401).json({ error: 'Token has been revoked' });
      }
    }

    // Add user info to request
    req.user = {
      id: decoded.user_id,
      email: decoded.email
    };

    next();
  } catch (error) {
    return res.status(401).json({ error: 'Invalid or expired token' });
  }
};

/**
 * GET /api/protected
 * Protected route that returns user information
 */
router.get('/protected', authenticateToken, async (req: AuthenticatedRequest, res: Response) => {
  try {
    const { user } = req;

    if (!user) {
      return res.status(401).json({ error: 'User not authenticated' });
    }

    // Get user from database to ensure they still exist and are active
    const userRecord = await db('users')
      .where('id', user.id)
      .where('is_active', true)
      .first();

    if (!userRecord) {
      return res.status(401).json({ error: 'User not found or inactive' });
    }

    // Return user information (excluding sensitive data)
    res.status(200).json({
      user_id: user.id,
      email: user.email,
      is_active: userRecord.is_active,
      email_verified: userRecord.email_verified,
      created_at: userRecord.created_at,
      updated_at: userRecord.updated_at
    });
  } catch (error) {
    console.error('Protected route error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * GET /api/user/profile
 * Returns user profile information
 */
router.get('/user/profile', authenticateToken, async (req: AuthenticatedRequest, res: Response) => {
  try {
    const { user } = req;

    if (!user) {
      return res.status(401).json({ error: 'User not authenticated' });
    }

    const userRecord = await db('users')
      .where('id', user.id)
      .where('is_active', true)
      .first();

    if (!userRecord) {
      return res.status(401).json({ error: 'User not found or inactive' });
    }

    // Return profile information (excluding sensitive data)
    res.status(200).json({
      user_id: user.id,
      email: user.email,
      is_active: userRecord.is_active,
      email_verified: userRecord.email_verified,
      created_at: userRecord.created_at,
      updated_at: userRecord.updated_at
    });
  } catch (error) {
    console.error('Profile route error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * PUT /api/user/profile
 * Updates user profile information
 */
router.put('/user/profile', authenticateToken, async (req: AuthenticatedRequest, res: Response) => {
  try {
    const { user } = req;
    const { email } = req.body;

    if (!user) {
      return res.status(401).json({ error: 'User not authenticated' });
    }

    // Validate email format if provided
    if (email) {
      const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
      if (!emailRegex.test(email)) {
        return res.status(400).json({ error: 'Invalid email format' });
      }

      // Check if email is already taken by another user
      const existingUser = await db('users')
        .where('email', email)
        .whereNot('id', user.id)
        .first();

      if (existingUser) {
        return res.status(400).json({ error: 'Email already in use' });
      }
    }

    // Update user profile
    const updateData: any = {
      updated_at: new Date()
    };

    if (email) {
      updateData.email = email;
      updateData.email_verified = false; // Reset verification when email changes
    }

    await db('users')
      .where('id', user.id)
      .update(updateData);

    res.status(200).json({ message: 'Profile updated successfully' });
  } catch (error) {
    console.error('Profile update error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

/**
 * POST /api/user/change-password
 * Changes user password
 */
router.post('/user/change-password', authenticateToken, async (req: AuthenticatedRequest, res: Response) => {
  try {
    const { user } = req;
    const { current_password, new_password } = req.body;

    if (!user) {
      return res.status(401).json({ error: 'User not authenticated' });
    }

    if (!current_password || !new_password) {
      return res.status(400).json({ error: 'Current password and new password are required' });
    }

    // Validate new password strength
    if (new_password.length < 8) {
      return res.status(400).json({ error: 'New password must be at least 8 characters long' });
    }

    // Get current user with password hash
    const userRecord = await db('users')
      .where('id', user.id)
      .where('is_active', true)
      .first();

    if (!userRecord) {
      return res.status(401).json({ error: 'User not found or inactive' });
    }

    // Verify current password
    const currentPasswordValid = await require('argon2').verify(userRecord.password_hash, current_password);
    if (!currentPasswordValid) {
      return res.status(401).json({ error: 'Current password is incorrect' });
    }

    // Hash new password
    const newPasswordHash = await require('argon2').hash(new_password, {
      type: require('argon2').argon2id,
      memoryCost: 2 ** 16,
      timeCost: 3,
      parallelism: 2,
    });

    // Update password
    await db('users')
      .where('id', user.id)
      .update({
        password_hash: newPasswordHash,
        updated_at: new Date()
      });

    // Revoke all refresh tokens for this user
    await db('refresh_tokens')
      .where('user_id', user.id)
      .update({ is_revoked: true });

    res.status(200).json({ message: 'Password changed successfully' });
  } catch (error) {
    console.error('Password change error:', error);
    res.status(500).json({ error: 'Internal server error' });
  }
});

export { router as protectedRoutes };
