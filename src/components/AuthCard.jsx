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
    <div className="min-h-screen flex items-center justify-center bg-neutral-light dark:bg-neutral-dark">
      <div className="auth-card bg-white dark:bg-gray-800 p-8 rounded-xl shadow-[0_4px_6px_rgba(0,0,0,0.1),inset_0_1px_0_rgba(255,255,255,0.2)] max-w-md w-full">
        <div className="flex justify-center mb-6">
          <button
            className={`px-4 py-2 ${isLogin ? 'bg-primary text-white' : 'bg-gray-200 dark:bg-gray-600'} rounded-l-md`}
            onClick={() => setIsLogin(true)}
          >
            Login
          </button>
          <button
            className={`px-4 py-2 ${!isLogin ? 'bg-primary text-white' : 'bg-gray-200 dark:bg-gray-600'} rounded-r-md`}
            onClick={() => setIsLogin(false)}
          >
            Sign Up
          </button>
        </div>
        {isLogin ? (
          <form onSubmit={handleLogin} role="form" aria-live="polite">
            <div className="mb-6">
              <label htmlFor="email" className="block text-text-gray dark:text-gray-300 mb-2">
                <LockClosedIcon className="h-5 w-5 inline mr-2" />
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="Email input"
                required
              />
            </div>
            <div className="mb-6">
              <label htmlFor="password" className="block text-text-gray dark:text-gray-300 mb-2">
                <KeyIcon className="h-5 w-5 inline mr-2" />
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="Password input"
                aria-describedby="password-tooltip"
                required
              />
              <Tooltip id="password-tooltip" content="8+ characters, include numbers/symbols" place="right" />
            </div>
            <div className="mb-6">
              <label htmlFor="totp" className="block text-text-gray dark:text-gray-300 mb-2">
                <QuestionMarkCircleIcon className="h-5 w-5 inline mr-2 cursor-help" data-tooltip-id="totp-tooltip" />
                TOTP Code
              </label>
              <input
                id="totp"
                type="text"
                maxLength={6}
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="TOTP code input"
                aria-describedby="totp-tooltip"
                required
              />
              <Tooltip id="totp-tooltip" content="Set up 2FA with Google Authenticator" place="right" />
            </div>
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full p-3 bg-primary text-white rounded-md hover:scale-105 transition-transform duration-150 disabled:opacity-50"
            >
              {isSubmitting ? 'Signing In...' : 'Sign In'}
            </button>
          </form>
        ) : totpQr ? (
          <form onSubmit={handleTotpConfirm} role="form" aria-live="polite">
            <div className="mb-6 text-center">
              <p className="text-text-gray dark:text-gray-300 mb-4">Scan this QR code with your authenticator app:</p>
              <QRCode value={totpQr} size={160} className="mx-auto" />
            </div>
            <div className="mb-6">
              <label htmlFor="totp" className="block text-text-gray dark:text-gray-300 mb-2">
                <QuestionMarkCircleIcon className="h-5 w-5 inline mr-2 cursor-help" data-tooltip-id="totp-tooltip" />
                TOTP Code
              </label>
              <input
                id="totp"
                type="text"
                maxLength={6}
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="TOTP code input"
                aria-describedby="totp-tooltip"
                required
              />
              <Tooltip id="totp-tooltip" content="Enter code from authenticator app" place="right" />
            </div>
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full p-3 bg-accent text-white rounded-md hover:scale-105 transition-transform duration-150 disabled:opacity-50"
            >
              {isSubmitting ? 'Confirming...' : 'Confirm Sign-Up'}
            </button>
          </form>
        ) : (
          <form onSubmit={handleSignUp} role="form" aria-live="polite">
            <div className="mb-6">
              <label htmlFor="email" className="block text-text-gray dark:text-gray-300 mb-2">
                <LockClosedIcon className="h-5 w-5 inline mr-2" />
                Email
              </label>
              <input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="Email input"
                aria-describedby="email-tooltip"
                required
              />
              <Tooltip id="email-tooltip" content="Must end with @securesystem.email" place="right" />
            </div>
            <div className="mb-6">
              <label htmlFor="password" className="block text-text-gray dark:text-gray-300 mb-2">
                <KeyIcon className="h-5 w-5 inline mr-2" />
                Password
              </label>
              <input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="Password input"
                aria-describedby="password-tooltip"
                required
              />
              <Tooltip id="password-tooltip" content="8+ characters, include numbers/symbols" place="right" />
            </div>
            <div className="mb-6">
              <label htmlFor="confirm-password" className="block text-text-gray dark:text-gray-300 mb-2">
                <KeyIcon className="h-5 w-5 inline mr-2" />
                Confirm Password
              </label>
              <input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full p-3 rounded-md border border-gray-300 dark:border-gray-600 bg-neutral-light dark:bg-gray-700"
                aria-label="Confirm password input"
                required
              />
            </div>
            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full p-3 bg-accent text-white rounded-md hover:scale-105 transition-transform duration-150 disabled:opacity-50"
            >
              {isSubmitting ? 'Signing Up...' : 'Sign Up'}
            </button>
          </form>
        )}
      </div>
      <ToastContainer position="top-right" autoClose={3000} />
    </div>
  );
};

export default AuthCard; 