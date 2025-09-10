import React, { useState } from 'react';
import {
  View,
  Text,
  ScrollView,
  KeyboardAvoidingView,
  Platform,
  StyleSheet,
  Alert,
  Clipboard,
  Share,
} from 'react-native';
import { useNavigation } from '@react-navigation/native';
import { WebsiteButton, WebsiteCard, WebsiteText } from '../components/WebsiteStyledComponents';
import { colors, spacing } from '../theme';

interface RecoveryKeyScreenProps {
  route: {
    params: {
      recoveryKey: string;
      username: string;
      email: string;
    };
  };
}

export default function RecoveryKeyScreen({ route }: RecoveryKeyScreenProps) {
  const navigation = useNavigation();
  const { recoveryKey, username, email } = route.params;
  const [hasCopied, setHasCopied] = useState(false);

  const handleCopyKey = async () => {
    try {
      await Clipboard.setString(recoveryKey);
      setHasCopied(true);
      Alert.alert('Copied!', 'Recovery key copied to clipboard');
    } catch (error) {
      Alert.alert('Error', 'Failed to copy recovery key');
    }
  };

  const handleShareKey = async () => {
    try {
      await Share.share({
        message: `SecureMail Recovery Key for ${username} (${email}):\n\n${recoveryKey}\n\n⚠️ IMPORTANT: Store this key securely. You will need it to recover your account if you lose access.`,
        title: 'SecureMail Recovery Key',
      });
    } catch (error) {
      Alert.alert('Error', 'Failed to share recovery key');
    }
  };

  const handleContinue = () => {
    navigation.navigate('Login' as never);
  };

  return (
    <KeyboardAvoidingView 
      style={styles.container}
      behavior={Platform.OS === 'ios' ? 'padding' : 'height'}
    >
      <ScrollView style={styles.scrollView} contentContainerStyle={styles.content}>
      <View style={styles.header}>
        <WebsiteText variant="h2" color="inverse" weight="bold" style={styles.title}>
          🔐 Save Your Recovery Key
        </WebsiteText>
        <WebsiteText variant="body" color="secondary" style={styles.subtitle}>
          This is the only time you'll see this recovery key
        </WebsiteText>
      </View>

      <WebsiteCard variant="glass" style={styles.warningCard}>
        <View style={styles.warningHeader}>
          <Text style={styles.warningIcon}>⚠️</Text>
          <WebsiteText variant="h5" color="warning" weight="bold" style={styles.warningTitle}>
            Critical Security Information
          </WebsiteText>
        </View>
        <WebsiteText variant="body" color="secondary" style={styles.warningText}>
          Your recovery key is the only way to regain access to your account if you lose your password or your primary email is compromised. Store it securely and never share it with anyone.
        </WebsiteText>
      </WebsiteCard>

      <WebsiteCard variant="glass" style={styles.keyCard}>
        <WebsiteText variant="h5" color="accent" weight="semibold" style={styles.keyLabel}>
          Your Recovery Key
        </WebsiteText>
        <View style={styles.keyContainer}>
          <Text style={styles.recoveryKey}>{recoveryKey}</Text>
        </View>
        <View style={styles.keyActions}>
          <WebsiteButton
            variant="outline"
            size="sm"
            onPress={handleCopyKey}
            style={styles.actionButton}
          >
            {hasCopied ? '✓ Copied' : '📋 Copy'}
          </WebsiteButton>
          <WebsiteButton
            variant="outline"
            size="sm"
            onPress={handleShareKey}
            style={styles.actionButton}
          >
            📤 Share
          </WebsiteButton>
        </View>
      </WebsiteCard>

      <WebsiteCard variant="glass" style={styles.infoCard}>
        <WebsiteText variant="h5" color="accent" weight="semibold" style={styles.infoTitle}>
          What to do with this key:
        </WebsiteText>
        <View style={styles.infoList}>
          <WebsiteText variant="body" color="secondary" style={styles.infoItem}>
            • Store it in a secure password manager
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.infoItem}>
            • Write it down and store it in a safe place
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.infoItem}>
            • Never store it in plain text on your computer
          </WebsiteText>
          <WebsiteText variant="body" color="secondary" style={styles.infoItem}>
            • You'll need it along with your fallback email to recover your account
          </WebsiteText>
        </View>
      </WebsiteCard>

      <View style={styles.footer}>
        <WebsiteButton
          variant="primary"
          size="lg"
          onPress={handleContinue}
          style={styles.continueButton}
        >
          I've Saved My Recovery Key - Continue
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
  content: {
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
    lineHeight: 22,
  },
  keyCard: {
    marginBottom: spacing[4],
  },
  keyLabel: {
    marginBottom: spacing[3],
    textAlign: 'center',
  },
  keyContainer: {
    backgroundColor: colors.primaryDark,
    borderRadius: 8,
    padding: spacing[3],
    marginBottom: spacing[3],
    borderWidth: 1,
    borderColor: colors.borderDarkMode,
  },
  recoveryKey: {
    fontFamily: 'monospace',
    fontSize: 14,
    color: colors.textInverse,
    textAlign: 'center',
    lineHeight: 20,
  },
  keyActions: {
    flexDirection: 'row',
    justifyContent: 'space-around',
  },
  actionButton: {
    flex: 1,
    marginHorizontal: spacing[1],
  },
  infoCard: {
    marginBottom: spacing[6],
  },
  infoTitle: {
    marginBottom: spacing[3],
  },
  infoList: {
    paddingLeft: spacing[2],
  },
  infoItem: {
    marginBottom: spacing[2],
    lineHeight: 20,
  },
  footer: {
    marginTop: spacing[4],
  },
  continueButton: {
    width: '100%',
  },
});
