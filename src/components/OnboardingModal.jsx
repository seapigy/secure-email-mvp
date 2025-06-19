import React, { useEffect, useState } from 'react';
import { gsap } from 'gsap';
import { XMarkIcon } from '@heroicons/react/24/outline';

const OnboardingModal = () => {
  const [isOpen, setIsOpen] = useState(false);

  useEffect(() => {
    if (!localStorage.getItem('onboarding_seen')) {
      setIsOpen(true);
      gsap.fromTo('.modal', { y: 50, opacity: 0 }, { y: 0, opacity: 1, duration: 0.35 });
    }
  }, []);

  const handleClose = () => {
    localStorage.setItem('onboarding_seen', 'true');
    setIsOpen(false);
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50 p-4">
      <div
        className="modal bg-white dark:bg-gray-800 bg-opacity-85 backdrop-blur-[10px] p-8 rounded-xl max-w-md w-full relative"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        {/* Close Button */}
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-primary hover:text-accent transition-colors duration-200 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700"
          aria-label="Close modal"
        >
          <XMarkIcon className="h-6 w-6" />
        </button>

        {/* Modal Content */}
        <div className="text-center">
          {/* Icon */}
          <div className="mb-6">
            <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary to-accent rounded-full flex items-center justify-center shadow-lg">
              <svg 
                className="w-8 h-8 text-white" 
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
                aria-hidden="true"
              >
                <path 
                  strokeLinecap="round" 
                  strokeLinejoin="round" 
                  strokeWidth={2} 
                  d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" 
                />
              </svg>
            </div>
          </div>

          {/* Title */}
          <h2 
            id="modal-title" 
            className="text-2xl font-bold text-text-gray dark:text-white mb-4"
          >
            Welcome to Secure Email!
          </h2>

          {/* Description */}
          <p className="text-text-gray dark:text-gray-300 mb-6 leading-relaxed">
            Your emails are encrypted with <strong>AES-256-GCM</strong> and auto-deleted after 3 failed access attempts for maximum security.
          </p>

          {/* Features List */}
          <div className="mb-8 text-left space-y-3">
            <div className="flex items-start">
              <div className="flex-shrink-0 w-5 h-5 bg-success rounded-full flex items-center justify-center mr-3 mt-0.5">
                <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
              </div>
              <span className="text-sm text-text-gray dark:text-gray-300">
                End-to-end encryption with AES-256-GCM
              </span>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-5 h-5 bg-success rounded-full flex items-center justify-center mr-3 mt-0.5">
                <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
              </div>
              <span className="text-sm text-text-gray dark:text-gray-300">
                Two-factor authentication with TOTP
              </span>
            </div>
            <div className="flex items-start">
              <div className="flex-shrink-0 w-5 h-5 bg-success rounded-full flex items-center justify-center mr-3 mt-0.5">
                <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
              </div>
              <span className="text-sm text-text-gray dark:text-gray-300">
                Secure link-based delivery system
              </span>
            </div>
          </div>

          {/* Action Button */}
          <button
            onClick={handleClose}
            className="w-full p-4 bg-accent text-white rounded-lg font-semibold hover:scale-105 transition-all duration-150 shadow-lg hover:shadow-xl"
            aria-label="Get started with Secure Email"
          >
            Get Started
          </button>
        </div>
      </div>
    </div>
  );
};

export default OnboardingModal; 