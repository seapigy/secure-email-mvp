/**
 * JWT Authentication Middleware
 * Handles JWT token verification and user authentication
 */

import { Request, Response, NextFunction } from 'express';
import * as jwt from 'jsonwebtoken';
import { db } from '../lib/db';

// JWT secret (in production, this should be in environment variables)
const JWT_SECRET = process.env.JWT_SECRET || 'test-secret-key';

// Extend Express Request type to include user property
export interface AuthenticatedRequest extends Request {
  user?: {
    id: string;
    email: string;
    type: string;
    jti: string;
  };
}

interface JWTPayload {
  user_id: string;
  email: string;
  type: string;
  jti: string;
}

/**
 * Middleware to authenticate JWT tokens
 */
export const authenticateToken = async (
  req: AuthenticatedRequest,
  res: Response,
  next: NextFunction
): Promise<void> => {
  try {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1]; // Bearer TOKEN

    if (!token) {
      res.status(401).json({ error: 'Access token required' });
      return;
    }

    // Verify JWT token
    const decoded = jwt.verify(token, JWT_SECRET) as JWTPayload;

    // Verify token type
    if (decoded.type !== 'access') {
      res.status(401).json({ error: 'Invalid token type' });
      return;
    }

    // Check if access token has been revoked (only if jti exists)
    if (decoded.jti) {
      const revokedToken = await db('refresh_tokens')
        .where('access_token_id', decoded.jti)
        .where('is_revoked', true)
        .first();

      if (revokedToken) {
        res.status(401).json({ error: 'Token has been revoked' });
        return;
      }
    }

    // Add user information to request
    req.user = {
      id: decoded.user_id,
      email: decoded.email,
      type: decoded.type,
      jti: decoded.jti
    };

    next();
  } catch (error) {
    if (error instanceof jwt.JsonWebTokenError) {
      res.status(401).json({ error: 'Invalid token' });
    } else if (error instanceof jwt.TokenExpiredError) {
      res.status(401).json({ error: 'Token expired' });
    } else {
      console.error('JWT authentication error:', error);
      res.status(500).json({ error: 'Internal server error' });
    }
  }
};

/**
 * Optional authentication middleware (doesn't fail if no token)
 */
export const optionalAuth = async (
  req: AuthenticatedRequest,
  _res: Response,
  next: NextFunction
): Promise<void> => {
  try {
    const authHeader = req.headers['authorization'];
    const token = authHeader && authHeader.split(' ')[1];

    if (token) {
      const decoded = jwt.verify(token, JWT_SECRET) as JWTPayload;
      req.user = {
        id: decoded.user_id,
        email: decoded.email,
        type: decoded.type,
        jti: decoded.jti
      };
    }

    next();
  } catch (error) {
    // Continue without authentication if token is invalid
    next();
  }
};
