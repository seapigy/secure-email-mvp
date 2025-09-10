import React, { useState } from 'react';
import {
  View,
  Text,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAuth } from '../contexts/AuthContext';
import { LoginRequest } from '../types';
import {
  WebsiteButton,
  WebsiteInput,
  WebsiteCard,
  WebsiteText,
  WebsiteLoadingSpinner,
  WebsiteSecurityNotice,
} from '../components/WebsiteStyledComponents';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';

export default function WebsiteLoginScreen() {
  const navigation = useNavigation();
  const { login, state } = useAuth();
  const [formData, setFormData] = useState<LoginRequest>({
    email: '',
    password: '',
    mfaCode: '',
    organizationId: '',
  });
  const [errors, setErrors] = useState<Partial<LoginRequest>>({});

  const isLoading = state === 'loading';

  const handleInputChange = (field: keyof LoginRequest, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Partial<LoginRequest> = {};

    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleLogin = async () => {
    if (!validateForm()) return;

    try {
      await login(formData);
      // Navigation will be handled by AuthContext
    } catch (error) {
      console.error('Login error:', error);
    }
  };

  const handleForgotPassword = () => {
    // TODO: Implement forgot password flow
    console.log('Forgot password clicked');
  };

  return (
    <KeyboardAvoidingView 
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      {/* Background matching website */}
      <View style={styles.background} />
      
      <View style={styles.content}>
        {/* Header Section */}
        <View style={styles.header}>
          <WebsiteText variant="h3" color="primary" weight="light" style={styles.subtitle}>
            <Text style={styles.accentText}>Sign in</Text> <Text style={styles.whiteText}>to your secure email</Text>
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.description}>
            Access your encrypted emails with{' '}
            <Text style={styles.accentText}>military-grade security</Text>
          </WebsiteText>
        </View>

        {/* Main Form Card */}
        <WebsiteCard variant="glass" style={styles.formCard}>
          <WebsiteText variant="h4" color="inverse" weight="bold" style={styles.formTitle}>
            Sign In
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.formSubtitle}>
            Enter your credentials to access your account
          </WebsiteText>

          {/* Form Fields */}
          <View style={styles.form}>
            <WebsiteInput
              label="Email Address"
              placeholder="Enter your email address"
              value={formData.email}
              onChangeText={(value) => handleInputChange('email', value)}
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              error={errors.email}
            />

            <WebsiteInput
              label="Password"
              placeholder="Enter your password"
              value={formData.password}
              onChangeText={(value) => handleInputChange('password', value)}
              secureTextEntry
              error={errors.password}
            />

            {/* MFA Code (if needed) */}
            {formData.mfaCode !== undefined && (
              <WebsiteInput
                label="MFA Code"
                placeholder="Enter 6-digit code"
                value={formData.mfaCode}
                onChangeText={(value) => handleInputChange('mfaCode', value)}
                keyboardType="numeric"
                maxLength={6}
                error={errors.mfaCode}
                helperText="Enter the code from your authenticator app"
              />
            )}

            {/* Organization ID (for enterprise users) */}
            {formData.organizationId !== undefined && (
              <WebsiteInput
                label="Organization ID"
                placeholder="Enter your organization ID"
                value={formData.organizationId}
                onChangeText={(value) => handleInputChange('organizationId', value)}
                autoCapitalize="none"
                autoCorrect={false}
                error={errors.organizationId}
                helperText="Required for enterprise accounts"
              />
            )}

            {/* Forgot Password Link */}
            <View style={styles.forgotPasswordContainer}>
              <WebsiteButton
                variant="outline"
                size="sm"
                onPress={handleForgotPassword}
                style={styles.forgotPasswordButton}
              >
                Forgot Password?
              </WebsiteButton>
            </View>

            {/* Account Recovery Link */}
            <View style={styles.recoveryContainer}>
              <WebsiteButton
                variant="outline"
                size="sm"
                onPress={() => navigation.navigate('AccountRecovery' as never)}
                style={styles.recoveryButton}
              >
                Account Recovery
              </WebsiteButton>
            </View>

            {/* Security Notice */}
            <WebsiteSecurityNotice variant="info">
              🔐 Your session is protected with end-to-end encryption. 
              All data remains encrypted and private.
            </WebsiteSecurityNotice>

            {/* Submit Button */}
            <WebsiteButton
              variant="primary"
              size="lg"
              onPress={handleLogin}
              disabled={isLoading}
              loading={isLoading}
              style={styles.submitButton}
            >
              {isLoading ? <WebsiteLoadingSpinner size={20} color={colors.textInverse} /> : 'Sign In'}
            </WebsiteButton>
          </View>
        </WebsiteCard>

        {/* Footer */}
        <View style={styles.footer}>
          <WebsiteText variant="body" color="secondary">
            Don't have an account?
          </WebsiteText>
          <WebsiteButton
            variant="outline"
            size="sm"
            onPress={() => navigation.navigate('PlanSelection' as never)}
            style={styles.signupButton}
          >
            Create Account
          </WebsiteButton>
        </View>
      </View>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.primary, // Dark background like website
  },
  background: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: colors.primary,
    // Add gradient effect similar to website
  },
  content: {
    flex: 1,
    padding: spacing[4],
    justifyContent: 'center',
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing[12],
  },
  title: {
    marginBottom: spacing[4],
    textAlign: 'center',
  },
  subtitle: {
    marginBottom: spacing[4],
    textAlign: 'center',
  },
  description: {
    maxWidth: 400,
    textAlign: 'center',
    lineHeight: 24,
  },
  formCard: {
    marginBottom: spacing[8],
    maxWidth: 500,
    alignSelf: 'center',
    width: '100%',
  },
  formTitle: {
    marginBottom: spacing[2],
    textAlign: 'center',
  },
  formSubtitle: {
    textAlign: 'center',
    marginBottom: spacing[8],
    lineHeight: 24,
  },
  form: {
    marginBottom: spacing[4],
  },
  forgotPasswordContainer: {
    alignItems: 'flex-end',
    marginBottom: spacing[3],
  },
  forgotPasswordButton: {
    paddingVertical: spacing[2],
    paddingHorizontal: spacing[4],
  },
  recoveryContainer: {
    alignItems: 'center',
    marginBottom: spacing[6],
  },
  recoveryButton: {
    paddingVertical: spacing[2],
    paddingHorizontal: spacing[4],
  },
  submitButton: {
    marginBottom: spacing[4],
    width: '100%',
  },
  footer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: spacing[4],
  },
  signupButton: {
    marginLeft: spacing[2],
  },
  // Gradient text styles matching marketing website
  gradientText: {
    color: colors.accent, // Blue gradient text like marketing site
    fontWeight: '900',
  },
  whiteText: {
    color: colors.textInverse, // White text for dark mode
    fontWeight: '900',
  },
  accentText: {
    color: colors.accent, // Blue accent color for highlighted text
    fontWeight: '600',
  },
});

