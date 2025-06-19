import React, { useEffect, useState, useRef } from 'react';
import { gsap } from 'gsap';
import { XMarkIcon } from '@heroicons/react/24/outline';

const OnboardingModal = () => {
  const [isOpen, setIsOpen] = useState(false);
  const modalRef = useRef(null);
  const overlayRef = useRef(null);

  useEffect(() => {
    // Check if onboarding has been seen
    const onboardingSeen = localStorage.getItem('onboarding_seen');
    if (!onboardingSeen) {
      setIsOpen(true);
      
      // Animate modal slide-up with GSAP
      gsap.fromTo(modalRef.current, 
        { y: 50, opacity: 0 }, 
        { y: 0, opacity: 1, duration: 0.4, ease: "power2.out" }
      );
      
      // Animate overlay fade-in
      gsap.fromTo(overlayRef.current,
        { opacity: 0 },
        { opacity: 1, duration: 0.3, ease: "power2.out" }
      );
    }
  }, []);

  const handleClose = () => {
    // Animate modal slide-down
    gsap.to(modalRef.current, {
      y: 50,
      opacity: 0,
      duration: 0.3,
      ease: "power2.in",
      onComplete: () => {
        localStorage.setItem('onboarding_seen', 'true');
        setIsOpen(false);
      }
    });
    
    // Animate overlay fade-out
    gsap.to(overlayRef.current, {
      opacity: 0,
      duration: 0.3,
      ease: "power2.in"
    });
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Escape') {
      handleClose();
    }
  };

  // Focus trap for accessibility
  useEffect(() => {
    if (isOpen) {
      const focusableElements = modalRef.current?.querySelectorAll(
        'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
      );
      
      if (focusableElements && focusableElements.length > 0) {
        focusableElements[0].focus();
      }
      
      // Add event listeners
      document.addEventListener('keydown', handleKeyDown);
      document.body.style.overflow = 'hidden';
      
      return () => {
        document.removeEventListener('keydown', handleKeyDown);
        document.body.style.overflow = 'unset';
      };
    }
  }, [isOpen]);

  if (!isOpen) return null;

  return (
    <div 
      ref={overlayRef}
      className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50 p-4"
      role="dialog"
      aria-modal="true"
      aria-labelledby="modal-title"
      aria-describedby="modal-description"
    >
      <div
        ref={modalRef}
        className="glassmorphic p-8 rounded-xl max-w-md w-full relative"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Close button */}
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-primary dark:text-gray-300 hover:text-accent dark:hover:text-accent transition-colors duration-200 p-1 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700"
          aria-label="Close modal"
        >
          <XMarkIcon className="h-6 w-6" />
        </button>

        {/* Modal content */}
        <div className="text-center">
          {/* Icon */}
          <div className="mb-6">
            <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary to-accent rounded-full flex items-center justify-center">
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
          <p 
            id="modal-description"
            className="text-text-gray dark:text-gray-300 mb-6 leading-relaxed"
          >
            Your emails are encrypted with <strong>AES-256-GCM</strong> for unbreakable privacy. 
            Every message is protected with military-grade encryption, ensuring your communications 
            remain secure and private.
          </p>

          {/* Features list */}
          <div className="mb-8 text-left">
            <div className="space-y-3">
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
          </div>

          {/* Action button */}
          <button
            onClick={handleClose}
            className="btn-accent w-full"
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