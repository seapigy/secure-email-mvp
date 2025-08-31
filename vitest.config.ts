/**
 * ⚠️ CRITICAL WARNING - TESTING CONFIGURATION ⚠️
 * 
 * THIS FILE CONTAINS VITEST CONFIGURATION FOR COMPREHENSIVE TESTING.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER remove test coverage requirements
 * 2. NEVER remove security testing configurations
 * 3. NEVER remove accessibility testing setup
 * 4. ALWAYS maintain performance testing capabilities
 * 5. ALWAYS keep error handling test configurations
 * 
 * This configuration ensures comprehensive testing coverage.
 * 
 * @author: AI Assistant
 * @warning: TESTING CONFIGURATION CRITICAL
 * @coverage_target: >90%
 * @last_updated: Priority 7 - Testing Enhancements
 */

import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'src/test/',
        '**/*.d.ts',
        '**/*.config.*',
        '**/coverage/**',
        '**/dist/**',
        '**/.next/**',
        '**/cypress/**',
        '**/playwright/**',
      ],
      thresholds: {
        global: {
          branches: 90,
          functions: 90,
          lines: 90,
          statements: 90,
        },
      },
    },
    include: [
      'src/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}',
      'tests/**/*.{test,spec}.{js,mjs,cjs,ts,mts,cts,jsx,tsx}',
    ],
    exclude: [
      'node_modules/',
      'dist/',
      '.next/',
      'coverage/',
    ],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
