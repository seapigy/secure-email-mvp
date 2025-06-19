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
        primary: '#10B981',    // Green for primary actions and text
        accent: '#10B981',     // Same green for consistency
        success: '#10B981',    // Green for success states
        error: '#EF4444',      // Red for errors
        'neutral-light': '#F8FAFC', // Light background
        'neutral-dark': '#1E293B',  // Dark background
        'text-gray': '#64748B',     // Text color
      },
      fontFamily: {
        'inter': ['Inter', 'sans-serif'],
      },
      boxShadow: {
        'card': '0px 4px 25px rgba(0, 0, 0, 0.05)',
        'input': '0px 1px 2px rgba(0, 0, 0, 0.05)',
      },
      borderRadius: {
        'xl': '1rem',
        '2xl': '1.5rem',
      },
    },
  },
  plugins: [],
} 