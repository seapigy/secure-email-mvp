/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE MAIN SIGNUP PAGE COMPONENT.
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
import { useNavigate } from 'react-router-dom';
import { log } from '@/lib/logger';
import AccountTypeSelector from '@/components/auth/AccountTypeSelector';
import SignupForm from '@/components/auth/SignupForm';
import CompanySignupForm from '@/components/auth/CompanySignupForm';
import PlanSelector from '@/components/auth/PlanSelector';
import PaymentForm from '@/components/auth/PaymentForm';

export type AccountType = 'free' | 'paid' | 'company';
export type SignupStep = 'account-type' | 'basic-info' | 'plan-selection' | 'payment' | 'company-info' | 'confirmation';

interface SignupData {
  accountType: AccountType;
  email: string;
  password: string;
  confirmPassword: string;
  fallbackEmail: string;
  companyName?: string;
  companyDomain?: string;
  selectedPlan?: string;
  paymentMethod?: string;
}

const SignupPage: React.FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState<SignupStep>('account-type');
  const [signupData, setSignupData] = useState<SignupData>({
    accountType: 'free',
    email: '',
    password: '',
    confirmPassword: '',
    fallbackEmail: '',
  });
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleAccountTypeSelect = (accountType: AccountType) => {
    log.info('Account type selected', { accountType }, 'SignupPage');
    setSignupData(prev => ({ ...prev, accountType }));
    
    if (accountType === 'free') {
      setCurrentStep('basic-info');
    } else if (accountType === 'paid') {
      setCurrentStep('plan-selection');
    } else {
      setCurrentStep('company-info');
    }
  };

  const handleBasicInfoSubmit = (data: Partial<SignupData>) => {
    log.info('Basic info submitted', { email: data.email }, 'SignupPage');
    setSignupData(prev => ({ ...prev, ...data }));
    
    if (signupData.accountType === 'free') {
      setCurrentStep('confirmation');
    } else if (signupData.accountType === 'paid') {
      setCurrentStep('plan-selection');
    }
  };

  const handlePlanSelection = (plan: string) => {
    log.info('Plan selected', { plan }, 'SignupPage');
    setSignupData(prev => ({ ...prev, selectedPlan: plan }));
    setCurrentStep('payment');
  };

  const handlePaymentSubmit = (paymentData: any) => {
    log.info('Payment submitted', { plan: signupData.selectedPlan }, 'SignupPage');
    setSignupData(prev => ({ ...prev, paymentMethod: paymentData.method }));
    setCurrentStep('confirmation');
  };

  const handleCompanyInfoSubmit = (data: Partial<SignupData>) => {
    log.info('Company info submitted', { companyName: data.companyName }, 'SignupPage');
    setSignupData(prev => ({ ...prev, ...data }));
    setCurrentStep('plan-selection');
  };

  const handleFinalSubmit = async () => {
    setIsLoading(true);
    setError(null);

    try {
      log.info('Final signup submission', { accountType: signupData.accountType }, 'SignupPage');
      
      // Call appropriate signup endpoint based on account type
      if (signupData.accountType === 'company') {
        // TODO: Implement company signup endpoint
        log.warn('Company signup endpoint not implemented yet', null, 'SignupPage');
      } else {
        // Use existing signup endpoint
        const response = await fetch('/api/auth/signup', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            email: signupData.email,
            password: signupData.password,
            fallback_email: signupData.fallbackEmail,
          }),
        });

        if (!response.ok) {
          const errorData = await response.json();
          throw new Error(errorData.error || 'Signup failed');
        }
      }

      log.info('Signup successful', { email: signupData.email }, 'SignupPage');
      
      // Redirect to login with success message
      navigate('/login', { 
        state: { 
          message: 'Account created successfully! Please check your fallback email for confirmation.' 
        } 
      });
    } catch (err) {
      log.error('Signup failed', err, 'SignupPage');
      setError(err instanceof Error ? err.message : 'Signup failed');
    } finally {
      setIsLoading(false);
    }
  };

  const goBack = () => {
    log.info('Going back in signup flow', { currentStep }, 'SignupPage');
    
    switch (currentStep) {
      case 'basic-info':
        setCurrentStep('account-type');
        break;
      case 'plan-selection':
        if (signupData.accountType === 'company') {
          setCurrentStep('company-info');
        } else {
          setCurrentStep('basic-info');
        }
        break;
      case 'payment':
        setCurrentStep('plan-selection');
        break;
      case 'confirmation':
        if (signupData.accountType === 'free') {
          setCurrentStep('basic-info');
        } else {
          setCurrentStep('payment');
        }
        break;
      default:
        setCurrentStep('account-type');
    }
  };

  const renderStep = () => {
    switch (currentStep) {
      case 'account-type':
        return (
          <AccountTypeSelector
            onSelect={handleAccountTypeSelect}
            selectedType={signupData.accountType}
          />
        );

      case 'basic-info':
        return (
          <SignupForm
            onSubmit={handleBasicInfoSubmit}
            initialData={signupData}
            onBack={goBack}
          />
        );

      case 'company-info':
        return (
          <CompanySignupForm
            onSubmit={handleCompanyInfoSubmit}
            initialData={signupData}
            onBack={goBack}
          />
        );

      case 'plan-selection':
        return (
          <PlanSelector
            accountType={signupData.accountType}
            onSelect={handlePlanSelection}
            selectedPlan={signupData.selectedPlan}
            onBack={goBack}
          />
        );

      case 'payment':
        return (
          <PaymentForm
            plan={signupData.selectedPlan!}
            onSubmit={handlePaymentSubmit}
            onBack={goBack}
          />
        );

      case 'confirmation':
        return (
          <div className="w-full max-w-md mx-auto">
            <div className="bg-white dark:bg-secondary-900 rounded-2xl shadow-xl border border-secondary-200 dark:border-secondary-700 p-8">
              <div className="text-center mb-8">
                <div className="mx-auto w-16 h-16 bg-green-100 dark:bg-green-900/20 rounded-full flex items-center justify-center mb-4">
                  <svg className="h-8 w-8 text-green-600 dark:text-green-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
                  </svg>
                </div>
                <h2 className="text-2xl font-bold text-secondary-900 dark:text-white mb-2">
                  Review Your Account
                </h2>
                <p className="text-secondary-600 dark:text-secondary-400">
                  Please review your information before creating your account
                </p>
              </div>

              <div className="space-y-4 mb-8">
                <div className="bg-secondary-50 dark:bg-secondary-800 rounded-lg p-4">
                  <h3 className="font-medium text-secondary-900 dark:text-white mb-2">Account Details</h3>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-secondary-600 dark:text-secondary-400">Account Type:</span>
                      <span className="text-secondary-900 dark:text-white capitalize">{signupData.accountType}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-secondary-600 dark:text-secondary-400">Email:</span>
                      <span className="text-secondary-900 dark:text-white">{signupData.email}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-secondary-600 dark:text-secondary-400">Fallback Email:</span>
                      <span className="text-secondary-900 dark:text-white">{signupData.fallbackEmail}</span>
                    </div>
                    {signupData.companyName && (
                      <div className="flex justify-between">
                        <span className="text-secondary-600 dark:text-secondary-400">Company:</span>
                        <span className="text-secondary-900 dark:text-white">{signupData.companyName}</span>
                      </div>
                    )}
                    {signupData.selectedPlan && (
                      <div className="flex justify-between">
                        <span className="text-secondary-600 dark:text-secondary-400">Plan:</span>
                        <span className="text-secondary-900 dark:text-white">{signupData.selectedPlan}</span>
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {error && (
                <div className="bg-error-50 dark:bg-error-900/20 border border-error-200 dark:border-error-800 rounded-lg p-4 mb-6">
                  <div className="flex">
                    <div className="flex-shrink-0">
                      <svg className="h-5 w-5 text-error-400" viewBox="0 0 20 20" fill="currentColor">
                        <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                      </svg>
                    </div>
                    <div className="ml-3">
                      <p className="text-sm text-error-800 dark:text-error-200">
                        {error}
                      </p>
                    </div>
                  </div>
                </div>
              )}

              <div className="flex space-x-4">
                <button
                  type="button"
                  onClick={goBack}
                  className="flex-1 px-4 py-2 text-sm font-medium text-secondary-700 dark:text-secondary-300 bg-white dark:bg-secondary-800 border border-secondary-300 dark:border-secondary-600 rounded-lg hover:bg-secondary-50 dark:hover:bg-secondary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 transition-colors"
                >
                  Back
                </button>
                <button
                  type="button"
                  onClick={handleFinalSubmit}
                  disabled={isLoading}
                  className="flex-1 px-4 py-2 text-sm font-medium text-white bg-primary-600 border border-transparent rounded-lg hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {isLoading ? (
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
                      Creating Account...
                    </>
                  ) : (
                    'Create Account'
                  )}
                </button>
              </div>
            </div>
          </div>
        );

      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-secondary-50 dark:from-secondary-900 dark:via-secondary-800 dark:to-secondary-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="mx-auto w-20 h-20 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center mb-6">
            <svg className="h-10 w-10 text-primary-600 dark:text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
            Join SecureMail
          </h1>
          <p className="text-secondary-600 dark:text-secondary-400">
            Create your secure email account
          </p>
        </div>
        
        {renderStep()}

        <div className="mt-6 text-center">
          <p className="text-sm text-secondary-600 dark:text-secondary-400">
            Already have an account?{' '}
            <button
              type="button"
              onClick={() => navigate('/login')}
              className="font-medium text-primary-600 hover:text-primary-500 dark:text-primary-400 dark:hover:text-primary-300 transition-colors"
            >
              Sign in
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default SignupPage;
