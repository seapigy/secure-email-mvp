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
    gsap.fromTo('.auth-card', { opacity: 0, y: 20 }, { opacity: 1, y: 0, duration: 0.3 });
  }, []);

  const validateLogin = () => {
    if (!/^[^@]+@securesystem.email$/.test(email)) return 'Invalid email';
    if (password.length < 8 || password.length > 128) return 'Password must be 8–128 chars';
    if (!/^\d{6}$/.test(totpCode)) return 'TOTP must be 6 digits';
    return null;
  };

  const validateSignUp = () => {
    if (!/^[^@]+@securesystem.email$/.test(email)) return 'Invalid email';
    if (password.length < 8 || password.length > 128) return 'Password must be 8–128 chars';
    if (password !== confirmPassword) return 'Passwords do not match';
    return null;
  };

  const handleLogin = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    const error = validateLogin();
    if (error) {
      toast.error(error);
      setIsSubmitting(false);
      return;
    }
    try {
      const response = await login({ email, password, totp_code: totpCode });
      sessionStorage.setItem('token', response.data.token);
      navigate('/inbox');
      toast.success('Logged in successfully!');
    } catch (err) {
      toast.error(err.response?.data?.error || 'Login failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleSignUp = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    const error = validateSignUp();
    if (error) {
      toast.error(error);
      setIsSubmitting(false);
      return;
    }
    try {
      const response = await signup({ email, password });
      setTotpQr(response.data.totp_qr);
      setTempId(response.data.temp_id);
    } catch (err) {
      toast.error(err.response?.data?.error || 'Sign-up failed');
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
      toast.success('Account created successfully!');
    } catch (err) {
      toast.error(err.response?.data?.error || 'TOTP verification failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-light dark:bg-neutral-dark p-4">
      <div className="auth-card w-full max-w-md bg-white/90 dark:bg-gray-800/90 rounded-xl shadow-neumorphic dark:shadow-neumorphic-dark p-8">
        <div className="flex justify-center mb-8">
          <div className="inline-flex bg-gray-100 dark:bg-gray-700/50 rounded-lg p-1">
            <button
              onClick={() => setIsLogin(true)}
              className={`px-6 py-2 rounded-lg transition-all duration-200 ${
                isLogin 
                  ? 'bg-primary text-white shadow-sm' 
                  : 'text-gray-600 dark:text-gray-300 hover:text-primary'
              }`}
            >
              Login
            </button>
            <button
              onClick={() => setIsLogin(false)}
              className={`px-6 py-2 rounded-lg transition-all duration-200 ${
                !isLogin 
                  ? 'bg-primary text-white shadow-sm' 
                  : 'text-gray-600 dark:text-gray-300 hover:text-primary'
              }`}
            >
              Sign Up
            </button>
          </div>
        </div>

        {isLogin ? (
          <form onSubmit={handleLogin} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                <LockClosedIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white transition-colors duration-200"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                <KeyIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white transition-colors duration-200"
                placeholder="••••••••"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                <QuestionMarkCircleIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                TOTP Code
              </label>
              <input
                type="text"
                maxLength={6}
                value={totpCode}
                onChange={(e) => setTotpCode(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-center tracking-widest transition-colors duration-200"
                placeholder="123456"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 bg-primary text-white rounded-lg font-medium hover:scale-[1.02] transition-transform duration-200 disabled:opacity-50 disabled:hover:scale-100"
            >
              {isSubmitting ? 'Signing In...' : 'Sign In'}
            </button>
          </form>
        ) : totpQr ? (
          <div className="space-y-6">
            <div className="text-center">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Set Up Two-Factor Authentication
              </h3>
              <div className="bg-white p-4 rounded-lg inline-block mb-4 shadow-sm">
                <QRCode value={totpQr} size={180} />
              </div>
              <p className="text-sm text-gray-600 dark:text-gray-300 mb-6">
                Scan this QR code with your authenticator app
              </p>
            </div>

            <form onSubmit={handleTotpConfirm} className="space-y-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                  <QuestionMarkCircleIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                  Enter TOTP Code
                </label>
                <input
                  type="text"
                  maxLength={6}
                  value={totpCode}
                  onChange={(e) => setTotpCode(e.target.value)}
                  className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-center tracking-widest transition-colors duration-200"
                  placeholder="123456"
                  required
                />
              </div>

              <button
                type="submit"
                disabled={isSubmitting}
                className="w-full py-3 bg-accent text-white rounded-lg font-medium hover:scale-[1.02] transition-transform duration-200 disabled:opacity-50 disabled:hover:scale-100"
              >
                {isSubmitting ? 'Verifying...' : 'Verify & Complete Sign Up'}
              </button>
            </form>
          </div>
        ) : (
          <form onSubmit={handleSignUp} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                <LockClosedIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                Email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white transition-colors duration-200"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                <KeyIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                Password
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white transition-colors duration-200"
                placeholder="••••••••"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                <KeyIcon className="w-4 h-4 inline-block mr-2 text-primary" />
                Confirm Password
              </label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
                className="w-full px-4 py-3 rounded-lg border border-gray-300 dark:border-gray-600 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white dark:bg-gray-700 text-gray-900 dark:text-white transition-colors duration-200"
                placeholder="••••••••"
                required
              />
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 bg-accent text-white rounded-lg font-medium hover:scale-[1.02] transition-transform duration-200 disabled:opacity-50 disabled:hover:scale-100"
            >
              {isSubmitting ? 'Creating Account...' : 'Create Account'}
            </button>
          </form>
        )}
      </div>
      <ToastContainer position="top-right" />
    </div>
  );
};

export default AuthCard; 