/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE PAYMENT FORM COMPONENT.
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
  ShieldCheckIcon,
  CreditCardIcon,
  LockClosedIcon,
} from '@heroicons/react/24/outline';

interface PaymentFormProps {
  plan: string;
  onSubmit: (paymentData: {
    method: string;
    cardNumber: string;
    expiryMonth: string;
    expiryYear: string;
    cvv: string;
    name: string;
  }) => void;
  onBack: () => void;
}

interface FormData {
  cardNumber: string;
  expiryMonth: string;
  expiryYear: string;
  cvv: string;
  name: string;
}

interface ValidationErrors {
  cardNumber?: string;
  expiryMonth?: string;
  expiryYear?: string;
  cvv?: string;
  name?: string;
}

const PaymentForm: React.FC<PaymentFormProps> = ({ plan, onSubmit, onBack }) => {
  const [formData, setFormData] = useState<FormData>({
    cardNumber: '',
    expiryMonth: '',
    expiryYear: '',
    cvv: '',
    name: '',
  });
  const [validationErrors, setValidationErrors] = useState<ValidationErrors>({});
  const [isProcessing, setIsProcessing] = useState(false);

  const getPlanPrice = (): string => {
    const planPrices: Record<string, string> = {
      'pro-monthly': '$9.99/month',
      'pro-yearly': '$95.88/year',
      'enterprise-starter': '$29.99/month',
      'enterprise-pro': '$79.99/month',
      'enterprise-enterprise': 'Contact Sales',
    };
    return planPrices[plan] || 'Varies';
  };

  const validateCardNumber = (cardNumber: string): string | null => {
    if (!cardNumber) return 'Card number is required';
    const cleaned = cardNumber.replace(/\s/g, '');
    if (!/^\d{13,19}$/.test(cleaned)) return 'Please enter a valid card number';
    return null;
  };

  const validateExpiryMonth = (month: string): string | null => {
    if (!month) return 'Expiry month is required';
    const monthNum = parseInt(month, 10);
    if (monthNum < 1 || monthNum > 12) return 'Please enter a valid month (1-12)';
    return null;
  };

  const validateExpiryYear = (year: string): string | null => {
    if (!year) return 'Expiry year is required';
    const yearNum = parseInt(year, 10);
    const currentYear = new Date().getFullYear();
    if (yearNum < currentYear || yearNum > currentYear + 20) {
      return `Please enter a valid year (${currentYear}-${currentYear + 20})`;
    }
    return null;
  };

  const validateCvv = (cvv: string): string | null => {
    if (!cvv) return 'CVV is required';
    if (!/^\d{3,4}$/.test(cvv)) return 'Please enter a valid CVV (3-4 digits)';
    return null;
  };

  const validateName = (name: string): string | null => {
    if (!name) return 'Cardholder name is required';
    if (name.length < 2) return 'Please enter a valid name';
    return null;
  };

  const handleInputChange = (field: keyof FormData, value: string) => {
    let processedValue = value;

    // Format card number with spaces
    if (field === 'cardNumber') {
      const cleaned = value.replace(/\s/g, '');
      processedValue = cleaned.replace(/(\d{4})/g, '$1 ').trim();
    }

    setFormData(prev => ({ ...prev, [field]: processedValue }));
    
    // Clear validation error when user starts typing
    if (validationErrors[field]) {
      setValidationErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const validateForm = (): boolean => {
    const errors: ValidationErrors = {};

    const cardNumberError = validateCardNumber(formData.cardNumber);
    if (cardNumberError) errors.cardNumber = cardNumberError;

    const expiryMonthError = validateExpiryMonth(formData.expiryMonth);
    if (expiryMonthError) errors.expiryMonth = expiryMonthError;

    const expiryYearError = validateExpiryYear(formData.expiryYear);
    if (expiryYearError) errors.expiryYear = expiryYearError;

    const cvvError = validateCvv(formData.cvv);
    if (cvvError) errors.cvv = cvvError;

    const nameError = validateName(formData.name);
    if (nameError) errors.name = nameError;

    setValidationErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!validateForm()) {
      log.warn('Payment form validation failed', validationErrors, 'PaymentForm');
      return;
    }

    setIsProcessing(true);
    log.info('Payment form submitted', { plan }, 'PaymentForm');

    try {
      // Simulate payment processing
      await new Promise(resolve => setTimeout(resolve, 2000));
      
      onSubmit({
        method: 'credit_card',
        ...formData
      });
    } catch (error) {
      log.error('Payment processing failed', error, 'PaymentForm');
    } finally {
      setIsProcessing(false);
    }
  };

  return (
    <div className="w-full max-w-md mx-auto">
      <div className="bg-white dark:bg-secondary-900 rounded-2xl shadow-xl border border-secondary-200 dark:border-secondary-700 p-8">
        <div className="text-center mb-8">
          <div className="mx-auto w-16 h-16 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center mb-4">
            <CreditCardIcon className="h-8 w-8 text-primary-600 dark:text-primary-400" />
          </div>
          <h2 className="text-2xl font-bold text-secondary-900 dark:text-white mb-2">
            Payment Information
          </h2>
          <p className="text-secondary-600 dark:text-secondary-400">
            Complete your {plan} subscription
          </p>
          <div className="mt-4 p-3 bg-secondary-50 dark:bg-secondary-800 rounded-lg">
            <p className="text-sm text-secondary-700 dark:text-secondary-300">
              <span className="font-medium">Plan:</span> {plan.replace('-', ' ').replace(/\b\w/g, l => l.toUpperCase())}
            </p>
            <p className="text-sm text-secondary-700 dark:text-secondary-300">
              <span className="font-medium">Price:</span> {getPlanPrice()}
            </p>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Card Number Field */}
          <div>
            <label htmlFor="cardNumber" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Card Number
            </label>
            <div className="relative">
              <input
                id="cardNumber"
                type="text"
                placeholder="1234 5678 9012 3456"
                value={formData.cardNumber}
                onChange={(e) => handleInputChange('cardNumber', e.target.value)}
                maxLength={19}
                className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                  validationErrors.cardNumber
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              />
              <CreditCardIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
            </div>
            {validationErrors.cardNumber && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.cardNumber}
              </p>
            )}
          </div>

          {/* Expiry Date and CVV */}
          <div className="grid grid-cols-3 gap-4">
            <div>
              <label htmlFor="expiryMonth" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                Month
              </label>
              <select
                id="expiryMonth"
                value={formData.expiryMonth}
                onChange={(e) => handleInputChange('expiryMonth', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 transition-colors ${
                  validationErrors.expiryMonth
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              >
                <option value="">MM</option>
                {Array.from({ length: 12 }, (_, i) => i + 1).map(month => (
                  <option key={month} value={month.toString().padStart(2, '0')}>
                    {month.toString().padStart(2, '0')}
                  </option>
                ))}
              </select>
              {validationErrors.expiryMonth && (
                <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                  {validationErrors.expiryMonth}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="expiryYear" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                Year
              </label>
              <select
                id="expiryYear"
                value={formData.expiryYear}
                onChange={(e) => handleInputChange('expiryYear', e.target.value)}
                className={`block w-full rounded-lg border px-3 py-2 text-sm focus:outline-none focus:ring-1 transition-colors ${
                  validationErrors.expiryYear
                    ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                    : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                }`}
                required
              >
                <option value="">YYYY</option>
                {Array.from({ length: 21 }, (_, i) => new Date().getFullYear() + i).map(year => (
                  <option key={year} value={year}>
                    {year}
                  </option>
                ))}
              </select>
              {validationErrors.expiryYear && (
                <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                  {validationErrors.expiryYear}
                </p>
              )}
            </div>

            <div>
              <label htmlFor="cvv" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                CVV
              </label>
              <div className="relative">
                <input
                  id="cvv"
                  type="text"
                  placeholder="123"
                  value={formData.cvv}
                  onChange={(e) => handleInputChange('cvv', e.target.value)}
                  maxLength={4}
                  className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors pr-10 ${
                    validationErrors.cvv
                      ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                      : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
                  }`}
                  required
                />
                <LockClosedIcon className="absolute right-3 top-1/2 transform -translate-y-1/2 h-5 w-5 text-secondary-400" />
              </div>
              {validationErrors.cvv && (
                <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                  {validationErrors.cvv}
                </p>
              )}
            </div>
          </div>

          {/* Cardholder Name */}
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Cardholder Name
            </label>
            <input
              id="name"
              type="text"
              placeholder="John Doe"
              value={formData.name}
              onChange={(e) => handleInputChange('name', e.target.value)}
              className={`block w-full rounded-lg border px-3 py-2 text-sm placeholder-secondary-500 focus:outline-none focus:ring-1 transition-colors ${
                validationErrors.name
                  ? 'border-error-300 bg-error-50 text-error-900 focus:border-error-500 focus:ring-error-500 dark:border-error-600 dark:bg-error-900/20 dark:text-error-200'
                  : 'border-secondary-300 bg-white text-secondary-900 focus:border-primary-500 focus:ring-primary-500 dark:border-secondary-600 dark:bg-secondary-800 dark:text-white'
              }`}
              required
            />
            {validationErrors.name && (
              <p className="mt-1 text-sm text-error-600 dark:text-error-400">
                {validationErrors.name}
              </p>
            )}
          </div>

          {/* Submit Button */}
          <div className="flex space-x-4">
            <button
              type="button"
              onClick={onBack}
              disabled={isProcessing}
              className="flex-1 px-4 py-2 text-sm font-medium text-secondary-700 dark:text-secondary-300 bg-white dark:bg-secondary-800 border border-secondary-300 dark:border-secondary-600 rounded-lg hover:bg-secondary-50 dark:hover:bg-secondary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors disabled:opacity-50"
            >
              Back
            </button>
            <button
              type="submit"
              disabled={isProcessing}
              className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-lg hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              {isProcessing ? (
                <>
                  <svg
                    className="animate-spin -ml-1 mr-2 h-4 w-4"
                    xmlns="http://www.w3.org/2000/svg"
                    fill="none"
                    viewBox="0 0 24 24"
                  >
                    <circle
                      className="opacity-25"
                      cx="12"
                      cy="12"
                      r="10"
                      stroke="currentColor"
                      strokeWidth="4"
                    />
                    <path
                      className="opacity-75"
                      fill="currentColor"
                      d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                    />
                  </svg>
                  Processing...
                </>
              ) : (
                'Complete Payment'
              )}
            </button>
          </div>
        </form>

        {/* Security Notice */}
        <div className="mt-6 p-4 bg-primary-50 dark:bg-primary-900/20 rounded-lg">
          <div className="flex items-start">
            <ShieldCheckIcon className="h-5 w-5 text-primary-600 dark:text-primary-400 mt-0.5 mr-3 flex-shrink-0" />
            <div>
              <h4 className="text-sm font-medium text-primary-800 dark:text-primary-200">
                Secure Payment
              </h4>
              <p className="text-xs text-primary-700 dark:text-primary-300 mt-1">
                Your payment information is encrypted and secure. We use industry-standard SSL encryption.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default PaymentForm;
