import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
  Dimensions,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { ACCOUNT_TYPES, AccountType } from '../config/api';
import {
  WebsiteButton,
  WebsiteCard,
  WebsiteText,
  WebsiteAccountTypeSelector,
} from '../components/WebsiteStyledComponents';
import { colors } from '../theme/colors';
import { spacing } from '../theme/spacing';

const { width: screenWidth } = Dimensions.get('window');

export default function PlanSelectionScreen() {
  const navigation = useNavigation();
  const [selectedPlan, setSelectedPlan] = useState<AccountType>(ACCOUNT_TYPES.FREE);

  // Account type options matching website pricing
  const accountTypeOptions = [
    {
      value: ACCOUNT_TYPES.FREE,
      label: 'Free',
      description: 'Perfect for personal use and getting started',
      features: [
        '1GB storage',
        'Military-grade encryption',
        'Standard support',
      ],
    },
    {
      value: ACCOUNT_TYPES.PREMIUM,
      label: 'Premium',
      description: 'Advanced features for professionals',
      price: '$9.99/month',
      features: [
        '10GB storage',
        'Military-grade encryption',
        'Priority support',
        'Custom domains',
      ],
    },
    {
      value: ACCOUNT_TYPES.ENTERPRISE,
      label: 'Enterprise',
      description: 'Complete solution for organizations',
      price: '$29.99/month',
      features: [
        'Unlimited storage',
        'Military-grade encryption',
        '24/7 dedicated support',
        'Custom domains',
        'Advanced analytics',
      ],
    },
  ];

  const handleContinue = () => {
    // Navigate to account information screen with selected plan
    navigation.navigate('Signup' as never, { selectedPlan } as never);
  };

  return (
    <KeyboardAvoidingView 
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      {/* Background matching website - dark mode */}
      <View style={styles.background} />
      
      <ScrollView 
        style={styles.scrollContainer}
        contentContainerStyle={styles.scrollContent}
        showsVerticalScrollIndicator={false}
      >
        {/* Logo Section */}
        <View style={styles.logoSection}>
          <WebsiteText variant="h1" gradient style={styles.logoText}>
            SecureMail
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.logoSubtext}>
            The World's Most Secure Email
          </WebsiteText>
        </View>

        {/* Header Section */}
        <View style={styles.header}>
          <WebsiteText variant="body" color="secondary" style={styles.description}>
            Experience <Text style={styles.accentText}>military-grade encryption</Text> with zero-knowledge architecture. 
            Your privacy is our priority.
          </WebsiteText>
        </View>

        {/* Plan Selection Card */}
        <WebsiteCard variant="glass" style={styles.formCard}>
          <WebsiteText variant="h4" color="inverse" weight="bold" style={styles.formTitle}>
            Choose Your Plan
          </WebsiteText>
          <WebsiteText variant="body" color="accent" style={styles.formSubtitle}>
            Choose your plan and start your secure email journey
          </WebsiteText>

          {/* Account Type Selection */}
          <View style={styles.section}>
            <WebsiteText variant="h5" color="accent" weight="semibold" style={styles.sectionTitle}>
              Choose Your Plan
            </WebsiteText>
            <WebsiteAccountTypeSelector
              options={accountTypeOptions}
              selectedValue={selectedPlan}
              onSelect={(value) => setSelectedPlan(value as AccountType)}
            />
          </View>

          {/* Continue Button */}
          <WebsiteButton
            variant="primary"
            size="lg"
            onPress={handleContinue}
            style={styles.continueButton}
          >
            Continue to Account Setup
          </WebsiteButton>

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
    paddingBottom: spacing[8], // Add bottom padding for footer
    maxWidth: 1200, // Match website max-width
    alignSelf: 'center',
    width: '100%',
    // Mobile responsiveness
    ...Platform.select({
      web: {
        minHeight: '100vh',
      },
      default: {
        paddingHorizontal: screenWidth > 768 ? spacing[8] : spacing[4],
      },
    }),
  },
  logoSection: {
    alignItems: 'center',
    marginBottom: spacing[6],
  },
  logoText: {
    marginBottom: spacing[1],
    textAlign: 'center',
    fontSize: 48, // Match website h1 size
    fontWeight: '900',
    color: colors.textInverse, // White text for dark mode
  },
  logoSubtext: {
    textAlign: 'center',
    fontSize: 16,
    color: colors.textSecondary, // Gray text for dark mode
  },
  header: {
    alignItems: 'center',
    marginBottom: spacing[8],
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
    marginBottom: spacing[4],
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
    marginBottom: spacing[8],
    lineHeight: 24,
    fontSize: 16,
  },
  section: {
    marginBottom: spacing[6],
  },
  sectionTitle: {
    marginBottom: spacing[4],
    fontSize: 20,
    fontWeight: '600',
  },
  continueButton: {
    marginTop: spacing[2],
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
  footer: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    marginTop: spacing[4],
  },
  loginButton: {
    marginLeft: spacing[2],
    borderColor: colors.surface, // White outline
    borderWidth: 1,
    backgroundColor: 'transparent',
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
