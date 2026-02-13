/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        mono: [
          "'JetBrains Mono'",
          "'Fira Code'",
          "ui-monospace",
          "SFMono-Regular",
          "Menlo",
          "Monaco",
          "Consolas",
          "'Liberation Mono'",
          "'Courier New'",
          "monospace",
        ],
      },
      colors: {
        terminal: {
          bg: "var(--bg)",
          fg: "var(--fg)",
          accent: "var(--accent)",
          muted: "var(--muted)",
          border: "var(--border)",
          "bg-secondary": "var(--bg-secondary)",
          dim: "var(--dim)",
        },
      },
      textColor: {
        terminal: "var(--fg)",
        accent: "var(--accent)",
        muted: "var(--muted)",
        dim: "var(--dim)",
      },
      backgroundColor: {
        terminal: "var(--bg)",
        "terminal-secondary": "var(--bg-secondary)",
      },
      borderColor: {
        terminal: "var(--border)",
      },
      animation: {
        blink: "blink 1s step-end infinite",
        "fade-in": "fadeIn 0.3s ease-in",
        glow: "glow 2s ease-in-out infinite alternate",
      },
      keyframes: {
        blink: {
          "0%, 100%": { opacity: "1" },
          "50%": { opacity: "0" },
        },
        fadeIn: {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" },
        },
        glow: {
          "0%": { textShadow: "0 0 5px var(--accent)" },
          "100%": { textShadow: "0 0 20px var(--accent), 0 0 40px var(--accent)" },
        },
      },
    },
  },
  plugins: [],
};
