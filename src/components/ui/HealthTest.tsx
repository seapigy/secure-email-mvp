import React from 'react';
import { healthCheck } from '@/lib/api';

/**
 * HealthTest Component
 * 
 * Simple test component to verify health check functionality.
 * This component can be used for debugging and testing.
 */
const HealthTest: React.FC = () => {
  const [testResult, setTestResult] = React.useState<string>('');
  const [isTesting, setIsTesting] = React.useState(false);

  const runTest = async () => {
    setIsTesting(true);
    setTestResult('Testing...');
    
    try {
      const response = await healthCheck();
      setTestResult(`✅ Success: ${JSON.stringify(response, null, 2)}`);
    } catch (error) {
      setTestResult(`❌ Error: ${error instanceof Error ? error.message : 'Unknown error'}`);
    } finally {
      setIsTesting(false);
    }
  };

  return (
    <div className="bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 p-4">
      <h3 className="text-lg font-semibold text-secondary-900 dark:text-white mb-4">
        Health Check Test
      </h3>
      
      <button
        onClick={runTest}
        disabled={isTesting}
        className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors duration-200 mb-4"
      >
        {isTesting ? 'Testing...' : 'Run Health Check Test'}
      </button>
      
      {testResult && (
        <div className="mt-4">
          <h4 className="text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
            Test Result:
          </h4>
          <pre className="bg-secondary-50 dark:bg-secondary-700 p-3 rounded text-xs overflow-auto">
            {testResult}
          </pre>
        </div>
      )}
    </div>
  );
};

export default HealthTest; 