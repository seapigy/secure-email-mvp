import React, { useState, useEffect, useCallback } from 'react';
import {
  ArrowPathIcon, Cog6ToothIcon, BellIcon, ShieldCheckIcon, ExclamationTriangleIcon,
  XMarkIcon, EyeIcon, UserGroupIcon,
} from '@heroicons/react/24/outline';
import { EnterpriseDashboardService } from '../../services/enterpriseDashboardService';
import {
  ZKIDLayerMetrics, PQCEncryptionMetrics, EmailDeliveryMetrics, SecurityComplianceMetrics,
  PerformanceOperationalMetrics, Alert, AuditLogEntry, DashboardConfig, RealTimeUpdate,
} from '../../types/admin';

// Import all monitoring panels
import ZKIDLayerPanel from './panels/ZKIDLayerPanel';
import PQCEncryptionPanel from './panels/PQCEncryptionPanel';
import EmailDeliveryPanel from './panels/EmailDeliveryPanel';
import SecurityCompliancePanel from './panels/SecurityCompliancePanel';
import PerformanceOperationalPanel from './panels/PerformanceOperationalPanel';
import AlertsPanel from './panels/AlertsPanel';
import AuditLogsPanel from './panels/AuditLogsPanel';
import AdminManagementPanel from './panels/AdminManagementPanel';
import AnalyticsDashboardPanel from './panels/AnalyticsDashboardPanel';
import ThreatAwarenessPanel from './panels/ThreatAwarenessPanel';

interface EnterpriseDashboardProps {
  adminToken: string;
}

const EnterpriseDashboard: React.FC<EnterpriseDashboardProps> = ({ adminToken }) => {
  // State for dashboard data
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshInterval, setRefreshInterval] = useState(30);
  const [showSettings, setShowSettings] = useState(false);
  const [showNotifications, setShowNotifications] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'connected' | 'disconnected' | 'reconnecting'>('connected');

  // State for metrics data
  const [zkidMetrics, setZkidMetrics] = useState<ZKIDLayerMetrics | null>(null);
  const [pqcMetrics, setPqcMetrics] = useState<PQCEncryptionMetrics | null>(null);
  const [emailMetrics, setEmailMetrics] = useState<EmailDeliveryMetrics | null>(null);
  const [securityMetrics, setSecurityMetrics] = useState<SecurityComplianceMetrics | null>(null);
  const [performanceMetrics, setPerformanceMetrics] = useState<PerformanceOperationalMetrics | null>(null);
  const [alerts, setAlerts] = useState<Alert[]>([]);
  const [auditLogs, setAuditLogs] = useState<AuditLogEntry[]>([]);
  const [dashboardConfig, setDashboardConfig] = useState<DashboardConfig | null>(null);

  // State for real-time updates
  const [realTimeUpdates, setRealTimeUpdates] = useState<RealTimeUpdate[]>([]);
  const [lastUpdate, setLastUpdate] = useState<Date>(new Date());

  const dashboardService = new EnterpriseDashboardService(adminToken);

  const fetchDashboardData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);

      // Fetch all metrics in parallel
      const [
        zkidData,
        pqcData,
        emailData,
        securityData,
        performanceData,
        alertsData,
        logsData,
        configData,
      ] = await Promise.all([
        dashboardService.getZKIDMetrics(),
        dashboardService.getPQCEncryptionMetrics(),
        dashboardService.getEmailDeliveryMetrics(),
        dashboardService.getSecurityComplianceMetrics(),
        dashboardService.getPerformanceOperationalMetrics(),
        dashboardService.getAlerts(),
        dashboardService.getAuditLogs(),
        dashboardService.getDashboardConfig(),
      ]);

      setZkidMetrics(zkidData);
      setPqcMetrics(pqcData);
      setEmailMetrics(emailData);
      setSecurityMetrics(securityData);
      setPerformanceMetrics(performanceData);
      setAlerts(alertsData);
      setAuditLogs(logsData);
      setDashboardConfig(configData);
      setLastUpdate(new Date());
      setConnectionStatus('connected');
    } catch (error: any) {
      setError(error.message || 'Failed to fetch dashboard data');
      setConnectionStatus('disconnected');
    } finally {
      setIsLoading(false);
    }
  }, [dashboardService]);

  // Initial data fetch
  useEffect(() => {
    fetchDashboardData();
  }, [fetchDashboardData]);

  // Auto-refresh setup
  useEffect(() => {
    if (!dashboardConfig?.auto_refresh_enabled) return;

    const interval = setInterval(() => {
      fetchDashboardData();
    }, refreshInterval * 1000);

    return () => clearInterval(interval);
  }, [refreshInterval, dashboardConfig?.auto_refresh_enabled, fetchDashboardData]);

  // Real-time updates setup
  useEffect(() => {
    const handleRealTimeUpdate = (update: RealTimeUpdate) => {
      setRealTimeUpdates(prev => [update, ...prev.slice(0, 9)]); // Keep last 10 updates
    };

    dashboardService.startRealTimeUpdates(handleRealTimeUpdate);

    return () => {
      dashboardService.stopRealTimeUpdates();
    };
  }, [dashboardService]);

  // Alert management handlers
  const handleDismissAlert = async (alertId: string) => {
    try {
      await dashboardService.acknowledgeAlert(alertId);
      setAlerts(prev => prev.map(alert => 
        alert.id === alertId 
          ? { ...alert, acknowledged: true, acknowledged_at: new Date().toISOString() }
          : alert
      ));
    } catch (error: any) {
      setError(error.message || 'Failed to dismiss alert');
    }
  };

  const handleResolveAlert = async (alertId: string) => {
    try {
      await dashboardService.resolveAlert(alertId);
      setAlerts(prev => prev.map(alert => 
        alert.id === alertId 
          ? { ...alert, resolved: true, resolved_at: new Date().toISOString() }
          : alert
      ));
    } catch (error: any) {
      setError(error.message || 'Failed to resolve alert');
    }
  };

  // Get current user and permissions
  const currentUser = dashboardService.getCurrentUser();
  const isReadOnlyAdmin = dashboardService.isReadOnlyAdmin();
  const canManageAdmins = dashboardService.canManageAdmins();
  const canViewSensitiveData = dashboardService.canViewSensitiveData();

  // Calculate system health status
  const getSystemHealthStatus = () => {
    const criticalAlerts = alerts.filter(alert => alert.severity === 'critical' && !alert.resolved);
    const highAlerts = alerts.filter(alert => alert.severity === 'high' && !alert.resolved);
    
    if (criticalAlerts.length > 0) return 'critical';
    if (highAlerts.length > 0) return 'warning';
    return 'healthy';
  };

  const systemHealthStatus = getSystemHealthStatus();

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <ShieldCheckIcon className="h-8 w-8 text-indigo-600" />
                <div>
                  <h1 className="text-xl font-semibold text-gray-900">Enterprise Admin Dashboard</h1>
                  <p className="text-sm text-gray-500">Multi-Admin Secure Email MVP Monitoring</p>
                </div>
              </div>
              
              {/* System Health Indicator */}
              <div className="flex items-center space-x-2">
                <div className={`w-3 h-3 rounded-full ${
                  systemHealthStatus === 'healthy' ? 'bg-green-500' :
                  systemHealthStatus === 'warning' ? 'bg-yellow-500' : 'bg-red-500'
                }`} />
                <span className="text-sm font-medium text-gray-700 capitalize">{systemHealthStatus}</span>
              </div>

              {/* Connection Status */}
              <div className="flex items-center space-x-2">
                <div className={`w-2 h-2 rounded-full ${
                  connectionStatus === 'connected' ? 'bg-green-500' :
                  connectionStatus === 'reconnecting' ? 'bg-yellow-500' : 'bg-red-500'
                }`} />
                <span className="text-xs text-gray-500 capitalize">{connectionStatus}</span>
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {/* Current User Info */}
              {currentUser && (
                <div className="flex items-center space-x-2 text-sm text-gray-700">
                  <UserGroupIcon className="h-4 w-4" />
                  <span>{currentUser.username}</span>
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    currentUser.role === 'primary_admin' ? 'bg-purple-100 text-purple-800' :
                    currentUser.role === 'full_admin' ? 'bg-blue-100 text-blue-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {currentUser.role.replace('_', ' ')}
                  </span>
                </div>
              )}

              {/* Alerts Badge */}
              <div className="relative">
                <button
                  onClick={() => setShowNotifications(!showNotifications)}
                  className="p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
                >
                  <BellIcon className="h-6 w-6" />
                  {alerts.filter(alert => !alert.resolved).length > 0 && (
                    <span className="absolute -top-1 -right-1 h-5 w-5 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                      {alerts.filter(alert => !alert.resolved).length}
                    </span>
                  )}
                </button>
              </div>

              {/* Settings */}
              <button
                onClick={() => setShowSettings(!showSettings)}
                className="p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
              >
                <Cog6ToothIcon className="h-6 w-6" />
              </button>

              {/* Refresh */}
              <button
                onClick={fetchDashboardData}
                disabled={isLoading}
                className="p-2 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded disabled:opacity-50"
              >
                <ArrowPathIcon className={`h-6 w-6 ${isLoading ? 'animate-spin' : ''}`} />
              </button>
            </div>
          </div>
        </div>
      </header>

      {/* Error Banner */}
      {error && (
        <div className="bg-red-50 border-b border-red-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <ExclamationTriangleIcon className="h-5 w-5 text-red-400" />
                <span className="text-sm text-red-700">{error}</span>
              </div>
              <button
                onClick={() => setError(null)}
                className="text-red-400 hover:text-red-600 focus:outline-none focus:ring-2 focus:ring-red-500 rounded"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Settings Panel */}
      {showSettings && (
        <div className="bg-white border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <label className="flex items-center space-x-2">
                  <span className="text-sm font-medium text-gray-700">Auto-refresh:</span>
                  <select
                    value={refreshInterval}
                    onChange={(e) => setRefreshInterval(parseInt(e.target.value))}
                    className="text-sm border border-gray-300 rounded-md px-2 py-1"
                  >
                    <option value={10}>10s</option>
                    <option value={30}>30s</option>
                    <option value={60}>1m</option>
                    <option value={300}>5m</option>
                  </select>
                </label>
              </div>
              <div className="text-sm text-gray-500">
                Last updated: {lastUpdate.toLocaleTimeString()}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Notifications Panel */}
      {showNotifications && (
        <div className="bg-white border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
            <h3 className="text-sm font-medium text-gray-900 mb-3">Recent Alerts</h3>
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {alerts.filter(alert => !alert.resolved).slice(0, 5).map((alert) => (
                <div key={alert.id} className="flex items-center justify-between p-2 bg-gray-50 rounded">
                  <div>
                    <p className="text-sm font-medium text-gray-900">{alert.title}</p>
                    <p className="text-xs text-gray-500">{alert.description}</p>
                  </div>
                  <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                    alert.severity === 'critical' ? 'bg-red-100 text-red-800' :
                    alert.severity === 'high' ? 'bg-orange-100 text-orange-800' :
                    alert.severity === 'medium' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-blue-100 text-blue-800'
                  }`}>
                    {alert.severity}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Main Dashboard Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
          {/* ZKID Layer Panel */}
          <div className="lg:col-span-1">
            <ZKIDLayerPanel metrics={zkidMetrics} isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
          </div>

          {/* PQC Encryption Panel */}
          <div className="lg:col-span-1">
            <PQCEncryptionPanel metrics={pqcMetrics} isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
          </div>

          {/* Email Delivery Panel */}
          <div className="lg:col-span-1">
            <EmailDeliveryPanel metrics={emailMetrics} isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
          </div>

          {/* Security & Compliance Panel */}
          <div className="lg:col-span-1">
            <SecurityCompliancePanel metrics={securityMetrics} isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
          </div>

          {/* Performance & Operational Panel */}
          <div className="lg:col-span-1">
            <PerformanceOperationalPanel metrics={performanceMetrics} isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
          </div>

          {/* Alerts Panel */}
          <div className="lg:col-span-1">
            <AlertsPanel 
              alerts={alerts} 
              isLoading={isLoading} 
              onDismissAlert={handleDismissAlert} 
              onResolveAlert={handleResolveAlert} 
              onRefresh={() => fetchDashboardData()} 
            />
          </div>

          {/* Admin Management Panel - Only visible to primary admins */}
          {canManageAdmins && (
            <div className="lg:col-span-1">
              <AdminManagementPanel isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
            </div>
          )}
        </div>

        {/* Analytics Dashboard Panel - Full Width */}
        <div className="mt-6">
          <AnalyticsDashboardPanel dashboardService={dashboardService} isReadOnly={isReadOnlyAdmin} />
        </div>

        {/* Threat Awareness Panel - Full Width */}
        <div className="mt-6">
          <ThreatAwarenessPanel dashboardService={dashboardService} isReadOnly={isReadOnlyAdmin} />
        </div>

        {/* Audit Logs Panel - Full Width */}
        <div className="mt-6">
          <AuditLogsPanel logs={auditLogs} isLoading={isLoading} onRefresh={() => fetchDashboardData()} />
        </div>

        {/* Real-time Updates */}
        {realTimeUpdates.length > 0 && (
          <div className="mt-6 bg-white rounded-lg shadow">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Real-time Updates</h3>
            </div>
            <div className="p-6">
              <div className="space-y-2 max-h-32 overflow-y-auto">
                {realTimeUpdates.map((update, index) => (
                  <div key={index} className="flex items-center justify-between text-sm">
                    <span className="text-gray-600">{update.type}</span>
                    <span className="text-gray-400">{new Date(update.timestamp).toLocaleTimeString()}</span>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Permission-based Information */}
        {isReadOnlyAdmin && (
          <div className="mt-6 bg-blue-50 rounded-lg p-4">
            <div className="flex items-center space-x-2">
              <EyeIcon className="h-5 w-5 text-blue-600" />
              <div>
                <h4 className="text-sm font-medium text-blue-900">Read-Only Mode</h4>
                <p className="text-sm text-blue-700">
                  You are in read-only mode. Contact your primary administrator for elevated permissions.
                </p>
              </div>
            </div>
          </div>
        )}

        {/* Zero-Knowledge Privacy Notice */}
        <div className="mt-6 bg-indigo-50 rounded-lg p-4">
          <div className="flex items-center space-x-2">
            <ShieldCheckIcon className="h-5 w-5 text-indigo-600" />
            <div>
              <h4 className="text-sm font-medium text-indigo-900">Zero-Knowledge Privacy</h4>
              <p className="text-sm text-indigo-700">
                All operations use UUID-only identifiers. No external emails or personal data are visible to administrators.
                {!canViewSensitiveData && ' Your access is restricted to non-sensitive metrics only.'}
              </p>
            </div>
          </div>
        </div>
      </main>
    </div>
  );
};

export default EnterpriseDashboard;
