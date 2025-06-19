import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { gsap } from 'gsap';
import { LockClosedIcon, KeyIcon, QuestionMarkCircleIcon } from '@heroicons/react/24/outline';
import { Tooltip } from 'react-tooltip';
import QRCode from 'qrcode.react';
import { login, signup, verifyTotp } from '../lib/api';

const AuthCard = () => {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [totpCode, setTotpCode] = useState('');
  const [totpQr, setTotpQr] = useState('');
  const [tempId, setTempId] = useState('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();

  useEffect(() => {
    gsap.fromTo('.auth-card', { opacity: 0 }, { opacity: 1, duration: 0.3 });
  }, []);

  const validateLogin = () => {
    if (!/^[^@]+@securesystem.email$/.test(email)) return 'Invalid email format';
    if (password.length < 8 || password.length > 128) return 'Password must be 8–128 characters';
    if (!/^\d{6}$/.test(totpCode)) return 'TOTP code must be 6 digits';
    return null;
  };

  const validateSignUp = () => {
    if (!/^[^@]+@securesystem.email$/.test(email)) return 'Invalid email format';
    if (password.length < 8 || password.length > 128) return 'Password must be 8–128 characters';
    if (password !== confirmPassword) return 'Passwords do not match';
    return null;
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    const error = validateLogin();
    if (error) {
      toast.error(error, { className: 'bg-error text-white' });
      setIsSubmitting(false);
      return;
    }
    try {
      const response = await login({ email, password, totp_code: totpCode });
      sessionStorage.setItem('token', response.data.token);
      navigate('/inbox');
      toast.success('Logged in successfully!', { className: 'bg-success text-white' });
    } catch (err) {
      toast.error(err.response?.data?.error || 'Login failed', { className: 'bg-error text-white' });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSignUp = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    const error = validateSignUp();
    if (error) {
      toast.error(error, { className: 'bg-error text-white' });
      setIsSubmitting(false);
      return;
    }
    try {
      const response = await signup({ email, password, confirm_password: confirmPassword });
      setTotpQr(response.data.totp_qr);
      setTempId(response.data.temp_id);
      toast.info('Scan the QR code with your authenticator app', { className: 'bg-primary text-white' });
    } catch (err) {
      toast.error(err.response?.data?.error || 'Sign-up failed', { className: 'bg-error text-white' });
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleTotpConfirm = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    try {
      const response = await verifyTotp({ temp_id: tempId, totp_code: totpCode });
      sessionStorage.setItem('token', response.data.token);
      navigate('/inbox');
      toast.success('Account created successfully!', { className: 'bg-success text-white' });
    } catch (err) {
      toast.error(err.response?.data?.error || 'TOTP verification failed', { className: 'bg-error text-white' });
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-light dark:bg-neutral-dark p-4">
      <div className="auth-card bg-white dark:bg-gray-800 p-8 rounded-xl shadow-[0_4px_6px_rgba(0,0,0,0.1),inset_0_1px_0_rgba(255,255,255,0.2)] max-w-md w-full">
        {/* Tab Navigation */}
        <div className="flex justify-center mb-8">
          <div className="flex bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
            <button
              className={`px-6 py-3 rounded-md font-medium transition-all duration-200 ${
                isLogin 
                  ? 'bg-primary text-white shadow-md' 
                  : 'text-text-gray dark:text-gray-300 hover:text-primary dark:hover:text-primary'
              }`}
              onClick={() => setIsLogin(true)}
            >
              Login
            </button>
            <button
              className={`px-6 py-3 rounded-md font-medium transition-all duration-200 ${
                !isLogin 
                  ? 'bg-primary text-white shadow-md' 
                  : 'text-text-gray dark:text-gray-300 hover:text-primary dark:hover:text-primary'
              }`}
              onClick={() => setIsLogin(false)}
            >
              Sign Up
            </button>
          </div>
        </div>

        {/* Login Form */}
        {isLogin ? (
          <form onSubmit={handleLogin} role="form" aria-live="polite" className="space-y-6">
            <div>
              <label htmlFor="email" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <LockClosedIcon className="h-5 w-5 inline mr-2 text-primary" />
                Email Address
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all duration-200"
                placeholder="user@securesystem.email"
                aria-label="Email input"
                required
              />
            </div>

            <div>
              <label htmlFor="password" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <KeyIcon className="h-5 w-5 inline mr-2 text-primary" />
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all duration-200"
                placeholder="Enter your password"
                aria-label="Password input"
                aria-describedby="password-tooltip"
                required
              />
              <Tooltip id="password-tooltip" content="8+ characters, include numbers/symbols" place="right" />
            </div>

            <div>
              <label htmlFor="totp" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <QuestionMarkCircleIcon 
                  className="h-5 w-5 inline mr-2 text-primary cursor-help" 
                  data-tooltip-id="totp-tooltip" 
                />
                TOTP Code
              </label>
              <input
                id="totp"
                type="text"
                maxLength={6}
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-primary focus:ring-2 focus:ring-primary/20 transition-all duration-200 text-center text-lg tracking-widest"
                placeholder="123456"
                aria-label="TOTP code input"
                aria-describedby="totp-tooltip"
                required
              />
              <Tooltip id="totp-tooltip" content="Set up 2FA with Google Authenticator" place="right" />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full p-4 bg-primary text-white rounded-lg font-semibold hover:scale-105 transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
              aria-label="Sign in"
            >
              {isSubmitting ? 'Signing In...' : 'Sign In'}
            </button>
          </form>
        ) : totpQr ? (
          /* TOTP Confirmation Form */
          <form onSubmit={handleTotpConfirm} role="form" aria-live="polite" className="space-y-6">
            <div className="text-center">
              <h3 className="text-lg font-semibold text-text-gray dark:text-white mb-4">
                Set Up Two-Factor Authentication
              </h3>
              <p className="text-text-gray dark:text-gray-300 mb-6">
                Scan this QR code with your authenticator app:
              </p>
              <div className="bg-white p-4 rounded-lg inline-block shadow-lg">
                <QRCode value={totpQr} size={160} className="mx-auto" />
              </div>
            </div>

            <div>
              <label htmlFor="totp-confirm" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <QuestionMarkCircleIcon 
                  className="h-5 w-5 inline mr-2 text-primary cursor-help" 
                  data-tooltip-id="totp-confirm-tooltip" 
                />
                Enter TOTP Code
              </label>
              <input
                id="totp-confirm"
                type="text"
                maxLength={6}
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-accent focus:ring-2 focus:ring-accent/20 transition-all duration-200 text-center text-lg tracking-widest"
                placeholder="123456"
                aria-label="TOTP code input"
                aria-describedby="totp-confirm-tooltip"
                required
              />
              <Tooltip id="totp-confirm-tooltip" content="Enter code from authenticator app" place="right" />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full p-4 bg-accent text-white rounded-lg font-semibold hover:scale-105 transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
              aria-label="Confirm sign-up"
            >
              {isSubmitting ? 'Confirming...' : 'Confirm Sign-Up'}
            </button>
          </form>
        ) : (
          /* Sign-Up Form */
          <form onSubmit={handleSignUp} role="form" aria-live="polite" className="space-y-6">
            <div>
              <label htmlFor="signup-email" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <LockClosedIcon className="h-5 w-5 inline mr-2 text-primary" />
                Email Address
              </label>
              <input
                id="signup-email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-accent focus:ring-2 focus:ring-accent/20 transition-all duration-200"
                placeholder="user@securesystem.email"
                aria-label="Email input"
                aria-describedby="email-tooltip"
                required
              />
              <Tooltip id="email-tooltip" content="Must end with @securesystem.email" place="right" />
            </div>

            <div>
              <label htmlFor="signup-password" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <KeyIcon className="h-5 w-5 inline mr-2 text-primary" />
                Password
              </label>
              <input
                id="signup-password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-accent focus:ring-2 focus:ring-accent/20 transition-all duration-200"
                placeholder="Create a strong password"
                aria-label="Password input"
                aria-describedby="signup-password-tooltip"
                required
              />
              <Tooltip id="signup-password-tooltip" content="8+ characters, include numbers/symbols" place="right" />
            </div>

            <div>
              <label htmlFor="confirm-password" className="block text-text-gray dark:text-gray-300 mb-3 font-medium">
                <KeyIcon className="h-5 w-5 inline mr-2 text-primary" />
                Confirm Password
              </label>
              <input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full p-4 rounded-lg border-2 border-gray-200 dark:border-gray-600 bg-neutral-light dark:bg-gray-700 text-text-gray dark:text-white focus:border-accent focus:ring-2 focus:ring-accent/20 transition-all duration-200"
                placeholder="Confirm your password"
                aria-label="Confirm password input"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full p-4 bg-accent text-white rounded-lg font-semibold hover:scale-105 transition-all duration-150 disabled:opacity-50 disabled:cursor-not-allowed shadow-lg hover:shadow-xl"
              aria-label="Sign up"
            >
              {isSubmitting ? 'Signing Up...' : 'Sign Up'}
            </button>
          </form>
        )}
      </div>
      <ToastContainer 
        position="top-right" 
        autoClose={3000}
        toastClassName="rounded-lg font-medium"
      />
    </div>
  );
};

export default AuthCard; 