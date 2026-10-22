/** @type {import('tailwindcss').Config} */
export default {
  darkMode: 'class',
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        primary: "#137fec",
        "background-light": "#f6f7f8",
        "background-dark": "#101922",
        "card-light": "#ffffff",
        "card-dark": "#1a2634",
        "text-main": "#111418",
        "text-sub": "#617589",
        "border-light": "#dbe0e6",
        "border-dark": "#2d3b4a",
      },
      fontFamily: {
        display: ["Manrope", "sans-serif"],
      },
      borderRadius: {
        DEFAULT: "0.25rem",
        lg: "0.5rem",
        xl: "0.75rem",
        full: "9999px",
      },
      animation: {
        float: "float 3s ease-in-out infinite",
        "scan-line": "scan-line 2s linear infinite",
        progress: "progress 1.2s ease-in-out infinite",
      },
      keyframes: {
        float: {
          "0%, 100%": { transform: "translateY(0px)" },
          "50%": { transform: "translateY(-8px)" },
        },
        "scan-line": {
          "0%": { top: "8px" },
          "50%": { top: "calc(100% - 8px)" },
          "100%": { top: "8px" },
        },
        progress: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(350%)" },
        },
      },
    },
  },
  plugins: [],
}
