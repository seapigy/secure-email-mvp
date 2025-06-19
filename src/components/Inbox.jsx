import React from 'react';
import { useNavigate } from 'react-router-dom';

const Inbox = () => {
  const navigate = useNavigate();

  const handleLogout = () => {
    sessionStorage.removeItem('token');
    sessionStorage.removeItem('user_id');
    navigate('/login');
  };

  return (
    <div className="min-h-screen bg-neutral-light dark:bg-neutral-dark p-4">
      <div className="max-w-4xl mx-auto">
        <div className="bg-white dark:bg-gray-800 rounded-xl shadow-lg p-8">
          <div className="text-center">
            <h1 className="text-3xl font-bold text-primary dark:text-blue-400 mb-4">
              Inbox Placeholder
            </h1>
            <p className="text-text-gray dark:text-gray-300 mb-8">
              This is a placeholder for the inbox. The full inbox functionality will be implemented in future iterations.
            </p>
            <div className="space-y-4">
              <p className="text-sm text-text-gray dark:text-gray-400">
                JWT Token: {sessionStorage.getItem('token') ? 'Present' : 'Not found'}
              </p>
              <p className="text-sm text-text-gray dark:text-gray-400">
                User ID: {sessionStorage.getItem('user_id') || 'Not found'}
              </p>
              <button
                onClick={handleLogout}
                className="btn-primary"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Inbox; 