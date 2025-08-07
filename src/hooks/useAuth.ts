import { useEffect } from 'react';
import { useAuthStore } from '@/stores/authStore';

export function useAuth() {
  const {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    signup,
    logout,
    refreshToken,
    setUser,
    setLoading,
    setError,
    clearError,
    updateUser,
  } = useAuthStore();

  // Auto-refresh token when needed
  useEffect(() => {
    if (isAuthenticated) {
      const checkTokenExpiry = () => {
        const accessToken = sessionStorage.getItem('accessToken');
        if (accessToken) {
          try {
            const payload = JSON.parse(atob(accessToken.split('.')[1]));
            const expiryTime = payload.exp * 1000; // Convert to milliseconds
            const currentTime = Date.now();
            const timeUntilExpiry = expiryTime - currentTime;
            
            // Refresh token if it expires in less than 5 minutes
            if (timeUntilExpiry < 5 * 60 * 1000) {
              refreshToken();
            }
          } catch (error) {
            console.error('Error parsing token:', error);
          }
        }
      };

      // Check token expiry every minute
      const interval = setInterval(checkTokenExpiry, 60 * 1000);
      checkTokenExpiry(); // Check immediately

      return () => clearInterval(interval);
    }
  }, [isAuthenticated, refreshToken]);

  // Check for existing session on mount
  useEffect(() => {
    const checkExistingSession = async () => {
      const accessToken = sessionStorage.getItem('accessToken');
      if (accessToken && !isAuthenticated) {
        try {
          const { getUserProfile } = await import('@/lib/api');
          const user = await getUserProfile();
          setUser(user);
        } catch (error) {
          console.error('Error checking session:', error);
          sessionStorage.removeItem('accessToken');
          sessionStorage.removeItem('refreshToken');
        }
      }
    };

    checkExistingSession();
  }, [isAuthenticated, setUser]);

  return {
    user,
    isAuthenticated,
    isLoading,
    error,
    login,
    signup,
    logout,
    refreshToken,
    setUser,
    setLoading,
    setError,
    clearError,
    updateUser,
  };
} 