import React, { useState, useEffect } from 'react';
import { 
  AlertTriangle, 
  Shield, 
  Eye, 
  MapPin, 
  Clock, 
  Globe, 
  Smartphone,
  Monitor,
  Tablet,
  Laptop,
  Server,
  Settings,
  CheckCircle,
  XCircle,
  Lock,
  Activity
} from 'lucide-react';
import { apiClient } from '../../lib/api';

interface SuspiciousAccessEvent {
  event_id: string;
  email_id: string;
  user_id: string;
  ip_address: string;
  user_agent: string;
  timestamp: string;
  event_type: 'failed_login' | 'multiple_attempts' | 'unusual_location' | 'unusual_device' | 'unusual_time' | 'suspicious_pattern';
  severity: 'low' | 'medium' | 'high' | 'critical';
  risk_score: number;
  country: string;
  city: string;
  device_type: string;
  browser: string;
  os: string;
  is_mobile: boolean;
  is_tor: boolean;
  is_vpn: boolean;
  is_proxy: boolean;
  threat_intel_match: boolean;
  threat_intel_details?: Record<string, unknown>;
  auto_blocked: boolean;
  manual_reviewed: boolean;
  review_notes?: string;
  created_at: string;
}

interface SuspiciousAccessSettings {
  enabled: boolean;
  riskThreshold: number;
  autoBlockEnabled: boolean;
  autoBlockThreshold: number;
  notificationEnabled: boolean;
  emailNotifications: boolean;
  smsNotifications: boolean;
  pushNotifications: boolean;
  geoFencingEnabled: boolean;
  allowedCountries: string[];
  blockedCountries: string[];
  timeRestrictionsEnabled: boolean;
  allowedTimeRanges: Array<{
    dayOfWeek: number;
    startTime: string;
    endTime: string;
  }>;
  deviceRestrictionsEnabled: boolean;
  allowedDeviceTypes: string[];
  vpnDetectionEnabled: boolean;
  torDetectionEnabled: boolean;
  proxyDetectionEnabled: boolean;
  threatIntelligenceEnabled: boolean;
  threatIntelligenceProviders: string[];
}

const SuspiciousAccessDashboard: React.FC = () => {
  const [events, setEvents] = useState<SuspiciousAccessEvent[]>([]);
  const [settings, setSettings] = useState<SuspiciousAccessSettings>({
    enabled: true,
    riskThreshold: 70,
    autoBlockEnabled: true,
    autoBlockThreshold: 85,
    notificationEnabled: true,
    emailNotifications: true,
    smsNotifications: false,
    pushNotifications: true,
    geoFencingEnabled: false,
    allowedCountries: [],
    blockedCountries: [],
    timeRestrictionsEnabled: false,
    allowedTimeRanges: [],
    deviceRestrictionsEnabled: false,
    allowedDeviceTypes: [],
    vpnDetectionEnabled: true,
    torDetectionEnabled: true,
    proxyDetectionEnabled: true,
    threatIntelligenceEnabled: true,
    threatIntelligenceProviders: ['abuseipdb', 'virustotal']
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);
  const [selectedEvent, setSelectedEvent] = useState<SuspiciousAccessEvent | null>(null);
  const [showEventDetails, setShowEventDetails] = useState(false);
  const [showSettings, setShowSettings] = useState(false);

  useEffect(() => {
    loadEvents();
    loadSettings();
  }, []);

  const loadEvents = async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await apiClient.get('/api/security/suspicious-access/events?limit=50');
      setEvents(response.data.events || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load suspicious access events');
    } finally {
      setLoading(false);
    }
  };

  const loadSettings = async () => {
    try {
      const response = await apiClient.get('/api/security/suspicious-access/settings');
      setSettings(response.data);
    } catch (err) {
      console.error('Failed to load settings:', err);
    }
  };

  const saveSettings = async () => {
    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      await apiClient.put('/api/security/suspicious-access/settings', settings);
      setSuccess('Settings saved successfully');
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to save settings');
    } finally {
      setLoading(false);
    }
  };

  const reviewEvent = async (eventId: string, action: 'allow' | 'block', notes?: string) => {
    try {
      await apiClient.post(`/api/security/suspicious-access/events/${eventId}/review`, {
        action,
        notes
      });
      
      // Refresh events
      loadEvents();
      setShowEventDetails(false);
      setSelectedEvent(null);
      setSuccess(`Event ${action}ed successfully`);
      setTimeout(() => setSuccess(null), 3000);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to review event');
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'text-red-600 bg-red-100 dark:bg-red-900/20';
      case 'high': return 'text-orange-600 bg-orange-100 dark:bg-orange-900/20';
      case 'medium': return 'text-yellow-600 bg-yellow-100 dark:bg-yellow-900/20';
      case 'low': return 'text-blue-600 bg-blue-100 dark:bg-blue-900/20';
      default: return 'text-gray-600 bg-gray-100 dark:bg-gray-900/20';
    }
  };

  const getEventTypeIcon = (eventType: string) => {
    switch (eventType) {
      case 'failed_login': return <Lock className="h-4 w-4" />;
      case 'multiple_attempts': return <Activity className="h-4 w-4" />;
      case 'unusual_location': return <MapPin className="h-4 w-4" />;
      case 'unusual_device': return <Smartphone className="h-4 w-4" />;
      case 'unusual_time': return <Clock className="h-4 w-4" />;
      case 'suspicious_pattern': return <AlertTriangle className="h-4 w-4" />;
      default: return <Eye className="h-4 w-4" />;
    }
  };

  const getDeviceIcon = (deviceType: string) => {
    switch (deviceType.toLowerCase()) {
      case 'mobile': return <Smartphone className="h-4 w-4" />;
      case 'tablet': return <Tablet className="h-4 w-4" />;
      case 'desktop': return <Monitor className="h-4 w-4" />;
      case 'laptop': return <Laptop className="h-4 w-4" />;
      case 'server': return <Server className="h-4 w-4" />;
      default: return <Monitor className="h-4 w-4" />;
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const maskIP = (ip: string) => {
    const parts = ip.split('.');
    return `${parts[0]}.${parts[1]}.*.*`;
  };

  const getRiskLevel = (score: number) => {
    if (score >= 90) return 'Critical';
    if (score >= 75) return 'High';
    if (score >= 50) return 'Medium';
    return 'Low';
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-red-100 dark:bg-red-900/20 rounded-lg">
              <Shield className="h-6 w-6 text-red-600 dark:text-red-400" />
            </div>
            <div>
              <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
                Suspicious Access Dashboard
              </h2>
              <p className="text-gray-600 dark:text-gray-400">
                Monitor and manage suspicious access attempts
              </p>
            </div>
          </div>
          <div className="flex space-x-2">
            <button
              onClick={() => setShowSettings(true)}
              className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-red-500"
            >
              <Settings className="h-4 w-4 mr-2 inline" />
              Settings
            </button>
            <button
              onClick={loadEvents}
              disabled={loading}
              className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50"
            >
              {loading ? 'Loading...' : 'Refresh'}
            </button>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
          <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg p-4">
            <div className="flex items-center">
              <AlertTriangle className="h-5 w-5 text-red-600 dark:text-red-400" />
              <div className="ml-3">
                <p className="text-sm font-medium text-red-600 dark:text-red-400">Critical Events</p>
                <p className="text-2xl font-bold text-red-600 dark:text-red-400">
                  {events.filter(e => e.severity === 'critical').length}
                </p>
              </div>
            </div>
          </div>
          <div className="bg-orange-50 dark:bg-orange-900/20 border border-orange-200 dark:border-orange-800 rounded-lg p-4">
            <div className="flex items-center">
              <Shield className="h-5 w-5 text-orange-600 dark:text-orange-400" />
              <div className="ml-3">
                <p className="text-sm font-medium text-orange-600 dark:text-orange-400">High Risk</p>
                <p className="text-2xl font-bold text-orange-600 dark:text-orange-400">
                  {events.filter(e => e.severity === 'high').length}
                </p>
              </div>
            </div>
          </div>
          <div className="bg-yellow-50 dark:bg-yellow-900/20 border border-yellow-200 dark:border-yellow-800 rounded-lg p-4">
            <div className="flex items-center">
              <Eye className="h-5 w-5 text-yellow-600 dark:text-yellow-400" />
              <div className="ml-3">
                <p className="text-sm font-medium text-yellow-600 dark:text-yellow-400">Auto Blocked</p>
                <p className="text-2xl font-bold text-yellow-600 dark:text-yellow-400">
                  {events.filter(e => e.auto_blocked).length}
                </p>
              </div>
            </div>
          </div>
          <div className="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
            <div className="flex items-center">
              <Activity className="h-5 w-5 text-blue-600 dark:text-blue-400" />
              <div className="ml-3">
                <p className="text-sm font-medium text-blue-600 dark:text-blue-400">Total Events</p>
                <p className="text-2xl font-bold text-blue-600 dark:text-blue-400">
                  {events.length}
                </p>
              </div>
            </div>
          </div>
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
      </div>

      {/* Events Table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="text-lg font-medium text-gray-900 dark:text-white">Recent Suspicious Events</h3>
        </div>
        
        {events.length === 0 ? (
          <div className="p-6 text-center text-gray-500 dark:text-gray-400">
            No suspicious access events found.
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
              <thead className="bg-gray-50 dark:bg-gray-700">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Event
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    IP Address
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Location
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Device
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Risk Score
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Time
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-300 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
                {events.map((event) => (
                  <tr key={event.event_id} className="hover:bg-gray-50 dark:hover:bg-gray-700">
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        {getEventTypeIcon(event.event_type)}
                        <div className="ml-3">
                          <div className="text-sm font-medium text-gray-900 dark:text-white">
                            {event.event_type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                          </div>
                          <div className="text-sm text-gray-500 dark:text-gray-400">
                            {event.email_id}
                          </div>
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                      {maskIP(event.ip_address)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <Globe className="h-4 w-4 text-gray-400 mr-1" />
                        <div className="text-sm text-gray-900 dark:text-white">
                          {event.city}, {event.country}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        {getDeviceIcon(event.device_type)}
                        <div className="ml-2 text-sm text-gray-900 dark:text-white">
                          {event.device_type}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      <div className="flex items-center">
                        <div className="text-sm font-medium text-gray-900 dark:text-white">
                          {event.risk_score}
                        </div>
                        <div className={`ml-2 px-2 py-1 text-xs font-medium rounded-full ${getSeverityColor(event.severity)}`}>
                          {getRiskLevel(event.risk_score)}
                        </div>
                      </div>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap">
                      {event.auto_blocked ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200">
                          <Lock className="h-3 w-3 mr-1" />
                          Blocked
                        </span>
                      ) : event.manual_reviewed ? (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200">
                          <CheckCircle className="h-3 w-3 mr-1" />
                          Reviewed
                        </span>
                      ) : (
                        <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900 dark:text-yellow-200">
                          <Eye className="h-3 w-3 mr-1" />
                          Pending
                        </span>
                      )}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                      {formatTimestamp(event.timestamp)}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <button
                        onClick={() => {
                          setSelectedEvent(event);
                          setShowEventDetails(true);
                        }}
                        className="text-red-600 hover:text-red-900 dark:text-red-400 dark:hover:text-red-300"
                      >
                        Review
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Event Details Modal */}
      {showEventDetails && selectedEvent && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-1/2 shadow-lg rounded-md bg-white dark:bg-gray-800">
            <div className="mt-3">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                  Event Details
                </h3>
                <button
                  onClick={() => {
                    setShowEventDetails(false);
                    setSelectedEvent(null);
                  }}
                  className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                >
                  <XCircle className="h-6 w-6" />
                </button>
              </div>
              
              <div className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Event Type</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.event_type}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Severity</label>
                    <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getSeverityColor(selectedEvent.severity)}`}>
                      {selectedEvent.severity}
                    </span>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">IP Address</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.ip_address}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Location</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.city}, {selectedEvent.country}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Device Type</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.device_type}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Browser</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.browser}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Operating System</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.os}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Risk Score</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.risk_score}</p>
                  </div>
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">User Agent</label>
                  <p className="text-sm text-gray-900 dark:text-white break-all">{selectedEvent.user_agent}</p>
                </div>
                
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">VPN Detection</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.is_vpn ? 'Yes' : 'No'}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Tor Network</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.is_tor ? 'Yes' : 'No'}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Proxy Detection</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.is_proxy ? 'Yes' : 'No'}</p>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Threat Intel Match</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.threat_intel_match ? 'Yes' : 'No'}</p>
                  </div>
                </div>
                
                {selectedEvent.threat_intel_details && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Threat Intelligence Details</label>
                    <pre className="text-sm text-gray-900 dark:text-white bg-gray-100 dark:bg-gray-700 p-2 rounded overflow-auto">
                      {JSON.stringify(selectedEvent.threat_intel_details, null, 2)}
                    </pre>
                  </div>
                )}
                
                {selectedEvent.review_notes && (
                  <div>
                    <label className="block text-sm font-medium text-gray-700 dark:text-gray-300">Review Notes</label>
                    <p className="text-sm text-gray-900 dark:text-white">{selectedEvent.review_notes}</p>
                  </div>
                )}
              </div>
              
              <div className="flex justify-end space-x-3 mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
                <button
                  onClick={() => reviewEvent(selectedEvent.event_id, 'allow')}
                  className="px-4 py-2 text-sm font-medium text-white bg-green-600 hover:bg-green-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-green-500"
                >
                  Allow
                </button>
                <button
                  onClick={() => reviewEvent(selectedEvent.event_id, 'block')}
                  className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500"
                >
                  Block
                </button>
                <button
                  onClick={() => {
                    setShowEventDetails(false);
                    setSelectedEvent(null);
                  }}
                  className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Settings Modal */}
      {showSettings && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full z-50">
          <div className="relative top-20 mx-auto p-5 border w-11/12 md:w-3/4 lg:w-2/3 shadow-lg rounded-md bg-white dark:bg-gray-800">
            <div className="mt-3">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900 dark:text-white">
                  Suspicious Access Settings
                </h3>
                <button
                  onClick={() => setShowSettings(false)}
                  className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
                >
                  <XCircle className="h-6 w-6" />
                </button>
              </div>
              
              <div className="space-y-6">
                {/* General Settings */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 dark:text-white mb-3">General Settings</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.enabled}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, enabled: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Enable suspicious access detection</span>
                      </label>
                    </div>
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.autoBlockEnabled}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, autoBlockEnabled: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Enable automatic blocking</span>
                      </label>
                    </div>
                  </div>
                </div>

                {/* Thresholds */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 dark:text-white mb-3">Risk Thresholds</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Detection Threshold
                      </label>
                      <input
                        type="number"
                        min="0"
                        max="100"
                        value={settings.riskThreshold}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, riskThreshold: parseInt(e.target.value) }))}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                        Auto-Block Threshold
                      </label>
                      <input
                        type="number"
                        min="0"
                        max="100"
                        value={settings.autoBlockThreshold}
                        onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, autoBlockThreshold: parseInt(e.target.value) }))}
                        className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:ring-2 focus:ring-red-500 focus:border-red-500 bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
                      />
                    </div>
                  </div>
                </div>

                {/* Notifications */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 dark:text-white mb-3">Notifications</h4>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.emailNotifications}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, emailNotifications: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Email notifications</span>
                      </label>
                    </div>
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.smsNotifications}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, smsNotifications: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">SMS notifications</span>
                      </label>
                    </div>
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.pushNotifications}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, pushNotifications: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Push notifications</span>
                      </label>
                    </div>
                  </div>
                </div>

                {/* Detection Features */}
                <div>
                  <h4 className="text-md font-medium text-gray-900 dark:text-white mb-3">Detection Features</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.vpnDetectionEnabled}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, vpnDetectionEnabled: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">VPN detection</span>
                      </label>
                    </div>
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.torDetectionEnabled}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, torDetectionEnabled: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Tor network detection</span>
                      </label>
                    </div>
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.proxyDetectionEnabled}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, proxyDetectionEnabled: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Proxy detection</span>
                      </label>
                    </div>
                    <div>
                      <label className="flex items-center">
                        <input
                          type="checkbox"
                          checked={settings.threatIntelligenceEnabled}
                          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSettings(prev => ({ ...prev, threatIntelligenceEnabled: e.target.checked }))}
                          className="rounded border-gray-300 text-red-600 focus:ring-red-500"
                        />
                        <span className="ml-2 text-sm text-gray-700 dark:text-gray-300">Threat intelligence</span>
                      </label>
                    </div>
                  </div>
                </div>
              </div>
              
              <div className="flex justify-end space-x-3 mt-6 pt-4 border-t border-gray-200 dark:border-gray-700">
                <button
                  onClick={() => setShowSettings(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-600 focus:outline-none focus:ring-2 focus:ring-gray-500"
                >
                  Cancel
                </button>
                <button
                  onClick={saveSettings}
                  disabled={loading}
                  className="px-4 py-2 text-sm font-medium text-white bg-red-600 hover:bg-red-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50"
                >
                  {loading ? 'Saving...' : 'Save Settings'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default SuspiciousAccessDashboard;

