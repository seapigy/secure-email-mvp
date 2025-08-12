import React, { useState, useEffect } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  Chip,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Switch,
  FormControlLabel,
  Alert,
  IconButton,
  Tooltip,
  Grid,
  Divider,
  List,
  ListItem,
  ListItemText,
  ListItemIcon,
} from '@mui/material';
import {
  Security,
  Warning,
  CheckCircle,
  Cancel,
  Visibility,
  Settings,
  Refresh,
  Flag,
  LocationOn,
  Computer,
  AccessTime,
  Person,
} from '@mui/icons-material';
import { api } from '../../lib/api';

interface SuspiciousAccessEvent {
  detection_id: string;
  email_id: string;
  detection_type: string;
  detection_rule: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  triggered_at: string;
  resolved_at?: string;
  resolved_by?: string;
  resolution_notes?: string;
  detection_metadata?: Record<string, any>;
}

interface SuspiciousActivityResponse {
  email_id: string;
  suspicious_flag: boolean;
  suspicious_flag_set_at?: string;
  detection_events: SuspiciousAccessEvent[];
  total_detections: number;
  unresolved_detections: number;
}

interface UserPreferences {
  user_id: string;
  enable_suspicious_detection: boolean;
  notify_on_suspicious_activity: boolean;
  auto_flag_suspicious_emails: boolean;
  minimum_severity_for_notification: 'low' | 'medium' | 'high' | 'critical';
  created_at: string;
  updated_at: string;
}

interface DetectionRule {
  rule_id: string;
  rule_name: string;
  rule_type: string;
  is_enabled: boolean;
  threshold_value: number;
  time_window_minutes: number;
  severity: 'low' | 'medium' | 'high' | 'critical';
  description: string;
  created_at: string;
  updated_at: string;
}

interface SuspiciousEmail {
  email_id: string;
  subject: string;
  suspicious_flag: boolean;
  suspicious_flag_set_at?: string;
  created_at: string;
}

const SuspiciousAccessDashboard: React.FC = () => {
  const [suspiciousEmails, setSuspiciousEmails] = useState<SuspiciousEmail[]>([]);
  const [selectedEmail, setSelectedEmail] = useState<string | null>(null);
  const [suspiciousActivity, setSuspiciousActivity] = useState<SuspiciousActivityResponse | null>(null);
  const [userPreferences, setUserPreferences] = useState<UserPreferences | null>(null);
  const [detectionRules, setDetectionRules] = useState<DetectionRule[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  // Dialog states
  const [activityDialogOpen, setActivityDialogOpen] = useState(false);
  const [preferencesDialogOpen, setPreferencesDialogOpen] = useState(false);
  const [rulesDialogOpen, setRulesDialogOpen] = useState(false);
  const [clearFlagDialogOpen, setClearFlagDialogOpen] = useState(false);
  const [resolveDialogOpen, setResolveDialogOpen] = useState(false);
  const [selectedDetection, setSelectedDetection] = useState<SuspiciousAccessEvent | null>(null);
  const [resolutionNotes, setResolutionNotes] = useState('');

  // Load suspicious emails
  const loadSuspiciousEmails = async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get('/api/suspicious/emails');
      setSuspiciousEmails(response.data.suspicious_emails || []);
    } catch (err) {
      setError('Failed to load suspicious emails');
      console.error('Error loading suspicious emails:', err);
    } finally {
      setLoading(false);
    }
  };

  // Load user preferences
  const loadUserPreferences = async () => {
    try {
      const response = await api.get('/api/suspicious/preferences');
      setUserPreferences(response.data);
    } catch (err) {
      console.error('Error loading user preferences:', err);
    }
  };

  // Load detection rules
  const loadDetectionRules = async () => {
    try {
      const response = await api.get('/api/suspicious/rules');
      setDetectionRules(response.data);
    } catch (err) {
      console.error('Error loading detection rules:', err);
    }
  };

  // Load suspicious activity for a specific email
  const loadSuspiciousActivity = async (emailId: string) => {
    setLoading(true);
    setError(null);
    try {
      const response = await api.get(`/api/suspicious/activity/${emailId}`);
      setSuspiciousActivity(response.data);
      setActivityDialogOpen(true);
    } catch (err) {
      setError('Failed to load suspicious activity');
      console.error('Error loading suspicious activity:', err);
    } finally {
      setLoading(false);
    }
  };

  // Update user preferences
  const updateUserPreferences = async (preferences: Partial<UserPreferences>) => {
    try {
      await api.put('/api/suspicious/preferences', preferences);
      setSuccess('User preferences updated successfully');
      await loadUserPreferences();
      setPreferencesDialogOpen(false);
    } catch (err) {
      setError('Failed to update user preferences');
      console.error('Error updating user preferences:', err);
    }
  };

  // Clear suspicious flag
  const clearSuspiciousFlag = async (emailId: string, notes?: string) => {
    try {
      await api.post(`/api/suspicious/clear-flag/${emailId}`, {
        resolution_notes: notes || 'Flag cleared by user'
      });
      setSuccess('Suspicious flag cleared successfully');
      await loadSuspiciousEmails();
      setClearFlagDialogOpen(false);
      setResolutionNotes('');
    } catch (err) {
      setError('Failed to clear suspicious flag');
      console.error('Error clearing suspicious flag:', err);
    }
  };

  // Resolve detection event
  const resolveDetectionEvent = async (detectionId: string, notes: string) => {
    try {
      await api.post(`/api/suspicious/resolve/${detectionId}`, {
        resolution_notes: notes
      });
      setSuccess('Detection event resolved successfully');
      if (selectedEmail) {
        await loadSuspiciousActivity(selectedEmail);
      }
      setResolveDialogOpen(false);
      setSelectedDetection(null);
      setResolutionNotes('');
    } catch (err) {
      setError('Failed to resolve detection event');
      console.error('Error resolving detection event:', err);
    }
  };

  // Get severity color
  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case 'critical':
        return 'error';
      case 'high':
        return 'warning';
      case 'medium':
        return 'info';
      case 'low':
        return 'success';
      default:
        return 'default';
    }
  };

  // Get detection type icon
  const getDetectionTypeIcon = (type: string) => {
    switch (type) {
      case 'multiple_failed_attempts':
        return <Warning />;
      case 'unusual_geolocation':
        return <LocationOn />;
      case 'rapid_multiple_ips':
        return <Computer />;
      case 'impossible_travel':
        return <AccessTime />;
      default:
        return <Security />;
    }
  };

  // Format detection metadata
  const formatDetectionMetadata = (metadata: Record<string, any>) => {
    if (!metadata) return null;

    const items = [];
    for (const [key, value] of Object.entries(metadata)) {
      if (Array.isArray(value)) {
        items.push(
          <ListItem key={key}>
            <ListItemIcon>{getDetectionTypeIcon(key)}</ListItemIcon>
            <ListItemText
              primary={key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
              secondary={`${value.length} items`}
            />
          </ListItem>
        );
      } else {
        items.push(
          <ListItem key={key}>
            <ListItemIcon>{getDetectionTypeIcon(key)}</ListItemIcon>
            <ListItemText
              primary={key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
              secondary={String(value)}
            />
          </ListItem>
        );
      }
    }
    return items;
  };

  useEffect(() => {
    loadSuspiciousEmails();
    loadUserPreferences();
    loadDetectionRules();
  }, []);

  return (
    <Box sx={{ p: 3 }}>
      <Typography variant="h4" gutterBottom>
        <Security sx={{ mr: 1, verticalAlign: 'middle' }} />
        Suspicious Access Detection
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      {success && (
        <Alert severity="success" sx={{ mb: 2 }} onClose={() => setSuccess(null)}>
          {success}
        </Alert>
      )}

      <Grid container spacing={3} sx={{ mb: 3 }}>
        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                <Flag sx={{ mr: 1, verticalAlign: 'middle' }} />
                Suspicious Emails
              </Typography>
              <Typography variant="h4" color="error">
                {suspiciousEmails.length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Emails flagged for suspicious activity
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                <Warning sx={{ mr: 1, verticalAlign: 'middle' }} />
                Detection Rules
              </Typography>
              <Typography variant="h4" color="info.main">
                {detectionRules.filter(rule => rule.is_enabled).length}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Active detection rules
              </Typography>
            </CardContent>
          </Card>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardContent>
              <Typography variant="h6" gutterBottom>
                <Settings sx={{ mr: 1, verticalAlign: 'middle' }} />
                Detection Status
              </Typography>
              <Typography variant="h4" color={userPreferences?.enable_suspicious_detection ? 'success.main' : 'error'}>
                {userPreferences?.enable_suspicious_detection ? 'Active' : 'Disabled'}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Suspicious detection system
              </Typography>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box sx={{ mb: 2, display: 'flex', gap: 1 }}>
        <Button
          variant="outlined"
          startIcon={<Refresh />}
          onClick={loadSuspiciousEmails}
          disabled={loading}
        >
          Refresh
        </Button>
        <Button
          variant="outlined"
          startIcon={<Settings />}
          onClick={() => setPreferencesDialogOpen(true)}
        >
          Preferences
        </Button>
        <Button
          variant="outlined"
          startIcon={<Security />}
          onClick={() => setRulesDialogOpen(true)}
        >
          Detection Rules
        </Button>
      </Box>

      <Card>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Suspicious Emails
          </Typography>
          
          {suspiciousEmails.length === 0 ? (
            <Typography variant="body2" color="text.secondary" sx={{ textAlign: 'center', py: 4 }}>
              No suspicious emails found
            </Typography>
          ) : (
            <TableContainer component={Paper}>
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Subject</TableCell>
                    <TableCell>Flagged At</TableCell>
                    <TableCell>Created</TableCell>
                    <TableCell>Actions</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {suspiciousEmails.map((email) => (
                    <TableRow key={email.email_id}>
                      <TableCell>{email.subject}</TableCell>
                      <TableCell>
                        {email.suspicious_flag_set_at ? new Date(email.suspicious_flag_set_at).toLocaleString() : 'N/A'}
                      </TableCell>
                      <TableCell>
                        {new Date(email.created_at).toLocaleString()}
                      </TableCell>
                      <TableCell>
                        <Tooltip title="View Suspicious Activity">
                          <IconButton
                            size="small"
                            onClick={() => loadSuspiciousActivity(email.email_id)}
                          >
                            <Visibility />
                          </IconButton>
                        </Tooltip>
                        <Tooltip title="Clear Suspicious Flag">
                          <IconButton
                            size="small"
                            color="success"
                            onClick={() => {
                              setSelectedEmail(email.email_id);
                              setClearFlagDialogOpen(true);
                            }}
                          >
                            <CheckCircle />
                          </IconButton>
                        </Tooltip>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          )}
        </CardContent>
      </Card>

      {/* Suspicious Activity Dialog */}
      <Dialog
        open={activityDialogOpen}
        onClose={() => setActivityDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          <Security sx={{ mr: 1, verticalAlign: 'middle' }} />
          Suspicious Activity Details
        </DialogTitle>
        <DialogContent>
          {suspiciousActivity && (
            <Box>
              <Typography variant="h6" gutterBottom>
                Email ID: {suspiciousActivity.email_id}
              </Typography>
              
              <Box sx={{ mb: 2 }}>
                <Chip
                  label={suspiciousActivity.suspicious_flag ? 'Flagged as Suspicious' : 'Not Flagged'}
                  color={suspiciousActivity.suspicious_flag ? 'error' : 'success'}
                  icon={suspiciousActivity.suspicious_flag ? <Warning /> : <CheckCircle />}
                />
              </Box>

              <Typography variant="h6" gutterBottom>
                Detection Events ({suspiciousActivity.total_detections})
              </Typography>

              {suspiciousActivity.detection_events.length === 0 ? (
                <Typography variant="body2" color="text.secondary">
                  No detection events found
                </Typography>
              ) : (
                <TableContainer component={Paper}>
                  <Table size="small">
                    <TableHead>
                      <TableRow>
                        <TableCell>Type</TableCell>
                        <TableCell>Severity</TableCell>
                        <TableCell>Triggered</TableCell>
                        <TableCell>Status</TableCell>
                        <TableCell>Actions</TableCell>
                      </TableRow>
                    </TableHead>
                    <TableBody>
                      {suspiciousActivity.detection_events.map((event) => (
                        <TableRow key={event.detection_id}>
                          <TableCell>
                            <Box sx={{ display: 'flex', alignItems: 'center' }}>
                              {getDetectionTypeIcon(event.detection_type)}
                              <Typography variant="body2" sx={{ ml: 1 }}>
                                {event.detection_type.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase())}
                              </Typography>
                            </Box>
                          </TableCell>
                          <TableCell>
                            <Chip
                              label={event.severity}
                              color={getSeverityColor(event.severity) as any}
                              size="small"
                            />
                          </TableCell>
                          <TableCell>
                            {new Date(event.triggered_at).toLocaleString()}
                          </TableCell>
                          <TableCell>
                            {event.resolved_at ? (
                              <Chip label="Resolved" color="success" size="small" icon={<CheckCircle />} />
                            ) : (
                              <Chip label="Active" color="warning" size="small" icon={<Warning />} />
                            )}
                          </TableCell>
                          <TableCell>
                            {!event.resolved_at && (
                              <Tooltip title="Resolve Detection">
                                <IconButton
                                  size="small"
                                  onClick={() => {
                                    setSelectedDetection(event);
                                    setResolveDialogOpen(true);
                                  }}
                                >
                                  <CheckCircle />
                                </IconButton>
                              </Tooltip>
                            )}
                          </TableCell>
                        </TableRow>
                      ))}
                    </TableBody>
                  </Table>
                </TableContainer>
              )}

              {selectedDetection && selectedDetection.detection_metadata && (
                <Box sx={{ mt: 2 }}>
                  <Typography variant="h6" gutterBottom>
                    Detection Metadata
                  </Typography>
                  <List dense>
                    {formatDetectionMetadata(selectedDetection.detection_metadata)}
                  </List>
                </Box>
              )}
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setActivityDialogOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* User Preferences Dialog */}
      <Dialog
        open={preferencesDialogOpen}
        onClose={() => setPreferencesDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>User Preferences</DialogTitle>
        <DialogContent>
          {userPreferences && (
            <Box sx={{ pt: 1 }}>
              <FormControlLabel
                control={
                  <Switch
                    checked={userPreferences.enable_suspicious_detection}
                    onChange={(e) => setUserPreferences({
                      ...userPreferences,
                      enable_suspicious_detection: e.target.checked
                    })}
                  />
                }
                label="Enable Suspicious Detection"
              />
              
              <FormControlLabel
                control={
                  <Switch
                    checked={userPreferences.notify_on_suspicious_activity}
                    onChange={(e) => setUserPreferences({
                      ...userPreferences,
                      notify_on_suspicious_activity: e.target.checked
                    })}
                  />
                }
                label="Notify on Suspicious Activity"
              />
              
              <FormControlLabel
                control={
                  <Switch
                    checked={userPreferences.auto_flag_suspicious_emails}
                    onChange={(e) => setUserPreferences({
                      ...userPreferences,
                      auto_flag_suspicious_emails: e.target.checked
                    })}
                  />
                }
                label="Auto-Flag Suspicious Emails"
              />
              
              <FormControl fullWidth sx={{ mt: 2 }}>
                <InputLabel>Minimum Severity for Notification</InputLabel>
                <Select
                  value={userPreferences.minimum_severity_for_notification}
                  onChange={(e) => setUserPreferences({
                    ...userPreferences,
                    minimum_severity_for_notification: e.target.value as any
                  })}
                  label="Minimum Severity for Notification"
                >
                  <MenuItem value="low">Low</MenuItem>
                  <MenuItem value="medium">Medium</MenuItem>
                  <MenuItem value="high">High</MenuItem>
                  <MenuItem value="critical">Critical</MenuItem>
                </Select>
              </FormControl>
            </Box>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setPreferencesDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={() => userPreferences && updateUserPreferences(userPreferences)}
            variant="contained"
          >
            Save Preferences
          </Button>
        </DialogActions>
      </Dialog>

      {/* Detection Rules Dialog */}
      <Dialog
        open={rulesDialogOpen}
        onClose={() => setRulesDialogOpen(false)}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>Detection Rules</DialogTitle>
        <DialogContent>
          <TableContainer component={Paper}>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Rule Name</TableCell>
                  <TableCell>Type</TableCell>
                  <TableCell>Threshold</TableCell>
                  <TableCell>Time Window</TableCell>
                  <TableCell>Severity</TableCell>
                  <TableCell>Status</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {detectionRules.map((rule) => (
                  <TableRow key={rule.rule_id}>
                    <TableCell>{rule.rule_name}</TableCell>
                    <TableCell>{rule.rule_type.replace(/_/g, ' ')}</TableCell>
                    <TableCell>{rule.threshold_value}</TableCell>
                    <TableCell>{rule.time_window_minutes} min</TableCell>
                    <TableCell>
                      <Chip
                        label={rule.severity}
                        color={getSeverityColor(rule.severity) as any}
                        size="small"
                      />
                    </TableCell>
                    <TableCell>
                      <Chip
                        label={rule.is_enabled ? 'Enabled' : 'Disabled'}
                        color={rule.is_enabled ? 'success' : 'default'}
                        size="small"
                      />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </TableContainer>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRulesDialogOpen(false)}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Clear Flag Dialog */}
      <Dialog
        open={clearFlagDialogOpen}
        onClose={() => setClearFlagDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Clear Suspicious Flag</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            multiline
            rows={3}
            label="Resolution Notes (Optional)"
            value={resolutionNotes}
            onChange={(e) => setResolutionNotes(e.target.value)}
            sx={{ mt: 1 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setClearFlagDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={() => selectedEmail && clearSuspiciousFlag(selectedEmail, resolutionNotes)}
            variant="contained"
            color="success"
          >
            Clear Flag
          </Button>
        </DialogActions>
      </Dialog>

      {/* Resolve Detection Dialog */}
      <Dialog
        open={resolveDialogOpen}
        onClose={() => setResolveDialogOpen(false)}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>Resolve Detection Event</DialogTitle>
        <DialogContent>
          <TextField
            fullWidth
            multiline
            rows={3}
            label="Resolution Notes"
            value={resolutionNotes}
            onChange={(e) => setResolutionNotes(e.target.value)}
            required
            sx={{ mt: 1 }}
          />
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setResolveDialogOpen(false)}>Cancel</Button>
          <Button
            onClick={() => selectedDetection && resolveDetectionEvent(selectedDetection.detection_id, resolutionNotes)}
            variant="contained"
            color="success"
            disabled={!resolutionNotes.trim()}
          >
            Resolve
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default SuspiciousAccessDashboard;

