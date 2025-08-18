import React, { useState, useEffect } from 'react';
import {
  ChartBarIcon, ArrowTrendingUpIcon, DocumentChartBarIcon, CalendarIcon, FunnelIcon, EyeIcon, EyeSlashIcon,
} from '@heroicons/react/24/outline';
import {
  AnalyticsDashboard, AnalyticsTimeRange,
} from '../../../types/admin';

interface AnalyticsDashboardPanelProps {
  dashboardService: any;
  isReadOnly?: boolean;
}

const AnalyticsDashboardPanel: React.FC<AnalyticsDashboardPanelProps> = ({
  dashboardService,
  isReadOnly = false,
}) => {
  const [analyticsData, setAnalyticsData] = useState<AnalyticsDashboard | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedTimeRange, setSelectedTimeRange] = useState<AnalyticsTimeRange>({
    start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
    end: new Date().toISOString(),
    granularity: 'day'
  });
  const [showTrends, setShowTrends] = useState(true);
  const [showDetails, setShowDetails] = useState(false);
  const [activeTab, setActiveTab] = useState<'overview' | 'email' | 'security' | 'zkid-pqc' | 'threats'>('overview');

  useEffect(() => {
    fetchAnalyticsData();
  }, [selectedTimeRange]);

  const fetchAnalyticsData = async () => {
    try {
      setIsLoading(true);
      setError(null);
      const data = await dashboardService.getAnalyticsDashboard({ time_range: selectedTimeRange });
      setAnalyticsData(data);
    } catch (err: any) {
      setError(err.message || 'Failed to fetch analytics data');
    } finally {
      setIsLoading(false);
    }
  };

  const formatBytes = (bytes: number): string => {
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
    if (bytes === 0) return '0 Bytes';
    const i = Math.floor(Math.log(bytes) / Math.log(1024));
    return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i];
  };

  const formatPercentage = (value: number): string => {
    return (value * 100).toFixed(1) + '%';
  };

  const getThreatLevelColor = (level: string): string => {
    switch (level) {
      case 'critical': return 'text-red-600 bg-red-50';
      case 'high': return 'text-orange-600 bg-orange-50';
      case 'medium': return 'text-yellow-600 bg-yellow-50';
      case 'low': return 'text-green-600 bg-green-50';
      default: return 'text-gray-600 bg-gray-50';
    }
  };

  const getSeverityColor = (severity: string): string => {
    switch (severity) {
      case 'critical': return 'bg-red-500';
      case 'high': return 'bg-orange-500';
      case 'medium': return 'bg-yellow-500';
      case 'low': return 'bg-green-500';
      default: return 'bg-gray-500';
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
          <ChartBarIcon className="h-12 w-12 mx-auto mb-2" />
          <p>Error loading analytics: {error}</p>
        </div>
      </div>
    );
  }

  if (!analyticsData) {
    return (
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="text-center text-gray-500">
          <ChartBarIcon className="h-12 w-12 mx-auto mb-2" />
          <p>No analytics data available</p>
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
            <ChartBarIcon className="h-6 w-6 text-blue-600" />
            <h3 className="text-lg font-semibold text-gray-900">Analytics Dashboard</h3>
            <span className="text-sm text-gray-500">
              Last updated: {new Date(analyticsData.last_updated).toLocaleString()}
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowTrends(!showTrends)}
              className="flex items-center space-x-1 px-3 py-1 text-sm text-gray-600 hover:text-gray-900"
            >
              {showTrends ? <EyeSlashIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
              <span>{showTrends ? 'Hide' : 'Show'} Trends</span>
            </button>
            <button
              onClick={() => setShowDetails(!showDetails)}
              className="flex items-center space-x-1 px-3 py-1 text-sm text-gray-600 hover:text-gray-900"
            >
              <FunnelIcon className="h-4 w-4" />
              <span>{showDetails ? 'Hide' : 'Show'} Details</span>
            </button>
          </div>
        </div>
      </div>

      {/* Time Range Selector */}
      <div className="px-6 py-3 bg-gray-50 border-b border-gray-200">
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <CalendarIcon className="h-4 w-4 text-gray-500" />
            <span className="text-sm font-medium text-gray-700">Time Range:</span>
          </div>
          <select
            value={selectedTimeRange.granularity}
            onChange={(e) => setSelectedTimeRange({
              ...selectedTimeRange,
              granularity: e.target.value as any
            })}
            className="text-sm border border-gray-300 rounded-md px-3 py-1"
            disabled={isReadOnly}
          >
            <option value="hour">Last 24 Hours</option>
            <option value="day">Last 7 Days</option>
            <option value="week">Last 4 Weeks</option>
            <option value="month">Last 3 Months</option>
          </select>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="px-6 py-3 border-b border-gray-200">
        <nav className="flex space-x-8">
          {[
            { id: 'overview', label: 'Overview', icon: ChartBarIcon },
            { id: 'email', label: 'Email Analytics', icon: DocumentChartBarIcon },
            { id: 'security', label: 'Security Events', icon: ArrowTrendingUpIcon },
            { id: 'zkid-pqc', label: 'ZKID & PQC', icon: ChartBarIcon },
            { id: 'threats', label: 'Threat Intelligence', icon: ArrowTrendingUpIcon },
          ].map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`flex items-center space-x-2 px-3 py-2 text-sm font-medium rounded-md transition-colors ${
                activeTab === tab.id
                  ? 'text-blue-600 bg-blue-50'
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
            {/* Key Metrics */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
              <div className="bg-gradient-to-r from-blue-50 to-blue-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-blue-600">Total Emails</p>
                    <p className="text-2xl font-bold text-blue-900">
                      {(analyticsData.email_usage.total_emails_sent + analyticsData.email_usage.total_emails_received).toLocaleString()}
                    </p>
                  </div>
                                     <DocumentChartBarIcon className="h-8 w-8 text-blue-600" />
                </div>
                <p className="text-xs text-blue-700 mt-2">
                  Sent: {analyticsData.email_usage.total_emails_sent.toLocaleString()} | 
                  Received: {analyticsData.email_usage.total_emails_received.toLocaleString()}
                </p>
              </div>

              <div className="bg-gradient-to-r from-red-50 to-red-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-red-600">Security Events</p>
                    <p className="text-2xl font-bold text-red-900">
                      {analyticsData.security_events.total_security_events.toLocaleString()}
                    </p>
                  </div>
                                     <ArrowTrendingUpIcon className="h-8 w-8 text-red-600" />
                </div>
                <p className="text-xs text-red-700 mt-2">
                  Critical: {analyticsData.security_events.events_by_severity.critical} | 
                  High: {analyticsData.security_events.events_by_severity.high}
                </p>
              </div>

              <div className="bg-gradient-to-r from-green-50 to-green-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-green-600">ZKID Operations</p>
                    <p className="text-2xl font-bold text-green-900">
                      {analyticsData.zkid_pqc.zkid_operations.total_mappings_created.toLocaleString()}
                    </p>
                  </div>
                  <ChartBarIcon className="h-8 w-8 text-green-600" />
                </div>
                <p className="text-xs text-green-700 mt-2">
                  Active: {analyticsData.zkid_pqc.zkid_operations.active_mappings.toLocaleString()} | 
                  Success Rate: {formatPercentage(analyticsData.zkid_pqc.security_metrics.zero_knowledge_compliance)}
                </p>
              </div>

              <div className="bg-gradient-to-r from-purple-50 to-purple-100 rounded-lg p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium text-purple-600">Threat Level</p>
                    <p className="text-2xl font-bold text-purple-900">
                      {analyticsData.threat_intelligence.threat_level.toUpperCase()}
                    </p>
                  </div>
                                     <ArrowTrendingUpIcon className="h-8 w-8 text-purple-600" />
                </div>
                <p className="text-xs text-purple-700 mt-2">
                  Score: {analyticsData.threat_intelligence.threat_score}/100 | 
                  Active: {analyticsData.threat_intelligence.active_threats}
                </p>
              </div>
            </div>

            {/* Trend Charts */}
            {showTrends && (
              <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                <div className="bg-gray-50 rounded-lg p-4">
                  <h4 className="text-sm font-medium text-gray-700 mb-3">Email Activity Trends</h4>
                  <div className="h-48 bg-white rounded border flex items-center justify-center">
                    <span className="text-gray-500 text-sm">Chart visualization would be rendered here</span>
                  </div>
                </div>
                <div className="bg-gray-50 rounded-lg p-4">
                  <h4 className="text-sm font-medium text-gray-700 mb-3">Security Events Trends</h4>
                  <div className="h-48 bg-white rounded border flex items-center justify-center">
                    <span className="text-gray-500 text-sm">Chart visualization would be rendered here</span>
                  </div>
                </div>
              </div>
            )}

            {/* Quick Stats */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Storage Usage</h4>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Total Storage:</span>
                    <span className="font-medium">{formatBytes(analyticsData.email_usage.storage_usage.total_bytes)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Growth Rate:</span>
                    <span className="font-medium">{analyticsData.email_usage.storage_usage.growth_rate_percent}%</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Compression:</span>
                    <span className="font-medium">{formatPercentage(analyticsData.email_usage.storage_usage.compression_ratio)}</span>
                  </div>
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Performance Metrics</h4>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Avg Delivery Time:</span>
                    <span className="font-medium">{analyticsData.email_usage.delivery_times.average_seconds}s</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>ZKID Latency:</span>
                    <span className="font-medium">{analyticsData.zkid_pqc.performance_metrics.zkid_creation_latency_ms}ms</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>PQC Latency:</span>
                    <span className="font-medium">{analyticsData.zkid_pqc.performance_metrics.pqc_encryption_latency_ms}ms</span>
                  </div>
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Security Score</h4>
                <div className="space-y-2">
                  <div className="flex justify-between text-sm">
                    <span>Encryption Success:</span>
                    <span className="font-medium">{formatPercentage(analyticsData.zkid_pqc.security_metrics.encryption_success_rate)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Key Rotation:</span>
                    <span className="font-medium">{formatPercentage(analyticsData.zkid_pqc.security_metrics.key_rotation_success_rate)}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span>Zero-Knowledge:</span>
                    <span className="font-medium">{formatPercentage(analyticsData.zkid_pqc.security_metrics.zero_knowledge_compliance)}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'email' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Email Types</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.email_usage.emails_by_type).map(([type, count]) => (
                    <div key={type} className="flex justify-between text-sm">
                      <span className="capitalize">{type.replace('_', ' ')}:</span>
                      <span className="font-medium">{count.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Email Status</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.email_usage.emails_by_status).map(([status, count]) => (
                    <div key={status} className="flex justify-between text-sm">
                      <span className="capitalize">{status}:</span>
                      <span className="font-medium">{count.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Top Senders</h4>
                <div className="space-y-2">
                  {analyticsData.email_usage.top_senders.map((sender, index) => (
                    <div key={index} className="flex justify-between text-sm">
                      <span className="truncate">{sender.email}</span>
                      <span className="font-medium">{sender.count.toLocaleString()} ({sender.percentage}%)</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Top Recipients</h4>
                <div className="space-y-2">
                  {analyticsData.email_usage.top_recipients.map((recipient, index) => (
                    <div key={index} className="flex justify-between text-sm">
                      <span className="truncate">{recipient.email}</span>
                      <span className="font-medium">{recipient.count.toLocaleString()} ({recipient.percentage}%)</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'security' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Events by Severity</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.security_events.events_by_severity).map(([severity, count]) => (
                    <div key={severity} className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <div className={`w-3 h-3 rounded-full ${getSeverityColor(severity)}`}></div>
                        <span className="text-sm capitalize">{severity}</span>
                      </div>
                      <span className="font-medium">{count.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Events by Component</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.security_events.events_by_component).map(([component, count]) => (
                    <div key={component} className="flex justify-between text-sm">
                      <span className="capitalize">{component.replace('_', ' ')}:</span>
                      <span className="font-medium">{count.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <h4 className="text-sm font-medium text-gray-700 mb-3">Threat Indicators</h4>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                {Object.entries(analyticsData.security_events.threat_indicators).map(([indicator, count]) => (
                  <div key={indicator} className="text-center">
                    <div className="text-2xl font-bold text-blue-600">{count}</div>
                    <div className="text-xs text-gray-500 capitalize">{indicator.replace('_', ' ')}</div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {activeTab === 'zkid-pqc' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">ZKID Operations</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.zkid_pqc.zkid_operations).map(([operation, count]) => (
                    <div key={operation} className="flex justify-between text-sm">
                      <span className="capitalize">{operation.replace(/_/g, ' ')}:</span>
                      <span className="font-medium">{count.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">PQC Operations</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.zkid_pqc.pqc_operations).map(([operation, count]) => (
                    <div key={operation} className="flex justify-between text-sm">
                      <span className="capitalize">{operation.replace(/_/g, ' ')}:</span>
                      <span className="font-medium">{count.toLocaleString()}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Performance Metrics</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.zkid_pqc.performance_metrics).map(([metric, value]) => (
                    <div key={metric} className="flex justify-between text-sm">
                      <span className="capitalize">{metric.replace(/_/g, ' ')}:</span>
                      <span className="font-medium">{value}ms</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Security Metrics</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.zkid_pqc.security_metrics).map(([metric, value]) => (
                    <div key={metric} className="flex justify-between text-sm">
                      <span className="capitalize">{metric.replace(/_/g, ' ')}:</span>
                      <span className="font-medium">{formatPercentage(value)}</span>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'threats' && (
          <div className="space-y-6">
            <div className="bg-white border border-gray-200 rounded-lg p-4">
              <div className="flex items-center justify-between mb-4">
                <h4 className="text-sm font-medium text-gray-700">Threat Intelligence Overview</h4>
                <span className={`px-2 py-1 rounded-full text-xs font-medium ${getThreatLevelColor(analyticsData.threat_intelligence.threat_level)}`}>
                  {analyticsData.threat_intelligence.threat_level.toUpperCase()}
                </span>
              </div>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="text-center">
                  <div className="text-2xl font-bold text-blue-600">{analyticsData.threat_intelligence.threat_score}</div>
                  <div className="text-xs text-gray-500">Threat Score</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-orange-600">{analyticsData.threat_intelligence.active_threats}</div>
                  <div className="text-xs text-gray-500">Active Threats</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-red-600">{analyticsData.threat_intelligence.emerging_threats}</div>
                  <div className="text-xs text-gray-500">Emerging Threats</div>
                </div>
                <div className="text-center">
                  <div className="text-2xl font-bold text-purple-600">{analyticsData.threat_intelligence.threat_indicators.length}</div>
                  <div className="text-xs text-gray-500">Indicators</div>
                </div>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Threat Categories</h4>
                <div className="space-y-2">
                  {Object.entries(analyticsData.threat_intelligence.threat_categories).map(([category, count]) => (
                    <div key={category} className="flex justify-between text-sm">
                      <span className="capitalize">{category.replace('_', ' ')}:</span>
                      <span className="font-medium">{count}</span>
                    </div>
                  ))}
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-4">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Recommendations</h4>
                <div className="space-y-2">
                  {analyticsData.threat_intelligence.recommendations.slice(0, 3).map((rec) => (
                    <div key={rec.id} className="text-sm">
                      <div className="flex items-center space-x-2">
                        <span className={`px-1 py-0.5 rounded text-xs font-medium ${getThreatLevelColor(rec.priority)}`}>
                          {rec.priority}
                        </span>
                        <span className="font-medium">{rec.title}</span>
                      </div>
                      <p className="text-gray-600 text-xs mt-1">{rec.description}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AnalyticsDashboardPanel;
