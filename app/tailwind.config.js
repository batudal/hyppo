/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./views/**/*.html"],
  theme: {
    fontFamily: {
      sans: ['"Inter"', 'sans-serif'],
      mono: ['"Fira Code"', 'monospace'],
    },
    extend: {
      colors: {
        hyppo: {
          primary: '#E11D48',
          50: '#F5F5F6',
          100: '#E9EAF1',
          200: '#C9CDE2',
          300: '#959CC6',
          400: '#4D5688',
          500: '#2D3770',
          600: '#474C67',
          700: '#323755',
          800: '#1E233F',
          900: '#101323',
          950: '#000312',
        },
      }
    },
  },
  plugins: [],
}
