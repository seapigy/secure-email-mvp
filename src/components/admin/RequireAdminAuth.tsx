import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { isAdminAuthenticated } from '../../lib/adminApi';

interface RequireAdminAuthProps {
  children: React.ReactNode;
}

const RequireAdminAuth: React.FC<RequireAdminAuthProps> = ({ children }) => {
  const navigate = useNavigate();
  const [isChecking, setIsChecking] = useState(true);

  useEffect(() => {
    const checkAuth = () => {
      if (!isAdminAuthenticated()) {
        navigate('/admin/login');
        return;
      }
      setIsChecking(false);
    };

    checkAuth();
  }, [navigate]);

  if (isChecking) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Verifying admin access...</p>
        </div>
      </div>
    );
  }

  return <>{children}</>;
};

export default RequireAdminAuth;







