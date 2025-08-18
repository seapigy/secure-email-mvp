import React, { useState } from 'react';
import {
  BellIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  EyeIcon,
  EyeSlashIcon,
  XMarkIcon,

  CheckIcon,
} from '@heroicons/react/24/outline';
import { Alert } from '../../../types/admin';

interface AlertsPanelProps {
  alerts: Alert[];
  isLoading: boolean;
  onDismissAlert: (alertId: string) => void;
  onResolveAlert: (alertId: string) => void;
  onRefresh: () => void;
}

const AlertsPanel: React.FC<AlertsPanelProps> = ({
  alerts,
  isLoading,
  onDismissAlert,
  onResolveAlert,
  onRefresh,
}) => {
  const [showDismissed, setShowDismissed] = useState(false);
  const [showEnhancedAlerts, setShowEnhancedAlerts] = useState(false);
  const [notificationSettings, setNotificationSettings] = useState({
    email: true,
    webhook: false,
    dashboard: true,
    slack: false,
  });

  const activeAlerts = alerts.filter(alert => !alert.resolved);
  const dismissedAlerts = alerts.filter(alert => alert.acknowledged && !alert.resolved);
  const resolvedAlerts = alerts.filter(alert => alert.resolved);

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'text-red-600 bg-red-100 border-red-200';
      case 'high':
        return 'text-orange-600 bg-orange-100 border-orange-200';
      case 'medium':
        return 'text-yellow-600 bg-yellow-100 border-yellow-200';
      case 'low':
        return 'text-blue-600 bg-blue-100 border-blue-200';
      default:
        return 'text-gray-600 bg-gray-100 border-gray-200';
    }
  };

  const getSeverityIcon = (severity: string) => {
    switch (severity) {
      case 'critical':
        return <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />;
      case 'high':
        return <ExclamationTriangleIcon className="h-4 w-4 text-orange-600" />;
      case 'medium':
        return <ExclamationTriangleIcon className="h-4 w-4 text-yellow-600" />;
      case 'low':
        return <BellIcon className="h-4 w-4 text-blue-600" />;
      default:
        return <BellIcon className="h-4 w-4 text-gray-600" />;
    }
  };

  const getCategoryColor = (category: string) => {
    switch (category) {
      case 'security':
        return 'bg-red-50 text-red-700';
      case 'performance':
        return 'bg-yellow-50 text-yellow-700';
      case 'system':
        return 'bg-blue-50 text-blue-700';
      case 'compliance':
        return 'bg-purple-50 text-purple-700';
      default:
        return 'bg-gray-50 text-gray-700';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-2">
            <BellIcon className="h-6 w-6 text-orange-600" />
            <h3 className="text-lg font-medium text-gray-900">Alerts</h3>
            <div className="px-2 py-1 rounded-full text-xs font-medium bg-orange-100 text-orange-800">
              {activeAlerts.length} Active
            </div>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowDismissed(!showDismissed)}
              className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
            >
              {showDismissed ? <EyeSlashIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
            </button>
            <button
              onClick={onRefresh}
              disabled={isLoading}
              className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded disabled:opacity-50"
            >
              <ArrowPathIcon className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-6">
        {/* Alert Summary */}
        <div className="mb-6">
          <div className="grid grid-cols-3 gap-4">
            <div className="bg-red-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-red-600" />
                <span className="text-xs font-medium text-red-900">Critical</span>
              </div>
              <div className="text-2xl font-bold text-red-600">
                {activeAlerts.filter(alert => alert.severity === 'critical').length}
              </div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-yellow-600" />
                <span className="text-xs font-medium text-yellow-900">Warning</span>
              </div>
              <div className="text-2xl font-bold text-yellow-600">
                {activeAlerts.filter(alert => alert.severity === 'medium' || alert.severity === 'high').length}
              </div>
            </div>
            <div className="bg-green-50 rounded-lg p-3">
              <div className="flex items-center space-x-2 mb-2">
                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                <span className="text-xs font-medium text-green-900">Resolved</span>
              </div>
              <div className="text-2xl font-bold text-green-600">
                {resolvedAlerts.length}
              </div>
            </div>
          </div>
        </div>

        {/* Active Alerts */}
        <div className="mb-6">
          <h4 className="text-sm font-medium text-gray-900 mb-3">Active Alerts</h4>
          {activeAlerts.length === 0 ? (
            <div className="text-center py-8 bg-green-50 rounded-lg">
              <CheckCircleIcon className="h-8 w-8 text-green-600 mx-auto mb-2" />
              <p className="text-sm text-green-700">No active alerts</p>
            </div>
          ) : (
            <div className="space-y-3">
              {activeAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className={`p-4 rounded-lg border ${getSeverityColor(alert.severity)}`}
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3">
                      {getSeverityIcon(alert.severity)}
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <h5 className="text-sm font-medium text-gray-900">{alert.title}</h5>
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${getCategoryColor(alert.category)}`}>
                            {alert.category}
                          </span>
                        </div>
                        <p className="text-sm text-gray-600 mb-2">{alert.description}</p>
                        <div className="flex items-center space-x-4 text-xs text-gray-500">
                          <span>{new Date(alert.timestamp).toLocaleString()}</span>
                          {alert.acknowledged && (
                            <span className="text-blue-600">
                              Acknowledged by {alert.acknowledged_by}
                            </span>
                          )}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center space-x-2">
                      {!alert.acknowledged && (
                        <button
                          onClick={() => onDismissAlert(alert.id)}
                          className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-indigo-500 rounded"
                          title="Dismiss"
                        >
                          <XMarkIcon className="h-4 w-4" />
                        </button>
                      )}
                      <button
                        onClick={() => onResolveAlert(alert.id)}
                        className="p-1 text-gray-400 hover:text-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 rounded"
                        title="Resolve"
                      >
                        <CheckIcon className="h-4 w-4" />
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Dismissed Alerts */}
        {showDismissed && dismissedAlerts.length > 0 && (
          <div className="mb-6">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Dismissed Alerts</h4>
            <div className="space-y-3">
              {dismissedAlerts.map((alert) => (
                <div
                  key={alert.id}
                  className="p-4 rounded-lg border border-gray-200 bg-gray-50"
                >
                  <div className="flex items-start justify-between">
                    <div className="flex items-start space-x-3">
                      <BellIcon className="h-4 w-4 text-gray-400" />
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-1">
                          <h5 className="text-sm font-medium text-gray-700">{alert.title}</h5>
                          <span className={`px-2 py-1 rounded-full text-xs font-medium ${getCategoryColor(alert.category)}`}>
                            {alert.category}
                          </span>
                        </div>
                        <p className="text-sm text-gray-500 mb-2">{alert.description}</p>
                        <div className="flex items-center space-x-4 text-xs text-gray-400">
                          <span>{new Date(alert.timestamp).toLocaleString()}</span>
                          <span className="text-blue-600">
                            Dismissed by {alert.acknowledged_by} at {alert.acknowledged_at && new Date(alert.acknowledged_at).toLocaleString()}
                          </span>
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={() => onResolveAlert(alert.id)}
                      className="p-1 text-gray-400 hover:text-green-600 focus:outline-none focus:ring-2 focus:ring-green-500 rounded"
                      title="Resolve"
                    >
                      <CheckIcon className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Resolved Alerts */}
        {showDismissed && resolvedAlerts.length > 0 && (
          <div>
            <h4 className="text-sm font-medium text-gray-900 mb-3">Resolved Alerts</h4>
            <div className="space-y-3">
              {resolvedAlerts.slice(0, 5).map((alert) => (
                <div
                  key={alert.id}
                  className="p-4 rounded-lg border border-green-200 bg-green-50"
                >
                  <div className="flex items-start space-x-3">
                    <CheckCircleIcon className="h-4 w-4 text-green-600" />
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-1">
                        <h5 className="text-sm font-medium text-green-900">{alert.title}</h5>
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getCategoryColor(alert.category)}`}>
                          {alert.category}
                        </span>
                      </div>
                      <p className="text-sm text-green-700 mb-2">{alert.description}</p>
                      <div className="flex items-center space-x-4 text-xs text-green-600">
                        <span>Created: {new Date(alert.timestamp).toLocaleString()}</span>
                        {alert.resolved_at && (
                          <span>Resolved: {new Date(alert.resolved_at).toLocaleString()}</span>
                        )}
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Enhanced Alert Monitoring */}
        <div className="mt-6">
          <div className="flex items-center justify-between mb-3">
            <h4 className="text-sm font-medium text-gray-900">Enhanced Alert Monitoring</h4>
            <button
              onClick={() => setShowEnhancedAlerts(!showEnhancedAlerts)}
              className="text-xs text-orange-600 hover:text-orange-800 focus:outline-none"
            >
              {showEnhancedAlerts ? 'Hide Details' : 'Show Details'}
            </button>
          </div>
          
          {showEnhancedAlerts && (
            <div className="space-y-4">
              {/* Alert Statistics */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Alert Statistics</h5>
                <div className="grid grid-cols-4 gap-3">
                  <div className="bg-red-50 rounded-lg p-3">
                    <div className="text-xs text-red-600">Critical</div>
                    <div className="text-lg font-bold text-red-900">
                      {alerts.filter(a => a.severity === 'critical' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-orange-50 rounded-lg p-3">
                    <div className="text-xs text-orange-600">High</div>
                    <div className="text-lg font-bold text-orange-900">
                      {alerts.filter(a => a.severity === 'high' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-yellow-50 rounded-lg p-3">
                    <div className="text-xs text-yellow-600">Medium</div>
                    <div className="text-lg font-bold text-yellow-900">
                      {alerts.filter(a => a.severity === 'medium' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-blue-50 rounded-lg p-3">
                    <div className="text-xs text-blue-600">Low</div>
                    <div className="text-lg font-bold text-blue-900">
                      {alerts.filter(a => a.severity === 'low' && !a.resolved).length}
                    </div>
                  </div>
                </div>
              </div>

              {/* Component-Based Alerts */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Component-Based Alerts</h5>
                <div className="grid grid-cols-3 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">ZKID</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.source_component === 'zkid' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">PQC</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.source_component === 'pqc' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Email</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.source_component === 'email' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Database</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.source_component === 'database' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Authentication</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.source_component === 'authentication' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">API</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.source_component === 'api' && !a.resolved).length}
                    </div>
                  </div>
                </div>
              </div>

              {/* Alert Types */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Alert Types</h5>
                <div className="grid grid-cols-2 gap-3">
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Threshold Exceeded</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.alert_type === 'threshold_exceeded' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Anomaly Detected</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.alert_type === 'anomaly_detected' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">System Failure</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.alert_type === 'system_failure' && !a.resolved).length}
                    </div>
                  </div>
                  <div className="bg-gray-50 rounded-lg p-3">
                    <div className="text-xs text-gray-600">Security Incident</div>
                    <div className="text-lg font-bold text-gray-900">
                      {alerts.filter(a => a.alert_type === 'security_incident' && !a.resolved).length}
                    </div>
                  </div>
                </div>
              </div>

              {/* Notification Settings */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Notification Settings</h5>
                <div className="grid grid-cols-2 gap-3">
                  {Object.entries(notificationSettings).map(([channel, enabled]) => (
                    <div key={channel} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                      <div className="flex items-center space-x-2">
                        <div className={`w-3 h-3 rounded-full ${enabled ? 'bg-green-500' : 'bg-gray-300'}`}></div>
                        <span className="text-xs font-medium text-gray-700 capitalize">{channel}</span>
                      </div>
                      <button
                        onClick={() => setNotificationSettings(prev => ({
                          ...prev,
                          [channel]: !enabled
                        }))}
                        className="text-xs text-blue-600 hover:text-blue-800"
                      >
                        {enabled ? 'Disable' : 'Enable'}
                      </button>
                    </div>
                  ))}
                </div>
              </div>

              {/* Escalation Levels */}
              <div>
                <h5 className="text-xs font-medium text-gray-700 mb-2">Escalation Levels</h5>
                <div className="grid grid-cols-4 gap-3">
                  {[1, 2, 3, 4].map(level => {
                    const count = alerts.filter(a => a.escalation_level === level && !a.resolved).length;
                    return (
                      <div key={level} className="bg-gray-50 rounded-lg p-3">
                        <div className="text-xs text-gray-600">Level {level}</div>
                        <div className="text-lg font-bold text-gray-900">{count}</div>
                      </div>
                    );
                  })}
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Alert Management */}
        <div className="mt-6 bg-blue-50 rounded-lg p-3">
          <div className="flex items-center space-x-2">
            <BellIcon className="h-5 w-5 text-blue-600" />
            <div>
              <h5 className="text-sm font-medium text-blue-900">Alert Management</h5>
              <p className="text-xs text-blue-700">
                {activeAlerts.length} active alerts | {dismissedAlerts.length} dismissed | {resolvedAlerts.length} resolved
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AlertsPanel;
