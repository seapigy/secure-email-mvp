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

  const handleKeyDown = (e) => {
    if (e.key === 'Escape') {
      handleClose();
    }
  };

  if (!isOpen) return null;

  return (
    <div 
      className="fixed inset-0 flex items-center justify-center bg-black/50 backdrop-blur-sm z-50"
      onClick={handleClose}
      onKeyDown={handleKeyDown}
      tabIndex={-1}
    >
      <div 
        className="modal bg-white/90 dark:bg-gray-800/90 backdrop-blur-md p-8 rounded-xl shadow-neumorphic max-w-sm w-full mx-4" 
        role="dialog"
        aria-labelledby="modal-title"
        aria-describedby="modal-description"
        onClick={(e) => e.stopPropagation()}
      >
        <button 
          onClick={handleClose} 
          className="absolute top-4 right-4 text-primary hover:text-primary/80 transition-colors duration-200" 
          aria-label="Close modal"
        >
          <XMarkIcon className="h-5 w-5" />
        </button>
        
        <div className="text-center">
          <h2 id="modal-title" className="text-xl font-semibold text-gray-900 dark:text-white mb-4">
            Welcome to SecureEmail!
          </h2>
          <p id="modal-description" className="text-gray-600 dark:text-gray-300 text-sm mb-6 leading-relaxed">
            Your emails are encrypted with AES-256-GCM and auto-deleted after 3 failed attempts.
          </p>
          <button 
            onClick={handleClose} 
            className="w-full py-3 bg-accent text-white rounded-lg font-medium hover:scale-[1.02] transition-transform duration-200"
          >
            Get Started
          </button>
        </div>
      </div>
    </div>
  );
};

export default OnboardingModal; 