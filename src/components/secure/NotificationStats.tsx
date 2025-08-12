import React, { useState, useEffect } from 'react';
import { 
  BarChart3, 
  Bell, 
  XCircle, 
  Clock, 
  Shield, 
  AlertTriangle,
  TrendingUp,
  TrendingDown,
  Info
} from 'lucide-react';

interface NotificationStatsProps {
  className?: string;
}

interface NotificationStats {
  total_events: number;
  suppressed_events: number;
  suppression_stats: {
    [key: string]: number;
  };
  delivery_frequency: string;
  rate_limit_info: {
    window_minutes: number;
    max_notifications: number;
  };
}

interface NotificationSuppression {
  suppression_id: string;
  email_id: string;
  user_id: string;
  event_id: string;
  suppression_reason: string;
  suppressed_at: string;
  original_event_type: string;
  ip_address?: string;
  user_agent?: string;
  country?: string;
  city?: string;
  device_type?: string;
  failure_reason?: string;
}

const NotificationStats: React.FC<NotificationStatsProps> = ({ className = '' }) => {
  const [stats, setStats] = useState<NotificationStats | null>(null);
  const [suppressions, setSuppressions] = useState<NotificationSuppression[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'suppressions'>('overview');

  useEffect(() => {
    loadStats();
    loadSuppressions();
  }, []);

  const loadStats = async () => {
    try {
      const response = await fetch('/api/notifications/stats', {
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to load notification statistics');
      }

      const data = await response.json();
      setStats(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load statistics');
    }
  };

  const loadSuppressions = async () => {
    try {
      const response = await fetch('/api/notifications/suppressions?limit=50', {
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to load notification suppressions');
      }

      const data = await response.json();
      setSuppressions(data);
    } catch (err) {
      console.error('Failed to load suppressions:', err);
    } finally {
      setIsLoading(false);
    }
  };

  const getSuppressionReasonIcon = (reason: string) => {
    switch (reason) {
      case 'rate_limited':
        return <Clock className="h-4 w-4 text-orange-500" />;
      case 'frequency_controlled':
        return <Bell className="h-4 w-4 text-blue-500" />;
      case 'threshold_not_met':
        return <AlertTriangle className="h-4 w-4 text-yellow-500" />;
      case 'first_attempt_only':
        return <Shield className="h-4 w-4 text-green-500" />;
      default:
        return <XCircle className="h-4 w-4 text-red-500" />;
    }
  };

  const getSuppressionReasonLabel = (reason: string) => {
    switch (reason) {
      case 'rate_limited':
        return 'Rate Limited';
      case 'frequency_controlled':
        return 'Frequency Controlled';
      case 'threshold_not_met':
        return 'Threshold Not Met';
      case 'first_attempt_only':
        return 'First Attempt Only';
      default:
        return reason.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase());
    }
  };

  const getSuppressionReasonDescription = (reason: string) => {
    switch (reason) {
      case 'rate_limited':
        return 'Notification was suppressed due to rate limiting';
      case 'frequency_controlled':
        return 'Notification was suppressed due to delivery frequency settings';
      case 'threshold_not_met':
        return 'Notification was suppressed because threshold was not reached';
      case 'first_attempt_only':
        return 'Notification was suppressed because it was not the first attempt';
      default:
        return 'Notification was suppressed';
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString();
  };

  if (isLoading) {
    return (
      <div className={`animate-pulse ${className}`}>
        <div className="bg-secondary-100 dark:bg-secondary-800 rounded-lg p-6">
          <div className="h-4 bg-secondary-200 dark:bg-secondary-700 rounded mb-4"></div>
          <div className="h-4 bg-secondary-200 dark:bg-secondary-700 rounded mb-4 w-3/4"></div>
          <div className="h-4 bg-secondary-200 dark:bg-secondary-700 rounded w-1/2"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4 ${className}`}>
        <div className="flex items-center space-x-2">
          <XCircle className="h-5 w-5 text-red-500" />
          <span className="text-red-700 dark:text-red-400">{error}</span>
        </div>
      </div>
    );
  }

  if (!stats) {
    return null;
  }

  const suppressionRate = stats.total_events > 0 
    ? ((stats.suppressed_events / stats.total_events) * 100).toFixed(1)
    : '0';

  return (
    <div className={`bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 ${className}`}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-primary-100 dark:bg-primary-900/20 rounded-lg">
              <BarChart3 className="h-5 w-5 text-primary-600 dark:text-primary-400" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-secondary-900 dark:text-white">
                Notification Statistics
              </h3>
              <p className="text-sm text-secondary-600 dark:text-secondary-400">
                Monitor notification delivery and suppression rates
              </p>
            </div>
          </div>
        </div>

        {/* Tabs */}
        <div className="border-b border-secondary-200 dark:border-secondary-700 mb-6">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab('overview')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'overview'
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
              }`}
            >
              Overview
            </button>
            <button
              onClick={() => setActiveTab('suppressions')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'suppressions'
                  ? 'border-primary-500 text-primary-600 dark:text-primary-400'
                  : 'border-transparent text-secondary-500 hover:text-secondary-700 hover:border-secondary-300'
              }`}
            >
              Suppressions ({suppressions.length})
            </button>
          </nav>
        </div>

        {activeTab === 'overview' && (
          <div className="space-y-6">
            {/* Key Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="bg-secondary-50 dark:bg-secondary-700/50 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-secondary-600 dark:text-secondary-400">
                      Total Events
                    </p>
                    <p className="text-2xl font-bold text-secondary-900 dark:text-white">
                      {stats.total_events}
                    </p>
                  </div>
                  <TrendingUp className="h-8 w-8 text-green-500" />
                </div>
              </div>

              <div className="bg-secondary-50 dark:bg-secondary-700/50 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-secondary-600 dark:text-secondary-400">
                      Suppressed Events
                    </p>
                    <p className="text-2xl font-bold text-secondary-900 dark:text-white">
                      {stats.suppressed_events}
                    </p>
                  </div>
                  <TrendingDown className="h-8 w-8 text-orange-500" />
                </div>
              </div>

              <div className="bg-secondary-50 dark:bg-secondary-700/50 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-secondary-600 dark:text-secondary-400">
                      Suppression Rate
                    </p>
                    <p className="text-2xl font-bold text-secondary-900 dark:text-white">
                      {suppressionRate}%
                    </p>
                  </div>
                  <BarChart3 className="h-8 w-8 text-blue-500" />
                </div>
              </div>
            </div>

            {/* Current Settings */}
            <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
              <div className="flex items-start space-x-3">
                <Info className="h-5 w-5 text-blue-500 mt-0.5 flex-shrink-0" />
                <div>
                  <h4 className="text-sm font-medium text-blue-900 dark:text-blue-100 mb-2">
                    Current Delivery Settings
                  </h4>
                  <div className="text-sm text-blue-800 dark:text-blue-200 space-y-1">
                    <p><strong>Frequency:</strong> {stats.delivery_frequency.replace('_', ' ').replace(/\b\w/g, l => l.toUpperCase())}</p>
                    <p><strong>Rate Limit:</strong> {stats.rate_limit_info.max_notifications} notifications per {stats.rate_limit_info.window_minutes} minutes</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Suppression Breakdown */}
            {Object.keys(stats.suppression_stats).length > 0 && (
              <div>
                <h4 className="text-sm font-medium text-secondary-900 dark:text-white mb-3">
                  Suppression Breakdown
                </h4>
                <div className="space-y-2">
                  {Object.entries(stats.suppression_stats).map(([reason, count]) => (
                    <div key={reason} className="flex items-center justify-between p-3 bg-secondary-50 dark:bg-secondary-700/50 rounded-lg">
                      <div className="flex items-center space-x-3">
                        {getSuppressionReasonIcon(reason)}
                        <div>
                          <p className="text-sm font-medium text-secondary-900 dark:text-white">
                            {getSuppressionReasonLabel(reason)}
                          </p>
                          <p className="text-xs text-secondary-600 dark:text-secondary-400">
                            {getSuppressionReasonDescription(reason)}
                          </p>
                        </div>
                      </div>
                      <span className="text-sm font-semibold text-secondary-900 dark:text-white">
                        {count}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}

        {activeTab === 'suppressions' && (
          <div className="space-y-4">
            {suppressions.length === 0 ? (
              <div className="text-center py-8">
                <XCircle className="h-12 w-12 text-secondary-400 mx-auto mb-4" />
                <p className="text-secondary-600 dark:text-secondary-400">
                  No suppressed notifications found
                </p>
              </div>
            ) : (
              <div className="space-y-3">
                {suppressions.map((suppression) => (
                  <div key={suppression.suppression_id} className="border border-secondary-200 dark:border-secondary-700 rounded-lg p-4">
                    <div className="flex items-start justify-between">
                      <div className="flex items-start space-x-3">
                        {getSuppressionReasonIcon(suppression.suppression_reason)}
                        <div className="flex-1">
                          <div className="flex items-center space-x-2 mb-1">
                            <span className="text-sm font-medium text-secondary-900 dark:text-white">
                              {getSuppressionReasonLabel(suppression.suppression_reason)}
                            </span>
                            <span className="text-xs text-secondary-500 dark:text-secondary-400">
                              {suppression.original_event_type}
                            </span>
                          </div>
                          <p className="text-xs text-secondary-600 dark:text-secondary-400 mb-2">
                            {getSuppressionReasonDescription(suppression.suppression_reason)}
                          </p>
                          <div className="grid grid-cols-2 gap-4 text-xs">
                            {suppression.ip_address && (
                              <div>
                                <span className="text-secondary-500">IP:</span> {suppression.ip_address}
                              </div>
                            )}
                            {suppression.country && (
                              <div>
                                <span className="text-secondary-500">Location:</span> {suppression.country}
                                {suppression.city && `, ${suppression.city}`}
                              </div>
                            )}
                            {suppression.device_type && (
                              <div>
                                <span className="text-secondary-500">Device:</span> {suppression.device_type}
                              </div>
                            )}
                            {suppression.failure_reason && (
                              <div>
                                <span className="text-secondary-500">Reason:</span> {suppression.failure_reason}
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <p className="text-xs text-secondary-500 dark:text-secondary-400">
                          {formatDate(suppression.suppressed_at)}
                        </p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default NotificationStats;
