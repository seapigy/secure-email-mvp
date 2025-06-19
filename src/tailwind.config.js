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