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
    <div className="fixed inset-0 flex items-center justify-center bg-black bg-opacity-50 z-50">
      <div
        className="modal bg-white dark:bg-gray-800 bg-opacity-85 backdrop-blur-[10px] p-8 rounded-xl max-w-md w-full"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-primary"
          aria-label="Close modal"
        >
          <XMarkIcon className="h-6 w-6" />
        </button>
        <h2 id="modal-title" className="text-2xl font-bold text-text-gray dark:text-white mb-4">
          Welcome!
        </h2>
        <p className="text-text-gray dark:text-gray-300 mb-6">
          Your emails are encrypted with AES-256-GCM and auto-deleted after 3 failed access attempts.
        </p>
        <button
          onClick={handleClose}
          className="w-full p-3 bg-accent text-white rounded-md hover:scale-105 transition-transform duration-150"
        >
          Get Started
        </button>
      </div>
    </div>
  );
};

export default OnboardingModal; 