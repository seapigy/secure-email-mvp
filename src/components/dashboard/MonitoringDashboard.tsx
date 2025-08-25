import React, { useState, useEffect } from 'react';
import Button from '@/components/ui/Button';
import { 
  Activity, 
  AlertTriangle, 
  BarChart3, 
  Clock, 
  RefreshCw, 
  CheckCircle
} from 'lucide-react';
import { useMonitoringStore } from '@/stores/monitoringStore';
import MetricsChart from './MetricsChart';

interface MetricsData {
  request_count: number;
  error_rate: number;
  average_latency: number;
  active_sessions: number;
  dlp_scans: number;
  watermarking_ops: number;
  security_alerts: number;
  last_updated: string;
  source_breakdown: Record<string, number>;
  error_breakdown: Record<string, number>;
}

interface HealthData {
  status: string;
  message: string;
  timestamp: string;
  request_count: number;
  error_rate: number;
  avg_latency: number;
  active_sessions: number;
}

const MonitoringDashboard: React.FC = () => {
  const [currentMetrics, setCurrentMetrics] = useState<MetricsData | null>(null);
  const [systemHealth, setSystemHealth] = useState<HealthData | null>(null);
  const [isConnected, setIsConnected] = useState(false);


  const { subscribeToStream } = useMonitoringStore();

  useEffect(() => {
    fetchMetrics();
    fetchHealth();
    
    // Connect to SSE stream
    const clientId = `dashboard_${Date.now()}`;
    subscribeToStream(clientId, (data: any) => {
      if (data.type === 'metrics_update') {
        setCurrentMetrics(data.data);
      } else if (data.type === 'health_update') {
        setSystemHealth(data.data);
      }
    });

    return () => {
      // Cleanup will be handled by the store
    };
  }, [subscribeToStream]);

  const fetchMetrics = async () => {
    try {
      const response = await fetch('/api/metrics');
      if (response.ok) {
        const data = await response.json();
        setCurrentMetrics(data.metrics);
        setIsConnected(true);
      } else {
        throw new Error(`HTTP ${response.status}`);
      }
          } catch (error) {
        console.error('Failed to fetch metrics:', error);
        setIsConnected(false);
      }
  };

  const fetchHealth = async () => {
    try {
      const response = await fetch('/api/metrics/health');
      if (response.ok) {
        const data = await response.json();
        setSystemHealth(data.health);
      }
    } catch (error) {
      console.error('Failed to fetch health:', error);
    }
  };

  // Generate chart data
  const generateChartData = () => {
    const now = new Date();
    const data = [];
    for (let i = 0; i < 10; i++) {
      const time = new Date(now.getTime() - (9 - i) * 60000);
      data.push({
        name: time.toLocaleTimeString(),
        value: Math.floor(Math.random() * 100) + 10
      });
    }
    return data;
  };

  const chartData = generateChartData();
  const errorData = [
    { name: '4xx Errors', value: 15 },
    { name: '5xx Errors', value: 5 },
    { name: 'Timeouts', value: 3 },
    { name: 'Network', value: 2 }
  ];
  const dlpData = [
    { name: 'PII Scans', value: currentMetrics?.dlp_scans || 0 },
    { name: 'Financial', value: 12 },
    { name: 'Healthcare', value: 8 },
    { name: 'Legal', value: 5 }
  ];
  const securityData = [
    { name: 'Failed Logins', value: currentMetrics?.security_alerts || 0 },
    { name: 'Suspicious IPs', value: 3 },
    { name: 'Rate Limits', value: 7 },
    { name: 'Policy Violations', value: 2 }
  ];

  const getHealthStatusIcon = (status: string) => {
    switch (status) {
      case 'healthy':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'warning':
        return <AlertTriangle className="h-5 w-5 text-yellow-500" />;
      case 'error':
        return <AlertTriangle className="h-5 w-5 text-red-500" />;
      default:
        return <Activity className="h-5 w-5 text-gray-500" />;
    }
  };

  return (
    <div className="p-6 space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">System Monitoring Dashboard</h1>
          <p className="text-gray-600 mt-2">Real-time metrics and system health monitoring</p>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <div className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
            <span className="text-sm text-gray-600">
              {isConnected ? 'Live Updates' : 'Polling Mode'}
            </span>
          </div>
          <Button onClick={fetchMetrics} variant="primary" size="sm">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* System Health Status */}
      {systemHealth && (
        <div className="bg-white p-6 rounded-lg shadow border-l-4 border-l-blue-500">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center space-x-2">
              {getHealthStatusIcon(systemHealth.status)}
              <span className="font-semibold">System Health</span>
              <span className={`px-2 py-1 rounded text-sm ${
                systemHealth.status === 'healthy' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
              }`}>
                {systemHealth.status.toUpperCase()}
              </span>
            </div>
            <span className="text-sm text-gray-500">
              Last updated: {new Date(systemHealth.timestamp).toLocaleTimeString()}
            </span>
          </div>
          <p className="text-gray-700">{systemHealth.message}</p>
        </div>
      )}

      {/* Key Metrics Grid */}
      {currentMetrics && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium">Request Rate</h3>
              <Activity className="h-4 w-4 text-gray-400" />
            </div>
            <div className="text-2xl font-bold">{currentMetrics.request_count}</div>
            <p className="text-xs text-gray-500">requests/min</p>
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium">Error Rate</h3>
              <AlertTriangle className="h-4 w-4 text-gray-400" />
            </div>
            <div className="text-2xl font-bold text-red-600">
              {currentMetrics.error_rate.toFixed(2)}%
            </div>
            <p className="text-xs text-gray-500">of total requests</p>
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium">Avg Latency</h3>
              <Clock className="h-4 w-4 text-gray-400" />
            </div>
            <div className="text-2xl font-bold">{currentMetrics.average_latency}ms</div>
            <p className="text-xs text-gray-500">response time</p>
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-sm font-medium">Active Sessions</h3>
              <BarChart3 className="h-4 w-4 text-gray-400" />
            </div>
            <div className="text-2xl font-bold">{currentMetrics.active_sessions}</div>
            <p className="text-xs text-gray-500">concurrent users</p>
          </div>
        </div>
      )}

      {/* Charts */}
      <div className="space-y-6">
        <div className="bg-white p-6 rounded-lg shadow">
          <h3 className="text-lg font-semibold mb-4">Request Trends</h3>
          <MetricsChart
            type="line"
            data={chartData}
            title="Requests per Minute"
            height={300}
          />
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-semibold mb-4">Error Distribution</h3>
            <MetricsChart
              type="pie"
              data={errorData}
              title="Error Types"
              height={300}
            />
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-semibold mb-4">DLP Activity</h3>
            <MetricsChart
              type="bar"
              data={dlpData}
              title="DLP Scans"
              height={300}
            />
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-semibold mb-4">Security Events</h3>
            <MetricsChart
              type="area"
              data={securityData}
              title="Security Alerts"
              height={300}
            />
          </div>

          <div className="bg-white p-6 rounded-lg shadow">
            <h3 className="text-lg font-semibold mb-4">Watermarking Operations</h3>
            <MetricsChart
              type="bar"
              data={[
                { name: 'Text', value: currentMetrics?.watermarking_ops || 0 },
                { name: 'Image', value: 8 },
                { name: 'Audio', value: 3 },
                { name: 'Video', value: 2 }
              ]}
              title="Watermarking Activity"
              height={300}
            />
          </div>
        </div>
      </div>

      {/* Connection Status */}
      <div className="fixed bottom-4 right-4">
        <div className={`flex items-center space-x-2 px-3 py-2 rounded-lg shadow-lg ${
          isConnected ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
        }`}>
          <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`} />
          <span className="text-sm font-medium">
            {isConnected ? 'Live Updates' : 'Disconnected'}
          </span>
        </div>
      </div>
    </div>
  );
};

export default MonitoringDashboard;
