import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
  Dimensions,
  TouchableOpacity,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { useAuth } from '../contexts/AuthContext';
import { ACCOUNT_TYPES, AccountType } from '../config/api';
import { SignupRequest } from '../types';
import {
  WebsiteButton,
  WebsiteInput,
  WebsiteCard,
  WebsiteText,
  WebsiteAccountTypeSelector,
  WebsiteLoadingSpinner,
  WebsiteSecurityNotice,
} from '../components/WebsiteStyledComponents';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';

const { width: screenWidth } = Dimensions.get('window');

export default function WebsiteSignupScreen({ route }: any) {
  const navigation = useNavigation();
  const { signup, state } = useAuth();
  const selectedPlan = route?.params?.selectedPlan || ACCOUNT_TYPES.FREE;
  
  const [formData, setFormData] = useState<SignupRequest & { confirmPassword: string; organizationName: string; organizationDomain: string; setupMFA: boolean }>({
    username: '',
    email: '',
    password: '',
    confirmPassword: '',
    accountType: selectedPlan,
    fallback_email: '', // Must be different from primary email
    organizationName: '',
    organizationDomain: '',
    setupMFA: false,
  });
  const [errors, setErrors] = useState<Partial<SignupRequest & { confirmPassword: string; organizationName: string; organizationDomain: string; setupMFA: boolean }>>({});

  const isLoading = state === 'loading';

  // Get plan display info
  const getPlanInfo = (planType: AccountType) => {
    switch (planType) {
      case ACCOUNT_TYPES.FREE:
        return { label: 'Free', price: '$0/month' };
      case ACCOUNT_TYPES.PREMIUM:
        return { label: 'Premium', price: '$9.99/month' };
      case ACCOUNT_TYPES.ENTERPRISE:
        return { label: 'Enterprise', price: '$29.99/month' };
      default:
        return { label: 'Free', price: '$0/month' };
    }
  };

  const planInfo = getPlanInfo(selectedPlan);

  const handleInputChange = (field: keyof SignupRequest, value: string) => {
    setFormData(prev => ({ ...prev, [field]: value }));
    if (errors[field]) {
      setErrors(prev => ({ ...prev, [field]: undefined }));
    }
  };

  const validateForm = (): boolean => {
    const newErrors: Partial<SignupRequest & { confirmPassword: string; organizationName: string; organizationDomain: string }> = {};

    if (!formData.username) {
      newErrors.username = 'Username is required';
    } else if (!/^[a-zA-Z0-9._-]+$/.test(formData.username)) {
      newErrors.username = 'Username can only contain letters, numbers, dots, underscores, and hyphens';
    }

    if (!formData.email) {
      newErrors.email = 'Email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.email)) {
      newErrors.email = 'Please enter a valid email address';
    }

    if (!formData.fallback_email) {
      newErrors.fallback_email = 'Fallback email is required';
    } else if (!/\S+@\S+\.\S+/.test(formData.fallback_email)) {
      newErrors.fallback_email = 'Please enter a valid fallback email address';
    } else if (formData.fallback_email.toLowerCase() === formData.email.toLowerCase()) {
      newErrors.fallback_email = 'Fallback email must be different from your primary email';
    }

    if (!formData.password) {
      newErrors.password = 'Password is required';
    } else {
      const password = formData.password;
      const errors: string[] = [];
      
      if (password.length < 8) {
        errors.push('at least 8 characters');
      }
      if (!/[a-z]/.test(password)) {
        errors.push('one lowercase letter');
      }
      if (!/[A-Z]/.test(password)) {
        errors.push('one uppercase letter');
      }
      if (!/[0-9]/.test(password)) {
        errors.push('one number');
      }
      if (!/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(password)) {
        errors.push('one special character');
      }
      
      if (errors.length > 0) {
        newErrors.password = `Password must contain ${errors.join(', ')}`;
      }
    }

    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match';
    }

    if (formData.accountType !== ACCOUNT_TYPES.FREE) {
      if (!formData.organizationName) {
        newErrors.organizationName = 'Organization name is required for paid plans';
      }
      if (!formData.organizationDomain) {
        newErrors.organizationDomain = 'Organization domain is required for paid plans';
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSignup = async () => {
    if (!validateForm()) return;

    try {
      // Use the provided fallback email (must be different from primary email)
      const signupData = {
        ...formData,
      };
      
      const response = await signup(signupData);
      // Navigate to email verification screen
      navigation.navigate('EmailVerification' as never, {
        userId: response.id,
        username: response.username,
        email: response.email,
      } as never);
    } catch (error) {
      console.error('Signup error:', error);
    }
  };

  return (
    <KeyboardAvoidingView 
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      {/* Background matching website - gradient from white to gray-50 */}
      <View style={styles.background} />
      
      <ScrollView 
        style={styles.scrollContainer}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Logo Section */}
        <View style={styles.logoSection}>
          <View style={styles.logoContainer}>
            <WebsiteText variant="h2" gradient style={styles.logoText}>
              SecureMail
            </WebsiteText>
            <WebsiteText variant="caption" color="secondary" style={styles.logoSubtext}>
              The World's Most Secure Email
            </WebsiteText>
          </View>
        </View>

        {/* Header Section */}
        <View style={styles.header}>
          <WebsiteText variant="h1" gradient style={styles.title}>
            <Text style={styles.gradientText}>Create</Text> Your Account
          </WebsiteText>
          <WebsiteText variant="h3" color="primary" weight="light" style={styles.subtitle}>
            <Text style={styles.accentText}>{planInfo.label}</Text> Plan - {planInfo.price}
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.description}>
            Complete your account setup to get started with secure email
          </WebsiteText>
        </View>

        {/* Main Form Card */}
        <WebsiteCard variant="glass" style={styles.formCard}>
          {/* Back Button */}
          <View style={styles.backButtonContainer}>
            <WebsiteButton
              variant="outline"
              size="sm"
              onPress={() => navigation.goBack()}
              style={styles.backButton}
            >
              ← Back to Plan Selection
            </WebsiteButton>
          </View>

          <WebsiteText variant="h4" color="inverse" weight="semibold" style={styles.formTitle}>
            Create Your Account
          </WebsiteText>
          <WebsiteText variant="body" color="accent" style={styles.formSubtitle}>
            Complete your account information to get started
          </WebsiteText>

          {/* Form Fields */}
          <View style={styles.section}>
            <WebsiteText variant="h5" color="accent" weight="semibold" style={styles.sectionTitle}>
              Account Information
            </WebsiteText>
            
            <WebsiteInput
              label="Username"
              placeholder="Choose a username"
              value={formData.username}
              onChangeText={(value) => handleInputChange('username', value)}
              autoCapitalize="none"
              autoCorrect={false}
              error={errors.username}
              helperText="Alphanumeric characters, dots, underscores, and hyphens only"
            />

            <WebsiteInput
              label="External Email Address"
              placeholder="Enter your Gmail, Outlook, or other email"
              value={formData.email}
              onChangeText={(value) => handleInputChange('email', value)}
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              error={errors.email}
              helperText="Your Gmail, Outlook, or other email for notifications and account recovery"
            />

            <WebsiteInput
              label="Fallback Email Address"
              placeholder="Enter a different email for account recovery"
              value={formData.fallback_email}
              onChangeText={(value) => handleInputChange('fallback_email', value)}
              keyboardType="email-address"
              autoCapitalize="none"
              autoCorrect={false}
              error={errors.fallback_email}
              helperText="Must be different from your primary email - used for account recovery"
            />

            <WebsiteInput
              label="Password"
              placeholder="Create a strong password"
              value={formData.password}
              onChangeText={(value) => handleInputChange('password', value)}
              secureTextEntry
              error={errors.password}
              helperText="Must contain: 8+ characters, uppercase, lowercase, number, special character"
            />

            <WebsiteInput
              label="Confirm Password"
              placeholder="Confirm your password"
              value={formData.confirmPassword}
              onChangeText={(value) => handleInputChange('confirmPassword', value)}
              secureTextEntry
              error={errors.confirmPassword}
            />

            {/* Organization fields for paid plans */}
            {formData.accountType !== ACCOUNT_TYPES.FREE && (
              <>
                <WebsiteInput
                  label="Organization Name"
                  placeholder="Enter your organization name"
                  value={formData.organizationName}
                  onChangeText={(value) => handleInputChange('organizationName', value)}
                  error={errors.organizationName}
                />

                <WebsiteInput
                  label="Organization Domain"
                  placeholder="yourcompany.com"
                  value={formData.organizationDomain}
                  onChangeText={(value) => handleInputChange('organizationDomain', value)}
                  autoCapitalize="none"
                  autoCorrect={false}
                  error={errors.organizationDomain}
                  helperText="We'll verify domain ownership before activation"
                />
              </>
            )}

            {/* MFA Setup Toggle */}
            <View style={styles.mfaSection}>
              <TouchableOpacity
                style={styles.mfaToggle}
                onPress={() => handleInputChange('setupMFA', !formData.setupMFA)}
              >
                <View style={[styles.toggleSwitch, formData.setupMFA && styles.toggleSwitchActive]}>
                  <View style={[styles.toggleThumb, formData.setupMFA && styles.toggleThumbActive]} />
                </View>
                <View style={styles.mfaTextContainer}>
                  <WebsiteText variant="body" color="inverse" weight="medium" style={styles.mfaLabel}>
                    Enable Two-Factor Authentication (2FA)
                  </WebsiteText>
                  <WebsiteText variant="caption" color="secondary" style={styles.mfaDescription}>
                    Add an extra layer of security with TOTP authenticator apps
                  </WebsiteText>
                </View>
              </TouchableOpacity>
            </View>
          </View>

          {/* Security Notice */}
          <WebsiteSecurityNotice variant="success">
            🔒 Your data is protected with AES-256-GCM encryption and zero-knowledge architecture. 
            We cannot see your emails, not even metadata.
          </WebsiteSecurityNotice>

          {/* Submit Button */}
          <WebsiteButton
            variant="primary"
            size="lg"
            onPress={handleSignup}
            disabled={isLoading}
            loading={isLoading}
            style={styles.submitButton}
          >
            {isLoading ? <WebsiteLoadingSpinner size={20} color={colors.textInverse} /> : 'Create Account'}
          </WebsiteButton>

          {/* Trial Notice for Paid Plans */}
          {formData.accountType !== ACCOUNT_TYPES.FREE && (
            <WebsiteCard variant="glass" style={styles.trialNotice}>
              <WebsiteText variant="body" color="accent" weight="medium" style={styles.trialText}>
                🎉 Start with a 1-month free trial! No credit card required.
              </WebsiteText>
              <WebsiteText variant="caption" color="secondary" style={styles.trialSubtext}>
                Cancel anytime during the trial period with no charges.
              </WebsiteText>
            </WebsiteCard>
          )}
        </WebsiteCard>

        {/* Footer */}
        <View style={styles.footer}>
          <WebsiteText variant="body" color="secondary">
            Already have an account?
          </WebsiteText>
          <WebsiteButton
            variant="outline"
            size="sm"
            onPress={() => navigation.navigate('Login' as never)}
            style={styles.loginButton}
          >
            Sign In
          </WebsiteButton>
        </View>
      </ScrollView>
    </KeyboardAvoidingView>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.primary, // Dark background matching website default
  },
  background: {
    position: 'absolute',
    top: 0,
    left: 0,
    right: 0,
    bottom: 0,
    backgroundColor: colors.primary, // Dark background matching website
    // Dark mode gradient effect matching website
  },
  scrollContainer: {
    flex: 1,
  },
  scrollContent: {
    flexGrow: 1,
    padding: spacing[4],
    paddingTop: spacing[20], // Account for header
    maxWidth: 1200, // Match website max-width
    alignSelf: 'center',
    width: '100%',
    // Mobile responsiveness
    ...Platform.select({
      web: {
        minHeight: '100vh',
      },
      default: {
        minHeight: screenWidth > 768 ? '100vh' : 'auto',
      },
    }),
  },
  logoSection: {
    alignItems: 'center',
    marginBottom: spacing[8],
  },
  logoContainer: {
    alignItems: 'center',
  },
  logoText: {
    marginBottom: spacing[1],
    textAlign: 'center',
  },
  logoSubtext: {
    textAlign: 'center',
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing[12],
  },
  title: {
    marginBottom: spacing[4],
    textAlign: 'center',
    fontSize: 48, // Match website h1 size
    fontWeight: '900',
    color: colors.textInverse, // White text for dark mode
  },
  subtitle: {
    marginBottom: spacing[4],
    textAlign: 'center',
    fontSize: 30, // Match website h3 size
    fontWeight: '700',
    color: colors.textInverse, // White text for dark mode
  },
  description: {
    maxWidth: 500,
    textAlign: 'center',
    lineHeight: 24,
    fontSize: 16,
    color: colors.textSecondary, // Gray text for dark mode
  },
  formCard: {
    marginBottom: spacing[8],
    maxWidth: 600,
    alignSelf: 'center',
    width: '100%',
    // Glass effect matching website dark mode
    backgroundColor: `${colors.primaryLight}80`, // 50% opacity for dark mode glass effect
    borderWidth: 1,
    borderColor: colors.borderDarkMode, // Dark mode border
    borderRadius: 24, // 2xl border radius
    padding: spacing[6],
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3, // Stronger shadow for dark mode
    shadowRadius: 8,
    elevation: 8,
    // Mobile responsiveness
    ...Platform.select({
      web: {
        minWidth: 400,
        backdropFilter: 'blur(12px)',
        WebkitBackdropFilter: 'blur(12px)',
      },
      default: {
        marginHorizontal: screenWidth > 768 ? 0 : spacing[2],
        padding: screenWidth > 768 ? spacing[6] : spacing[4],
      },
    }),
  },
  formTitle: {
    marginBottom: spacing[2],
    textAlign: 'center',
    fontSize: 24,
    fontWeight: '600',
  },
  formSubtitle: {
    textAlign: 'center',
    marginBottom: spacing[4],
    lineHeight: 24,
    fontSize: 16,
  },
  backButtonContainer: {
    alignItems: 'flex-start',
    marginBottom: spacing[4],
  },
  backButton: {
    borderColor: colors.surface, // White outline
    borderWidth: 1,
    backgroundColor: 'transparent',
  },
  section: {
    marginBottom: spacing[8],
  },
  sectionTitle: {
    marginBottom: spacing[4],
    fontSize: 20,
    fontWeight: '600',
  },
  submitButton: {
    marginBottom: spacing[4],
    width: '100%',
    // Match website button styling
    backgroundColor: colors.accent,
    borderRadius: 20, // xl border radius
    paddingVertical: spacing[4],
    paddingHorizontal: spacing[8],
    shadowColor: colors.accent,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  trialNotice: {
    marginBottom: spacing[4],
    backgroundColor: `${colors.accent}15`, // 15% accent color
    borderColor: colors.accent,
    borderWidth: 1,
    borderRadius: 16, // lg border radius
    padding: spacing[4],
  },
  trialText: {
    textAlign: 'center',
    marginBottom: spacing[2],
    fontSize: 16,
    fontWeight: '500',
  },
  trialSubtext: {
    textAlign: 'center',
    fontSize: 14,
  },
  footer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: spacing[4],
  },
  loginButton: {
    marginLeft: spacing[2],
  },
  // Gradient text styles matching marketing website
  gradientText: {
    color: colors.accent, // Blue gradient text like marketing site
    fontWeight: '900',
  },
  accentText: {
    color: colors.accent, // Blue accent color for highlighted text
    fontWeight: '600',
  },
  mfaSection: {
    marginTop: spacing[4],
    marginBottom: spacing[4],
  },
  mfaToggle: {
    flexDirection: 'row',
    alignItems: 'center',
    paddingVertical: spacing[3],
    paddingHorizontal: spacing[4],
    backgroundColor: colors.surface,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: colors.borderDarkMode,
  },
  toggleSwitch: {
    width: 50,
    height: 28,
    borderRadius: 14,
    backgroundColor: colors.gray400,
    justifyContent: 'center',
    paddingHorizontal: 2,
  },
  toggleSwitchActive: {
    backgroundColor: colors.accent,
  },
  toggleThumb: {
    width: 24,
    height: 24,
    borderRadius: 12,
    backgroundColor: colors.textInverse,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.2,
    shadowRadius: 2,
    elevation: 2,
  },
  toggleThumbActive: {
    transform: [{ translateX: 22 }],
  },
  mfaTextContainer: {
    flex: 1,
    marginLeft: spacing[3],
  },
  mfaLabel: {
    fontSize: 16,
    fontWeight: '500',
    marginBottom: 2,
  },
  mfaDescription: {
    fontSize: 14,
    lineHeight: 18,
  },
});

