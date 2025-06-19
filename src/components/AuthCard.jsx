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
      const response = await signup({ email, password, confirm_password: confirmPassword });
      if (response.data.totp_qr) {
        // Show TOTP setup modal
        sessionStorage.setItem('temp_id', response.data.temp_id);
        sessionStorage.setItem('totp_qr', response.data.totp_qr);
        navigate('/setup-totp');
      }
    } catch (err) {
      toast.error(err.response?.data?.error || 'Sign-up failed');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="flex items-center justify-center min-h-screen px-4">
      <div className="w-80 max-w-xs mx-auto bg-white/10 backdrop-blur-lg rounded-xl shadow-lg border border-white/20 px-4 py-6">
        {isLogin ? (
          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <label className="block text-sm mb-1 text-white font-medium">
                Your email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="text-sm px-3 py-2 rounded-md bg-white/20 text-white w-full placeholder-gray-300 border border-white/20 focus:border-white/40 focus:outline-none transition-colors"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm mb-1 text-white font-medium">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="text-sm px-3 py-2 rounded-md bg-white/20 text-white w-full pr-10 placeholder-gray-300 border border-white/20 focus:border-white/40 focus:outline-none transition-colors"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  {showPassword ? (
                    <EyeSlashIcon className="w-5 h-5 text-gray-300" />
                  ) : (
                    <EyeIcon className="w-5 h-5 text-gray-300" />
                  )}
                </button>
              </div>
              <div className="flex justify-end mt-1">
                <button
                  type="button"
                  className="text-sm text-gray-300 hover:text-white transition-colors"
                  onClick={() => toast.info('Password reset coming soon!')}
                >
                  Forgot password?
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full bg-blue-900 text-white text-sm font-bold rounded-full py-2 hover:bg-blue-800 transition-colors disabled:opacity-50"
            >
              {isSubmitting ? 'Signing in...' : 'Login'}
            </button>

            <div className="relative mt-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-white/20"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white/10 text-gray-300">or</span>
              </div>
            </div>

            <div className="text-center mt-6">
              <span className="text-sm text-gray-300">Don't have an account? </span>
              <button
                type="button"
                onClick={() => setIsLogin(false)}
                className="text-sm text-white font-medium hover:text-blue-200 transition-colors"
              >
                Register new
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleSignUp} className="space-y-4">
            <div>
              <label className="block text-sm mb-1 text-white font-medium">
                Your email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="text-sm px-3 py-2 rounded-md bg-white/20 text-white w-full placeholder-gray-300 border border-white/20 focus:border-white/40 focus:outline-none transition-colors"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm mb-1 text-white font-medium">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="text-sm px-3 py-2 rounded-md bg-white/20 text-white w-full pr-10 placeholder-gray-300 border border-white/20 focus:border-white/40 focus:outline-none transition-colors"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  {showPassword ? (
                    <EyeSlashIcon className="w-5 h-5 text-gray-300" />
                  ) : (
                    <EyeIcon className="w-5 h-5 text-gray-300" />
                  )}
                </button>
              </div>
            </div>

            <div>
              <label className="block text-sm mb-1 text-white font-medium">
                Confirm Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="text-sm px-3 py-2 rounded-md bg-white/20 text-white w-full pr-10 placeholder-gray-300 border border-white/20 focus:border-white/40 focus:outline-none transition-colors"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2"
                  aria-label={showPassword ? "Hide password" : "Show password"}
                >
                  {showPassword ? (
                    <EyeSlashIcon className="w-5 h-5 text-gray-300" />
                  ) : (
                    <EyeIcon className="w-5 h-5 text-gray-300" />
                  )}
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full bg-blue-900 text-white text-sm font-bold rounded-full py-2 hover:bg-blue-800 transition-colors disabled:opacity-50"
            >
              {isSubmitting ? 'Creating Account...' : 'Create Account'}
            </button>

            <div className="relative mt-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-white/20"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-white/10 text-gray-300">or</span>
              </div>
            </div>

            <div className="text-center mt-6">
              <span className="text-sm text-gray-300">Already have an account? </span>
              <button
                type="button"
                onClick={() => setIsLogin(true)}
                className="text-sm text-white font-medium hover:text-blue-200 transition-colors"
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