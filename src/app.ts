/**
 * Express application for testing the secure email API
 * This file sets up the Express server with all necessary middleware and routes
 */

import express from 'express';
import cors from 'cors';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import { json } from 'body-parser';

// Import route handlers
import { authRoutes } from './routes/auth';
import { protectedRoutes } from './routes/protected';

// Create Express app
const app = express();

// Security middleware
app.use(helmet({
  contentSecurityPolicy: {
    directives: {
      defaultSrc: ["'self'"],
      styleSrc: ["'self'", "'unsafe-inline'"],
      scriptSrc: ["'self'"],
      imgSrc: ["'self'", "data:", "https:"],
    },
  },
  hsts: {
    maxAge: 31536000,
    includeSubDomains: true,
    preload: true
  }
}));

// CORS configuration
app.use(cors({
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ['http://localhost:3000'],
  credentials: true,
  methods: ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'],
  allowedHeaders: ['Content-Type', 'Authorization', 'X-Requested-With']
}));

// Rate limiting
const authLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 10, // limit each IP to 10 requests per windowMs
  message: {
    error: 'Too many authentication attempts, please try again later.'
  },
  standardHeaders: true,
  legacyHeaders: false,
});

const generalLimiter = rateLimit({
  windowMs: 15 * 60 * 1000, // 15 minutes
  max: 100, // limit each IP to 100 requests per windowMs
  message: {
    error: 'Too many requests, please try again later.'
  },
  standardHeaders: true,
  legacyHeaders: false,
});

// Body parsing middleware
app.use(json({ limit: '10mb' }));

// Apply rate limiting to auth routes
app.use('/api/auth', authLimiter);

// Apply general rate limiting to all other routes
app.use(generalLimiter);

// Health check endpoint
app.get('/health', (_req, res) => {
  res.status(200).json({ 
    status: 'healthy',
    timestamp: new Date().toISOString(),
    version: process.env.APP_VERSION || '1.0.0'
  });
});

// API routes
app.use('/api/auth', authRoutes);
app.use('/api', protectedRoutes);

// 404 handler
app.use('*', (_req, res) => {
  res.status(404).json({ error: 'Endpoint not found' });
});

// Global error handler
app.use((err: any, _req: express.Request, res: express.Response, _next: express.NextFunction) => {
  console.error('Global error handler:', err);
  
  // Don't expose internal errors in production
  const isProduction = process.env.NODE_ENV === 'production';
  
  if (err.type === 'entity.parse.failed') {
    return res.status(400).json({ error: 'Invalid JSON format' });
  }
  
  if (err.type === 'entity.too.large') {
    return res.status(413).json({ error: 'Request entity too large' });
  }
  
  res.status(500).json({ 
    error: isProduction ? 'Internal server error' : err.message 
  });
});

export { app };
