# Testing Guide

## Overview

This document provides comprehensive guidance for testing the Secure Email application, with a focus on security, accessibility, performance, and reliability.

## Testing Strategy

### 1. Unit Testing
- **Coverage Target**: >90% for all critical components
- **Framework**: Vitest with React Testing Library
- **Focus**: Individual component functionality and edge cases

### 2. Integration Testing
- **Scope**: Component interactions and API integration
- **Focus**: End-to-end user workflows and data flow

### 3. Security Testing
- **XSS Prevention**: Test all user inputs for XSS vulnerabilities
- **SQL Injection**: Validate input sanitization
- **File Upload Security**: Test malicious file detection
- **Authentication**: Test security feature implementations

### 4. Accessibility Testing
- **WCAG 2.1 AA Compliance**: All components must meet accessibility standards
- **Screen Reader Support**: Test with NVDA, JAWS, VoiceOver
- **Keyboard Navigation**: Complete keyboard-only operation
- **Focus Management**: Proper focus handling and indicators

### 5. Performance Testing
- **Render Performance**: Components must render within 16ms (60fps)
- **Memory Usage**: Monitor for memory leaks
- **Bundle Size**: Track component bundle sizes
- **User Interaction**: Test debounced inputs and optimizations

## Test Structure

### File Naming Convention
```
ComponentName.test.tsx          # Unit tests
ComponentName.integration.test.tsx  # Integration tests
ComponentName.security.test.tsx     # Security tests
ComponentName.accessibility.test.tsx # Accessibility tests
ComponentName.performance.test.tsx  # Performance tests
```

### Test Organization
```typescript
describe('ComponentName', () => {
  describe('Basic Functionality', () => {
    // Core functionality tests
  });

  describe('Security Features', () => {
    // Security-specific tests
  });

  describe('Accessibility', () => {
    // Accessibility tests
  });

  describe('Performance', () => {
    // Performance tests
  });

  describe('Error Handling', () => {
    // Error scenarios
  });

  describe('Integration', () => {
    // Integration tests
  });
});
```

## Testing Utilities

### Security Testing
```typescript
import { createSecurityTestPayload } from '@/test/setup';

// Test XSS prevention
const xssPayloads = createSecurityTestPayload('xss');
xssPayloads.forEach(payload => {
  // Test that payload is sanitized
});

// Test SQL injection prevention
const sqlPayloads = createSecurityTestPayload('sql');
sqlPayloads.forEach(payload => {
  // Test that payload is blocked
});

// Test file upload security
const maliciousFiles = createSecurityTestPayload('file');
maliciousFiles.forEach(file => {
  // Test that file is rejected
});
```

### Accessibility Testing
```typescript
import { checkAccessibility } from '@/test/setup';

it('meets accessibility standards', async () => {
  const { container } = render(<Component />);
  const issues = await checkAccessibility(container);
  expect(issues).toHaveLength(0);
});
```

### Performance Testing
```typescript
import { measurePerformance } from '@/test/setup';

it('renders within performance budget', async () => {
  const renderTime = await measurePerformance(() => {
    render(<Component />);
  });
  expect(renderTime).toBeLessThan(16); // 60fps threshold
});
```

### Form Testing
```typescript
import { fillFormFields } from '@/test/setup';

it('handles form submission correctly', async () => {
  const { container } = render(<FormComponent />);
  
  await fillFormFields(container, {
    email: 'test@example.com',
    subject: 'Test Subject',
    body: 'Test message'
  });
  
  // Test form submission
});
```

## Test Data

### Standard Test Data
```typescript
import { TEST_DATA } from '@/test/setup';

// Valid data
TEST_DATA.VALID_EMAIL
TEST_DATA.STRONG_PASSWORD
TEST_DATA.VALID_CITY
TEST_DATA.VALID_COUNTRY

// Invalid data
TEST_DATA.INVALID_EMAIL
TEST_DATA.WEAK_PASSWORD
TEST_DATA.INVALID_CITY
TEST_DATA.INVALID_COUNTRY
```

### Security Test Data
```typescript
// XSS payloads
'<script>alert("xss")</script>'
'javascript:alert("xss")'
'<img src=x onerror=alert(1)>'

// SQL injection payloads
"'; DROP TABLE users; --"
"' OR '1'='1"

// Malicious files
{ name: 'malicious.exe', size: 1024, type: 'application/octet-stream' }
{ name: 'script.js', size: 1024, type: 'application/javascript' }
```

## Running Tests

### Basic Commands
```bash
# Run all tests
npm run test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run specific test categories
npm run test:security
npm run test:accessibility
npm run test:performance
npm run test:integration
```

### Test Coverage Requirements
- **Statements**: >90%
- **Branches**: >90%
- **Functions**: >90%
- **Lines**: >90%

## Best Practices

### 1. Test Organization
- Group related tests using `describe` blocks
- Use descriptive test names that explain the scenario
- Keep tests focused and atomic
- Avoid test interdependence

### 2. Security Testing
- Always test with malicious inputs
- Verify input sanitization
- Test file upload security
- Validate authentication flows

### 3. Accessibility Testing
- Test with screen readers
- Verify keyboard navigation
- Check focus management
- Validate ARIA attributes

### 4. Performance Testing
- Monitor render times
- Test with large datasets
- Verify memory usage
- Check bundle sizes

### 5. Error Handling
- Test all error scenarios
- Verify error messages
- Test recovery mechanisms
- Validate user feedback

## Common Patterns

### Component Testing
```typescript
describe('ComponentName', () => {
  let mockProps: ComponentProps;
  let mockOnAction: ReturnType<typeof vi.fn>;

  beforeEach(() => {
    mockOnAction = vi.fn();
    mockProps = {
      onAction: mockOnAction,
      // ... other props
    };
  });

  it('renders correctly', () => {
    render(<Component {...mockProps} />);
    expect(screen.getByText('Expected Text')).toBeInTheDocument();
  });

  it('handles user interactions', async () => {
    const user = userEvent.setup();
    render(<Component {...mockProps} />);
    
    const button = screen.getByRole('button');
    await user.click(button);
    
    expect(mockOnAction).toHaveBeenCalled();
  });
});
```

### API Testing
```typescript
describe('API Integration', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('handles successful API calls', async () => {
    mockApi.mockResolvedValue({ success: true });
    
    // Test component behavior
    await waitFor(() => {
      expect(mockApi).toHaveBeenCalled();
    });
  });

  it('handles API errors', async () => {
    mockApi.mockRejectedValue(new Error('API Error'));
    
    // Test error handling
    await waitFor(() => {
      expect(screen.getByText('Error message')).toBeInTheDocument();
    });
  });
});
```

## Continuous Integration

### Pre-commit Hooks
- Run unit tests
- Check test coverage
- Run security tests
- Validate accessibility

### CI Pipeline
1. **Install Dependencies**
2. **Run Linting**
3. **Run Type Checking**
4. **Run Unit Tests**
5. **Run Integration Tests**
6. **Run Security Tests**
7. **Generate Coverage Report**
8. **Deploy if all tests pass**

## Troubleshooting

### Common Issues

1. **Test Timeouts**
   - Increase timeout for async operations
   - Use `waitFor` for async assertions
   - Mock external dependencies

2. **Mock Issues**
   - Ensure mocks are properly set up
   - Clear mocks between tests
   - Use `vi.clearAllMocks()` in `afterEach`

3. **Accessibility Failures**
   - Check ARIA attributes
   - Verify keyboard navigation
   - Test with screen readers

4. **Performance Failures**
   - Optimize component rendering
   - Use React.memo for expensive components
   - Implement proper memoization

### Debugging Tests
```typescript
// Enable debug logging
import { screen } from '@testing-library/react';

// Debug component output
screen.debug();

// Debug specific element
screen.debug(screen.getByRole('button'));

// Log test data
console.log('Test data:', testData);
```

## Resources

- [Vitest Documentation](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Jest DOM Matchers](https://github.com/testing-library/jest-dom)
- [WCAG 2.1 Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)
- [OWASP Testing Guide](https://owasp.org/www-project-web-security-testing-guide/)
