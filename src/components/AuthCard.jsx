import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast, ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { EyeIcon, EyeSlashIcon } from '@heroicons/react/24/outline';
import { login, signup } from '../lib/api';

const AuthCard = () => {
  const [isLogin, setIsLogin] = useState(true);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const navigate = useNavigate();

  const validateLogin = () => {
    if (!/^[^@]+@securesystem\.email$/.test(email)) return 'Invalid email';
    if (password.length < 8 || password.length > 128) return 'Password must be 8–128 chars';
    return null;
  };

  const validateSignUp = () => {
    if (!/^[^@]+@securesystem\.email$/.test(email)) return 'Invalid email';
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
      const response = await login({ email, password });
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
      sessionStorage.setItem('token', response.data.token);
      navigate('/inbox');
      toast.success('Account created successfully!');
    } catch (err) {
      toast.error(err.response?.data?.error || 'Sign-up failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-neutral-light dark:bg-neutral-dark p-6">
      <div className="w-full max-w-md bg-white dark:bg-neutral-dark rounded-2xl shadow-card p-8">
        {isLogin ? (
          <form onSubmit={handleLogin} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-primary mb-2">
                Your email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white text-gray-900 shadow-input"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-primary mb-2">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white text-gray-900 shadow-input pr-12"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5" />
                  ) : (
                    <EyeIcon className="h-5 w-5" />
                  )}
                </button>
              </div>
              <div className="flex justify-end mt-2">
                <button
                  type="button"
                  className="text-sm text-gray-500 hover:text-primary"
                  onClick={() => toast.info('Password reset coming soon!')}
                >
                  Forgot password?
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 bg-primary text-white rounded-xl font-medium hover:bg-primary/90 transition-colors duration-200 disabled:opacity-50"
            >
              {isSubmitting ? 'Signing in...' : 'Login'}
            </button>

            <div className="relative mt-8">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-200"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-4 bg-white text-gray-500">or</span>
              </div>
            </div>

            <div className="text-center mt-6">
              <span className="text-gray-600">Don't have an account? </span>
              <button
                type="button"
                onClick={() => setIsLogin(false)}
                className="text-primary hover:text-primary/80 font-medium"
              >
                Register new
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleSignUp} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-primary mb-2">
                Your email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white text-gray-900 shadow-input"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-primary mb-2">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white text-gray-900 shadow-input pr-12"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5" />
                  ) : (
                    <EyeIcon className="h-5 w-5" />
                  )}
                </button>
              </div>
            </div>

            <div>
              <label className="block text-sm font-medium text-primary mb-2">
                Confirm Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-4 py-3 rounded-xl border border-gray-200 focus:ring-2 focus:ring-primary/20 focus:border-primary bg-white text-gray-900 shadow-input pr-12"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5" />
                  ) : (
                    <EyeIcon className="h-5 w-5" />
                  )}
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-3 bg-primary text-white rounded-xl font-medium hover:bg-primary/90 transition-colors duration-200 disabled:opacity-50"
            >
              {isSubmitting ? 'Creating Account...' : 'Create Account'}
            </button>

            <div className="relative mt-8">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-gray-200"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-4 bg-white text-gray-500">or</span>
              </div>
            </div>

            <div className="text-center mt-6">
              <span className="text-gray-600">Already have an account? </span>
              <button
                type="button"
                onClick={() => setIsLogin(true)}
                className="text-primary hover:text-primary/80 font-medium"
              >
                Login
              </button>
            </div>
          </form>
        )}
      </div>
      <ToastContainer position="top-right" />
    </div>
  );
};

export default AuthCard; 