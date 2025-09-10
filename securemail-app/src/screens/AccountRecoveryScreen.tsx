import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
  Alert,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { WebsiteButton, WebsiteCard, WebsiteInput, WebsiteText } from '../components/WebsiteStyledComponents';
import { colors, spacing } from '../theme';
import { apiService } from '../services/api';

interface RecoveryFormData {
  fallback_email: string;
  recovery_key: string;
  new_password: string;
  confirm_password: string;
  action: 'reset_password' | 'reset_email';
  new_email: string;
}

export default function AccountRecoveryScreen() {
  const navigation = useNavigation();
  const [formData, setFormData] = useState<RecoveryFormData>({
    fallback_email: '',
    recovery_key: '',
    new_password: '',
    confirm_password: '',
    action: 'reset_password',
    new_email: '',
  });
  const [errors, setErrors] = useState<Partial<RecoveryFormData>>({});
  const [isLoading, setIsLoading] = useState(false);

  const handleInputChange = (field: keyof RecoveryFormData, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    // Clear error when user starts typing
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Partial<RecoveryFormData> = {};

    if (!formData.fallback_email) {
      newErrors.fallback_email = 'Fallback email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.fallback_email)) {
      newErrors.fallback_email = 'Please enter a valid email address';
    }

    if (!formData.recovery_key) {
      newErrors.recovery_key = 'Recovery key is required';
    }

    if (formData.action === 'reset_password') {
      if (!formData.new_password) {
        newErrors.new_password = 'New password is required';
      } else if (formData.new_password.length < 8) {
        newErrors.new_password = 'Password must be at least 8 characters';
      }

      if (formData.new_password !== formData.confirm_password) {
        newErrors.confirm_password = 'Passwords do not match';
      }
    } else if (formData.action === 'reset_email') {
      if (!formData.new_email) {
        newErrors.new_email = 'New email is required';
      } else if (!/\S+@\S+\.\S+/.test(formData.new_email)) {
        newErrors.new_email = 'Please enter a valid email address';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleRecovery = async () => {
    if (!validateForm()) return;

    setIsLoading(true);
    try {
      const recoveryData = {
        fallback_email: formData.fallback_email,
        recovery_key: formData.recovery_key,
        action: formData.action,
        ...(formData.action === 'reset_password' 
          ? { new_password: formData.new_password }
          : { new_email: formData.new_email }
        ),
      };

      await apiService.recoverAccount(recoveryData);
      
      Alert.alert(
        'Recovery Successful',
        formData.action === 'reset_password' 
          ? 'Your password has been reset successfully. You can now log in with your new password.'
          : 'Your email has been reset successfully. Please check your new email for verification.',
        [
          {
            text: 'OK',
            onPress: () => navigation.navigate('Login' as never),
          },
        ]
      );
    } catch (error: any) {
      console.error('Recovery error:', error);
      Alert.alert('Recovery Failed', error.message || 'Account recovery failed. Please check your fallback email and recovery key.');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <KeyboardAvoidingView 
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView 
        style={styles.scrollView}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        <View style={styles.header}>
          <WebsiteText variant="h2" color="inverse" weight="bold" style={styles.title}>
            🔐 Account Recovery
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.subtitle}>
            Recover your account using your fallback email and recovery key
          </WebsiteText>
        </View>

        <WebsiteCard variant="glass" style={styles.formCard}>
          <WebsiteText variant="h4" color="inverse" weight="semibold" style={styles.formTitle}>
            Recovery Information
          </WebsiteText>
          
          <WebsiteInput
            label="Fallback Email Address"
            placeholder="Enter your fallback email"
            value={formData.fallback_email}
            onChangeText={(value) => handleInputChange('fallback_email', value)}
            keyboardType="email-address"
            autoCapitalize="none"
            autoCorrect={false}
            error={errors.fallback_email}
            helperText="The external email address (Gmail, Outlook, etc.) you provided during signup"
          />

          <WebsiteInput
            label="Recovery Key"
            placeholder="Enter your recovery key"
            value={formData.recovery_key}
            onChangeText={(value) => handleInputChange('recovery_key', value)}
            autoCapitalize="none"
            autoCorrect={false}
            error={errors.recovery_key}
            helperText="The recovery key you received during signup"
          />

          <View style={styles.actionSection}>
            <WebsiteText variant="h5" color="accent" weight="semibold" style={styles.sectionTitle}>
              Recovery Action
            </WebsiteText>
            
            <View style={styles.actionButtons}>
              <WebsiteButton
                variant={formData.action === 'reset_password' ? 'primary' : 'outline'}
                size="sm"
                onPress={() => handleInputChange('action', 'reset_password')}
                style={styles.actionButton}
              >
                Reset Password
              </WebsiteButton>
              <WebsiteButton
                variant={formData.action === 'reset_email' ? 'primary' : 'outline'}
                size="sm"
                onPress={() => handleInputChange('action', 'reset_email')}
                style={styles.actionButton}
              >
                Reset Email
              </WebsiteButton>
            </View>
          </View>

          {formData.action === 'reset_password' && (
            <>
              <WebsiteInput
                label="New Password"
                placeholder="Enter your new password"
                value={formData.new_password}
                onChangeText={(value) => handleInputChange('new_password', value)}
                secureTextEntry
                error={errors.new_password}
                helperText="Minimum 8 characters with mixed case, numbers, and symbols"
              />

              <WebsiteInput
                label="Confirm New Password"
                placeholder="Confirm your new password"
                value={formData.confirm_password}
                onChangeText={(value) => handleInputChange('confirm_password', value)}
                secureTextEntry
                error={errors.confirm_password}
              />
            </>
          )}

          {formData.action === 'reset_email' && (
            <WebsiteInput
              label="New Email Address"
              placeholder="Enter your new email address"
              value={formData.new_email}
              onChangeText={(value) => handleInputChange('new_email', value)}
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              error={errors.new_email}
              helperText="You'll need to verify this email address after recovery"
            />
          )}

          <WebsiteButton
            variant="primary"
            size="lg"
            onPress={handleRecovery}
            disabled={isLoading}
            style={styles.submitButton}
          >
            {isLoading ? 'Processing...' : 'Recover Account'}
          </WebsiteButton>
        </WebsiteCard>

        <View style={styles.footer}>
          <WebsiteButton
            variant="outline"
            size="sm"
            onPress={() => navigation.goBack()}
            style={styles.backButton}
          >
            ← Back to Login
          </WebsiteButton>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.primary,
  },
  scrollView: {
    flex: 1,
  },
  scrollContent: {
    padding: spacing[4],
    paddingBottom: spacing[8],
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing[6],
  },
  title: {
    textAlign: 'center',
    marginBottom: spacing[2],
  },
  subtitle: {
    textAlign: 'center',
    fontSize: 16,
    lineHeight: 24,
  },
  formCard: {
    marginBottom: spacing[6],
  },
  formTitle: {
    textAlign: 'center',
    marginBottom: spacing[4],
  },
  actionSection: {
    marginVertical: spacing[4],
  },
  sectionTitle: {
    marginBottom: spacing[3],
  },
  actionButtons: {
    flexDirection: 'row',
    justifyContent: 'space-between',
  },
  actionButton: {
    flex: 1,
    marginHorizontal: spacing[1],
  },
  submitButton: {
    marginTop: spacing[4],
  },
  footer: {
    alignItems: 'center',
  },
  backButton: {
    borderColor: colors.surface,
    borderWidth: 1,
    backgroundColor: 'transparent',
  },
});
