// Dashboard Screen for SecureMail
import React, { useState, useEffect } from 'react';
import {
  View,
  Text,
  StyleSheet,
  ScrollView,
  TouchableOpacity,
  Alert,
} from 'react-native';
import { useAuth } from '../contexts/AuthContext';
import { apiService } from '../services/api';
import { TrialWarning } from '../types';

export default function DashboardScreen() {
  const { state, logout } = useAuth();
  const [trialWarning, setTrialWarning] = useState<TrialWarning | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    if (state.isAuthenticated && state.token) {
      loadTrialWarning();
    }
  }, [state.isAuthenticated, state.token]);

  const loadTrialWarning = async () => {
    if (!state.token) return;
    
    try {
      const response = await apiService.getTrialWarning(state.token);
      if (response.has_warning && response.warning) {
        setTrialWarning(response.warning);
      }
    } catch (error) {
      console.error('Error loading trial warning:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleExtendTrial = async () => {
    if (!state.token) return;
    
    try {
      const response = await apiService.extendTrial(state.token);
      Alert.alert('Success', response.message);
      loadTrialWarning(); // Reload warning
    } catch (error: any) {
      Alert.alert('Error', error.message || 'Failed to extend trial');
    }
  };

  const handleLogout = () => {
    Alert.alert(
      'Sign Out',
      'Are you sure you want to sign out?',
      [
        { text: 'Cancel', style: 'cancel' },
        { text: 'Sign Out', style: 'destructive', onPress: logout },
      ]
    );
  };

  const getTrialWarningStyle = (level: string) => {
    switch (level) {
      case 'critical':
        return styles.trialWarningCritical;
      case 'warning':
        return styles.trialWarningWarning;
      default:
        return styles.trialWarningInfo;
    }
  };

  const getTrialWarningIcon = (level: string) => {
    switch (level) {
      case 'critical':
        return '⚠️';
      case 'warning':
        return '⚠️';
      default:
        return 'ℹ️';
    }
  };

  return (
    <ScrollView style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>Dashboard</Text>
        <TouchableOpacity onPress={handleLogout} style={styles.logoutButton}>
          <Text style={styles.logoutText}>Sign Out</Text>
        </TouchableOpacity>
      </View>

      {state.user && (
        <View style={styles.userInfo}>
          <Text style={styles.welcomeText}>
            Welcome back, {state.user.username}!
          </Text>
          <Text style={styles.accountTypeText}>
            Account: {state.user.accountType.charAt(0).toUpperCase() + state.user.accountType.slice(1)}
          </Text>
          {state.user.organizationId && (
            <Text style={styles.organizationText}>
              Organization: {state.user.organizationRole}
            </Text>
          )}
        </View>
      )}

      {trialWarning && (
        <View style={[styles.trialWarning, getTrialWarningStyle(trialWarning.warningLevel)]}>
          <Text style={styles.trialWarningIcon}>
            {getTrialWarningIcon(trialWarning.warningLevel)}
          </Text>
          <View style={styles.trialWarningContent}>
            <Text style={styles.trialWarningTitle}>
              Trial Expires in {trialWarning.daysRemaining} Days
            </Text>
            <Text style={styles.trialWarningText}>
              Your {trialWarning.plan} trial expires on {new Date(trialWarning.expiryDate).toLocaleDateString()}.
            </Text>
            <TouchableOpacity 
              style={styles.extendTrialButton}
              onPress={handleExtendTrial}
            >
              <Text style={styles.extendTrialText}>Extend Trial</Text>
            </TouchableOpacity>
          </View>
        </View>
      )}

      <View style={styles.features}>
        <Text style={styles.featuresTitle}>Features</Text>
        
        <View style={styles.featureCard}>
          <Text style={styles.featureTitle}>📧 Secure Email</Text>
          <Text style={styles.featureDescription}>
            Send and receive encrypted emails with end-to-end security.
          </Text>
        </View>

        <View style={styles.featureCard}>
          <Text style={styles.featureTitle}>🔐 Zero-Knowledge</Text>
          <Text style={styles.featureDescription}>
            We can't read your emails. Your privacy is protected.
          </Text>
        </View>

        <View style={styles.featureCard}>
          <Text style={styles.featureTitle}>🛡️ Quantum-Resistant</Text>
          <Text style={styles.featureDescription}>
            Advanced cryptography that's future-proof against quantum computers.
          </Text>
        </View>

        {state.user?.accountType !== 'free' && (
          <View style={styles.featureCard}>
            <Text style={styles.featureTitle}>🏢 Organization Management</Text>
            <Text style={styles.featureDescription}>
              Manage team members and organization settings.
            </Text>
          </View>
        )}
      </View>
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#F5F5F5',
  },
  header: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: 20,
    backgroundColor: '#FFFFFF',
    borderBottomWidth: 1,
    borderBottomColor: '#E0E0E0',
  },
  title: {
    fontSize: 24,
    fontWeight: 'bold',
    color: '#333333',
  },
  logoutButton: {
    padding: 8,
  },
  logoutText: {
    color: '#007AFF',
    fontSize: 16,
  },
  userInfo: {
    backgroundColor: '#FFFFFF',
    padding: 20,
    margin: 20,
    borderRadius: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  welcomeText: {
    fontSize: 20,
    fontWeight: '600',
    color: '#333333',
    marginBottom: 8,
  },
  accountTypeText: {
    fontSize: 16,
    color: '#666666',
    marginBottom: 4,
  },
  organizationText: {
    fontSize: 16,
    color: '#666666',
  },
  trialWarning: {
    flexDirection: 'row',
    margin: 20,
    padding: 16,
    borderRadius: 12,
    alignItems: 'flex-start',
  },
  trialWarningInfo: {
    backgroundColor: '#E3F2FD',
    borderColor: '#2196F3',
  },
  trialWarningWarning: {
    backgroundColor: '#FFF3E0',
    borderColor: '#FF9800',
  },
  trialWarningCritical: {
    backgroundColor: '#FFEBEE',
    borderColor: '#F44336',
  },
  trialWarningIcon: {
    fontSize: 24,
    marginRight: 12,
  },
  trialWarningContent: {
    flex: 1,
  },
  trialWarningTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333333',
    marginBottom: 4,
  },
  trialWarningText: {
    fontSize: 14,
    color: '#666666',
    marginBottom: 12,
  },
  extendTrialButton: {
    backgroundColor: '#007AFF',
    paddingHorizontal: 16,
    paddingVertical: 8,
    borderRadius: 6,
    alignSelf: 'flex-start',
  },
  extendTrialText: {
    color: '#FFFFFF',
    fontSize: 14,
    fontWeight: '600',
  },
  features: {
    padding: 20,
  },
  featuresTitle: {
    fontSize: 20,
    fontWeight: '600',
    color: '#333333',
    marginBottom: 16,
  },
  featureCard: {
    backgroundColor: '#FFFFFF',
    padding: 16,
    borderRadius: 12,
    marginBottom: 12,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.1,
    shadowRadius: 4,
    elevation: 3,
  },
  featureTitle: {
    fontSize: 16,
    fontWeight: '600',
    color: '#333333',
    marginBottom: 8,
  },
  featureDescription: {
    fontSize: 14,
    color: '#666666',
    lineHeight: 20,
  },
});
