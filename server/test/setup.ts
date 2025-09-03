/**
 * ⚠️ CRITICAL WARNING - TEST SETUP PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS ESSENTIAL TEST SETUP AND UTILITIES.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER remove testing library imports
 * 2. NEVER remove accessibility testing setup
 * 3. NEVER remove security testing utilities
 * 4. NEVER remove performance testing helpers
 * 5. ALWAYS maintain error handling test setup
 * 
 * This setup file ensures all tests have proper configuration.
 * 
 * @author: AI Assistant
 * @warning: TEST SETUP PRESERVATION CRITICAL
 * @last_updated: Priority 7 - Testing Enhancements
 */

import '@testing-library/jest-dom';
import { vi } from 'vitest';

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
});

// Mock IntersectionObserver
(globalThis as unknown as { IntersectionObserver: unknown }).IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock ResizeObserver
(globalThis as unknown as { ResizeObserver: unknown }).ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}));

// Mock performance API
Object.defineProperty(window, 'performance', {
  writable: true,
  value: {
    now: vi.fn(() => Date.now()),
    mark: vi.fn(),
    measure: vi.fn(),
    getEntriesByType: vi.fn(() => []),
    memory: {
      usedJSHeapSize: 50 * 1024 * 1024, // 50MB
      totalJSHeapSize: 100 * 1024 * 1024, // 100MB
      jsHeapSizeLimit: 200 * 1024 * 1024, // 200MB
    },
  },
});

// Mock requestIdleCallback
Object.defineProperty(window, 'requestIdleCallback', {
  value: vi.fn((callback) => {
    // Execute callback immediately in test environment
    setTimeout(callback, 0);
    return 1; // Return a mock ID
  }),
  writable: true,
});

// Mock PerformanceObserver
Object.defineProperty(window, 'PerformanceObserver', {
  value: class MockPerformanceObserver {
    constructor(_callback: unknown) {
      // TODO: Use callback in mock implementation
      // this._callback = callback;
    }
    observe(_options: unknown) {
      // Mock implementation
    }
    disconnect() {
      // Mock implementation
    }
    // TODO: Use callback in mock implementation
    // private _callback: any;
  },
  writable: true,
});

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
(globalThis as unknown as { localStorage: unknown }).localStorage = localStorageMock;

// Mock sessionStorage
const sessionStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
(globalThis as unknown as { sessionStorage: unknown }).sessionStorage = sessionStorageMock;

// Security testing utilities
export const createSecurityTestPayload = (type: 'xss' | 'sql' | 'file') => {
  const payloads = {
    xss: [
      '<script>alert("xss")</script>',
      'javascript:alert("xss")',
      '<img src=x onerror=alert(1)>',
      '<iframe src="javascript:alert(1)"></iframe>',
      'data:text/html,<script>alert(1)</script>',
    ],
    sql: [
      "'; DROP TABLE users; --",
      "' OR '1'='1",
      "'; INSERT INTO users VALUES ('hacker', 'password'); --",
    ],
    file: [
      { name: 'malicious.exe', size: 1024, type: 'application/octet-stream' },
      { name: 'script.js', size: 1024, type: 'application/javascript' },
      { name: 'test.bat', size: 1024, type: 'application/x-msdownload' },
    ],
  };
  return payloads[type];
};

// Accessibility testing utilities
export const checkAccessibility = async (container: HTMLElement) => {
  // Check for proper ARIA labels
  const elementsWithoutAria = container.querySelectorAll('button, input, select, textarea');
  const accessibilityIssues: string[] = [];

  elementsWithoutAria.forEach((element) => {
    const hasLabel = element.hasAttribute('aria-label') || 
                    element.hasAttribute('aria-labelledby') ||
                    element.hasAttribute('title');
    
    if (!hasLabel && element.tagName !== 'BUTTON') {
      accessibilityIssues.push(`Element ${element.tagName} missing accessibility label`);
    }
  });

  // Check for proper heading structure
  const headings = container.querySelectorAll('h1, h2, h3, h4, h5, h6');
  let previousLevel = 0;
  
  headings.forEach((heading) => {
    const level = parseInt(heading.tagName.charAt(1));
    if (level > previousLevel + 1) {
      accessibilityIssues.push(`Heading structure issue: ${heading.tagName} follows ${previousLevel}`);
    }
    previousLevel = level;
  });

  return accessibilityIssues;
};

// Performance testing utilities
export const measurePerformance = async (callback: () => void | Promise<void>) => {
  const startTime = performance.now();
  await callback();
  const endTime = performance.now();
  return endTime - startTime;
};

// Error handling testing utilities
export const createMockError = (type: 'network' | 'validation' | 'security' | 'api') => {
  const errors = {
    network: new Error('Network error: Failed to fetch'),
    validation: new Error('Validation error: Invalid input'),
    security: new Error('Security error: Access denied'),
    api: new Error('API error: Internal server error'),
  };
  return errors[type];
};

// Form testing utilities
export const fillFormFields = async (container: HTMLElement, data: Record<string, string>) => {
  const user = (await import('@testing-library/user-event')).default.setup();
  
  for (const [fieldName, value] of Object.entries(data)) {
    const field = container.querySelector(`[name="${fieldName}"], [id="${fieldName}"], [data-testid="${fieldName}"]`) as HTMLInputElement;
    if (field) {
      await user.type(field, value);
    }
  }
};

// File upload testing utilities
export const createMockFile = (name: string, size: number, type: string): File => {
  const file = new File(['test content'], name, { type });
  Object.defineProperty(file, 'size', { value: size });
  Object.defineProperty(file, 'lastModified', { value: Date.now() });
  return file;
};

// Mock toast notifications
export const mockToast = {
  error: vi.fn(),
  success: vi.fn(),
  warning: vi.fn(),
  info: vi.fn(),
};

// Test data constants
export const TEST_DATA = {
  VALID_EMAIL: 'test@example.com',
  INVALID_EMAIL: 'invalid-email',
  LONG_SUBJECT: 'A'.repeat(201),
  LONG_BODY: 'A'.repeat(10001),
  WEAK_PASSWORD: '123456',
  STRONG_PASSWORD: 'SecurePass123!',
  VALID_CITY: 'New York',
  INVALID_CITY: 'New York<script>alert("xss")</script>',
  VALID_COUNTRY: 'United States',
  INVALID_COUNTRY: 'United States<img src=x onerror=alert(1)>',
  VALID_DECOY: 'mysecret',
  INVALID_DECOY: 'password123',
};

// Cleanup after each test
// TODO: Import afterEach from vitest
// afterEach(() => {
//   vi.clearAllMocks();
//   localStorageMock.clear();
//   sessionStorageMock.clear();
// });



