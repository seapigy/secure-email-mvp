// Web-specific configuration for SecureMail Frontend
export const WEB_CONFIG = {
  // Web deployment settings
  BASE_URL: process.env.EXPO_PUBLIC_WEB_BASE_URL || 'https://app.securesystem.email',
  
  // CORS configuration
  CORS_ENABLED: process.env.EXPO_PUBLIC_CORS_ENABLED === 'true',
  
  // Security headers
  SECURITY_HEADERS: {
    'X-Content-Type-Options': 'nosniff',
    'X-Frame-Options': 'DENY',
    'X-XSS-Protection': '1; mode=block',
    'Strict-Transport-Security': 'max-age=31536000; includeSubDomains',
    'Content-Security-Policy': "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline';",
  },
  
  // Session storage configuration
  SESSION_STORAGE: {
    TOKEN_KEY: 'securemail_auth_token',
    USER_KEY: 'securemail_user_data',
    EXPIRY_KEY: 'securemail_token_expiry',
  },
  
  // Analytics configuration
  ANALYTICS: {
    ENABLED: process.env.EXPO_PUBLIC_ENABLE_ANALYTICS === 'true',
    DEBUG: process.env.EXPO_PUBLIC_DEBUG === 'true',
  },
} as const;
