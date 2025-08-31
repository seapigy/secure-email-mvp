// =============================================================================
// SECURE EMAIL MVP - NOTIFICATION PREFERENCES COMPONENT
// =============================================================================
// Component for managing notification preferences and access event history.
// =============================================================================

import React, { useState, useEffect } from 'react';
import { log } from '@/lib/logger';
import { Bell, Mail, Smartphone, MapPin, Monitor, Save, History, Eye, EyeOff } from 'lucide-react';
import { toast } from 'react-toastify';
import {
  getNotificationPreferences,
  updateNotificationPreferences,
  getAccessEventHistory
} from '@/lib/api';

// TODO: Define these types properly
interface NotificationPreferences {
  email_notifications: boolean;
  sms_notifications: boolean;
  push_notifications: boolean;
  digest_enabled: boolean;
  digest_frequency: 'daily' | 'weekly' | 'monthly';
  notify_on_success: boolean;
  notify_on_failure: boolean;
  notify_on_blocked: boolean;
  include_geolocation: boolean;
  include_device_info: boolean;
}

interface AccessEvent {
  event_id: string;
  email_id: string;
  timestamp: string;
  event_type: string;
  ip_address: string;
  user_agent: string;
  country: string;
  city: string;
  device_type?: string;
  failure_reason?: string;
}

interface NotificationPreferencesProps {
  isOpen: boolean;
  onClose: () => void;
}

const NotificationPreferences: React.FC<NotificationPreferencesProps> = ({
  isOpen,
  onClose
}) => {
  const [preferences, setPreferences] = useState<NotificationPreferences | null>(null);
  const [accessHistory, setAccessHistory] = useState<AccessEvent[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [activeTab, setActiveTab] = useState<'preferences' | 'history'>('preferences');

  useEffect(() => {
    if (isOpen) {
      loadPreferences();
      loadAccessHistory();
    }
  }, [isOpen]);

  const loadPreferences = async () => {
    try {
      setLoading(true);
      const prefs = await getNotificationPreferences();
      setPreferences(prefs);
    } catch (error) {
      log.error('Failed to load notification preferences:', error, 'NotificationPreferences');
      toast.error('Failed to load notification preferences');
    } finally {
      setLoading(false);
    }
  };

  const loadAccessHistory = async () => {
    try {
      const history = await getAccessEventHistory();
      setAccessHistory(history);
    } catch (error) {
      log.error('Failed to load access history:', error, 'NotificationPreferences');
      toast.error('Failed to load access history');
    }
  };

  const handlePreferenceChange = (key: keyof NotificationPreferences, value: boolean) => {
    if (!preferences) return;
    setPreferences({
      ...preferences,
      [key]: value
    });
  };

  const handleSavePreferences = async () => {
    if (!preferences) return;

    try {
      setSaving(true);
      await updateNotificationPreferences(preferences);
      toast.success('Notification preferences updated successfully');
    } catch (error) {
      log.error('Failed to update notification preferences:', error, 'NotificationPreferences');
      toast.error('Failed to update notification preferences');
    } finally {
      setSaving(false);
    }
  };

  const getEventTypeIcon = (eventType: string) => {
    switch (eventType) {
      case 'success':
        return <Eye className="w-4 h-4 text-green-500" />;
      case 'failure':
        return <EyeOff className="w-4 h-4 text-red-500" />;
      case 'blocked':
        return <EyeOff className="w-4 h-4 text-orange-500" />;
      default:
        return <Eye className="w-4 h-4 text-gray-500" />;
    }
  };

  const getEventTypeLabel = (eventType: string) => {
    switch (eventType) {
      case 'success':
        return 'Success';
      case 'failure':
        return 'Failure';
      case 'blocked':
        return 'Blocked';
      default:
        return eventType;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow-xl max-w-4xl w-full mx-4 max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-gray-200 dark:border-gray-700">
          <div className="flex items-center space-x-3">
            <Bell className="w-6 h-6 text-blue-600 dark:text-blue-400" />
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white">
              Notification Settings
            </h2>
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
          >
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-gray-200 dark:border-gray-700">
          <button
            onClick={() => setActiveTab('preferences')}
            className={`flex-1 py-3 px-4 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'preferences'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            <Bell className="w-4 h-4 inline mr-2" />
            Preferences
          </button>
          <button
            onClick={() => setActiveTab('history')}
            className={`flex-1 py-3 px-4 text-sm font-medium border-b-2 transition-colors ${
              activeTab === 'history'
                ? 'border-blue-500 text-blue-600 dark:text-blue-400'
                : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
            }`}
          >
            <History className="w-4 h-4 inline mr-2" />
            Access History
          </button>
        </div>

        {/* Content */}
        <div className="p-6 overflow-y-auto max-h-[calc(90vh-140px)]">
          {loading ? (
            <div className="flex items-center justify-center py-8">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
            </div>
          ) : activeTab === 'preferences' ? (
            <div className="space-y-6">
              {/* Notification Channels */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                  Notification Channels
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <Mail className="w-5 h-5 text-blue-600 dark:text-blue-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">Email Notifications</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          Receive notifications via email
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.email_notifications || false}
                        onChange={(e) => handlePreferenceChange('email_notifications', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <Smartphone className="w-5 h-5 text-green-600 dark:text-green-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">SMS Notifications</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          Receive notifications via SMS
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.sms_notifications || false}
                        onChange={(e) => handlePreferenceChange('sms_notifications', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>
                </div>
              </div>

              {/* Event Types */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                  Event Types
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <Eye className="w-5 h-5 text-green-600 dark:text-green-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">Successful Access</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          When someone successfully accesses your email
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.notify_on_success || false}
                        onChange={(e) => handlePreferenceChange('notify_on_success', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <EyeOff className="w-5 h-5 text-red-600 dark:text-red-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">Failed Access</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          When someone fails to access your email
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.notify_on_failure || false}
                        onChange={(e) => handlePreferenceChange('notify_on_failure', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <EyeOff className="w-5 h-5 text-orange-600 dark:text-orange-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">Blocked Access</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          When access is blocked by security measures
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.notify_on_blocked || false}
                        onChange={(e) => handlePreferenceChange('notify_on_blocked', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>
                </div>
              </div>

              {/* Information to Include */}
              <div>
                <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                  Information to Include
                </h3>
                <div className="space-y-4">
                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <MapPin className="w-5 h-5 text-purple-600 dark:text-purple-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">Geolocation</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          Include location information in notifications
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.include_geolocation || false}
                        onChange={(e) => handlePreferenceChange('include_geolocation', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <Monitor className="w-5 h-5 text-indigo-600 dark:text-indigo-400" />
                      <div>
                        <h4 className="font-medium text-gray-900 dark:text-white">Device Information</h4>
                        <p className="text-sm text-gray-500 dark:text-gray-400">
                          Include device type and browser information
                        </p>
                      </div>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input
                        type="checkbox"
                        checked={preferences?.include_device_info || false}
                        onChange={(e) => handlePreferenceChange('include_device_info', e.target.checked)}
                        className="sr-only peer"
                      />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 dark:peer-focus:ring-blue-800 rounded-full peer dark:bg-gray-700 peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all dark:border-gray-600 peer-checked:bg-blue-600"></div>
                    </label>
                  </div>
                </div>
              </div>

              {/* Save Button */}
              <div className="flex justify-end pt-4">
                <button
                  onClick={handleSavePreferences}
                  disabled={saving}
                  className="flex items-center space-x-2 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  <Save className="w-4 h-4" />
                  <span>{saving ? 'Saving...' : 'Save Preferences'}</span>
                </button>
              </div>
            </div>
          ) : (
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900 dark:text-white mb-4">
                Access Event History
              </h3>
              
              {accessHistory.length === 0 ? (
                <div className="text-center py-8 text-gray-500 dark:text-gray-400">
                  <History className="w-12 h-12 mx-auto mb-4 opacity-50" />
                  <p>No access events found</p>
                </div>
              ) : (
                <div className="space-y-3">
                  {accessHistory.map((event) => (
                    <div
                      key={event.event_id}
                      className="p-4 border border-gray-200 dark:border-gray-700 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex items-center space-x-3">
                          {getEventTypeIcon(event.event_type)}
                          <div>
                            <div className="flex items-center space-x-2">
                              <span className="font-medium text-gray-900 dark:text-white">
                                {getEventTypeLabel(event.event_type)}
                              </span>
                              <span className="text-sm text-gray-500 dark:text-gray-400">
                                {formatTimestamp(event.timestamp)}
                              </span>
                            </div>
                            <div className="text-sm text-gray-600 dark:text-gray-300 mt-1">
                              <p>IP: {event.ip_address}</p>
                              {event.country && (
                                <p>Location: {event.city ? `${event.city}, ` : ''}{event.country}</p>
                              )}
                              {event.device_type && (
                                <p>Device: {event.device_type}</p>
                              )}
                              {event.failure_reason && (
                                <p className="text-red-600 dark:text-red-400">
                                  Reason: {event.failure_reason}
                                </p>
                              )}
                            </div>
                          </div>
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
    </div>
  );
};

export default NotificationPreferences;
