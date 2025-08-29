/**
 * ⚠️ CRITICAL WARNING - DESIGN PRESERVATION ⚠️
 * 
 * THIS FILE CONTAINS THE MAIN APPLICATION ROUTING AND LAYOUT.
 * 
 * 🚨 CRITICAL RULES:
 * 1. NEVER change the visual layout or design of any components
 * 2. NEVER modify the routing structure that affects the ComposeModal design
 * 3. NEVER alter the component imports that could affect the "perfect" design
 * 4. ONLY add new routes or functionality that doesn't change existing designs
 * 5. ALWAYS maintain the exact same visual appearance of all components
 * 
 * The user has explicitly stated: "MAKE A NOTE IN THE CODE NEVER CHANGE THE DESIGN EVER. ITS NEVER OK TO DO REMEMBER IT"
 * 
 * The ComposeModal design was restored from commit e291daf and represents the "perfect" design.
 * Any changes to the visual design will result in immediate user dissatisfaction.
 * 
 * ⚠️ IF YOU ARE CONSIDERING CHANGING THE DESIGN, STOP IMMEDIATELY ⚠️
 * 
 * @author: AI Assistant
 * @warning: DESIGN PRESERVATION CRITICAL
 * @user_feedback: "This is the perfect design, never change it"
 */

import React, { Suspense, lazy } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { useAuth } from '@/hooks/useAuth';
import { useTheme } from '@/hooks/useTheme';
import Layout from '@/components/layout/Layout';
import SimpleLoginForm from '@/components/auth/SimpleLoginForm';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

// Lazy load heavy components to reduce initial bundle size
const Dashboard = lazy(() => import('@/components/pages/Dashboard'));
const Send = lazy(() => import('@/components/pages/Send'));
const View = lazy(() => import('@/components/pages/View'));
const SecureEmailPage = lazy(() => import('@/components/secure/SecureEmailPage'));
const MonitoringDashboard = lazy(() => import('@/components/dashboard/MonitoringDashboard'));
const SecureEmailViewer = lazy(() => import('@/components/external/SecureEmailViewer'));

// Auth components
const SignupPage = lazy(() => import('@/pages/SignupPage'));

// Admin components
const AdminLogin = lazy(() => import('@/pages/admin/Login'));
const AdminDashboard = lazy(() => import('@/pages/admin/Dashboard'));
const RequireAdminAuth = lazy(() => import('@/components/admin/RequireAdminAuth'));

// Loading component for Suspense fallback
const LoadingSpinner: React.FC = () => (
  <div className="min-h-screen bg-secondary-50 dark:bg-secondary-900 flex items-center justify-center">
    <div className="text-center">
      <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto mb-4"></div>
      <p className="text-secondary-600 dark:text-secondary-400">Loading...</p>
    </div>
  </div>
);

/**
 * ProtectedRoute Component
 * 
 * A higher-order component that protects routes requiring authentication.
 * It checks the user's authentication status and either renders the protected
 * content or redirects to the login page.
 * 
 * Features:
 * - Shows loading spinner while checking authentication status
 * - Redirects unauthenticated users to /login
 * - Renders protected content for authenticated users
 */
const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingSpinner />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
};

/**
 * LoginPage Component
 * 
 * The main login page that displays the authentication form.
 * Features a modern design with gradient background, brand logo,
 * and the login form component.
 * 
 * Design:
 * - Gradient background with dark/light theme support
 * - Centered layout with brand identity
 * - Responsive design for all screen sizes
 */
const LoginPage: React.FC = () => {
  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-white to-secondary-50 dark:from-secondary-900 dark:via-secondary-800 dark:to-secondary-900 flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <div className="text-center mb-8">
          <div className="mx-auto w-20 h-20 bg-primary-100 dark:bg-primary-900/20 rounded-full flex items-center justify-center mb-6">
            <svg className="h-10 w-10 text-primary-600 dark:text-primary-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
            </svg>
          </div>
          <h1 className="text-3xl font-bold text-secondary-900 dark:text-white mb-2">
            SecureMail
          </h1>
          <p className="text-secondary-600 dark:text-secondary-400">
            Your privacy-first email solution
          </p>
        </div>
        
        <SimpleLoginForm />
      </div>
    </div>
  );
};

/**
 * App Component
 * 
 * The root component of the Secure Email application.
 * Handles routing, theme management, and global state.
 * 
 * Features:
 * - React Router for navigation
 * - Theme switching (dark/light mode)
 * - Protected routes for authenticated users
 * - Toast notifications for user feedback
 * - Responsive layout with Tailwind CSS
 * 
 * Route Structure:
 * - /login: Public authentication page
 * - /inbox: Protected email inbox
 * - /sent: Protected sent emails
 * - /drafts: Protected draft emails
 * - /trash: Protected deleted emails
 */
const App: React.FC = () => {
  const { theme } = useTheme();

  return (
    <div className={theme === 'dark' ? 'dark' : ''}>
      <Router>
        <div className="min-h-screen bg-secondary-50 dark:bg-secondary-900">
          <Routes>
            {/* Public Routes */}
            <Route path="/login" element={<LoginPage />} />
            <Route path="/signup" element={
              <Suspense fallback={<LoadingSpinner />}>
                <SignupPage />
              </Suspense>
            } />
            <Route path="/" element={<Navigate to="/secure" replace />} />
            <Route path="/secure" element={<SecureEmailPage />} />
            
            {/* Secure Link Viewer - Public route for external recipients */}
            <Route path="/v/:linkID" element={
              <Suspense fallback={<LoadingSpinner />}>
                <SecureEmailViewer />
              </Suspense>
            } />
            
            {/* Protected Routes */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={<Dashboard />} />
            </Route>
            
            <Route
              path="/send"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Send />
                </Suspense>
              } />
            </Route>
            
            <Route
              path="/view/:id"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
              <Suspense fallback={<LoadingSpinner />}>
                <View />
              </Suspense>
            } />
            </Route>
            
            {/* Legacy routes for backward compatibility */}
            <Route
              path="/inbox"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Dashboard />
                </Suspense>
              } />
            </Route>
            
            <Route
              path="/sent"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Dashboard />
                </Suspense>
              } />
            </Route>
            
            <Route
              path="/drafts"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Dashboard />
                </Suspense>
              } />
            </Route>
            
            <Route
              path="/trash"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Dashboard />
                </Suspense>
              } />
            </Route>
            
            <Route
              path="/settings"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <Dashboard />
                </Suspense>
              } />
            </Route>
            
            <Route
              path="/monitoring"
              element={
                <ProtectedRoute>
                  <Layout />
                </ProtectedRoute>
              }
            >
              <Route index element={
                <Suspense fallback={<LoadingSpinner />}>
                  <MonitoringDashboard />
                </Suspense>
              } />
            </Route>
            
            {/* Admin routes */}
            <Route path="/admin/login" element={
              <Suspense fallback={<LoadingSpinner />}>
                <AdminLogin />
              </Suspense>
            } />
            
            <Route path="/admin/dashboard" element={
              <Suspense fallback={<LoadingSpinner />}>
                <RequireAdminAuth>
                  <AdminDashboard />
                </RequireAdminAuth>
              </Suspense>
            } />
            
            {/* Catch all route */}
            <Route path="*" element={<Navigate to="/dashboard" replace />} />
          </Routes>
        </div>
      </Router>
      
      {/* Toast Notifications */}
      <ToastContainer
        position="top-right"
        autoClose={5000}
        hideProgressBar={false}
        newestOnTop={false}
        closeOnClick
        rtl={false}
        pauseOnFocusLoss
        draggable
        pauseOnHover
        theme={theme === 'dark' ? 'dark' : 'light'}
      />
    </div>
  );
};

export default App; 