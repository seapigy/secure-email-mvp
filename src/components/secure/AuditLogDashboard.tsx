// =============================================================================
// SECURE EMAIL MVP - AUDIT LOG DASHBOARD
// =============================================================================
// React component for viewing and managing audit logs with filtering and export.
// =============================================================================

import React, { useState, useEffect } from 'react';
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableRow, 
  Paper, 
  Typography, 
  Box, 
  Button, 
  TextField, 
  FormControl, 
  InputLabel, 
  Select, 
  MenuItem, 
  Chip, 
  IconButton, 
  Dialog, 
  DialogTitle, 
  DialogContent, 
  DialogActions,
  Alert,
  CircularProgress,
  Pagination,
  Grid,
  Card,
  CardContent,
  Divider,
  Tooltip
} from '@mui/material';
import {
  FilterList,
  Download,
  Refresh,
  Visibility,
  Delete,
  Settings,
  FileDownload,
  History,
  Security,
  Event
} from '@mui/icons-material';
import { DatePicker } from '@mui/x-date-pickers/DatePicker';
import { LocalizationProvider } from '@mui/x-date-pickers/LocalizationProvider';
import { AdapterDateFns } from '@mui/x-date-pickers/AdapterDateFns';

// Types
interface AuditEvent {
  log_id: string;
  timestamp: string;
  event_type: string;
  user_id?: string;
  ip_address?: string;
  user_agent?: string;
  related_email_id?: string;
  outcome: 'success' | 'failure' | 'blocked';
  details?: Record<string, any>;
  severity: 'info' | 'warning' | 'error' | 'critical';
  session_id?: string;
  request_id?: string;
  country?: string;
  city?: string;
  device_type?: string;
  created_at: string;
}

interface AuditLogQuery {
  events: AuditEvent[];
  total: number;
  page: number;
  page_size: number;
  has_more: boolean;
}

interface ExportRequest {
  export_id: string;
  user_id: string;
  export_type: 'csv' | 'json';
  date_from?: string;
  date_to?: string;
  event_types?: string[];
  filters?: any;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  file_path?: string;
  file_size?: number;
  error_message?: string;
  created_at: string;
  completed_at?: string;
  expires_at?: string;
}

interface RetentionPolicy {
  retention_id: string;
  event_type: string;
  retention_days: number;
  auto_purge: boolean;
  created_at: string;
  updated_at: string;
}

// API functions
const api = {
  async getAuditLogs(params: any): Promise<AuditLogQuery> {
    const queryString = new URLSearchParams(params).toString();
    const response = await fetch(`/api/audit/logs?${queryString}`, {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to fetch audit logs');
    return response.json();
  },

  async getEventTypes(): Promise<string[]> {
    const response = await fetch('/api/audit/event-types', {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to fetch event types');
    const data = await response.json();
    return data.event_types;
  },

  async getUserEvents(limit: number = 10): Promise<{ events: AuditEvent[]; total: number }> {
    const response = await fetch(`/api/audit/user-events?limit=${limit}`, {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to fetch user events');
    return response.json();
  },

  async createExport(exportData: any): Promise<ExportRequest> {
    const response = await fetch('/api/audit/exports', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(exportData)
    });
    if (!response.ok) throw new Error('Failed to create export');
    return response.json();
  },

  async getExports(limit: number = 10): Promise<{ exports: ExportRequest[]; total: number }> {
    const response = await fetch(`/api/audit/exports?limit=${limit}`, {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to fetch exports');
    return response.json();
  },

  async getExport(exportId: string): Promise<ExportRequest> {
    const response = await fetch(`/api/audit/exports/${exportId}`, {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to fetch export');
    return response.json();
  },

  async deleteExport(exportId: string): Promise<void> {
    const response = await fetch(`/api/audit/exports/${exportId}`, {
      method: 'DELETE',
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to delete export');
  },

  async getRetentionPolicies(): Promise<RetentionPolicy[]> {
    const response = await fetch('/api/audit/retention-policies', {
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to fetch retention policies');
    const data = await response.json();
    return data.policies;
  },

  async updateRetentionPolicy(eventType: string, policy: any): Promise<void> {
    const response = await fetch(`/api/audit/retention-policies/${eventType}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify(policy)
    });
    if (!response.ok) throw new Error('Failed to update retention policy');
  },

  async purgeExpiredLogs(): Promise<void> {
    const response = await fetch('/api/audit/purge-expired', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to purge expired logs');
  },

  async cleanupExpiredExports(): Promise<void> {
    const response = await fetch('/api/audit/cleanup-exports', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    });
    if (!response.ok) throw new Error('Failed to cleanup expired exports');
  }
};

// Component
const AuditLogDashboard: React.FC = () => {
  // State
  const [auditLogs, setAuditLogs] = useState<AuditLogQuery | null>(null);
  const [userEvents, setUserEvents] = useState<AuditEvent[]>([]);
  const [exports, setExports] = useState<ExportRequest[]>([]);
  const [retentionPolicies, setRetentionPolicies] = useState<RetentionPolicy[]>([]);
  const [eventTypes, setEventTypes] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Filter state
  const [filters, setFilters] = useState({
    dateFrom: null as Date | null,
    dateTo: null as Date | null,
    eventTypes: [] as string[],
    outcomes: [] as string[],
    severities: [] as string[],
    searchTerm: ''
  });

  // Pagination state
  const [pagination, setPagination] = useState({
    page: 1,
    pageSize: 25
  });

  // Dialog state
  const [exportDialogOpen, setExportDialogOpen] = useState(false);
  const [retentionDialogOpen, setRetentionDialogOpen] = useState(false);
  const [selectedEvent, setSelectedEvent] = useState<AuditEvent | null>(null);
  const [eventDetailsOpen, setEventDetailsOpen] = useState(false);

  // Export state
  const [exportData, setExportData] = useState({
    exportType: 'json' as 'csv' | 'json',
    dateFrom: null as Date | null,
    dateTo: null as Date | null,
    eventTypes: [] as string[]
  });

  // Load initial data
  useEffect(() => {
    loadInitialData();
  }, []);

  // Load audit logs when filters or pagination change
  useEffect(() => {
    loadAuditLogs();
  }, [filters, pagination]);

  const loadInitialData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [eventTypesData, userEventsData, exportsData, policiesData] = await Promise.all([
        api.getEventTypes(),
        api.getUserEvents(5),
        api.getExports(5),
        api.getRetentionPolicies()
      ]);

      setEventTypes(eventTypesData);
      setUserEvents(userEventsData.events);
      setExports(exportsData.exports);
      setRetentionPolicies(policiesData);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load data');
    } finally {
      setLoading(false);
    }
  };

  const loadAuditLogs = async () => {
    setLoading(true);
    setError(null);
    try {
      const params: any = {
        page: pagination.page,
        page_size: pagination.pageSize
      };

      if (filters.dateFrom) {
        params.date_from = filters.dateFrom.toISOString();
      }
      if (filters.dateTo) {
        params.date_to = filters.dateTo.toISOString();
      }
      if (filters.eventTypes.length > 0) {
        params.event_types = filters.eventTypes.join(',');
      }
      if (filters.outcomes.length > 0) {
        params.outcomes = filters.outcomes.join(',');
      }
      if (filters.severities.length > 0) {
        params.severities = filters.severities.join(',');
      }
      if (filters.searchTerm) {
        params.search_term = filters.searchTerm;
      }

      const data = await api.getAuditLogs(params);
      setAuditLogs(data);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load audit logs');
    } finally {
      setLoading(false);
    }
  };

  const handleCreateExport = async () => {
    try {
      const exportRequest = await api.createExport({
        export_type: exportData.exportType,
        filter: {
          date_from: exportData.dateFrom?.toISOString(),
          date_to: exportData.dateTo?.toISOString(),
          event_types: exportData.eventTypes
        }
      });

      setExportDialogOpen(false);
      setExportData({
        exportType: 'json',
        dateFrom: null,
        dateTo: null,
        eventTypes: []
      });

      // Refresh exports list
      const exportsData = await api.getExports(5);
      setExports(exportsData.exports);

      // Poll for completion
      pollExportStatus(exportRequest.export_id);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create export');
    }
  };

  const pollExportStatus = async (exportId: string) => {
    const maxAttempts = 30; // 30 seconds
    let attempts = 0;

    const poll = async () => {
      try {
        const exportData = await api.getExport(exportId);
        if (exportData.status === 'completed' || exportData.status === 'failed') {
          // Refresh exports list
          const exportsData = await api.getExports(5);
          setExports(exportsData.exports);
          return;
        }
      } catch (err) {
        console.error('Failed to poll export status:', err);
      }

      attempts++;
      if (attempts < maxAttempts) {
        setTimeout(poll, 1000);
      }
    };

    setTimeout(poll, 1000);
  };

  const handleDownloadExport = (exportId: string) => {
    window.open(`/api/audit/exports/${exportId}/download`, '_blank');
  };

  const handleDeleteExport = async (exportId: string) => {
    try {
      await api.deleteExport(exportId);
      const exportsData = await api.getExports(5);
      setExports(exportsData.exports);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete export');
    }
  };

  const handlePurgeExpiredLogs = async () => {
    try {
      await api.purgeExpiredLogs();
      loadAuditLogs();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to purge expired logs');
    }
  };

  const handleCleanupExports = async () => {
    try {
      await api.cleanupExpiredExports();
      const exportsData = await api.getExports(5);
      setExports(exportsData.exports);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to cleanup exports');
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical': return 'error';
      case 'error': return 'error';
      case 'warning': return 'warning';
      case 'info': return 'info';
      default: return 'default';
    }
  };

  const getOutcomeColor = (outcome: string) => {
    switch (outcome) {
      case 'success': return 'success';
      case 'failure': return 'error';
      case 'blocked': return 'warning';
      default: return 'default';
    }
  };

  const formatTimestamp = (timestamp: string) => {
    return new Date(timestamp).toLocaleString();
  };

  const maskSensitiveData = (data: string) => {
    // Mask IP addresses, user IDs, etc.
    return data.replace(/\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/g, '***.***.***.***');
  };

  if (loading && !auditLogs) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="400px">
        <CircularProgress />
      </Box>
    );
  }

  return (
    <LocalizationProvider dateAdapter={AdapterDateFns}>
      <Box p={3}>
        <Typography variant="h4" gutterBottom>
          <Security sx={{ mr: 1, verticalAlign: 'middle' }} />
          Audit Log Dashboard
        </Typography>

        {error && (
          <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
            {error}
          </Alert>
        )}

        {/* Summary Cards */}
        <Grid container spacing={3} sx={{ mb: 3 }}>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Total Events
                </Typography>
                <Typography variant="h4">
                  {auditLogs?.total || 0}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Recent User Events
                </Typography>
                <Typography variant="h4">
                  {userEvents.length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Active Exports
                </Typography>
                <Typography variant="h4">
                  {exports.filter(e => e.status === 'pending' || e.status === 'processing').length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid item xs={12} sm={6} md={3}>
            <Card>
              <CardContent>
                <Typography color="textSecondary" gutterBottom>
                  Retention Policies
                </Typography>
                <Typography variant="h4">
                  {retentionPolicies.length}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        {/* Filters */}
        <Paper sx={{ p: 2, mb: 3 }}>
          <Box display="flex" alignItems="center" mb={2}>
            <FilterList sx={{ mr: 1 }} />
            <Typography variant="h6">Filters</Typography>
          </Box>
          
          <Grid container spacing={2}>
            <Grid item xs={12} sm={6} md={3}>
              <DatePicker
                label="Date From"
                value={filters.dateFrom}
                onChange={(date) => setFilters(prev => ({ ...prev, dateFrom: date }))}
                renderInput={(params) => <TextField {...params} fullWidth />}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <DatePicker
                label="Date To"
                value={filters.dateTo}
                onChange={(date) => setFilters(prev => ({ ...prev, dateTo: date }))}
                renderInput={(params) => <TextField {...params} fullWidth />}
              />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth>
                <InputLabel>Event Types</InputLabel>
                <Select
                  multiple
                  value={filters.eventTypes}
                  onChange={(e) => setFilters(prev => ({ ...prev, eventTypes: e.target.value as string[] }))}
                  renderValue={(selected) => (
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                      {selected.map((value) => (
                        <Chip key={value} label={value} size="small" />
                      ))}
                    </Box>
                  )}
                >
                  {eventTypes.map((type) => (
                    <MenuItem key={type} value={type}>{type}</MenuItem>
                  ))}
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <FormControl fullWidth>
                <InputLabel>Outcomes</InputLabel>
                <Select
                  multiple
                  value={filters.outcomes}
                  onChange={(e) => setFilters(prev => ({ ...prev, outcomes: e.target.value as string[] }))}
                  renderValue={(selected) => (
                    <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                      {selected.map((value) => (
                        <Chip key={value} label={value} size="small" />
                      ))}
                    </Box>
                  )}
                >
                  <MenuItem value="success">Success</MenuItem>
                  <MenuItem value="failure">Failure</MenuItem>
                  <MenuItem value="blocked">Blocked</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <TextField
                fullWidth
                label="Search"
                value={filters.searchTerm}
                onChange={(e) => setFilters(prev => ({ ...prev, searchTerm: e.target.value }))}
                placeholder="Search in details or user agent..."
              />
            </Grid>
            <Grid item xs={12} sm={6} md={3}>
              <Button
                variant="outlined"
                onClick={() => setFilters({
                  dateFrom: null,
                  dateTo: null,
                  eventTypes: [],
                  outcomes: [],
                  severities: [],
                  searchTerm: ''
                })}
                sx={{ height: 56 }}
              >
                Clear Filters
              </Button>
            </Grid>
          </Grid>
        </Paper>

        {/* Actions */}
        <Box display="flex" gap={2} mb={2}>
          <Button
            variant="contained"
            startIcon={<Download />}
            onClick={() => setExportDialogOpen(true)}
          >
            Export Logs
          </Button>
          <Button
            variant="outlined"
            startIcon={<Refresh />}
            onClick={loadInitialData}
            disabled={loading}
          >
            Refresh
          </Button>
          <Button
            variant="outlined"
            startIcon={<Settings />}
            onClick={() => setRetentionDialogOpen(true)}
          >
            Retention Settings
          </Button>
          <Button
            variant="outlined"
            color="warning"
            onClick={handlePurgeExpiredLogs}
          >
            Purge Expired
          </Button>
          <Button
            variant="outlined"
            color="warning"
            onClick={handleCleanupExports}
          >
            Cleanup Exports
          </Button>
        </Box>

        {/* Audit Logs Table */}
        <Paper sx={{ width: '100%', overflow: 'hidden' }}>
          <Table stickyHeader>
            <TableHead>
              <TableRow>
                <TableCell>Timestamp</TableCell>
                <TableCell>Event Type</TableCell>
                <TableCell>User ID</TableCell>
                <TableCell>IP Address</TableCell>
                <TableCell>Outcome</TableCell>
                <TableCell>Severity</TableCell>
                <TableCell>Related Email</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {auditLogs?.events.map((event) => (
                <TableRow key={event.log_id} hover>
                  <TableCell>{formatTimestamp(event.timestamp)}</TableCell>
                  <TableCell>
                    <Chip 
                      label={event.event_type} 
                      size="small" 
                      icon={<Event />}
                    />
                  </TableCell>
                  <TableCell>
                    {event.user_id ? maskSensitiveData(event.user_id) : '-'}
                  </TableCell>
                  <TableCell>
                    {event.ip_address ? maskSensitiveData(event.ip_address) : '-'}
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={event.outcome} 
                      color={getOutcomeColor(event.outcome) as any}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    <Chip 
                      label={event.severity} 
                      color={getSeverityColor(event.severity) as any}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    {event.related_email_id ? maskSensitiveData(event.related_email_id) : '-'}
                  </TableCell>
                  <TableCell>
                    <Tooltip title="View Details">
                      <IconButton
                        size="small"
                        onClick={() => {
                          setSelectedEvent(event);
                          setEventDetailsOpen(true);
                        }}
                      >
                        <Visibility />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>

        {/* Pagination */}
        {auditLogs && (
          <Box display="flex" justifyContent="center" mt={2}>
            <Pagination
              count={Math.ceil(auditLogs.total / pagination.pageSize)}
              page={pagination.page}
              onChange={(_, page) => setPagination(prev => ({ ...prev, page }))}
              color="primary"
            />
          </Box>
        )}

        {/* Recent User Events */}
        <Paper sx={{ p: 2, mt: 3 }}>
          <Typography variant="h6" gutterBottom>
            <History sx={{ mr: 1, verticalAlign: 'middle' }} />
            Recent User Events
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Time</TableCell>
                <TableCell>Event</TableCell>
                <TableCell>Outcome</TableCell>
                <TableCell>IP Address</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {userEvents.map((event) => (
                <TableRow key={event.log_id}>
                  <TableCell>{formatTimestamp(event.timestamp)}</TableCell>
                  <TableCell>{event.event_type}</TableCell>
                  <TableCell>
                    <Chip 
                      label={event.outcome} 
                      color={getOutcomeColor(event.outcome) as any}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{event.ip_address ? maskSensitiveData(event.ip_address) : '-'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>

        {/* Recent Exports */}
        <Paper sx={{ p: 2, mt: 3 }}>
          <Typography variant="h6" gutterBottom>
            <FileDownload sx={{ mr: 1, verticalAlign: 'middle' }} />
            Recent Exports
          </Typography>
          <Table size="small">
            <TableHead>
              <TableRow>
                <TableCell>Created</TableCell>
                <TableCell>Type</TableCell>
                <TableCell>Status</TableCell>
                <TableCell>File Size</TableCell>
                <TableCell>Actions</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {exports.map((exportItem) => (
                <TableRow key={exportItem.export_id}>
                  <TableCell>{formatTimestamp(exportItem.created_at)}</TableCell>
                  <TableCell>{exportItem.export_type.toUpperCase()}</TableCell>
                  <TableCell>
                    <Chip 
                      label={exportItem.status} 
                      color={exportItem.status === 'completed' ? 'success' : 
                             exportItem.status === 'failed' ? 'error' : 'warning'}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>
                    {exportItem.file_size ? `${(exportItem.file_size / 1024).toFixed(1)} KB` : '-'}
                  </TableCell>
                  <TableCell>
                    {exportItem.status === 'completed' && (
                      <Tooltip title="Download">
                        <IconButton
                          size="small"
                          onClick={() => handleDownloadExport(exportItem.export_id)}
                        >
                          <Download />
                        </IconButton>
                      </Tooltip>
                    )}
                    <Tooltip title="Delete">
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => handleDeleteExport(exportItem.export_id)}
                      >
                        <Delete />
                      </IconButton>
                    </Tooltip>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </Paper>

        {/* Export Dialog */}
        <Dialog open={exportDialogOpen} onClose={() => setExportDialogOpen(false)} maxWidth="sm" fullWidth>
          <DialogTitle>Export Audit Logs</DialogTitle>
          <DialogContent>
            <Grid container spacing={2} sx={{ mt: 1 }}>
              <Grid item xs={12}>
                <FormControl fullWidth>
                  <InputLabel>Export Type</InputLabel>
                  <Select
                    value={exportData.exportType}
                    onChange={(e) => setExportData(prev => ({ ...prev, exportType: e.target.value as 'csv' | 'json' }))}
                  >
                    <MenuItem value="json">JSON</MenuItem>
                    <MenuItem value="csv">CSV</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={12} sm={6}>
                <DatePicker
                  label="Date From"
                  value={exportData.dateFrom}
                  onChange={(date) => setExportData(prev => ({ ...prev, dateFrom: date }))}
                  renderInput={(params) => <TextField {...params} fullWidth />}
                />
              </Grid>
              <Grid item xs={12} sm={6}>
                <DatePicker
                  label="Date To"
                  value={exportData.dateTo}
                  onChange={(date) => setExportData(prev => ({ ...prev, dateTo: date }))}
                  renderInput={(params) => <TextField {...params} fullWidth />}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth>
                  <InputLabel>Event Types</InputLabel>
                  <Select
                    multiple
                    value={exportData.eventTypes}
                    onChange={(e) => setExportData(prev => ({ ...prev, eventTypes: e.target.value as string[] }))}
                    renderValue={(selected) => (
                      <Box sx={{ display: 'flex', flexWrap: 'wrap', gap: 0.5 }}>
                        {selected.map((value) => (
                          <Chip key={value} label={value} size="small" />
                        ))}
                      </Box>
                    )}
                  >
                    {eventTypes.map((type) => (
                      <MenuItem key={type} value={type}>{type}</MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setExportDialogOpen(false)}>Cancel</Button>
            <Button onClick={handleCreateExport} variant="contained">Create Export</Button>
          </DialogActions>
        </Dialog>

        {/* Event Details Dialog */}
        <Dialog open={eventDetailsOpen} onClose={() => setEventDetailsOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>Event Details</DialogTitle>
          <DialogContent>
            {selectedEvent && (
              <Box>
                <Typography variant="h6" gutterBottom>Event Information</Typography>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Log ID</Typography>
                    <Typography variant="body1">{selectedEvent.log_id}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Timestamp</Typography>
                    <Typography variant="body1">{formatTimestamp(selectedEvent.timestamp)}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Event Type</Typography>
                    <Typography variant="body1">{selectedEvent.event_type}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Outcome</Typography>
                    <Chip label={selectedEvent.outcome} color={getOutcomeColor(selectedEvent.outcome) as any} />
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Severity</Typography>
                    <Chip label={selectedEvent.severity} color={getSeverityColor(selectedEvent.severity) as any} />
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">User ID</Typography>
                    <Typography variant="body1">{selectedEvent.user_id ? maskSensitiveData(selectedEvent.user_id) : '-'}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">IP Address</Typography>
                    <Typography variant="body1">{selectedEvent.ip_address ? maskSensitiveData(selectedEvent.ip_address) : '-'}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Country</Typography>
                    <Typography variant="body1">{selectedEvent.country || '-'}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">City</Typography>
                    <Typography variant="body1">{selectedEvent.city || '-'}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Device Type</Typography>
                    <Typography variant="body1">{selectedEvent.device_type || '-'}</Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="body2" color="textSecondary">Session ID</Typography>
                    <Typography variant="body1">{selectedEvent.session_id ? maskSensitiveData(selectedEvent.session_id) : '-'}</Typography>
                  </Grid>
                  <Grid item xs={12}>
                    <Typography variant="body2" color="textSecondary">User Agent</Typography>
                    <Typography variant="body1" sx={{ wordBreak: 'break-all' }}>
                      {selectedEvent.user_agent ? maskSensitiveData(selectedEvent.user_agent) : '-'}
                    </Typography>
                  </Grid>
                  {selectedEvent.details && Object.keys(selectedEvent.details).length > 0 && (
                    <Grid item xs={12}>
                      <Typography variant="body2" color="textSecondary">Details</Typography>
                      <Paper sx={{ p: 1, bgcolor: 'grey.50' }}>
                        <pre style={{ margin: 0, fontSize: '0.875rem' }}>
                          {JSON.stringify(selectedEvent.details, null, 2)}
                        </pre>
                      </Paper>
                    </Grid>
                  )}
                </Grid>
              </Box>
            )}
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setEventDetailsOpen(false)}>Close</Button>
          </DialogActions>
        </Dialog>

        {/* Retention Policies Dialog */}
        <Dialog open={retentionDialogOpen} onClose={() => setRetentionDialogOpen(false)} maxWidth="md" fullWidth>
          <DialogTitle>Retention Policies</DialogTitle>
          <DialogContent>
            <Typography variant="body2" color="textSecondary" sx={{ mb: 2 }}>
              Configure how long different types of audit events are retained before automatic deletion.
            </Typography>
            <Table size="small">
              <TableHead>
                <TableRow>
                  <TableCell>Event Type</TableCell>
                  <TableCell>Retention Days</TableCell>
                  <TableCell>Auto Purge</TableCell>
                  <TableCell>Last Updated</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {retentionPolicies.map((policy) => (
                  <TableRow key={policy.retention_id}>
                    <TableCell>{policy.event_type}</TableCell>
                    <TableCell>{policy.retention_days}</TableCell>
                    <TableCell>
                      <Chip 
                        label={policy.auto_purge ? 'Yes' : 'No'} 
                        color={policy.auto_purge ? 'success' : 'default'}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>{formatTimestamp(policy.updated_at)}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </DialogContent>
          <DialogActions>
            <Button onClick={() => setRetentionDialogOpen(false)}>Close</Button>
          </DialogActions>
        </Dialog>
      </Box>
    </LocalizationProvider>
  );
};

export default AuditLogDashboard;

