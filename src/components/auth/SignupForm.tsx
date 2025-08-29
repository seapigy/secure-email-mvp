/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE BASIC SIGNUP FORM COMPONENT.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the visual layout or design of any components
 * 2. NEVER modify the routing structure that affects existing designs
 * 3. NEVER alter the component imports that could affect the "perfect" design
 * 4. ONLY add new functionality that doesn't change existing designs
 * 5. ALWAYS maintain the exact same visual appearance of all components
 * 
 * The user has explicitly stated: "MAKE A NOTE IN THE CODE NEVER CHANGE THE DESIGN EVER. ITS NEVER OK TO DO REMEMBER IT"
 * 
 * The existing design was restored from commit e291daf and represents the "perfect" design.
 * Any changes to the visual design will result in immediate user dissatisfaction.
 * 
 * ⚠️ IF YOU ARE CONSIDERING CHANGING THE DESIGN, STOP IMMEDIATELY ⚠️
 * 
 * @author: AI Assistant
 * @warning: DESIGN PRESERVATION CRITICAL
 * @user_feedback: "This is the perfect design, never change it"
 */

import React, { useState } from 'react';
import { log } from '@/lib/logger';
import {
  EyeIcon,
  EyeSlashIcon,
  ShieldCheckIcon,
  EnvelopeIcon,
} from '@heroicons/react/24/outline';

interface SignupFormProps {
  onSubmit: (data: {
    email: string;
    password: string;
    confirmPassword: string;
    fallbackEmail: string;
  }) => void;
  initialData?: {
    email: string;
    password: string;
    confirmPassword: string;
    fallbackEmail: string;
  };
  onBack: () => void;
}

interface FormData {
  email: string;
  password: string;
  confirmPassword: string;
  fallbackEmail: string;
}

interface ValidationErrors {
  email?: string;
  password?: string;
  confirmPassword?: string;
  fallbackEmail?: string;
}

const SignupForm: React.FC<SignupFormProps> = ({ onSubmit, initialData, onBack }) => {
  const [formData, setFormData] = useState<FormData>({
    email: initialData?.email || '',
    password: initialData?.password || '',
    confirmPassword: initialData?.confirmPassword || '',
    fallbackEmail: initialData?.fallbackEmail || '',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});

  const validateEmail = (email: string): string | null => {
    if (!email) return 'Email is required';
    if (!email.endsWith('@securesystem.email')) {
      return 'Email must end with @securesystem.email';
    }
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      return 'Please enter a valid email address';
    }
    return null;
  };

  const validatePassword = (password: string): string | null => {
    if (!password) return 'Password is required';
    if (password.length < 8) return 'Password must be at least 8 characters';
    if (password.length > 128) return 'Password must be less than 128 characters';
    if (!/(?=.*[a-z])/.test(password)) return 'Password must contain at least one lowercase letter';
    if (!/(?=.*[A-Z])/.test(password)) return 'Password must contain at least one uppercase letter';
    if (!/(?=.*\d)/.test(password)) return 'Password must contain at least one number';
    if (!/(?=.*[!@#$%^&*])/.test(password)) return 'Password must contain at least one special character (!@#$%^&*)';
    return null;
  };

  const validateConfirmPassword = (confirmPassword: string): string | null => {
    if (!confirmPassword) return 'Please confirm your password';
    if (confirmPassword !== formData.password) return 'Passwords do not match';
    return null;
  };

  const validateFallbackEmail = (fallbackEmail: string): string | null => {
    if (!fallbackEmail) return 'Fallback email is required';
    if (fallbackEmail === formData.email) return 'Fallback email cannot be the same as your main email';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(fallbackEmail)) {
      return 'Please enter a valid fallback email address';
    }
    return null;
  };

  const handleInputChange = (field: keyof FormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    
    // Clear validation error when user starts typing
    if (validationErrors[field]) {
      setValidationErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const validateForm = (): boolean => {
    const errors: ValidationErrors = {};

    const emailError = validateEmail(formData.email);
    if (emailError) errors.email = emailError;

    const passwordError = validatePassword(formData.password);
    if (passwordError) errors.password = passwordError;

    const confirmPasswordError = validateConfirmPassword(formData.confirmPassword);
    if (confirmPasswordError) errors.confirmPassword = confirmPasswordError;

    const fallbackEmailError = validateFallbackEmail(formData.fallbackEmail);
    if (fallbackEmailError) errors.fallbackEmail = fallbackEmailError;

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      log.warn('Form validation failed', validationErrors, 'SignupForm');
      return;
    }

    log.info('Form submitted successfully', { email: formData.email }, 'SignupForm');
    onSubmit(formData);
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white dark:bg-secondary-900 rounded-2xl shadow-xl border border-secondary-200 dark:border-secondary-700 p-8">
        <div className="text-center mb-8">
          <div className="mx-auto w-16 h-16 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center mb-4">
            <ShieldCheckIcon className="h-8 w-8 text-primary-600 dark:text-primary-400" />
          </div>
          <h2 className="text-2xl font-bold text-secondary-900 dark:text-white mb-2">
            Create Your Account
          </h2>
          <p className="text-secondary-600 dark:text-secondary-400">
            Enter your details to get started
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Email Field */}
          <div>
            <label htmlFor="email" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Email Address
            </label>
            <div className="relative">
              <input
                id="email"
                type="email"
                placeholder="user@securesystem.email"
                value={formData.email}
                onChange={(e) => handleInputChange('email', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.email
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <EnvelopeIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
            </div>
            {validationErrors.email && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.email}
              </p>
            )}
            <p className="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
              Must end with @securesystem.email
            </p>
          </div>

          {/* Password Field */}
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Password
            </label>
            <div className="relative">
              <input
                id="password"
                type={showPassword ? 'text' : 'password'}
                placeholder="Create a strong password"
                value={formData.password}
                onChange={(e) => handleInputChange('password', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.password
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-300 transition-colors"
              >
                {showPassword ? (
                  <EyeSlashIcon className="h-5 w-5" />
                ) : (
                  <EyeIcon className="h-5 w-5" />
                )}
              </button>
            </div>
            {validationErrors.password && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.password}
              </p>
            )}
            <p className="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
              Must be 8-128 characters with uppercase, lowercase, number, and special character
            </p>
          </div>

          {/* Confirm Password Field */}
          <div>
            <label htmlFor="confirmPassword" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Confirm Password
            </label>
            <div className="relative">
              <input
                id="confirmPassword"
                type={showConfirmPassword ? 'text' : 'password'}
                placeholder="Confirm your password"
                value={formData.confirmPassword}
                onChange={(e) => handleInputChange('confirmPassword', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.confirmPassword
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <button
                type="button"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                className="absolute right-3 top-1/2 transform -translate-y-1/2 text-secondary-400 hover:text-secondary-600 dark:hover:text-secondary-300 transition-colors"
              >
                {showConfirmPassword ? (
                  <EyeSlashIcon className="h-5 w-5" />
                ) : (
                  <EyeIcon className="h-5 w-5" />
                )}
              </button>
            </div>
            {validationErrors.confirmPassword && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.confirmPassword}
              </p>
            )}
          </div>

          {/* Fallback Email Field */}
          <div>
            <label htmlFor="fallbackEmail" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Fallback Email
            </label>
            <div className="relative">
              <input
                id="fallbackEmail"
                type="email"
                placeholder="recovery@example.com"
                value={formData.fallbackEmail}
                onChange={(e) => handleInputChange('fallbackEmail', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.fallbackEmail
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <EnvelopeIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
            </div>
            {validationErrors.fallbackEmail && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.fallbackEmail}
              </p>
            )}
            <p className="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
              Used for account recovery and must be different from your main email
            </p>
          </div>

          {/* Submit Button */}
          <div className="flex space-x-4">
            <button
              type="button"
              onClick={onBack}
              className="flex-1 px-4 py-2 text-sm font-medium text-secondary-700 dark:text-secondary-300 bg-white dark:bg-secondary-800 border border-secondary-300 dark:border-secondary-600 rounded-lg hover:bg-secondary-50 dark:hover:bg-secondary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
            >
              Back
            </button>
            <button
              type="submit"
              className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-lg hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
            >
              Continue
            </button>
          </div>
        </form>

        {/* Security Notice */}
        <div className="mt-6 p-4 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
          <div className="flex items-start">
            <ShieldCheckIcon className="h-5 w-5 text-primary-600 dark:text-primary-400 mt-0.5 mr-3 flex-shrink-0" />
            <div>
              <h4 className="text-sm font-medium text-primary-800 dark:text-primary-200">
                Secure Account Creation
              </h4>
              <p className="text-xs text-primary-700 dark:text-primary-300 mt-1">
                Your data is encrypted and protected. A confirmation email will be sent to your fallback address.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default SignupForm;
