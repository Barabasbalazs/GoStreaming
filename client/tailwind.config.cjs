const { buildErrorMessage } = require('vite');

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        green: '#38CB89',
        buttonGray: '#E5E5E5',
        red: '#FF481F',
        blue: '#38CBCB',
        yellow: '#FFAB00',
      },
      textColor: {
        grey: '#6D6D6D',
        ultragrey: '#333333',
        midgrey: '#757575',
      },
    },
  },
  plugins: [],
};