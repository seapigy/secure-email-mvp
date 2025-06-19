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
    <div className="min-h-screen flex items-start justify-center pt-16 bg-[#F8F9FA]">
      <div className="w-full max-w-sm">
        {isLogin ? (
          <form onSubmit={handleLogin} className="space-y-4">
            <div>
              <label className="block text-sm text-black mb-1">
                Your email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 rounded border border-black text-black text-sm"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-black mb-1">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-3 py-2 rounded border border-black text-black text-sm pr-10"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-1/2 -translate-y-1/2"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5 text-black" />
                  ) : (
                    <EyeIcon className="h-5 w-5 text-black" />
                  )}
                </button>
              </div>
              <div className="flex justify-start mt-1">
                <button
                  type="button"
                  className="text-sm text-black hover:underline"
                  onClick={() => toast.info('Password reset coming soon!')}
                >
                  Forgot password?
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-2 bg-black text-white rounded text-sm font-normal disabled:opacity-50"
            >
              {isSubmitting ? 'Signing in...' : 'Login'}
            </button>

            <div className="relative mt-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-black"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-[#F8F9FA] text-black">or</span>
              </div>
            </div>

            <div className="text-center mt-6">
              <span className="text-sm text-black">Don't have an account? </span>
              <button
                type="button"
                onClick={() => setIsLogin(false)}
                className="text-sm text-black bg-black/10 rounded px-2 py-1 hover:bg-black/20"
              >
                Register new
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleSignUp} className="space-y-4">
            <div>
              <label className="block text-sm text-black mb-1">
                Your email
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                className="w-full px-3 py-2 rounded border border-black text-black text-sm"
                placeholder="user@securesystem.email"
                required
              />
            </div>

            <div>
              <label className="block text-sm text-black mb-1">
                Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full px-3 py-2 rounded border border-black text-black text-sm pr-10"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-1/2 -translate-y-1/2"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5 text-black" />
                  ) : (
                    <EyeIcon className="h-5 w-5 text-black" />
                  )}
                </button>
              </div>
            </div>

            <div>
              <label className="block text-sm text-black mb-1">
                Confirm Password
              </label>
              <div className="relative">
                <input
                  type={showPassword ? "text" : "password"}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  className="w-full px-3 py-2 rounded border border-black text-black text-sm pr-10"
                  placeholder="••••••••"
                  required
                />
                <button
                  type="button"
                  onClick={() => setShowPassword(!showPassword)}
                  className="absolute right-2 top-1/2 -translate-y-1/2"
                >
                  {showPassword ? (
                    <EyeSlashIcon className="h-5 w-5 text-black" />
                  ) : (
                    <EyeIcon className="h-5 w-5 text-black" />
                  )}
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isSubmitting}
              className="w-full py-2 bg-black text-white rounded text-sm font-normal disabled:opacity-50"
            >
              {isSubmitting ? 'Creating Account...' : 'Create Account'}
            </button>

            <div className="relative mt-6">
              <div className="absolute inset-0 flex items-center">
                <div className="w-full border-t border-black"></div>
              </div>
              <div className="relative flex justify-center text-sm">
                <span className="px-2 bg-[#F8F9FA] text-black">or</span>
              </div>
            </div>

            <div className="text-center mt-6">
              <span className="text-sm text-black">Already have an account? </span>
              <button
                type="button"
                onClick={() => setIsLogin(true)}
                className="text-sm text-black bg-black/10 rounded px-2 py-1 hover:bg-black/20"
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