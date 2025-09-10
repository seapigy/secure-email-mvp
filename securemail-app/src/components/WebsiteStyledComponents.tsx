import React from 'react';
import {
  View,
  Text,
  TextInput,
  TouchableOpacity,
  StyleSheet,
  ViewStyle,
  TextStyle,
  TextInputProps,
  TouchableOpacityProps,
  Platform,
} from 'react-native';
import { colors } from '../theme/colors';
import { spacing, borderRadius } from '../theme/spacing';

// Website-style Button Component
interface WebsiteButtonProps extends TouchableOpacityProps {
  variant?: 'primary' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
  loading?: boolean;
}

export const WebsiteButton: React.FC<WebsiteButtonProps> = ({
  variant = 'primary',
  size = 'md',
  children,
  loading = false,
  style,
  disabled,
  ...props
}) => {
  const buttonStyle = [
    styles.button,
    styles[`button_${variant}`],
    styles[`button_${size}`],
    disabled && styles.buttonDisabled,
    style,
  ];

  const textStyle = [
    styles.buttonText,
    styles[`buttonText_${variant}`],
    styles[`buttonText_${size}`],
    disabled && styles.buttonTextDisabled,
  ];

  return (
    <TouchableOpacity
      style={buttonStyle}
      disabled={disabled || loading}
      activeOpacity={0.8}
      {...props}
    >
      <Text style={textStyle}>
        {loading ? 'Loading...' : children}
      </Text>
    </TouchableOpacity>
  );
};

// Website-style Input Component
interface WebsiteInputProps extends TextInputProps {
  label?: string;
  error?: string;
  helperText?: string;
}

export const WebsiteInput: React.FC<WebsiteInputProps> = ({
  label,
  error,
  helperText,
  style,
  ...props
}) => {
  const inputStyle = [
    styles.input,
    error && styles.inputError,
    style,
  ];

  return (
    <View style={styles.inputContainer}>
      {label && <Text style={styles.inputLabel}>{label}</Text>}
      <TextInput
        style={inputStyle}
        placeholderTextColor={colors.textTertiary}
        {...props}
      />
      {error && <Text style={styles.inputErrorText}>{error}</Text>}
      {helperText && !error && <Text style={styles.inputHelperText}>{helperText}</Text>}
    </View>
  );
};

// Website-style Glass Card Component
interface WebsiteCardProps {
  children: React.ReactNode;
  variant?: 'glass' | 'elevated' | 'default';
  style?: ViewStyle;
}

export const WebsiteCard: React.FC<WebsiteCardProps> = ({
  children,
  variant = 'glass',
  style,
}) => {
  const cardStyle = [
    styles.card,
    styles[`card_${variant}`],
    style,
  ];

  return <View style={cardStyle}>{children}</View>;
};

// Website-style Text Component with gradient support
interface WebsiteTextProps {
  children: React.ReactNode;
  variant?: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'body' | 'caption' | 'label';
  color?: 'primary' | 'secondary' | 'tertiary' | 'inverse' | 'accent' | 'success' | 'warning' | 'error';
  weight?: 'light' | 'normal' | 'medium' | 'semibold' | 'bold' | 'black';
  align?: 'left' | 'center' | 'right';
  gradient?: boolean;
  style?: TextStyle;
}

export const WebsiteText: React.FC<WebsiteTextProps> = ({
  children,
  variant = 'body',
  color = 'primary',
  weight = 'normal',
  align = 'left',
  gradient = false,
  style,
}) => {
  const textStyle = [
    styles.text,
    styles[`text_${variant}`],
    gradient ? styles.textGradient : styles[`text_${color}`],
    styles[`text_${weight}`],
    { textAlign: align },
    style,
  ];

  return <Text style={textStyle}>{children}</Text>;
};

// Website-style Account Type Selector
interface AccountTypeOption {
  value: string;
  label: string;
  description: string;
  price?: string;
  features?: string[];
}

interface WebsiteAccountTypeSelectorProps {
  options: AccountTypeOption[];
  selectedValue: string;
  onSelect: (value: string) => void;
}

export const WebsiteAccountTypeSelector: React.FC<WebsiteAccountTypeSelectorProps> = ({
  options,
  selectedValue,
  onSelect,
}) => {
  return (
    <View style={styles.accountTypeContainer}>
      {options.map((option) => (
        <TouchableOpacity
          key={option.value}
          style={[
            styles.accountTypeOption,
            selectedValue === option.value && styles.accountTypeOptionSelected,
          ]}
          onPress={() => onSelect(option.value)}
          activeOpacity={0.8}
        >
          <View style={styles.accountTypeContent}>
            <View style={styles.accountTypeHeader}>
              <WebsiteText 
                variant="h4" 
                weight="semibold" 
                color={selectedValue === option.value ? 'accent' : 'primary'}
              >
                {option.label}
              </WebsiteText>
              {option.price && (
                <WebsiteText 
                  variant="body" 
                  color={selectedValue === option.value ? 'primary' : 'secondary'}
                  weight="medium"
                >
                  {option.price}
                </WebsiteText>
              )}
            </View>
            <WebsiteText 
              variant="caption" 
              color={selectedValue === option.value ? 'secondary' : 'secondary'}
            >
              {option.description}
            </WebsiteText>
            {option.features && (
              <View style={styles.accountTypeFeatures}>
                {option.features.map((feature, index) => (
                  <WebsiteText 
                    key={index} 
                    variant="caption" 
                    color={selectedValue === option.value ? 'secondary' : 'secondary'}
                  >
                    • {feature}
                  </WebsiteText>
                ))}
              </View>
            )}
          </View>
          <View style={[
            styles.radioButton,
            selectedValue === option.value && styles.radioButtonSelected,
          ]} />
        </TouchableOpacity>
      ))}
    </View>
  );
};

// Website-style Loading Spinner
export const WebsiteLoadingSpinner: React.FC<{ size?: number; color?: string }> = ({
  size = 24,
  color = colors.accent,
}) => {
  return (
    <View style={[styles.spinner, { width: size, height: size }]}>
      <View style={[styles.spinnerInner, { borderColor: color }]} />
    </View>
  );
};

// Website-style Security Notice
interface WebsiteSecurityNoticeProps {
  children: React.ReactNode;
  variant?: 'success' | 'info' | 'warning';
}

export const WebsiteSecurityNotice: React.FC<WebsiteSecurityNoticeProps> = ({
  children,
  variant = 'info',
}) => {
  const noticeStyle = [
    styles.securityNotice,
    styles[`securityNotice_${variant}`],
  ];

  return (
    <View style={noticeStyle}>
      <WebsiteText variant="caption" color="inverse" weight="medium">
        {children}
      </WebsiteText>
    </View>
  );
};

// Styles - Matching Website Design System
const styles = StyleSheet.create({
  // Button styles - matching website .btn-primary and .btn-secondary
  button: {
    borderRadius: borderRadius.xl,
    alignItems: 'center',
    justifyContent: 'center',
    flexDirection: 'row',
    fontFamily: 'Inter',
    fontWeight: '600',
  },
  button_primary: {
    backgroundColor: colors.accent,
    shadowColor: colors.accent,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3,
    shadowRadius: 8,
    elevation: 8,
  },
  button_secondary: {
    backgroundColor: 'transparent',
    borderWidth: 2,
    borderColor: colors.accent,
  },
  button_outline: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: colors.surface, // White outline
  },
  button_sm: {
    paddingVertical: spacing[2],
    paddingHorizontal: spacing[4],
  },
  button_md: {
    paddingVertical: spacing[3],
    paddingHorizontal: spacing[8],
  },
  button_lg: {
    paddingVertical: spacing[4],
    paddingHorizontal: spacing[8],
  },
  buttonDisabled: {
    backgroundColor: colors.gray300,
    borderColor: colors.gray300,
    shadowOpacity: 0,
    elevation: 0,
  },
  buttonText: {
    fontFamily: 'Inter, system-ui, sans-serif', // Exact font match to website
    fontWeight: '600',
  },
  buttonText_primary: {
    color: colors.textInverse,
  },
  buttonText_secondary: {
    color: colors.accent,
  },
  buttonText_outline: {
    color: colors.accent, // Blue text for outline buttons
  },
  buttonText_sm: {
    fontSize: 14,
  },
  buttonText_md: {
    fontSize: 16,
  },
  buttonText_lg: {
    fontSize: 18,
  },
  buttonTextDisabled: {
    color: colors.textTertiary,
  },

  // Input styles - matching website form styling
  inputContainer: {
    marginBottom: spacing[4],
  },
  inputLabel: {
    fontSize: 14,
    fontWeight: '500',
    color: colors.textInverse, // White text for dark mode
    marginBottom: spacing[2],
    fontFamily: 'Inter, system-ui, sans-serif', // Exact font match
  },
  input: {
    borderWidth: 1,
    borderColor: colors.borderDarkMode, // Dark mode border
    borderRadius: borderRadius.lg,
    paddingVertical: spacing[3],
    paddingHorizontal: spacing[4],
    backgroundColor: `${colors.primaryLight}60`, // Dark mode input background
    fontSize: 16,
    color: colors.textInverse, // White text for dark mode
    fontFamily: 'Inter, system-ui, sans-serif', // Exact font match
    ...Platform.select({
      web: {
        backdropFilter: 'blur(8px)',
      },
    }),
  },
  inputError: {
    borderColor: colors.error,
  },
  inputErrorText: {
    fontSize: 14,
    color: colors.error,
    marginTop: spacing[1],
    fontFamily: 'Inter, system-ui, sans-serif', // Exact font match
  },
  inputHelperText: {
    fontSize: 14,
    color: colors.textSecondary,
    marginTop: spacing[1],
    fontFamily: 'Inter, system-ui, sans-serif', // Exact font match
  },

  // Card styles - matching website .glass-effect and .feature-card
  card: {
    borderRadius: borderRadius['2xl'],
    padding: spacing[6],
    fontFamily: 'Inter',
  },
  card_glass: {
    backgroundColor: `${colors.primaryLight}80`, // 50% opacity for dark mode glass effect
    borderWidth: 1,
    borderColor: colors.borderDarkMode, // Dark mode border
    ...Platform.select({
      web: {
        backdropFilter: 'blur(12px)',
        WebkitBackdropFilter: 'blur(12px)',
      },
    }),
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.3, // Stronger shadow for dark mode
    shadowRadius: 8,
    elevation: 8,
  },
  card_elevated: {
    backgroundColor: colors.surface,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.15,
    shadowRadius: 15,
    elevation: 8,
  },
  card_default: {
    backgroundColor: colors.surface,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
    elevation: 4,
  },

  // Text styles - matching website typography
  text: {
    fontFamily: 'Inter, system-ui, sans-serif', // Exact font match
  },
  text_h1: {
    fontSize: 48,
    fontWeight: '900',
    lineHeight: 60,
  },
  text_h2: {
    fontSize: 36,
    fontWeight: '700',
    lineHeight: 45,
  },
  text_h3: {
    fontSize: 30,
    fontWeight: '700',
    lineHeight: 37,
  },
  text_h4: {
    fontSize: 24,
    fontWeight: '600',
    lineHeight: 30,
  },
  text_h5: {
    fontSize: 20,
    fontWeight: '600',
    lineHeight: 25,
  },
  text_body: {
    fontSize: 16,
    lineHeight: 24,
  },
  text_caption: {
    fontSize: 14,
    lineHeight: 21,
  },
  text_label: {
    fontSize: 14,
    fontWeight: '500',
    lineHeight: 21,
  },
  text_primary: { color: colors.textPrimary },
  text_secondary: { color: colors.textSecondary },
  text_tertiary: { color: colors.textTertiary },
  text_inverse: { color: colors.textInverse },
  text_accent: { color: colors.accent },
  text_success: { color: colors.success },
  text_warning: { color: colors.warning },
  text_error: { color: colors.error },
  textGradient: { color: colors.accent }, // Gradient text effect
  text_light: { fontWeight: '300' },
  text_normal: { fontWeight: '400' },
  text_medium: { fontWeight: '500' },
  text_semibold: { fontWeight: '600' },
  text_bold: { fontWeight: '700' },
  text_black: { fontWeight: '900' },

  // Account type selector - matching website feature cards
  accountTypeContainer: {
    marginBottom: spacing[6],
  },
  accountTypeOption: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: spacing[4],
    borderWidth: 1,
    borderColor: colors.borderLight,
    borderRadius: borderRadius['2xl'],
    marginBottom: spacing[3],
    backgroundColor: `${colors.surface}CC`, // Glass effect
    ...Platform.select({
      web: {
        backdropFilter: 'blur(8px)',
        WebkitBackdropFilter: 'blur(8px)',
      },
    }),
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  accountTypeOptionSelected: {
    borderColor: colors.accent,
    backgroundColor: `${colors.accent}10`, // 10% accent color
    shadowColor: colors.accent,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.2,
    shadowRadius: 8,
    elevation: 4,
  },
  accountTypeContent: {
    flex: 1,
  },
  accountTypeHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: spacing[1],
  },
  accountTypeFeatures: {
    marginTop: spacing[2],
  },
  radioButton: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: colors.borderMedium,
    marginLeft: spacing[3],
  },
  radioButtonSelected: {
    borderColor: colors.accent,
    backgroundColor: colors.accent,
  },

  // Loading spinner
  spinner: {
    justifyContent: 'center',
    alignItems: 'center',
  },
  spinnerInner: {
    width: '100%',
    height: '100%',
    borderRadius: 50,
    borderWidth: 2,
    borderTopColor: 'transparent',
    borderRightColor: 'transparent',
    borderBottomColor: 'transparent',
  },

  // Security notice - matching website styling
  securityNotice: {
    padding: spacing[4],
    borderRadius: borderRadius.lg,
    marginBottom: spacing[4],
    // Glass effect matching website
    ...Platform.select({
      web: {
        backdropFilter: 'blur(8px)',
        WebkitBackdropFilter: 'blur(8px)',
      },
    }),
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.05,
    shadowRadius: 4,
    elevation: 2,
  },
  securityNotice_success: {
    backgroundColor: `${colors.success}15`, // 15% opacity for subtle effect
    borderWidth: 1,
    borderColor: colors.success,
  },
  securityNotice_info: {
    backgroundColor: `${colors.accent}15`, // 15% opacity for subtle effect
    borderWidth: 1,
    borderColor: colors.accent,
  },
  securityNotice_warning: {
    backgroundColor: `${colors.warning}15`, // 15% opacity for subtle effect
    borderWidth: 1,
    borderColor: colors.warning,
  },
});

