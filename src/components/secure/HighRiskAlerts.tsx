import React, { useState, useEffect } from 'react';
import { 
  AlertTriangle, 
  CheckCircle, 
  Clock, 
  MapPin, 
  Shield,
  Eye,
  EyeOff,
  RefreshCw
} from 'lucide-react';

interface HighRiskAlert {
  alert_id: string;
  email_id: string;
  alert_type: string;
  severity: string;
  triggered_channels: string;
  alert_data: string;
  acknowledged: boolean;
  created_at: string;
}

interface HighRiskAlertsProps {
  className?: string;
}

const HighRiskAlerts: React.FC<HighRiskAlertsProps> = ({ className = '' }) => {
  const [alerts, setAlerts] = useState<HighRiskAlert[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAcknowledged, setShowAcknowledged] = useState(false);

  // Load alerts on mount
  useEffect(() => {
    loadAlerts();
  }, []);

  const loadAlerts = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/multichannel/alerts?limit=50', {
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to load high-risk alerts');
      }

      const data = await response.json();
      setAlerts(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load alerts');
    } finally {
      setIsLoading(false);
    }
  };

  const acknowledgeAlert = async (alertId: string) => {
    try {
      const response = await fetch(`/api/multichannel/alerts/${alertId}/acknowledge`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to acknowledge alert');
      }

      // Update the alert in the list
      setAlerts(prev => prev.map(alert => 
        alert.alert_id === alertId 
          ? { ...alert, acknowledged: true }
          : alert
      ));
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to acknowledge alert');
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity.toLowerCase()) {
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

  const getAlertTypeIcon = (alertType: string) => {
    switch (alertType.toLowerCase()) {
      case 'multiple_failures':
        return <Shield className="w-5 h-5 text-red-500" />;
      case 'suspicious_location':
        return <MapPin className="w-5 h-5 text-orange-500" />;
      case 'unusual_time':
        return <Clock className="w-5 h-5 text-yellow-500" />;
      case 'known_attacker_ip':
        return <AlertTriangle className="w-5 h-5 text-red-500" />;
      default:
        return <AlertTriangle className="w-5 h-5 text-gray-500" />;
    }
  };

  const formatAlertType = (alertType: string) => {
    return alertType.split('_').map(word => 
      word.charAt(0).toUpperCase() + word.slice(1)
    ).join(' ');
  };

  const parseAlertData = (alertData: string) => {
    try {
      return JSON.parse(alertData);
    } catch {
      return null;
    }
  };

  const filteredAlerts = showAcknowledged 
    ? alerts 
    : alerts.filter(alert => !alert.acknowledged);

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 ${className}`}>
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center">
          <AlertTriangle className="w-6 h-6 text-red-600 mr-3" />
          <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
            High-Risk Security Alerts
          </h2>
        </div>
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setShowAcknowledged(!showAcknowledged)}
            className="flex items-center px-3 py-1 text-sm text-gray-600 hover:text-gray-800"
          >
            {showAcknowledged ? <EyeOff className="w-4 h-4 mr-1" /> : <Eye className="w-4 h-4 mr-1" />}
            {showAcknowledged ? 'Hide Acknowledged' : 'Show All'}
          </button>
          <button
            onClick={loadAlerts}
            disabled={isLoading}
            className="flex items-center px-3 py-1 text-sm text-blue-600 hover:text-blue-800 disabled:opacity-50"
          >
            <RefreshCw className={`w-4 h-4 mr-1 ${isLoading ? 'animate-spin' : ''}`} />
            Refresh
          </button>
        </div>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
          {error}
        </div>
      )}

      {isLoading ? (
        <div className="flex justify-center py-8">
          <RefreshCw className="w-6 h-6 text-blue-600 animate-spin" />
        </div>
      ) : filteredAlerts.length === 0 ? (
        <div className="text-center py-8 text-gray-500">
          <Shield className="w-12 h-12 mx-auto mb-3 text-gray-300" />
          <p>No high-risk alerts found</p>
        </div>
      ) : (
        <div className="space-y-4">
          {filteredAlerts.map((alert) => {
            const alertData = parseAlertData(alert.alert_data);
            const createdDate = new Date(alert.created_at);
            
            return (
              <div
                key={alert.alert_id}
                className={`border rounded-lg p-4 ${
                  alert.acknowledged 
                    ? 'border-gray-200 bg-gray-50' 
                    : 'border-red-200 bg-red-50'
                }`}
              >
                <div className="flex items-start justify-between">
                  <div className="flex items-start space-x-3">
                    {getAlertTypeIcon(alert.alert_type)}
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-2">
                        <h3 className="font-medium text-gray-900 dark:text-white">
                          {formatAlertType(alert.alert_type)}
                        </h3>
                        <span className={`px-2 py-1 text-xs font-medium rounded-full border ${getSeverityColor(alert.severity)}`}>
                          {alert.severity.toUpperCase()}
                        </span>
                        {alert.acknowledged && (
                          <span className="px-2 py-1 text-xs font-medium rounded-full bg-green-100 text-green-600 border border-green-200">
                            ACKNOWLEDGED
                          </span>
                        )}
                      </div>
                      
                      <div className="text-sm text-gray-600 dark:text-gray-300 mb-2">
                        <p>Email ID: {alert.email_id}</p>
                        <p>Channels: {alert.triggered_channels}</p>
                        <p>Time: {createdDate.toLocaleString()}</p>
                      </div>

                      {alertData && (
                        <div className="bg-white dark:bg-gray-700 rounded p-3 text-sm">
                          <h4 className="font-medium mb-2">Alert Details:</h4>
                          {alertData.ip_address && (
                            <p><strong>IP Address:</strong> {alertData.ip_address}</p>
                          )}
                          {alertData.location && (
                            <p><strong>Location:</strong> {alertData.location}</p>
                          )}
                          {alertData.device_info && (
                            <p><strong>Device:</strong> {alertData.device_info}</p>
                          )}
                          {alertData.failure_count && (
                            <p><strong>Failure Count:</strong> {alertData.failure_count}</p>
                          )}
                          {alertData.time_of_day && (
                            <p><strong>Time of Day:</strong> {alertData.time_of_day}</p>
                          )}
                          {alertData.threat_level && (
                            <p><strong>Threat Level:</strong> {alertData.threat_level}</p>
                          )}
                          {alertData.recommendations && alertData.recommendations.length > 0 && (
                            <div>
                              <strong>Recommendations:</strong>
                              <ul className="list-disc list-inside mt-1">
                                {alertData.recommendations.map((rec: string, index: number) => (
                                  <li key={index}>{rec}</li>
                                ))}
                              </ul>
                            </div>
                          )}
                        </div>
                      )}
                    </div>
                  </div>

                  {!alert.acknowledged && (
                    <button
                      onClick={() => acknowledgeAlert(alert.alert_id)}
                      className="flex items-center px-3 py-1 text-sm text-green-600 hover:text-green-800 bg-green-100 hover:bg-green-200 rounded"
                    >
                      <CheckCircle className="w-4 h-4 mr-1" />
                      Acknowledge
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      <div className="mt-6 text-sm text-gray-500 text-center">
        Showing {filteredAlerts.length} of {alerts.length} alerts
        {!showAcknowledged && alerts.filter(a => a.acknowledged).length > 0 && (
          <span> ({alerts.filter(a => a.acknowledged).length} acknowledged)</span>
        )}
      </div>
    </div>
  );
};

export default HighRiskAlerts;
