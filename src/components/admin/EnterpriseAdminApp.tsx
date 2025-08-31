import React, { useState, useEffect } from 'react';
import { log } from '@/lib/logger';
import EnterpriseAdminLogin from './EnterpriseAdminLogin';
import EnterpriseDashboard from './EnterpriseDashboard';


const EnterpriseAdminApp: React.FC = () => {
  const [adminToken, setAdminToken] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);


  useEffect(() => {
    // Check for stored admin token on component mount
    const storedToken = localStorage.getItem('enterprise_admin_token');
    if (storedToken) {
      setAdminToken(storedToken);
    }
    setIsLoading(false);
  }, []);

  // Listen for auth errors from the service
  useEffect(() => {
    const handleAuthError = () => {
      log.error('Admin authentication error detected', null, 'EnterpriseAdminApp');
      handleLogout();
    };

    window.addEventListener('adminAuthError', handleAuthError);
    return () => {
      window.removeEventListener('adminAuthError', handleAuthError);
    };
  }, []);

  const handleLogin = (token: string, user: unknown) => {
    setAdminToken(token);
    localStorage.setItem('enterprise_admin_token', token);
    // User data can be used if needed
    log.info('User logged in', user, 'EnterpriseAdminApp');
  };

  const handleLogout = () => {
    setAdminToken(null);
    // MFA setup cleared
    localStorage.removeItem('enterprise_admin_token');
  };



  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto"></div>
          <p className="mt-4 text-gray-600">Initializing Enterprise Admin Dashboard...</p>
        </div>
      </div>
    );
  }

  if (!adminToken) {
    return (
      <EnterpriseAdminLogin
        onLoginSuccess={handleLogin}
      />
    );
  }

  return (
    <EnterpriseDashboard
      adminToken={adminToken}
    />
  );
};

export default EnterpriseAdminApp;
