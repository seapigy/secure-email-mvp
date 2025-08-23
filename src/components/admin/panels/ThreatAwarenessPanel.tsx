import React, { useState, useEffect, useCallback } from 'react';
import {
  ShieldExclamationIcon, EyeIcon, Cog6ToothIcon, BellIcon,
  ExclamationTriangleIcon, ArrowPathIcon, PencilIcon, TrashIcon,
} from '@heroicons/react/24/outline';
import {
  ThreatIntelligence, ThreatAlert, ThreatFeed, ThreatRule,
  ThreatAwarenessConfig,
} from '../../../types/admin';
import { EnterpriseDashboardService } from '../../../services/enterpriseDashboardService';

interface ThreatAwarenessPanelProps {
  dashboardService: EnterpriseDashboardService;
  isReadOnly?: boolean;
}

const ThreatAwarenessPanel: React.FC<ThreatAwarenessPanelProps> = ({
  dashboardService,
  isReadOnly = false,
}) => {
  const [threatIntelligence, setThreatIntelligence] = useState<ThreatIntelligence | null>(null);
  const [threatAlerts, setThreatAlerts] = useState<ThreatAlert[]>([]);
  const [threatFeeds, setThreatFeeds] = useState<ThreatFeed[]>([]);
  const [threatRules, setThreatRules] = useState<ThreatRule[]>([]);
  const [config, setConfig] = useState<ThreatAwarenessConfig | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'alerts' | 'feeds' | 'rules' | 'config'>('overview');
  const [selectedAlert, setSelectedAlert] = useState<ThreatAlert | null>(null);
  const [showAlertModal, setShowAlertModal] = useState(false);

  const fetchThreatData = useCallback(async () => {
    try {
      setIsLoading(true);
      setError(null);
      const [intelligence, alerts, feeds, rules, threatConfig] = await Promise.all([
        dashboardService.getThreatIntelligence(),
        dashboardService.getThreatAlerts(),
        dashboardService.getThreatFeeds(),
        dashboardService.getThreatRules(),
        dashboardService.getThreatAwarenessConfig(),
      ]);
      setThreatIntelligence(intelligence);
      setThreatAlerts(alerts);
      setThreatFeeds(feeds);
      setThreatRules(rules);
      setConfig(threatConfig);
    } catch (err: unknown) {
      const error = err as Error;
      setError(error.message || 'Failed to fetch threat data');
    } finally {
      setIsLoading(false);
    }
  }, [dashboardService]);

  useEffect(() => {
    fetchThreatData();
  }, [fetchThreatData]);

  const handleUpdateAlert = async (alertId: string, updates: Partial<ThreatAlert>) => {
    try {
      const updatedAlert = await dashboardService.updateThreatAlert(alertId, updates);
      setThreatAlerts(prev => prev.map(alert => 
        alert.id === alertId ? updatedAlert : alert
      ));
      setShowAlertModal(false);
      setSelectedAlert(null);
    } catch (err: unknown) {
      const error = err as Error;
      setError(error.message || 'Failed to update alert');
    }
  };

  const getThreatLevelColor = (level: string): string => {
    switch (level) {
      case 'critical': return 'text-red-600 bg-red-50 border-red-200';
      case 'high': return 'text-orange-600 bg-orange-50 border-orange-200';
      case 'medium': return 'text-yellow-600 bg-yellow-50 border-yellow-200';
      case 'low': return 'text-green-600 bg-green-50 border-green-200';
      default: return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const getStatusColor = (status: string): string => {
    switch (status) {
      case 'active': return 'text-red-600 bg-red-50';
      case 'investigating': return 'text-yellow-600 bg-yellow-50';
      case 'mitigated': return 'text-blue-600 bg-blue-50';
      case 'resolved': return 'text-green-600 bg-green-50';
      default: return 'text-gray-600 bg-gray-50';
    }
  };

  const getFeedStatusColor = (status: string): string => {
    switch (status) {
      case 'active': return 'text-green-600 bg-green-50';
      case 'inactive': return 'text-gray-600 bg-gray-50';
      case 'error': return 'text-red-600 bg-red-50';
      default: return 'text-gray-600 bg-gray-50';
    }
  };

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded w-1/4 mb-4"></div>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {[...Array(4)].map((_, i) => (
              <div key={i} className="h-24 bg-gray-200 rounded"></div>
            ))}
          </div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="text-center text-red-600">
          <ShieldExclamationIcon className="h-12 w-12 mx-auto mb-2" />
          <p>Error loading threat data: {error}</p>
        </div>
      </div>
    );
  }

  if (!threatIntelligence || !config) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="text-center text-gray-500">
          <ShieldExclamationIcon className="h-12 w-12 mx-auto mb-2" />
          <p>No threat data available</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border border-gray-200">
      {/* Header */}
      <div className="px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <ShieldExclamationIcon className="h-6 w-6 text-red-600" />
            <h3 className="text-lg font-semibold text-gray-900">Threat Awareness</h3>
            <span className={`px-2 py-1 rounded-full text-xs font-medium ${getThreatLevelColor(threatIntelligence.threat_level)}`}>
              {threatIntelligence.threat_level.toUpperCase()} THREAT LEVEL
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={fetchThreatData}
              className="flex items-center space-x-1 px-3 py-1 text-sm text-gray-600 hover:text-gray-900"
            >
              <ArrowPathIcon className="h-4 w-4" />
              <span>Refresh</span>
            </button>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="px-6 py-3 border-b border-gray-200">
        <nav className="flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: EyeIcon },
            { id: 'alerts', label: 'Threat Alerts', icon: BellIcon },
            { id: 'feeds', label: 'Threat Feeds', icon: Cog6ToothIcon },
            { id: 'rules', label: 'Detection Rules', icon: ShieldExclamationIcon },
            { id: 'config', label: 'Configuration', icon: Cog6ToothIcon },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as 'overview' | 'alerts' | 'feeds' | 'rules' | 'config')}
              className={`flex items-center space-x-2 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                activeTab === tab.id
                  ? 'text-red-600 bg-red-50'
                  : 'text-gray-500 hover:text-gray-700 hover:bg-gray-50'
              }`}
            >
              <tab.icon className="h-4 w-4" />
              <span>{tab.label}</span>
            </button>
          ))}
        </nav>
      </div>

      {/* Content */}
      <div className="p-6">
        {activeTab === 'overview' && (
          <div className="space-y-6">
            {/* Threat Intelligence Overview */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <div className="bg-gradient-to-r from-red-50 to-red-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-red-600">Threat Score</p>
                    <p className="text-2xl font-bold text-red-900">{threatIntelligence.threat_score}/100</p>
                  </div>
                  <ShieldExclamationIcon className="h-8 w-8 text-red-600" />
                </div>
                <p className="text-xs text-red-700 mt-2">
                  Level: {threatIntelligence.threat_level.toUpperCase()}
                </p>
              </div>

              <div className="bg-gradient-to-r from-orange-50 to-orange-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-orange-600">Active Threats</p>
                    <p className="text-2xl font-bold text-orange-900">{threatIntelligence.active_threats}</p>
                  </div>
                  <ExclamationTriangleIcon className="h-8 w-8 text-orange-600" />
                </div>
                <p className="text-xs text-orange-700 mt-2">
                  Emerging: {threatIntelligence.emerging_threats}
                </p>
              </div>

              <div className="bg-gradient-to-r from-blue-50 to-blue-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-blue-600">Threat Indicators</p>
                    <p className="text-2xl font-bold text-blue-900">{threatIntelligence.threat_indicators.length}</p>
                  </div>
                  <EyeIcon className="h-8 w-8 text-blue-600" />
                </div>
                <p className="text-xs text-blue-700 mt-2">
                  High Confidence: {threatIntelligence.threat_indicators.filter(ti => ti.confidence > 0.8).length}
                </p>
              </div>

              <div className="bg-gradient-to-r from-purple-50 to-purple-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-purple-600">Active Alerts</p>
                    <p className="text-2xl font-bold text-purple-900">{threatAlerts.filter(a => a.status === 'active').length}</p>
                  </div>
                  <BellIcon className="h-8 w-8 text-purple-600" />
                </div>
                <p className="text-xs text-purple-700 mt-2">
                  Total: {threatAlerts.length}
                </p>
              </div>
            </div>

            {/* Threat Categories */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Threat Categories</h4>
                <div className="space-y-2">
                  {Object.entries(threatIntelligence.threat_categories).map(([category, count]) => (
                    <div key={category} className="flex justify-between text-sm">
                      <span className="capitalize">{category.replace('_', ' ')}:</span>
                      <span className="font-medium">{count}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">System Status</h4>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Threat Awareness:</span>
                    <span className={`font-medium ${config.enabled ? 'text-green-600' : 'text-red-600'}`}>
                      {config.enabled ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Auto Blocking:</span>
                    <span className={`font-medium ${config.auto_blocking ? 'text-green-600' : 'text-yellow-600'}`}>
                      {config.auto_blocking ? 'Enabled' : 'Disabled'}
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Active Feeds:</span>
                    <span className="font-medium">{threatFeeds.filter(f => f.status === 'active').length}/{threatFeeds.length}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Active Rules:</span>
                    <span className="font-medium">{threatRules.filter(r => r.enabled).length}/{threatRules.length}</span>
                  </div>
                </div>
              </div>
            </div>

            {/* Recent Threat Indicators */}
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h4 className="text-sm font-medium text-gray-700 mb-3">Recent Threat Indicators</h4>
              <div className="space-y-3">
                {threatIntelligence.threat_indicators.slice(0, 5).map((indicator) => (
                  <div key={indicator.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <div className="flex items-center space-x-3">
                      <div className={`w-3 h-3 rounded-full ${getThreatLevelColor(indicator.threat_level).split(' ')[0]}`}></div>
                      <div>
                        <p className="text-sm font-medium">{indicator.value}</p>
                        <p className="text-xs text-gray-500">{indicator.description}</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-gray-500">Confidence: {(indicator.confidence * 100).toFixed(0)}%</p>
                      <p className="text-xs text-gray-500">
                        {new Date(indicator.last_seen).toLocaleDateString()}
                      </p>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Recommendations */}
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h4 className="text-sm font-medium text-gray-700 mb-3">Security Recommendations</h4>
              <div className="space-y-3">
                {threatIntelligence.recommendations.slice(0, 3).map((rec) => (
                  <div key={rec.id} className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg">
                    <div className={`w-2 h-2 rounded-full mt-2 ${getThreatLevelColor(rec.priority).split(' ')[0]}`}></div>
                    <div className="flex-1">
                      <div className="flex items-center space-x-2 mb-1">
                        <span className="text-sm font-medium">{rec.title}</span>
                        <span className={`px-1 py-0.5 rounded text-xs font-medium ${getThreatLevelColor(rec.priority)}`}>
                          {rec.priority}
                        </span>
                      </div>
                      <p className="text-xs text-gray-600 mb-2">{rec.description}</p>
                      <div className="flex items-center space-x-4 text-xs text-gray-500">
                        <span>Category: {rec.category}</span>
                        <span>Effort: {rec.estimated_effort}</span>
                        {rec.action_required && (
                          <span className="text-red-600 font-medium">Action Required</span>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'alerts' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h4 className="text-lg font-medium text-gray-900">Threat Alerts</h4>
              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-500">
                  {threatAlerts.filter(a => a.status === 'active').length} active
                </span>
              </div>
            </div>

            <div className="space-y-4">
              {threatAlerts.map((alert) => (
                <div key={alert.id} className="bg-white border border-gray-200 rounded-lg p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getThreatLevelColor(alert.threat_level)}`}>
                          {alert.threat_level.toUpperCase()}
                        </span>
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getStatusColor(alert.status)}`}>
                          {alert.status.toUpperCase()}
                        </span>
                        <span className="text-xs text-gray-500">
                          {new Date(alert.timestamp).toLocaleString()}
                        </span>
                      </div>
                      <h5 className="text-sm font-medium text-gray-900 mb-1">{alert.title}</h5>
                      <p className="text-sm text-gray-600 mb-3">{alert.description}</p>
                      
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                        <div>
                          <p className="font-medium text-gray-700 mb-1">Indicators:</p>
                          <div className="flex flex-wrap gap-1">
                            {alert.indicators.map((indicator, index) => (
                              <span key={index} className="px-2 py-1 bg-gray-100 rounded text-xs">
                                {indicator}
                              </span>
                            ))}
                          </div>
                        </div>
                        <div>
                          <p className="font-medium text-gray-700 mb-1">Affected Components:</p>
                          <div className="flex flex-wrap gap-1">
                            {alert.affected_components.map((component, index) => (
                              <span key={index} className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs">
                                {component}
                              </span>
                            ))}
                          </div>
                        </div>
                      </div>

                      <div className="mt-3">
                        <p className="font-medium text-gray-700 mb-1">Recommended Actions:</p>
                        <ul className="list-disc list-inside text-sm text-gray-600 space-y-1">
                          {alert.recommended_actions.map((action, index) => (
                            <li key={index}>{action}</li>
                          ))}
                        </ul>
                      </div>

                      {alert.notes && (
                        <div className="mt-3 p-2 bg-yellow-50 border border-yellow-200 rounded">
                          <p className="text-sm text-yellow-800">
                            <strong>Notes:</strong> {alert.notes}
                          </p>
                        </div>
                      )}
                    </div>

                    {!isReadOnly && (
                      <button
                        onClick={() => {
                          setSelectedAlert(alert);
                          setShowAlertModal(true);
                        }}
                        className="ml-4 px-3 py-1 text-sm text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded"
                      >
                        <PencilIcon className="h-4 w-4" />
                      </button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'feeds' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h4 className="text-lg font-medium text-gray-900">Threat Intelligence Feeds</h4>
              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-500">
                  {threatFeeds.filter(f => f.status === 'active').length} active feeds
                </span>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {threatFeeds.map((feed) => (
                <div key={feed.id} className="bg-white border border-gray-200 rounded-lg p-4">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h5 className="text-sm font-medium text-gray-900">{feed.name}</h5>
                      <p className="text-xs text-gray-500">{feed.description}</p>
                    </div>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${getFeedStatusColor(feed.status)}`}>
                      {feed.status.toUpperCase()}
                    </span>
                  </div>

                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-500">URL:</span>
                      <span className="font-mono text-xs truncate max-w-32">{feed.url}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Format:</span>
                      <span className="font-medium">{feed.format.toUpperCase()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Update Frequency:</span>
                      <span className="font-medium">{feed.update_frequency}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Last Update:</span>
                      <span className="font-medium">{new Date(feed.last_update).toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Next Update:</span>
                      <span className="font-medium">{new Date(feed.next_update).toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Indicators:</span>
                      <span className="font-medium">{feed.threat_indicators_count.toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-500">Confidence:</span>
                      <span className="font-medium">{(feed.confidence_score * 100).toFixed(0)}%</span>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'rules' && (
          <div className="space-y-6">
            <div className="flex items-center justify-between">
              <h4 className="text-lg font-medium text-gray-900">Detection Rules</h4>
              <div className="flex items-center space-x-2">
                <span className="text-sm text-gray-500">
                  {threatRules.filter(r => r.enabled).length} active rules
                </span>
              </div>
            </div>

            <div className="space-y-4">
              {threatRules.map((rule) => (
                <div key={rule.id} className="bg-white border border-gray-200 rounded-lg p-4">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center space-x-3 mb-2">
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${getThreatLevelColor(rule.priority)}`}>
                          {rule.priority.toUpperCase()}
                        </span>
                        <span className={`px-2 py-1 rounded-full text-xs font-medium ${rule.enabled ? 'text-green-600 bg-green-50' : 'text-gray-600 bg-gray-50'}`}>
                          {rule.enabled ? 'ENABLED' : 'DISABLED'}
                        </span>
                      </div>
                      <h5 className="text-sm font-medium text-gray-900 mb-1">{rule.name}</h5>
                      <p className="text-sm text-gray-600 mb-3">{rule.description}</p>

                      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                        <div>
                          <p className="font-medium text-gray-700 mb-1">Conditions:</p>
                          <div className="space-y-1">
                            {rule.conditions.map((condition, index) => (
                              <div key={index} className="px-2 py-1 bg-gray-100 rounded text-xs">
                                {condition.field} {condition.operator} {condition.value}
                              </div>
                            ))}
                          </div>
                        </div>
                        <div>
                          <p className="font-medium text-gray-700 mb-1">Actions:</p>
                          <div className="space-y-1">
                            {rule.actions.map((action, index) => (
                              <div key={index} className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs">
                                {action.type}: {JSON.stringify(action.parameters)}
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>

                      <div className="mt-3 text-xs text-gray-500">
                        Created: {new Date(rule.created_at).toLocaleDateString()} | 
                        Updated: {new Date(rule.updated_at).toLocaleDateString()}
                      </div>
                    </div>

                    {!isReadOnly && (
                      <div className="ml-4 flex space-x-2">
                        <button className="p-1 text-blue-600 hover:text-blue-800 hover:bg-blue-50 rounded">
                          <PencilIcon className="h-4 w-4" />
                        </button>
                        <button className="p-1 text-red-600 hover:text-red-800 hover:bg-red-50 rounded">
                          <TrashIcon className="h-4 w-4" />
                        </button>
                      </div>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'config' && (
          <div className="space-y-6">
            <h4 className="text-lg font-medium text-gray-900">Threat Awareness Configuration</h4>

            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h5 className="text-sm font-medium text-gray-700 mb-3">General Settings</h5>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
                <div className="flex justify-between">
                  <span>Threat Awareness Enabled:</span>
                  <span className={`font-medium ${config.enabled ? 'text-green-600' : 'text-red-600'}`}>
                    {config.enabled ? 'Yes' : 'No'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Auto Blocking:</span>
                  <span className={`font-medium ${config.auto_blocking ? 'text-green-600' : 'text-yellow-600'}`}>
                    {config.auto_blocking ? 'Enabled' : 'Disabled'}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span>Update Interval:</span>
                  <span className="font-medium">{config.update_interval_minutes} minutes</span>
                </div>
                <div className="flex justify-between">
                  <span>Notification Channels:</span>
                  <span className="font-medium">{config.notification_channels.join(', ')}</span>
                </div>
              </div>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h5 className="text-sm font-medium text-gray-700 mb-3">Alert Thresholds</h5>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                <div className="text-center">
                  <div className="text-2xl font-bold text-green-600">{config.alert_thresholds.low_threshold}</div>
                  <div className="text-xs text-gray-500">Low</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-yellow-600">{config.alert_thresholds.medium_threshold}</div>
                  <div className="text-xs text-gray-500">Medium</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-600">{config.alert_thresholds.high_threshold}</div>
                  <div className="text-xs text-gray-500">High</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-red-600">{config.alert_thresholds.critical_threshold}</div>
                  <div className="text-xs text-gray-500">Critical</div>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Alert Update Modal */}
      {showAlertModal && selectedAlert && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Update Threat Alert</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
                <select
                  value={selectedAlert.status}
                  onChange={(e) => setSelectedAlert({...selectedAlert, status: e.target.value as 'active' | 'investigating' | 'mitigated' | 'resolved'})}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                >
                  <option value="active">Active</option>
                  <option value="investigating">Investigating</option>
                  <option value="mitigated">Mitigated</option>
                  <option value="resolved">Resolved</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Assigned To</label>
                <input
                  type="text"
                  value={selectedAlert.assigned_to || ''}
                  onChange={(e) => setSelectedAlert({...selectedAlert, assigned_to: e.target.value})}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  placeholder="admin@company.com"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Notes</label>
                <textarea
                  value={selectedAlert.notes || ''}
                  onChange={(e) => setSelectedAlert({...selectedAlert, notes: e.target.value})}
                  className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm"
                  rows={3}
                  placeholder="Add investigation notes..."
                />
              </div>
            </div>
            <div className="flex justify-end space-x-3 mt-6">
              <button
                onClick={() => {
                  setShowAlertModal(false);
                  setSelectedAlert(null);
                }}
                className="px-4 py-2 text-sm text-gray-600 hover:text-gray-800"
              >
                Cancel
              </button>
              <button
                onClick={() => handleUpdateAlert(selectedAlert.id, {
                  status: selectedAlert.status,
                  assigned_to: selectedAlert.assigned_to,
                  notes: selectedAlert.notes,
                })}
                className="px-4 py-2 text-sm bg-blue-600 text-white rounded-md hover:bg-blue-700"
              >
                Update Alert
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ThreatAwarenessPanel;
