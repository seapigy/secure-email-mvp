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
} from 'react-native';
import { colors } from '../theme/colors';
import { spacing, borderRadius } from '../theme/spacing';

// Styled Button Component
interface StyledButtonProps extends TouchableOpacityProps {
  variant?: 'primary' | 'secondary' | 'outline';
  size?: 'sm' | 'md' | 'lg';
  children: React.ReactNode;
  loading?: boolean;
}

export const StyledButton: React.FC<StyledButtonProps> = ({
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

// Styled Input Component
interface StyledInputProps extends TextInputProps {
  label?: string;
  error?: string;
  helperText?: string;
  variant?: 'default' | 'filled' | 'outlined';
}

export const StyledInput: React.FC<StyledInputProps> = ({
  label,
  error,
  helperText,
  variant = 'default',
  style,
  ...props
}) => {
  const inputStyle = [
    styles.input,
    styles[`input_${variant}`],
    error && styles.inputError,
    style,
  ];

  return (
    <View style={styles.inputContainer}>
      {label && <Text style={styles.inputLabel}>{label}</Text>}
      <TextInput
        style={inputStyle}
        placeholderTextColor={theme.colors.textTertiary}
        {...props}
      />
      {error && <Text style={styles.inputErrorText}>{error}</Text>}
      {helperText && !error && <Text style={styles.inputHelperText}>{helperText}</Text>}
    </View>
  );
};

// Styled Card Component
interface StyledCardProps {
  children: React.ReactNode;
  variant?: 'default' | 'glass' | 'elevated';
  style?: ViewStyle;
}

export const StyledCard: React.FC<StyledCardProps> = ({
  children,
  variant = 'default',
  style,
}) => {
  const cardStyle = [
    styles.card,
    styles[`card_${variant}`],
    style,
  ];

  return <View style={cardStyle}>{children}</View>;
};

// Styled Text Component
interface StyledTextProps {
  children: React.ReactNode;
  variant?: 'h1' | 'h2' | 'h3' | 'h4' | 'body' | 'caption' | 'label';
  color?: 'primary' | 'secondary' | 'tertiary' | 'inverse' | 'accent' | 'success' | 'warning' | 'error';
  weight?: 'light' | 'normal' | 'medium' | 'semibold' | 'bold';
  align?: 'left' | 'center' | 'right';
  style?: TextStyle;
}

export const StyledText: React.FC<StyledTextProps> = ({
  children,
  variant = 'body',
  color = 'primary',
  weight = 'normal',
  align = 'left',
  style,
}) => {
  const textStyle = [
    styles.text,
    styles[`text_${variant}`],
    styles[`text_${color}`],
    styles[`text_${weight}`],
    { textAlign: align },
    style,
  ];

  return <Text style={textStyle}>{children}</Text>;
};

// Account Type Selector Component
interface AccountTypeOption {
  value: string;
  label: string;
  description: string;
  price?: string;
  features?: string[];
}

interface AccountTypeSelectorProps {
  options: AccountTypeOption[];
  selectedValue: string;
  onSelect: (value: string) => void;
}

export const AccountTypeSelector: React.FC<AccountTypeSelectorProps> = ({
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
              <StyledText variant="h4" weight="semibold" color={selectedValue === option.value ? 'accent' : 'primary'}>
                {option.label}
              </StyledText>
              {option.price && (
                <StyledText variant="body" color={selectedValue === option.value ? 'accent' : 'secondary'}>
                  {option.price}
                </StyledText>
              )}
            </View>
            <StyledText variant="caption" color={selectedValue === option.value ? 'accent' : 'secondary'}>
              {option.description}
            </StyledText>
            {option.features && (
              <View style={styles.accountTypeFeatures}>
                {option.features.map((feature, index) => (
                  <StyledText key={index} variant="caption" color={selectedValue === option.value ? 'accent' : 'tertiary'}>
                    • {feature}
                  </StyledText>
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

// Loading Spinner Component
export const LoadingSpinner: React.FC<{ size?: number; color?: string }> = ({
  size = 24,
  color = theme.colors.accent,
}) => {
  return (
    <View style={[styles.spinner, { width: size, height: size }]}>
      <View style={[styles.spinnerInner, { borderColor: color }]} />
    </View>
  );
};

// Styles
const styles = StyleSheet.create({
  // Button styles
  button: {
    borderRadius: borderRadius.xl,
    alignItems: 'center',
    justifyContent: 'center',
    flexDirection: 'row',
  },
  button_primary: {
    backgroundColor: colors.accent,
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 1,
  },
  button_secondary: {
    backgroundColor: colors.surface,
    borderWidth: 2,
    borderColor: colors.accent,
  },
  button_outline: {
    backgroundColor: 'transparent',
    borderWidth: 1,
    borderColor: colors.borderMedium,
  },
  button_sm: {
    paddingVertical: spacing[2],
    paddingHorizontal: spacing[4],
  },
  button_md: {
    paddingVertical: spacing[3],
    paddingHorizontal: spacing[6],
  },
  button_lg: {
    paddingVertical: spacing[4],
    paddingHorizontal: spacing[8],
  },
  buttonDisabled: {
    backgroundColor: theme.colors.gray300,
    borderColor: theme.colors.gray300,
  },
  buttonText: {
    fontWeight: theme.typography.fontWeight.semibold,
  },
  buttonText_primary: {
    color: theme.colors.textInverse,
  },
  buttonText_secondary: {
    color: theme.colors.accent,
  },
  buttonText_outline: {
    color: theme.colors.textPrimary,
  },
  buttonText_sm: {
    fontSize: theme.typography.fontSize.sm,
  },
  buttonText_md: {
    fontSize: theme.typography.fontSize.base,
  },
  buttonText_lg: {
    fontSize: theme.typography.fontSize.lg,
  },
  buttonTextDisabled: {
    color: theme.colors.textTertiary,
  },

  // Input styles
  inputContainer: {
    marginBottom: theme.spacing[4],
  },
  inputLabel: {
    fontSize: theme.typography.fontSize.sm,
    fontWeight: theme.typography.fontWeight.medium,
    color: theme.colors.textPrimary,
    marginBottom: theme.spacing[2],
  },
  input: {
    borderWidth: 1,
    borderColor: theme.colors.borderLight,
    borderRadius: theme.borderRadius.lg,
    paddingVertical: theme.spacing[3],
    paddingHorizontal: theme.spacing[4],
    backgroundColor: theme.colors.surface,
    fontSize: theme.typography.fontSize.base,
    color: theme.colors.textPrimary,
  },
  input_default: {
    backgroundColor: theme.colors.surface,
  },
  input_filled: {
    backgroundColor: theme.colors.gray50,
    borderColor: 'transparent',
  },
  input_outlined: {
    backgroundColor: 'transparent',
    borderWidth: 2,
  },
  inputError: {
    borderColor: theme.colors.error,
  },
  inputErrorText: {
    fontSize: theme.typography.fontSize.sm,
    color: theme.colors.error,
    marginTop: theme.spacing[1],
  },
  inputHelperText: {
    fontSize: theme.typography.fontSize.sm,
    color: theme.colors.textSecondary,
    marginTop: theme.spacing[1],
  },

  // Card styles
  card: {
    borderRadius: theme.borderRadius['2xl'],
    padding: theme.spacing[6],
  },
  card_default: {
    backgroundColor: theme.colors.surface,
    ...theme.shadows.md,
  },
  card_glass: {
    backgroundColor: `${theme.colors.surface}CC`, // 80% opacity
    borderWidth: 1,
    borderColor: theme.colors.borderLight,
  },
  card_elevated: {
    backgroundColor: theme.colors.surface,
    ...theme.shadows.lg,
  },

  // Text styles
  text: {
    fontFamily: theme.typography.fontFamily.sans[0],
  },
  text_h1: {
    fontSize: theme.typography.fontSize['4xl'],
    fontWeight: theme.typography.fontWeight.bold,
    lineHeight: theme.typography.lineHeight.tight * theme.typography.fontSize['4xl'],
  },
  text_h2: {
    fontSize: theme.typography.fontSize['3xl'],
    fontWeight: theme.typography.fontWeight.bold,
    lineHeight: theme.typography.lineHeight.tight * theme.typography.fontSize['3xl'],
  },
  text_h3: {
    fontSize: theme.typography.fontSize['2xl'],
    fontWeight: theme.typography.fontWeight.semibold,
    lineHeight: theme.typography.lineHeight.snug * theme.typography.fontSize['2xl'],
  },
  text_h4: {
    fontSize: theme.typography.fontSize.xl,
    fontWeight: theme.typography.fontWeight.semibold,
    lineHeight: theme.typography.lineHeight.snug * theme.typography.fontSize.xl,
  },
  text_body: {
    fontSize: theme.typography.fontSize.base,
    lineHeight: theme.typography.lineHeight.normal * theme.typography.fontSize.base,
  },
  text_caption: {
    fontSize: theme.typography.fontSize.sm,
    lineHeight: theme.typography.lineHeight.normal * theme.typography.fontSize.sm,
  },
  text_label: {
    fontSize: theme.typography.fontSize.sm,
    fontWeight: theme.typography.fontWeight.medium,
    lineHeight: theme.typography.lineHeight.normal * theme.typography.fontSize.sm,
  },
  text_primary: { color: theme.colors.textPrimary },
  text_secondary: { color: theme.colors.textSecondary },
  text_tertiary: { color: theme.colors.textTertiary },
  text_inverse: { color: theme.colors.textInverse },
  text_accent: { color: theme.colors.accent },
  text_success: { color: theme.colors.success },
  text_warning: { color: theme.colors.warning },
  text_error: { color: theme.colors.error },
  text_light: { fontWeight: theme.typography.fontWeight.light },
  text_normal: { fontWeight: theme.typography.fontWeight.normal },
  text_medium: { fontWeight: theme.typography.fontWeight.medium },
  text_semibold: { fontWeight: theme.typography.fontWeight.semibold },
  text_bold: { fontWeight: theme.typography.fontWeight.bold },

  // Account type selector styles
  accountTypeContainer: {
    marginBottom: theme.spacing[6],
  },
  accountTypeOption: {
    flexDirection: 'row',
    alignItems: 'center',
    padding: theme.spacing[4],
    borderWidth: 1,
    borderColor: theme.colors.borderLight,
    borderRadius: theme.borderRadius.lg,
    marginBottom: theme.spacing[3],
    backgroundColor: theme.colors.surface,
  },
  accountTypeOptionSelected: {
    borderColor: theme.colors.accent,
    backgroundColor: `${theme.colors.accent}10`, // 10% opacity
    ...theme.shadows.sm,
  },
  accountTypeContent: {
    flex: 1,
  },
  accountTypeHeader: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: theme.spacing[1],
  },
  accountTypeFeatures: {
    marginTop: theme.spacing[2],
  },
  radioButton: {
    width: 20,
    height: 20,
    borderRadius: 10,
    borderWidth: 2,
    borderColor: theme.colors.borderMedium,
    marginLeft: theme.spacing[3],
  },
  radioButtonSelected: {
    borderColor: theme.colors.accent,
    backgroundColor: theme.colors.accent,
  },

  // Loading spinner styles
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
});
