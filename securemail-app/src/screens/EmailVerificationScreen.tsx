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

interface EmailVerificationScreenProps {
  route: {
    params: {
      userId: string;
      username: string;
      email: string;
    };
  };
}

export default function EmailVerificationScreen({ route }: EmailVerificationScreenProps) {
  const navigation = useNavigation();
  const { userId, username, email } = route.params;
  const [verificationCode, setVerificationCode] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleVerifyEmail = async () => {
    if (!verificationCode.trim()) {
      setError('Please enter the verification code');
      return;
    }

    setIsLoading(true);
    setError('');

    try {
      const response = await apiService.verifyEmail({
        user_id: userId,
        code: verificationCode.trim(),
      });

      if (response.success) {
        Alert.alert(
          'Email Verified!',
          'Your email has been verified successfully. Your recovery key has been sent to your email.',
          [
            {
              text: 'OK',
              onPress: () => navigation.navigate('RecoveryKey' as never, {
                recoveryKey: response.recovery_key,
                username: username,
                email: email,
              } as never),
            },
          ]
        );
      }
    } catch (error: any) {
      console.error('Email verification error:', error);
      setError(error.message || 'Email verification failed. Please check your code and try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const handleResendCode = async () => {
    try {
      await apiService.resendVerification();
      Alert.alert('Code Sent', 'A new verification code has been sent to your email.');
    } catch (error: any) {
      Alert.alert('Error', 'Failed to resend verification code. Please try again.');
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
            📧 Verify Your Email
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.subtitle}>
            We've sent a verification code to your external email
          </WebsiteText>
        </View>

        <WebsiteCard variant="glass" style={styles.infoCard}>
          <WebsiteText variant="h5" color="accent" weight="semibold" style={styles.infoTitle}>
            Check Your Email
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.infoText}>
            We've sent a 6-digit verification code to:
          </WebsiteText>
          <WebsiteText variant="body" color="inverse" weight="semibold" style={styles.emailText}>
            {email}
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.infoText}>
            Please enter the code below to verify your email.
          </WebsiteText>
        </WebsiteCard>

        <WebsiteCard variant="glass" style={styles.warningCard}>
          <View style={styles.warningHeader}>
            <Text style={styles.warningIcon}>⚠️</Text>
            <WebsiteText variant="h5" color="warning" weight="bold" style={styles.warningTitle}>
              Critical Security Notice
            </WebsiteText>
          </View>
          <WebsiteText variant="body" color="secondary" style={styles.warningText}>
            After verification, you will receive your RECOVERY KEY via email. This key is EXTREMELY IMPORTANT:
          </WebsiteText>
          <View style={styles.warningList}>
            <WebsiteText variant="body" color="secondary" style={styles.warningItem}>
              • Store it in a secure password manager
            </WebsiteText>
            <WebsiteText variant="body" color="secondary" style={styles.warningItem}>
              • Write it down and store it safely
            </WebsiteText>
            <WebsiteText variant="body" color="secondary" style={styles.warningItem}>
              • Never share it with anyone
            </WebsiteText>
            <WebsiteText variant="body" color="secondary" style={styles.warningItem}>
              • You'll need it to recover your account
            </WebsiteText>
            <WebsiteText variant="body" color="secondary" style={styles.warningItem}>
              • This is the ONLY time you'll receive it
            </WebsiteText>
          </View>
        </WebsiteCard>

        <WebsiteCard variant="glass" style={styles.formCard}>
          <WebsiteText variant="h4" color="inverse" weight="semibold" style={styles.formTitle}>
            Enter Verification Code
          </WebsiteText>
          
          <WebsiteInput
            label="Verification Code"
            placeholder="Enter 6-digit code"
            value={verificationCode}
            onChangeText={(value) => {
              setVerificationCode(value);
              if (error) setError('');
            }}
            keyboardType="numeric"
            maxLength={6}
            error={error}
            helperText="Enter the 6-digit code sent to your email"
          />

          <WebsiteButton
            variant="primary"
            size="lg"
            onPress={handleVerifyEmail}
            disabled={isLoading}
            style={styles.verifyButton}
          >
            {isLoading ? 'Verifying...' : 'Verify Email'}
          </WebsiteButton>

          <View style={styles.resendContainer}>
            <WebsiteText variant="body" color="secondary" style={styles.resendText}>
              Didn't receive the code?
            </WebsiteText>
            <WebsiteButton
              variant="outline"
              size="sm"
              onPress={handleResendCode}
              style={styles.resendButton}
            >
              Resend Code
            </WebsiteButton>
          </View>
        </WebsiteCard>

        <View style={styles.footer}>
          <WebsiteButton
            variant="outline"
            size="sm"
            onPress={() => navigation.goBack()}
            style={styles.backButton}
          >
            ← Back to Signup
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
  infoCard: {
    marginBottom: spacing[4],
  },
  infoTitle: {
    marginBottom: spacing[3],
  },
  infoText: {
    marginBottom: spacing[2],
    lineHeight: 20,
  },
  emailText: {
    marginBottom: spacing[3],
    fontSize: 16,
  },
  warningCard: {
    marginBottom: spacing[4],
    borderColor: colors.warning,
    borderWidth: 2,
  },
  warningHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    marginBottom: spacing[3],
  },
  warningIcon: {
    fontSize: 24,
    marginRight: spacing[2],
  },
  warningTitle: {
    flex: 1,
  },
  warningText: {
    marginBottom: spacing[3],
    lineHeight: 22,
  },
  warningList: {
    paddingLeft: spacing[2],
  },
  warningItem: {
    marginBottom: spacing[2],
    lineHeight: 20,
  },
  formCard: {
    marginBottom: spacing[6],
  },
  formTitle: {
    textAlign: 'center',
    marginBottom: spacing[4],
  },
  verifyButton: {
    marginTop: spacing[4],
  },
  resendContainer: {
    alignItems: 'center',
    marginTop: spacing[4],
  },
  resendText: {
    marginBottom: spacing[2],
  },
  resendButton: {
    borderColor: colors.surface,
    borderWidth: 1,
    backgroundColor: 'transparent',
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
