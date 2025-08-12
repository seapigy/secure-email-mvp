import React, { useState, useEffect } from 'react';
import { 
  Bell, 
  Clock, 
  Shield, 
  AlertTriangle, 
  Settings,
  CheckCircle,
  XCircle,
  Info
} from 'lucide-react';

interface NotificationDeliveryControlsProps {
  emailId?: string; // Optional - if provided, shows per-email settings
  onSettingsChange?: (settings: NotificationDeliverySettings) => void;
  className?: string;
}

export interface NotificationDeliverySettings {
  deliveryFrequency: 'immediate' | 'daily_digest' | 'first_attempt_only' | 'threshold_trigger';
  thresholdAttempts: number;
  rateLimitWindowMinutes: number;
  rateLimitMaxNotifications: number;
  digestDeliveryTime: string;
  digestEmailEnabled: boolean;
  digestSMSEnabled: boolean;
  inheritGlobalSettings: boolean;
}

const NotificationDeliveryControls: React.FC<NotificationDeliveryControlsProps> = ({
  emailId,
  onSettingsChange,
  className = ''
}) => {
  const [settings, setSettings] = useState<NotificationDeliverySettings>({
    deliveryFrequency: 'immediate',
    thresholdAttempts: 3,
    rateLimitWindowMinutes: 15,
    rateLimitMaxNotifications: 5,
    digestDeliveryTime: '08:00',
    digestEmailEnabled: true,
    digestSMSEnabled: false,
    inheritGlobalSettings: true
  });

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Load current settings on mount
  useEffect(() => {
    loadSettings();
  }, [emailId]);

  const loadSettings = async () => {
    setIsLoading(true);
    setError(null);

    try {
      const endpoint = emailId 
        ? `/api/notifications/email/${emailId}/preferences`
        : '/api/notifications/preferences';

      const response = await fetch(endpoint, {
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        }
      });

      if (!response.ok) {
        throw new Error('Failed to load notification settings');
      }

      const data = await response.json();
      
      if (emailId) {
        // Per-email settings
        setSettings({
          deliveryFrequency: data.delivery_frequency || 'immediate',
          thresholdAttempts: data.threshold_attempts || 3,
          rateLimitWindowMinutes: data.rate_limit_window_minutes || 15,
          rateLimitMaxNotifications: data.rate_limit_max_notifications || 5,
          digestDeliveryTime: data.digest_delivery_time || '08:00',
          digestEmailEnabled: data.digest_email_enabled !== false,
          digestSMSEnabled: data.digest_sms_enabled === true,
          inheritGlobalSettings: data.inherit_global_settings !== false
        });
      } else {
        // Global settings
        setSettings({
          deliveryFrequency: data.delivery_frequency || 'immediate',
          thresholdAttempts: data.threshold_attempts || 3,
          rateLimitWindowMinutes: data.rate_limit_window_minutes || 15,
          rateLimitMaxNotifications: data.rate_limit_max_notifications || 5,
          digestDeliveryTime: data.digest_delivery_time || '08:00',
          digestEmailEnabled: data.digest_email_enabled !== false,
          digestSMSEnabled: data.digest_sms_enabled === true,
          inheritGlobalSettings: false // Not applicable for global settings
        });
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load settings');
    } finally {
      setIsLoading(false);
    }
  };

  const saveSettings = async () => {
    setIsLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const endpoint = emailId 
        ? `/api/notifications/email/${emailId}/preferences`
        : '/api/notifications/preferences';

      const response = await fetch(endpoint, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          delivery_frequency: settings.deliveryFrequency,
          threshold_attempts: settings.thresholdAttempts,
          rate_limit_window_minutes: settings.rateLimitWindowMinutes,
          rate_limit_max_notifications: settings.rateLimitMaxNotifications,
          digest_delivery_time: settings.digestDeliveryTime,
          digest_email_enabled: settings.digestEmailEnabled,
          digest_sms_enabled: settings.digestSMSEnabled,
          inherit_global_settings: settings.inheritGlobalSettings
        })
      });

      if (!response.ok) {
        throw new Error('Failed to save notification settings');
      }

      setSuccess('Notification settings saved successfully');
      onSettingsChange?.(settings);

      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save settings');
    } finally {
      setIsLoading(false);
    }
  };

  const handleSettingChange = (key: keyof NotificationDeliverySettings, value: any) => {
    setSettings(prev => ({
      ...prev,
      [key]: value
    }));
  };

  const getDeliveryFrequencyDescription = (frequency: string) => {
    switch (frequency) {
      case 'immediate':
        return 'Send notification for every access attempt';
      case 'daily_digest':
        return 'Send a daily summary of all access events';
      case 'first_attempt_only':
        return 'Send notification only for the first access attempt from each IP';
      case 'threshold_trigger':
        return 'Send notification only after multiple failed attempts';
      default:
        return '';
    }
  };

  if (isLoading && !settings.deliveryFrequency) {
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

  return (
    <div className={`bg-white dark:bg-secondary-800 rounded-lg border border-secondary-200 dark:border-secondary-700 ${className}`}>
      <div className="p-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-primary-100 dark:bg-primary-900/20 rounded-lg">
              <Bell className="h-5 w-5 text-primary-600 dark:text-primary-400" />
            </div>
            <div>
              <h3 className="text-lg font-semibold text-secondary-900 dark:text-white">
                {emailId ? 'Email-Specific' : 'Global'} Notification Controls
              </h3>
              <p className="text-sm text-secondary-600 dark:text-secondary-400">
                Control how and when you receive access notifications
              </p>
            </div>
          </div>
          {emailId && (
            <div className="flex items-center space-x-2">
              <input
                type="checkbox"
                id="inheritGlobal"
                checked={settings.inheritGlobalSettings}
                onChange={(e) => handleSettingChange('inheritGlobalSettings', e.target.checked)}
                className="rounded border-secondary-300 text-primary-600 focus:ring-primary-500"
              />
              <label htmlFor="inheritGlobal" className="text-sm text-secondary-700 dark:text-secondary-300">
                Inherit global settings
              </label>
            </div>
          )}
        </div>

        {/* Error/Success Messages */}
        {error && (
          <div className="mb-4 p-3 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg flex items-center space-x-2">
            <XCircle className="h-4 w-4 text-red-500" />
            <span className="text-sm text-red-700 dark:text-red-400">{error}</span>
          </div>
        )}

        {success && (
          <div className="mb-4 p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg flex items-center space-x-2">
            <CheckCircle className="h-4 w-4 text-green-500" />
            <span className="text-sm text-green-700 dark:text-green-400">{success}</span>
          </div>
        )}

        {/* Settings Form */}
        <div className={`space-y-6 ${emailId && settings.inheritGlobalSettings ? 'opacity-50 pointer-events-none' : ''}`}>
          {/* Delivery Frequency */}
          <div>
            <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
              Delivery Frequency
            </label>
            <div className="space-y-3">
              {[
                { value: 'immediate', label: 'Immediate', icon: Bell },
                { value: 'daily_digest', label: 'Daily Digest', icon: Clock },
                { value: 'first_attempt_only', label: 'First Attempt Only', icon: Shield },
                { value: 'threshold_trigger', label: 'Threshold Trigger', icon: AlertTriangle }
              ].map((option) => {
                const IconComponent = option.icon;
                return (
                  <label key={option.value} className="flex items-center space-x-3 cursor-pointer">
                    <input
                      type="radio"
                      name="deliveryFrequency"
                      value={option.value}
                      checked={settings.deliveryFrequency === option.value}
                      onChange={(e) => handleSettingChange('deliveryFrequency', e.target.value)}
                      className="text-primary-600 focus:ring-primary-500"
                    />
                    <IconComponent className="h-4 w-4 text-secondary-500" />
                    <div>
                      <div className="text-sm font-medium text-secondary-900 dark:text-white">
                        {option.label}
                      </div>
                      <div className="text-xs text-secondary-600 dark:text-secondary-400">
                        {getDeliveryFrequencyDescription(option.value)}
                      </div>
                    </div>
                  </label>
                );
              })}
            </div>
          </div>

          {/* Threshold Settings (only for threshold_trigger) */}
          {settings.deliveryFrequency === 'threshold_trigger' && (
            <div>
              <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                Failed Attempts Threshold
              </label>
              <div className="flex items-center space-x-3">
                <input
                  type="number"
                  min="1"
                  max="10"
                  value={settings.thresholdAttempts}
                  onChange={(e) => handleSettingChange('thresholdAttempts', parseInt(e.target.value))}
                  className="w-20 px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-secondary-700 text-secondary-900 dark:text-white"
                />
                <span className="text-sm text-secondary-600 dark:text-secondary-400">
                  failed attempts before notification
                </span>
              </div>
            </div>
          )}

          {/* Rate Limiting */}
          <div className="border-t border-secondary-200 dark:border-secondary-700 pt-6">
            <div className="flex items-center space-x-2 mb-4">
              <Settings className="h-4 w-4 text-secondary-500" />
              <h4 className="text-sm font-medium text-secondary-900 dark:text-white">Rate Limiting</h4>
            </div>
            
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                  Time Window (minutes)
                </label>
                <input
                  type="number"
                  min="1"
                  max="1440"
                  value={settings.rateLimitWindowMinutes}
                  onChange={(e) => handleSettingChange('rateLimitWindowMinutes', parseInt(e.target.value))}
                  className="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-secondary-700 text-secondary-900 dark:text-white"
                />
              </div>
              
              <div>
                <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                  Max Notifications
                </label>
                <input
                  type="number"
                  min="1"
                  max="100"
                  value={settings.rateLimitMaxNotifications}
                  onChange={(e) => handleSettingChange('rateLimitMaxNotifications', parseInt(e.target.value))}
                  className="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-secondary-700 text-secondary-900 dark:text-white"
                />
              </div>
            </div>
            
            <div className="mt-3 p-3 bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg flex items-start space-x-2">
              <Info className="h-4 w-4 text-blue-500 mt-0.5 flex-shrink-0" />
              <div className="text-xs text-blue-700 dark:text-blue-400">
                Rate limiting prevents notification spam by limiting notifications to{' '}
                <strong>{settings.rateLimitMaxNotifications}</strong> per{' '}
                <strong>{settings.rateLimitWindowMinutes} minutes</strong> per IP address.
              </div>
            </div>
          </div>

          {/* Daily Digest Settings (only for daily_digest) */}
          {settings.deliveryFrequency === 'daily_digest' && (
            <div className="border-t border-secondary-200 dark:border-secondary-700 pt-6">
              <div className="flex items-center space-x-2 mb-4">
                <Clock className="h-4 w-4 text-secondary-500" />
                <h4 className="text-sm font-medium text-secondary-900 dark:text-white">Daily Digest Settings</h4>
              </div>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                <div>
                  <label className="block text-sm font-medium text-secondary-700 dark:text-secondary-300 mb-2">
                    Delivery Time (UTC)
                  </label>
                  <input
                    type="time"
                    value={settings.digestDeliveryTime}
                    onChange={(e) => handleSettingChange('digestDeliveryTime', e.target.value)}
                    className="w-full px-3 py-2 border border-secondary-300 dark:border-secondary-600 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 bg-white dark:bg-secondary-700 text-secondary-900 dark:text-white"
                  />
                </div>
                
                <div className="flex flex-col space-y-3">
                  <label className="flex items-center space-x-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.digestEmailEnabled}
                      onChange={(e) => handleSettingChange('digestEmailEnabled', e.target.checked)}
                      className="text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-secondary-700 dark:text-secondary-300">
                      Send digest via email
                    </span>
                  </label>
                  
                  <label className="flex items-center space-x-3 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={settings.digestSMSEnabled}
                      onChange={(e) => handleSettingChange('digestSMSEnabled', e.target.checked)}
                      className="text-primary-600 focus:ring-primary-500"
                    />
                    <span className="text-sm text-secondary-700 dark:text-secondary-300">
                      Send digest via SMS
                    </span>
                  </label>
                </div>
              </div>
              
              <div className="p-3 bg-green-50 dark:bg-green-900/20 border border-green-200 dark:border-green-800 rounded-lg flex items-start space-x-2">
                <CheckCircle className="h-4 w-4 text-green-500 mt-0.5 flex-shrink-0" />
                <div className="text-xs text-green-700 dark:text-green-400">
                  Daily digests will be sent at <strong>{settings.digestDeliveryTime} UTC</strong> and include a summary of all access events from the previous day.
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Action Buttons */}
        <div className="flex justify-end space-x-3 mt-6 pt-6 border-t border-secondary-200 dark:border-secondary-700">
          <button
            onClick={loadSettings}
            disabled={isLoading}
            className="px-4 py-2 text-sm font-medium text-secondary-700 dark:text-secondary-300 bg-white dark:bg-secondary-700 border border-secondary-300 dark:border-secondary-600 rounded-lg hover:bg-secondary-50 dark:hover:bg-secondary-600 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50"
          >
            Reset
          </button>
          <button
            onClick={saveSettings}
            disabled={isLoading || (emailId && settings.inheritGlobalSettings)}
            className="px-4 py-2 text-sm font-medium text-white bg-primary-600 hover:bg-primary-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            {isLoading ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default NotificationDeliveryControls;
