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
        primary: '#1E40AF',      // Blue for buttons
        accent: '#F472B6',       // Pink for accents
        success: '#34D399',      // Green for success
        error: '#EF4444',        // Red for errors
        'neutral-light': '#F9FAFB', // Light background
        'neutral-dark': '#111827',   // Dark background
        'text-gray': '#4B5563',      // Text color
      },
      fontFamily: {
        'inter': ['Inter', 'sans-serif'],
      },
      boxShadow: {
        'neumorphic': '12px 12px 24px rgba(0, 0, 0, 0.12), -12px -12px 24px rgba(255, 255, 255, 0.6), inset 0px 1px 1px rgba(255, 255, 255, 0.4)',
        'neumorphic-dark': '12px 12px 24px rgba(0, 0, 0, 0.3), -12px -12px 24px rgba(255, 255, 255, 0.05), inset 0px 1px 1px rgba(255, 255, 255, 0.1)',
        'neumorphic-inset': 'inset 2px 2px 5px rgba(0, 0, 0, 0.1), inset -3px -3px 7px rgba(255, 255, 255, 0.7)',
        'neumorphic-inset-dark': 'inset 2px 2px 5px rgba(0, 0, 0, 0.3), inset -3px -3px 7px rgba(255, 255, 255, 0.05)',
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-in-out',
        'slide-up': 'slideUp 0.35s ease-out',
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
      },
    },
  },
  plugins: [],
} 