/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'bg-light': '#F5F7FF',
        'primary': '#10B981',
        'primary-dark': '#0EA472',
        'accent': '#8B5CF6',
        'accent-dark': '#7C3AED',
        'neutral-light': '#F9FAFB',
        'neutral-dark': '#111827',
        'text-gray': '#374151',
        'error': '#EF4444',
        'success': '#10B981',
      },
      spacing: {
        '4.5': '1.125rem',
      },
      fontSize: {
        'sm': '0.875rem',
      },
      borderRadius: {
        'lg': '0.625rem',
        '3xl': '1.5rem',
      },
      boxShadow: {
        'card': '0 2px 15px -3px rgba(0,0,0,0.2), 0 10px 20px -2px rgba(0,0,0,0.2)',
        'neumorphic': '0 4px 6px rgba(0, 0, 0, 0.1), inset 0 1px 0 rgba(255, 255, 255, 0.2), 0 1px 3px rgba(0, 0, 0, 0.05)',
        'glassmorphic': '0 8px 32px rgba(0, 0, 0, 0.1)',
      }
    },
  },
  plugins: [],
} 