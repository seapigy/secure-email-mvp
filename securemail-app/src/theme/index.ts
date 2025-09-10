// SecureMail Design System - Matching Website Theme
export const colors = {
  // Primary colors matching website
  primary: '#1B1F23',      // Dark base
  secondary: '#3A3F45',    // Mid-tone accents
  accent: '#4CAFEE',       // Links, buttons, highlights
  success: '#28A745',      // Success states, confirmations
  warning: '#FFC107',      // Alerts, warnings
  error: '#DC3545',        // Error states
  background: '#F9FAFB',   // Page background
  surface: '#FFFFFF',      // Cards, panels
  
  // Extended palette for variations
  primaryDark: '#0F1215',
  primaryLight: '#2A2E32',
  secondaryLight: '#4A4F55',
  accentLight: '#6BC5F0',
  accentDark: '#3A9BC7',
  successLight: '#34CE57',
  successDark: '#1E7E34',
  warningLight: '#FFD43B',
  warningDark: '#E0A800',
  errorLight: '#E74C3C',
  errorDark: '#C82333',
  
  // Neutral colors
  gray50: '#F9FAFB',
  gray100: '#F3F4F6',
  gray200: '#E5E7EB',
  gray300: '#D1D5DB',
  gray400: '#9CA3AF',
  gray500: '#6B7280',
  gray600: '#4B5563',
  gray700: '#374151',
  gray800: '#1B1F23',
  gray900: '#0F1215',
  
  // Text colors
  textPrimary: '#1B1F23',
  textSecondary: '#6B7280',
  textTertiary: '#9CA3AF',
  textInverse: '#FFFFFF',
  
  // Border colors
  borderLight: '#E5E7EB',
  borderMedium: '#D1D5DB',
  borderDark: '#9CA3AF',
};

export const typography = {
  fontFamily: {
    sans: ['Inter', 'system-ui', 'sans-serif'],
    mono: ['JetBrains Mono', 'monospace'],
  },
  fontSize: {
    xs: 12,
    sm: 14,
    base: 16,
    lg: 18,
    xl: 20,
    '2xl': 24,
    '3xl': 30,
    '4xl': 36,
    '5xl': 48,
    '6xl': 60,
    '7xl': 72,
    '8xl': 96,
  },
  fontWeight: {
    light: '300',
    normal: '400',
    medium: '500',
    semibold: '600',
    bold: '700',
    extrabold: '800',
    black: '900',
  },
  lineHeight: {
    tight: 1.25,
    snug: 1.375,
    normal: 1.5,
    relaxed: 1.625,
    loose: 2,
  },
};

export const spacing = {
  0: 0,
  1: 4,
  2: 8,
  3: 12,
  4: 16,
  5: 20,
  6: 24,
  8: 32,
  10: 40,
  12: 48,
  16: 64,
  20: 80,
  24: 96,
  32: 128,
  40: 160,
  48: 192,
  56: 224,
  64: 256,
};

export const borderRadius = {
  none: 0,
  sm: 4,
  base: 8,
  md: 12,
  lg: 16,
  xl: 20,
  '2xl': 24,
  '3xl': 32,
  full: 9999,
};

export const shadows = {
  sm: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.05,
    shadowRadius: 2,
    elevation: 1,
  },
  base: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 1 },
    shadowOpacity: 0.1,
    shadowRadius: 3,
    elevation: 2,
  },
  md: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.1,
    shadowRadius: 6,
    elevation: 4,
  },
  lg: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 10 },
    shadowOpacity: 0.15,
    shadowRadius: 15,
    elevation: 8,
  },
  xl: {
    shadowColor: '#000',
    shadowOffset: { width: 0, height: 20 },
    shadowOpacity: 0.25,
    shadowRadius: 25,
    elevation: 12,
  },
  glow: {
    shadowColor: colors.accent,
    shadowOffset: { width: 0, height: 0 },
    shadowOpacity: 0.3,
    shadowRadius: 10,
    elevation: 5,
  },
};

export const animations = {
  duration: {
    fast: 150,
    normal: 300,
    slow: 500,
  },
  easing: {
    easeInOut: 'ease-in-out',
    easeOut: 'ease-out',
    easeIn: 'ease-in',
  },
};

// Component-specific styles
export const components = {
  button: {
    primary: {
      backgroundColor: colors.accent,
      borderRadius: borderRadius.xl,
      paddingVertical: spacing[3],
      paddingHorizontal: spacing[8],
      ...shadows.sm,
    },
    secondary: {
      backgroundColor: 'transparent',
      borderWidth: 2,
      borderColor: colors.accent,
      borderRadius: borderRadius.xl,
      paddingVertical: spacing[3],
      paddingHorizontal: spacing[8],
    },
    disabled: {
      backgroundColor: colors.gray300,
      borderRadius: borderRadius.xl,
      paddingVertical: spacing[3],
      paddingHorizontal: spacing[8],
    },
  },
  input: {
    base: {
      borderWidth: 1,
      borderColor: colors.borderLight,
      borderRadius: borderRadius.lg,
      paddingVertical: spacing[3],
      paddingHorizontal: spacing[4],
      backgroundColor: colors.surface,
      fontSize: typography.fontSize.base,
      color: colors.textPrimary,
    },
    focused: {
      borderColor: colors.accent,
      ...shadows.glow,
    },
    error: {
      borderColor: colors.error,
    },
  },
  card: {
    base: {
      backgroundColor: colors.surface,
      borderRadius: borderRadius['2xl'],
      padding: spacing[6],
      ...shadows.md,
    },
    glass: {
      backgroundColor: `${colors.surface}CC`, // 80% opacity
      borderRadius: borderRadius['2xl'],
      padding: spacing[6],
      borderWidth: 1,
      borderColor: colors.borderLight,
    },
  },
};

export default {
  colors,
  typography,
  spacing,
  borderRadius,
  shadows,
  animations,
  components,
};

