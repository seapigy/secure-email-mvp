/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE ACCOUNT TYPE SELECTOR COMPONENT.
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

import React from 'react';
import { log } from '@/lib/logger';
import type { AccountType } from '@/pages/SignupPage';

interface AccountTypeSelectorProps {
  onSelect: (accountType: AccountType) => void;
  selectedType: AccountType;
}

interface AccountTypeOption {
  type: AccountType;
  title: string;
  description: string;
  features: string[];
  price: string;
  icon: React.ReactNode;
  popular?: boolean;
}

const AccountTypeSelector: React.FC<AccountTypeSelectorProps> = ({ onSelect, selectedType }) => {
  const accountTypes: AccountTypeOption[] = [
    {
      type: 'free',
      title: 'Free Plan',
      description: 'Perfect for personal use and getting started',
      features: [
        '1GB storage',
        'Basic email encryption',
        'Standard support',
        'Up to 100 emails/month'
      ],
      price: 'Free',
      icon: (
        <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
        </svg>
      )
    },
    {
      type: 'paid',
      title: 'Pro Plan',
      description: 'Advanced features for power users and professionals',
      features: [
        '10GB storage',
        'Advanced encryption',
        'Priority support',
        'Unlimited emails',
        'Custom domains',
        'Advanced security features'
      ],
      price: '$9.99/month',
      icon: (
        <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
      ),
      popular: true
    },
    {
      type: 'company',
      title: 'Enterprise Plan',
      description: 'Complete solution for teams and organizations',
      features: [
        'Unlimited storage',
        'Enterprise encryption',
        '24/7 support',
        'Team management',
        'Custom branding',
        'Advanced analytics',
        'API access',
        'SSO integration'
      ],
      price: '$29.99/month',
      icon: (
        <svg className="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 21V5a2 2 0 00-2-2H7a2 2 0 00-2 2v16m14 0h2m-2 0h-5m-9 0H3m2 0h5M9 7h1m-1 4h1m4-4h1m-1 4h1m-5 10v-5a1 1 0 011-1h2a1 1 0 011 1v5m-4 0h4" />
        </svg>
      )
    }
  ];

  const handleSelect = (accountType: AccountType) => {
    log.info('Account type selected', { accountType }, 'AccountTypeSelector');
    onSelect(accountType);
  };

  return (
    <div className="w-full max-w-8xl mx-auto px-4">
      <div className="bg-white dark:bg-secondary-900 rounded-2xl shadow-xl border border-secondary-200 dark:border-secondary-700 p-8">
        <div className="text-center mb-8">
          <div className="mx-auto w-16 h-16 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center mb-4">
            <svg className="h-8 w-8 text-primary-600 dark:text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-secondary-900 dark:text-white mb-2">
            Choose Your Plan
          </h2>
          <p className="text-secondary-600 dark:text-secondary-400">
            Select the plan that best fits your needs
          </p>
        </div>

        <div className="flex flex-col lg:flex-row gap-4">
          {accountTypes.map((option) => (
                         <div
               key={option.type}
                               className={`relative p-4 border-2 rounded-xl cursor-pointer transition-all duration-200 flex-1 flex flex-col min-w-[250px] ${
                 selectedType === option.type
                   ? 'border-primary-500 bg-primary-50 dark:bg-primary-900/20'
                   : 'border-secondary-200 dark:border-secondary-700 hover:border-primary-300 dark:hover:border-primary-600'
               }`}
               onClick={() => handleSelect(option.type)}
             >
              {option.popular && (
                <div className="absolute -top-3 left-1/2 transform -translate-x-1/2">
                  <span className="bg-primary-600 text-white text-xs font-medium px-3 py-1 rounded-full">
                    Most Popular
                  </span>
                </div>
              )}

                             <div className="text-center mb-4">
                 <div className={`inline-flex p-3 rounded-lg mb-3 ${
                   selectedType === option.type
                     ? 'bg-primary-100 dark:bg-primary-900/30 text-primary-600 dark:text-primary-400'
                     : 'bg-secondary-100 dark:bg-secondary-800 text-secondary-600 dark:text-secondary-400'
                 }`}>
                   {option.icon}
                 </div>
                 <h3 className="text-lg font-semibold text-secondary-900 dark:text-white mb-1">
                   {option.title}
                 </h3>
                 <p className="text-sm text-secondary-600 dark:text-secondary-400 mb-2">
                   {option.description}
                 </p>
                 <div className="text-2xl font-bold text-secondary-900 dark:text-white">
                   {option.price}
                 </div>
               </div>

                             <ul className="space-y-2 flex-1">
                 {option.features.map((feature, index) => (
                   <li key={index} className="flex items-center text-sm text-secondary-700 dark:text-secondary-300">
                     <svg className="h-4 w-4 text-green-500 mr-2 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                       <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                     </svg>
                     {feature}
                   </li>
                 ))}
               </ul>

               <div className="mt-auto pt-4">
                <button
                  type="button"
                  className={`w-full py-2 px-4 rounded-lg font-medium transition-colors ${
                    selectedType === option.type
                      ? 'bg-primary-600 text-white hover:bg-primary-700'
                      : 'bg-secondary-100 dark:bg-secondary-800 text-secondary-700 dark:text-secondary-300 hover:bg-secondary-200 dark:hover:bg-secondary-700'
                  }`}
                >
                  {selectedType === option.type ? 'Selected' : 'Select Plan'}
                </button>
              </div>
            </div>
          ))}
        </div>

        <div className="mt-6 p-4 bg-secondary-50 dark:bg-secondary-800 rounded-lg">
          <div className="flex items-start">
            <svg className="h-5 w-5 text-secondary-500 dark:text-secondary-400 mt-0.5 mr-3 flex-shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            <div>
              <h4 className="text-sm font-medium text-secondary-800 dark:text-secondary-200">
                All plans include
              </h4>
              <p className="text-xs text-secondary-600 dark:text-secondary-400 mt-1">
                End-to-end encryption, secure email delivery, and 30-day money-back guarantee.
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AccountTypeSelector;
