/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{vue,js,ts,jsx,tsx}",
  ],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: "#7c83ff",
        green: "#31d0aa",
        orange: "#ff8a65",
        background: "#0a0b18",
        panel: "#14162b",
      },
      boxShadow: {
        custom: "0 20px 40px rgba(0,0,0,0.28)",
      },
    },
  },
  plugins: [],
};