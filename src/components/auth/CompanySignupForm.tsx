/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE COMPANY SIGNUP FORM COMPONENT.
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
  BuildingOfficeIcon,
  GlobeAltIcon,
  UserIcon,
} from '@heroicons/react/24/outline';

interface CompanySignupFormProps {
  onSubmit: (data: {
    companyName: string;
    companyDomain: string;
    adminEmail: string;
    password: string;
    confirmPassword: string;
    fallbackEmail: string;
  }) => void;
  initialData?: {
    companyName?: string;
    companyDomain?: string;
    adminEmail?: string;
    password?: string;
    confirmPassword?: string;
    fallbackEmail?: string;
  };
  onBack: () => void;
}

interface FormData {
  companyName: string;
  companyDomain: string;
  adminEmail: string;
  password: string;
  confirmPassword: string;
  fallbackEmail: string;
}

interface ValidationErrors {
  companyName?: string;
  companyDomain?: string;
  adminEmail?: string;
  password?: string;
  confirmPassword?: string;
  fallbackEmail?: string;
}

const CompanySignupForm: React.FC<CompanySignupFormProps> = ({ onSubmit, initialData, onBack }) => {
  const [formData, setFormData] = useState<FormData>({
    companyName: initialData?.companyName || '',
    companyDomain: initialData?.companyDomain || '',
    adminEmail: initialData?.adminEmail || '',
    password: initialData?.password || '',
    confirmPassword: initialData?.confirmPassword || '',
    fallbackEmail: initialData?.fallbackEmail || '',
  });
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});

  const validateCompanyName = (name: string): string | null => {
    if (!name) return 'Company name is required';
    if (name.length < 2) return 'Company name must be at least 2 characters';
    if (name.length > 100) return 'Company name must be less than 100 characters';
    if (!/^[a-zA-Z0-9\s\-&.]+$/.test(name)) {
      return 'Company name can only contain letters, numbers, spaces, hyphens, ampersands, and periods';
    }
    return null;
  };

  const validateCompanyDomain = (domain: string): string | null => {
    if (!domain) return 'Company domain is required';
    if (!/^[a-zA-Z0-9][a-zA-Z0-9-]{1,61}[a-zA-Z0-9]\.[a-zA-Z]{2,}$/.test(domain)) {
      return 'Please enter a valid domain (e.g., example.com)';
    }
    return null;
  };

  const validateAdminEmail = (email: string): string | null => {
    if (!email) return 'Admin email is required';
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      return 'Please enter a valid email address';
    }
    // Check if admin email matches company domain
    if (formData.companyDomain && !email.endsWith(`@${formData.companyDomain}`)) {
      return `Admin email should match your company domain (@${formData.companyDomain})`;
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
    if (fallbackEmail === formData.adminEmail) return 'Fallback email cannot be the same as your admin email';
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

    const companyNameError = validateCompanyName(formData.companyName);
    if (companyNameError) errors.companyName = companyNameError;

    const companyDomainError = validateCompanyDomain(formData.companyDomain);
    if (companyDomainError) errors.companyDomain = companyDomainError;

    const adminEmailError = validateAdminEmail(formData.adminEmail);
    if (adminEmailError) errors.adminEmail = adminEmailError;

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
      log.warn('Company form validation failed', validationErrors, 'CompanySignupForm');
      return;
    }

    log.info('Company form submitted successfully', { 
      companyName: formData.companyName,
      adminEmail: formData.adminEmail 
    }, 'CompanySignupForm');
    onSubmit(formData);
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white dark:bg-secondary-900 rounded-2xl shadow-xl border border-secondary-200 dark:border-secondary-700 p-8">
        <div className="text-center mb-8">
          <div className="mx-auto w-16 h-16 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center mb-4">
            <BuildingOfficeIcon className="h-8 w-8 text-primary-600 dark:text-primary-400" />
          </div>
          <h2 className="text-2xl font-bold text-secondary-900 dark:text-white mb-2">
            Company Account Setup
          </h2>
          <p className="text-secondary-600 dark:text-secondary-400">
            Create your organization&apos;s secure email account
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Company Name Field */}
          <div>
            <label htmlFor="companyName" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Company Name
            </label>
            <div className="relative">
              <input
                id="companyName"
                type="text"
                placeholder="Your Company Inc."
                value={formData.companyName}
                onChange={(e) => handleInputChange('companyName', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.companyName
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <BuildingOfficeIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
            </div>
            {validationErrors.companyName && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.companyName}
              </p>
            )}
          </div>

          {/* Company Domain Field */}
          <div>
            <label htmlFor="companyDomain" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Company Domain
            </label>
            <div className="relative">
              <input
                id="companyDomain"
                type="text"
                placeholder="example.com"
                value={formData.companyDomain}
                onChange={(e) => handleInputChange('companyDomain', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.companyDomain
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <GlobeAltIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
            </div>
            {validationErrors.companyDomain && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.companyDomain}
              </p>
            )}
            <p className="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
              Your company&apos;s domain (e.g., example.com)
            </p>
          </div>

          {/* Admin Email Field */}
          <div>
            <label htmlFor="adminEmail" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Admin Email
            </label>
            <div className="relative">
              <input
                id="adminEmail"
                type="email"
                placeholder="admin@example.com"
                value={formData.adminEmail}
                onChange={(e) => handleInputChange('adminEmail', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.adminEmail
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <UserIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
            </div>
            {validationErrors.adminEmail && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.adminEmail}
              </p>
            )}
            <p className="mt-1 text-xs text-secondary-500 dark:text-secondary-400">
              Should match your company domain
            </p>
          </div>

          {/* Password Field */}
          <div>
            <label htmlFor="password" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Admin Password
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
              Used for account recovery and must be different from your admin email
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
                Enterprise Security
              </h4>
              <p className="text-xs text-primary-700 dark:text-primary-300 mt-1">
                Your company data is encrypted and protected. You&apos;ll be able to invite team members after setup.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default CompanySignupForm;
