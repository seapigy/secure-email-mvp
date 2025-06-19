import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import axios from 'axios';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { gsap } from 'gsap';
import { 
  LockClosedIcon, 
  LockOpenIcon, 
  QuestionMarkCircleIcon,
  SunIcon,
  MoonIcon
} from '@heroicons/react/24/outline';
import { Tooltip } from 'react-tooltip';

const Login = () => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [isDarkMode, setIsDarkMode] = useState(false);
  const navigate = useNavigate();
  const cardRef = useRef(null);
  const buttonRef = useRef(null);

  // Dark mode detection and toggle
  useEffect(() => {
    const isDark = document.documentElement.classList.contains('dark');
    setIsDarkMode(isDark);
  }, []);

  const toggleDarkMode = () => {
    const newMode = !isDarkMode;
    setIsDarkMode(newMode);
    if (newMode) {
      document.documentElement.classList.add('dark');
      localStorage.setItem('theme', 'dark');
    } else {
      document.documentElement.classList.remove('dark');
      localStorage.setItem('theme', 'light');
    }
  };

  // GSAP animations
  useEffect(() => {
    // Animate card fade-in
    gsap.fromTo(cardRef.current, 
      { opacity: 0, y: 20 }, 
      { opacity: 1, y: 0, duration: 0.3, ease: "power2.out" }
    );
  }, []);

  // Button hover animation
  const handleButtonHover = () => {
    gsap.to(buttonRef.current, { scale: 1.05, duration: 0.2 });
  };

  const handleButtonLeave = () => {
    gsap.to(buttonRef.current, { scale: 1, duration: 0.2 });
  };

  // Input validation
  const validateInputs = () => {
    const emailRegex = /^[^@]+@securesystem\.email$/;
    if (!emailRegex.test(email)) {
      return 'Invalid email format. Must be user@securesystem.email';
    }
    if (password.length < 8 || password.length > 128) {
      return 'Password must be 8–128 characters';
    }
    if (!/^\d{6}$/.test(totpCode)) {
      return 'TOTP code must be 6 digits';
    }
    return null;
  };

  // Form submission
  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);

    const error = validateInputs();
    if (error) {
      toast.error(error, { 
        theme: 'colored', 
        className: 'toast-error',
        position: "top-right",
        autoClose: 5000
      });
      setIsSubmitting(false);
      return;
    }

    try {
      const response = await axios.post('https://api.securesystem.email/api/auth/login', {
        email,
        password,
        totp_code: totpCode,
      });

      // Store JWT token
      sessionStorage.setItem('token', response.data.token);
      sessionStorage.setItem('user_id', response.data.user_id);

      // Show success toast
      toast.success('Login successful! Redirecting...', {
        theme: 'colored',
        className: 'toast-success',
        position: "top-right",
        autoClose: 2000
      });

      // Redirect to inbox
      setTimeout(() => {
        navigate('/inbox');
      }, 1000);

    } catch (err) {
      let message = 'Login failed';
      
      if (err.response) {
        switch (err.response.status) {
          case 401:
            message = 'Invalid credentials. Please check your email, password, and TOTP code.';
            break;
          case 429:
            message = 'Too many login attempts. Please wait a minute before trying again.';
            break;
          case 400:
            message = 'Invalid request format. Please check your input.';
            break;
          default:
            message = err.response.data?.error || 'Login failed. Please try again.';
        }
      } else if (err.request) {
        message = 'Unable to connect to server. Please check your internet connection.';
      }

      toast.error(message, { 
        theme: 'colored', 
        className: 'toast-error',
        position: "top-right",
        autoClose: 5000
      });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-light dark:bg-neutral-dark p-4">
      {/* Dark mode toggle */}
      <button
        onClick={toggleDarkMode}
        className="absolute top-4 right-4 p-2 rounded-full bg-white dark:bg-gray-800 shadow-lg hover:scale-105 transition-transform duration-200"
        aria-label={isDarkMode ? 'Switch to light mode' : 'Switch to dark mode'}
      >
        {isDarkMode ? (
          <SunIcon className="h-6 w-6 text-yellow-500" />
        ) : (
          <MoonIcon className="h-6 w-6 text-gray-600" />
        )}
      </button>

      {/* Login card */}
      <div 
        ref={cardRef}
        className="neumorphic-card bg-white dark:bg-gray-800 p-8 max-w-md w-full"
      >
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold text-primary dark:text-blue-400 mb-2">
            Secure Email
          </h1>
          <p className="text-text-gray dark:text-gray-300">
            Sign in to your encrypted inbox
          </p>
        </div>

        {/* Login form */}
        <form onSubmit={handleSubmit} role="form" aria-live="polite">
          {/* Email field */}
          <div className="mb-6">
            <label htmlFor="email" className="block text-text-gray dark:text-gray-300 mb-2 font-medium">
              <LockClosedIcon className="h-5 w-5 inline mr-2" aria-hidden="true" />
              Email Address
            </label>
            <input
              id="email"
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="input-field"
              placeholder="user@securesystem.email"
              aria-label="Email input"
              aria-describedby="email-help"
              required
            />
            <p id="email-help" className="sr-only">
              Enter your email address in the format user@securesystem.email
            </p>
          </div>

          {/* Password field */}
          <div className="mb-6">
            <label htmlFor="password" className="block text-text-gray dark:text-gray-300 mb-2 font-medium">
              <LockOpenIcon className="h-5 w-5 inline mr-2" aria-hidden="true" />
              Password
            </label>
            <input
              id="password"
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="input-field"
              placeholder="Enter your password"
              aria-label="Password input"
              aria-describedby="password-help"
              required
            />
            <p id="password-help" className="sr-only">
              Enter your password (8-128 characters)
            </p>
          </div>

          {/* TOTP field */}
          <div className="mb-8">
            <label htmlFor="totp" className="block text-text-gray dark:text-gray-300 mb-2 font-medium">
              <QuestionMarkCircleIcon
                className="h-5 w-5 inline mr-2 cursor-help"
                data-tooltip-id="totp-tooltip"
                aria-hidden="true"
              />
              TOTP Code
            </label>
            <input
              id="totp"
              type="text"
              maxLength={6}
              value={totpCode}
              onChange={(e) => setTotpCode(e.target.value.replace(/\D/g, ''))}
              className="input-field text-center text-lg tracking-widest"
              placeholder="123456"
              aria-label="TOTP code input"
              aria-describedby="totp-help"
              required
            />
            <p id="totp-help" className="sr-only">
              Enter the 6-digit code from your authenticator app
            </p>
            <Tooltip 
              id="totp-tooltip" 
              content="Scan QR code in authenticator app" 
              place="right"
              className="bg-gray-800 text-white p-2 rounded shadow-lg"
            />
          </div>

          {/* Submit button */}
          <button
            ref={buttonRef}
            type="submit"
            disabled={isSubmitting}
            onMouseEnter={handleButtonHover}
            onMouseLeave={handleButtonLeave}
            className="btn-primary w-full"
            aria-label="Sign in"
          >
            {isSubmitting ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white mr-2"></div>
                Signing In...
              </div>
            ) : (
              'Sign In'
            )}
          </button>
        </form>

        {/* Footer */}
        <div className="mt-6 text-center">
          <p className="text-sm text-text-gray dark:text-gray-400">
            Your emails are encrypted with AES-256-GCM
          </p>
        </div>
      </div>

      {/* Toast container */}
      <ToastContainer />
    </div>
  );
};

export default Login; 