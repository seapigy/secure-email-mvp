import { useState, useEffect, useCallback } from 'react';
import { healthCheck, HealthCheckResponse } from '@/lib/api';

export interface HealthStatus {
  isHealthy: boolean;
  isLoading: boolean;
  error: string | null;
  lastChecked: Date | null;
  response: HealthCheckResponse | null;
}

export const useHealthCheck = (autoCheck: boolean = true) => {
  const [status, setStatus] = useState<HealthStatus>({
    isHealthy: false,
    isLoading: false,
    error: null,
    lastChecked: null,
    response: null,
  });

  const checkHealth = useCallback(async () => {
    setStatus(prev => ({
      ...prev,
      isLoading: true,
      error: null,
    }));

    try {
      const response = await healthCheck();
      setStatus({
        isHealthy: response.status === 'ok',
        isLoading: false,
        error: null,
        lastChecked: new Date(),
        response,
      });
    } catch (error) {
      let errorMessage = 'Health check failed';
      
      if (error instanceof Error) {
        errorMessage = error.message;
      } else if (typeof error === 'object' && error !== null && 'message' in error) {
        errorMessage = (error as any).message;
      }
      
      setStatus({
        isHealthy: false,
        isLoading: false,
        error: errorMessage,
        lastChecked: new Date(),
        response: null,
      });
    }
  }, []);

  // Auto-check health on mount if enabled
  useEffect(() => {
    if (autoCheck) {
      checkHealth();
    }
  }, [autoCheck, checkHealth]);

  return {
    ...status,
    checkHealth,
  };
}; 