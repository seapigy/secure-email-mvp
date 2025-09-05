/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
                colors: {
            // SecureMail Official Color System
            primary: '#1B1F23',      // Dark base
            secondary: '#3A3F45',    // Mid-tone accents
            accent: '#4CAFEE',       // Links, buttons, highlights
            success: '#28A745',      // Success states, confirmations
            warning: '#FFC107',      // Alerts, warnings
            error: '#DC3545',        // Error states
            background: '#F9FAFB',   // Page background
            surface: '#FFFFFF',      // Cards, panels
            
            // Extended palette for variations
            'primary-dark': '#0F1215',
            'primary-light': '#2A2E32',
            'secondary-light': '#4A4F55',
            'accent-light': '#6BC5F0',
            'accent-dark': '#3A9BC7',
            'success-light': '#34CE57',
            'success-dark': '#1E7E34',
            'warning-light': '#FFD43B',
            'warning-dark': '#E0A800',
            'error-light': '#E74C3C',
            'error-dark': '#C82333',
            
            // Legacy support (mapped to new system)
            secure: {
              50: '#F0F9F4',
              100: '#DCFCE7',
              200: '#BBF7D0',
              300: '#86EFAC',
              400: '#4ADE80',
              500: '#28A745',  // Maps to success
              600: '#1E7E34',
              700: '#15803D',
              800: '#166534',
              900: '#14532D',
            },
            dark: {
              50: '#F9FAFB',   // Maps to background
              100: '#F3F4F6',
              200: '#E5E7EB',
              300: '#D1D5DB',
              400: '#9CA3AF',
              500: '#6B7280',
              600: '#4B5563',
              700: '#374151',
              800: '#1B1F23',  // Maps to primary
              900: '#0F1215',
            }
          },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
      animation: {
        'fade-in': 'fadeIn 0.5s ease-in-out',
        'slide-up': 'slideUp 0.5s ease-out',
        'bounce-gentle': 'bounceGentle 2s infinite',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        bounceGentle: {
          '0%, 100%': { transform: 'translateY(0)' },
          '50%': { transform: 'translateY(-10px)' },
        },
      },
      backdropBlur: {
        xs: '2px',
      },
      boxShadow: {
        'glow': '0 0 20px rgba(34, 197, 94, 0.3)',
        'glow-lg': '0 0 40px rgba(34, 197, 94, 0.4)',
      },
    },
  },
  plugins: [
    require('@tailwindcss/forms'),
    require('@tailwindcss/typography'),
  ],
}
