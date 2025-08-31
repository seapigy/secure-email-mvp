/**
 * TOTP (Time-based One-Time Password) implementation
 * RFC 6238 compliant TOTP generation and validation
 */

import { createHmac, timingSafeEqual } from 'crypto';

export interface TOTPConfig {
  period: number;      // Time step in seconds (default: 30)
  digits: number;      // Number of digits in code (default: 6)
  algorithm: string;   // Hash algorithm (default: 'sha1')
  window: number;      // Time window for validation (default: 1)
}

const DEFAULT_CONFIG: TOTPConfig = {
  period: 30,
  digits: 6,
  algorithm: 'sha1',
  window: 1
};

/**
 * Generate a TOTP code for the given secret and time
 * @param secret - Base32 encoded secret
 * @param time - Time to generate code for (defaults to current time)
 * @param config - TOTP configuration
 * @returns TOTP code as string
 */
export function generateTOTP(secret: string, time: Date = new Date(), config: Partial<TOTPConfig> = {}): string {
  const finalConfig = { ...DEFAULT_CONFIG, ...config };
  
  // Convert time to Unix timestamp
  const timestamp = Math.floor(time.getTime() / 1000);
  
  // Calculate time step
  const timeStep = Math.floor(timestamp / finalConfig.period);
  
  // Convert time step to buffer (8 bytes, big-endian)
  const timeBuffer = Buffer.alloc(8);
  timeBuffer.writeBigUInt64BE(BigInt(timeStep), 0);
  
  // Decode base32 secret
  const secretBuffer = base32Decode(secret);
  
  // Generate HMAC
  const hmac = createHmac(finalConfig.algorithm, secretBuffer);
  hmac.update(timeBuffer);
  const hash = hmac.digest();
  
  // Generate code using DT (Dynamic Truncation)
  const offset = hash[hash.length - 1] & 0xf;
  const code = ((hash[offset] & 0x7f) << 24) |
               ((hash[offset + 1] & 0xff) << 16) |
               ((hash[offset + 2] & 0xff) << 8) |
               (hash[offset + 3] & 0xff);
  
  // Convert to string with specified number of digits
  const codeStr = (code % Math.pow(10, finalConfig.digits)).toString();
  return codeStr.padStart(finalConfig.digits, '0');
}

/**
 * Verify a TOTP code against a secret
 * @param secret - Base32 encoded secret
 * @param token - TOTP code to verify
 * @param config - TOTP configuration
 * @returns true if code is valid, false otherwise
 */
export function verifyTOTP(secret: string, token: string, config: Partial<TOTPConfig> = {}): boolean {
  const finalConfig = { ...DEFAULT_CONFIG, ...config };
  const now = new Date();
  
  // Check current time step and surrounding steps within window
  for (let i = -finalConfig.window; i <= finalConfig.window; i++) {
    const checkTime = new Date(now.getTime() + (i * finalConfig.period * 1000));
    const expectedCode = generateTOTP(secret, checkTime, finalConfig);
    
    if (timingSafeEqual(Buffer.from(token), Buffer.from(expectedCode))) {
      return true;
    }
  }
  
  return false;
}

/**
 * Decode base32 string to buffer
 * @param str - Base32 encoded string
 * @returns Decoded buffer
 */
function base32Decode(str: string): Buffer {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  
  // Remove padding and convert to uppercase
  str = str.replace(/=+$/, '').toUpperCase();
  
  let bits = 0;
  let value = 0;
  const output: number[] = [];
  
  for (let i = 0; i < str.length; i++) {
    const char = str[i];
    const index = alphabet.indexOf(char);
    
    if (index === -1) {
      throw new Error(`Invalid base32 character: ${char}`);
    }
    
    value = (value << 5) | index;
    bits += 5;
    
    if (bits >= 8) {
      output.push((value >>> (bits - 8)) & 0xff);
      bits -= 8;
    }
  }
  
  return Buffer.from(output);
}

/**
 * Generate a random base32 secret
 * @param length - Length of secret in bytes (default: 20)
 * @returns Base32 encoded secret
 */
export function generateSecret(length: number = 20): string {
  const randomBytes = require('crypto').randomBytes(length);
  return base32Encode(randomBytes);
}

/**
 * Encode buffer to base32 string
 * @param buffer - Buffer to encode
 * @returns Base32 encoded string
 */
function base32Encode(buffer: Buffer): string {
  const alphabet = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ234567';
  const padding = '=';
  
  let bits = 0;
  let value = 0;
  let output = '';
  
  for (let i = 0; i < buffer.length; i++) {
    value = (value << 8) | buffer[i];
    bits += 8;
    
    while (bits >= 5) {
      output += alphabet[(value >>> (bits - 5)) & 0x1f];
      bits -= 5;
    }
  }
  
  if (bits > 0) {
    output += alphabet[(value << (5 - bits)) & 0x1f];
  }
  
  // Add padding
  const padLength = (8 - (output.length % 8)) % 8;
  output += padding.repeat(padLength);
  
  return output;
}
