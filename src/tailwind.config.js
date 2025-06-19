/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        'bg-light': '#F8F9FA',
      },
      spacing: {
        '4.5': '1.125rem',
      },
      fontSize: {
        'sm': '0.875rem',
      },
    },
  },
  plugins: [],
} 