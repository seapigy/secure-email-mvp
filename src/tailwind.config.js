/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'primary': '#1E40AF', // blue-900
        'accent': '#F472B6', // pink-400
        'success': '#34D399', // green-400
        'error': '#EF4444', // red-500
        'neutral-dark': '#111827', // gray-900
        'neutral-light': '#F9FAFB', // gray-100
        'text-gray': '#4B5563', // gray-600
      },
      fontFamily: {
        'inter': ['Inter', 'sans-serif'],
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
    },
  },
  plugins: [],
} 