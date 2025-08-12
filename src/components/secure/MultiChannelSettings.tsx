import React, { useState, useEffect } from 'react';
import { 
  Bell, 
  Shield, 
  Smartphone, 
  MessageCircle, 
  Settings,
  CheckCircle,
  XCircle,
  Info,
  AlertTriangle,
  Zap
} from 'lucide-react';

interface MultiChannelSettingsProps {
  emailId?: string; // Optional - if provided, shows per-email settings
  onSettingsChange?: (settings: MultiChannelSettings) => void;
  className?: string;
}

export interface MultiChannelSettings {
  pushNotificationsEnabled: boolean;
  signalEnabled: boolean;
  matrixEnabled: boolean;
  telegramEnabled: boolean;
  discordEnabled: boolean;
  pushDeviceToken: string;
  signalPhone: string;
  matrixUserId: string;
  matrixHomeserver: string;
  telegramChatId: string;
  discordWebhookUrl: string;
  highRiskChannels: string;
  highRiskThreshold: number;
  highRiskTimeoutMinutes: number;
}

const MultiChannelSettings: React.FC<MultiChannelSettingsProps> = ({
  emailId,
  onSettingsChange,
  className = ''
}) => {
  const [settings, setSettings] = useState<MultiChannelSettings>({
    pushNotificationsEnabled: false,
    signalEnabled: false,
    matrixEnabled: false,
    telegramEnabled: false,
    discordEnabled: false,
    pushDeviceToken: '',
    signalPhone: '',
    matrixUserId: '',
    matrixHomeserver: '',
    telegramChatId: '',
    discordWebhookUrl: '',
    highRiskChannels: 'email,sms',
    highRiskThreshold: 3,
    highRiskTimeoutMinutes: 30
  });

  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [verificationStatus, setVerificationStatus] = useState<{[key: string]: string}>({});

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
        throw new Error('Failed to load multi-channel settings');
      }

      const data = await response.json();
      
      setSettings({
        pushNotificationsEnabled: data.push_notifications_enabled === true,
        signalEnabled: data.signal_enabled === true,
        matrixEnabled: data.matrix_enabled === true,
        telegramEnabled: data.telegram_enabled === true,
        discordEnabled: data.discord_enabled === true,
        pushDeviceToken: data.push_device_token || '',
        signalPhone: data.signal_phone || '',
        matrixUserId: data.matrix_user_id || '',
        matrixHomeserver: data.matrix_homeserver || '',
        telegramChatId: data.telegram_chat_id || '',
        discordWebhookUrl: data.discord_webhook_url || '',
        highRiskChannels: data.high_risk_channels || 'email,sms',
        highRiskThreshold: data.high_risk_threshold || 3,
        highRiskTimeoutMinutes: data.high_risk_timeout_minutes || 30
      });
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
          push_notifications_enabled: settings.pushNotificationsEnabled,
          signal_enabled: settings.signalEnabled,
          matrix_enabled: settings.matrixEnabled,
          telegram_enabled: settings.telegramEnabled,
          discord_enabled: settings.discordEnabled,
          push_device_token: settings.pushDeviceToken,
          signal_phone: settings.signalPhone,
          matrix_user_id: settings.matrixUserId,
          matrix_homeserver: settings.matrixHomeserver,
          telegram_chat_id: settings.telegramChatId,
          discord_webhook_url: settings.discordWebhookUrl,
          high_risk_channels: settings.highRiskChannels,
          high_risk_threshold: settings.highRiskThreshold,
          high_risk_timeout_minutes: settings.highRiskTimeoutMinutes
        })
      });

      if (!response.ok) {
        throw new Error('Failed to save multi-channel settings');
      }

      setSuccess('Multi-channel settings saved successfully');
      onSettingsChange?.(settings);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save settings');
    } finally {
      setIsLoading(false);
    }
  };

  const verifyChannel = async (channelType: string, identifier: string) => {
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch('/api/multichannel/verify', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          channel_type: channelType,
          channel_identifier: identifier
        })
      });

      if (!response.ok) {
        throw new Error('Failed to initiate channel verification');
      }

      const data = await response.json();
      setVerificationStatus(prev => ({
        ...prev,
        [channelType]: 'pending'
      }));

      // Show verification code input
      const code = prompt(`Enter the verification code sent to your ${channelType} channel:`);
      if (code) {
        await confirmVerification(data.verification_id, code);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to verify channel');
    } finally {
      setIsLoading(false);
    }
  };

  const confirmVerification = async (verificationId: string, code: string) => {
    try {
      const response = await fetch('/api/multichannel/verify/confirm', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${sessionStorage.getItem('accessToken')}`,
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          verification_id: verificationId,
          verification_code: code
        })
      });

      if (!response.ok) {
        throw new Error('Invalid verification code');
      }

      setVerificationStatus(prev => ({
        ...prev,
        [verificationId]: 'verified'
      }));
      setSuccess('Channel verified successfully');
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Verification failed');
    }
  };

  const handleInputChange = (field: keyof MultiChannelSettings, value: string | boolean | number) => {
    setSettings(prev => ({
      ...prev,
      [field]: value
    }));
  };

  return (
    <div className={`bg-white dark:bg-gray-800 rounded-lg shadow-md p-6 ${className}`}>
      <div className="flex items-center mb-6">
        <Zap className="w-6 h-6 text-blue-600 mr-3" />
        <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
          Multi-Channel Alert Settings
        </h2>
      </div>

      {error && (
        <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded flex items-center">
          <XCircle className="w-4 h-4 mr-2" />
          {error}
        </div>
      )}

      {success && (
        <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded flex items-center">
          <CheckCircle className="w-4 h-4 mr-2" />
          {success}
        </div>
      )}

      <div className="space-y-6">
        {/* Push Notifications */}
        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center">
              <Bell className="w-5 h-5 text-blue-600 mr-2" />
              <h3 className="font-medium text-gray-900 dark:text-white">Push Notifications</h3>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.pushNotificationsEnabled}
                onChange={(e) => handleInputChange('pushNotificationsEnabled', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
          {settings.pushNotificationsEnabled && (
            <div className="space-y-2">
              <input
                type="text"
                placeholder="Device Token"
                value={settings.pushDeviceToken}
                onChange={(e) => handleInputChange('pushDeviceToken', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={() => verifyChannel('push', settings.pushDeviceToken)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Verify Device
              </button>
            </div>
          )}
        </div>

        {/* Signal */}
        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center">
              <MessageCircle className="w-5 h-5 text-green-600 mr-2" />
              <h3 className="font-medium text-gray-900 dark:text-white">Signal</h3>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.signalEnabled}
                onChange={(e) => handleInputChange('signalEnabled', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
          {settings.signalEnabled && (
            <div className="space-y-2">
              <input
                type="text"
                placeholder="Phone Number"
                value={settings.signalPhone}
                onChange={(e) => handleInputChange('signalPhone', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={() => verifyChannel('signal', settings.signalPhone)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Verify Phone
              </button>
            </div>
          )}
        </div>

        {/* Matrix */}
        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center">
              <MessageCircle className="w-5 h-5 text-purple-600 mr-2" />
              <h3 className="font-medium text-gray-900 dark:text-white">Matrix</h3>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.matrixEnabled}
                onChange={(e) => handleInputChange('matrixEnabled', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
          {settings.matrixEnabled && (
            <div className="space-y-2">
              <input
                type="text"
                placeholder="User ID (@user:server.com)"
                value={settings.matrixUserId}
                onChange={(e) => handleInputChange('matrixUserId', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <input
                type="text"
                placeholder="Homeserver URL"
                value={settings.matrixHomeserver}
                onChange={(e) => handleInputChange('matrixHomeserver', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={() => verifyChannel('matrix', settings.matrixUserId)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Verify Account
              </button>
            </div>
          )}
        </div>

        {/* Telegram */}
        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center">
              <MessageCircle className="w-5 h-5 text-blue-500 mr-2" />
              <h3 className="font-medium text-gray-900 dark:text-white">Telegram</h3>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.telegramEnabled}
                onChange={(e) => handleInputChange('telegramEnabled', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
          {settings.telegramEnabled && (
            <div className="space-y-2">
              <input
                type="text"
                placeholder="Chat ID"
                value={settings.telegramChatId}
                onChange={(e) => handleInputChange('telegramChatId', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={() => verifyChannel('telegram', settings.telegramChatId)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Verify Chat
              </button>
            </div>
          )}
        </div>

        {/* Discord */}
        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex items-center justify-between mb-3">
            <div className="flex items-center">
              <MessageCircle className="w-5 h-5 text-indigo-600 mr-2" />
              <h3 className="font-medium text-gray-900 dark:text-white">Discord</h3>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                checked={settings.discordEnabled}
                onChange={(e) => handleInputChange('discordEnabled', e.target.checked)}
                className="sr-only peer"
              />
              <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
            </label>
          </div>
          {settings.discordEnabled && (
            <div className="space-y-2">
              <input
                type="text"
                placeholder="Webhook URL"
                value={settings.discordWebhookUrl}
                onChange={(e) => handleInputChange('discordWebhookUrl', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <button
                onClick={() => verifyChannel('discord', settings.discordWebhookUrl)}
                className="text-sm text-blue-600 hover:text-blue-800"
              >
                Test Webhook
              </button>
            </div>
          )}
        </div>

        {/* High-Risk Alert Settings */}
        <div className="border border-gray-200 rounded-lg p-4">
          <div className="flex items-center mb-3">
            <AlertTriangle className="w-5 h-5 text-red-600 mr-2" />
            <h3 className="font-medium text-gray-900 dark:text-white">High-Risk Alert Settings</h3>
          </div>
          <div className="space-y-3">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                High-Risk Channels
              </label>
              <input
                type="text"
                placeholder="email,sms,push"
                value={settings.highRiskChannels}
                onChange={(e) => handleInputChange('highRiskChannels', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                High-Risk Threshold (failed attempts)
              </label>
              <input
                type="number"
                min="1"
                max="10"
                value={settings.highRiskThreshold}
                onChange={(e) => handleInputChange('highRiskThreshold', parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Timeout (minutes)
              </label>
              <input
                type="number"
                min="5"
                max="60"
                value={settings.highRiskTimeoutMinutes}
                onChange={(e) => handleInputChange('highRiskTimeoutMinutes', parseInt(e.target.value))}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>
        </div>
      </div>

      <div className="mt-6 flex justify-end">
        <button
          onClick={saveSettings}
          disabled={isLoading}
          className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
        >
          {isLoading ? 'Saving...' : 'Save Settings'}
        </button>
      </div>
    </div>
  );
};

export default MultiChannelSettings;
